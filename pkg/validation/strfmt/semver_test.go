/*
Copyright 2025 The Kubernetes Authors.

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

package strfmt

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSemver(t *testing.T) {
	validSemvers := []string{
		"1.0.0",
		"2.3.4",
		"1.0.0-alpha",
		"1.0.0-alpha.1",
		"1.0.0-0.3.7",
		"1.0.0-x.7.z.92",
		"1.0.0-beta+exp.sha.5114f85",
		"1.0.0+20130313144700",
		"1.0.0-beta+exp.sha.5114f85",
		"11.200.300-alpha+meta",
	}

	invalidSemvers := []string{
		"",              // empty
		"1",             // missing minor and patch
		"1.0",           // missing patch
		"1.a.2",         // non-numeric version parts
		"1.0.0beta",     // prerelease without hyphen
		"v1.0.0",        // with v prefix
		"1.0.0-",        // empty prerelease
		"1.0.0+",        // empty build metadata
		"1.0.0-+",       // empty prerelease and build metadata
		"1.0.0-alpha_1", // invalid character in prerelease
		"1.0.0+alpha_1", // invalid character in build metadata
		"-1.0.0",        // negative major version
		"1.-2.0",        // negative minor version
		"1.0.-3",        // negative patch version
	}

	for _, v := range validSemvers {
		t.Run(v, func(t *testing.T) {
			assert.True(t, isSemver(v), "Expected %q to be a valid semver", v)
		})
	}

	for _, v := range invalidSemvers {
		t.Run(v, func(t *testing.T) {
			assert.False(t, isSemver(v), "Expected %q to be an invalid semver", v)
		})
	}
}

func TestDeepCopySemver(t *testing.T) {
	semver := Semver("1.0.0-alpha+001")
	in := &semver

	out := new(Semver)
	in.DeepCopyInto(out)
	assert.Equal(t, in, out)

	out2 := in.DeepCopy()
	assert.Equal(t, in, out2)

	var inNil *Semver
	out3 := inNil.DeepCopy()
	assert.Nil(t, out3)
}

func TestSemverJSON(t *testing.T) {
	semver := Semver("1.0.0-alpha+001")

	// Test marshaling
	data, err := json.Marshal(semver)
	assert.NoError(t, err)
	assert.Equal(t, `"1.0.0-alpha+001"`, string(data))

	// Test unmarshaling
	var s Semver
	err = json.Unmarshal(data, &s)
	assert.NoError(t, err)
	assert.Equal(t, semver, s)
}
