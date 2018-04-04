# Kube OpenAPI

This repo is the home for Kubernetes OpenAPI discovery spec generation. The goal 
is to support a subset of OpenAPI features to satisfy kubernetes use-cases but 
implement that subset with little to no assumption about the structure of the 
code or routes. Thus, there should be no kubernetes specific code in this repo. 


There are two main parts: 
 - A model generator that goes through .go files, find and generate model 
definitions. 
 - The spec generator that is responsible for dynamically generate 
the final OpenAPI spec using web service routes or combining other 
OpenAPI/Json specs.

## Development

After cloning this repo, within the top-level directory do the following to
resolve dependencies:

```
$ go get -t -v ./...
```

### Unit testing

From top-level directory:

```
$ go test ./pkg/...
```

### Testing with Kubernetes

Assuming you already have a `k8s.io/kubernetes` repo, and `k8s.io/kube-openapi` repo:

* Remove vendored `kube-openapi` within `k8s.io/kubernetes`. From top-level directory:


```
$ rm -rf vendor/k8s.io/kube-openapi
```

* Link to `kube-openapi` repo. From top-level `k8s.io/kubernetes` repo:

```
$ ln -s <path to kubeopen-api root>/k8s.io/kube-openapi vendor/k8s.io/kube-openapi
```

* Run unit tests in `k8s.io/kubernetes`:

```
$ make test
$ make test WHAT=k8s.io/kubernetes/vendor/k8s.io/kube-openapi/pkg/generators
$ make test WHAT=k8s.io/kubernetes/vendor/k8s.io/kube-openapi/pkg/builder
$ go test ./vendor/k8s.io/kube-openapi/pkg/...
```

* Run integration tests in `k8s.io/kubernetes`:

```
$ make test-integration
$ make test-integration WHAT=k8s.io/kubernetes/test/integration/apiserver
```
