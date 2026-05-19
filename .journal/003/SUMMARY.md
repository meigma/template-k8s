---
id: 003
title: Gate release dry runs to release PRs
date: 2026-05-19
status: complete
repos_touched: [template-k8s]
related_sessions: []
---

## Goal
Port the latest merged `../template-go` workflow change into `template-k8s`.
The source commit was `d421cc5` (`ci: gate release dry runs to release-please PRs (#13)`).

## Outcome
The goal was met. PR #10 was reviewed, squash-merged, local `master` was fast-forwarded to `49f7395`, and the session worktree plus remote feature branch were removed.

## Key Decisions
- Port only the release dry-run gating change -> the rest of the `template-go` workflow diff was repo-specific binary/container smoke behavior that does not apply cleanly to `template-k8s`.
- Keep the broad `pull_request` trigger -> required check contexts still appear on ordinary PRs while the expensive release rehearsal jobs skip unless the run is manual or the branch is a Release Please branch.

## Changes
- `.github/workflows/release-dry-run.yml` - added explanatory comments and job-level guards for the binary and container dry-run jobs.
- `.journal/TECH_NOTES.md` - recorded the durable Release Dry Run gating rule for future workflow edits.

## Open Threads
- None.

## References
- PR: https://github.com/meigma/template-k8s/pull/10
- Merged commit: `49f7395962765b50363994da33f2cf38dd81477b`
- Source template-go commit: `d421cc5ead9e227698da5a7d48358ad17e25d4aa`
