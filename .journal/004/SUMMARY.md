---
id: 004
title: Manager CLI and startup refactor
date: 2026-05-19
status: complete
repos_touched: [template-k8s]
related_sessions: [001, 002]
---

## Goal
The session started with research into modern Kubebuilder/controller-runtime operator patterns, then narrowed into improving this template's generated manager startup surface. The concrete goal became replacing the generated `flag`/zap setup with Kong plus `slog`, and making `cmd/main.go` readable enough for future operator templates.

## Outcome
The goal was met. PR #11 was reviewed, squash-merged, local `master` was fast-forwarded to `ca321f9`, and the feature worktree plus remote branch were removed.

## Key Decisions
- Use Kong for manager options -> the manager has a typed, testable CLI surface instead of ad hoc standard-library flag registration.
- Use `slog` through `logr.FromSlogHandler` -> controller-runtime still receives a `logr.Logger`, while the template avoids carrying zap-specific runtime flags.
- Drop `--zap-*` compatibility -> no deployed manifest used those flags, and the template benefits from a clean `--log-format` / `--log-level` surface.
- Split startup by responsibility -> `main()` now reads as parse options, configure logging, create manager, register controllers, register health checks, start manager.
- Keep Kubebuilder scaffold markers together with scheme and controller registration -> future Kubebuilder commands still have predictable patch points.

## Changes
- `cmd/options.go` - added Kong-backed manager options and `slog` logger construction with JSON logs as the default.
- `cmd/main.go` - reduced startup to the high-level execution outline.
- `cmd/manager.go` - moved TLS, webhook, metrics, and manager option construction into focused helpers.
- `cmd/setup.go` - grouped scheme registration, controller registration, health checks, start logic, and Kubebuilder scaffold markers.
- `cmd/options_test.go` and `cmd/manager_test.go` - covered option defaults, current manifest args, zap-flag rejection, logger construction, TLS behavior, metrics auth/certs, webhook certs, and manager option wiring.
- `moon.yml` - changed manager build/run tasks to target `./cmd` now that the command is split across multiple files.
- `go.mod` / `go.sum` - added direct Kong, logr, and Testify dependencies required by the refactor and tests.

## Open Threads
- The operator-practice source survey in `NOTES.md` has not yet been converted into template policy or implementation slices.
- Future operator sessions can still evaluate field indexes, status-only predicates, partial metadata watches, class/shard selection, and stronger status helper patterns.

## References
- PR: https://github.com/meigma/template-k8s/pull/11
- Merge commit: `ca321f9273f6d6713f3bb2673ff3fb5631940648` (`refactor(cli): use kong and slog for manager startup`)
- Prior operator prototype session: `.journal/001/SUMMARY.md`
- Prior Chainsaw/e2e session: `.journal/002/SUMMARY.md`

## Lessons
- Once a Go command package grows beyond one file, Moon build/run tasks must target the package path (`./cmd`) rather than a single `cmd/main.go` file.
- Kubebuilder's generated `main.go` can be made much easier to follow without hiding controller-runtime details behind a new package boundary.
