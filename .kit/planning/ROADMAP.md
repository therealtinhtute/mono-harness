# ROADMAP: Harness Subtraction Pass

## Planning Basis
- source spec: `.kit/planning/SPEC.md`
- planning mode: `full`
- entry phase (recommended): **write-boundary**
- execution mode: `work full` (high-risk lane — proof matrix requires unit + integration + command-output + manual-check)
- sequencing: **linear** — phases 2 and 3 both edit `score.go`/`audit.go`/`check` playbook; phase 4 depends on the final playbook text. Linear order avoids same-file conflicts.

## Phase 1: write-boundary
**Status:** done — implemented + gated APPROVED 2026-07-21 (commit `32cb60c`, check report `.kit/reports/check/20260721-1044-write-boundary.md`). Bookkeeping only was stale (this ROADMAP + the phase PLAN/CONTEXT status fields, and the harness `stories.status` DB row, which the CLI has no command to transition — a confirmed gap, not yet fixed).

**Goal:** The CLI owns 100% of harness writes — no playbook hand-authors changeset JSONL.

**Deliverables:**
- `zharness run create` command: creates the run row + sets `meta.latest_run_id` atomically (one changeset, one tx).
- `check record` sets `meta.latest_check_id` itself (default behavior; no separate hand-authored meta changeset).
- `work.md` + `check.md` embedded playbooks rewritten to call the new commands; hand-authored-JSONL steps deleted.

**Dependencies:**
- none (entry phase)

**Risks / Watch-fors:**
- `run create` must reproduce the exact two-line semantics (create run + meta pointer) the playbook did by hand, or replay/resume drifts.
- Simple-mode must keep skipping DB registration (FK constraint on `runs.story_slug`) — `run create` is full-mode only.

## Phase 2: dead-surface-removal
**Goal:** Remove built-but-unused surface: `decision`, `backlog`, `tool`, `propose`, `score-context`.

**Deliverables:**
- Cobra subcommands + interfaces + application code + entity rows/tables/columns for the five removed.
- `migrations.go` schema change dropping unused tables + `schema_version` bump.
- Their tests deleted; `go build`/`go test` green.

**Dependencies:**
- write-boundary (settles the CLI command set first; sequencing only)

**Risks / Watch-fors:**
- Schema change must not break replay of the existing committed changesets (none reference dropped entities — verify by grep before deleting) or `import` of a legacy `.kit/`.
- `propose` lives in `audit.go`, `score-context` in `score.go` — the same files Phase 3 edits; do Phase 2 first to keep deletions cohesive.

## Phase 3: scoring-removal
**Goal:** Delete the meaningless "deterministic verdict" scoring; keep the lane×proof matrix as the real verdict.

**Deliverables:**
- `ScoreTrace` tier logic + `score-trace` command + `entropy_score` field removed from `audit --json`.
- `check.md` Step 4 no longer loops `score-trace`; the proof matrix remains the gate.
- `CONTRACT.md` / `SCHEMA.md` updated for the changed `audit --json` shape.

**Dependencies:**
- dead-surface-removal (shares `score.go`/`audit.go`)

**Risks / Watch-fors:**
- The lane×proof gate's pass/fail outcome MUST be unchanged — only the vestigial score output goes. Prove with a check-gate run that still FAILs on a missing required proof cell.

## Phase 4: single-source-playbooks
**Goal:** `.kit/docs/playbooks/*` is a pure projection of the Go embed; drift is caught by a test.

**Deliverables:**
- `init` writes playbooks as a projection of `cli/docs/embedded/playbooks/`; documented that humans edit only the embed.
- A test (or `zharness playbooks verify`) failing when a scaffolded copy diverges from the embed.

**Dependencies:**
- write-boundary + scoring-removal (playbook text must be final before locking the projection)

**Risks / Watch-fors:**
- The drift test must compare against the embed as the single source of truth, not freeze a stale `.kit/docs/` copy.

## Next Steps
- **write-boundary is done** (see Phase 1 Status above); `work full phase dead-surface-removal` is next.
- `check full` gates each phase (high-risk proof matrix).
- `git` / `handoff` to wrap up.
