---
id: 011
title: Kyverno image verification chart option
date: 2026-05-20
status: complete
repos_touched: [template-k8s]
related_sessions: [008, 012]
---

## Goal
Add an optional Helm chart input that creates a Kyverno policy validating the signed operator image produced by the release pipeline, then prove it functionally in Kind with Kyverno installed.

## Outcome
The goal was met. PR #30 was squash-merged as `22b6f50751bbcd604dcb67955ba108f8f15f46a6`, adding a disabled-by-default `kyverno.imageVerification` chart option that renders a Kyverno `ClusterPolicy` for the operator image.

Manual Kind/Kyverno validation first exposed that `v0.1.1` had branch provenance because it was recovered through `workflow_dispatch` on `refs/heads/master`. After `v0.1.2` was cut with normal tag provenance, the temporary branch-trust fallback was removed and the policy passed against the tag-provenance image.

## Key Decisions
- Gate the policy behind `kyverno.imageVerification.enabled` - image policy admission is useful for higher-trust installs but should not make default chart installs depend on Kyverno.
- Use Kyverno `verifyImages` with `type: SigstoreBundle` - this matches the GitHub Artifact Attestation format produced by the release workflow.
- Require SLSA provenance `buildDefinition.buildType == https://actions.github.io/buildtypes/workflow/v1` - this keeps the policy tied to GitHub Actions workflow provenance, not just any keyless signature.
- Default the keyless identity to the release workflow at SemVer tag refs only - `refs/heads/master` was a recovery artifact from `v0.1.1`, and `v0.1.2` confirmed the normal release path now produces tag provenance.

## Changes
- `charts/template-k8s/values.yaml` - added disabled-by-default Kyverno image verification values and tag-only GitHub Actions identity defaults.
- `charts/template-k8s/values.schema.json` - added schema coverage for the optional Kyverno values.
- `charts/template-k8s/templates/kyverno-image-policy.yaml` - added the optional Kyverno `ClusterPolicy`.
- `charts/template-k8s/templates/_helpers.tpl` - added helper support for stable policy naming.
- `test/chart/rbac_test.go` - added chart render coverage for default-disabled and enabled policy output.
- `README.md` - documented the optional downstream chart flag.
- `DELETE_ME.md` - added downstream setup guidance for replacing the workflow and image identity.

## Open Threads
- Downstream repositories generated from this template must replace the default image repository and GitHub workflow identity before enabling the policy.
- If the release workflow identity, attestation predicate, or image publication shape changes, rerun the Kind/Kyverno smoke test before changing the chart defaults.

## Lessons
- Verify policy defaults against the actual published attestation, not only the intended release workflow shape. The `v0.1.1` manual recovery path produced valid provenance with the wrong source ref for a tag-only policy.
- Do not broaden long-term trust to branch refs to accommodate a one-off release recovery when cutting a corrected tag release is viable.

## References
- PR #30: https://github.com/meigma/template-k8s/pull/30
- `v0.1.2` release: https://github.com/meigma/template-k8s/releases/tag/v0.1.2
- Prior release workflow session: `.journal/008/SUMMARY.md`
- Release `v0.1.2` notes: `.journal/012/NOTES.md`
