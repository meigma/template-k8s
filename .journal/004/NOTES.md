---
id: 004
title: Operator best-practice research
started: 2026-05-19
---

## 2026-05-19 10:38 — Kickoff
Goal for the session: research modern open source Kubernetes operators, especially operators built around Kubebuilder/controller-runtime, and collect interesting design choices or deviations that could inform this template.
Current state of the world: `master` is at `49f7395` after the release dry-run gating change. The template has a working `NginxDeployment` prototype, Moon-fronted generation/test/lint/e2e tasks, Chainsaw smoke coverage, and repo-local Kubernetes operator guidance in `.agents/skills/k8s-operator/SKILL.md`.
Plan: start with current open source operator codebases and primary docs, prefer evidence from active Kubebuilder/controller-runtime operators, then summarize candidate practices that are small enough to prototype before becoming template policy.

## 2026-05-19 10:48 — Source survey
Surveyed current upstream code from Cluster API, External Secrets Operator, cert-manager, Kueue, Flux source-controller, and OpenTelemetry Operator. Patterns worth considering for the template are field indexes for reverse lookups, patch helpers with owned conditions, status-only update predicates, partial-metadata watches, typed manager configuration, class/shard selection, leader-aware queue draining, artifact/content-observation status, webhook-backed invariants, and generated-desired-object pruning.

## 2026-05-19 11:18 — Close
Merged PR #11 (`refactor(cli): use kong and slog for manager startup`) after user approval. The landed work replaced generated flag/zap startup with Kong plus `slog`, split manager startup into ordered helpers, updated Moon build/run tasks to target `./cmd`, and added focused helper tests. Local `master` is fast-forwarded to `ca321f9`, the `feat/kong-slog-cli` worktree was removed, and the remote feature branch was deleted. The earlier operator-practice source survey remains useful context but was not converted into template policy in this session.
