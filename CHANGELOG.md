# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [v0.15.1] — 2026-08-28

### Fixed

- Pre-commit guard: the verdict token is read from a Validation entry's first
  line and now matches the repository's `verdict \`APPROVED\`` grammar —
  previously every real APPROVED entry was silently skipped by the guard.
- Pre-commit guard: Validation entries without a leading timestamp are now
  visible to both fail-closed guards.
- Installer: `.zharness/base/original/` naming is collision-free (with a
  legacy fallback so v0.15.0-captured originals still restore), and diff3
  falls back to a conservative whole-side hunk above an 8M-cell LCS cap.
- Non-spine `git`/`interview` skills no longer reference binary commands
  deleted in v0.15.
- Pre-commit guard: proof re-execution no longer depends on GNU `timeout`
  being installed. The wrapper is resolved at call time (`timeout`, else
  `gtimeout`, else the command runs unbounded with a warning), so a proof's
  verdict follows its own exit code. On stock macOS every proof previously
  exited 127 (`timeout: command not found`) and the guard rejected each
  honest APPROVED entry.
- Guard fixture suite: the entry-count assertion compares numerically, so
  BSD `wc -l` padding can no longer produce a false FAIL on macOS.

## [v0.15.0] — 2026-08-28 (breaking)

### Changed

- **BREAKING**: the entire `zharness` lifecycle command surface (init,
  migrate, import, db rebuild/status, query views, intake, story, trace add,
  decision add, memory, run create, plan complete/abandon, resume, preflight,
  check record, handoff record, validate, audit) and SQLite are deleted from
  source. The binary is being reduced to install / update / uninstall.
- State lives in git-committed markdown alone (`docs/plans/**`). Lifecycle
  bookkeeping is hand-appended markdown gated by the repository's pre-commit
  hook, which now owns both fail-closed guarantees: proof re-execution for
  APPROVED Validation entries and the independent-judge rule for high-risk
  lanes (`scripts/install-git-hooks.sh`; CI re-runs the same guards).

### Removed

- The SQLite database (`harness.db`) is no longer created, read, or written
  by the binary.

### Upgrade guidance

- **Pin `v0.14.x`** if you still need the old lifecycle binary — that line
  keeps a fully working product with its own `harness.db`.
- Your existing `harness.db` (and its `-wal`/`-shm` sidecars) is
  **consumer-owned bytes**: nothing in this or future releases deletes it.
- New consumers need no initialization at all: markdown plus git is the whole
  system of record.

## [Unreleased]

### Added

- `LICENSE` (MIT), matching the license already declared in `README.md`.
- `SECURITY.md` with vulnerability reporting instructions.
- `CONTRIBUTING.md` with setup and pre-commit gate instructions.
- `CODE_OF_CONDUCT.md` (Contributor Covenant v2.1).
- `.github/dependabot.yml` for Go modules and GitHub Actions.
