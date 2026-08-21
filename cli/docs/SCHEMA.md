# DB + Changeset Schema v1

SQLite schema for `harness.db` and the changeset line/file format that reproduces it. Every database-mutating command in `CONTRACT.md` writes one changeset before touching the DB (R7); replaying all changesets from empty yields identical state. File-only `scaffold` is outside this database mutation contract.

> **Table-count note**: SPEC R13 and this phase's own PLAN.md list `phases` as a table separate from `stories`. `harness-contracts-CONTEXT.md`'s Locked Decisions resolve SPEC's Open Question 1 explicitly: "one `zharness` story per `to-plan` phase, story slug = phase slug... no longer an open assumption." A separate `phases` table would duplicate that mapping and require keeping two rows in sync on every phase transition — the opposite of what R13's own idempotency intent wants. This schema defines **one** `stories` table carrying phase status; there is no `phases` table. Flagged in `.kit/implementation-notes.md`, not a new open question — applying an already-locked decision, not creating a deviation.

## Tables

### `meta` (single row)
| Column | Type | Notes |
|---|---|---|
| `schema_version` | INTEGER | bumped on breaking table/entity change |
| `current_phase` | TEXT, nullable | FK `stories.slug` |
| `entry_phase` | TEXT, nullable | FK `stories.slug` |
| `latest_run_id` | TEXT, nullable | FK `runs.id` |
| `latest_check_id` | TEXT, nullable | FK `checks.id` |
| `last_applied_changeset` | TEXT, nullable | ULID of the most recently applied changeset file — the replay fence marker, see Epoch-Fence Adaptation below |
| `docs_version` | TEXT, nullable | added in migration `0002_meta_docs_version`; stamps the embedded-docs version written into a project by `init`/`init --refresh-docs` — the CLI's own version string for release builds, `"dev"` for unreleased builds. Set via the same changeset-first path as the other meta columns (`metaColumns` allowlist in `changeset.go`). |

### Workflow entities

#### `stories` (carries phase semantics — see table-count note above)
| Column | Type | Notes |
|---|---|---|
| `id` | TEXT PK | ULID |
| `slug` | TEXT UNIQUE | phase slug; frontmatter `phase` key resolves here |
| `goal` | TEXT | |
| `status` | TEXT | enum `planned\|in-progress\|checked\|done` (STATE.md) |
| `depends_on` | TEXT, nullable | FK `stories.slug` |
| `created_at` | TEXT | ULID-derived timestamp |

#### `runs`
| Column | Type | Notes |
|---|---|---|
| `id` | TEXT PK | ULID |
| `story_slug` | TEXT | FK `stories.slug` |
| `plan_id` | TEXT, nullable | Optional phase PLAN ULID supplied by `run create --plan-id`; not a foreign key and not part of current DB lifecycle-link validation |
| `trace_ids` | TEXT | Legacy JSON array column, currently stored as `[]`; `trace add` persists links in `traces.run_id` instead of updating this field |
| `artifact_path` | TEXT NOT NULL | Optional/deprecated legacy metadata; absence is encoded as `""` for additive compatibility with migration `0001_init`. No filesystem path or artifact existence is required. |
| `created_at` | TEXT | |

#### `checks`
| Column | Type | Notes |
|---|---|---|
| `id` | TEXT PK | ULID |
| `run_id` | TEXT | FK `runs.id` |
| `verdict` | TEXT | enum `APPROVED\|APPROVE_WITH_REQUESTS\|REQUEST_CHANGES` |
| `judge` | TEXT, nullable | added in migration `0006_check_judge`; enum `independent\|same-session`; `NULL` for a check recorded before this migration |
| `judge_model` | TEXT, nullable | added in migration `0006_check_judge`; free-text model identifier that produced the verdict; `NULL` for a check recorded before this migration |
| `proof_links` | TEXT | JSON array of `{command, output_ref, artifact_path}`; each `artifact_path` is optional/deprecated legacy metadata and is not a filesystem requirement |
| `artifact_path` | TEXT, nullable | Optional/deprecated legacy check-artifact metadata; not a filesystem requirement |
| `created_at` | TEXT | |

#### `handoffs`
| Column | Type | Notes |
|---|---|---|
| `id` | TEXT PK | ULID |
| `run_id` | TEXT, nullable | FK `runs.id` |
| `check_id` | TEXT, nullable | FK `checks.id` |
| `anchors` | TEXT | Durable JSON `{latest_run_id, latest_check_id, open_items, exact_next_action}`; run/check/next-action keys are present only when supplied. `exact_next_action` persists the plan's Current State field (D1, `docs/audit/workflow-harness-ceremony-audit.md`) — no migration needed, this column was already free-form JSON |
| `created_at` | TEXT | |

Handoff anchors and open items are durable database state. There is no fixed `.kit/HANDOFF.md` convention or required handoff markdown path in the schema.

### Ported entities

#### `intakes`
| Column | Type | Notes |
|---|---|---|
| `id` | TEXT PK | ULID |
| `type` | TEXT | enum `new-spec\|spec-slice\|change-request\|new-initiative\|maintenance\|harness-improvement` |
| `summary` | TEXT | |
| `lane` | TEXT | enum `tiny\|normal\|high-risk` |
| `plan_path` | TEXT, nullable | Optional repository-relative path to the initiative's evolving plan; added by migration `0005_intake_plan_path` |
| `plan_id` | TEXT, nullable | Optional plan ULID, same value as `runs.plan_id`; added by migration `0009_intake_plan_id`. Not a foreign key, not part of lifecycle-link validation — the join `check record` uses to resolve a run's lane and gate `--judge` for `high-risk` (G2, `docs/audit/workflow-harness-ceremony-audit.md`/V2) |
| `created_at` | TEXT | |

#### `interventions`
| Column | Type | Notes |
|---|---|---|
| `id` | TEXT PK | ULID |
| `verdict_id` | TEXT | FK `checks.id` — the check verdict being overridden |
| `reason` | TEXT | |
| `created_at` | TEXT | |

#### `traces`
| Column | Type | Notes |
|---|---|---|
| `id` | TEXT PK | ULID |
| `run_id` | TEXT, nullable | FK `runs.id` |
| `wave` | INTEGER | |
| `summary` | TEXT | |
| `task` | TEXT, nullable | added in migration `0008_trace_task_granularity`; task name for a task-level trace, `NULL` for wave-level |
| `task_status` | TEXT, nullable | added in migration `0008_trace_task_granularity`; enum `DONE\|DONE_WITH_CONCERNS\|NEEDS_CONTEXT\|BLOCKED` (docs/playbooks/work.md's Status Routing), `NULL` for wave-level |
| `created_at` | TEXT | |

#### `decisions`
| Column | Type | Notes |
|---|---|---|
| `id` | TEXT PK | ULID |
| `run_id` | TEXT, nullable | FK `runs.id`; optional, shared across a `decision add` batch |
| `phase` | TEXT, nullable | FK `stories.slug` |
| `task` | TEXT, nullable | free text, not a modeled entity — tasks exist only in plan markdown |
| `decision` | TEXT NOT NULL | |
| `rationale` | TEXT NOT NULL | |
| `created_at` | TEXT | |

Re-created by migration `0007_decisions`, which re-adds a table `0003_drop_dead_surface` dropped as dead surface with no writer (`git log -S "decision add"` returns nothing before `0007`) — not a reversal of that migration's judgment, since it was never wired to anything. The schema differs from the original (`id, summary, rationale, rejected, created_at`, no phase/task/run linkage): `phase`/`task`/`run_id` tie a decision to the work that produced it, matching what `work.md`'s `## Decisions` section already records in markdown. No historical changeset references the `decision` entity, so this migration is purely additive to replay.

### Managed infrastructure

#### `managed_docs`
| Column | Type | Notes |
|---|---|---|
| `id` | TEXT PK | stable repository-relative path |
| `path` | TEXT UNIQUE | managed root doc path |
| `installed_sha256` | TEXT | embedded bytes last installed without conflict |
| `docs_version` | TEXT | CLI docs version that recorded the hash |
| `updated_at` | TEXT | last metadata update |

The writable database is the ignored root `harness.db`. Tracked replay deltas remain under `.kit/changesets/`; no baseline or second database exists.

#### `plan_index`
| Column | Type | Notes |
|---|---|---|
| `id` | TEXT PK | ULID minted when the path is first indexed |
| `path` | TEXT UNIQUE | active plan path (`docs/plans/active/{slug}.md`) |
| `sha256` | TEXT | hash of plan content as of the last index refresh |
| `status` | TEXT | plan frontmatter `status` as of the last index refresh |
| `updated_at` | TEXT | last index refresh time |

Not a changeset entity — no command writes `plan_index` directly. The read path refreshes a row whenever the on-disk hash differs from the indexed `sha256`, so staleness is a comparison against the file's real content, never a timestamp guess (R9, `docs/decisions/0001-markdown-as-source-of-truth.md`).

#### `memories`
| Column | Type | Notes |
|---|---|---|
| `id` | TEXT PK | ULID minted by `memory add` |
| `path` | TEXT UNIQUE | entry path (`docs/memory/{id}.md`) |
| `type` | TEXT | free-text memory type, set by `memory add --type` |
| `scope` | TEXT | enum `plan\|global` |
| `plan_id` | TEXT, nullable | set only when `scope = plan`; ties the entry to one initiative plan |
| `sha256` | TEXT | hash of the entry's markdown content as of the last write/rebuild |
| `created_at` | TEXT | ULID-derived timestamp |

Markdown-first, same as every other durable write path: `memory add` writes `docs/memory/{id}.md` before inserting this row, and `db rebuild` reconstructs the whole table from committed `docs/memory/*.md` content alone — no memory content is ever DB-only (R1/R4, `docs/decisions/0001-markdown-as-source-of-truth.md`).

## Table ↔ Changeset Entity Type

Every table maps to exactly one changeset `entity` string (the value in `{op, entity, id, fields, at}`). Table names are plural (SQL convention, SPEC R13's own wording); entity strings are singular, matching the CONTRACT.md command that produces them.

| Table | Entity string | Producing command(s) |
|---|---|---|
| `meta` | `meta` | side-effect of any command that updates `current_phase`/`latest_run_id`/etc. — never created directly |
| `stories` | `story` | `story`; `run create` updates status to `in-progress`, clean `check record` updates it to `checked`, and closing `handoff record --close-phase` updates it to `done` |
| `runs` | `run` | `run create` |
| `checks` | `check` | `check record` |
| `handoffs` | `handoff` | `handoff record` |
| `intakes` | `intake` | `intake` |
| `interventions` | `intervention` | `intervention` |
| `traces` | `trace` | `trace add` |
| `decisions` | `decision` | `decision add` |
| `managed_docs` | `managed_doc` | `init`, `init --refresh-docs`, layout migration |
| `memories` | `memory` | `memory add` |

Cross-check against CONTRACT.md: database-mutating commands include `intake`, `story`, `intervention`, `trace add`, `decision add`, `run create`, `check record`, and `handoff record`, plus infrastructure mutations from `init`/`import`/`migrate`. Lifecycle commands may write multiple entity lines atomically for status and meta-pointer transitions while each created row uses the singular entity string above. `resume`, `query`, `validate`, `audit`, and `preflight` are read-only and write no changeset; `scaffold` is file-only and writes no database entity or changeset.

## Changeset Format

- Line shape: `{"op": "create"|"update", "entity": "...", "id": "ulid", "fields": {...}, "at": "ISO8601"}`
- File naming: `{ulid}.changeset.jsonl`, one file per command invocation (R8) — a command that writes to two tables (e.g., `work` completing a wave: a `trace` row plus a `stories.status` update) still writes both lines into the **same** file, since both mutations belong to one command batch
- `op: "create"` — `id` is freshly minted for this line, `fields` is the full row
- `op: "update"` — `id` references an existing row, `fields` is a partial patch (e.g., `{"status": "in-progress"}` against an existing `stories` row)

## Replay Ordering

Changeset files replay in **ULID filename order** (R8) — ULIDs are lexically sortable and timestamp-prefixed, so filename order reproduces original write order exactly. Replay is a full drop-and-rebuild: `harness.db` is discarded, every `.changeset.jsonl` under `.kit/changesets/` is applied in ULID order from empty. This is what `continuity`'s fresh-machine rebuild does (`db changeset apply` per file, in order).

## Idempotency Key Rules

Two layers, matching R7 ("applying the same changeset twice changes nothing"):

1. **File-level (primary path)**: `meta.last_applied_changeset` holds the ULID of the most recently applied changeset file — the replay fence. `db changeset apply <path>` compares the path's ULID against the fence: ULID ≤ fence → the whole file is skipped, reported as `skipped_already_applied` (CONTRACT.md). This is the normal-path check; it never opens the file's lines at all once skipped.
2. **Line-level (forced/manual re-apply)**: if a file is force-applied despite already being past the fence, `create` lines are idempotent via `INSERT OR IGNORE` keyed on (`entity`, `id`) — `id` is minted once when the original command ran and is never re-minted on replay, so a duplicate `create` line is a guaranteed no-op. `update` lines are idempotent via plain `UPDATE ... WHERE id = ?` — setting a column to a value it already holds changes nothing. Ordering correctness for `update` lines depends on ULID-order replay (layer above); an out-of-order manual replay is exactly the condition STATE.md's `out_of_order` stale-pointer rule detects and flags, not something this layer re-solves.

## Epoch-Fence Adaptation

SPEC's architecture note (R2) places "epoch-fence logic" in `internal/infrastructure`. The term originates upstream in `harness-experimental/scripts/harness-epoch-transition.py`: a crash-recoverable activation of one database/changeset-log **generation pair**, using a checksummed journal (payload + SHA-256, written via tmp-file + fsync + atomic rename + directory fsync) and four ordered rename steps, so a crash mid-transition can be resolved by replaying the journal (`recover forward` finishes the named pair; `recover compensate` restores the prior pair exactly). The guarantee it provides: the active DB and its active changeset log are always a matched, consistent pair — never a stale DB read against a newer log or vice versa.

`zharness` doesn't have upstream's use case — there is no repository-separation cutover, no two full (DB, log) **generations** to swap between. There is one `harness.db` and one ever-growing `.kit/changesets/` directory per project. The adaptation keeps the guarantee, not the machinery:

- The fence is `meta.last_applied_changeset` (Idempotency Key Rules, layer 1) — everything with ULID ≤ fence is "replayed state," everything above is "pending." This is the direct analog of upstream's active-pair pointer.
- The fence update and the changeset's row mutations happen inside **one SQLite transaction** per `db changeset apply <path>` call: `BEGIN`, apply every line in the file, `UPDATE meta SET last_applied_changeset = ?`, `COMMIT`. SQLite's own WAL gives the all-or-nothing guarantee upstream builds by hand with a journal file and four renames — a crash mid-apply leaves either nothing committed (file is still "pending," safe to retry) or everything committed including the fence (file is "applied," idempotent per R7). No separate `recover forward`/`recover compensate` step is needed, because there is no second generation to reconcile against.
- This logic — transactional apply + co-located fence update — is what lives in `internal/infrastructure` per R2; `internal/domain` and `internal/application` never see partial-apply states.

This is a deliberate scale-down, not a full port: `harness-experimental`'s dual-generation rename dance solves a problem (swapping between two complete DB/log generations during a repository cutover) that `zharness` does not have. Porting the rename+journal machinery as-is would be over-engineering for a single ever-growing log (PLAN.md's own `avoid: premature index/optimization design`).

## Verification

- Nine persisted tables are active: `meta`, seven typed lifecycle tables, and `managed_docs`; each non-meta table has one allowlisted changeset entity mapping.
- Repository replay tests migrate an empty database to the current schema and apply every tracked `.kit/changesets/*.changeset.jsonl` file in ULID order.
