# 2026-09-22 — zharness installer temp-file writes follow planted symlinks

- id: 2026-09-22-zharness-atomic-write-symlink
- created: 2026-09-22
- topic: cli/src/installer, write_file_atomic, symlink-following temp path, deferred defect

## Fact

`write_file_atomic` (`cli/src/installer/mod.rs:120`) writes to the predictable sibling path `<path>.tmp-zharness` with `fs::write`, which follows a symlink planted at that path. A consumer repository can carry such a symlink, so `zharness install` can overwrite an arbitrary file the user can write. Two sibling write sites share the class: `cli/src/installer/uninstall.rs:143` (`fs::write(&dst_p, o)` restoring a captured pre-install original through a symlink at the managed path) and one more site named in the 2026-09-22T06:02:22Z full-check entry of the cli-rust-rewrite plan. The behavior is inherited verbatim from Go v0.23.1 (`cli/internal/installer/installer.go:241-245`), so it is a latent defect, not a port regression; an independent judge reproduced both sites.

## Consequence

The fix is `OpenOptions::new().create_new(true)` on the temp path (or `O_NOFOLLOW`) at all three sites. It was deliberately deferred: p4-release-proof's surface forbids source changes, so it needs a new initiative, and it must land before the final non-rc `cli/v0.24.0` tag is cut. The ten review requests from the same full entry are also still open.

## Source

- `docs/plans/completed/cli-rust-rewrite.md` — Validation entry 2026-09-22T06:02:22Z (mode: full, judge: independent), requests 1-2 and prose notes
- `cli/src/installer/mod.rs`, `cli/src/installer/uninstall.rs`
