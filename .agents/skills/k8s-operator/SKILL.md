---
name: k8s-operator
description: Use when building, reviewing, or testing Kubernetes operators in this repository, especially Kubebuilder/controller-runtime APIs, CRDs, reconcile loops, owned child resources, status conditions, envtest specs, Moon tasks, or Kind-backed e2e smoke tests.
---

# Kubernetes Operator Work

Use this skill to keep operator work prototype-friendly but correct enough to
teach the right patterns. Prefer the smallest working slice that proves the
workflow, then tighten behavior from what the prototype exposes.

## API Shape

- If `0` is a valid value, do not combine a scalar field with `omitempty` and a
  CRD default. Use a pointer field and a helper that supplies the in-controller
  default when the pointer is nil.
- Keep defaults close to the reconciler too. API-server defaulting helps cluster
  objects, but tests and typed clients may construct objects directly.
- If child resources are named after the custom resource or the custom resource
  name is reused in labels, intentionally validate the CR name length or derive
  child names and selector labels from a stable label-safe hash.
- After changing API field types, regenerate deepcopy code and manifests. A Go
  pointer change may mainly show up in `zz_generated.deepcopy.go`.

## Reconcile Ownership

- Name simple owned children after the custom resource unless there is a real
  reason not to. It keeps smoke tests and cleanup obvious.
- Put stable labels on every owned child and use a narrower selector label set
  for pods/services when needed.
- Always set controller references for owned resources and register `.Owns(...)`
  watches in `SetupWithManager`.
- In `CreateOrPatch`, set the fields the controller owns instead of replacing
  entire specs by default. Broad spec replacement can fight API defaults and
  makes idempotence harder to reason about.
- When config changes should restart pods, hash the effective config into a pod
  template annotation.
- If status depends on a just-created or patched child, refetch the child before
  deriving status so generation and defaulted fields are current.
- Demo workloads should be compatible with Restricted Pod Security by default:
  use an unprivileged image and high port, set pod/container security contexts,
  drop Linux capabilities, disable privilege escalation, set seccomp, and add
  resource requests.
- RBAC markers should match current behavior. Do not leave generated primary-CR
  write verbs or finalizer verbs unless the reconciler actually uses them.

## Status

- Never mark a parent resource available from stale child status. For a managed
  Deployment, require `deployment.Status.ObservedGeneration >= deployment.Generation`
  before trusting ready replica counts.
- For positive availability, require the child availability condition as well
  as the desired ready replica count. Ready counts alone can be stale or
  incomplete.
- Set the parent condition `observedGeneration` to the custom resource
  generation.
- Patch status only when it changed. Noisy status writes create avoidable
  reconciles and hide meaningful transitions.
- Treat scale-to-zero as available when it is explicitly supported and the child
  status is fresh.

## Tests

- Envtest should prove owned child creation, owner refs, labels/selectors,
  desired images/ports/replicas, mounted config, restricted-compatible pod
  settings, resource requests, rollout hash annotations, and update behavior.
- Add stale-status tests when parent status depends on child status. Make the
  child available, change the parent spec, reconcile, and assert the parent does
  not stay available until the child observes the new generation.
- When manually setting Deployment status in envtest, set both
  `Status.ObservedGeneration = Generation` and a `DeploymentAvailable=True`
  condition before expecting the parent to report `Available=True`.
- Add a scale-to-zero test when zero replicas are supported, especially when the
  API field is optional or defaulted.

## E2E And Moon

- Moon is the task front door. Do not reintroduce generated Makefile paths into
  template tests or docs.
- Keep one runnable smoke path that installs the CRD/controller, applies the
  sample custom resource in a Restricted-enforced namespace, waits for the
  parent condition, and verifies the owned workload/service exist.
- For Kind-backed tests with locally loaded images, ensure the Deployment uses
  the exact loaded tag and `imagePullPolicy: IfNotPresent` before readiness
  waits. Default `:latest` behavior can force remote pulls.
- Prefer e2e task cleanup that removes only a Kind cluster the task created; do
  not delete a pre-existing developer cluster with the same name.
- Cluster-scoped e2e resources such as `ClusterRoleBinding` must be created
  idempotently or cleared before creation, then cleaned up in suite teardown.

## Verification

For ordinary controller/API changes, run:

```sh
moon run root:generate
moon run root:manifests
go test ./...
moon run root:lint
moon ci --summary minimal
git diff --check
```

When changing e2e wiring, also run:

```sh
go test -tags=e2e ./test/e2e -run '^$'
moon run root:test-e2e
```
