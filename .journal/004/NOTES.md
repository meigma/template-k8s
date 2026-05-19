---
id: 004
title: Operator best-practice research
started: 2026-05-19
---

## 2026-05-19 10:38 — Kickoff
Goal for the session: research modern open source Kubernetes operators, especially operators built around Kubebuilder/controller-runtime, and collect interesting design choices or deviations that could inform this template.
Current state of the world: `master` is at `49f7395` after the release dry-run gating change. The template has a working `NginxDeployment` prototype, Moon-fronted generation/test/lint/e2e tasks, Chainsaw smoke coverage, and repo-local Kubernetes operator guidance in `.agents/skills/k8s-operator/SKILL.md`.
Plan: start with current open source operator codebases and primary docs, prefer evidence from active Kubebuilder/controller-runtime operators, then summarize candidate practices that are small enough to prototype before becoming template policy.
