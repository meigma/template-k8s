---
id: 014
title: Moon task surface slimming
date: 2026-05-20
status: complete
repos_touched: [template-k8s]
related_sessions: [013]
---

## Goal
Critically examine `moon.yml` for recipe sprawl, remove low-value commands, and
leave a smaller Moon task surface that avoids recipe fatigue while preserving
the real operator validation paths.

## Outcome
The goal was met. PR #33 was squash-merged as
`1620d08df77083058ebe68da48e34f4967b81a54`, local `master` was
fast-forwarded, the `feat/moon-task-slimming` worktree and branch were removed,
and the remote feature branch was deleted.

## Key Decisions
- Collapse static validation into `root:check` -> the previous separate
  `fmt`, `lint`, `lint-config`, `generated-check`, `chart-validate`, and
  `chainsaw-lint` commands created recipe fatigue without representing separate
  maintainer workflows.
- Keep `root:test` separate -> envtest-backed Go tests have a distinct runtime
  dependency on `setup-envtest` and should stay independently runnable.
- Keep only durable maintainer-facing tasks in Moon -> the final exposed task
  list is `check`, `test`, `generate`, `deploy`, `undeploy`, `dev-up`,
  `dev-down`, and `test-e2e`.
- Remove `dev-stack-smoke` -> the user did not see enough value in a separate
  destructive dev-stack proof once the normal e2e path already validates the
  packaged operator flow.
- Move longer command bodies to scripts -> this kept `moon.yml` readable while
  preserving the existing deploy, undeploy, check, and e2e behavior.

## Changes
- `moon.yml` - reduced the root task list from 24 tasks to 8, removed narrow
  recipes, made `root:check` the static validation gate, and kept `root:test`
  independent.
- `scripts/check.sh` - added the consolidated static check script covering Go
  formatting, golangci-lint, generated artifact drift, Helm chart validation,
  and Chainsaw manifest linting.
- `scripts/deploy.sh`, `scripts/undeploy.sh`, and `scripts/test-e2e.sh` -
  moved longer operational shell from Moon tasks into focused scripts.
- `dev/stack-smoke.sh` - removed the redundant local dev stack smoke test.
- `.github/workflows/ci.yml` - replaced `moon ci`, generated-check, and chart
  validation steps with explicit `root:check` and `root:test` runs.
- `AGENTS.md`, `README.md`, and `DELETE_ME.md` - updated the documented command
  surface and added agent guidance discouraging future Moon recipe sprawl.
- `Tiltfile` and `test/chainsaw/nginx-smoke/chainsaw-test.yaml` - updated
  stale task references after removing `dev-cluster-*` and `uninstall`.

## Open Threads
- None for this slice. Future task additions should satisfy the new
  maintainer-facing workflow rule in `AGENTS.md`.

## Lessons
- A single-project operator template should avoid exposing implementation
  details as top-level recipes. Make the common path obvious and push longer
  mechanics into scripts when they still need to be versioned.
- The local Moon CLI is now past the earlier `2.0.0-rc.1` script-task parser
  issue; script tasks were visible through `moon tasks root` during this
  session.

## References
- PR #33: https://github.com/meigma/template-k8s/pull/33
- Merge commit: `1620d08df77083058ebe68da48e34f4967b81a54`
- Local branch/worktree removed: `feat/moon-task-slimming`
- Prior local dev stack session: `.journal/013/SUMMARY.md`
