---
id: 015
title: Reproduce go-template-api Mise/Chainguard/release tooling changes
started: 2026-06-28
---

## 2026-06-28 13:19 — Kickoff
Goal for the session: Review the full scope of session 015 in the sibling repo
`~/code/meigma/go-template-api` (which was originally sourced from this template
family), understand the major tooling changes adopted there — namely Mise
adoption, Chainguard tooling, and release-flow updates — and fully reproduce
those changes in this `template-k8s` repo to the extent they concern what is
present here.

Current state of the world:
- `template-k8s` is a Kubebuilder/controller-runtime Go operator template.
- Tooling today: Moon is the task front door (`root:check`, `root:test`,
  `root:generate`, `root:deploy`, `root:undeploy`, `root:dev-up`,
  `root:dev-down`, `root:test-e2e`); repo-local tools come via Proto +
  `.prototools` / `.moon/proto/*.toml`, surfaced on PATH through `.envrc`
  (direnv) locally and `moonrepo/setup-toolchain` in CI.
- Local dev stack (session 013) is Proto-managed ctlptl/Kind/Tilt/ko.
- Release flow (sessions 003/008/011/012): Release Please drives versioning;
  the tag release workflow publishes binaries, a multi-platform container
  image, and the Helm chart as `oci://ghcr.io/meigma/template-k8s/chart`;
  Release Dry Run gates binary/container dry-run jobs to release PRs; optional
  Kyverno image verification trusts `.github/workflows/release.yml` keyless
  identity at SemVer tag refs.
- So the relevant deltas to investigate: Mise (likely replacing/augmenting
  Proto for tool management), Chainguard tooling (e.g. apko/melange/chainctl,
  or Chainguard base images / wolfi for the container image), and release-flow
  updates.

Plan (initial investigation, no code changes yet):
1. Read go-template-api session 015 journal (SUMMARY.md + NOTES.md + INDEX) to
   capture the full scope and intent of the changes there.
2. Diff the relevant files between go-template-api and template-k8s (tool
   config, CI/release workflows, container build, dev stack) to see what
   actually changed and what maps to this repo.
3. Separate "applies here" from "go-template-api specific" given this repo is a
   k8s operator with a Helm chart + container image + dev stack.
4. Produce an assessment for the user before doing the reproduction work.
