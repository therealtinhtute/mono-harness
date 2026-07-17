# State Contract v1

Defines how `zharness` represents workflow position and status, replacing the legacy `.kit/workflow-state.yml` pointer file with DB-backed state plus a generated human view.

## State Model

- `schema_version` — integer, bumped on any breaking table/entity change; stored in a single-row `meta` table
- `current_phase` — story slug of the phase actively being executed or resumed (`none` if no phase started)
- `entry_phase` — story slug recommended as the next phase when execution has not started yet
- `phase.status` enum: `planned | in-progress | checked | done`
  - `planned` — story exists, no run recorded yet
  - `in-progress` — at least one `run` recorded, no clean `check` yet
  - `checked` — latest `check` for the phase has verdict `APPROVED` or `APPROVE_WITH_REQUESTS`
  - `done` — phase explicitly closed (handoff recorded against it, or superseded by roadmap advance)
- Artifact pointers are **by ULID**, not by path — `zharness query state --json` resolves ULIDs to their current file path at read time, so a path rename never orphans a pointer the way the legacy yml did
- Every mutating entity (`run`, `check`, `handoff`, `story`) carries its own ULID; `current` pointers in `meta` reference the latest ULID per entity type per phase

## Writer / Reader Ownership

| Command / Skill | Writes | Reads |
|---|---|---|
| `to-plan` (`story`, `init`) | `story` rows (phase creation) | `spec`, prior `story` rows |
| `work` (`trace add`, run completion) | `run` rows, phase status → `in-progress` | `story`, `latest run/check` |
| `check` (`check record`) | `check` rows, phase status → `checked` | `run`, `story` |
| `handoff` (`handoff record`) | `handoff` rows, phase status → `done` (when the handoff closes a phase) | `run`, `check`, `story` |
| `watzup` (`resume`) | nothing | everything — read-only snapshot |
| `git` (`query check --latest`) | nothing | `check` (latest only) |

No command outside this table writes phase/status/pointer state. Skills that only read (`watzup`, `git`) never call a mutating command for state purposes.

## Stale-Pointer Rules

| Condition | Detection | Recovery action |
|---|---|---|
| Missing file | ULID resolves to a path that doesn't exist on disk | `resume` reports `drift: missing_file` with the last-known path; recovery = `to-plan phase {slug}` to regenerate, or `git checkout` if it was tracked and deleted accidentally |
| Unknown phase slug | `current_phase` references a story slug not in the `story` table | `drift: unknown_phase`; recovery = `to-plan phase {slug}` to create the missing story, or correct `current_phase` via `to-plan` if it was a typo |
| Out-of-order run/check IDs | A `check` row's `created_at` (ULID-ordered) predates the `run` it claims to gate | `drift: out_of_order`; recovery = re-run `check full` against the latest `run` — the stale `check` is left in place (append-only), superseded by the new one |
| DB unreadable / absent | `zharness resume` cannot open `harness.db` | readiness state `no-harness`; recovery = `zharness init` (new project) or `zharness import` (legacy `.kit/` present) |

## Legacy Field Mapping (`workflow-state.yml` → DB)

| Legacy field | DB representation |
|---|---|
| `current_phase` | `meta.current_phase` (story slug, FK to `story.slug`) |
| `entry_phase` | `meta.entry_phase` (story slug) |
| `spec` | dropped — `spec` path is fixed at `.kit/planning/SPEC.md` by convention, no longer a pointer that can drift |
| `roadmap` | dropped — same reasoning, fixed at `.kit/planning/ROADMAP.md` |
| `active_context` | derived, not stored — `query state` resolves to `.kit/planning/phases/{current_phase}/{current_phase}-CONTEXT.md` from `current_phase` directly |
| `active_plan` | derived, not stored — same derivation pattern for `-PLAN.md` |
| `latest_cook_run` | `meta.latest_run_id` (ULID, FK to `run.id`) — resolved to path via `run.artifact_path` at read time |
| `latest_check_report` | `meta.latest_check_id` (ULID, FK to `check.id`) — resolved via `check.artifact_path` |
| `handoff` | dropped as a fixed pointer — `handoff` is now a queryable entity list (`query handoff --latest`); the human `HANDOFF.md` path stays fixed at `.kit/HANDOFF.md` by convention like `spec`/`roadmap` |
| `last_updated` | dropped as a manual field — every entity carries its own ULID-derived timestamp; `resume` reports the most recent write across all entities instead of one hand-maintained field |

No legacy field is silently lost: five fields map 1:1 to DB columns, four become derived/convention paths (still resolvable, just no longer separately stored and thus no longer able to drift from reality), one (`handoff` as a single pointer) is superseded by a queryable list which is strictly more capable.
