/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package handler_test

import (
	"errors"
	"testing"

	"k8s.io/kube-openapi/pkg/internal/handler"
)

func TestCache(t *testing.T) {
	calledCount := 0
	expectedBytes := []byte("ABC")
	cacheObj := handler.HandlerCache{
		BuildCache: func() ([]byte, error) {
			calledCount++
			return expectedBytes, nil
		},
	}
	bytes, _ := cacheObj.Get()
	if string(bytes) != string(expectedBytes) {
		t.Fatalf("got value of %q from cache (expected %q)", bytes, expectedBytes)
	}
	cacheObj.Get()
	if calledCount != 1 {
		t.Fatalf("expected BuildCache to be called once (called %d times)", calledCount)
	}
}

func TestCacheError(t *testing.T) {
	cacheObj := handler.HandlerCache{
		BuildCache: func() ([]byte, error) {
			return nil, errors.New("cache error")
		},
	}
	_, err := cacheObj.Get()
	if err == nil {
		t.Fatalf("expected non-nil err from cache.Get()")
	}
}

func TestCacheRefresh(t *testing.T) {
	// check that returning an error while having no prior cached value results in a nil value from cache.Get()
	cacheObj := (&handler.HandlerCache{}).New(func() ([]byte, error) {
		return nil, errors.New("returning nil bytes")
	})
	// make multiple calls to Get() to ensure errors are preserved across subsequent calls
	for i := 0; i < 4; i++ {
		value, err := cacheObj.Get()
		if value != nil {
			t.Fatalf("expected nil bytes (got %s)", value)
		}
		if err == nil {
			t.Fatalf("expected non-nil err from cache.Get()")
		}
	}
	// check that we can call New() multiple times and get the last known cache value while also returning any errors
	lastGoodVal := []byte("last good value")
	cacheObj = cacheObj.New(func() ([]byte, error) {
		return lastGoodVal, nil
	})
	// call Get() once, so lastGoodVal is cached
	cacheObj.Get()
	for i := 0; i < 4; i++ {
		cacheObj = cacheObj.New(func() ([]byte, error) {
			return nil, errors.New("check that c.bytes is preserved across New() calls")
		})
		value, err := cacheObj.Get()
		if err == nil {
			t.Fatalf("expected non-nil err from cache.Get()")
		}
		if string(value) != string(lastGoodVal) {
			t.Fatalf("expected previous value for cache to be returned (got %s, expected %s)", value, lastGoodVal)
		}
	}
	// check that if we successfully renew the cache the old last known value is flushed
	newVal := []byte("new good value")
	cacheObj = cacheObj.New(func() ([]byte, error) {
		return newVal, nil
	})
	value, err := cacheObj.Get()
	if err != nil {
		t.Fatalf("expected nil err from cache.Get()")
	}
	if string(value) != string(newVal) {
		t.Fatalf("got value of %s from cache (expected %s)", value, newVal)
	}
}
