---
id: 013
title: Local operator development stack
date: 2026-05-20
status: complete
repos_touched: [template-k8s]
related_sessions: [009]
---

## Goal
Establish and prove the local development flow that downstream operator
repositories should inherit from this template. The target workflow was a fast
Kind feedback loop where `ctlptl` owns cluster/registry lifecycle, Tilt owns
watch/build/deploy orchestration, and `ko` builds the manager image.

## Outcome
The goal was met. PR #32 was squash-merged as
`e38b98899bdba357ad02e17b754bc2d326b78dd6`, local `master` was
fast-forwarded, and the `feat/tilt-dev-flow` worktree was removed.

## Key Decisions
- Keep cluster lifecycle outside the `Tiltfile` -> `ctlptl` makes the Kind
  cluster and local registry reproducible while Tilt stays focused on the dev
  loop.
- Use a local registry instead of `kind load` -> Tilt can push immutable image
  refs and avoid slow image-load loops when rebuilding the controller.
- Render the existing Helm chart in Tilt -> local development exercises the
  same packaged Deployment/RBAC/CRD shape downstream users will rely on.
- Build `./cmd` with `ko` through a small Tilt `custom_build` wrapper -> the
  dev loop uses Go-native image builds while still letting Tilt inject the
  expected image reference into Helm-rendered YAML.
- Document the flow in `AGENTS.md` -> future agents need explicit ownership
  boundaries for `ctlptl`, Tilt, and `ko`, not just task names.

## Changes
- `.prototools` and `.moon/proto/{tilt,ko,kind,ctlptl}.toml` - added
  repo-managed local-dev tools.
- `dev/ctlptl.yaml` - added the `template-k8s-dev` Kind cluster and
  `template-k8s-registry` local registry on `127.0.0.1:5005`.
- `Tiltfile`, `.tiltignore`, `.ko.yaml`, and `dev/ko-build.sh` - added the
  Tilt/ko manager image build and Helm-rendered deploy loop.
- `dev/stack-smoke.sh` and `moon.yml` - added Moon tasks for cluster
  up/down, Tilt up/down, and a full non-interactive smoke proof.
- `README.md`, `DELETE_ME.md`, and `AGENTS.md` - documented the downstream
  development workflow, generated-repo cleanup items, and agent-facing local
  stack instructions.

## Open Threads
- Downstream repositories must rename the dev cluster, registry, image
  selector, chart path, namespace, and sample CR smoke assertions when they
  copy from this template.
- The local stack currently assumes Docker or a compatible container runtime is
  available and healthy.

## Lessons
- `tilt ci` works well for proving this dev stack when the cluster and registry
  are created first by `ctlptl`; it can build through `ko`, inject the pushed
  image into Helm output, wait for the controller Deployment, and then allow a
  script to prove the sample CR and service.
- Port-forward smoke checks should bind and curl `localhost` explicitly; using
  mixed loopback addresses can make the HTTP proof fail even when the service
  is healthy.

## References
- PR #32: https://github.com/meigma/template-k8s/pull/32
- Merge commit: `e38b98899bdba357ad02e17b754bc2d326b78dd6`
- Local branch/worktree removed: `feat/tilt-dev-flow`
- Prior downstream template guidance: `.journal/009/SUMMARY.md`
