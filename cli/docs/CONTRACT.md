# CLI Contract

`zharness` command surface. Every command supports `--json`; human-readable output is the default without the flag. Exit code `0` = success; non-zero = failure, and every JSON error response carries a stable machine-readable `code` string (`snake_case`) alongside the numeric exit code, so scripts can branch on `code` without parsing text.

> **Resolved 2026-07-17**: SPEC R6 named exactly 19 commands and omitted `handoff`, despite R18 ("handoff records entity") and continuity's Wave 1 T1 depending on one existing. This contract now documents a 20th command, `handoff record` (see "Domain — workflow additions" below), added via a narrow reopening of the already-gated cli-domain phase (`cli-domain-CONTEXT.md`'s "Reopened 2026-07-17" note) rather than a full `brainstorm refine` of R6's count — the underlying entity/table already existed (Phase 4 built them ahead of the gap being hit), only the write-path command was missing.

## Error Codes

| Exit | Meaning |
|---|---|
| 0 | success |
| 1 | validation/user error (bad args, bad fixture, gate FAIL) |
| 2 | system error (db unreadable, fs permission, network for `gh` calls) |

Every non-zero JSON response: `{"error": {"code": "snake_case_string", "message": "human text"}}`.

---

## Core (Phase 3 — cli-core)

Repository coordination is cross-process and repository-root scoped. `preflight`, `query`, `next`, `resume`, `validate`, `audit`, `db changeset status`, and `db status` hold a shared directory-inode lock for the complete read-only SQLite handle lifetime. `init`, `migrate`, `migrate layout` (including dry-run), `import`, `db changeset apply`, `db rebuild`, `intake`, `story`, `intervention`, `trace add`, `decision add`, `run create`, `check record`, and `handoff record` hold the exclusive lock from before DB probing/application validation through SQLite close. Lock acquisition creates no file, times out after five seconds with `repository_lock_timeout` (2), and reports unsupported platforms as `repository_lock_unsupported` (2); Linux and Darwin are supported.

Append-producing mutations allocate canonical changeset filenames strictly above both `meta.last_applied_changeset` and every canonical existing filename. A pending earlier changeset blocks ordinary mutation with `changeset_recovery_required` (1) before any new file or DB change; recover by applying the named earliest file. Direct apply enforces the same earliest-first order. `init` replay remains sorted and exempt, and an apply failure leaves the file pending with its transaction and fence unchanged.

### `id`
- Args: none — mints one fresh ULID without reading or mutating harness state; does not require `init` or a database
- Human output: the ULID followed by a newline
- `--json`: `{"id": "ulid"}`
- Errors: standard Cobra argument error if positional arguments are supplied
- Consumer: playbooks that author artifacts or changeset filenames directly (`work`, `check`, `to-plan`); mint a separate ID for each semantic object and each changeset filename

### `scaffold <run|check|handoff|spec|plan>`
- Args: one artifact kind plus required `--path {destination}`; creates parent directories, fills an absent or empty destination from the embedded skeleton, and refuses to overwrite a non-empty file
- File-only: writes no database row or changeset; the corresponding lifecycle command owns durable registration
- Human output: `scaffolded {kind} → {path}`
- `--json`: `{"kind":"run|check|handoff|spec|plan","path":"...","bytes":N}`
- Errors: `unknown_kind` (1), `missing_required_field` (1, no `--path`), `file_exists` (1, non-empty destination), `scaffold_failed` (2, template/directory/write failure)
- Consumer: workflow playbooks that need an editable lifecycle or planning skeleton before filling its fields and sections

### `preflight <stage>`
- Stages: `brainstorm|to-plan|work|check|handoff|watzup|git|interview`; optional `--mode` is validated per stage
- Read-only: inspects DB/docs presence and docs-version compatibility without creating paths, enabling WAL, writing the database, or appending a changeset
- `--json`: `{"stage":"...","mode":"reduced|durable","version":"...","db":"ready|missing|unreadable","docs":"ready|missing|stale","playbook":"...","readiness":"ready|reduced|blocked","stop":{"code":"...","message":"...","recovery":"..."},"context":{...}|omitted}`; `stop`, `playbook`, and `context` are omitted when not applicable. `version` is the running CLI's own version string (`"dev"` for unreleased builds) — the same value compared against `meta.docs_version` to derive `docs`.
- Reduced stages may continue when DB/docs are missing or stale; durable stages return a named stop and recovery. An unreadable existing DB blocks every stage. `work --mode auto` resolves to durable when at least one non-empty `docs/plans/active/*.md` plan exists and reduced otherwise. `work --mode bounded` and `check --mode bounded` always resolve to reduced preflight and perform zero lifecycle, changeset, or markdown writes. `playbook` is omitted unless the managed docs are ready, so reduced mode never points at a known-missing file.
- `context` (R4, `docs/audit/workflow-harness-ceremony-audit.md`'s F1/F4 findings): a stage-shaped memory packet in place of the resume/query-phases/query-traces menu a stage would otherwise assemble by hand — present for `watzup`, `work`, and `handoff` (and only once the DB is `ready`), and, since R6 (`docs/audit/sdlc-token-cache-audit.md`), for `check` when its resolved mode is `durable` (`gate`/`full`/`auto`/omitted `--mode`) — `check`'s response-only `review`/`bounded`/`simple` modes still never receive one, since they touch no lifecycle position. This supersedes the ceremony audit's own earlier NG2, which kept `check`'s reads entirely separate; `check.md` step 1 now reads position from this packet instead of a standalone `zharness resume --json` call. Shape: `{"position":{"current_phase":"slug"|null,"status":"..."|null},"latest_run_id":"ulid"|null,"latest_check_id":"ulid"|null,"latest_handoff_id":"ulid"|null,"drift":[...],"readiness":"clean"|"in-progress"|"drifted","phases":[...]|omitted,"traces":[...]|omitted,"omitted":[{"field":"...","reason":"...","fetch":"..."}]|omitted}` — `position`/`latest_*_id`/`drift`/`readiness` are exactly `resume --json`'s own fields (the same `Resume` derivation, so the two can never disagree). `phases` (the same shape as `query phases`) is present only for `work`, `handoff`, and durable-mode `check`, whose playbooks reference it; `watzup` never gets it. `traces` (the same shape as `query traces`) is the current phase's own trace history, windowed to the 30 most recent (P4 wave 2's window policy — bounds packet size across phase counts instead of the pre-existing 6.4x-per-phase-history read-cost scaling); absent when there is no current phase. `omitted` (R5) declares any field the packet bounded rather than returning in full, plus an exact command to fetch the rest — populated today only when a phase's trace count exceeds the 30-cap.
- Errors: `invalid_stage` (1), `invalid_mode` (1), `preflight_failed` (2, also covers a failure building `context` for an eligible stage)
- Consumer: every skill under `skills/workflow/`; stage-specific commands still own lifecycle mutation

### `init`
- Args: none; `--force` reinitializes the root database; `--refresh-docs` safely refreshes managed root docs; `--force-docs` explicitly permits overwriting locally changed managed docs
- Creates the single ignored root `harness.db`, replays any existing `.kit/changesets/` before writing new docs metadata, projects `docs/WORKFLOW.md` + `docs/playbooks/**`, updates only the marked `AGENTS.md` block, and appends root DB/WAL/SHM plus `.kit/cache/`/`.kit/conflicts/` ignore entries
- Managed file hashes and docs version live in `managed_docs`. Refresh updates untouched files, preserves local-only edits, and returns `docs_conflict` after staging upstream content under `.kit/conflicts/` when local and embedded content both changed.
- Refuses to create a second database when legacy `.kit/harness.db` exists; recovery is `zharness migrate layout --to v2`
- Playbooks are edited in `cli/docs/embedded/playbooks/` only; root managed docs are the projection checked by `TestProjectionDrift_RootDocsMatchEmbed`
- `--json`: `{"status": "created"|"exists", "db_path": "harness.db", "schema_version": N}`
- Errors: `layout_migration_required` (1), `docs_conflict` (1), `db_not_writable` (2)
- Consumer: durable workflow stages

### `migrate`
- No args: applies pending schema migrations to root `harness.db`; `--json` returns `{"applied": [...], "schema_version": N}`
- `migrate layout --to v2 [--dry-run]`: migrates legacy `.kit/harness.db` by migrating an empty temporary DB, replaying tracked changesets, applying one staged semantic backfill when legacy materialized state is missing from the log, proving normalized parity, safely projecting docs, then atomically activating the root DB and retiring the legacy DB
- Dry-run creates no root DB or managed docs. Apply leaves legacy state active on parity/docs/activation failure.
- `--json` (layout): `{"status":"dry-run|migrated|already-v2","source_db":"...","target_db":"...","replayed":N,"backfilled":N,"parity":true,"docs_written":bool,"dry_run":bool,"schema_version":N}`
- Errors: `invalid_layout` (1), `docs_conflict` (1), `migration_conflict` / `layout_migration_failed` (2)
- Consumer: schema repair and reusable v1→v2 repository migration

### `import`
- Args: `[path]` — defaults to `.kit/`; parses legacy `workflow-state.yml` + planning markdown into changesets + DB rows, per `STATE.md`'s legacy mapping. One-shot idempotent: re-running produces zero new changesets.
- `--json`: `{"imported": N, "skipped": N, "changesets_written": ["ulid.changeset.jsonl", ...]}`
- Errors: `legacy_field_unmapped` (1, a yml field has no destination in STATE.md)
- Consumer: `to-plan` (first run on a legacy project), `pilot-migration` (dogfooding this repo's own `.kit/`)

### `db changeset apply <path>`
- Args: path to a `.jsonl` changeset file; idempotent — re-applying an already-applied changeset is a no-op
- `--json`: `{"applied": N, "skipped_already_applied": N}`
- Errors: `changeset_malformed` (1), `changeset_out_of_order` (1, ULID predates the last-applied changeset — replay only, not a hard block, but flagged), `changeset_recovery_required` (1, an earlier canonical pending file must be applied first)
- Consumer: `continuity` cross-machine resume (rebuild db from committed changesets on a fresh clone)

### `db changeset status`
- Args: none
- `--json`: `{"pending": ["ulid.changeset.jsonl", ...], "applied_count": N, "last_applied": "ulid", "unverified_below_fence": ["ulid.changeset.jsonl", ...]}`
- `unverified_below_fence` names changesets at or below the fence whose create-op entity ids are missing from their tables — proof they were never actually applied despite the fence assuming order-based coverage (F5, `docs/audit/workflow-harness-ceremony-audit.md`). Update-only changesets cannot be verified this way and are never listed here. Empty in the ordinary single-fence-advance case; non-empty is a signal to run `db rebuild`.
- Errors: `db_unreadable` (2)
- Consumer: `continuity` (verify rebuild completeness), `watzup` (indirect, via `resume`)

### `db rebuild --yes`
- Args: `--yes` required — without it, the command refuses with `confirmation_required` and mutates nothing
- Deletes `harness.db` and its `-wal`/`-shm` sidecars if present, re-migrates to the current schema, then replays every changeset under `.kit/changesets/` in ULID order from empty. Database-only: unlike `init`, it never touches `docs/`, `AGENTS.md`, or `.gitignore`.
- `--json`: `{"status":"rebuilt","schema_version":N,"replayed":N}`
- Errors: `confirmation_required` (1), `db_not_writable` (2), `changeset_replay_failed` (2)
- Consumer: recovery after `db changeset status`/`db status` reports `unverified_below_fence`, or after any suspected DB/changeset divergence

### `db status`
- Args: none
- `--json`: `{"schema_version":N,"fence":"ulid","rows":{"stories":N,"runs":N,"checks":N,"handoffs":N,"intakes":N,"interventions":N,"traces":N,"managed_docs":N},"pending":[...],"unverified_below_fence":[...],"context_cost_estimate":{"active_plan_path":"...","active_plan_bytes":N,"stages":{"watzup":{"playbook_bytes":N,"estimated_tokens_today":N}, ...},"note":"..."}}`
- `rows` is introspected from the schema (every table except `meta`), so a later migration's new table appears automatically without a contract change here.
- `context_cost_estimate` reports, per spine stage, the byte size of that stage's playbook plus the current active plan — the `bytes/4` heuristic documented in `docs/audit/workflow-harness-ceremony-audit.md` — answering "how expensive would resuming be right now" (addresses that audit's G4). It reflects today's full-plan-read path, not the compressed-index read path this initiative is adding.
- Errors: `db_unreadable` (2)
- Consumer: an agent or operator checking harness health before deciding whether `db rebuild` is warranted

### `query <view>`
- Views: `state`, `phases`, `artifacts`, `check --latest`, `checks`, `traces`, `decisions`, `handoff --latest`, `plan --section {current-state|phase}`
- Args: `state` (no args); `phases` (no args, lists all stories + status); `artifacts` (`--phase {slug}` optional filter); `check --latest` (`--latest` flag, returns most recent check verdict); `checks` (`--phase {slug}` optional filter joining through `runs.story_slug`, `--tail {N}` optional cap, 0/omitted = unbounded); `traces` (`--phase {slug}` optional filter joining through `runs.story_slug`, or `--run-id {ulid}` optional filter to one run — mutually exclusive, `--phase` wins if both are given; `--tail {N}` optional cap on the most recent entries, 0/omitted = unbounded); `decisions` (`--phase {slug}` optional filter, `--tail {N}` optional cap, 0/omitted = unbounded); `handoff --latest` (`--latest` flag, returns most recent handoff's anchors flattened); `plan` (`--section {current-state|phase}` required; `phase` also requires `--phase {slug}`) — the only view that reads no database: it resolves the single plan under `docs/plans/active/*.md` and slices its markdown directly
- `--json` (`state`): `{"current_phase":"slug"|null,"entry_phase":"slug"|null,"schema_version":N,"latest_run_id":"ulid"|null,"latest_check_id":"ulid"|null}`
- `--json` (`phases`): `[{"slug":"...","goal":"...","status":"planned|in-progress|checked|done","depends_on":"slug"|null,"created_at":"..."}, ...]`
- `--json` (`artifacts`): `[{"id":"ulid","story_slug":"slug","artifact_path":"","created_at":"..."}, ...]`; `artifact_path` is optional/deprecated lifecycle metadata encoded as a string that may be empty and is never resolved on disk
- `--json` (`check --latest`): `{"id":"ulid","verdict":"APPROVED"|"APPROVE_WITH_REQUESTS"|"REQUEST_CHANGES","phase":"slug","judge":"independent"|"same-session"|null,"judge_model":"..."|null}`; `judge`/`judge_model` are `null` for a check recorded before the eval-layer initiative
- `--json` (`checks`): `[{"id":"ulid","run_id":"ulid","phase":"slug","verdict":"APPROVED"|"APPROVE_WITH_REQUESTS"|"REQUEST_CHANGES","judge":"independent"|"same-session"|null,"judge_model":"..."|null,"created_at":"..."}, ...]` in chronological order; the compressed-index counterpart of a plan's `## Validation` entries — `check --latest`'s locked shape is unchanged and unaffected by this view (P6, docs/audit/workflow-harness-ceremony-audit.md's own success-signal verification: every append-only markdown section needs a query returning its full compressed form, not only its latest entry)
- `--json` (`traces`): `[{"id":"ulid","run_id":"ulid"|null,"wave":N,"summary":"...","task":"..."|null,"task_status":"DONE"|"DONE_WITH_CONCERNS"|"NEEDS_CONTEXT"|"BLOCKED"|null,"created_at":"..."}, ...]` in chronological order; the compressed-index counterpart of a plan's `## Progress` entries (see `docs/audit/workflow-harness-ceremony-audit.md`); `task`/`task_status` are `null` for a wave-level trace recorded before migration `0008_trace_task_granularity` or without `--task`/`--task-status`
- `--json` (`decisions`): `[{"id":"ulid","run_id":"ulid"|null,"phase":"slug"|null,"task":"..."|null,"decision":"...","rationale":"...","created_at":"..."}, ...]` in chronological order; the compressed-index counterpart of a plan's `## Decisions` markdown entries
- `--json` (`handoff --latest`): `{"id":"ulid","run_id":"ulid"|null,"check_id":"ulid"|null,"open_items":["...",...],"exact_next_action":"..."|null,"created_at":"..."}`; flattens the most recent handoff's `anchors` JSON column — the read half of `handoff record --next-action`'s round trip
- `--json` (`plan`): `{"path":"docs/plans/active/{slug}.md","section":"current-state"|"phase","phase":"slug"|omitted,"content":"...","degraded":false}`; `content` is the requested markdown slice verbatim (`current-state`: the `## Current State and Next Action` body; `phase`: the phase's block within `## Phases and Verification`, matched either as a `### phase_slug: \`{slug}\`` heading or the scaffold template's `  - phase_slug: {slug}` list item — heading form wins if a plan somehow has both) with no reformatting. `degraded:true` means neither form's block (nor, for `current-state`, the section) could be found — `content` then carries the whole plan file instead of failing the call, so a malformed hand-edited plan degrades the caller to a full read rather than blocking it (P3, `docs/audit/workflow-harness-ceremony-audit.md`; list-form matching added by R1, `docs/audit/sdlc-token-cache-audit.md`'s F2, after the heading-only regex was found to degrade on every plan `to-plan` fills in without headings)
- Errors: `unknown_view` (1), `no_check_found` (1), `no_handoff_found` (1, `handoff --latest` with zero handoff rows), `unknown_phase` (1, `--phase` given to `checks`/`traces`/`decisions` naming no story row — distinct from an empty result, which means the phase exists but has no rows yet; R2, `docs/audit/sdlc-gap-analysis.md`'s C4), `db_unreadable` (2); `plan` view only: `unknown_section` (1, `--section` not `current-state`/`phase`), `missing_required_field` (1, `--section phase` without `--phase`), `no_active_plan` (1, no non-empty file under `docs/plans/active/*.md`), `ambiguous_active_plan` (1, more than one), `plan_unreadable` (2, filesystem error other than the file simply not existing)
- Consumer: `to-plan` (phase status), `git` (`query check --latest`, warn-not-block on FAIL/missing), `continuity`, `watzup`/`work`/`handoff` (`traces`/`decisions`/`handoff --latest`, reading wave history, prior decisions, and the last recorded next action without opening the plan file)

---

## Domain — 4 ported (Phase 4 — cli-domain)

### `intake`
- Args: `--type {new-spec|spec-slice|change-request|new-initiative|maintenance|harness-improvement} --summary "..." --lane {tiny|normal|high-risk} [--plan-path docs/plans/active/{slug}.md] [--plan-id {ulid}]`; `--plan-path` is optional repository-relative metadata for the initiative's evolving plan
- `--plan-id`, added by migration `0009_intake_plan_id`, is optional and — when supplied — must be the same plan ULID passed to `run create --plan-id` for that initiative's runs. It enables `check record` to resolve a run's lane and gate `--judge` for `high-risk` (G2, `docs/audit/workflow-harness-ceremony-audit.md`/V2). It carries no FK, matching `runs.plan_id`'s own precedent.
- `--json`: `{"id": "ulid"}`
- Errors: `invalid_lane` (1), `invalid_type` (1), `missing_required_field` (1), `invalid_plan_id` (1, `--plan-id` given but not a valid ULID)
- Consumer: `brainstorm` (fires at SPEC lock; ULID written to SPEC frontmatter `intake_id`)

### `story`
- Args: `--slug {phase-slug} --goal "..." [--depends-on {slug}]`
- `--json`: `{"id": "ulid", "slug": "...", "status": "planned"}`
- Errors: `duplicate_slug` (1, story already exists for this phase), `unknown_dependency` (1)
- Consumer: `to-plan` (one story per roadmap phase, slug = phase slug — locked decision, Phase 1 gap-matrix)

### `intervention`
- Args: `--verdict-id {ulid} --reason "..."`
- `--json`: `{"id": "ulid"}`
- Errors: `unknown_verdict_id` (1)
- Consumer: `check` (documented escalation path when a human overrides a FAIL verdict — validation-gate T3 step 4; not auto-invoked, a human decision)

### `trace`
- Args (single-entry form): `trace add --wave N --summary "..." [--run-id {ulid}] [--task "..."] [--task-status DONE|DONE_WITH_CONCERNS|NEEDS_CONTEXT|BLOCKED]`
- `--task` and `--task-status` are both optional and independent of each other: a wave-level trace (fires once per wave) omits both; a task-level trace (added by migration `0008_trace_task_granularity` to close G1, `docs/audit/workflow-harness-ceremony-audit.md` — a mid-wave interruption used to leave the index blind to tasks the plan's `## Progress` markdown already recorded) sets both, one call per attempted task, matching `docs/playbooks/work.md`'s Status Routing values.
- `--json`: `{"id": "ulid"}`
- Args (batch form): `trace add --wave N --tasks '[{"task":"...","task_status":"DONE|DONE_WITH_CONCERNS|NEEDS_CONTEXT|BLOCKED","summary":"..."}, ...]' [--run-id {ulid}]` — 1-20 entries, mutually exclusive with `--task`/`--task-status`/`--summary`. Added for R5 (`docs/audit/sdlc-token-cache-audit.md`): several tasks completing within one wave previously cost one round trip each even though most land clean; the batch flushes them in one call. Every element is a task-level entry — `task`, `task_status`, and `summary` are all required per element, unlike the single-entry form's optional task fields.
- `--json` (batch form): `[{"id": "ulid"}, ...]` in input order
- Errors: `unknown_run_id` (1, if `--run-id` given but not found), `invalid_task_status` (1, `--task-status`/an element's `task_status` given but not one of the four values), `empty_tasks` (1, `--tasks` is `[]` or omitted with `--tasks` set), `too_many_tasks` (1, more than 20 elements), `invalid_tasks` (1, `--tasks` is malformed JSON), `missing_required_field` (1, a batch element is missing `task` or `summary`), `invalid_arguments` (1, `--tasks` combined with `--task`/`--task-status`/`--summary`), `internal_error` (1, exactly one active plan exists but is missing its `## Progress` heading — not a named contract code, since it signals a malformed plan file rather than bad command input)
- Atomic side effect: when exactly one plan is `active` under `docs/plans/active/`, appends a matching entry (one per batch element, in the batch case) to its `## Progress` section in the same operation as the DB write — the entry text is computed and the section's writability validated before the DB write, so the common failure (missing section) has zero side effects; with zero or multiple active plans, the markdown write is a no-op (P3, `docs/audit/workflow-harness-ceremony-audit.md`'s "DB is a compressed index, not a copy" mental model). A batch write is all-or-nothing: one invalid element fails the whole call before any row or markdown line is written.
- Consumer: `work` (fires at each wave completion, or per task for finer-grained history — batched via `--tasks` when several tasks complete within the same wave; RUN artifact frontmatter carries the returned `trace_ids`)

### `decision`
- Args: `decision add --decisions '[{"decision":"...","rationale":"...","phase":"slug","task":"..."}, ...]' [--run-id {ulid}]`
- `--decisions` is a JSON array, one object per decision; `decision` and `rationale` are required per object, `phase` and `task` are optional. At least one array element is required. `--run-id` is optional and shared across the whole batch, matching `trace add`'s `--run-id` pattern — this is a batch of decisions produced by one unit of work, not decisions from different runs.
- Batching exists so a wave surfacing several decisions costs one call, not one per decision (docs/audit/workflow-harness-ceremony-audit.md's ceremony finding).
- `--json`: `{"ids": ["ulid", ...]}`, one id per decision in array order
- Errors: `empty_decisions` (1, `--decisions` is `[]` or omitted), `invalid_decisions` (1, malformed JSON), `missing_required_field` (1, an element is missing `decision` or `rationale`), `unknown_run_id` (1, `--run-id` given but not found), `unknown_phase` (1, an element's `phase` slug not found), `internal_error` (1, exactly one active plan exists but is missing its `## Decisions` heading)
- Atomic side effect: when exactly one plan is `active` under `docs/plans/active/`, appends one formatted line per batch element to its `## Decisions` section in the same operation as the DB write; with zero or multiple active plans, the markdown write is a no-op (P3, same "one writer" pattern as `trace add`)
- Consumer: `work`/`to-plan`/`check` (recording a plan gap, trade-off, deviation, or wrong assumption discovered during execution — the compressed-index counterpart of a plan's `## Decisions` markdown section, re-adding the `decisions` table migration `0003_drop_dead_surface` dropped as unwritten dead surface; see `docs/audit/workflow-harness-ceremony-audit.md`)

## Domain — workflow additions (Phase 4 — cli-domain)

### `run create`
- Args: `--slug {phase-slug} [--artifact-path {legacy run file path}] [--plan-id {ulid}]` — full-mode only; only `--slug` is required, while `--artifact-path` is optional/deprecated metadata and may be omitted
- `--json`: `{"id": "ulid"}`
- Errors: `missing_required_field` (1, missing `--slug`), `invalid_plan_id` (1, non-empty `--plan-id` is not a strict ULID), `unknown_story` (1, `--slug` has no matching story), `story_not_runnable` (1, story is already `checked` or `done`), `db_unreadable` / `db_not_writable` (2)
- Atomic side effects: for a `planned` or `in-progress` story, creates the run, leaves/moves the story at `in-progress`, points `meta.latest_run_id` at the new run, and points `meta.current_phase` at the story slug in one changeset/transaction
- Consumer: `work` full mode; bounded/reduced work performs preflight only and skips run, story, pointer, changeset, and markdown writes

### `resume`
- Args: none
- `--json`: `{"position":{"current_phase":"..."|null,"status":"..."|null},"latest_run_id":"ulid"|null,"latest_check_id":"ulid"|null,"latest_handoff_id":"ulid"|null,"drift":[{"type":"unknown_phase"|"out_of_order"|"stale_docs","detail":"...","recovery":"..."}],"readiness":"clean"|"in-progress"|"drifted"|"no-harness"}`
- Lifecycle drift is derived from database stories, run/check links, meta pointers, and managed-docs version. Legacy lifecycle artifact paths are metadata only and are never resolved or existence-checked on disk.
- Errors: `db_unreadable` (2, distinct from `readiness: no-harness` — that's a valid successful response, not an error)
- Consumer: `watzup` (renders 1:1 from this snapshot, no independent prose re-derivation)

### `check record`
- Args: `--verdict {APPROVED|APPROVE_WITH_REQUESTS|REQUEST_CHANGES} --run-id {ulid} --judge {independent|same-session} --judge-model {model identifier} --proof-links '[{"command":"...","output_ref":"...","artifact_path":"..."}]'`; each proof link's `artifact_path` is optional/deprecated legacy metadata and is not a filesystem requirement
- `--judge` and `--judge-model` are required for every verdict, including `REQUEST_CHANGES` — declares whether the verdict was produced by an independent judge or by the same session that authored the diff under review, and which model produced it (eval-layer initiative)
- **Proof verification (`docs/audit/workflow-harness-ceremony-audit.md` §15):** for `APPROVED` and `APPROVE_WITH_REQUESTS` only, every `proof_links[].command` is re-executed (`sh -c`, cwd = the process's own, 5-minute timeout per command) and must exit 0 before the check is recorded — a failure rejects the whole call before any DB write, naming the failing command and its captured output tail. Only the exit code is checked, not `output_ref`/stdout text, since a command's exact output is not guaranteed identical between runs. `REQUEST_CHANGES` proof is never re-executed: it commonly cites a failing command as the evidence of the problem being reported, so requiring exit 0 there would reject exactly the proof it needs to carry. Proof commands must therefore be safe to run a second time (tests, builds, lints — not anything with a side effect that shouldn't repeat).
- `--json`: `{"id":"ulid","verdict":"..."}`
- Errors: `unknown_run_id` (1), `story_not_checkable` (1, story is not `in-progress`), `run_not_latest` (1, run is stale for its story), `invalid_verdict` (1), `invalid_judge` (1, `--judge` is not `independent`/`same-session`), `independent_judge_required` (1, the run resolves via `runs.plan_id` -> `intakes.plan_id` to a `high-risk` lane and `--judge` is not `independent` — G2, `docs/audit/workflow-harness-ceremony-audit.md`/V2; unresolvable or non-`high-risk` lanes are unaffected), `invalid_proof_links` (1, malformed `--proof-links` JSON, or any entry whose `command` is empty/whitespace-only — checked regardless of verdict, since a blank command isn't proof of anything whether or not it's re-executed), `empty_proof_links` (1, verdict other than REQUEST_CHANGES with zero proof links), `proof_verification_failed` (1, `APPROVED`/`APPROVE_WITH_REQUESTS` only — a proof command exited non-zero on re-execution), `missing_required_field` (1, empty `--run-id` or `--judge-model`), `internal_error` (1, exactly one active plan exists but is missing its `## Validation` heading), `db_unreadable` / `db_not_writable` (2)
- Atomic side effects: for the latest run of an `in-progress` story, records the check and points `meta.latest_check_id` at it; `APPROVED` and `APPROVE_WITH_REQUESTS` move the story to `checked`, while `REQUEST_CHANGES` leaves it `in-progress`. When exactly one plan is `active` under `docs/plans/active/`, also appends a matching entry (verdict, judge, phase, nested proof-link sub-bullets) to its `## Validation` section; the section's writability is checked before the DB write, so the common failure (missing section) has zero side effects. With zero or multiple active plans, the markdown write is a no-op (P3, same "one writer" pattern as `trace add`/`decision add`)
- Consumer: `check` (deterministic — no free-text-only verdicts, R19)

### `handoff record`
- Args: `[--run-id {ulid}] [--check-id {ulid}] [--open-items '["...","..."]'] [--next-action "..."] [--close-phase]`; anchors are optional for a non-closing handoff and `--open-items` defaults to `[]`
- `--next-action` is optional and, when supplied, persists the plan's Current State `exact_next_action` into `anchors.exact_next_action` — no migration needed, `anchors` is already free-form JSON (docs/audit/workflow-harness-ceremony-audit.md, D1). Readable back via `query handoff --latest`.
- `--close-phase` requires both anchors, requires `--run-id` to be the latest run for its story, requires `--check-id` to be the latest check for that run and to carry a clean verdict (`APPROVED` or `APPROVE_WITH_REQUESTS`), and requires the story to be `checked`; a successful close moves that story to `done` in the same changeset/transaction
- `--json`: `{"id":"ulid"}`
- Errors: `invalid_open_items` (1, malformed JSON or an empty-string entry), `unknown_run_id` (1), `unknown_check_id` (1), `missing_required_field` (1, closing without both anchors), `check_run_mismatch` (1), `check_not_clean` (1), `run_not_latest` (1), `check_not_latest` (1), `phase_not_checked` (1), `internal_error` (1, exactly one active plan exists but is missing its `## Progress` heading), `db_unreadable` / `db_not_writable` (2)
- The CLI durably records DB anchors and open items; it does not write `.kit/HANDOFF.md`. When exactly one plan is `active` under `docs/plans/active/`, it additionally appends an event-log entry (handoff id, run, check, next action, open items) to that plan's `## Progress` section — not a rewrite of the snapshot-style `## Current State and Next Action` section, which the P3 markdown writer does not touch. With zero or multiple active plans, the markdown write is a no-op (P3, same "one writer" pattern as `trace add`/`decision add`/`check record`)
- Consumer: `handoff` skill; `resume`'s `latest_handoff_id` reads the latest durable record unchanged

### `validate`
- Args: none — validates the durable lifecycle graph in the database: lifecycle entity ULIDs, story dependencies, story→run→check→handoff links, compatible handoff run/check anchors, and meta pointers
- `--json`: `{"valid":true|false,"findings":[{"link":"...","issue":"missing_key"|"broken_link"|"stale_pointer","detail":"..."}]}`
- A missing database returns the same envelope with `valid:false`, a `DB->LIFECYCLE` finding, and exit 1. An unreadable or unqueryable database returns `db_unreadable` (2).
- RUN, CHECK, and HANDOFF markdown files are not required; legacy artifact paths and frontmatter chains are not read or existence-checked
- Consumer: lifecycle integrity checks and `audit`'s `contract_violations` composition

---

## Research (Phase 6 — validation-gate)

### `audit`
- Args: none — composes the database-backed `resume` drift reader with `validate` lifecycle/link findings
- `--json`: `{"pointer_drift":[...],"contract_violations":[...],"unlinked_proofs":[]}`, stable ordering; `unlinked_proofs` remains an always-empty compatibility field
- Proof-link and check artifact paths are optional/deprecated metadata and are not existence-checked
- Errors: `db_unreadable` (2)
- Consumer: `check` (wired into gate flow)

---

## Workflow-Step → CLI-Action Mapping

| Skill step | CLI action |
|---|---|
| brainstorm: SPEC lock | `intake --type … --summary … --lane … [--plan-path docs/plans/active/{slug}.md] --json` |
| to-plan: roadmap init | `init` (if no db) |
| to-plan: per-phase | `story --slug {phase} --goal … --json` |
| to-plan: status render | `query state --json`, `query phases --json` |
| work: bounded execution | `preflight work --mode bounded --json`; no lifecycle entity, changeset, or markdown artifact is written |
| work: run registration (full mode) | `run create --slug {phase} [--artifact-path …] [--plan-id …] --json` |
| work: wave completion | `trace add --wave N --summary … --json` |
| check: bounded review | `preflight check --mode bounded --json`; no lifecycle entity, changeset, or markdown artifact is written |
| check: gate evaluation | `audit --json` → matrix → `check record --verdict … --json` |
| check: human override | `intervention --verdict-id … --reason …` (manual, documented escalation only) |
| watzup: recap render | `resume --json` (1:1, no fallback) |
| handoff: close phase | `handoff record --run-id … --check-id … --open-items … --close-phase --json` |
| git: pre-commit/PR | `query check --latest --json` (warn-not-block on FAIL/missing) |
| continuity: fresh-machine rebuild | `db changeset apply <path>` (per committed changeset, ULID order) |
