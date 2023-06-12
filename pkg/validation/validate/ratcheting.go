package validate

import (
	"fmt"
	"reflect"

	"k8s.io/kube-openapi/pkg/validation/spec"
	"k8s.io/kube-openapi/pkg/validation/strfmt"
)

type ratchetingSchemaValidator struct {
	Schema       *spec.Schema
	Root         interface{}
	Path         string
	KnownFormats strfmt.Registry
	Options      []Option
}

func NewRatchetingSchemaValidator(schema *spec.Schema, rootSchema interface{}, root string, formats strfmt.Registry, options ...Option) *ratchetingSchemaValidator {
	return &ratchetingSchemaValidator{
		Schema:       schema,
		Root:         rootSchema,
		Path:         root,
		KnownFormats: formats,
		Options:      options,
	}
}

func (r *ratchetingSchemaValidator) ValidateUpdate(old, new interface{}) *Result {
	opts := append([]Option{
		r.enableRatchetingOption(old),
	}, r.Options...)

	s := NewSchemaValidator(r.Schema, r.Root, r.Path, r.KnownFormats, opts...)

	res := s.Validate(new)

	if res.IsValid() {
		return res
	}

	if reflect.DeepEqual(old, new) {
		//!TODO: only consider errors with paths on "."
		newRes := &Result{}
		newRes.MergeAsWarnings(res)
		return newRes
	}

	return res
}

func (r *ratchetingSchemaValidator) enableRatchetingOption(old interface{}) Option {
	stub := &ratchetThunk{
		oldValue:                  old,
		ratchetingSchemaValidator: r,
	}

	return func(svo *SchemaValidatorOptions) {
		svo.subIndexValidator = stub.SubIndexValidator
		svo.subPropertyValidator = stub.SubPropertyValidator
	}
}

type ratchetThunk struct {
	*ratchetingSchemaValidator
	oldValue interface{}
}

func (r ratchetThunk) Applies(value interface{}, kind reflect.Kind) bool {
	return true
}

func (s ratchetThunk) SetPath(path string) {
	s.ratchetingSchemaValidator.Path = path
}

// Validate validates the value.
func (r ratchetThunk) Validate(value interface{}) *Result {
	return r.ValidateUpdate(r.oldValue, value)
}

func (r ratchetThunk) SubPropertyValidator(field string, sch *spec.Schema) valueValidator {
	// Find correlated old value
	if asMap, ok := r.oldValue.(map[string]interface{}); ok {
		return ratchetThunk{
			oldValue:                  asMap[field],
			ratchetingSchemaValidator: NewRatchetingSchemaValidator(sch, r.Root, r.Path+"."+field, r.KnownFormats, r.Options...),
		}
	}

	return NewSchemaValidator(sch, r.ratchetingSchemaValidator.Root, r.ratchetingSchemaValidator.Path+"."+field, r.ratchetingSchemaValidator.KnownFormats, r.ratchetingSchemaValidator.Options...)

}

func (r ratchetThunk) SubIndexValidator(index int, sch *spec.Schema) valueValidator {
	//!TODO: implement slice ratcheting which considers the x-kubernetes extensions
	// Some notes
	// 1. Check if this schema uses map-keys
	// 2. If it does not, use index
	// 3. If it does, find other entry with map

	// if the list is a set, just find the element in the old value which equals it
	// if it exists. Sets can only be used on scalars so this is find

	return NewSchemaValidator(sch, r.Root, fmt.Sprintf("%s[%d]", r.Path, index), r.KnownFormats, r.Options...)

}
