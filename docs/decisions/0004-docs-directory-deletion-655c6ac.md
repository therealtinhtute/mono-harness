# 0004 — Recovery position after commit `655c6ac` deleted `docs/`

## Status

Accepted. Restoration executed in phase `p1-doc-authority` of `docs/plans/completed/docs-architecture.md`.

## Context

Commit `655c6ac`, "Delete docs directory" (2026-08-16), removed all 26 files under `docs/` — 4,285 lines. It was authored through the GitHub web UI, which runs no local gate, so `scripts/verify-doc-links.sh` never saw it.

The next `zharness init` regenerated exactly the 8 files present in the CLI's embedded doc set: `docs/WORKFLOW.md` and the 7 files under `docs/playbooks/`. The other 18 files, roughly 3,733 lines, had no embedded source and were not regenerated. The correspondence is exact, with no exception.

That is what made the loss invisible. The managed half of `docs/` reappeared on its own, so the directory looked healthy, while the authored half stayed gone. It surfaced only as 16 broken cross-references reported by the repository's own link gate — which meant the gate was red on every commit until this initiative.

## Decision

Restore only what is required. Mark the rest, and record where it lives.

**Restored — 2 files**, both because something in the repository issues an unexecutable instruction without them:

| File | Lines | Why required |
|---|---|---|
| `docs/prompt-engineering-principles.md` | 143 | `CLAUDE.md` orders it read before any skill or rule edit |
| `docs/workflow-harness/migration.md` | 52 | the sole pointer for adopting the harness on a legacy project |

Both were corrected on restore rather than restored verbatim. `migration.md` in particular described a changeset layer that no longer exists (see [0001](0001-markdown-as-source-of-truth.md)), a `.kit/docs/` scaffold target that moved to `docs/`, and a repository that has since been renamed.

**Staying deleted — 8 cited files.** Each is cited only as provenance for rationale that is already written inline at the citation site, so the citation loses a footnote, not a meaning:

| File | Lines |
|---|---|
| docs/audit/workflow-harness-ceremony-audit.md | 399 |
| docs/audit/sdlc-token-cache-audit.md | 212 |
| docs/audit/sdlc-gap-analysis.md | 126 |
| docs/plans/completed/harness-memory-ceremony-convergence.md | 885 |
| docs/plans/completed/workflow-harness-history-2026-07.md | 314 |
| docs/plans/completed/harness-convergence-pass-v3.md | 247 |
| docs/workflow-harness/pilot-evidence/2026-07-17-lab-skills-import.md | 66 |
| docs/evals/failures.md | 12 |

Eight further files were deleted and are cited by nothing: the three under docs/audit/cost-model/, docs/plans/completed/eval-layer.md, docs/plans/completed/sdlc-token-optimization.md, docs/workflow-harness/gap-matrix.md, and two more pilot-evidence logs from 2026-07-19. They are left deleted without further comment.

Citations to the 8 cited files are handled in two places. Three of them are cited from text that lives in `cli/docs/embedded/playbooks/`, so the citation cannot be removed without a binary change and a release — those are marked in `.claimignore`, each entry naming `655c6ac` and deferring removal to phase `p2-consumer-scaffold`. The rest are cited from `skills/workflow/README.md`, which is authored and freely editable, and are retargeted at this ADR.

## Consequences

- Nothing is lost. Every one of the 26 files remains retrievable with `git checkout 655c6ac^ -- <path>`.
- `.claimignore` gains three entries that are honest markers of deferred work, not silencers of real breakage. They must be removed when `p2-consumer-scaffold` edits the embedded playbooks.
- `.claimignore` matches by substring across the whole repository, so the entry for the ceremony audit also hides that same citation in `skills/workflow/README.md`. The retarget there is still required by hand and is verified by a direct grep rather than by the link gate.
- The underlying gap is unfixed: a deletion made through the GitHub web UI still bypasses every local gate. This ADR records the incident; it does not prevent the next one.

## Authority

- Commit `655c6ac` — `git show --stat 655c6ac` lists all 26 files and 4,285 deletions.
- `scripts/verify-doc-links.sh` — the gate that reported the 16 broken references.
- `docs/plans/completed/docs-architecture.md` — phase `p1-doc-authority`, requirements R10, R11, R12.
