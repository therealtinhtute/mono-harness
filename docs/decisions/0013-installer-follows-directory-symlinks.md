# 0013 — The installer follows symlinks at directory components

## Status

Accepted. 2026-09-22. Authority: the owner's call of 2026-09-22, recorded in the `full` review of the cli-rust-rewrite follow-ups.

## Context

The v0.24 follow-up initiative closed the file-level symlink class. `write_file_atomic` (`cli/src/installer/mod.rs:122`) now unlinks a stale temp entry — `unlink` removes the link itself, never its target — and opens it with `create_new`, and the three non-test write sites route through it, each pinned by a `refuses_symlink` regression test.

The complete review then reproduced a variant that fix does not cover: a symlink planted at a **directory component** of a managed path is still followed. Observed at `45cfcbe`: with `docs` a symlink to a directory outside the repository, `zharness install --root <repo>` exits 0 and writes `PROJECT.md`, `WORKFLOW.md` and `playbooks/` into that outside directory, leaving `docs` a symlink. `write_file_atomic` calls `fs::create_dir_all(dir)` before it touches the temp path, and `create_dir_all` accepts an existing symlink-to-directory as a satisfied component.

The behavior is inherited from Go v0.23.1, not a regression of the Rust port or of the follow-up fix.

## Decision

Defer the fix to its own initiative. The escape is pre-existing, it is not one of the three write sites the follow-up's R3 named, and closing it changes the installer's path policy — a decision that deserves its own surfaces, its own acceptance and its own gate rather than a drive-by edit inside a closure. Rejected: fixing it inside the follow-up initiative, because that phase's surfaces and acceptance were already locked and gated, and widening them after the fact would have made its verdict describe a tree it never saw.

The fix direction, when it is taken up: resolve each managed path's parent components and refuse — or re-anchor — when a component is a symlink, instead of relying on `create_dir_all`'s acceptance of one.

## Consequences

Easy: nothing changes for a repository whose managed directories are real directories, which is every repository the installer has been run against. Hard: the escape stays open until a later initiative closes it, and nothing guards it — the three `refuses_symlink` tests pin the file-level sites only, so a regression or a new directory-level site would be silent.

## Authority

- Reproduced 2026-09-22 at `45cfcbe` in a scratch repository whose `docs` is a symlink to a directory outside it: `zharness install --root <repo>` exits 0, and the outside directory receives `PROJECT.md`, `WORKFLOW.md` and `playbooks/`.
- `cli/src/installer/mod.rs:122` — `write_file_atomic`, its `create_dir_all` call and the `create_new` temp open.
