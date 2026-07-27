# State Contract v1

Defines how `zharness` represents workflow position and status in the harness database. The legacy `.kit/workflow-state.yml` file is import input only; it is not live state.

## State Model

- `schema_version` — integer stored in the single-row `meta` table and bumped for breaking table or entity changes.
- `docs_version` — embedded-docs version stamped by `init` or `init --refresh-docs`; release builds use the CLI version and unreleased builds use `"dev"`. `"dev"` never triggers staleness.
- `current_phase` — story slug of the phase currently being executed or resumed; null when no phase has started.
- `entry_phase` — story slug recommended as the next phase before execution starts.
- `latest_run_id` and `latest_check_id` — database IDs stored in `meta`; readers return the IDs directly.
- `latest_handoff_id` — derived by `resume` from the newest handoff row by `created_at`, then ID. It is not a `meta` pointer.
- Story status enum: `planned | in-progress | checked | done`.
  - `planned` — `story` created the phase and no run has started it.
  - `in-progress` — `run create` registered a run for the phase.
  - `checked` — `check record` recorded `APPROVED` or `APPROVE_WITH_REQUESTS` for the run.
  - `done` — `handoff record --close-phase` closed the run's story using a clean check for that same run.
- Artifact paths on runs, checks, and proof links are optional, deprecated legacy metadata. They are never pointer targets, are never resolved by state readers, and do not participate in drift or lifecycle validity checks.

## Writer / Reader Ownership

| Command | Writes | Reads |
|---|---|---|
| `story` | Story row with status `planned` | Optional dependency story |
| `run create` | Run row; story status → `in-progress`; `meta.current_phase`; `meta.latest_run_id` | Target story |
| `trace add` | Trace row and optional run linkage | Optional target run |
| `check record` | Check row; `meta.latest_check_id`; story status → `checked` for a clean verdict | Gated run and story |
| `handoff record` | Handoff row; with `--close-phase`, story status → `done` | Optional run/check; closing requires a clean check for the supplied run |
| `resume`, `query state`, `query phases`, `query artifacts`, `query check --latest`, `validate`, `audit` | Nothing | DB lifecycle state |

Only `run create`, `check record`, and closing `handoff record` advance an existing phase's durable status or current/latest pointers. Read-only skills and commands do not mutate state.

## Drift and Error Rules

| Condition | Detection | Result and recovery |
|---|---|---|
| Unknown phase slug | `meta.current_phase` has no matching story row | `resume` reports `drift: unknown_phase`; create the story or correct the DB pointer through the owning workflow command. Normal CLI writes are protected by foreign keys. |
| Latest check gates a different run | `meta.latest_check_id` resolves to a check whose `run_id` differs from `meta.latest_run_id` | `resume` reports `drift: out_of_order`; record a check for the latest run or correct the latest run/check pointers. This check compares the linked IDs, not timestamps or artifact paths. |
| Stale DB pointer | A populated `meta` phase/run/check pointer has no referenced DB row | `validate` reports `stale_pointer`; repair the referenced lifecycle row or pointer. |
| Harness DB absent | No root `harness.db` exists | `resume` returns readiness `no-harness` with no drift findings; run `zharness init` for a new project or `zharness import` for legacy state. |
| Harness DB unreadable | The DB exists but cannot be opened or queried | Command fails with the `db_unreadable` system error; no readiness state is fabricated. |
| Docs stale | `meta.docs_version` and the running CLI version differ, both are non-empty, and neither is `dev` | `resume` reports `drift: stale_docs`; recover with `zharness init --refresh-docs`. Missing versions are unversioned, not drifted. |

## Legacy Field Mapping (`workflow-state.yml` → DB)

| Legacy field | DB disposition |
|---|---|
| `current_phase` | `meta.current_phase` story slug |
| `entry_phase` | `meta.entry_phase` story slug |
| `spec` | Not stored; planning files are not state pointers |
| `roadmap` | Not stored; import may read it only to recover story goals |
| `active_context` | Not stored; context paths are not state pointers |
| `active_plan` | Not stored; plan paths are not state pointers |
| `latest_cook_run` | Imported as a run row plus `meta.latest_run_id` when the legacy filename can identify the phase and timestamp. The old path may be retained only as deprecated `run.artifact_path` metadata; readers return the DB ID without path resolution. |
| `latest_check_report` | Maps semantically to `meta.latest_check_id`. Legacy import does not synthesize a check from a path without a verdict, so the pointer remains unset until a check row exists; readers return the DB ID without path resolution. |
| `handoff` | Legacy single-file pointer is not stored. Handoffs are DB rows, and `resume` derives the latest row's ID. |
| `last_updated` | Not stored; lifecycle rows carry their own timestamps. |

Import rejects unknown legacy keys. Known path-only fields have the explicit non-pointer dispositions above and never become live drift targets.
