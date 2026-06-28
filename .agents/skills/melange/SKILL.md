---
name: melange
description: >
  Build the operator's signed Wolfi apk with melange. Use when editing melange.yaml or its
  go/build pipeline, adding a build-time package or Go toolchain, signing the apk, or debugging
  `dev/image-build.sh` or the release/dry-run `melange-build` job. This is the source-to-signed-apk
  step that apko later turns into the operator's runtime image.
---

# Melange

melange has exactly one job in this repo: compile the operator's controller-manager binary into
a signed [Wolfi](https://github.com/wolfi-dev) apk described by `melange.yaml`. That apk is the
only artifact apko assembles into the runtime image. There is no Dockerfile (this replaced the
former multi-stage Dockerfile). Ground every command in `--help` and the repo files below, not
memory.

## Verified against

- melange `v0.54.0` (GitCommit `7fb1d6a`), pinned in `mise.toml` as
  `aqua:chainguard-dev/melange = "0.54.0"` and locked per-platform in `mise.lock`. Run it via
  mise (`mise exec -- melange ...`) or an activated mise shell; do not install it any other way.
- Grounded in local `melange --help` / `melange build --help` and the repo files: `melange.yaml`,
  `mise.toml`, `dev/image-build.sh`, `.github/workflows/release.yml`, `release-dry-run.yml`, and
  `security-scan.yml`.
- Sibling skills: `mise` provisions melange and puts it on PATH; `apko` consumes the apk this
  step produces. See those skills for their lanes.

## Use this skill when

- Editing `melange.yaml` or the `go/build` pipeline.
- Adding a build-time package or Go toolchain to the apk.
- Signing an apk or reasoning about the melange-to-apko key handoff.
- Debugging a failed `dev/image-build.sh` (local) or the release / dry-run `melange-build` job.

## melange's lane (non-negotiable)

1. melange does ONE thing: source to signed Wolfi apk. It does not build the image (that is
   apko) and it does not push anything anywhere.
2. The apk is the single artifact. apko reads it from `./packages` via `@local`; nothing else
   consumes, tags, or distributes it.
3. Never hand-edit `package.version` in `melange.yaml`. It carries `# x-release-please-version`
   and release-please owns it (registered via `extra-files` in `release-please-config.json`). Do
   not run `melange bump` in this repo.
4. This operator stamps NO build vars. `cmd` has no `main.version`/`commit`/`date` symbol, so
   `melange.yaml` has no `vars:` block, no `--vars-file`, and no `-X` ldflags. The only ldflag is
   `-buildid=` (for reproducibility). Do not add version/commit/date plumbing unless `cmd` first
   grows the variables to receive it.
5. Never commit signing keys. `melange*.rsa`, `melange*.rsa.pub`, `melange-vars.yaml`, and
   `/packages/` are gitignored. Keys are ephemeral, minted per build; the private key never
   leaves the machine that built the apk.
6. No Dockerfile, no `RUN`, no `apt`. Build-time tools come from
   `environment.contents.packages` (the `go-1.26` Wolfi package). Runtime dependencies belong in
   `apko.yaml`, not here.
7. Always pass `--runner docker`. Do not rely on the platform default runner.

## melange.yaml anatomy

- `package`: `name: template-k8s`, `version: "0.1.2"` (release-please marker), `epoch: 0`, and a
  `description`.
- `environment.contents`: the Wolfi `os` repository + its signing keyring, plus
  `packages: [go-1.26]` (the build toolchain). `environment.environment.CGO_ENABLED: "0"`.
- `pipeline: - uses: go/build` with `packages: ./cmd`, `output: manager`, `go-package: go-1.26`,
  `modroot: .`, `strip: "-s -w"`, `ldflags: "-buildid="`, and
  `extra-args: "-mod=readonly -buildvcs=false"`. The `go/build` builtin auto-adds `-trimpath` and
  installs to `/usr/bin/<output>` — apko's entrypoint `/usr/bin/manager` depends on that path.
  `--source-dir .` mounts the module so `go/build` compiles `./cmd` against the rest of the
  operator's packages (`api/`, `internal/`).
- There is no `vars:` block. Unlike a version-stamped build there is nothing to override per
  build; this mirrors the prior Dockerfile (`go build -o manager ./cmd`), which also stamped no
  version symbols.

## Build the apk locally

`dev/image-build.sh` is the supported local path; it runs melange then apko and loads the image
into Docker (default tag `template-k8s:dev`, override with `IMG`). `scripts/test-e2e.sh` calls it
with the e2e image tag, and the `moon root:test-e2e` task lists it as an input. The melange
portion is:

```bash
arch="$(go env GOARCH)"
melange keygen melange.rsa
melange build melange.yaml \
  --arch "$arch" \
  --runner docker \
  --signing-key melange.rsa \
  --source-dir .
```

This drops a signed apk under `./packages/<apkdir>/` (default `--out-dir` is `./packages/`).
`<apkdir>` is the Wolfi arch name, not the Go arch: `amd64` to `x86_64`, `arm64` to `aarch64`.
apko then reads the whole `./packages` directory as the `@local` repository. There is no
`--vars-file`: nothing is stamped.

## How the release / CI build differs

`release.yml` (`melange-build`) and `release-dry-run.yml` (`melange-build-dry-run`) run the same
`melange build` invocation under a matrix, one arch per NATIVE runner — `amd64` on `ubuntu-24.04`,
`arm64` on `ubuntu-24.04-arm`. No QEMU. Differences from local:

- Each runner mints its own ephemeral key with a distinct name: `melange keygen melange-<arch>.rsa`.
- Each runner builds with `--arch <arch> --runner docker --signing-key melange-<arch>.rsa
  --source-dir .` — the same invocation as local, still no vars file.
- Each runner uploads `packages/<apkdir>/**` plus its `melange-<arch>.rsa.pub` (artifact
  `apk-<arch>`). The private key never leaves the runner; apko later trusts the apk via the
  uploaded public keys.
- The dry-run job is identical to the release build except it never pushes, signs, or attests —
  it just feeds `apko build` (not `apko publish`) to assemble and smoke-test the image.

`security-scan.yml` builds only `amd64` for the Trivy scan — the same `melange build`, but it
mints a single `melange.rsa` key (not the per-arch `melange-<arch>.rsa` naming above).

## Signing and the apko handoff

`--signing-key` signs the apk during the build. apko must trust that signature to install the
`@local` apk, so the matching public key is appended to apko's keyring with `--keyring-append`.
Locally that is `--keyring-append ./melange.rsa.pub`; in the release/dry-run jobs apko appends both
arches' keys (`./melange-amd64.rsa.pub`, `./melange-arm64.rsa.pub`). Omit the public key and apko
rejects the apk as untrusted. See the `apko` skill.

## Gotchas

- `--runner docker` is required wherever melange runs: it needs a Linux build sandbox, and on
  macOS/Docker Desktop the Docker runner provides it. Docker must be running. Valid runners are
  `bubblewrap`, `docker`, `qemu`; this repo always uses `docker`.
- Wolfi arch name is not the Go arch in output paths: `amd64`→`x86_64`, `arm64`→`aarch64`. Match
  these when globbing `packages/<apkdir>/**`.
- melange produces an SBOM for the apk (`--namespace` sets its package-URL namespace). The
  image-level SBOM (syft) and SLSA provenance (`attest.yml`) are produced later, NOT by melange.
  Do not add `--generate-provenance` to chase that; the repo does not use it.
- No build vars: this operator does not stamp version/commit/date. Do not add a `vars:` block, a
  `--vars-file`, or `-X` ldflags unless `cmd` first gains the variables to receive them. The lone
  `-buildid=` ldflag is for reproducibility, not metadata.
- Neither local nor CI uses melange's `--build-date` flag (that controls in-image file timestamps
  for reproducibility, a separate concern).
- Do not override `output:` in the pipeline without updating `apko.yaml`'s `entrypoint.command`
  and `contents.packages` — the binary path `/usr/bin/manager` is contractual.
- To add a build-time tool or a different Go toolchain, edit `environment.contents.packages`
  (Wolfi package names) in `melange.yaml`. Do not install inside the sandbox.

## Command reference

See [references/melange-commands.md](references/melange-commands.md) for the version-stamped
command and flag reference.
