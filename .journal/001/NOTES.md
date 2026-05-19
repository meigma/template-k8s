---
id: 001
title: Working example operator prototype
started: 2026-05-18
---

## 2026-05-18 17:39 — Kickoff
Goal for the session: turn the repository's example operator into something that actually works, keeping the effort intentionally simple so it can validate the real template workflows.
Current state of the world: this is the first official session. The personal journal branch `journal/jmgilman` exists and contains the root journal scaffold. The default branch is `master` at `4d91d99` (`feat: bootstrap Kubernetes operator template`). No substantive operator work has been started in this session.
Plan: after session priming, inspect the existing example operator, choose the smallest useful working prototype, implement it in an isolated Worktrunk branch, and verify it with the repo's normal checks.

## 2026-05-18 17:43 — Started nginx API scaffold
Created implementation worktree `feat/nginx-api` at `.wt/feat-nginx-api`. The requested first slice is to remove the generated Widget API/controller and use Kubebuilder to create a new base API object and controller for the nginx prototype, without adding reconcile logic yet.

## 2026-05-18 17:58 — Nginx API scaffold committed
Replaced the Widget scaffold with a Kubebuilder-generated `NginxDeployment` API and controller on branch `feat/nginx-api` (`625ddb2`). The API currently carries simple prototype fields for `image`, `replicas`, `port`, `config`, and status `readyReplicas`/`conditions`; the reconciler remains generated/no-op as requested. Regenerated deepcopy code, CRD, RBAC, and sample manifests. Verification passed: `moon run root:build`, `go test ./...` after installing ignored envtest assets under `bin/k8s/1.35.0-darwin-arm64`, `moon run root:lint`, and `git diff --cached --check`.

## 2026-05-18 19:19 — Controller implementation committed
Implemented the first real `NginxDeployment` reconciliation loop on `feat/nginx-api` (`1c35e43`). The controller now owns a same-named ConfigMap, Deployment, and ClusterIP Service; mounts `nginx.conf`; rolls pods with a config hash annotation; watches owned children; and projects Deployment readiness into `status.readyReplicas` plus an `Available` condition. Regenerated manager RBAC for ConfigMaps, Services, and Deployments. Verification passed: `moon run root:manifests`, `moon run root:build`, `go test ./...`, `moon run root:lint`, and `git diff --check`.

## 2026-05-18 19:50 — Review feedback addressed
Addressed review feedback on `feat/nginx-api` (`ed0b774`). `spec.replicas` is now a pointer so typed clients can preserve explicit scale-to-zero; Deployment status projection now gates `Available=True` on the owned Deployment having observed its current generation and reporting an available condition; and Deployment reconciliation now sets only the fields this controller owns instead of replacing the whole `DeploymentSpec`. Replaced e2e `make` calls with Moon tasks, added a Kind-backed sample `NginxDeployment` smoke, fixed local-image e2e deployment with `LOCAL_IMAGE=true`, and refreshed the README prototype state. Verification passed: `go test ./...`, `moon run root:lint`, `moon ci --summary minimal`, `go test -tags=e2e ./test/e2e -run '^$'`, full `moon run root:test-e2e`, and `git diff --check`.

## 2026-05-18 19:57 — K8s operator skill added
Added a repo-local `k8s-operator` skill under `.agents/skills/k8s-operator/` on `feat/nginx-api` (`092005f`). The skill captures the lessons from the controller review pass: pointer fields for zero-capable defaulted API values, child observed-generation gating before reporting parent availability, narrower owned-field reconciliation, envtest patterns for stale Deployment status, Moon as the task front door, and Kind local-image e2e handling.

## 2026-05-18 20:12 — Prototype hardening review addressed
Addressed the second review pass on `feat/nginx-api` (`980eed7`). The demo now defaults to `nginxinc/nginx-unprivileged:stable` on port 8080, reconciles Restricted-compatible pod/container security settings plus resource requests, and the sample config includes writable PID/temp paths for unprivileged nginx. Tightened controller RBAC markers to read/watch the primary CR, patch status, and create/patch owned children only. Added CRD validation plus envtest coverage for the <=63 character name constraint used by same-named children and selector labels. The Kind e2e sample now runs in its own Restricted-enforced namespace, and the metrics `ClusterRoleBinding` is cleared before creation and deleted during teardown. Verification passed: `go test ./...`, `moon run root:build`, `moon run root:lint`, `moon ci --summary minimal`, `go test -tags=e2e ./test/e2e -run '^$'`, full `moon run root:test-e2e`, and `git diff --check`.

## 2026-05-18 22:14 — API validation review addressed
Addressed the follow-up API validation review on `feat/nginx-api` (`e16278d`). Because the prototype still names child resources after the `NginxDeployment`, the CRD now rejects names that are longer than 63 characters or are not Service-safe DNS labels, including otherwise-valid CR names with dots. Inline `spec.config` is capped at 64 KiB before it is copied into an owned ConfigMap. Added envtest coverage for dotted/too-long names and oversized config, refreshed the README, and updated the repo-local `k8s-operator` skill with these lessons. Verification passed: `moon run root:generate`, `moon run root:manifests`, `go test ./...`, `moon run root:build`, `go test -tags=e2e ./test/e2e -run '^$'`, `kubectl kustomize config/default`, `moon run root:lint`, `moon ci --summary minimal`, `git diff --check`, and full `moon run root:test-e2e`.

## 2026-05-19 07:19 — Close
Closed the session after PR #8 was approved and squash-merged: https://github.com/meigma/template-k8s/pull/8. Local `master` was fast-forwarded to merge commit `5f9fe08`, the remote `feat/nginx-api` branch was deleted, and the `feat/nginx-api` worktree/branch was removed with Worktrunk. Closeout artifacts were written in this journal branch: `.journal/001/SUMMARY.md`, `.journal/INDEX.md`, and `.journal/TECH_NOTES.md`.
