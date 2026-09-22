# 2026-09-22 — zharness symlink class: file level closed, directory level open

- id: 2026-09-22-zharness-symlink-class-status
- created: 2026-09-22
- topic: cli/src/installer, write_file_atomic, symlink following, class status

## Fact

The file-level symlink class is **closed** as of `557d69d` (the cli-rust-rewrite follow-ups, phase `p2-symlink-hardening`): `write_file_atomic` (`cli/src/installer/mod.rs:122`) unlinks a stale temp entry — `unlink` removes the link itself, never its target — and opens it with `create_new`, and the three non-test write sites route through it: the temp path (`cli/src/installer/mod.rs:120`), the uninstall restore of a captured pre-install original (`cli/src/installer/uninstall.rs:145`) and the AGENTS.md creation branch (`cli/src/installer/mod.rs:517`). Three `refuses_symlink` regression tests pin them and all three fail against the pre-fix code; a crate-wide search finds no fourth non-test write primitive.

The directory-component variant is **open** and recorded in ADR 0013: a symlink planted at a directory component of a managed path is still followed, because `write_file_atomic`'s `fs::create_dir_all` accepts an existing symlink-to-directory as a satisfied component. Reproduced at `45cfcbe` — with `docs` a symlink to a directory outside the repository, `zharness install --root <repo>` exits 0 and writes `PROJECT.md`, `WORKFLOW.md` and `playbooks/` into that outside directory. Inherited from Go v0.23.1, not a port regression; no guard covers it.

## Consequence

Do not re-derive the file-level class as open — it is fixed and tested. The open work is the directory-component case; its fix direction is in ADR 0013 (resolve each managed path's parent components and refuse or re-anchor when one is a symlink, instead of relying on `create_dir_all`).

## Source

- `docs/decisions/0013-installer-follows-directory-symlinks.md`
- `cli/src/installer/mod.rs`, `cli/src/installer/uninstall.rs` — the three sites and their `refuses_symlink` tests
- `docs/plans/completed/cli-rust-rewrite-followups.md`, Validation entry 2026-09-22T09:33:51Z, request 1
