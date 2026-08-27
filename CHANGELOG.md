# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [v0.15] — UNRELEASED (breaking)

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
