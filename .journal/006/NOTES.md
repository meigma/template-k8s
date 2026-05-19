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

## 2026-05-19 14:45 — Initial Helm chart
Added `charts/template-k8s` as the first centralized Helm chart for the operator. The chart includes conventional `Chart.yaml`, `.helmignore`, `values.yaml`, `values.schema.json`, plain generated CRD YAML under `crds/`, and readable templates for ServiceAccount, manager RBAC, leader-election RBAC, metrics-reader RBAC, Deployment, and metrics Service.
The CRD was generated directly from Go API markers with `controller-gen crd paths="./..." output:crd:artifacts:config=charts/template-k8s/crds`; no files were copied or read from `config/` for the chart.
Defaults render `ghcr.io/meigma/template-k8s:v0.1.0`, secure metrics on `:8443`, leader election enabled, restricted-compatible pod/container security settings, and digest-pinned image support via `image.digest`.
Validation passed: `python3 -m json.tool charts/template-k8s/values.schema.json`, `helm lint charts/template-k8s`, `helm template template-k8s charts/template-k8s --namespace template-k8s-system --include-crds | kubeconform -strict -ignore-missing-schemas`, `helm install template-k8s charts/template-k8s --namespace template-k8s-system --dry-run=client --server-side=false`, and `helm package charts/template-k8s --destination /tmp`.
Implementation commit: `9ee217d` (`feat(chart): add operator helm chart`).

## 2026-05-19 15:14 — Review fixes
Addressed review findings against the initial chart in `ceea249` (`fix(chart): address helm review findings`).
Changes: `root:manifests` now also generates chart CRDs; `root:chart-validate` validates the chart; `root:generated-check` verifies generated API artifacts are current; CI explicitly runs both new checks and `git diff --check`; `.prototools` now pins Helm `4.0.4`.
Release Please now updates `charts/template-k8s/Chart.yaml` through a generic extra-file updater, with inline markers for both chart `version` and `appVersion`; the default image tag continues to follow `appVersion`.
Chart fixes: resource-specific name helpers reserve suffix length and fail early for too-long explicit overrides; manager bind args now derive from the public port values; metrics auth RBAC is separate from manager RBAC; leader-election RBAC includes ConfigMaps, Leases, and Events; admin/editor/viewer helper ClusterRoles are templated behind `rbac.customResourceRoles.create`.
Validation passed: `moon run root:generated-check`, `moon run root:chart-validate`, `moon run root:test`, `moon ci --summary minimal`, `git diff --check`, Helm schema rejection for old `metrics.bindAddress` and `manager.healthProbe.bindAddress`, long-release render name-length check, port override render check, custom-resource helper role disable check, and Helm skill validation via `quick_validate.py`.
