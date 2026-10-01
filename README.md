# Kube OpenAPI

This repo is the home for Kubernetes OpenAPI discovery spec generation. The goal 
is to support a subset of OpenAPI features to satisfy kubernetes use-cases but 
implement that subset with little to no assumption about the structure of the 
code or routes. Thus, there should be no kubernetes specific code in this repo. 

Both OpenAPI v2 (Swagger 2.0) and OpenAPI v3 are supported. Kubernetes uses
these packages to generate its API models and to build and serve the specs
behind its `/openapi/v2` and `/openapi/v3` endpoints.

## Components

- `cmd/openapi-gen`: a code generator, built on
  [gengo](https://github.com/kubernetes/gengo), that scans Go types annotated
  with `+k8s:openapi-gen=true` and generates `OpenAPIDefinition` code for
  them. The generation logic lives in `pkg/generators`, which also contains
  the API linting rules (`pkg/generators/rules`).
- `pkg/builder` and `pkg/builder3`: build a complete OpenAPI v2 or v3 spec
  from web service routes and the generated definitions.
- `pkg/handler` and `pkg/handler3`: HTTP handlers that serve the resulting
  specs as JSON or protobuf, with etag and gzip support.
- `pkg/validation/spec` and `pkg/spec3`: the Go types for OpenAPI v2 and v3
  documents. `pkg/validation` is a fork of the
  [go-openapi](https://github.com/go-openapi) libraries and also contains
  document validation (`validate`), error types (`errors`), and string
  formats (`strfmt`).

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for instructions on how to contribute.
