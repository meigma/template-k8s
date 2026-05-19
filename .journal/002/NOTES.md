---
id: 002
title: Chainsaw e2e migration
started: 2026-05-19
---

## 2026-05-19 07:26 — Kickoff
Goal for the session: Attempt to replace the current Go-based e2e tests for the example `NginxDeployment` operator with Kyverno Chainsaw.
Current state of the world: `master` is at `5f9fe08` (`feat: add working nginx deployment operator prototype (#8)`). Session 001 landed the working nginx operator prototype, Moon task wiring, Go envtest coverage, and a Kind-backed `moon run root:test-e2e` smoke path. Journal technical notes identify Moon as the task front door and the current e2e smoke as part of the operator template baseline.
Plan: Inspect the existing e2e and Moon wiring, ground Chainsaw usage against current docs/tooling, prototype the smallest equivalent smoke test, then replace the old e2e path only after the prototype proves it can install the operator, apply the sample CR, wait for readiness, and clean up.
