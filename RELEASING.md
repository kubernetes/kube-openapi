# Releasing kube-openapi into kubernetes/kubernetes

kube-openapi is consumed by `kubernetes/kubernetes` (k/k) via Go modules. A
"release" is a commit-pin bump of `k8s.io/kube-openapi` in k/k's `go.mod`,
followed by regeneration of OpenAPI artifacts.

Sample PR: https://github.com/kubernetes/kubernetes/pull/138931

Placeholders:
- `<NEW-SHA>` — new kube-openapi commit to pin. Typically the latest commit on
  `master`: https://github.com/kubernetes/kube-openapi/commits/master
- `<OLD-SHA>` — previous pin (read from the `k8s.io/kube-openapi` version
  string in k/k `go.mod`; the trailing hex is the commit SHA).

## Steps

Run from the root of a k/k checkout:

```bash
# 1. Pin the new commit. Rewrites go.mod with the new kube-openapi version.
hack/pin-dependency.sh k8s.io/kube-openapi <NEW-SHA>

# 2. Rebuild vendor. Touches every staging go.mod/go.sum and vendor/.
hack/update-vendor.sh

# 3. Regenerate zz_generated.openapi.go files.
hack/update-codegen.sh

# 4. Regenerate api/openapi-spec/ (etcd must be in PATH, see hack/install-etcd.sh).
hack/update-openapi-spec.sh
```

## Commit split

Land as two commits, in order:

1. `Bump kube-openapi to <short-SHA>`
2. `Regenerate OpenAPI`

## PR

Include a compare link in the PR description as the change list:

```
https://github.com/kubernetes/kube-openapi/compare/<OLD-SHA>...<NEW-SHA>
```

