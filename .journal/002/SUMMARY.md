---
id: 002
title: Chainsaw e2e migration
date: 2026-05-19
status: complete
repos_touched: [template-k8s]
related_sessions: [001]
---

## Goal
Replace the current Ginkgo-based Go e2e suite for the example operator with a Kyverno Chainsaw smoke test while keeping Moon as the developer-facing task front door.

## Outcome
The goal was met. PR #9 landed the Chainsaw migration on `master`, kept `moon run root:test-e2e` as the local e2e entrypoint, added Chainsaw schema linting to `moon ci`, and added a `root:test` envtest wrapper so agents do not rely on a broken plain `go test ./...` path.

## Key Decisions
- Keep Moon responsible for Kind, image build/load, and task orchestration -> preserves the repository's template workflow and avoids making Chainsaw own local cluster lifecycle.
- Pin Chainsaw through Proto -> keeps the test runner reproducible alongside the other repo-managed CLIs.
- Export a temp kubeconfig for the selected Kind cluster -> prevents Chainsaw and nested `kubectl` scripts from accidentally targeting the ambient kube context.
- Restore the authenticated metrics probe in Chainsaw -> keeps the template smoke useful for metrics TLS/auth/RBAC regressions instead of only checking Service existence.
- Add `root:test` around `setup-envtest` -> makes the correct envtest binary discovery path the normal Moon task and lets `moon ci` run it.

## Changes
- `.moon/proto/chainsaw.toml` and `.prototools` - added the pinned Chainsaw v0.2.15 tool.
- `moon.yml` - replaced the Go e2e task internals with Kind plus Chainsaw orchestration, added `root:chainsaw-lint`, and added `root:test` for envtest.
- `test/chainsaw/chainsaw-config.yaml` - added the Chainsaw config with namespace templating and Restricted Pod Security labels.
- `test/chainsaw/nginx-smoke/chainsaw-test.yaml` - added the sequential smoke test, controller deploy, manager readiness wait, authenticated metrics curl pod, sample reconciliation waits, cleanup, and failure diagnostics.
- `test/e2e/**` and `test/utils/utils.go` - removed the obsolete Go e2e harness.
- `README.md` - refreshed task docs to describe Chainsaw e2e, Chainsaw lint, and the Moon envtest task.
- `.agents/skills/k8s-operator/SKILL.md` - updated repo-local operator guidance for Chainsaw e2e and `root:test`.

## Open Threads
- The Kind-backed Chainsaw smoke remains local/manual with `runInCI: false`, matching the previous e2e gate. CI parses Chainsaw YAML and runs envtest but does not create Kind.
- The metrics probe is now covered, but deeper reconciler correctness remains in envtest rather than expanding the Chainsaw smoke.

## References
- PR: https://github.com/meigma/template-k8s/pull/9
- Merge commit: `67a6402` (`test: replace e2e suite with chainsaw (#9)`)
- Prior session: `.journal/001/SUMMARY.md`
- Kyverno Chainsaw docs: https://kyverno.github.io/chainsaw/

## Lessons
- In this template, plain `go test ./...` is not a reliable verification command unless envtest assets are already discoverable. Prefer `moon run root:test`.
- For local e2e tests that create or reuse named Kind clusters, always run test tooling through a kubeconfig exported for the selected cluster instead of trusting the ambient context.
