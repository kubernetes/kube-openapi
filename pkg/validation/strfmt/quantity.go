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
	"k8s.io/apimachinery/pkg/api/resource"
)

type Quantity string

func init() {
	quantity := Quantity("")
	Default.Add("quantity", &quantity, isQuantity)
}

// String converts this value to a string
func (q Quantity) String() string {
	return string(q)
}

// DeepCopyInto copies the receiver into out. out must be non-nil.
func (q *Quantity) DeepCopyInto(out *Quantity) {
	*out = *q
}

// DeepCopy creates a deep copy of Semver
func (q *Quantity) DeepCopy() *Quantity {
	if q == nil {
		return nil
	}
	out := new(Quantity)
	q.DeepCopyInto(out)
	return out
}

// MarshalText turns this instance into text
func (q Quantity) MarshalText() ([]byte, error) {
	return []byte(q), nil
}

// UnmarshalText hydrates this instance from text
func (q *Quantity) UnmarshalText(data []byte) error {
	*(q) = Quantity(data)
	return nil
}

func isQuantity(str string) bool {
	_, err := resource.ParseQuantity(str)
	return err == nil
}
