# COOK RUN

Run ID: work-20260717-1100-harness-contracts
Mode: full
Status: done
Spec: .kit/planning/SPEC.md
Roadmap: .kit/planning/ROADMAP.md
Workflow State: .kit/workflow-state.yml
Phase: harness-contracts
Plan: .kit/planning/phases/harness-contracts/harness-contracts-PLAN.md
Started At: 2026-07-17 11:00

## Preflight
- scope drift: no — working tree carries harness-concept's already-gated, uncommitted changes (README.md, skills/workflow/README.md, docs/workflow-harness/); none overlap harness-contracts' allowed surfaces (cli/docs/*, 4 template files)
- working tree note: harness-concept phase uncommitted (clean gate, not yet committed) — pre-existing, not this phase's concern
- required artifacts present: yes (SPEC.md locked, harness-contracts-CONTEXT.md, harness-contracts-PLAN.md refreshed 2026-07-17)
- selected phase: harness-contracts (goal: freeze all machine contracts before Go code — STATE.md, CONTRACT.md, SCHEMA.md, 4 template frontmatter contracts)

## Wave / Task Log
### Wave 1
#### T1 — cli/docs/STATE.md (state contract v1)
- status: DONE
- changed files:
  - cli/docs/STATE.md (new)
- verification:
  - `grep -iE 'TBD|TODO' cli/docs/STATE.md` → no matches (pass)
  - all 10 `workflow-state.yml` fields present in mapping table (pass)
- notes:
  - none

### T2 — Artifact frontmatter contracts in 4 templates
- status: DONE_WITH_CONCERNS
- changed files:
  - skills/workflow/brainstorm/references/spec-template.md
  - skills/workflow/work/references/run-artifact-template.md
  - skills/workflow/check/references/report-template.md
  - skills/workflow/handoff/references/handoff-template.md
- verification:
  - `grep -l '^id:' skills/workflow/{brainstorm/references/spec-template.md,work/references/run-artifact-template.md,check/references/report-template.md,handoff/references/handoff-template.md}` → all 4 listed (pass)
- notes:
  - PLAN→SPEC cross-link has no carrier file this phase (`phase-plan-template.md` not in Allowed Surfaces) — logged in `.kit/implementation-notes.md`
  - handoff-template.md carries two frontmatter key sets (human + machine contract) with a documented manual-sync rule

### Wave 2
#### T3 — cli/docs/CONTRACT.md (full command surface)
- status: DONE_WITH_CONCERNS
- changed files:
  - cli/docs/CONTRACT.md (new)
- verification:
  - `grep -c '^### \`' cli/docs/CONTRACT.md` → 21 headings (20 commands, `db` has 2 subcommand headings)
  - `grep -c '\-\-json' cli/docs/CONTRACT.md` → 23 (every command documents a `--json` shape; `query` documents 2)
  - `grep -c '^- Errors:' cli/docs/CONTRACT.md` → 21 (1:1 with headings)
- notes:
  - documented 20 commands, not SPEC R6's literal 19 — `handoff` added, deviation flagged in a blockquote at the top of CONTRACT.md and in `.kit/implementation-notes.md` (R6 vs R18 SPEC-internal inconsistency)
  - `validate` section carries a "Known gap" note for the unimplementable SPEC→PLAN link (T2 concern above)

#### T4 — cli/docs/SCHEMA.md (DB + changesets)
- status: DONE_WITH_CONCERNS
- changed files:
  - cli/docs/SCHEMA.md (new)
- verification:
  - 11 table headings (`meta` + 10 data tables) match 11 rows in the Table ↔ Changeset Entity Type mapping (pass)
  - every CONTRACT.md mutating command (9 entity-writing + 3 `meta`-only) names an entity from SCHEMA.md; 7 read-only commands write none; 9+3+7=19 matches CONTRACT.md's 19 non-`db` commands (pass, cross-check inspection)
- notes:
  - resolved a second SPEC-vs-locked-decision inconsistency: SPEC R13 and this phase's own PLAN.md list `phases` as a table separate from `stories`, but `harness-contracts-CONTEXT.md`'s Locked Decisions already resolve SPEC Open Question 1 ("one story per phase, slug = phase slug — no longer an open assumption"). Implemented one `stories` table carrying phase status, no separate `phases` table — applying an already-locked decision, not a new deviation. Logged in `.kit/implementation-notes.md`.
  - "epoch-fence adaptation" (PLAN T4 step 3) grounded by reading `~/Lab/harness-experimental/scripts/harness-epoch-transition.py` directly (not fabricated): upstream is a crash-recoverable rename-based swap between two full DB/changeset-log generations, guarded by a checksummed journal. Adapted to zharness's single-generation topology as a SQLite-transaction fence (`meta.last_applied_changeset` updated in the same transaction as the changeset's row mutations) — same crash-safety guarantee, no journal/rename machinery, since zharness never swaps generations. Full reasoning in `.kit/implementation-notes.md`.

## Summary
All 4 tasks (T1–T4) complete: `cli/docs/STATE.md`, `cli/docs/CONTRACT.md`, `cli/docs/SCHEMA.md` (new), plus frontmatter contracts added to 4 skill templates (`spec-template.md`, `run-artifact-template.md`, `report-template.md`, `handoff-template.md`). Every machine contract needed before cli-core writes real Go code now exists: state model, 20-command surface (19 SPEC-listed + `handoff`, flagged deviation), 11-table SQLite schema with changeset entity mapping, replay/idempotency/epoch-fence rules, and artifact frontmatter cross-links.

2 tasks landed `DONE_WITH_CONCERNS` (T2, T3), 1 landed `DONE_WITH_CONCERNS` with a resolved-not-open note (T4) — no task `BLOCKED` or `NEEDS_CONTEXT`. Three genuine SPEC-level inconsistencies were found and handled by grounded reading rather than silent guessing, each logged in full in `.kit/implementation-notes.md`:
1. PLAN→SPEC cross-link has no carrier file this phase (`phase-plan-template.md` out of Allowed Surfaces) — real gap, deferred to skill-adapters or cli-domain escalation.
2. SPEC R6 (19 commands) vs R18/continuity-CONTEXT.md (requires `handoff record`) — resolved by implementing 20, flagged, not editing SPEC.md.
3. SPEC R13/PLAN wording (`phases` as separate table) vs the already-locked story↔phase mapping decision — resolved by collapsing into one `stories` table; this one applies an existing lock rather than creating a new deviation.

No working-tree scope drift: only `cli/docs/*` (new) and the 4 named template files changed, matching harness-contracts' Allowed Surfaces exactly.

## Next Recommended Action
Run `check full` against this phase's diff (harness-contracts is a docs-only phase — no build/test/lint exists yet at repo root; gate evidence will be `grep`/inspection-based per each task's verification, plus scope/artifact-alignment review). On a clean gate, advance to `cli-core` (first phase with real Go code — `go build`/`go test`/`go vet` become actual gate evidence from here on).

## Gate Result — check full
- verdict (1st pass): APPROVE WITH REQUESTS — 🟠 Major: `CONTRACT.md` documented 20 commands vs the phase's own locked 19 (`CONTEXT.md` Goal + `PLAN.md` T3's explicit `avoid`/verification clauses), bypassing T3's named escalation path (`continuity-CONTEXT.md`: "if the CLI lacks it, that is a Phase 4 gap — escalate, don't shim"). 💡 Suggestion (no action): PLAN→SPEC cross-link gap, already correctly handled.
- remediation applied: trimmed `CONTRACT.md` back to exactly 19 commands (20 `###` headings, `db` split into 2 subcommand headings); replaced the "deviation flag (implemented)" blockquote with an "escalation note (deferred)"; updated `SCHEMA.md`'s `handoffs` entity row + closing verification math to match. Logged in `.kit/implementation-notes.md` (`check gate — reverted the 20th command`).
- verdict (2nd pass, re-verified): `grep -c '^### \`' cli/docs/CONTRACT.md` → 20 (pass, matches 19 commands + db split); `grep -c '^- Errors:' cli/docs/CONTRACT.md` → 20 (1:1 with headings, pass); SCHEMA.md verification math recomputed (8 entity-producing + 3 meta-only + 1 replay-special + 7 read-only = 19, pass). Scope: on target, 7 files, all within Allowed Surfaces. No 🔴/🟠 findings remain. **APPROVED.**

## Phase Status: CLOSED — clean gate, ready to advance to cli-core.
