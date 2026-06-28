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

## 2026-06-28 13:28 — Investigation complete; assessment ready
Source is `~/code/meigma/template-go-api` (the user said "go-template-api"; actual
repo is `template-go-api`). Its **session 015** = "Swap dev tooling to mise +
moon-system + melange/apko (+ SLSA L3)". Shipped as 3 PRs (#24 Proto→mise; #25
Dockerfile→melange/apko + keyless cosign + syft SBOM; #26 provenance in a reusable
`attest.yml` for SLSA L3), proven by a forced release that shook out 3 pipeline
bugs and ended at a verified v1.0.4. Session 016 then added `.agents/skills/{mise,
melange,apko}`.

Ran a 4-agent parallel comparison (workflow wf_1fa81ebc-1e4) across both repos:
tooling, container, release/SLSA, skills/docs. Then verified the agents' biggest
unknown (mise backend availability) empirically with the local `mise 2026.6.14`.

KEY GROUNDED FINDINGS:
- **Tool management (Proto→mise): FULLY FEASIBLE with integrity intact.** All 11
  pinned tools migrate. Verified via `mise registry` / `mise ls-remote`:
  - aqua (verifying — checksum + cosign/SLSA/attestation): golangci-lint,
    kubebuilder, setup-envtest (`kubernetes-sigs/controller-runtime/setup-envtest`),
    kubectl, helm, tilt, ko, kind, ctlptl — plus chainsaw via the FULL path
    `aqua:kyverno/chainsaw` (works; just not in the curated short list).
  - **controller-gen is the ONLY tool with no aqua package** → use
    `ubi:kubernetes-sigs/controller-tools` (resolves 0.21.0; UBI records a
    checksum in mise.lock, so still fail-closed). `go:` is an alt but builds from
    source / no lock hash.
  - go = core. Add melange/apko/cosign (new). No `proto run` wrappers exist in
    moon.yml/scripts (already bare), so task bodies need no edits.
  - template-k8s-specific extra: it uses `.envrc`/direnv (CLAUDE.md "direnv allow")
    that template-go-api never had → mise replaces it.
- **Container (Dockerfile→melange/apko): applies to the RELEASE image only.** The
  release image is built from `Dockerfile` via `docker buildx` (distroless
  static:nonroot, USER 65532, ENTRYPOINT /manager, provenance: mode=max, sbom:
  true). `ko` is ONLY the local Tilt dev builder and stays untouched. melange
  compiles ./cmd → signed Wolfi apk; apko assembles minimal multi-arch nonroot
  image (accounts uid/gid 65532, entrypoint /manager). Faithful reproduction also
  ADDS keyless cosign + syft SBOM (template-k8s has NEITHER today).
- **Release/SLSA L3: reusable `attest.yml`, but template-k8s attests THREE artifact
  classes** (binaries + container image + **Helm chart OCI** at
  `oci://ghcr.io/meigma/template-k8s/chart`) vs template-go-api's two. All three are
  attested inline today via `actions/attest`. Decision: add a third `attest-chart`
  caller (Option A, full L3) or leave the chart inline (Option B).
- **LANDMINE (high): the Kyverno ClusterPolicy is coupled to the signer workflow.**
  `charts/template-k8s/templates/kyverno-image-policy.yaml` verifies the image
  attestation (SigstoreBundle) against `attestor.subject`/`subjectRegExp` (default
  trusts `.github/workflows/release.yml` at tag refs) plus a
  `buildDefinition.buildType` condition. Moving image provenance into `attest.yml`
  changes the attestation SAN to `…/attest.yml@refs/tags/vX`, so the chart's Kyverno
  defaults in `values.yaml` MUST move in lockstep, or image verification breaks for
  downstream installs that enable the policy (validationFailureAction: Enforce).
  Same ripple hits the release-inspection-summary prose + any dry-run signer
  assertions. template-go-api had the signer ripple (ghd.toml/stage script/dry-run)
  but NOT the Kyverno dimension — that's unique to this repo.
- 3 hard-won release bugs to pre-empt (from template-go-api's forced release):
  (a) every attest.yml caller must grant ≥ the reusable job's declared scopes
  (packages: write on the binary caller too); (b) attest-build-provenance
  --push-to-registry needs its OWN docker/login-action inside attest.yml;
  (c) apko --sbom-path <dir> needs the dir pre-created (mkdir -p).

RECOMMENDED SEQUENCING (mirror template-go-api's PR split, fold in k8s extras):
1. PR1 — Proto→mise + moon on system + `jdx/mise-action` in CI (+ release.yml/
   release-dry-run.yml helm job toolchain setup, which use moonrepo/setup-toolchain).
   Delete .prototools/.moon/proto/*/.envrc/.go-version. Lowest risk, self-contained.
2. PR2 — Dockerfile→melange/apko (+ keyless cosign + syft SBOM), native multi-arch.
   Update the Kyverno default identity + buildType to match, and the dry-run image
   build. Delete Dockerfile/.dockerignore.
3. PR3 — provenance into reusable attest.yml (SLSA L3); decide chart attestation
   Option A/B; update Kyverno signer identity to attest.yml + summary/dry-run prose.
4. PR4 — port `.agents/skills/{mise,melange,apko}` (adapt for operator tool set,
   ./cmd, Helm/Kyverno) + README/DELETE_ME/AGENTS/CLAUDE prose.
5. (Optional) a forced/throwaway-tag release rehearsal — only a real tag exercises
   attest.yml + publish + cosign + Kyverno; template-k8s is at v0.1.2.

OPEN DECISIONS for the user (assessment delivered; awaiting direction, no code yet):
- Reproduction fidelity: also ADD keyless cosign + syft SBOM (faithful) — recommend yes.
- Helm chart attestation: move into attest.yml (Option A) vs inline (Option B) — recommend A.
- Forced-release rehearsal: do it (burns a patch tag) vs stop at merged config — TBD.
- PR sequencing: 4 PRs as above vs different bundling.

## 2026-06-28 13:30 — Decisions made; starting PR1
User answered (AskUserQuestion):
1. **Fidelity = Faithful**: add keyless cosign + syft SBOM (on top of existing
   GitHub attestation) in PR2, matching session 015 exactly.
2. **Chart attestation = Option A**: move the Helm chart OCI provenance into the
   reusable `attest.yml` (third `attest-chart` caller) for full SLSA L3 across all
   three artifact classes.
3. **Proceed = Start PR1 (Proto→mise) now.**
Rehearsal decision deferred to after PR3.

PR1 scope: `.prototools` + `.moon/proto/*` (+ `.envrc`, `.go-version`) → `mise.toml`
+ `mise.lock`; moon on system toolchain; CI `moonrepo/setup-toolchain` →
`jdx/mise-action` (ci.yml AND the helm jobs in release.yml/release-dry-run.yml that
use setup-toolchain). Backends settled: 10 aqua (incl `aqua:kyverno/chainsaw` full
path) + `ubi:kubernetes-sigs/controller-tools` for controller-gen. Branch
`build/proto-to-mise` off master, PR + squash merge.
