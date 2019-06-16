// Copyright 2015 go-swagger maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package strfmt

import (
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/globalsign/mgo/bson"
)

func init() {
	var id ObjectId
	// register this format in the default registry
	Default.Add("bsonobjectid", &id, IsBSONObjectID)
}

// IsBSONObjectID returns true when the string is a valid BSON.ObjectId
func IsBSONObjectID(str string) bool {
	return bson.IsObjectIdHex(str)
}

// ObjectId represents a BSON object ID (alias to github.com/globalsign/mgo/bson.ObjectId)
//
// swagger:strfmt bsonobjectid
type ObjectId bson.ObjectId

// NewObjectId creates a ObjectId from a Hex String
func NewObjectId(hex string) ObjectId {
	return ObjectId(bson.ObjectIdHex(hex))
}

// MarshalText turns this instance into text
func (id *ObjectId) MarshalText() ([]byte, error) {
	return []byte(bson.ObjectId(*id).Hex()), nil
}

// UnmarshalText hydrates this instance from text
func (id *ObjectId) UnmarshalText(data []byte) error { // validation is performed later on
	*id = ObjectId(bson.ObjectIdHex(string(data)))
	return nil
}

// Scan read a value from a database driver
func (id *ObjectId) Scan(raw interface{}) error {
	var data []byte
	switch v := raw.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return fmt.Errorf("cannot sql.Scan() strfmt.URI from: %#v", v)
	}

	return id.UnmarshalText(data)
}

// Value converts a value to a database driver value
func (id *ObjectId) Value() (driver.Value, error) {
	return driver.Value(string(*id)), nil
}

func (id *ObjectId) String() string {
	return string(*id)
}

// MarshalJSON returns the ObjectId as JSON
func (id *ObjectId) MarshalJSON() ([]byte, error) {
	return bson.ObjectId(*id).MarshalJSON()
}

// UnmarshalJSON sets the ObjectId from JSON
func (id *ObjectId) UnmarshalJSON(data []byte) error {
	var obj bson.ObjectId
	if err := obj.UnmarshalJSON(data); err != nil {
		return err
	}
	*id = ObjectId(obj)
	return nil
}

// GetBSON returns the hex representation of the ObjectId as a bson.M{} map.
func (id *ObjectId) GetBSON() (interface{}, error) {
	return bson.M{"data": bson.ObjectId(*id).Hex()}, nil
}

// SetBSON sets the ObjectId from raw bson data
func (id *ObjectId) SetBSON(raw bson.Raw) error {
	var m bson.M
	if err := raw.Unmarshal(&m); err != nil {
		return err
	}

	if data, ok := m["data"].(string); ok {
		*id = NewObjectId(data)
		return nil
	}

	return errors.New("couldn't unmarshal bson raw value as ObjectId")
}

// DeepCopyInto copies the receiver and writes its value into out.
func (id *ObjectId) DeepCopyInto(out *ObjectId) {
	*out = *id
}

// DeepCopy copies the receiver into a new ObjectId.
func (id *ObjectId) DeepCopy() *ObjectId {
	if id == nil {
		return nil
	}
	out := new(ObjectId)
	id.DeepCopyInto(out)
	return out
}
