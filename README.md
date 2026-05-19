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
- Chainsaw

Quick smoke checks:

```sh
kubebuilder version
golangci-lint version
controller-gen --version
setup-envtest version
kubectl version --client=true
helm version --short
chainsaw version
```

## Starter Operator

The current scaffold is a Kubebuilder `go/v4` operator with one prototype API:

- group: `example.meigma.io`
- version: `v1alpha1`
- kind: `NginxDeployment`

The first prototype target is a minimal nginx deployment operator. Each
`NginxDeployment` reconciles to an owned ConfigMap, Deployment, and ClusterIP
Service, with Deployment readiness projected back into status.

Because this prototype names child resources after the custom resource, a
`NginxDeployment` name must be a Service-safe DNS label no longer than 63
characters. Inline nginx config is capped at 64 KiB before it is copied into an
owned ConfigMap.

## Moon Tasks

Moon is the template task front door:

```sh
moon run root:manifests
moon run root:generate
moon run root:fmt
moon run root:vet
moon run root:test
moon run root:lint
moon run root:lint-fix
moon run root:lint-config
moon run root:chainsaw-lint
moon run root:build
moon run root:run
moon run root:test-e2e
```

These tasks are the Moon equivalents of Kubebuilder's generated `make manifests`
and `make generate` targets plus the basic Go format, vet, envtest, lint,
build, local run, and Kind-backed e2e smoke paths. The e2e task builds the local
manager image, loads it into Kind, and runs the Chainsaw smoke tests in
`test/chainsaw/`. They use tools from the activated Proto-managed environment.
