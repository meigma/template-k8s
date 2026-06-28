---
name: apko
description: >
  Assemble the runtime OCI image for the template-k8s operator with apko (`apko.yaml`) —
  the Dockerfile-free, multi-arch, nonroot image built from the melange-produced apk plus a
  minimal Wolfi base. Use when changing image contents, the `dev/image-build.sh` local build,
  the `apko build`/`apko publish` steps in `.github/workflows/release.yml`,
  `release-dry-run.yml`, or `security-scan.yml`, runtime packages, the nonroot user, OCI
  annotations, the per-build SBOM, or anything that moves the published image digest the
  chart's Kyverno provenance check binds to. Pairs with the `melange` skill (the apk) and the
  `mise` skill (the CLI).
---

# apko

apko owns exactly one step in this repo: turn the melange-built apk plus a small set of Wolfi
base packages into the operator runtime image. It is the modern, distroless-equivalent
replacement for `gcr.io/distroless/static:nonroot` — no Dockerfile, no `RUN`, no shell. The
entire image is declared in `apko.yaml`; everything else (signing, attestation, the apk
itself) belongs to adjacent tools.

This is the **operator manager image** (`ghcr.io/meigma/template-k8s`), the one the Helm chart
runs as the controller. Its published multi-arch digest is the subject of the cosign
signature, the SBOM/provenance attestations, and the chart's optional Kyverno image-verify
policy — so the apko/publish/attest path and the chart's `kyverno.imageVerification` defaults
are coupled (see "Supply-chain tie-in" below).

## Verified against

- apko `v1.2.19` (pinned in `mise.toml` as `aqua:chainguard-dev/apko`, locked in `mise.lock`),
  alongside melange `0.54.0`.
- Grounded in the local `apko --help` for this version and the repo files (`apko.yaml`,
  `mise.toml`, `.github/workflows/release.yml`, `.github/workflows/release-dry-run.yml`,
  `.github/workflows/security-scan.yml`, `dev/image-build.sh`, `scripts/test-e2e.sh`,
  `charts/template-k8s/values.yaml`), not from memory.
- Run apko through mise so the pinned binary is used: `mise exec -- apko <sub> --help`.

## Use this skill when

- Adding or removing a runtime dependency (a Wolfi package in `contents.packages`).
- Editing the nonroot account, entrypoint, archs, or OCI annotations in `apko.yaml`.
- Touching `apko build`/`apko publish` in the release, release-dry-run, or security-scan
  workflows, or the `dev/image-build.sh` local/e2e build.
- Changing anything that moves the published image digest — because the chart's Kyverno
  policy verifies provenance bound to that digest.
- Debugging the published image (wrong arch set, missing apk, untrusted `@local` package,
  SBOM directory errors, manager that does not start nonroot/read-only).

## apko's lane (do not cross it)

1. The image is defined **only** in `apko.yaml`. There is no Dockerfile and there must not be
   one. Never add `RUN`, `apt`, `apk add`, shell steps, or a base-image `FROM`. The inner-loop
   dev stack (`moon run root:dev-up`) builds the manager with **ko**, not apko — do not fold
   the two image paths together.
2. Add a runtime dependency by adding a **Wolfi package** to `contents.packages` — nothing
   else. Keep the set minimal; every package is CVE surface in a distroless image.
3. apko consumes the apk; it never builds it. Source → apk is the `melange` skill's job
   (`melange.yaml`). If the operator code changed, rebuild the apk first, then apko.
4. apko does not sign or attest. `cosign sign` and the `attest.yml` provenance are separate
   release steps. Do not fold them into apko invocations.
5. Keep it nonroot. The image runs as uid/gid 65532 with no shell, matching the chart's
   `containerSecurityContext` (`runAsNonRoot`, `runAsUser: 65532`, `readOnlyRootFilesystem`).
   Do not add a shell, package manager, or root entrypoint for convenience.
6. Do not reintroduce `apko.lock.json` / `apko lock`. The Wolfi base floats by design (see
   below). Pinning is recorded in the per-build SBOM + provenance, not a committed lockfile.

## How the image is wired (`apko.yaml` anatomy)

Read `apko.yaml` before changing anything. The load-bearing parts:

- `contents.repositories`: the Wolfi os repo **and** `@local ./packages`. The `@local`
  repository is melange's output directory — that is how the just-built apk is found.
- `contents.keyring`: the Wolfi signing key URL. The **ephemeral melange public key(s)** are
  not listed here; they are appended at build/publish time with `--keyring-append` and are
  never committed. Without the matching pub key in the keyring, apko refuses the `@local` apk
  as unsigned.
- `contents.packages`: `wolfi-baselayout`, `ca-certificates-bundle`, `tzdata`, and
  `template-k8s@local`. The `@local` suffix pins the package to the `@local` repository, i.e.
  the apk melange just built (not anything from the Wolfi index).
- `accounts`: defines group+user `nonroot` (gid/uid **65532**) and `run-as: 65532`. Wolfi has
  **no `nonroot` package**, so the user is created here. This mirrors `distroless:nonroot` and
  must stay aligned with the chart's `containerSecurityContext.runAsUser: 65532`.
- `entrypoint.command: /usr/bin/manager` — where the `go/build` melange pipeline installs the
  operator manager binary.
- `archs: [amd64, arm64]` — the index architectures.
- `annotations`: OCI labels (`title: template-k8s`, `source: https://github.com/meigma/template-k8s`).
  `org.opencontainers.image.version` carries the `# x-release-please-version` marker;
  release-please bumps it. Do not hand-edit it.

## Local build (`dev/image-build.sh`)

`dev/image-build.sh` builds a single host-arch image with melange + apko and `docker load`s
it, so local e2e exercises the same Wolfi nonroot image that ships. It defaults the tag to
`template-k8s:dev` and honors `IMG` to override it. `scripts/test-e2e.sh` (the
`root:test-e2e` Moon task) drives it with the e2e tag and then `kind load docker-image`s the
result. The apko step is:

```bash
apko build apko.yaml "$image" image.tar \
  --arch "$arch" \
  --keyring-append ./melange.rsa.pub
docker load < image.tar
docker tag "${image}-${arch}" "$image"
```

Non-obvious points:

- **The retag is required, not cosmetic.** A single-arch `apko build` loads into Docker under
  an arch-suffixed tag (`<image>-amd64` / `-arm64`, using the Go arch name from
  `go env GOARCH`). Consumers expect the plain `<image>` tag, so the script retags.
  `release-dry-run.yml` does the same (`template-k8s:dry-run-amd64` → `template-k8s:dry-run`)
  and `security-scan.yml` does the same (`template-k8s:security-scan-amd64` →
  `template-k8s:security-scan`).
- `apko build` writes a tarball for `docker load`. The positional output can also be an
  `oci-layout-dir/`, but the repo uses a `.tar`.
- `--keyring-append ./melange.rsa.pub` makes apko trust the locally-signed `@local` apk. The
  local script mints one ephemeral key (`melange.rsa`); only its single pub key is appended.
- `--arch "$arch"` uses the explicit Go arch name. `host` is also accepted by apko, but the
  script passes the resolved value.
- melange must have run first (`packages/` must contain the apk). The script does this in
  order; if you run apko by hand, build the apk first — see the `melange` skill.
- For Kind-backed e2e, the image must be `kind load docker-image`ed and the Deployment must
  use the exact loaded tag with `imagePullPolicy: IfNotPresent`, or Kind tries a remote pull.

## CI publish (`release.yml` → multi-arch index)

The `container-image-release` job assembles and pushes the multi-arch image:

```bash
mkdir -p sbom            # apko does NOT create --sbom-path; it must pre-exist
apko publish apko.yaml "$IMAGE_TAG" \
  --arch amd64,arm64 \
  --keyring-append ./melange-amd64.rsa.pub \
  --keyring-append ./melange-arm64.rsa.pub \
  --sbom-path ./sbom
```

Non-obvious points:

- **`--sbom-path` must pre-exist.** apko moves SBOMs into the directory but does not create
  it; a missing `sbom/` is a real release failure. The job runs `mkdir -p sbom` first.
- **Two keyring keys.** Each arch's apk was signed on its own native runner with its own
  ephemeral key, so both `melange-amd64.rsa.pub` and `melange-arm64.rsa.pub` must be appended;
  apko verifies each per-arch apk against the matching key.
- **`docker login` is a precondition.** `apko publish` authenticates via the Docker keychain.
  release.yml runs `docker/login-action` against `ghcr.io` first; without it, the push fails.
- **The authoritative digest is resolved from the registry, NOT parsed from apko stdout.**
  The next step runs `docker buildx imagetools inspect "$IMAGE_TAG" --format '{{json .}}'` and
  takes `jq -r '.manifest.digest'` — the multi-arch **index** digest the tag points to. It
  also asserts the platform set is exactly `linux/amd64,linux/arm64`. apko's `--image-refs`
  flag could write refs to a file, but the repo deliberately does not trust apko stdout here:
  cosign, the SBOM + provenance attestations, and the chart's Kyverno check all bind to this
  index digest, so it must come from the registry's own view.
- The image is pushed even while the GitHub release is still a draft — GHCR has no draft
  state. That is expected.
- apko emits its own SBOM (`--sbom-path ./sbom`, spdx by default). The **attested** image SBOM
  is generated separately by `syft <ref> -o spdx-json=image.spdx.json` and attached with
  `actions/attest-sbom`. Two different SBOMs; do not conflate them.

After publish + digest resolution, release.yml (adjacent steps, not apko's job): smoke-tests
`docker run --rm <ref> --help` (accepting exit code 0 or 2, since the manager's flag parser
may exit non-zero on `--help`), then `cosign sign --yes` (keyless, Sigstore/Fulcio via OIDC),
`syft` image SBOM attestation, and finally the isolated `attest.yml` SLSA-L3 provenance via
the `attest-image` caller (passing the resolved `image-name` + `image-digest`,
`push-to-registry: true`).

## Supply-chain tie-in (operator-specific)

The published image is the controller the chart deploys, and the chart ships an **optional
Kyverno image-verification policy** (`kyverno.imageVerification` in
`charts/template-k8s/values.yaml`, default `enabled: false`). When enabled, it verifies the
image's **SLSA provenance** before admitting the operator:

- attestation type `https://slsa.dev/provenance/v1`
- build type `https://actions.github.io/buildtypes/workflow/v1`
- signed by `attest.yml` (the attestor `subjectRegExp` pins
  `…/.github/workflows/attest.yml@refs/tags/vX.Y.Z`), keyless via Fulcio/Rekor.

That provenance is produced by the `attest-image` job over the **same index digest** apko
published. So the build/sign/attest chain and the chart's Kyverno defaults are coupled:
if you change how the image is published (digest resolution), who signs the provenance (the
`attest.yml` signer-workflow identity), or the attestation type/build type, you may need to
update `kyverno.imageVerification` defaults so the policy still matches real releases. Treat a
change to the apko → digest → attest path as a potential chart-policy change.

## Multi-arch model

apko does not emulate. It assembles a 2-arch index from per-arch apks that **melange already
built natively** (amd64 on `ubuntu-24.04`, arm64 on `ubuntu-24.04-arm`, no QEMU). The `archs:`
in `apko.yaml` and `--arch amd64,arm64` must line up with the apks present under `packages/`
(Wolfi arch dirs `x86_64`/`aarch64`). If an arch's apk is missing, publish fails. The
single-arch `apko build` paths (`dev/image-build.sh`, `release-dry-run.yml`,
`security-scan.yml`) only assemble one arch for a quick load/smoke/scan; they are not the
multi-arch artifact.

## Read-only inspection

Use these to reason about the image without building it:

```bash
apko show-config apko.yaml      # the fully-derived config apko will act on
apko show-packages apko.yaml    # exact packages + versions that would install
```

`show-packages` resolves the live Wolfi index, so it shows what the floating base would pull
right now — the right tool to preview a CVE bump or confirm a new package resolves. Both accept
`--keyring-append`/`--repository-append` if you need them to see the `@local` apk.

## Why `apko lock` is deliberately unused

`apko lock` exists (it writes a `.lock.json` of pinned package versions) but the repo does not
use it, on purpose:

- The app package is a per-build `@local` apk; a committed lock would pin a stale/foreign
  checksum for it.
- The Wolfi base (`ca-certificates-bundle`, `tzdata`, …) is meant to float to latest for a
  fresh CA bundle/timezones and low CVE surface. Pinning fights that model.
- Reproducibility comes from recording the exact resolved versions in the per-build SBOM +
  provenance attestation, not from a lockfile. Do not add `apko.lock.json`.

## Gotchas

- Single-arch `apko build` Docker-loads under an **arch-suffixed tag**; retag before using it
  (e2e, dry-run, scan). The suffix is the value passed to `--arch`; the repo passes the Go
  arch name, e.g. `-amd64`, not the Wolfi `-x86_64`.
- `--sbom-path` directory must already exist (`mkdir -p sbom`).
- `--keyring-append` is mandatory for the `@local` apk and must cover **every** arch being
  published.
- `apko publish` needs a prior `docker login`; `apko build` (tarball) does not.
- The release digest is the **registry index digest** from `imagetools inspect`, not apko
  stdout. Do not switch the release to parse apko output — it would break the digest the
  cosign/attest/Kyverno chain depends on.
- `org.opencontainers.image.version` in `apko.yaml` is release-please-owned — never hand-edit.
- Wolfi has no `nonroot` package; the uid/gid 65532 user is created via `accounts` in
  `apko.yaml`. Keep it in sync with the chart's `runAsUser: 65532`.
- Local build artifacts (`packages/`, `sbom/`, `image.tar`, `*.oci`, `*.spdx.json`,
  `melange*.rsa*`) are git-ignored; `apko.yaml` and `melange.yaml` ARE committed.
- Adding a runtime dep means a Wolfi package in `contents.packages`, then rebuild the apk and
  the image. There is no Dockerfile to edit.
- apko is pinned by mise; invoke it via `mise exec -- apko …` (or a mise task / shimmed PATH),
  not a system install. Bumping apko is a `mise.toml` + `mise lock` change — see the `mise`
  skill.

See [references/apko-commands.md](references/apko-commands.md) for the version-stamped command
and flag reference.
