---
id: 007
title: Test boundaries and operator observability
date: 2026-05-19
status: complete
repos_touched: [template-k8s]
related_sessions: [001, 002, 004, 005, 006]
---

## Goal
Evaluate whether envtest still earns its place beside Chainsaw e2e tests, encode a durable boundary between those layers, and add template-quality operator observability examples for metrics, events, and reconcile logging.

## Outcome
The goal was met. PR #15 established the envtest/Chainsaw split in repo guidance and added manager-backed envtest coverage for parent and owned-child watch wiring. PR #16 added injectable operator-specific telemetry, Kubernetes Events, signal-oriented reconcile logging, focused envtest coverage, and Chainsaw smoke assertions for the new external signals.

## Key Decisions
- Keep envtest and Chainsaw -> envtest is the fast controller/API behavior matrix, while Chainsaw remains a small installed-operator smoke path.
- Add manager-backed envtest coverage now -> this template should demonstrate watch wiring before later field indexes, predicates, or external reference watches make the gap harder to notice.
- Keep metrics finite-label and controller-specific -> template metrics should demonstrate useful instrumentation without encouraging namespace/name labels or per-object gauges.
- Emit Kubernetes Events through a small recorder abstraction -> the controller can demonstrate user-visible state-change events without scattering direct event-recorder calls through reconciliation logic.
- Log only durable signal at info level -> child apply side effects and persisted condition transitions are useful default logs, while reconcile lifecycle/no-op details belong behind `V(1)`.
- Avoid OpenMetrics/OpenTelemetry changes -> controller-runtime already exposes Prometheus on `/metrics`, and changing internal metric binding was not worth the complexity for this template.

## Changes
- `AGENTS.md` - clarified the envtest versus Chainsaw testing boundary for future agents.
- `.agents/skills/k8s-operator/SKILL.md` - added local rules for test layering, observability labels/events, and reconcile logging verbosity.
- `internal/controller/nginxdeployment_controller.go` - added injectable telemetry, child apply operation reporting, status transition instrumentation, Kubernetes Events, defaulted child fields to avoid no-op churn, and controller-runtime context logging.
- `internal/controller/telemetry/` - added reusable metrics and recorder primitives with a no-op fallback.
- `internal/controller/nginxdeployment_controller_test.go` - added manager-backed envtest coverage plus metrics/events/no-op behavior tests.
- `cmd/setup.go` - registered operator metrics with controller-runtime's metrics registry and wired the Kubernetes Event recorder.
- `charts/template-k8s/templates/rbac-manager.yaml` - added event create/patch RBAC for the controller while keeping leader-election event RBAC separate.
- `test/chainsaw/nginx-smoke/chainsaw-test.yaml` - added smoke assertions for the new metric families and emitted events.
- `go.mod` - promoted Prometheus client usage to a direct dependency.

## Open Threads
- None for this session. The release dry-run checks remain intentionally skipped for ordinary PRs by the existing release-branch gating policy.

## References
- PR #15: https://github.com/meigma/template-k8s/pull/15
- PR #16: https://github.com/meigma/template-k8s/pull/16
- Merge commit #15: `750470cb4eae9b4035c4428cdc2d5d2c95f45523`
- Merge commit #16: `93278412e61a5bb597f66875e60827d9178d2e0b`
- `.journal/001/SUMMARY.md`
- `.journal/002/SUMMARY.md`
- `.journal/004/SUMMARY.md`
- `.journal/005/SUMMARY.md`
- `.journal/006/SUMMARY.md`
