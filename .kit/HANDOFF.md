---
id: 01KY4D5F1FRHN3NPJ6NBB891NZ
type: handoff
phase: single-source-playbooks
lane: high-risk
run_id: 01KY4BCT00MYGZFJAW9JP650JE
check_id: 01KY4BSGY48T93BYNHVGNF3PED
created: 2026-07-22
updated: 2026-07-22
session-date: 2026-07-22
branch: master
status: clean
continuity-mode: full-harness
active-phase: none — Harness Subtraction Pass closed (4/4 phases done)
last-updated: 2026-07-22
---

# Session Handoff — master

## Current State

**Branch**: `master`, working tree clean, **pushed to origin/master** (HEAD `aad8de9`, in sync with `origin/master`, no ahead/behind).
**Status**: clean — **Harness Subtraction Pass is fully closed**. All 4 phases (write-boundary, dead-surface-removal, scoring-removal, single-source-playbooks) done, gated APPROVED, committed, and pushed.
**Continuity Mode**: full-harness (SPEC/ROADMAP/phase chain present; latest RUN + CHECK both resolve cleanly via `zharness resume --json`, drift empty).
**Active Work**: none. No queued next phase exists in `.kit/planning/ROADMAP.md`.
**Last Commits** (this session): `b17840f` test(embedded) drift-guard test, `049dd33` docs embed-only rule, `7ee007f` chore(kit) bank single-source-playbooks run+gate, `aad8de9` docs(kit) close out ROADMAP Next Steps — all pushed.

## What We're Building

**Harness Subtraction Pass** (`.kit/planning/ROADMAP.md`, spec locked, lane: high-risk, execution mode `work full`) — 4 linear phases, **all done**:
1. **write-boundary** — done (commit `32cb60c`).
2. **dead-surface-removal** — done (`73bc50c`/`7b4b009`).
3. **scoring-removal** — done (`2d6e2fc`).
4. **single-source-playbooks** — done this session: `.kit/docs/playbooks/*` is now a drift-tested projection of the Go embed (`cli/internal/embedded/projection_drift_test.go` → `TestProjectionDrift_KitDocsMatchesEmbed`), plus a one-line "edit the embed only" rule added to `CONTRACT.md`, `README.md`, `docs/workflow-harness/migration.md`. Gated APPROVED, committed (`b17840f`/`049dd33`/`7ee007f`), pushed. ROADMAP.md's own "Next Steps" section rewritten (`aad8de9`) to state the pass is closed instead of pointing at an already-finished phase.

No further phase is defined for this initiative — this is a genuine end state, not a pause.

## Continuity Anchors

**Latest harness RUN/CHECK**: single-source-playbooks run `01KY4BCT00MYGZFJAW9JP650JE` + check `01KY4BSGY48T93BYNHVGNF3PED` (APPROVED, 0 critical/major, 4 minor/informational findings).
**Gate report**: `.kit/reports/check/20260722-1445-single-source-playbooks.md` — Next Action said "ready for PR — suggest `git` then `handoff`"; both completed this session.
**Handoff entity**: `01KY4D5F1FRHN3NPJ6NBB891NZ` (this record), superseding prior handoff `01KY493H9GATGD9QSP0K4GJTBG`.
**Proof / Drift Notes**: `zharness resume --json` `drift` is **empty** ✓.

## Progress This Session

### Completed ✓
- **Phase 4 `single-source-playbooks`**, both waves DONE:
  - Wave 1 (T1): added `TestProjectionDrift_KitDocsMatchesEmbed` (`cli/internal/embedded/projection_drift_test.go`) — compares the embed manifest against the real, git-tracked `.kit/docs/` tree byte-for-byte. Proven both directions: PASS on clean tree, FAIL naming the exact drifted path when `.kit/docs/playbooks/check.md` was deliberately corrupted, then restored and re-verified via `diff`.
  - Wave 2 (T2): documented the "edit the embed only" rule in `cli/docs/CONTRACT.md`, `README.md`, `docs/workflow-harness/migration.md`.
  - Phase gate: `check full` → APPROVED, check id `01KY4BSGY48T93BYNHVGNF3PED`.
- **Committed and pushed** — split into 3 commits by type (`test`, `docs`, `chore(kit)`) plus one follow-up ROADMAP cleanup commit — all pushed to `origin/master`.
- **ROADMAP.md Next Steps rewritten** (`aad8de9`) to state the pass is closed, per explicit user confirmation via `AskUserQuestion`.
- **Handoff recorded** — entity `01KY4D5F1FRHN3NPJ6NBB891NZ`, this file rewritten to match.

### In Progress ⏳
- None. Clean stopping point — the initiative itself is finished.

### Not Started
- Nothing queued. No Phase 5 exists for the Harness Subtraction Pass.

## Key Decisions

1. **Drift-guard test compares against the real project's `.kit/docs/`, not a fresh scaffold** — the pre-existing `TestInit_FreshScratchDir_FullIntegration` was judged tautological (fresh-scaffold-vs-embed always matches); the meaningful guard is against the actual git-tracked copy, which is genuinely at risk of hand-edit drift.
2. **Test placed in a new file, not the plan's suggested `embedded_test.go`** — `application` already imports `embedded`; importing `application`'s `ScaffoldDocs` back into an `embedded_test.go` would cycle. `projection_drift_test.go` (same package) reads the filesystem directly instead — same spirit, no cycle.
3. **`.gitignore`/`migration.md` contradiction about `.kit/docs/` tracking status flagged, not fixed** — `.kit/docs/` has been git-tracked since `77ed8bb`, contradicting stated "should be ignored" guidance; out of Phase 4's stated scope, but it's exactly why the new test has real teeth (a tracked, hand-editable copy is a live drift vector). Left for a future small cleanup.
4. **3-way commit split** (test / docs / chore(kit)) — matches this repo's established convention of separating harness bookkeeping (`chore(kit)`) from the actual code/docs change, rather than one bundled commit.
5. **ROADMAP.md Next Steps rewrite confirmed via `AskUserQuestion` before editing** — treated as a real (if small) decision point rather than auto-applying it, consistent with the "ask before commit-adjacent doc edits" pattern this session leaned on.

## Blockers & Issues

None blocking. One pre-existing, non-blocking item carried forward (not this pass's responsibility to fix, flagged for visibility):
- **Confirmed CLI gap** (carried forward across the whole pass): no command transitions a `story` (phase) row's `status` from `planned`→`done`. `zharness query state --json`/`resume --json` still show `current_phase: write-boundary`/`planned` even though all 4 phases are done — this is expected bookkeeping drift, not a sign of unfinished work. `.kit/planning/ROADMAP.md` and the check reports are the source of truth for phase completion, not `query state`.
- **Minor, flagged-not-fixed**: `.gitignore` + `docs/workflow-harness/migration.md` line 37 say `.kit/docs/` should be ignored; it has been tracked since `77ed8bb`. Worth a small future cleanup reconciling the two.
- **`audit --json` debt tail**: long-standing `contract_violations`/`unlinked_proofs` from phases predating this initiative (slim-playbooks, cli-release, etc.) — none touch this pass's files.

## Technical Context

**Approach (single-source-playbooks, closed)**: guard with a test, not a new command or git hook (explicitly rejected alternatives) — the embed (`cli/internal/embedded/`) is canonical, `.kit/docs/` is generated output, never hand-edited.

**Key Files**:
- `.kit/planning/ROADMAP.md` — Harness Subtraction Pass, now marked fully closed
- `cli/internal/embedded/projection_drift_test.go` — the drift guard
- `cli/docs/CONTRACT.md`, `README.md`, `docs/workflow-harness/migration.md` — the "edit the embed only" rule, stated in 3 places

## Next Steps

1. **→ START HERE**: nothing queued for this initiative. If new work is wanted, it starts fresh with `/brainstorm` (no existing SPEC/ROADMAP covers anything beyond this closed pass).
2. **Optional cleanup** (low priority, carried forward across the whole pass): add a story-status-update command so `query state`/`resume` stop showing stale phase status; reconcile the `.gitignore`/`migration.md` contradiction about `.kit/docs/` tracking.
3. No dependency chain is open — this pass has no continuation.

## Notes

This closes the Harness Subtraction Pass. Unlike prior handoffs in this chain, there is no deferred phase to resume — the next session's first move should be deciding what new work to start, not continuing this one.

---

*Generated by handoff on 2026-07-22*
