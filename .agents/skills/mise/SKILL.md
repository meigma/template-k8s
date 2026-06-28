---
name: mise
description: >
  Operate mise as the single source of truth for tool versions and integrity in
  template-k8s. Use when touching mise.toml or mise.lock, bumping or adding a
  pinned tool (go, moon, golangci-lint, controller-gen, kubebuilder, setup-envtest,
  kubectl, helm, chainsaw, ko, tilt, ctlptl, kind, melange, apko, cosign),
  resolving "command not found"/PATH problems, fixing locked/trust failures, or
  wiring mise into moon, the CI workflow, or the local dev/image-build tasks.
---

# mise

mise owns the lifecycle of every pinned tool and the project's tool-related env in
this repo. It replaced Proto (`.prototools`, `.moon/proto/*`) and the `.envrc`
direnv shim activation. Treat `mise.toml` + `mise.lock` as the only place a
toolchain version is declared; everything else (moon, CI, the container image
build, the Kind dev stack) consumes what mise puts on PATH.

## Verified against

- `mise 2026.6.14` (`macos-arm64`, build 2026-06-25), grounded in the captured
  `--help` for
  `install/use/ls/lock/exec/run/trust/outdated/upgrade/settings/current/activate/env/which`
  and `mise doctor --help`, plus this repo's `mise.toml`, `mise.lock`, `moon.yml`,
  `.moon/toolchains.yml`, and `.github/workflows/ci.yml`.
- Advice is grounded in the local CLI and these files, not memory. Re-verify on a
  mise minor/major bump.

## Use this skill when

- Bumping or adding a tool, or reviewing a diff that touches `mise.toml`/`mise.lock`.
- A tool is missing from PATH, or `mise install` fails closed under `locked`.
- mise prompts for trust (commonly inside a `.wt/` worktree that nests under the repo).
- Explaining how moon, `ci.yml`, `moon run root:dev-up`, or `moon run root:test-e2e`
  get their binaries.

## mise's lane (non-negotiables)

mise manages **tool + env lifecycle only**. State these as rules:

1. mise is **not the task runner and not the CI gate** — that is moon. Do not move
   build/lint/test/codegen into mise tasks.
2. **This repo defines no mise `[tasks]`.** mise is purely `[tools]`, `[env]`, and
   `[settings]`. The local conveniences — the melange/apko container image build and
   the Kind dev stack — are **moon** tasks (`root:dev-up`, `root:dev-down`,
   `root:test-e2e`) that run mise-provisioned binaries. Do not add general-purpose
   mise tasks.
3. **Every tool an engineer needs goes through mise.** Never `go install`,
   `go tool`, `brew install`, `apt`, `npm -g`, `cargo install`, or a manual
   download for project tooling. Add it to `[tools]` and `mise lock` instead.
4. **Force the verifying backend.** Pin CLIs with an explicit `aqua:` ref, e.g.
   `"aqua:kyverno/chainsaw" = "0.2.15"`. A bare short name (`chainsaw`, `kubectl`,
   `helm`) is not in mise's curated registry and/or resolves through a backend with
   no recorded checksum — always use the explicit `aqua:<owner>/<repo>` ref so the
   tool lands with a pinned URL + checksum in `mise.lock`. The one deliberate
   exception is `controller-gen` (see rule 5).
5. **`controller-gen` is the only non-aqua tool**, pinned via the **go: backend**
   (`"go:sigs.k8s.io/controller-tools/cmd/controller-gen" = "0.21.0"`) because no
   aqua package exists. Its integrity comes from the **Go module checksum database**
   (`go.sum`/sumdb), not from a `mise.lock` URL+checksum. Keep it on the go: backend;
   do not try to invent an aqua ref for it.
6. **Bump = edit `mise.toml`, then `mise lock`, then commit both together.** Never
   hand-edit `mise.lock` (`# @generated`) except the one documented moon macos-x64
   workaround below, and never commit one file without the other.

## How mise is wired here

`mise.toml`:

- `[tools]`: `go = "1.26.3"` (core backend, authoritative, matches `go.mod`'s
  `go 1.26.3`); `controller-gen` via the **go: backend** (version-only lock entry);
  and fourteen CLIs pinned via explicit `aqua:` refs, grouped by job:
  - task runner / CI gate: `moonrepo/moon`
  - lint: `golangci/golangci-lint`
  - Kubebuilder operator toolchain: `kubernetes-sigs/kubebuilder`,
    `kubernetes-sigs/controller-runtime/setup-envtest` (deeper `owner/repo/subpath`
    aqua ref), `kubernetes/kubernetes/kubectl`, `helm/helm`, `kyverno/chainsaw`
  - local Kind dev stack: `ko-build/ko`, `tilt-dev/tilt`, `tilt-dev/ctlptl`,
    `kubernetes-sigs/kind`
  - supply-chain / image build: `chainguard-dev/melange`, `chainguard-dev/apko`,
    `sigstore/cosign`
- `[env] GOTOOLCHAIN = "local"`: never auto-download a Go toolchain other than the
  pinned one; matches `go.mod`'s `go 1.26.3`. mise `[env]` is **not** carried by the
  CI action's shims, so `ci.yml` also sets `GOTOOLCHAIN: local` at job level — keep
  both in sync.
- `[settings] lockfile = true` (read/write `mise.lock`) and `locked = true` (the
  integrity gate; equivalent to the `--locked` flag / `MISE_LOCKED=1`).

moon consumes mise, it does not duplicate it: `.moon/toolchains.yml` declares no
language toolchain and `moon.yml` sets `toolchains.default: system`, so every moon
task command is a bare binary (`go`, `golangci-lint`, `controller-gen`,
`setup-envtest`, `chainsaw`, `helm`, `kubectl`, `ctlptl`, `tilt`, `ko`, `kind`)
resolved from PATH. `moon.yml` also collects `mise.toml` + `mise.lock` into the
`toolchainConfig` fileGroup and lists it as an input of every task. Under the
`system` toolchain moon does **not** hash a tool binary's version, so those pins are
the task inputs that force a re-run when a tool bumps (every task here runs
`cache: false`, so this is purely re-trigger, not result-cache invalidation). See
the `worktrunk` skill for worktree mechanics, the `k8s-operator` skill for the Moon
task surface, and the `melange`/`apko` skills for the image build those pinned tools
feed.

CI (`.github/workflows/ci.yml`) installs via
`jdx/mise-action@…v4.2.0 with: version: 2026.6.14, cache: true`. The action installs
every tool from `mise.toml` honoring `mise.lock` (locked → fail closed), including
`moon`, and prepends the shim dir to PATH so moon's `system` tasks find the
binaries. CI uses mise-action, **not** `moonrepo/setup-toolchain`.

## The lockfile, precisely

`mise.lock` is `# @generated`. Per tool it records a `[[tools."<ref>"]]` block
(`version`, `backend`) and one `[tools."<ref>"."platforms.<plat>"]` table for each
of the four platforms: `linux-x64`, `linux-arm64`, `macos-x64`, `macos-arm64`.

- Every platform entry carries a `url`. **`locked = true` requires a pre-resolved
  `url` per platform** and fails closed otherwise (per `mise install --help`: it
  prevents API calls to GitHub/aqua at install time).
- Every aqua platform entry in this repo also carries an enforced
  `checksum = "sha256:…"`. There is no missing-checksum exception here — do not
  assume a tool may legitimately ship without one.
- A subset additionally records verification provenance, reflecting what the aqua
  registry applies for that tool. In this repo: `provenance = "github-attestations"`
  on `golangci-lint`; `provenance = "cosign"` on `cosign`; and
  `[…"platforms.<plat>".provenance.slsa]` subtables on `ko` and `chainsaw`. The
  remaining tools (`go`, `moon`, `kubebuilder`, `setup-envtest`, `kubectl`, `helm`,
  `tilt`, `ctlptl`, `kind`, `melange`, `apko`) carry a pinned `url` + `checksum` but
  no `provenance` field. Do **not** claim every tool is attestation/SLSA/cosign
  verified; the always-on guarantees are the pinned `url` and `checksum`.
- **`controller-gen` is the single tool whose integrity lives outside `mise.lock`.**
  Its block is `version` + `backend` only — no platform tables, no `url`, no
  `checksum` — because the go: backend builds it from source and verifies against the
  Go checksum database. `mise lock` will not (and should not) write platform
  url/checksum rows for it.

### moon macos-x64 lockfile quirk

`mise lock` resolves moon's `macos-x64` artifact but does **not** persist its
`[tools."aqua:moonrepo/moon"."platforms.macos-x64"]` table — a mise write quirk.
That table is **hand-added** (with an inline comment) from moon's published v2.3.5
`.sha256` sidecar. This is the one sanctioned hand-edit to `mise.lock`. After any
re-lock that touches moon, confirm the macos-x64 entry is still present and re-add it
from the upstream `.sha256` if it disappeared, before committing.

## Bumping a tool (the canonical operation)

```bash
# 1. edit the version in mise.toml (keep the aqua: / go: ref)
# 2. re-resolve url/checksum for all four platforms
mise lock --platform linux-x64,linux-arm64,macos-x64,macos-arm64
# 3. commit mise.toml + mise.lock together
```

- `mise outdated` (add `--bump` to see latest across major lines, `-J` for JSON)
  shows what could move before you decide.
- `mise upgrade <tool> --bump` is the one-shot equivalent (edits `mise.toml` and
  re-locks), but the repo's committed convention is the explicit edit + `mise lock`
  so the version change is a reviewable diff.
- After locking, confirm all four platform tables are present for each changed aqua
  tool, and **re-check the moon macos-x64 entry** (it may have been dropped); do not
  ship a partial lock entry.
- Bumping `controller-gen` is just the version edit in `mise.toml`; `mise lock`
  records only its version/backend (no platform rows), and integrity follows from the
  Go proxy/sumdb on next install.

## Adding a tool

1. Add `"aqua:<owner>/<repo>" = "<version>"` to `[tools]` in `mise.toml` (or, only
   when no aqua package exists, a `"go:<module/path>" = "<version>"` go: entry).
2. `mise lock --platform linux-x64,linux-arm64,macos-x64,macos-arm64` to populate
   url/checksum for all platforms (a no-op for a go: entry beyond version/backend).
3. If a moon task uses it, it is already covered by the `toolchainConfig` input
   fileGroup; add the tool's binary to any narrower task input only if needed.
4. `mise install` locally to materialize it, then commit `mise.toml` + `mise.lock`.

## Worktree trust gotcha

`.wt/` worktrees nest **under** the repo, so mise's upward config search loads both
the worktree's `mise.toml` and the parent checkout's `mise.toml`
(`/Users/josh/code/meigma/template-k8s/mise.toml` exists). When mise prompts, trust
both:

```bash
mise trust --all      # trust this dir and its parents
mise trust --show     # inspect trust status without changing it
```

## Inspection / read-only ops

```bash
mise ls                       # installed + active tool versions (-J for JSON)
mise current                  # active versions only, script-friendly
mise which controller-gen     # resolved bin path; --version for just the version
mise which chainsaw
mise outdated                 # what could bump
mise doctor                   # diagnose install/PATH problems (doctor path prints PATH)
mise exec -- kubebuilder version   # run a pinned tool ad hoc, no shell activation
```

## Gotchas

- `mise install` installs but does **not** activate — tools are not on PATH until
  `mise activate` runs in the shell, or you go through `mise exec` / shims. CI relies
  on mise-action prepending the shim dir; locally use `eval "$(mise activate zsh)"`
  once, or prefix one-off commands with `mise exec --`.
- `mise.local.toml` / `.mise.local.toml` are gitignored per-developer overrides.
  Never commit them and never put project pins there — project pins belong in the
  committed `mise.toml`.
- The moon `macos-x64` lock entry is the one hand-maintained row in `mise.lock`
  (see above). A re-lock can silently drop it; re-add it before committing.
- `controller-gen` is the only go: backend tool. Treat the absence of platform
  url/checksum rows for it as correct, not as an incomplete lock.
- There are no mise tasks here. The container image build (`melange.yaml`/`apko.yaml`
  via `dev/image-build.sh`) and the Kind dev stack (`ctlptl` + `tilt` + `ko` + `kind`)
  run through moon (`root:test-e2e`, `root:dev-up`, `root:dev-down`). See the
  `melange`/`apko` and `k8s-operator` skills.

## Command reference

See [references/mise-commands.md](references/mise-commands.md) for the version-stamped
command and flag map.
