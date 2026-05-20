---
id: 013
title: Session kickoff
started: 2026-05-20
---

## 2026-05-20 09:31 — Kickoff
Goal for the session: Start a fresh journal session for future `template-k8s` work; no substantive implementation request has been made yet.
Current state of the world: `master` is clean at `22b6f50751bbcd604dcb67955ba108f8f15f46a6` (`feat(chart): add optional Kyverno image verification (#30)`). The personal journal worktree is `journal/jmgilman` at `/Users/josh/code/meigma/template-k8s/.wt/journal-jmgilman`, clean and up to date before this note. Required journal files exist, required skills are `git` and `worktrunk`, and the latest closed summaries read for startup were sessions `009`, `010`, and `011`. `.journal/012` already exists without a `SUMMARY.md`, so this new session was assigned ID `013`.
Plan: Wait for the user's actual work request before changing code or repository state beyond this session kickoff.

## 2026-05-20 09:43 — Tilt proto toolchain
Goal for the checkpoint: Begin the operator development-flow work by making Tilt a repo-managed Proto tool.
What changed: Created implementation branch `feat/tilt-dev-flow` at `/Users/josh/code/meigma/template-k8s/.wt/feat-tilt-dev-flow`, added `tilt = "=0.37.3"` to `.prototools`, and added `.moon/proto/tilt.toml` using Tilt's official GitHub release archive/checksum layout.
Validation: `proto install tilt --quiet` installed the tool successfully, `proto run tilt -- version` reported `v0.37.3, built 2026-04-30`, and `git diff --check` passed. The implementation commit is `53948be` (`chore(proto): add tilt toolchain`).
Next: Continue building out the local operator development flow on the same branch.

## 2026-05-20 09:46 — Ko proto toolchain
Goal for the checkpoint: Add `ko` to the same repo-managed Proto toolchain for the operator development flow.
What changed: Added `ko = "=0.18.1"` to `.prototools` and added `.moon/proto/ko.toml` using the official `ko-build/ko` GitHub release archive/checksum layout.
Validation: `proto install ko --quiet` installed the tool successfully, `proto run ko -- version` reported `0.18.1`, and `git diff --check` passed. The implementation commit is `90cb640` (`chore(proto): add ko toolchain`).
Next: Continue building out the Tilt/ko-backed development flow on `feat/tilt-dev-flow`.

## 2026-05-20 09:52 — Tilt ko kind assessment
Goal for the checkpoint: Research whether Tilt + ko + Kind is the right foundation for this template's local operator development flow.
Findings: The best first prototype is a root `Tiltfile` that renders the existing Helm chart with Tilt's `helm()`/`k8s_yaml()` path, uses `ko` as a Tilt custom image builder for `./cmd`, and runs against a Kind cluster with a discoverable local registry. Avoid making `ko apply` own deployment because it bypasses Tilt's strongest resource graph and image injection behavior, and avoid `kind load` for iterative development because Tilt falls back to that slower path when it cannot find a local registry.
Repo fit: The chart already exposes image repository/tag/digest and pull policy controls, `moon.yml` already has Kind/Chainsaw e2e proof, and the current Dockerfile is not essential for local Go operator loops once ko is introduced. The missing pieces are repo-managed Kind/registry setup, a `Tiltfile`, likely `.tiltignore`, and possibly `.ko.yaml` if the dev image needs to mirror release base/platform choices.
Next: Prototype a small Tiltfile plus Kind local-registry setup, then prove that Go source changes rebuild with ko and roll the manager Deployment in Kind.

## 2026-05-20 10:20 — Local development stack implemented
Goal for the checkpoint: Implement the approved `ctlptl` + `kind` + `tilt` + `ko` development stack plan on `feat/tilt-dev-flow`.
What changed: Added Proto-managed `kind` and `ctlptl`, a `dev/ctlptl.yaml` Kind+registry config, `.ko.yaml`, root `Tiltfile`, `.tiltignore`, `dev/ko-build.sh`, `dev/stack-smoke.sh`, Moon `dev-*` tasks, and downstream README/DELETE_ME guidance. The implementation commit is `90b5895` (`feat(dev): add local operator development stack`).
Validation: `proto install --quiet`, `proto run kind -- version`, `proto run ctlptl -- version`, `proto run tilt -- version`, `proto run ko -- version`, `moon run root:chart-validate`, `moon run root:generate`, `bash -n dev/ko-build.sh dev/stack-smoke.sh`, `git diff --check`, `moon run root:dev-stack-smoke`, and a no-op `moon run root:dev-down` passed. The full smoke created the dev Kind cluster and registry, `tilt ci` built the controller image through ko, the manager Deployment became healthy, the sample `NginxDeployment` and owned Deployment became Available, and the port-forwarded Service returned `hello from template-k8s`.
Notes: The first smoke run exposed a localhost binding mismatch in the HTTP proof; `dev/stack-smoke.sh` now uses `kubectl port-forward --address localhost` and curls `http://localhost:<port>/`.
Next: Push/open review or continue refining the development flow as requested.

## 2026-05-20 10:52 — Close
Merged PR #32 (`feat(dev): add local operator development stack`) via squash as `e38b98899bdba357ad02e17b754bc2d326b78dd6` after user review. Local `master` is fast-forwarded to the merge commit, GitHub CI and Kusari Inspector passed on the final PR commit, and the `feat/tilt-dev-flow` worktree was removed.

Closeout artifacts written: `.journal/013/SUMMARY.md`, an `INDEX.md` row for session 013, and a compact `TECH_NOTES.md` entry covering the Proto-managed `ctlptl`/Kind/Tilt/ko development stack.
