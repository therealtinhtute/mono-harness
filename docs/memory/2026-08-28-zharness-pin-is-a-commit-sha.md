# 2026-08-28 — zharness:pin is a git commit sha, not a content hash

- id: 2026-08-28-zharness-pin-is-a-commit-sha
- created: 2026-08-28
- topic: docs/ARCHITECTURE.md pin, drift audit, repinning

## Fact

The `<!-- zharness:pin <sha> -->` declaration at the top of `docs/ARCHITECTURE.md` is a **git commit sha**, verified with `git rev-parse --verify <sha>^{commit}` — not a hash of the document body. The drift audit (deleted with the v0.15 CLI in commit 4fa4103; last lived at `cli/internal/application/audit.go` in history — see `pinResolves`, `measureCitation`, `pinDriftFinding`) compared each `path:line` citation extracted from the pinned doc against git history since the pin commit, reporting lines added/removed per citation.

## Consequence

"Repin" means: set the pin to the current HEAD (or any commit at/after the last change to every cited path). Editing the doc without touching cited files does NOT create drift; changing a cited file after the pin does.

## Source

- `git show 4fc8481:cli/internal/application/audit.go` (pre-deletion source)
- `docs/plans/completed/consumer-doc-drift-gate.md` (design record)
