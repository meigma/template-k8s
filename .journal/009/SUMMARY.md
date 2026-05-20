---
id: 009
title: Downstream operator template setup guidance
date: 2026-05-19
status: complete
repos_touched: [template-k8s]
related_sessions: [008]
---

## Goal
Make the repository more usable as a template for downstream Kubernetes
operators by adding the same kind of first-repository setup guidance and GitHub
repository administration surface that `template-go` provides.

## Outcome
The goal was met. PR #24 was squash-merged, local `master` was fast-forwarded,
and the `feat/github-repo-settings` worktree and remote branch were removed.
The root README is now a downstream-facing operator README template, the root
`DELETE_ME.md` is the generated-repository cleanup checklist, and the reusable
repository settings helper/config from `template-go` is now present under
`.github/`.

## Key Decisions
- Keep the README generic -> this repository is itself the template, so copied
  repositories need a README they can fill in rather than docs about
  `template-k8s` internals.
- Make `DELETE_ME.md` operator-specific -> the template-go file was useful as a
  shape reference, but this repo needs a checklist for Kubebuilder metadata,
  API types, reconcilers, Helm chart paths, release workflows, Chainsaw, and
  repo-local agent guidance.
- Copy repository settings tooling unchanged first -> the initial slice keeps
  `.github/scripts/configure_github_repo.py` and
  `.github/repository-settings.toml` byte-for-byte aligned with `template-go`
  before any operator-specific customization.

## Changes
- `README.md` - replaced the repo-specific starter text with a downstream
  README template using placeholders for operator name, API group/kind,
  repository, namespace, installation, usage, configuration, development,
  release, contribution, security, and license details.
- `DELETE_ME.md` - added a generated-repository setup checklist covering project
  identity, API/reconciler replacement, Helm chart/manifests, release and
  repository automation, tests/samples, agent guidance, artifact-shape choices,
  final validation, and cleanup.
- `.github/scripts/configure_github_repo.py` and
  `.github/repository-settings.toml` - copied the reusable GitHub repository
  settings automation from `template-go`.

## Open Threads
- The repository settings file was added but not applied to GitHub during this
  session. A future session can customize or apply it if the live repository
  settings should be reconciled from this file.
- `DELETE_ME.md` calls out that `Helm Chart Dry Run` may need to become a
  required protected check if chart publication should be enforced for release
  PRs; the copied settings currently retain the template-go required contexts.

## References
- PR #24: https://github.com/meigma/template-k8s/pull/24
- Merge commit: `8b9bd5fbf17c19b3adc3089eaa0679ba0dea3215`
- Local branch (removed): `feat/github-repo-settings`
- Source files copied from `/Users/josh/code/meigma/template-go/.github/`
- `.journal/008/SUMMARY.md` - prior Helm OCI chart release and published chart
  workflow baseline.
