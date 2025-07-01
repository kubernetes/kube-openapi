package structtype

// +k8s:openapi-gen=true
type Struct struct {
	Field      ContainedStruct
	OtherField int
}

func (Struct) OpenAPIModelName() string {
	return "com.example.Struct"
}

// +k8s:openapi-gen=true
type ContainedStruct struct{}

func (ContainedStruct) OpenAPIModelName() string {
	return "com.example.ContainedStruct"
}

// +k8s:openapi-gen=true
type AtomicStruct struct {
	Field int
}

func (AtomicStruct) OpenAPIModelName() string {
	return "com.example.AtomicStruct"
}
