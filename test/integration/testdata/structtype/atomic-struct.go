package structtype

// +k8s:openapi-gen=true
type AtomicStruct struct {
	// +structType=atomic
	Field      string
	OtherField int
}
