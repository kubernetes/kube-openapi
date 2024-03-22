package valuevalidation

// Dummy type to test the openapi-gen API rule checker.
// The API rule violations are in format of:
// -> +k8s:validation:[validation rule]=[value]

// +k8s:openapi-gen=true
type NameFormats struct {
	// +k8s:validation:nameFormat="dns1123Label"
	dns string
	// +k8s:validation:nameFormat="dns1123Subdomain"
	subdomain string
	// +k8s:validation:nameFormat="httpPath"
	path string
	// +k8s:validation:nameFormat="qualifiedName"
	qualified string
	// +k8s:validation:nameFormat="wildcardDNS1123Subdomain"
	wildcard string
	// +k8s:validation:nameFormat="cIdentifier"
	identifier string
	// +k8s:validation:nameFormat="dns1035Label"
	label string
	// +k8s:validation:nameFormat="labelValue"
	value string
}
