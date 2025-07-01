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

	"github.com/blang/semver/v4"
)

func init() {
	semver := Semver("")
	Default.Add("semver", &semver, isSemver)
}

// Semver represents a semantic version string that follows the semver.org specification.
//
// swagger:strfmt semver
type Semver string

// MarshalText turns this instance into text
func (s Semver) MarshalText() ([]byte, error) {
	return []byte(s), nil
}

// UnmarshalText hydrates this instance from text
func (s *Semver) UnmarshalText(data []byte) error {
	*(s) = Semver(data)
	return nil
}

// String converts this value to a string
func (s Semver) String() string {
	return string(s)
}

// MarshalJSON returns the Semver as JSON
func (s Semver) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

// UnmarshalJSON sets the Semver from JSON
func (s *Semver) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	return s.UnmarshalText([]byte(str))
}

// DeepCopyInto copies the receiver into out. out must be non-nil.
func (s *Semver) DeepCopyInto(out *Semver) {
	*out = *s
}

// DeepCopy creates a deep copy of Semver
func (s *Semver) DeepCopy() *Semver {
	if s == nil {
		return nil
	}
	out := new(Semver)
	s.DeepCopyInto(out)
	return out
}

func isSemver(str string) bool {
	_, err := semver.Parse(str)
	return err == nil
}
