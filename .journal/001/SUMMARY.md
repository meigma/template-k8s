---
id: 001
title: Working example operator prototype
date: 2026-05-19
status: complete
repos_touched: [template-k8s]
related_sessions: []
---

## Goal
Turn the repository's generated example operator into a small working prototype that can exercise real template workflows without over-designing the final architecture.

## Outcome
The goal was met. PR #8 landed a working `NginxDeployment` operator prototype on `master` and removed the placeholder Widget scaffold.

## Key Decisions
- Use a single `NginxDeployment` API -> enough surface area to test CRD generation, controller reconciliation, status, samples, RBAC, envtest, and Kind e2e without adding unrelated operator complexity.
- Keep children same-named with the custom resource -> simple for smoke tests and owner-reference cleanup, then validate CR names as Service-safe DNS labels to avoid delayed reconciliation failures.
- Use pointer `spec.replicas` -> preserves explicit scale-to-zero with API defaulting and typed clients.
- Gate parent availability on fresh Deployment status -> avoids marking the CR available from stale child `ReadyReplicas` after spec changes.
- Default to an unprivileged nginx image on port 8080 -> keeps the sample compatible with Restricted Pod Security.
- Capture review lessons in a repo-local `k8s-operator` skill -> future operator work can reuse the patterns without turning the README into a manual.

## Changes
- `.agents/skills/k8s-operator/SKILL.md` - added a repo-local operator skill covering API shape, ownership, status, envtest, e2e, RBAC, Pod Security, and validation lessons.
- `api/v1alpha1/nginxdeployment_types.go` - replaced Widget with `NginxDeployment`, including image, replicas, port, config, status, CRD validation, and status subresource markers.
- `internal/controller/nginxdeployment_controller.go` - implemented reconciliation for owned ConfigMap, Deployment, and ClusterIP Service resources plus config-hash rollouts and availability status projection.
- `internal/controller/nginxdeployment_controller_test.go` - added envtest coverage for owned children, updates, defaults, stale Deployment status, scale-to-zero, restricted workload settings, name validation, and config-size validation.
- `config/**` - regenerated CRD, RBAC, sample, and kustomization artifacts for the nginx API and controller.
- `moon.yml` - added Moon task wiring for install, deploy, undeploy, docker build, and Kind-backed e2e smoke.
- `test/e2e/**` and `test/utils/utils.go` - moved e2e onto Moon tasks, added sample CR reconciliation under a Restricted-enforced namespace, handled local Kind images, and cleaned cluster-scoped test resources.
- `README.md` - updated the template state and documented the current prototype constraints.

## Open Threads
- No controller finalizer, webhook, Ingress, TLS, advanced nginx behavior, or ConfigMapRef pattern was added in this slice.
- Future API work can replace same-named children with normalized/hash-derived child names if broader CR name support becomes useful.
- The inline config cap is intentionally small for the prototype; a reference-based config model may be better for a real operator.

## References
- PR: https://github.com/meigma/template-k8s/pull/8
- Merge commit: `5f9fe08` (`feat: add working nginx deployment operator prototype (#8)`)
- Kubernetes Service naming: https://kubernetes.io/docs/concepts/services-networking/service/
- Kubernetes object names: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/
- Kubernetes ConfigMaps: https://kubernetes.io/docs/concepts/configuration/configmap/
- Kubernetes RBAC good practices: https://kubernetes.io/docs/concepts/security/rbac-good-practices/
- Kubernetes Pod Security Standards: https://kubernetes.io/docs/concepts/security/pod-security-standards/

## Lessons
- Child resource naming constraints should be encoded at the CRD boundary when the reconciler intentionally reuses the CR name.
- Parent status should not trust child status until the child has observed its own latest generation.
- Template examples should teach least-privilege RBAC and Restricted-compatible workloads even when the operator itself is only a prototype.
