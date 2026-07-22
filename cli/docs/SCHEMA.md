# DB + Changeset Schema v1

SQLite schema for `harness.db` and the changeset line/file format that reproduces it. Every mutating command in `CONTRACT.md` writes one changeset before touching the DB (R7); replaying all changesets from empty yields identical state.

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
| `plan_id` | TEXT, nullable | frontmatter PLAN→SPEC cross-link; **known gap** — `phase-plan-template.md` doesn't carry `spec_id` yet (see CONTRACT.md `validate`), so this column exists but stays unpopulated until that template is updated |
| `trace_ids` | TEXT | JSON array of trace ULIDs, appended per `trace add` call |
| `artifact_path` | TEXT | `.kit/runs/work/{...}.md` |
| `created_at` | TEXT | |

#### `checks`
| Column | Type | Notes |
|---|---|---|
| `id` | TEXT PK | ULID |
| `run_id` | TEXT | FK `runs.id` |
| `verdict` | TEXT | enum `APPROVED\|APPROVE_WITH_REQUESTS\|REQUEST_CHANGES` |
| `proof_links` | TEXT | JSON array of `{command, output_ref, artifact_path}` |
| `artifact_path` | TEXT, nullable | `.kit/reports/check/{...}.md`, when persisted |
| `created_at` | TEXT | |

#### `handoffs`
| Column | Type | Notes |
|---|---|---|
| `id` | TEXT PK | ULID |
| `run_id` | TEXT, nullable | FK `runs.id` |
| `check_id` | TEXT, nullable | FK `checks.id` |
| `anchors` | TEXT | JSON `{latest_run_id, latest_check_id, open_items}` |
| `created_at` | TEXT | human `.kit/HANDOFF.md` path is fixed by convention (STATE.md), not stored here |

### Ported entities

#### `intakes`
| Column | Type | Notes |
|---|---|---|
| `id` | TEXT PK | ULID |
| `type` | TEXT | enum `new-spec\|spec-slice\|change-request\|new-initiative\|maintenance\|harness-improvement` |
| `summary` | TEXT | |
| `lane` | TEXT | enum `tiny\|normal\|high-risk` |
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
| `created_at` | TEXT | |

## Table ↔ Changeset Entity Type

Every table maps to exactly one changeset `entity` string (the value in `{op, entity, id, fields, at}`). Table names are plural (SQL convention, SPEC R13's own wording); entity strings are singular, matching the CONTRACT.md command that produces them.

| Table | Entity string | Producing command(s) |
|---|---|---|
| `meta` | `meta` | side-effect of any command that updates `current_phase`/`latest_run_id`/etc. — never created directly |
| `stories` | `story` | `story`, plus `status` updates from `work`/`check`/`handoff` (STATE.md Writer/Reader Ownership) |
| `runs` | `run` | `trace add` (run completion), `work` |
| `checks` | `check` | `check record` |
| `handoffs` | `handoff` | **none yet** — SPEC R13 mandates this table, but no command among R6's 19 produces it (R6/R18 gap, see `CONTRACT.md`'s escalation note); table exists now so the schema doesn't need a breaking migration once the command lands |
| `intakes` | `intake` | `intake` |
| `interventions` | `intervention` | `intervention` |
| `traces` | `trace` | `trace add` |

Cross-check against CONTRACT.md: every mutating command listed there (`intake, story, intervention, trace, check record, handoff record`, plus `init`/`import`/`migrate` which write `meta` only) names exactly one entity string from this table. `resume`, `query`, `validate`, `audit` are read-only — they write no changeset.

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

- Every table above has exactly one row in Table ↔ Changeset Entity Type: 8 tables (`meta` + 7 data tables), 8 entity-string mappings — count matches. `handoffs` is the one row with no current producing command (see note above); the table is still defined now per SPEC R13, so no breaking migration is needed once R6/R18's gap is resolved.
- Cross-check against CONTRACT.md's 13 commands: entity-producing (`intake, story, intervention, trace, check record` = 5) + `meta`-only (`init, import, migrate` = 3) + replay/special (`db` — applies existing changesets, doesn't mint a new entity type = 1) + read-only (`resume, query, validate, audit` = 4) = 5 + 3 + 1 + 4 = **13**, matching CONTRACT.md exactly (post `scoring-removal`; the pre-existing `handoff record` undercount noted above predates this phase and is unrelated).
