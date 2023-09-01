package defaults

// +k8s:openapi-gen=true
type Defaulted struct {
	// +default="bar"
	Field string `json:"Field,omitempty"`
	// +default=0
	OtherField int
	// +default=["foo", "bar"]
	List []Item
	// +default={"s": "foo", "i": 5}
	Sub *SubStruct

	OtherSub SubStruct

	// +default={"foo": "bar"}
	Map map[string]Item

	// +default=ref(ConstantValue)
	LocalSymbolReference string `json:"localSymbolReference,omitempty"`
	// +default=ref(k8s.io/kube-openapi/test/integration/testdata/defaults.ConstantValue)
	FullyQualifiedSymbolReference string `json:"fullyQualifiedSymbolReference,omitempty"`
	// +default=ref(k8s.io/kube-openapi/test/integration/testdata/enumtype.FruitApple)
	ExternalSymbolReference string `json:"externalSymbolReference,omitempty"`
	// +default=ref(k8s.io/kube-openapi/test/integration/testdata/enumtype.FruitApple)
	PointerConversionSymbolReference *DefaultedItem `json:"pointerConversionSymbolReference,omitempty"`
	DefaultedAliasSymbolReference    DefaultedItem  `json:"defaultedAliasSymbolReference,omitempty"`
}

const ConstantValue string = "SymbolConstant"

// +default=ref(ConstantValue)
type DefaultedItem string

// +k8s:openapi-gen=true
type Item string

// +k8s:openapi-gen=true
type SubStruct struct {
	S string
	// +default=1
	I int `json:"I,omitempty"`
}
