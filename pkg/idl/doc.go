/*
Copyright 2018 The Kubernetes Authors.

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

/*
The IDL package describes comment directives that may be applied to API types.
*/
package idl

// ListType annotates a list to further describe how it works.
// Currently, it can have 3 possible values: "atomic", "map", or "set".
// Note that there is no default, and eventually the generation step
// will fail if a list is found that doesn't have that tag. This tag
// MUST be used on lists, or the generation step will fail.
//
// Atomic
//
// The tag looks like this:
//  +listType=atomic
//
// Atomic is a list that should be entirely replaced when changed. It
// should usually be treated as a scalar type. This tag can be used on
// any type of list (struct, scalar, ...).
//
// Using that tag will generate the following OpenAPI extension:
//  "x-kubernetes-list-type": "atomic"
//
// Map
//
// The tag looks like this:
//  +listType=map
//
// These lists are like maps. Order is preserved upon merge. Using the
// map tag on a non-struct will report an error during the generation
// step.
//
// Using that tag will generate the following OpenAPI extension:
//  "x-kubernetes-list-type": "map"
//
// Set
//
// The tag looks like this:
//  +listType=set
//
// Sets are lists that can't have multiple times the same value. Each
// value must be a scalar or an atomic list and can't be map, set or struct.
//
// Using that tag will generate the following OpenAPI extension:
//  "x-kubernetes-list-type": "set"
type ListType string
