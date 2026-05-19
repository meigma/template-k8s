---
id: 006
title: Evaluate removing Kustomize
date: 2026-05-19
status: complete
repos_touched: [template-k8s]
related_sessions: []
---

## Goal
Evaluate practical options for removing Kustomize from the Kubernetes operator template, then replace the Kustomize deployment surface with a centralized Helm chart if it proved viable.

## Outcome
The goal was met. PR #13 was squash-merged, and `master` now uses `charts/template-k8s` as the deployment surface with no tracked `config/` Kustomize tree.

## Key Decisions
- Use Helm as the single deployment manifest surface -> this keeps operator installs readable and avoids split Kustomize overlays.
- Keep generated CRDs in the chart `crds/` directory -> Helm can install CRDs first, while Moon applies them before Helm deploys to handle upgrades explicitly.
- Generate RBAC only for validation -> chart RBAC stays templated, and the test compares it against temporary controller-gen output to avoid reintroducing checked-in generated RBAC.
- Set chart `kubeVersion` to `>= 1.29.0-0` -> the CRD uses stable Kubernetes CEL validation rules, so unsupported clusters should fail early.
- Defer chart publication -> Release Please updates `Chart.yaml`, but publishing an OCI chart remains a future release-workflow decision.

## Changes
- `charts/template-k8s` - added a Helm v4-oriented operator chart with CRDs, values schema, manager deployment, RBAC, metrics Service, validation helpers, and release-version markers.
- `.agents/skills/helm` - added repo-local Helm chart guidance for future operator chart work.
- `moon.yml` - routes generation, validation, install/deploy, CI, and e2e through the Helm chart and chart CRDs.
- `test/chart/rbac_test.go` - verifies chart manager RBAC against temporary controller-gen RBAC output.
- `test/chainsaw` - smoke test installs through Helm and applies the sample custom resource from test fixtures.
- `config/` - removed the old Kustomize and generated manifest tree.

## Open Threads
- Helm chart publication is not wired yet. Releases still publish binaries and the container image; publishing a packaged/OCI chart is a separate follow-up.
- The chart is ready as the template baseline, but future real operator repos may need additional values and templates for webhooks, ServiceMonitors, network policies, or controller sharding.

## References
- PR: https://github.com/meigma/template-k8s/pull/13
- Merge commit: `2d6bd2920b34050f6113e37ccbda4de24e5386b1`
- Local branch: `feat/helm-skill`
