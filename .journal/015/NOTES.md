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

## 2026-06-28 13:50 — PR1 shipped (PR #42, open, CI running)
**PR #42** `build(tooling): replace proto with mise and run moon on system
binaries` — branch `build/proto-to-mise` (commit `5477d30`), off master.

Backend reality (verified with local `mise 2026.6.14`, supersedes the assessment):
- 10 verifying `aqua:` backends: golangci-lint, kubebuilder, setup-envtest
  (`kubernetes-sigs/controller-runtime/setup-envtest@0.24.1`), kubectl, helm,
  `aqua:kyverno/chainsaw` (full path), ko, tilt, ctlptl, kind — plus
  `aqua:moonrepo/moon@2.3.5`.
- **controller-gen → Go backend** `go:sigs.k8s.io/controller-tools/cmd/controller-gen@0.21.0`,
  NOT ubi. ubi is DEPRECATED in mise (removed 2027.1.0) AND mis-names the binary
  `controller-tools`; the Go backend is canonical and yields a correct
  `controller-gen` (integrity via the Go checksum DB). It locks as a version-only
  entry and `locked=true` tolerates it.
- melange 0.54.0 / apko 1.2.19 / cosign 3.1.1 pinned now (for PR2/PR3).
- mise quirk reproduced: `mise lock` dropped moon's `macos-x64` entry; hand-added
  from moon's published `.sha256` (`ffc0bf6e…`), with a comment.

What landed: `mise.toml` + `mise.lock` (16 tools, fail-closed `locked=true`);
`.moon/toolchains.yml` → comments-only; `moon.yml` gains `toolchains.default:
system` and the `toolchainConfig` fileGroup now tracks mise.toml/mise.lock;
`ci.yml` + the Helm jobs in `release.yml`/`release-dry-run.yml` →
`jdx/mise-action@v4.2.0`, `GOTOOLCHAIN: local`, cache keys on `mise.lock`;
`setup-go` → `go.mod`; deleted `.prototools`/`.moon/proto/*`/`.envrc`/`.go-version`;
README Prerequisites + AGENTS dev-setup → mise.

Local proof (clean env): `moon run root:check` ✓ and `moon run root:test` ✓
(envtest k8s 1.35.x via mise setup-envtest). NOTE — local validation needs the
machine's stray go 1.26.4 neutralized: this host leaks `~/.proto/bin/go`,
`~/.goenv/shims/go`, and exported `GOPATH=~/go/1.26.4`; combined with mise's
exported `GOROOT=…/1.26.3` that yields a golangci-lint "go1.26.3 does not match
go tool version go1.26.4" typecheck error. Run gates via
`mise exec -- bash -c 'PATH without /.proto/shims; unset GOROOT; moon run …'`.
A clean CI runner has none of this. `go.sum` churn from a stray `-mod=mod`
diagnostic was reverted; final tree leaves go.sum pristine.

PR1 merged to master `f708c83` (squash). Worktree/branch cleaned.

## 2026-06-28 14:12 — PR2 built + validated (commit b352449, review running)
Branch `build/melange-apko` off master, commit `b352449`
`build(release): build the container image with melange + apko`.

Grounded design facts (verified against both repos before authoring):
- Release image is built from `Dockerfile` + `docker buildx` (NOT ko); ko is only
  the Tilt dev builder → untouched. Swap targets the release path.
- **Kyverno needs NO change in PR2.** values.yaml defaults trust
  `release.yml@refs/tags/v…` + `buildType .../workflow/v1` + type
  `slsa.dev/provenance/v1`. The Kyverno-verified attestation is the inline
  `actions/attest@v4.1.0` step (empirically produces SLSA provenance; no
  predicate-type needed). PR2 keeps that step on the apko index digest, signer
  stays release.yml → contract preserved. Signer flip is PR3.
- cmd has NO version vars → melange ldflags `-buildid=` + strip only (matches the
  old Dockerfile's no-stamp behavior); no `-X`/vars-file plumbing.
- master is NOT branch-protected → free to rename dry-run jobs
  (container-image-platform-dry-run → melange-build-dry-run).
- `scripts/test-e2e.sh` did `docker build .` → had to repoint e2e at the apko
  image (new `dev/image-build.sh`, melange+apko host-arch → kind load).

What landed: `melange.yaml` (./cmd → /usr/bin/manager apk), `apko.yaml` (Wolfi
nonroot 65532, ca-certs/tzdata, amd64+arm64, entrypoint /usr/bin/manager);
release.yml `container-image-build`(buildx matrix)→`melange-build` + apko-publish
`container-image-release` with keyless cosign + syft SBOM attest + preserved
`actions/attest` provenance; release-dry-run.yml + security-scan.yml → melange/apko;
`dev/image-build.sh` + test-e2e.sh; release-please extra-files += melange/apko;
deleted Dockerfile/.dockerignore; gitignore local apk/key/sbom artifacts.

LOCAL PROOF: `dev/image-build.sh` built the apk (`cmd:manager=0.1.2-r0`) + apko
image; `docker run template-k8s:dev --help` → "Kubernetes controller manager"
exit 0; **User=65532, Entrypoint /usr/bin/manager, ~51MB, no shell** (matches the
former distroless:nonroot). `helm template --set kyverno...enabled=true` still
renders the policy trusting release.yml + workflow/v1 buildType. `root:check` ✓,
`git diff --check` clean, no artifact leaks.

Running an adversarial review workflow (4 reviewers: release-graph,
supplychain/kyverno, melange-apko-build, completeness) before pushing. apko index
digest parse (`tail -n1`) + buildx-imagetools-without-setup-buildx are inherited
from the proven template-go-api v1.0.4 pipeline.

## 2026-06-28 14:29 — PR2 shipped + CI/dry-run green (PR #43, commit 83e16e5)
Adversarial review (wf_04275e19): 3/4 dimensions clean. Two flagged:
- **BLOCKER "actions/attest needs predicate-type" = FALSE POSITIVE.** Reviewer
  reasoned from general knowledge. Checked ground truth: `gh attestation verify
  oci://ghcr.io/meigma/template-k8s:v0.1.2` → predicateType
  `slsa.dev/provenance/v1`, buildType `actions.github.io/buildtypes/workflow/v1`.
  So `actions/attest@v4.1.0` (no predicate-type) DOES emit SLSA build provenance —
  exactly the Kyverno contract. The step I preserved is correct. (Lesson: verify
  supply-chain claims against the published artifact, not action docs from memory.)
- **HIGH "is the attested digest the multi-arch INDEX digest?" = FIXED.** Switched
  from parsing apko stdout (`tail -n1`) to resolving the index digest
  authoritatively from the registry (`docker buildx imagetools inspect |
  jq .manifest.digest`) — the old buildx job's pattern. cosign/syft-SBOM/provenance
  + job outputs all bind to `steps.manifest.outputs.{digest,ref,name}`.
- Medium (cosign OIDC in release.yml) = not a bug; release.yml is tag-gated +
  resolve-release validates the SemVer tag, and release.yml IS the intended cosign
  signer identity.

PR #43 `build(release): build the container image with melange + apko` (commit
`83e16e5`). PR `ci` ✓ + Kusari ✓. Dispatched `release-dry-run` (workflow_dispatch
on the branch, since dry-run jobs gate to release-please/dispatch) → **ALL GREEN**:
Binary Dry Run ✓, **Melange Build Dry Run amd64 ✓ + arm64 ✓ (native runners, no
QEMU)**, Container Image Dry Run (apko assemble + smoke) ✓, Helm Chart Dry Run ✓.
That validates melange/apko on real CI incl. the ubuntu-24.04-arm runner. The
publish→cosign→attest tag path is unreachable by dry-run (only a real tag runs it).

PR2 open + green, awaiting review/merge. Next: PR3 (provenance → reusable
`attest.yml` for SLSA L3, chart attestation Option A into attest.yml, and the
Kyverno signer flip release.yml→attest.yml in values.yaml — the high-stakes
cross-cutting change).

PR2 merged to master `800bbfb` (squash). Worktree/branch cleaned.

## 2026-06-28 14:49 — PR3 built (commit 41184f1, review running)
Branch `ci/slsa-l3-provenance` off master, commit `41184f1`
`ci(release): generate provenance in an isolated reusable workflow (SLSA L3)`.

Design (Option A — chart provenance also isolated):
- `.github/workflows/attest.yml` (new, `workflow_call`): one isolated `attest` job
  → binary checksums (`actions/attest --subject-checksums`) + OCI provenance
  (`actions/attest-build-provenance@v4.1.1`, used for BOTH image and chart). Own
  GHCR `docker/login` gated on subject-digest. Job declares id-token/attestations/
  contents:read/**packages:write**.
- release.yml: dropped the 3 inline `actions/attest` provenance steps; added
  `attest-binaries` / `attest-image` / `attest-chart` caller jobs. binary job now
  uploads `checksums.txt` as `release-checksums` artifact; **keyless cosign + syft
  image SBOM attest STAY inline** (separate controls, not SLSA provenance).
- 3 hard-won bugs pre-empted: (a) every caller grants `packages: write` (incl. the
  binary one) since the shared reusable job declares it; (b) attest.yml has its own
  GHCR login; (c) reusable-workflow `with:` uses **needs-outputs not env** — so
  helm-chart-release gained a `chart-name` output (`env.CHART_NAME` is illegal in a
  reusable `with:`). image uses container-image-release.outputs.image-{name,digest}.
- **Kyverno signer flip (the landmine):** values.yaml `attestor.subjectRegExp`
  release.yml→attest.yml. `attest-build-provenance` emits the SAME predicate
  (`slsa.dev/provenance/v1`) + buildType (`buildtypes/workflow/v1`) the policy
  already requires — verified live against the v0.1.2 image — so ONLY the signer
  identity changes. Confirmed `helm template --set kyverno…enabled=true` now renders
  attest.yml. values.schema.json unaffected (free-string regex).
- Summary: 3 `gh attestation verify` → attest.yml; added a `cosign verify` line
  (signer stays release.yml — cosign signing is unchanged). DELETE_ME: documented
  attest.yml as the provenance signer + the Kyverno coupling. (DELETE_ME melange/apko
  residue from PR2 — stale "Docker cache scopes", missing melange/apko entries —
  deferred to PR4's docs pass.)

`root:check` ✓, `git diff --check` clean, YAML + job graph valid (all needs resolve;
3 callers wired). The attest.yml path is unreachable by dry-run (only a real tag runs
it) → the throwaway-tag rehearsal is the final check. Adversarial review running
(wf_ef01f372): reusable-workflow perms/contexts, kyverno predicate alignment,
graph/completeness.

## 2026-06-28 14:58 — PR3 review clean, CI caught a test, fixed, green (PR #45)
Adversarial review (wf_ef01f372): **all 3 dimensions ZERO findings** — callers grant
the full perm set, `with:` uses needs-outputs not env, attest.yml self-sufficient
GHCR login, kyverno predicate preserved, graph complete. Also ran **actionlint**
(via `MISE_LOCKED=0 mise x aqua:rhysd/actionlint@1.7.12`) on all 5 workflows → exit
0 (validates reusable-workflow calls/contexts structurally; mise verified
actionlint's own GitHub attestations on install — PR1 integrity in action).

PR #45 opened; dispatched dry-run **all green** (binary + melange amd64/arm64 + apko
+ helm). But PR `ci` **FAILED on root:test**:
`TestKyvernoImageVerificationPolicyRendersGitHubAttestationPolicy`
(test/chart/rbac_test.go) asserts the rendered Kyverno `subjectRegExp`, still
expecting `release.yml`. **MY MISS: for PR3 I ran `root:check` but not `root:test`.**
`root:check`'s chart test is only `go test ./test/chart -run
TestManagerRBACMatchesControllerGen` (RBAC drift); the Kyverno-render assertion lives
in the FULL `go test ./...` that only `root:test` runs. Fixed the test to expect
`attest.yml`, re-ran **full root:test ✓**, amended (`f8b2696`), force-pushed. PR #45
`ci` now **pass** + Kusari pass.

LESSON (durable, added to TECH_NOTES): chart `values.yaml`/template changes must be
validated with `root:test` (full chart suite), not just `root:check` (RBAC-drift
subset). The render-assertion test correctly catching the signer flip is the system
working as intended.

PR3 open + green, awaiting review/merge. Next: **PR4** — port
`.agents/skills/{mise,melange,apko}` (adapt for the operator tool set + Helm/Kyverno)
+ the docs pass (README/AGENTS/DELETE_ME), including the PR2 DELETE_ME melange/apko
residue (stale "Docker cache scopes", missing melange/apko entries) deferred here.
