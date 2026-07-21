---
id: 01KY1BD7D5P2PQTCRYDP5KJKPC
type: run
phase: write-boundary
lane: high-risk
mode: full
plan_id: none
trace_ids: [01KY1BKJN50ZVG2C1X03SFH05N, 01KY1BQDX0ZNY02VFAW9TJPW8H, 01KY1BTHY7WSP68D4JXQ5EN8PR]
created: 2026-07-21
updated: 2026-07-21
---

# COOK RUN

Run ID: work-20260721-1027-write-boundary
Mode: full
Status: passed
Spec: .kit/planning/SPEC.md
Roadmap: .kit/planning/ROADMAP.md
Phase: write-boundary
Plan: .kit/planning/phases/write-boundary/write-boundary-PLAN.md
Started At: 2026-07-21 10:27

## Preflight
- scope drift: no
- working tree note: 3 pre-existing modified files unrelated to this initiative (README.md, assets/spec-plan-workflow.svg, docs/workflow-harness/migration.md — prior readme-workflow-refresh work); untracked planning/report artifacts from this initiative's brainstorm/to-plan steps
- required artifacts present: yes
- selected phase / source prompt: write-boundary (entry phase, first incomplete per roadmap)

## Wave / Task Log
### Wave 1
#### T1 — Implement `zharness run create`
- status: DONE
- changed files:
  - cli/internal/application/run_create.go (new)
  - cli/internal/interfaces/run.go (new)
  - cli/internal/interfaces/root.go (register `run` command)
  - cli/internal/application/run_create_test.go (new)
- verification:
  - `go test ./internal/application/ -run RunCreate -v` → pass (TestRunCreate, TestRunCreateUnknownStory, TestRunCreateMissingArtifactPath)
  - manual: `zharness run create --slug demo-phase --artifact-path .kit/runs/work/x.md --json` → `{"id":"01KY1BHWSVNGHTZD1M8EJED23Q"}`; `zharness query state --json` showed `latest_run_id` = that id (isolated tmp dir, cleaned up after)
- notes:
  - none

#### T2 — `check record` sets `latest_check_id`
- status: DONE
- changed files:
  - cli/internal/application/check_record.go
  - cli/internal/application/check_record_test.go
- verification:
  - `go test ./internal/application/ -run CheckRecord -v` → pass (5/5)
  - `go test ./...` (full suite) → all packages ok
- notes:
  - default-on, no `--no-set-latest` flag added (locked decision leaned default; flag was optional-not-required, skipped per simplicity)

### Wave 2
#### T3 — Rewrite `work.md` embed step 2
- status: DONE
- changed files:
  - cli/docs/embedded/playbooks/work.md
- verification:
  - `grep -c '"op":"create","entity":"run"' cli/docs/embedded/playbooks/work.md` → 0
- notes:
  - reordered id-minting: full mode now gets the RUN id from `run create`'s JSON response (the command mints it internally, no longer accepts a pre-minted id); simple mode keeps `zharness id --json` unchanged. This is a small deviation from the plan's literal wording ("replace step-2 hand-authored block with a single `run create` call") but was necessary — `run create` mints its own ULID rather than accepting an externally-minted one, so the "mint id first, write artifact, then register" sequencing had to become "register (mints id), then write artifact using the returned id" for full mode.

#### T4 — Rewrite `check.md` embed step 4
- status: DONE
- changed files:
  - cli/docs/embedded/playbooks/check.md
- verification:
  - `grep -c 'latest_check_id.*changeset apply' cli/docs/embedded/playbooks/check.md` → 0
- notes:
  - touched only the meta-pointer lines (step 4 prose + Command Reference); scoring/score-trace content left untouched for Phase 3

#### Re-scaffold `.kit/docs/`
- status: DONE
- changed files:
  - none tracked (generated: `.kit/docs/playbooks/work.md`, `.kit/docs/playbooks/check.md`)
- verification:
  - built dev binary (`go build -o /tmp/zharness-test ./cmd/zharness`), ran `zharness init --refresh-docs --json` from repo root → `.kit/docs/playbooks/{work,check}.md` now byte-identical to the edited embeds (verified via `diff`); real DB state (`query state --json`) unchanged (`current_phase`/`latest_run_id`/`latest_check_id` all intact)
- notes:
  - first invocation was accidentally run from `cli/` and created a stray `cli/.kit/`; removed before re-running from repo root

### Wave 3
#### T5 — Integration + replay-safety proof
- status: DONE
- changed files:
  - cli/internal/application/run_create_replay_test.go (new)
  - cli/internal/embedded/embedded_test.go
- verification:
  - `go test ./internal/application/ -run TestRunCreateReplaySafety -v` → pass
  - `cd cli && go build ./... && go test ./...` → all packages ok
  - `go vet ./...` → clean
- notes:
  - Two pre-existing embedded-doc tests (`TestWorkPlaybook_UsesIDCommand`, `TestPlaybooks_ManualIDsUseIDCommand`) locked in the OLD hand-authored-changeset phrasing this phase deliberately removes; they failed after the T3/T4 embed edits. Updated their expected phrases to match the new `run create`/`check record` contract and dropped the `check.md` case from the manual-ID table (check.md no longer mints IDs manually — `check record` owns the pointer). This is outside the plan's literal `touches` list for T5 but is a direct, necessary consequence of T3/T4's scope, not new scope.

## Summary
- passed tasks: T1, T2, T3, T4, T5
- blocked tasks: none
- unresolved concerns:
  - the embedded_test.go update above (surfaced for awareness, not blocking)

## Next Recommended Action
- `check full` — high-risk proof matrix gate for write-boundary
