---
id: 005
title: Operator guidance and repo-local tooling refresh
date: 2026-05-19
status: complete
repos_touched: [template-k8s]
related_sessions: [001, 002, 003, 004]
---

## Goal
Refresh the repository's agent-facing guidance and repo-local workflow assets after researching mature Kubernetes operator practices, while keeping the repo suitable as a copied starting point for real operator work.

## Outcome
The goal was met and merged via PR #12. The repository now tracks repo-local skills, has updated agent guidance, documents additional Kubebuilder/operator best practices, removes stale template artifacts, removes the unused `ghd` package config, and simplifies linting to use the normal PATH-resolved `golangci-lint` binary.

## Key Decisions
- Document mature-operator patterns in the repo-local `k8s-operator` skill -> agents need actionable examples without turning AGENTS.md into a large reference manual.
- Keep `AGENTS.md` repo-specific and avoid calling the repo a template -> the file is copied into downstream repos and should read as local guidance.
- Track `.agents/` and `.claude/skills` -> repo-local skills are part of the operating contract and must travel with the repository.
- Remove custom `golangci-lint` module-plugin wiring -> building a custom linter binary was too much CI/tooling friction for one Kubernetes logging check.
- Use PATH-resolved `golangci-lint` in Moon tasks -> `.envrc` and GitHub CI already activate/install repo-pinned tools, so task commands should call the tool directly.

## Changes
- `.agents/skills/k8s-operator/SKILL.md` - added best-practice guidance with examples for status patch helpers, field indexes, predicates, metadata-only watches, typed manager config, controller classes, desired object builders, pruning, and webhooks.
- `.agents/skills/{git,worktrunk,session-*}` - restored repo-local workflow skills from the main checkout so agents can operate without relying on external skill state.
- `.claude/skills` - added the relative symlink to the tracked repo-local skills.
- `.gitignore` - stopped ignoring `.agents/` and `.claude/` so skills are committed.
- `AGENTS.md` - replaced old generated guidance with concise repo-specific guidance covering Moon, Kubebuilder/controller-runtime, Kong, `slog`, envtest, and Chainsaw.
- `AGENTS.md.old` and `Makefile.old` - removed stale backup/generated files.
- `ghd.toml` - removed package metadata because operators are not expected to be distributed through `ghd`.
- `.custom-gcl.yml`, `.golangci.yml`, and `moon.yml` - removed custom `golangci-lint` binary generation, replaced plugin `logcheck` with built-in `loggercheck`, and made Moon lint tasks call `golangci-lint` directly from PATH.

## Open Threads
- The `template-go` golangci-lint config has not yet been fully ported. Future work should merge its stricter rules carefully, with explicit Kubebuilder exceptions for scheme globals, generated-style init registration, Ginkgo tests, and Kubernetes map types.
- PR #12 had a local cleanup wrinkle: `gh pr merge --squash --delete-branch` merged remotely but failed while trying to switch to the `master` worktree. The remote branch was deleted manually and the local worktree was removed afterward.

## References
- PR #12: https://github.com/meigma/template-k8s/pull/12
- `.journal/001/SUMMARY.md`
- `.journal/002/SUMMARY.md`
- `.journal/003/SUMMARY.md`
- `.journal/004/SUMMARY.md`
