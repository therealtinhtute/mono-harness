# Context: scoring-removal

Phase: scoring-removal
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: medium
Expected Proof: unit, integration, command-output, manual-check

## Goal
Delete the "deterministic verdict" scoring — `ScoreTrace` tier logic, the `score-trace` command, and `entropy_score` from `audit --json` — because it measures string length and finding-counts, not evidence. Keep the lane×proof-class matrix in `check` as the sole, meaningful verdict. The gate's pass/fail outcome must not change.

## Scope Boundary
### Allowed Surfaces
- `cli/internal/application/score.go` — remove `ScoreTrace` (and `TraceScore` if unused after)
- `cli/internal/application/audit.go` — remove `entropyScore` + the `EntropyScore` field from `AuditReport`
- `cli/internal/interfaces/score.go` — remove the `score-trace` command
- `cli/docs/embedded/playbooks/check.md` — remove Step 4's `score-trace` loop; keep the matrix
- `cli/docs/CONTRACT.md`, `SCHEMA.md` — update `audit --json` shape (no `entropy_score`), remove `score-trace`
- related tests

### Forbidden Surfaces
- the lane×proof matrix logic and its pass/fail behavior (must stay identical)
- `check record` / meta pointers (Phase 1)
- dropped entities/tables (Phase 2)
- changeset format / replay

## Spec Hooks
- Requirement R3 (remove scoring, keep matrix)
- Constraint: **no change to the gate's pass/fail outcome** — only the vestigial score output is removed.
- Acceptance: `audit --json` has no `entropy_score`; `check.md` doesn't call `score-trace`; a check-gate run still FAILs on a missing required proof cell.

## Locked Decisions
- **Remove, not repair.** Chosen over enriching the trace schema to make scoring valid (rejected — schema-touching, out of this slice per SPEC Key Decisions).
- The matrix's `command-output` + `manual-check` proof classes already cover what `score-trace` gated ("does this trace count as evidence"), so removing tier scoring loses no real gate signal.
- `score-context` is removed in Phase 2 (dead surface); this phase handles only `score-trace` + `entropy_score`.

## Assumptions
- No skill other than `check` consumes `score-trace` (verify: grep playbooks — only `check.md` Step 4 references it).
- `entropy_score` has no downstream consumer beyond being printed (audit callers don't branch on it).

## Canonical Refs
- `cli/internal/application/score.go` (`ScoreTrace`), `audit.go` (`entropyScore`, `AuditReport.EntropyScore`)
- `.kit/docs/playbooks/check.md` Step 4 (score-trace loop) + Validation Matrix (the part that stays)

## Rejected Options
- Keep `entropy_score` as "informational" — the audit showed it's non-actionable; keeping it invites treating a finding-count as a health metric.
- Repair scoring by adding trace evidence fields — deferred to a possible future initiative.

## Deferred Ideas
- A genuinely evidence-based trace quality score (would need schema enrichment) — out of scope, noted in SPEC Deferred Ideas.

## Escalate If
- Removing `score-trace` changes any matrix pass/fail result → STOP (that means the matrix secretly depended on tier scoring; rescope via `to-plan`).
- A second consumer of `entropy_score` surfaces → STOP, reassess.
