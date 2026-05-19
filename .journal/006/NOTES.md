---
id: 006
title: Evaluate removing Kustomize
started: 2026-05-19
---

## 2026-05-19 13:23 — Kickoff
Goal for the session: evaluate practical options for removing Kustomize from the Kubernetes operator template.
Current state of the world: the repo is on `master` after PR #12, with Moon as the workflow front door, repo-local session and Kubernetes operator skills restored, generated Kubernetes assets still part of the source tree, and recent guidance favoring pragmatic operator patterns over generated-tooling ceremony.
Plan: first map where Kustomize is currently used, then compare lightweight replacement options against Kubebuilder/controller-runtime expectations, and prototype only enough to prove the most plausible path before making a recommendation.
