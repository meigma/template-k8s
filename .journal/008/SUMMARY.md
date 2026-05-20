---
id: 008
title: Helm OCI chart release and first published release
date: 2026-05-19
status: complete
repos_touched: [template-k8s]
related_sessions: [006, 007]
---

## Goal
Make the Helm chart a first-class release artifact for the operator template, using the same release version as the operator, then prove the release process end to end with the first published release.

## Outcome
The goal was met. The chart now publishes to GHCR as `oci://ghcr.io/meigma/template-k8s/chart`, Release Please keeps the operator and chart versions aligned, the release workflow builds binaries plus a multi-platform image, and `v0.1.1` was published and smoke-tested from the published Helm chart with the published operator image.

## Key Decisions
- Rename only chart metadata to `chart` -> Helm derives the OCI basename from `Chart.yaml:name`, while helpers preserve rendered Kubernetes identity as `template-k8s`.
- Publish the chart only to GHCR OCI -> this keeps GitHub release assets focused on binaries, checksums, and SBOMs while making Helm installation use the canonical registry path.
- Split container image builds by platform and run arm64 on native GitHub runners -> this removed the slow QEMU path and made release dry-runs practical.
- Add Docker GHCR login before Helm chart attestation -> `helm registry login` is not enough for `actions/attest` when pushing an OCI attestation to GHCR.
- Make chart push recovery digest-aware -> a rerun after a partial release can recover the already-published chart digest with `helm show chart` and still complete attestation.

## Changes
- `.github/workflows/release.yml` - added Helm chart packaging, OCI push, chart attestation, native platform container builds, manifest assembly, digest-aware chart retry behavior, and release summary commands.
- `.github/workflows/release-dry-run.yml` - added a Helm chart dry-run and split container platform dry-runs to match the release workflow shape without publishing.
- `.github/scripts/stage_release_assets.py` and `.github/scripts/test_stage_release_assets.py` - extracted release asset staging and checksum validation from workflow shell into a tested script.
- `charts/template-k8s/Chart.yaml` - renamed chart metadata to `chart` and kept Release Please-owned `version` and `appVersion` aligned with the operator release.
- `charts/template-k8s/templates/_helpers.tpl` - preserved `template-k8s` as the default rendered resource and app identity while allowing the package name to be `chart`.
- `.release-please-manifest.json`, `CHANGELOG.md`, and `charts/template-k8s/Chart.yaml` - Release Please cut `v0.1.1`.

## Open Threads
- Release Please opened PR #22, `chore(master): release 0.1.2`, after the Helm attestation recovery fix. It is green and clean, but it was left unmerged for normal review rather than folded into this closeout.
- GitHub Actions warns that the pinned `actions/upload-artifact` and `actions/download-artifact` actions still run on Node.js 20. Update those pins before GitHub's Node 24 migration becomes disruptive.
- The successful recovery attestation run for `v0.1.1` was manually dispatched from `refs/heads/master`; the next normal tag-triggered release should verify with `--source-ref refs/tags/vX.Y.Z`.

## Lessons
- Helm OCI publication path and chart metadata are coupled: `helm push chart-0.1.1.tgz oci://ghcr.io/meigma/template-k8s` publishes `ghcr.io/meigma/template-k8s/chart:0.1.1` because the package basename comes from `Chart.yaml:name`.
- GitHub's draft release UI can be misleading: draft releases may show temporary `untagged-*` URLs even while `gh release view` can see the tag and assets.
- A release rerun can encounter partially-published artifacts; release workflows should either be idempotent or explicitly recover the published digest before attesting.

## References
- PR #19: https://github.com/meigma/template-k8s/pull/19
- PR #20: https://github.com/meigma/template-k8s/pull/20
- PR #14: https://github.com/meigma/template-k8s/pull/14
- PR #21: https://github.com/meigma/template-k8s/pull/21
- Open PR #22: https://github.com/meigma/template-k8s/pull/22
- Release `v0.1.1`: https://github.com/meigma/template-k8s/releases/tag/v0.1.1
- Successful recovery release run: https://github.com/meigma/template-k8s/actions/runs/26141230186
- Published chart: `oci://ghcr.io/meigma/template-k8s/chart --version 0.1.1`
