---
id: 002
title: Chainsaw e2e migration
started: 2026-05-19
---

## 2026-05-19 07:26 — Kickoff
Goal for the session: Attempt to replace the current Go-based e2e tests for the example `NginxDeployment` operator with Kyverno Chainsaw.
Current state of the world: `master` is at `5f9fe08` (`feat: add working nginx deployment operator prototype (#8)`). Session 001 landed the working nginx operator prototype, Moon task wiring, Go envtest coverage, and a Kind-backed `moon run root:test-e2e` smoke path. Journal technical notes identify Moon as the task front door and the current e2e smoke as part of the operator template baseline.
Plan: Inspect the existing e2e and Moon wiring, ground Chainsaw usage against current docs/tooling, prototype the smallest equivalent smoke test, then replace the old e2e path only after the prototype proves it can install the operator, apply the sample CR, wait for readiness, and clean up.

## 2026-05-19 07:43 — Chainsaw research proposal
Targeted research found Chainsaw v0.2.15 as the current release, with explicit `chainsaw-test.yaml` resources, declarative apply/assert/wait operations, ephemeral test namespaces with templates, automatic cleanup, script/command escape hatches, JUnit/JSON reporting, and catch/cleanup diagnostics. The local Go e2e suite currently mixes Kind/image orchestration, manager deployment, metrics probing, and the sample `NginxDeployment` smoke. Recommended direction: keep Moon as the e2e front door and Kind/image orchestrator, add Chainsaw as a proto-managed tool, replace the Go e2e suite with one sequential Chainsaw smoke that deploys the manager, asserts readiness/metrics service, applies the sample into a restricted Chainsaw namespace, waits for parent and child readiness, and collects logs/events/describes on failure.

## 2026-05-19 08:05 — Chainsaw implementation
Implemented the migration on `feat/chainsaw-e2e` at `24bd714` (`test: replace e2e suite with chainsaw`). The branch pins Chainsaw v0.2.15 through Proto, rewires `moon run root:test-e2e` to create/reuse Kind, build/load the manager image, and run `chainsaw test`, replaces `test/e2e` plus `test/utils` with `test/chainsaw`, and refreshes README plus the repo-local operator skill. Validation passed: `proto run chainsaw -- version`, `moon run root:test-e2e`, `KUBEBUILDER_ASSETS="$(proto run setup-envtest -- use 1.35.x -p path)" go test ./...`, `moon ci --summary minimal`, and `git diff --check` / `git diff --cached --check`.

## 2026-05-19 08:26 — Review response
Responded to review feedback on `feat/chainsaw-e2e` with `5035af5` (`test: harden chainsaw e2e coverage`). The follow-up exports a Kind-cluster-specific temp kubeconfig before invoking Chainsaw so nested scripts run against the intended cluster, restores the authenticated metrics endpoint smoke with a restricted curl pod plus `metrics-reader` binding, and adds a `root:chainsaw-lint` Moon task that `moon ci` now runs. Validation passed: `proto run chainsaw -- version`, direct Chainsaw config/test lint, `moon run root:chainsaw-lint --summary minimal`, `moon run root:test-e2e`, `KUBEBUILDER_ASSETS="$(proto run setup-envtest -- use 1.35.x -p path)" go test ./...`, `moon ci --summary minimal`, and `git diff --check` / `git diff --cached --check`.

## 2026-05-19 08:52 — Envtest Moon task
Responded to the final review note with `d514b86` (`test: add moon envtest task`). The branch now has `moon run root:test`, which wraps `KUBEBUILDER_ASSETS="$(setup-envtest use 1.35.x -p path)" go test ./...`, participates in `moon ci`, and replaces stale plain `go test ./...` guidance in README and the repo-local operator skill. Validation passed: `moon task root:test`, `moon run root:test --summary minimal`, `moon ci --summary minimal`, `moon run root:test-e2e --summary minimal`, and `git diff --check` / `git diff --cached --check`.
