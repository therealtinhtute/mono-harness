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

### `id`
- Args: none — mints one fresh ULID without reading or mutating harness state; does not require `init` or a database
- Human output: the ULID followed by a newline
- `--json`: `{"id": "ulid"}`
- Errors: standard Cobra argument error if positional arguments are supplied
- Consumer: playbooks that author artifacts or changeset filenames directly (`work`, `check`, `to-plan`); mint a separate ID for each semantic object and each changeset filename

### `init`
- Args: none; `--force` reinitializes an empty db if one already exists (refuses on a non-empty db without `--force`); `--refresh-docs` forces `.kit/docs/**` to be rewritten from the embedded doc set and `docs_version` re-stamped even if docs already exist — canonical overwrite, never a merge with user edits; independent of `--force` and the db's own status
- Side effects beyond the db, run on every invocation and each individually idempotent (a repeat call with nothing to do performs zero writes): creates `.kit/` if missing; scaffolds `.kit/docs/**` from the CLI's embedded doc set if `.kit/docs` doesn't already exist or `--refresh-docs` was passed, and stamps `meta.docs_version` (the CLI's own version, `"dev"` for unreleased builds) whenever it changes; writes a root `AGENTS.md` shim only if the repo has none (never overwrites an existing one — prints a notice naming the canonical copy under `.kit/docs/AGENTS.md` instead); appends any of `.kit/harness.db`, `.kit/cache/` missing from the root `.gitignore` (append-only, existing content untouched)
- `--json`: `{"status": "created"|"exists", "db_path": "...", "schema_version": N}` — unchanged by the side effects above; they are not reflected in JSON output
- Errors: `db_not_writable` (2)
- Consumer: `to-plan` (creates db if absent, before first `story`)

### `migrate`
- Args: none — applies all pending versioned migrations to the current schema_version
- `--json`: `{"applied": ["0002_...", ...], "schema_version": N}`
- Errors: `migration_conflict` (2, schema_version ahead of binary's known migrations)
- Consumer: internal / invoked by `init` implicitly, exposed standalone for post-upgrade repair

### `import`
- Args: `[path]` — defaults to `.kit/`; parses legacy `workflow-state.yml` + planning markdown into changesets + DB rows, per `STATE.md`'s legacy mapping. One-shot idempotent: re-running produces zero new changesets.
- `--json`: `{"imported": N, "skipped": N, "changesets_written": ["ulid.changeset.jsonl", ...]}`
- Errors: `legacy_field_unmapped` (1, a yml field has no destination in STATE.md)
- Consumer: `to-plan` (first run on a legacy project), `pilot-migration` (dogfooding this repo's own `.kit/`)

### `db changeset apply <path>`
- Args: path to a `.jsonl` changeset file; idempotent — re-applying an already-applied changeset is a no-op
- `--json`: `{"applied": N, "skipped_already_applied": N}`
- Errors: `changeset_malformed` (1), `changeset_out_of_order` (1, ULID predates the last-applied changeset — replay only, not a hard block, but flagged)
- Consumer: `continuity` cross-machine resume (rebuild db from committed changesets on a fresh clone)

### `db changeset status`
- Args: none
- `--json`: `{"pending": ["ulid.changeset.jsonl", ...], "applied_count": N, "last_applied": "ulid"}`
- Errors: `db_unreadable` (2)
- Consumer: `continuity` (verify rebuild completeness), `watzup` (indirect, via `resume`)

### `query <view>`
- Views: `state`, `phases`, `artifacts`, `check --latest`
- Args: `state` (no args); `phases` (no args, lists all stories + status); `artifacts` (`--phase {slug}` optional filter); `check --latest` (`--latest` flag, returns most recent check verdict)
- `--json` (`state`): `{"current_phase": "slug"|null, "entry_phase": "slug"|null, "schema_version": N, "latest_run_id": "ulid"|null, "latest_check_id": "ulid"|null}`
- `--json` (`check --latest`): `{"id": "ulid", "verdict": "APPROVED"|"APPROVE_WITH_REQUESTS"|"REQUEST_CHANGES", "phase": "slug"}`
- Errors: `unknown_view` (1), `db_unreadable` (2)
- Consumer: `to-plan` (phase status), `git` (`query check --latest`, warn-not-block on FAIL/missing), `continuity`

---

## Domain — 4 ported (Phase 4 — cli-domain)

### `intake`
- Args: `--type {new-spec|spec-slice|change-request|new-initiative|maintenance|harness-improvement} --summary "..." --lane {tiny|normal|high-risk}`
- `--json`: `{"id": "ulid"}`
- Errors: `invalid_lane` (1), `invalid_type` (1)
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
- Args: `trace add --wave N --summary "..." [--run-id {ulid}]`
- `--json`: `{"id": "ulid"}`
- Errors: `unknown_run_id` (1, if `--run-id` given but not found)
- Consumer: `work` (fires at each wave completion; RUN artifact frontmatter carries the returned `trace_ids`)

## Domain — workflow additions (Phase 4 — cli-domain)

### `run create`
- Args: `--slug {phase-slug} --artifact-path {run file path} [--plan-id {ulid}]` — full-mode only
- `--json`: `{"id": "ulid"}`
- Errors: `missing_required_field` (1, `--slug`/`--artifact-path`), `unknown_story` (1, `--slug` has no matching story)
- Consumer: `work` (full mode only; mints the run row and points `meta.latest_run_id` at it atomically, one changeset/tx — simple mode still skips DB registration entirely, `runs.story_slug` has no row for it to reference)

### `resume`
- Args: none
- `--json`: `{"position": {"current_phase": "...", "status": "..."}, "latest_run_id": "ulid"|null, "latest_check_id": "ulid"|null, "latest_handoff_id": "ulid"|null, "drift": [{"type": "missing_file"|"unknown_phase"|"out_of_order"|"stale_docs", "detail": "...", "recovery": "..."}], "readiness": "clean"|"in-progress"|"drifted"|"no-harness"}`
- Errors: `db_unreadable` (2, distinct from `readiness: no-harness` — that's a valid successful response, not an error)
- Consumer: `watzup` (renders 1:1 from this snapshot, no independent prose re-derivation)

### `check record`
- Args: `--verdict {APPROVED|APPROVE_WITH_REQUESTS|REQUEST_CHANGES} --run-id {ulid} --proof-links '[{"command":"...","output_ref":"...","artifact_path":"..."}]'`
- `--json`: `{"id": "ulid", "verdict": "..."}`
- Errors: `unknown_run_id` (1), `invalid_verdict` (1), `empty_proof_links` (1, verdict other than REQUEST_CHANGES with zero proof links)
- Consumer: `check` (deterministic — no free-text-only verdicts, R19)
- Side effect: also points `meta.latest_check_id` at the new check row, atomically in the same changeset/tx (no separate manual pointer update needed)

### `handoff record`
- Args: `--run-id {ulid} --check-id {ulid} --open-items '["...","..."]'` — `--run-id`/`--check-id` optional (anchors are nullable pointers), `--open-items` defaults `[]`
- `--json`: `{"id": "ulid"}`
- Errors: `invalid_open_items` (1, malformed JSON or an empty-string entry), `unknown_run_id` (1, `--run-id` given but not found), `unknown_check_id` (1, `--check-id` given but not found)
- Consumer: `handoff` skill (close-out — writes both this entity and the human-readable `.kit/HANDOFF.md`); `resume`'s `latest_handoff_id` reads it back unchanged

### `validate`
- Args: none — walks SPEC→PLAN→RUN→CHECK→HANDOFF by frontmatter ULID cross-links
- `--json`: `{"valid": true|false, "findings": [{"link": "SPEC->PLAN", "issue": "missing_key"|"broken_link"|"stale_pointer"|"not_yet_implemented", "detail": "..."}]}`
- Errors: exits 1 (`!valid`) when any finding's `issue` is not `not_yet_implemented` — a `not_yet_implemented`-only findings list is a passing (`valid: true`) response, not a violation
- Consumer: `skill-adapters` T4 sample-chain proof, `pilot-migration` evidence bundle
- **Known gap**: the `SPEC->PLAN` link cannot currently be validated — `phase-plan-template.md` was not in harness-contracts' Allowed Surfaces, so PLAN artifacts don't yet carry a `spec_id` field. `validate` skips that one link with a `not_yet_implemented` finding rather than a hard failure, until a later phase adds it (see `.kit/implementation-notes.md`).
- **Mode-aware carve-out** (`harness-mode-parity`, GitHub #38): RUN and CHECK artifacts carry a `mode: full|simple` frontmatter field (`work.md`/`check.md`). `mode: simple` artifacts are phase-less, plan-less, and never DB-registered by design — `runs.story_slug` and `checks.run_id` are both `NOT NULL` FKs with no story/run to reference in simple mode, so `work`/`check` skip DB registration entirely rather than crash. `validate` reflects this: a `mode: simple` RUN is exempt from the `phase`-existence check and the `plan_id` ULID requirement; a `mode: simple` RUN or CHECK is exempt from its DB stale-pointer check. `id` remains a required, well-formed ULID unconditionally — the carve-out is about DB/phase/plan linkage, not artifact hygiene. Missing `mode` (artifacts predating this field) or `mode: full` keeps every check at full strictness, unchanged from before this phase.

---

## Research (Phase 6 — validation-gate)

### `score-trace <trace-id>`
- Args: trace ULID
- `--json`: `{"tier": "...", "reasons": ["..."]}`
- Errors: `unknown_trace_id` (1)
- Consumer: `check` (wired into gate flow: `audit` + `score-trace` → matrix evaluation → `check record`)

### `audit`
- Args: none
- `--json`: `{"pointer_drift": [...], "contract_violations": [...], "unlinked_proofs": [...], "entropy_score": N}`, stable ordering (determinism requirement)
- Errors: `db_unreadable` (2)
- Consumer: `check` (wired into gate flow)

---

## Workflow-Step → CLI-Action Mapping

| Skill step | CLI action |
|---|---|
| brainstorm: SPEC lock | `intake --type … --summary … --lane … --json` |
| to-plan: roadmap init | `init` (if no db) |
| to-plan: per-phase | `story --slug {phase} --goal … --json` |
| to-plan: status render | `query state --json`, `query phases --json` |
| work: run registration (full mode) | `run create --slug {phase} --artifact-path … [--plan-id …] --json` |
| work: wave completion | `trace add --wave N --summary … --json` |
| check: gate evaluation | `audit --json` + `score-trace <id> --json` → matrix → `check record --verdict … --json` |
| check: human override | `intervention --verdict-id … --reason …` (manual, documented escalation only) |
| watzup: recap render | `resume --json` (1:1, no fallback) |
| handoff: close-out | `handoff record --run-id … --check-id … --open-items … --json` |
| git: pre-commit/PR | `query check --latest --json` (warn-not-block on FAIL/missing) |
| continuity: fresh-machine rebuild | `db changeset apply <path>` (per committed changeset, ULID order) |
