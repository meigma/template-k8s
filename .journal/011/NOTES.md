---
id: 011
title: Session kickoff
started: 2026-05-19
---

## 2026-05-19 22:39 — Kickoff
Goal for the session: Start a new journal session and wait for the user's substantive request.
Current state of the world: `master` is clean at `8b9bd5f` (`docs: add downstream operator setup guidance (#24)`), the personal journal worktree is on `journal/jmgilman` and was already up to date, and recent closed sessions cover the Helm OCI release path, downstream operator setup guidance, and the godoc sweep. `.journal/TECH_NOTES.md` records the current `NginxDeployment` operator baseline, Moon task expectations, chart/release shape, and the known moon 2.0.0-rc.1 local script-task regression.
Plan: Hold here until the user gives the actual implementation, review, or research request; update this notes file at meaningful checkpoints during that work.

## 2026-05-19 22:51 — Kyverno image verification chart option
Goal for the slice: Add an optional Helm chart input that creates a Kyverno policy validating the signed operator image produced by the release pipeline.
Current state of the world: The release workflow attests the final multi-platform GHCR image digest with `actions/attest` and `push-to-registry: true`. Current Kyverno docs show GitHub Artifact Attestation verification through `verifyImages` with `type: SigstoreBundle`, GitHub Actions OIDC issuer, Rekor, SLSA provenance type, and `buildDefinition.buildType == https://actions.github.io/buildtypes/workflow/v1`.
What changed: Added disabled-by-default `kyverno.imageVerification` values and schema, rendered an optional `ClusterPolicy` from `charts/template-k8s/templates/kyverno-image-policy.yaml`, defaulted image matches to the chart's operator repository tag and digest forms, documented the Helm flag in the downstream README template, updated `DELETE_ME.md` for downstream workflow identity replacement, and added chart tests for default-disabled and enabled policy output.
Validation: `go test ./test/chart -count=1`, `moon run root:chart-validate`, enabled Kyverno `helm install --dry-run=client`, `git diff --check`, and `moon ci --summary minimal` all pass after fixing the initial test lint issues.

## 2026-05-20 07:00 — Kind/Kyverno functional test
Goal for the checkpoint: Manually prove the Helm chart's optional Kyverno policy accepts the released signed operator image in a real Kind cluster.
What was tested: Created Kind cluster `template-k8s-kyverno` with Kubernetes v1.35.0, installed Kyverno chart `3.8.1` / Kyverno `v1.18.1`, installed the local chart with `kyverno.imageVerification.enabled=true`, and forced fresh manager Pod admissions after the policy was active.
What failed first: The original chart default expected the keyless subject `.../release.yml@refs/tags/*`, but `gh attestation verify oci://ghcr.io/meigma/template-k8s:v0.1.1` showed the current released image's GitHub Artifact Attestation was produced by workflow dispatch from `refs/heads/master`. Kyverno correctly denied a fresh manager Pod with `sigstore bundle verification failed: no matching signatures found`.
Fix and result: Changed the default keyless identity to a `subjectRegExp` accepting the release workflow at either SemVer tag refs or `refs/heads/master`. After reinstalling those chart defaults and forcing a new Pod, Kyverno logged `image attestations verification succeeded` with `verifiedCount=1`, admitted `template-k8s-controller-manager-957b55bb6-zdpbh`, and mutated its image to `ghcr.io/meigma/template-k8s:v0.1.1@sha256:fe721300df5d0e11d07b4304f51c9222618ad65a42e9578aff842b2341961f46`.
Validation: `go test ./test/chart -count=1`, `moon run root:chart-validate`, `git diff --check`, and `moon ci --summary minimal` pass after the fix.

## 2026-05-20 08:41 — v0.1.2 provenance retest
Goal for the checkpoint: Remove the temporary `refs/heads/master` trust fallback and retest the Kyverno chart option against a release with normal tag provenance.
What changed: Updated the chart default and chart test so `kyverno.imageVerification.attestor.subjectRegExp` only accepts the release workflow identity at SemVer tag refs, specifically `refs/tags/v*`, with no `master` branch fallback.
Release evidence: `gh attestation verify oci://ghcr.io/meigma/template-k8s:v0.1.2 --repo meigma/template-k8s --signer-workflow meigma/template-k8s/.github/workflows/release.yml --source-ref refs/tags/v0.1.2 --deny-self-hosted-runners --format json` passed. The verified identity was `.../release.yml@refs/tags/v0.1.2`, the source ref was `refs/tags/v0.1.2`, the event was `push`, and the attested image digest was `sha256:ec3e8c33a7822195ee84b59575eaeac1a873e78c0dcd293fd1ebeab148cad775`.
Functional result: In Kind cluster `template-k8s-kyverno` with Kubernetes v1.35.0 and Kyverno chart `3.8.1` / Kyverno `v1.18.1`, the local chart installed with `kyverno.imageVerification.enabled=true` and `image.tag=v0.1.2`. After scaling the manager deployment down and back up to force a fresh admission after the policy was ready, Kyverno logged `image attestations verification succeeded` with `requiredCount=1` and `verifiedCount=1`, logged `validation passed`, and admitted `template-k8s-controller-manager-9bc4cc9b7-2vkqb` with image `ghcr.io/meigma/template-k8s:v0.1.2@sha256:ec3e8c33a7822195ee84b59575eaeac1a873e78c0dcd293fd1ebeab148cad775`.
Validation: `go test ./test/chart -count=1`, `python3 -m json.tool charts/template-k8s/values.schema.json`, enabled-policy `helm template`, `git diff --check`, and `moon ci --summary minimal` pass. The temporary Kind cluster was deleted afterward; the unrelated `oidc-smoke` Kind cluster was left in place.
