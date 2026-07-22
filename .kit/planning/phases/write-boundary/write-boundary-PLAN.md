# Plan: write-boundary

Phase: write-boundary
Status: done — implemented + gated APPROVED 2026-07-21 (commit `32cb60c`, check report `.kit/reports/check/20260721-1044-write-boundary.md`, run artifact `.kit/runs/work/20260721-1027-write-boundary.md`)
Wave Count: 3
Execution Owner: work
Updated At: 2026-07-22

## Goal
Add `zharness run create` and make `check record` set `latest_check_id`; rewire `work.md`/`check.md` to call them; delete hand-authored-JSONL steps.

## Inputs
- `cli/internal/infrastructure/changeset.go` (WriteChangeset/ApplyChangeset, entityColumns for `runs`)
- existing `zharness story` / `trace add` command as pattern
- `.kit/docs/playbooks/work.md` step 2, `check.md` step 4

## Wave 1
### T1 — Implement `zharness run create`
- type: implementation
- inputs:
  - existing mutating-command pattern (`story`, `trace add`)
  - `runs` table columns: story_slug, plan_id, trace_ids, artifact_path, created_at
- touches:
  - `cli/internal/interfaces/` (new command), `cli/internal/application/` (logic)
- avoid:
  - `score.go`, `audit.go`, `migrations.go`, changeset format
- steps:
  1. Add `run create` cobra command: flags `--slug` (story_slug, required), `--artifact-path` (required), `--plan-id` (optional), `--json`.
  2. In application layer, build a two-line changeset: line 1 `create` run (mint run ULID via existing id logic), line 2 `update meta.latest_run_id`. Write via `WriteChangeset`, apply via `ApplyChangeset` in one tx.
  3. Return the run `id` as JSON.
  4. Guard: require an existing story matching `--slug` (fail `unknown_story` if absent) — mirrors the FK the hand-authored path hit.
- expected outputs:
  - `zharness run create --slug write-boundary --artifact-path .kit/runs/work/x.md --json` prints `{"id":"<ULID>"}` and sets `latest_run_id`
- verification:
  - `go test ./cli/internal/application/ -run RunCreate` (new unit test) → pass
  - manual: `zharness run create ... --json && zharness query state --json | grep latest_run_id`
- stop if:
  - reproducing two-line semantics needs a changeset-format change
- escalate to:
  - to-plan phase write-boundary

### T2 — `check record` sets `latest_check_id`
- type: implementation
- inputs:
  - existing `check record` command (`cli/internal/interfaces/check.go`, `application/check_record.go`)
- touches:
  - `cli/internal/interfaces/check.go`, `cli/internal/application/check_record.go`
- avoid:
  - scoring logic, audit shape
- steps:
  1. Extend `check record` so its changeset also updates `meta.latest_check_id` to the new check id (same tx).
  2. Default on; optionally add `--no-set-latest` (only if trivial).
- expected outputs:
  - after `check record`, `query state --json` shows `latest_check_id` = new check id
- verification:
  - `go test ./cli/internal/application/ -run CheckRecord` → pass
- stop if:
  - `check record` is used by legacy import in a way that conflicts → note and escalate
- escalate to:
  - to-plan phase write-boundary

## Wave 2
### T3 — Rewrite `work.md` embed step 2
- type: docs
- inputs:
  - `cli/docs/embedded/playbooks/work.md`
- touches:
  - `cli/docs/embedded/playbooks/work.md`
- avoid:
  - `.kit/docs/playbooks/work.md` (generated — re-scaffold instead)
- steps:
  1. Replace the full-mode hand-authored two-line changeset block with a single `zharness run create --slug ... --artifact-path ... --json` call.
  2. Keep the simple-mode carve-out (no DB registration) intact.
  3. Run `zharness init --refresh-docs` (or equivalent) to re-scaffold `.kit/docs/`.
- expected outputs:
  - `work.md` step 2 references `run create`, no JSONL literal to author
- verification:
  - `grep -c '"op":"create","entity":"run"' cli/docs/embedded/playbooks/work.md` → 0 hand-author instructions (only command usage remains)
- stop if:
  - unclear how simple-mode text should change → ask
- escalate to:
  - user clarification

### T4 — Rewrite `check.md` embed step 4
- type: docs
- inputs:
  - `cli/docs/embedded/playbooks/check.md`
- touches:
  - `cli/docs/embedded/playbooks/check.md`
- avoid:
  - scoring content (Phase 3 owns that)
- steps:
  1. Replace the hand-authored `latest_check_id` meta changeset with the note that `check record` sets it.
  2. Re-scaffold `.kit/docs/`.
- expected outputs:
  - `check.md` step 4 has no hand-author meta-changeset instruction
- verification:
  - `grep -c 'latest_check_id.*changeset apply' cli/docs/embedded/playbooks/check.md` → no hand-author block
- stop if:
  - step 4 also entangles score-trace (Phase 3) → touch only the meta-pointer lines, leave scoring for Phase 3
- escalate to:
  - to-plan phase scoring-removal

## Wave 3
### T5 — Integration + replay-safety proof
- type: test
- inputs:
  - new commands from Wave 1
- touches:
  - `cli/internal/application/` integration test
- avoid:
  - unrelated fixtures
- steps:
  1. Add an integration test: `init` → create story → `run create` → rebuild DB from changesets → assert `resume --json` identical.
  2. Run full `go test ./...` and `go build ./...`.
- expected outputs:
  - green build + tests; replay reproduces state
- verification:
  - `cd cli && go test ./... && go build ./...` → pass (capture output)
- stop if:
  - replay diverges after `run create`
- escalate to:
  - to-plan phase write-boundary

## Risks / Watch-fors
- Keep `run create` full-mode only; simple mode must not gain a run row.
- work.md/check.md edits go to the **embed**, then re-scaffold — never edit `.kit/docs/` directly.
