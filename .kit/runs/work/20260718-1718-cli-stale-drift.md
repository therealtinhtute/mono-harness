---
id: 01KXTBQXH36QHNEBC2CXRDKQ4G
type: run
phase: cli-stale-drift
lane: high-risk
plan_id: null
trace_ids: [01KXTC2A16YCCJSE0MKAJX7ZJ3, 01KXTNV9NECFMK4BTCKFKNZEMH]
created: 2026-07-18
updated: 2026-07-18
---

# COOK RUN

Run ID: work-20260718-1718-cli-stale-drift
Mode: full
Status: passed
Spec: .kit/planning/SPEC.md
Roadmap: .kit/planning/ROADMAP.md
Phase: cli-stale-drift
Plan: .kit/planning/phases/cli-stale-drift/cli-stale-drift-PLAN.md
Started At: 2026-07-18 17:18

## Preflight
- scope drift: no — working tree carries only phase 1 + phase 2's own (already-gated) uncommitted output; no files outside cli-stale-drift's Allowed Surfaces changed yet
- working tree note: phase 1 + phase 2 remain uncommitted per `work`'s own "suggest, never auto-commit" rule; both were suggested for commit at phase 2's gate, not yet actioned
- required artifacts present: yes — SPEC.md locked, ROADMAP.md present, cli-stale-drift-CONTEXT.md + cli-stale-drift-PLAN.md both present and internally consistent (dates, no placeholders)
- selected phase: cli-stale-drift (3rd of 6), depends_on cli-embed-scaffold (gated APPROVE_WITH_REQUESTS — dependency satisfied)
- note: `plan_id: null` above is the same pre-existing, out-of-scope gap already documented in cli-embed-scaffold's gate (PLAN.md artifacts carry no frontmatter `id` at all yet — CONTRACT.md's known `SPEC->PLAN not_yet_implemented` gap, one link further down the chain). Not this phase's fix.

## Wave / Task Log
### Wave 1
#### T1 — stale_docs drift in resume
- status: DONE
- changed files:
  - cli/internal/application/resume.go (new `StaleDocsRecovery` constant + `docsVersionMinSchema` guard constant; `Resume` now takes `cliVersion string`; new stale_docs check appended into the existing drift-detection flow, guarded on `state.SchemaVersion >= docsVersionMinSchema` before ever reading `meta.docs_version` — avoids a raw "no such column" SQL error against a db migrated by an older binary that predates migration 0002)
  - cli/internal/application/audit.go (`Audit`/`Propose` now take `cliVersion string`, threaded into their own `Resume(db, cliVersion)` call — `audit --json`'s `pointer_drift` is `Resume`'s `Drift` verbatim, so stale_docs must flow through here too or `audit` would never surface it)
  - cli/internal/interfaces/resume.go, audit.go, root.go (thread the CLI's own `version` string, already available at root, into `newResumeCmd`/`newAuditCmd`/`newProposeCmd` — same pattern `newInitCmd(version)` already established in cli-embed-scaffold)
  - cli/internal/application/resume_stale_docs_test.go (new — 6 cases: match, differ, dev-on-written-side, dev-on-cli-side, missing/unversioned, pre-migration schema_version)
  - cli/internal/application/resume_test.go, audit_test.go, scaffold_integration_test.go (updated existing `Resume(db)`/`Audit(db, root)`/`Propose(db, root)` call sites to the new signatures — `"dev"` for the pre-existing fixtures, `"test-integration"` for scaffold_integration_test.go to match what it actually stamps)
  - cli/docs/CONTRACT.md (added `stale_docs` to `resume --json`'s drift `type` enum — not in this phase's Allowed Surfaces, same documented-exception pattern as cli-embed-scaffold's CONTRACT.md sync; see notes)
- decision: "the CLI's embedded docs version" (CONTEXT.md's Locked Decision wording) is the same `version` string already threaded everywhere else (`main.go`'s ldflags-injected release version, `"dev"` for local builds) — not a separate value read from the embedded FS. `docsVersion` stamped by `init`/`--refresh-docs` IS this same string, so comparing `meta.docs_version` against the running binary's own `version` is exactly the firing rule CONTEXT.md specifies. No new storage needed.
- decision: "Missing meta.docs_version ... resume may note docs as `unversioned` informationally (or omit)" — chose omit. Adding an informational field would mean extending `resume --json`'s locked top-level shape (a bigger CONTRACT.md change than the drift-type enum addition, and CONTEXT.md's own wording explicitly offers omission as a valid choice), so the simpler option was taken per Hard Rule 5.
- verification:
  - `cd cli && go test ./internal/application/... -run TestResumeStaleDocs -v` → pass (6/6)
  - `cd cli && go build ./... && go vet ./... && gofmt -l .` → clean
  - `cd cli && go test ./...` → pass, all packages, no regressions in existing Resume/Audit/Propose tests
- notes:
  - CONTRACT.md exception: same judgment call already resolved via `AskUserQuestion` in cli-embed-scaffold ("update CONTRACT.md now" was the user's explicit, recorded preference for this exact category of decision within this initiative) — applied without re-asking since the precedent is direct and low-stakes (documentation accuracy, not behavior).

#### T2 — STATE.md drift-table row
- status: DONE
- changed files:
  - cli/docs/STATE.md (added the `stale_docs` row to the Stale-Pointer Rules table, quoting `application.StaleDocsRecovery`'s literal value)
- verification:
  - `grep -o "zharness init --refresh-docs" cli/docs/STATE.md cli/internal/application/resume.go` → both hits, byte-identical string confirmed
- notes:
  - none

### Wave 2
#### T3 — Clearing-semantics test
- status: DONE
- changed files:
  - cli/internal/application/clearing_semantics_test.go (new — `TestClearingSemantics_RefreshDocsResolvesStaleDocsDrift`: seeds a story+run+meta pointers with a real on-disk artifact_path, scaffolds docs at "0.2.0", proves `Resume(db, "0.3.0")` fires stale_docs with the exact `StaleDocsRecovery` string, runs `ScaffoldDocs(..., "0.3.0", refresh=true)`, proves `Resume(db, "0.3.0")` afterward is clean with zero drift, and asserts the story status/run artifact_path/meta current_phase+entry_phase+latest_run_id are all byte-stable across the refresh — only `.kit/docs/**` + the `docs_version` stamp changed)
- verification:
  - `cd cli && go test ./internal/application/... -run TestClearingSemantics -v` → pass
  - `cd cli && go build ./... && go vet ./... && gofmt -l .` → clean
  - `cd cli && go test ./...` → pass, all packages
- notes:
  - first attempt failed: used the shared `seedRun` test helper, whose fixed fake `artifact_path` (`.kit/runs/work/x.md`) doesn't exist on disk and trips an unrelated `missing_file` drift, muddying the stale_docs-specific assertion. Fixed by seeding the story/run manually with a real on-disk artifact_path instead of reusing `seedRun`.
  - second attempt failed: seeded story status `"in-progress"`, which drives `Resume`'s readiness to `"in-progress"` regardless of drift — the test's "clean after refresh" assertion needs a status that isn't `planned`/`in-progress` per the existing readiness switch. Fixed by seeding status `"done"`.

#### T4 — Scratch-dir lifecycle integration suite
- status: DONE_WITH_CONCERNS
- changed files:
  - cli/cmd/zharness/lifecycle_test.go (new — `TestLifecycle_ScratchDirFullChain`: builds the real `zharness` binary once via `go build`, then drives it through `os/exec` on a scratch dir through the full chain `init → intake → story → run registration (changeset) → trace add → check record → handoff record → resume/validate/audit`, exactly as PLAN.md T4 step 1 lists — real CLI flag parsing, real JSON output, real exit codes, not internal Go calls. Reusable as-is for cli-release's own smoke test per the PLAN's risk note.)
- decision (escalated via `AskUserQuestion`, user confirmed "Honest lifecycle test, mark DONE_WITH_CONCERNS"): PLAN.md T4 step 2 literally asks to assert "readiness transitions (clean → in-progress → checked)". Investigation found this is not achievable through the CLI as it exists today, and the gap is *larger* than initially scoped in the question: `resume`'s readiness only reads `Position.Status`, which only populates when `meta.current_phase` is set — and **nothing** in the current CLI surface ever writes `current_phase`, not even `work`'s own documented run-registration changeset (confirmed empirically: the test's own run-registration step, copied verbatim from `work`'s execution-loop step 2, left readiness at `"clean"`, not `"in-progress"` as first expected — the test's assertions were corrected to match this real, observed behavior rather than the assumption in the original question). Separately, no command ever transitions `story.status` past `"planned"` either (STATE.md's Writer/Reader Ownership table documents `check record`/`work` writing phase-status transitions; no production code does). Both are the same class of gap already recorded as backlog item `01KXTBG4JZYTW528Y5XZQK8FEH` during cli-embed-scaffold's gate — not widened or fixed here, just fully confirmed empirically and documented precisely.
- verification:
  - `cd cli && go test ./cmd/zharness/... -run TestLifecycle -v -timeout 120s` → pass
  - `cd cli && go build ./... && go vet ./... && gofmt -l .` → clean
  - `cd cli && time go test ./...` → pass, all packages, ~0.9s total wall time (no build tag needed — not slow enough to warrant excluding from the default suite, per PLAN.md's own "verification: go test ./..." line)
  - Confirmed via `zharness query state --json` (installed v0.1.0 binary) that the real repo's own `.kit/harness.db` stayed at `schema_version=1` throughout — all of T4's binary builds and CLI invocations ran inside Go's own `t.TempDir()` scratch dirs via `os/exec`, never touching repo root
- notes:
  - `validate --json`/`audit --json` assertions ARE fully clean per PLAN.md's literal ask (both report exactly the one already-known, already-accepted `not_yet_implemented` SPEC->PLAN finding and nothing else) — only the readiness-transition part of T4's acceptance criteria is scoped down, not the whole task.

## Summary
- passed tasks: T1, T2, T3
- concern tasks: T4 (DONE_WITH_CONCERNS — see decision note above; user-confirmed scope adaptation via AskUserQuestion, not a silent gap)
- blocked tasks: none
- unresolved concerns:
  - PLAN.md T4's literal "readiness transitions clean → in-progress → checked" is unimplementable through the current CLI surface — confirmed empirically, not just plausible. Root cause: `meta.current_phase` is never written by any documented command (including `work`'s own run-registration changeset), and `story.status` never transitions past `planned` either. Both tracked as backlog item `01KXTBG4JZYTW528Y5XZQK8FEH` (recorded during cli-embed-scaffold's gate, confirmed and detailed further here). Not this phase's fix — CONTEXT.md's Goal is drift detection + lifecycle *testing*, not building missing lifecycle-management commands; that's a separate, unplanned initiative-sized change.
  - CONTRACT.md was touched again (`resume --json`'s drift `type` enum, adding `stale_docs`) despite not being in this phase's Allowed Surfaces — same documented-exception precedent as cli-embed-scaffold, applied without re-asking since the category of decision was already resolved once this initiative.

## Next Recommended Action
- Gate this phase via `check full` (dogfooding `cli/docs/embedded/playbooks/check.md`) before advancing to `cli-release`. T4's DONE_WITH_CONCERNS status and the backlog item should both be surfaced explicitly at the gate, not silently passed.
