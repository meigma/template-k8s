---
id: 015
title: Reproduce template-go-api session 015 tooling — mise + melange/apko + SLSA L3, proven by a release rehearsal
date: 2026-06-28
status: complete
repos_touched: [template-k8s]
related_sessions: ["008", "011", "013", "014"]
---

## Goal
Reproduce in `template-k8s` the tooling migration from the sibling
`template-go-api` session 015 — **so far as it concerns what is present here**:
adopt **mise** (replacing Proto), build the operator image with **Chainguard
melange/apko** (replacing the Dockerfile), and isolate release provenance in a
reusable workflow for **SLSA Build L3**. Investigate first, deliver an assessment,
then implement, and (developer choice) prove it with a real release rehearsal.

## Outcome
**Met in full and proven end-to-end.** Shipped as four squash-merged PRs and
validated with a release rehearsal that passed **clean on the first real tag**
(the sibling `template-go-api` needed four versions; `template-mcp` needed two):
- **#42** `f708c83` — Proto → mise; moon on the `system` toolchain.
- **#43** `800bbfb` — Dockerfile → melange/apko (+ keyless cosign, syft SBOM).
- **#45** `d26d07f` — provenance → reusable `attest.yml` (SLSA L3) + Kyverno signer flip.
- **#46** `a9cc2dc` — port mise/melange/apko agent skills + docs.
- **Rehearsal:** `v0.1.3` — `release.yml` ran all 10 jobs green first try; the image
  `ghcr.io/meigma/template-k8s:v0.1.3` (linux/amd64+arm64) is cosign-verified and
  carries SLSA-provenance attestations whose signer is `attest.yml@refs/tags/v0.1.3`
  (verified — confirms L3 isolation); the chart is L3-attested too. GitHub release
  left a **draft** awaiting a human publish.

## Key Decisions
- **Four-PR structure mirroring the sibling** (mise → melange/apko → reusable
  attest.yml → skills/docs), each provable independently via `moon run root:check`
  + a dispatched `release-dry-run`. PR2 keeps the SLSA provenance attestation
  **in-job** (signer release.yml) and adds cosign + syft SBOM; PR3 extracts
  provenance into the reusable workflow and flips the Kyverno signer.
- **Faithful reproduction** (developer choice): added keyless cosign + syft SBOM
  (template-k8s had neither) and routed the **Helm chart** provenance through
  `attest.yml` too (Option A — the chart is a third artifact class the sibling
  lacks).
- **controller-gen on the `go:` backend, not aqua/ubi** — no aqua package exists;
  ubi is deprecated in mise and mis-names the binary `controller-tools`. The Go
  backend yields a correct `controller-gen` (integrity via the Go checksum DB).
- **`ko` stays the local Tilt dev builder; melange/apko are release-only.** The
  Kind e2e (`scripts/test-e2e.sh`) was repointed off the deleted Dockerfile to the
  apko image via the new `dev/image-build.sh`.
- **Kyverno values flipped release.yml → attest.yml in lockstep with PR3** —
  `attest-build-provenance` preserves the exact `slsa.dev/provenance/v1` predicate +
  `actions/buildtypes/workflow/v1` buildType the policy checks (verified against the
  live v0.1.2 attestation before flipping), so only the signer identity changes.
- **Rehearsed via release-please, not a bare tag** — `release.yml`'s `resolve-release`
  waits for a release-please draft, so a throwaway tag cannot trigger it. A stale
  pending release PR (#31, predating PR2–PR4, missing the melange/apko bumps) was
  regenerated into a clean PR #47 (close + delete branch + re-run release-please)
  rather than merged as-is.

## Changes
- `mise.toml` + `mise.lock` (new): 16 tools, fail-closed `locked=true`; moon's
  `macos-x64` lock entry hand-added (mise write quirk). Deleted `.prototools`,
  `.moon/proto/*`, `.envrc`, `.go-version`. `moon.yml`/`.moon/toolchains.yml` →
  `system`. `ci.yml` + the Helm jobs in `release.yml`/`release-dry-run.yml` →
  `jdx/mise-action`; `setup-go` → `go.mod`.
- `melange.yaml` + `apko.yaml` (new); deleted `Dockerfile`/`.dockerignore`.
  `release.yml`/`release-dry-run.yml`/`security-scan.yml` → melange/apko; the
  multi-arch index digest is resolved authoritatively via `imagetools inspect`.
  `dev/image-build.sh` (new) + `scripts/test-e2e.sh` build the apko image for e2e.
  `release-please-config.json` `extra-files` += [melange.yaml, apko.yaml].
- `.github/workflows/attest.yml` (new reusable, SLSA L3); `release.yml` gained
  `attest-binaries`/`attest-image`/`attest-chart` callers and dropped the inline
  provenance steps (cosign + syft SBOM stay inline). `charts/template-k8s/values.yaml`
  Kyverno `subjectRegExp` → `attest.yml`; `test/chart/rbac_test.go` updated to match.
- `.agents/skills/{mise,melange,apko}/` (new). `AGENTS.md`/`CLAUDE.md`/`DELETE_ME.md`
  refreshed (skills refs + downstream melange/apko guidance; dropped stale
  Dockerfile-era notes).

## Open Threads
- **`v0.1.3` GitHub release is a DRAFT awaiting a human publish** — the image + chart
  are already live on GHCR (no draft state there). Publish or reject after inspection.
- Optional future convergence: SLSA L3 here is GitHub-self-asserted (reusable-workflow
  isolation), not a `slsa-verifier` builder ID — the deliberate trade for keeping
  `gh attestation verify`.
- Stale dependabot PRs (#37/#38/#40/#44) are unrelated, left for normal triage.

## Lessons
- **The first real tag passed clean because the sibling's three tag-only-path bugs
  were pre-applied:** (1) every `attest.yml` caller (incl. the binary one) must grant
  `packages: write` to match the shared job; (2) `apko publish --sbom-path <dir>`
  needs the dir to pre-exist (`mkdir -p`); (3) `attest.yml` needs its OWN
  `docker/login-action`. The dry-run reaches none of these — only a real tag does.
- **Verify supply-chain claims against the published artifact, not action docs from
  memory.** A reviewer flagged `actions/attest@v4.1.0` as not producing SLSA
  provenance; `gh attestation verify` on the live v0.1.2 image disproved it
  (predicate `slsa.dev/provenance/v1` + buildType `workflow/v1`).
- **`root:check` runs only the RBAC-drift chart test**; the Kyverno render assertion
  is in the full `root:test`. CI caught the coupled `subjectRegExp` test that the
  signer flip broke — validate chart `values.yaml`/template changes with `root:test`.
- **release-please can leave a stale release PR** when no version-changing commit
  lands after it (build/ci/docs don't bump); regenerate it (close + delete branch +
  re-run) so `extra-files` like melange/apko get bumped. Check for pre-existing tags
  first (v0.1.3 was free here; it collided in `template-mcp`).
- Local validation needs the host's stray go neutralized (`~/.proto/bin/go`,
  `~/.goenv/shims/go`, `GOPATH=~/go/1.26.4` vs mise's exported `GOROOT=…/1.26.3`):
  run gates via `mise exec -- bash -c 'strip /.proto/shims from PATH; unset GOROOT; moon run …'`.

## References
- PRs: #42 (mise), #43 (melange/apko), #45 (SLSA L3), #46 (skills/docs);
  release-please #47 (`chore(master): release 0.1.3`).
- Released (draft): `v0.1.3` — `ghcr.io/meigma/template-k8s:v0.1.3`
  (index `sha256:d713787b782ab3e1250b80eb10886c638d9674cc1c564b5d35ee6f949a7d717d`);
  chart `…/chart@sha256:fe92439562af8dac235f45128b7e25d4e0eb989f1278a5ae0d6e9389a7168fad`.
  `release.yml` run `28338367187` = full success.
- Source: `template-go-api/.wt/journal-jmgilman/.journal/015/` (SUMMARY + NOTES);
  parallel reproduction `template-mcp/.wt/journal-jmgilman/.journal/001/`.
- Builds on: `.journal/008` (OCI release), `.journal/011` (Kyverno), `.journal/013`
  (dev stack), `.journal/014` (moon slimming). Session log: `.journal/015/NOTES.md`.
