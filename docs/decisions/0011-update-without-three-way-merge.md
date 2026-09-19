# 0011 — `update` drops three-way merge: hash-guarded AGENTS block, write-once PROJECT.md

## Status

Accepted. 2026-09-19. Supersedes the `Merge: true` half of
[ADR 0007](0007-fresh-overwrite-for-playbooks-and-workflow.md) and decision 5
(stash transaction) of [ADR 0008](0008-recorded-ownership-and-transactional-recovery.md).
ADR 0008 decisions 1–4 (ownership ledger, `unknown` means keep, legacy seeding,
honest uninstall summary) are unchanged. Authority: R4–R7 of
`docs/plans/active/zharness-slim.md`.

## Context

After ADR 0007, three-way merge protected exactly two surfaces: `docs/PROJECT.md`
and the `AGENTS.md` marked block. Keeping it cost a diff3 implementation, an
upstream blob store under `.zharness/base/upstream/`, an update stash with its
own on-disk format and recovery rules, a conflict file, and the
`--continue`/`--abort` verbs — the largest share of `cli/internal/installer/`.

Neither surface needs a merge:

- `docs/PROJECT.md` is scaffolded once and then owned by the project. Upstream
  changes to its template are new questions, not edits to the project's answers.
  They are surfaced as missing headings by `zharness update --check`.
- The `AGENTS.md` block is harness text between markers; the project's own prose
  lives outside them. A hand edit inside the block is the only case that needs
  protection, and detecting it needs a hash, not an ancestor.

## Decision

1. **PROJECT.md is write-once.** `install` and `update` write it only when absent
   and never read, merge, or overwrite an existing one
   (`cli/internal/installer/installer.go:77`, `cli/internal/installer/update.go:103`).
2. **The AGENTS block is replaced between its markers, guarded by a hash.** The
   manifest already records the sha256 of the canonical block last written. Before
   any write, `update` compares the on-disk block with that hash. A mismatch that is
   not already the new block means a hand edit: `update` prints a unified diff,
   writes nothing, and exits non-zero. `--force` replaces the block anyway. A
   repository with no recorded hash is accepted and recorded. Prose outside the
   markers is never touched (`cli/internal/installer/update.go:88`,
   `cli/internal/installer/update.go:156`).
3. **All refusal checks run before the first write**, so a refused `update`
   leaves every file as it was. There is no stash to restore
   (`cli/internal/installer/update.go:88` precedes the first write at `update.go:109`).
4. **Removed:** three-way merge, `.zharness/base/upstream/` (deleted on the next
   `update`), the update stash, `.zharness/conflicts.json`, and
   `--continue`/`--abort`.
5. **The manifest's per-file sha256 replaces the blob store as the recorded base**
   for ADR 0008 uninstall and legacy seeding. Both only ever asked "are these bytes
   exactly what the installer last wrote?", which a hash answers
   (`cli/internal/installer/installer.go:196`, `cli/internal/installer/installer.go:215`).

## Consequences

- A project that hand-edited `docs/PROJECT.md` no longer receives template
  changes as merged text; `update --check` names the headings it lacks.
- A hand edit inside the AGENTS block stops `update` until the owner either moves
  the edit outside the markers or passes `--force`.
- Downgrading to a pre-0011 binary after an `update` finds no upstream blobs. The
  supported path is `zharness uninstall` with the new binary, then `install` with
  the old one; the ownership ledger keeps that uninstall safe.
- A conflict list left by an interrupted pre-0011 update makes `update` refuse
  and `update --check` report drift until the owner resolves the markers and
  deletes it and the stash (`cli/internal/installer/update.go:70`,
  `cli/internal/installer/check.go:24`); `uninstall` removes both
  (`cli/internal/installer/uninstall.go:48`).
- `install` rewrites the block without the hash guard: rerunning it is the
  explicit reset, and the guard protects only `update`.
- A block checked out with CRLF line endings compares as LF, so `autocrlf` is
  not mistaken for a hand edit.

## Authority

- `docs/plans/active/zharness-slim.md` — R4, R5, R6, R7; owner approval 2026-09-19.
- `README.md` — *Safe to adopt and to leave*.
- `docs/decisions/0007-fresh-overwrite-for-playbooks-and-workflow.md`,
  `docs/decisions/0008-recorded-ownership-and-transactional-recovery.md` — the
  decisions narrowed here.
