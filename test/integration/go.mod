module k8s.io/kube-openapi/test/integration

go 1.21

require (
	github.com/emicklei/go-restful/v3 v3.8.0
	github.com/getkin/kin-openapi v0.76.0
	github.com/onsi/ginkgo/v2 v2.1.4
	github.com/onsi/gomega v1.19.0
	k8s.io/kube-openapi v0.0.0
	k8s.io/utils v0.0.0-20230726121419-3b25d923346b
)

require (
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.6 // indirect
	github.com/go-openapi/jsonreference v0.20.1 // indirect
	github.com/go-openapi/swag v0.22.3 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/gnostic-models v0.6.8 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/mod v0.7.0 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	golang.org/x/tools v0.4.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/gengo v0.0.0-20230829151522-9cce18d56c01 // indirect
	k8s.io/klog/v2 v2.80.1 // indirect
)

// Use the relative local source of the github.com/google/cadvisor library to build
replace k8s.io/kube-openapi => ../../
