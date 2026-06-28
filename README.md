# <operator-name>

`<operator-name>` is a Kubernetes operator for `<managed system or workload>`.
It reconciles `<Kind>` custom resources into `<owned Kubernetes resources or
external system state>` and reports current state through Kubernetes status
conditions.

Replace bracketed placeholders before publishing this README.

## Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Development](#development)
- [Release](#release)
- [Contributing](#contributing)
- [Security](#security)
- [License](#license)

## Features

- Reconciles `<api-group>/<version>` `<Kind>` resources.
- Manages `<owned resources, external resources, or integration points>`.
- Publishes operator status with Kubernetes conditions and Events.
- Ships a controller image and Helm chart for cluster installation.

## Prerequisites

- [mise](https://mise.jdx.dev) — provisions every pinned tool from `mise.toml` +
  `mise.lock`: Go, Moon, the Kubebuilder operator CLIs (`controller-gen`,
  `setup-envtest`, `kubebuilder`, `kubectl`, `helm`, `chainsaw`), the local dev
  stack (`ko`, `tilt`, `ctlptl`, `kind`), and `melange`/`apko`/`cosign` for the
  release image. Run `mise install` once; there is nothing else to install by hand.
- Docker or another container runtime for local image builds.
- A Kubernetes cluster for deployed testing. Kind is recommended for local e2e
  checks.

Tool versions live in `mise.toml`; `mise.lock` records a per-platform download URL
and checksum for each (and, for the aqua-backed CLIs, cosign/SLSA/GitHub-attestation
verification). `mise install` runs with `locked = true`, so it **fails closed** if a
tool lacks a pre-resolved, checksummed entry for the current platform. To bump a
tool, edit its version in `mise.toml`, run
`mise lock --platform linux-x64,linux-arm64,macos-x64,macos-arm64`, and commit
`mise.toml` + `mise.lock` together.

## Installation

Install the released Helm chart:

```sh
helm install <release-name> oci://ghcr.io/<org>/<repo>/chart \
  --version <version> \
  --namespace <namespace> \
  --create-namespace
```

Install from a local checkout:

```sh
moon run root:deploy
```

Useful local deployment overrides:

```sh
HELM_RELEASE=<release-name> HELM_NAMESPACE=<namespace> moon run root:deploy
IMG=ghcr.io/<org>/<repo>:<tag> moon run root:deploy
LOCAL_IMAGE=true IMG=<local-image>:<tag> moon run root:deploy
```

Uninstall the local deployment:

```sh
moon run root:undeploy
```

## Usage

Create a custom resource:

```sh
kubectl apply -f - <<'EOF'
apiVersion: <api-group>/<version>
kind: <Kind>
metadata:
  name: example
  namespace: <namespace>
spec:
  # Add the fields this operator reconciles.
EOF
```

Inspect the reconciled resource and its status:

```sh
kubectl get <resource-plural> example -n <namespace> -o yaml
kubectl describe <resource-plural> example -n <namespace>
```

Describe the expected spec fields, owned resources, and status conditions here
after the project defines its API contract.

## Configuration

Runtime configuration is exposed through the Helm chart values in
`charts/<chart-directory>/values.yaml`.

Common settings to document for this operator:

- Controller image repository, tag, or digest.
- Resource requests and limits.
- Metrics and health probe settings.
- Leader election settings.
- Optional Kyverno image verification for the released controller image.
- Any provider credentials, watched namespaces, or external service endpoints.

If Kyverno is installed in the target cluster, the chart can install an
optional `ClusterPolicy` that verifies the operator image's GitHub Artifact
Attestation:

```sh
helm install <release-name> oci://ghcr.io/<org>/<repo>/chart \
  --version <version> \
  --namespace <namespace> \
  --create-namespace \
  --set kyverno.imageVerification.enabled=true
```

## Development

Regenerate code and manifests after API changes:

```sh
moon run root:generate
```

Run the normal local checks and tests:

```sh
moon run root:check
moon run root:test
```

Run the full local e2e smoke path:

```sh
moon run root:test-e2e
```

Run the local development stack with Kind, ctlptl, Tilt, and ko:

```sh
moon run root:dev-up
```

Stop and delete the local development stack:

```sh
moon run root:dev-down
```

## Release

Releases publish the controller binaries, container image, and Helm chart. Use
the release workflow and Release Please configuration in this repository as the
source of truth for versioning and publication.

Install released charts from:

```text
oci://ghcr.io/<org>/<repo>/chart
```

## Contributing

Before opening a pull request:

- Keep generated API code and CRDs up to date.
- Add envtest coverage for API and reconciler behavior.
- Add or update Chainsaw coverage only for installed-operator behavior.
- Run `moon run root:check` and `moon run root:test`.

## Security

Report security issues through the project's private disclosure process. Do not
open public issues for vulnerabilities until maintainers have had a chance to
triage them.

## License

TODO: Add the project's license name and license file before publishing.
