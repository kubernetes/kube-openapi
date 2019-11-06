package structtype

// +k8s:openapi-gen=true
type GranularStruct struct {
	// +structType=granular
	Field      string
	OtherField int
}
