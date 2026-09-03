# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

## [v0.17.0] — 2026-09-03

### Fixed

- `zharness update`: a playbook or `docs/WORKFLOW.md` with no recorded
  merge base — `.zharness/base/` lost, or predating base tracking — used to
  stay stuck on stale content forever (R18 refused to invent an ancestor).
  These files are pure upstream mirrors, never a legitimate consumer
  customization surface, so `install`/`update` now fresh-overwrite them
  unconditionally: no diff, no compare, no conflict. Three-way merge stays
  for `docs/PROJECT.md` and the marked `AGENTS.md` block, the two managed
  files consumers do own. See `docs/decisions/0007-fresh-overwrite-for-playbooks-and-workflow.md`.

### Changed

- Hand-edits to a playbook or `WORKFLOW.md` are now silently discarded on
  the next `zharness update`, with no conflict marker warning first — the
  accepted cost of the fix above.

## [v0.16.3] — 2026-09-02

### Changed

- `check.md` gate: if `scripts/record-check.sh` exists, run proofs through
  it (timeout → gtimeout → unbounded; 3-line pass / 10-line fail tail).
  If it is absent, capture to a temp file, print the same tails, and keep
  the command's exit code. Nested Validation bullets still cite the raw
  commands so the hook re-executes them.

### Fixed

- `scripts/record-check.sh`: same timeout resolver as the pre-commit hook,
  so stock macOS no longer fails with `timeout: command not found`.

### Added

- `docs/audit/wave-session-ab-protocol.md`: paired worktree A/B for
  same-session vs wave-boundary restart. Does not change `work.md` step 11.

## [v0.16.2] — 2026-08-30

### Changed

- Handoff close: after absorb and `git mv`, a completed plan is a run log
  (cite the ADR or guard, never the completed path). It may later be deleted
  when no recovery audience remains; if unsure, keep. Deletion is not part
  of close.

## [v0.16.1] — 2026-08-30

### Added

- `escalate_when` on `work.md` and `WORKFLOW.md`: ask the owner and stop
  when locked schema or requirements would change, the same verification
  command failed twice, or a product rule conflicts. Retry cap stays one
  targeted fix. Not a plan field. Not a hook.

## [v0.16.0] — 2026-08-30

### Added

- Handoff absorb gate: an initiative cannot `git mv` to `completed/` without
  an `absorb:` line in `## Decisions` (`absorb: none` is valid). A
  class-of-failure or expensive-to-reverse decision must already live in an
  ADR or a native guard.
- Pre-commit R5: at most one non-empty file under `docs/plans/active/`; zero
  is a valid idle state.
- Pre-commit H3: a newly added `mode: full` Validation entry that also
  declares `judge: same-session` is rejected. Full checks require an
  independent judge.
- Source-repo skills `encode-invariant` and `improve-harness` (not part of
  `zharness install`).
- `LICENSE` (MIT), `SECURITY.md`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`,
  and `.github/dependabot.yml`.

### Changed

- Consumer `README.md` and `AGENTS.md` rewritten around the three-verb
  installer. Workflow diagrams no longer mention `zharness init` or a local
  database.
- Live contract and architecture describe the v0.16 protocol. The binary
  surface is unchanged from v0.15: `install` / `update` / `uninstall`.

### Fixed

- Guard entry hashes no longer depend on trailing blank lines, which had
  re-executed the previous Validation entry on every append.
- Skill validation actually runs on `skills/<category>/<name>/SKILL.md`.
  Load-time errors stay fail-closed; format checks are warnings.

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
- Pre-commit guard: the old-side entry set no longer uses a bash 4
  associative array. The hook's shebang is `#!/bin/bash`, which on macOS is
  bash 3.2: `local -A` failed there, the hex hash was then evaluated as an
  arithmetic array subscript, and the shell died with "value too great for
  base" on the first old-side entry — so every commit touching a plan that
  already had a Validation entry was rejected with an opaque error. Membership
  is now a hash file plus `grep -Fxq`, identical on every bash.
- Guard fixture suite: the decisive accept and reject cases are re-run under a
  legacy bash 3.x when one is present, so the guard core cannot regress into a
  bash-4-only construct.

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

