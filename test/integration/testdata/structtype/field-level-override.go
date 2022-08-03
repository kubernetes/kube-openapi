package structtype

// +k8s:openapi-gen=true
type FieldLevelOverrideStruct struct {
	// +structType=atomic
	Field DeclaredAtomicStruct

	OtherField int
}
