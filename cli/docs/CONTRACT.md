# CLI Contract

`zharness` command surface. Every command supports `--json`; human-readable output is the default without the flag. Exit code `0` = success; non-zero = failure, and every JSON error response carries a stable machine-readable `code` string (`snake_case`) alongside the numeric exit code, so scripts can branch on `code` without parsing text.

> **Escalation note**: SPEC R6 names exactly 19 commands, and this contract documents exactly those 19 — no `handoff` command is included here, even though R18 ("handoff records entity") and continuity's Wave 1 T1 depend on one existing. `continuity-CONTEXT.md`'s own Assumption pre-answers this: "if the CLI lacks it, that is a Phase 4 gap — escalate, don't shim." This is a real SPEC-internal inconsistency (R6's enumerated list vs. R18's requirement) — see `.kit/implementation-notes.md` for the full reasoning — but the fix belongs to whichever phase actually hits the gap (cli-domain or continuity), via `brainstorm refine`, not to a preemptive addition here that would violate this phase's own locked 19-command scope.

## Error Codes

| Exit | Meaning |
|---|---|
| 0 | success |
| 1 | validation/user error (bad args, bad fixture, gate FAIL) |
| 2 | system error (db unreadable, fs permission, network for `gh` calls) |

Every non-zero JSON response: `{"error": {"code": "snake_case_string", "message": "human text"}}`.

---

## Core (Phase 3 — cli-core)

### `init`
- Args: none; `--force` reinitializes an empty db if one already exists (refuses on a non-empty db without `--force`)
- `--json`: `{"status": "created"|"exists", "db_path": "...", "schema_version": N}`
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

## Domain — 7 ported (Phase 4 — cli-domain)

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

### `decision`
- Args: `--summary "..." --rationale "..." [--rejected "..."]`
- `--json`: `{"id": "ulid"}`
- Errors: `missing_required_field` (1)
- Consumer: none mandated inline — general-purpose, available for ad hoc recording outside the fixed skill flows (ported behavior, not adopted into a specific skill step)

### `backlog`
- Args: `--summary "..." [--priority tiny|normal|high-risk]`
- `--json`: `{"id": "ulid"}`
- Errors: `missing_required_field` (1)
- Consumer: none mandated inline — general-purpose (same as `decision`)

### `tool`
- Args: `--name "..." --purpose "..."`
- `--json`: `{"id": "ulid"}`
- Errors: `missing_required_field` (1)
- Consumer: none mandated inline — general-purpose (records tool/capability usage, ported behavior)

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

### `resume`
- Args: none
- `--json`: `{"position": {"current_phase": "...", "status": "..."}, "latest_run_id": "ulid"|null, "latest_check_id": "ulid"|null, "latest_handoff_id": "ulid"|null, "drift": [{"type": "missing_file"|"unknown_phase"|"out_of_order", "detail": "...", "recovery": "..."}], "readiness": "clean"|"in-progress"|"drifted"|"no-harness"}`
- Errors: `db_unreadable` (2, distinct from `readiness: no-harness` — that's a valid successful response, not an error)
- Consumer: `watzup` (renders 1:1 from this snapshot, no independent prose re-derivation)

### `check record`
- Args: `--verdict {APPROVED|APPROVE_WITH_REQUESTS|REQUEST_CHANGES} --run-id {ulid} --proof-links '[{"command":"...","output_ref":"...","artifact_path":"..."}]'`
- `--json`: `{"id": "ulid", "verdict": "..."}`
- Errors: `unknown_run_id` (1), `invalid_verdict` (1), `empty_proof_links` (1, verdict other than REQUEST_CHANGES with zero proof links)
- Consumer: `check` (deterministic — no free-text-only verdicts, R19)

### `validate`
- Args: none — walks SPEC→PLAN→RUN→CHECK→HANDOFF by frontmatter ULID cross-links
- `--json`: `{"valid": true|false, "findings": [{"link": "SPEC->PLAN", "issue": "missing_key"|"broken_link"|"stale_pointer", "detail": "..."}]}`
- Errors: exits 1 with non-empty `findings` on any violation
- Consumer: `skill-adapters` T4 sample-chain proof, `pilot-migration` evidence bundle
- **Known gap**: the `SPEC->PLAN` link cannot currently be validated — `phase-plan-template.md` was not in harness-contracts' Allowed Surfaces, so PLAN artifacts don't yet carry a `spec_id` field. `validate` should skip that one link with a `not_yet_implemented` finding rather than a hard failure, until a later phase adds it (see `.kit/implementation-notes.md`).

---

## Research (Phase 6 — validation-gate)

### `score-trace <trace-id>`
- Args: trace ULID
- `--json`: `{"tier": "...", "reasons": ["..."]}`
- Errors: `unknown_trace_id` (1)
- Consumer: `check` (wired into gate flow: `audit` + `score-trace` → matrix evaluation → `check record`)

### `score-context <trace-id>`
- Args: trace ULID
- `--json`: `{"score": N, "reasons": ["..."]}`
- Errors: `unknown_trace_id` (1)
- Consumer: **reserved — documented only**, not adopted into any skill this initiative (validation-gate Forbidden Surfaces)

### `audit`
- Args: none
- `--json`: `{"pointer_drift": [...], "contract_violations": [...], "unlinked_proofs": [...], "entropy_score": N}`, stable ordering (determinism requirement)
- Errors: `db_unreadable` (2)
- Consumer: `check` (wired into gate flow)

### `propose`
- Args: none
- `--json`: `{"proposals": [{"pattern": "...", "suggestion": "..."}]}`
- Errors: `db_unreadable` (2)
- Consumer: **reserved — documented only**, not adopted into any skill this initiative (validation-gate Forbidden Surfaces)

---

## Workflow-Step → CLI-Action Mapping

| Skill step | CLI action |
|---|---|
| brainstorm: SPEC lock | `intake --type … --summary … --lane … --json` |
| to-plan: roadmap init | `init` (if no db) |
| to-plan: per-phase | `story --slug {phase} --goal … --json` |
| to-plan: status render | `query state --json`, `query phases --json` |
| work: wave completion | `trace add --wave N --summary … --json` |
| check: gate evaluation | `audit --json` + `score-trace <id> --json` → matrix → `check record --verdict … --json` |
| check: human override | `intervention --verdict-id … --reason …` (manual, documented escalation only) |
| watzup: recap render | `resume --json` (1:1, no fallback) |
| handoff: close-out | **not yet in the 19-command surface** — R18 requires this, no verb exists among R6's 19; see escalation note at top of this doc |
| git: pre-commit/PR | `query check --latest --json` (warn-not-block on FAIL/missing) |
| continuity: fresh-machine rebuild | `db changeset apply <path>` (per committed changeset, ULID order) |
