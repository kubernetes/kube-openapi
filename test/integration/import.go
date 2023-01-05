// Copyright 2023 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build tools

package integration

// This file, which is not built (because of the build constraint), is just
// there to import "k8s.io/kube-openapi/cmd/openapi-gen", so that
// "go mod tidy" will not remove the openapi-gen's dependencies.
// openapi-gen is indeed built as part of the integration tests.
import _ "k8s.io/kube-openapi/cmd/openapi-gen"
