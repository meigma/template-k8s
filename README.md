# template-k8s

`template-k8s` is the Meigma starter for Kubernetes operator projects.

This first slice only wires the local operator-development toolchain. It pins
Kubebuilder through Proto and lets direnv activate the repo environment when
you enter the checkout.

## Local Environment

Prerequisites:

- [proto](https://moonrepo.dev/proto)
- [direnv](https://direnv.net/)

Enable the environment:

```sh
direnv allow
proto status
```

The local toolchain is pinned in `.prototools` and installed through repo-local
Proto plugins under `.moon/proto/`.

Current tools:

- Kubebuilder
- golangci-lint
- controller-gen
- setup-envtest
- kubectl
- Helm

Quick smoke checks:

```sh
kubebuilder version
golangci-lint version
controller-gen --version
setup-envtest version
kubectl version --client=true
helm version --short
```

## Starter Operator

The current scaffold is a stock Kubebuilder `go/v4` operator with one dummy API:

- group: `example.meigma.io`
- version: `v1alpha1`
- kind: `Widget`

Generated workflow files such as the Makefile and Kustomize manifests are still
present for now. They are intentionally left as scaffold output so the next
slice can replace them with the template workflow.

## Moon Tasks

Moon is the template task front door:

```sh
moon run root:manifests
moon run root:generate
moon run root:fmt
moon run root:vet
moon run root:lint
moon run root:lint-fix
moon run root:lint-config
moon run root:build
moon run root:run
```

These tasks are the Moon equivalents of Kubebuilder's generated `make manifests`
and `make generate` targets plus the basic Go format, vet, lint, build, and
local run targets. They use tools from the activated Proto-managed environment.
