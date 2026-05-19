---
id: 006
title: Evaluate removing Kustomize
started: 2026-05-19
---

## 2026-05-19 13:23 — Kickoff
Goal for the session: evaluate practical options for removing Kustomize from the Kubernetes operator template.
Current state of the world: the repo is on `master` after PR #12, with Moon as the workflow front door, repo-local session and Kubernetes operator skills restored, generated Kubernetes assets still part of the source tree, and recent guidance favoring pragmatic operator patterns over generated-tooling ceremony.
Plan: first map where Kustomize is currently used, then compare lightweight replacement options against Kubebuilder/controller-runtime expectations, and prototype only enough to prove the most plausible path before making a recommendation.

## 2026-05-19 14:29 — Helm skill
Created implementation branch/worktree `feat/helm-skill` at `.wt/feat-helm-skill`.
Compiled Helm v4-focused chart guidance from current official Helm documentation, including the Helm 4 overview, chart best-practices pages, chart/CRD topic docs, OCI registry guidance, chart tests, and provenance notes.
Added `.agents/skills/helm/SKILL.md` as a concise repo-local skill for modern Helm v4 chart work, especially replacing Kustomize with a centralized chart in operator repos.
Validation: `uv run --with pyyaml python /Users/josh/.codex/skills/.system/skill-creator/scripts/quick_validate.py .agents/skills/helm` passed; `git diff --cached --check` passed before commit.
Implementation commit: `7c4c95d` (`docs(skills): add helm chart guidance`).
