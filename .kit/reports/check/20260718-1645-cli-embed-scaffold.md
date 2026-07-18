---
id: 01KXTBJ0QAYZHKYMA32FNWCW8Q
type: check
phase: cli-embed-scaffold
lane: high-risk
run_id: 01KXT8WSQ5KJR03JGKJZHCGSN8
proof_links: [{"command":"cd cli && go build ./... && go vet ./... && gofmt -l .","output_ref":"all clean, zero output/errors","artifact_path":"cli/"},{"command":"cd cli && go test ./... -v","output_ref":"pass, all packages (domain, embedded, infrastructure, application) - includes 6 TestScaffoldDocs_* cases, TestBuildManifest_MatchesDiskTree, TestInit_FreshScratchDir_FullIntegration, TestMetaDocsVersionRoundTrip, TestMigrate","artifact_path":"cli/internal/application/init_test.go"},{"command":"zharness audit --json (installed v0.1.0, read-only, against real repo db)","output_ref":"1 pointer_drift out_of_order (self-resolved after this check record landed and updated latest_check_id — confirmed clean on re-run); 2 pre-existing contract_violations unrelated to this phase (SPEC->PLAN not_yet_implemented documented gap; PLAN->RUN missing_key from PLAN.md never carrying an id, outside this phase's Allowed Surfaces); 2 pre-existing CHECK->HANDOFF missing_key from stale phase-1 HANDOFF.md","artifact_path":".kit/harness.db"},{"command":"manual dev-binary scratch-dir testing: init idempotency matrix (no-.kit / db-only / docs-present cells) + init --refresh-docs canonical overwrite","output_ref":"all 3 matrix cells plus refresh path verified against real binary in isolated scratch dirs; repo's own .kit/harness.db confirmed untouched via git status --short after each round","artifact_path":".kit/runs/work/20260718-1628-cli-embed-scaffold.md"},{"command":"zharness trace add (backfill, wave 1 and wave 2) + zharness score-trace","output_ref":"both traces recorded and scored tier=detailed","artifact_path":".kit/runs/work/20260718-1628-cli-embed-scaffold.md"}]
created: 2026-07-18
updated: 2026-07-18
---

# CHECK REPORT

Run ID: check-20260718-1645-cli-embed-scaffold
Scope: full
Artifact Alignment: aligned (one explicit, user-approved boundary exception — see notes)
Review Verdict: APPROVE with requests
Phase: cli-embed-scaffold
Spec: .kit/planning/SPEC.md
Plan: .kit/planning/phases/cli-embed-scaffold/cli-embed-scaffold-PLAN.md
Cook Run: .kit/runs/work/20260718-1628-cli-embed-scaffold.md
Created At: 2026-07-18 16:45

## Gate Evidence
- tests: `cd cli && go test ./... -v` → pass (all 4 packages; interfaces has no test files, consistent with repo convention)
- types/vet: `cd cli && go vet ./...` → pass
- lint: `cd cli && gofmt -l .` → pass (clean)
- build: `cd cli && go build ./...` → pass

## Artifact Alignment
- status: aligned
- notes:
  - spec coverage: R1 (embed mechanism), R2 (init scaffolding + .gitignore), R3 (docs_version stamp + --refresh-docs) all implemented and tested; matches CONTEXT.md's Spec Hooks exactly
  - all 5 PLAN.md tasks (T1-T5) DONE, verified individually with real command output
  - Locked Decisions cross-checked against code: write-out target `.kit/docs/` ✅, AGENTS.md never-overwrite ✅, docs_version storage + `dev` semantics ✅, idempotency matrix (3 cells, all tested) ✅, `--refresh-docs` canonical-overwrite-only semantics ✅ (proven not to touch other meta pointers), `.gitignore` append-only ✅
  - boundary compliance: every changed file falls inside Allowed Surfaces (`cli/internal/**`, `cli/docs/embedded/**` consumed, `cli/docs/SCHEMA.md`/`STATE.md`, Go test files) with **one exception**: `cli/docs/CONTRACT.md` is not in this phase's Allowed Surfaces but was updated twice (T3, T4) to keep the `init` contract in sync with real behavior — this was explicitly surfaced via `AskUserQuestion` mid-phase and the user chose "update CONTRACT.md now." Documented exception, not silent drift.
  - proof trail: intact — run artifact documents verification for every task; `trace_ids` was found empty at gate time (wave-completion `trace add` was never invoked during execution) and was backfilled during this gate (see Findings)

## Findings
### Critical
- none

### Major
- **[found + remediated during this gate]** Wave-completion `zharness trace add` (work's own execution-loop step 9) was never invoked for either wave during execution — run artifact `trace_ids` was `[]`. Backfilled both waves (`01KXTBDEE399NBVS0M28T9P4MW` wave 1, `01KXTBDEEBTQJGJR90JG9HGNGB` wave 2 — both scored tier=detailed), appended to the run artifact's frontmatter. Did not block the harness gate's proof matrix (unit/integration/manual-check/command-output were independently satisfied by real test evidence), but is a real process gap worth naming given this initiative's own goal is harness self-tracking.

### Minor / Suggestions
- PLAN.md declares Wave Count: 3 (Wave 1 = T1+T2, Wave 2 = T3+T4, Wave 3 = T5); execution logged T3+T4+T5 together under one "Wave 2" heading. No task ran out of order and all were correctly dependency-sequenced, so this is a labeling granularity mismatch, not a correctness issue. Documented via a process note in the run artifact.
- (advisory, deferred to `cli-stale-drift` per CONTEXT.md's own scoping) `copyFS`/`ScaffoldDocs` overwrites per-file on `--refresh-docs` but never prunes a doc present on disk but no longer in the embedded manifest. This repo's own `cli/docs/embedded/` is guarded by `TestBuildManifest_MatchesDiskTree`; a *consumer* project's `.kit/docs/` has no equivalent orphan guard yet. Known boundary, not a silent gap — noted in the run artifact.
- (out of scope, recorded not fixed) `meta.current_phase`/`story.status` never advance past their initial `to-plan` values anywhere in the currently documented CLI/skill flow — confirmed all 6 roadmap stories (including the already-gated `playbook-authoring`) still read `status: planned`, and `meta.current_phase` still reads `playbook-authoring` after this phase's own gate. `audit`'s `unknown_phase` check doesn't catch this since the slug is still a valid story. Not this phase's Allowed Surface (story/resume domain logic) — recorded as backlog item `01KXTBG4JZYTW528Y5XZQK8FEH` for future scoping rather than silently dropped.

## Harness Gate Flow (Validation Matrix, lane=high-risk)
- `unit`: required → satisfied — `TestScaffoldDocs_*` (6 cases covering the full idempotency matrix + refresh semantics), `TestBuildManifest_MatchesDiskTree`, `TestMetaDocsVersionRoundTrip`, `TestMigrate`
- `integration`: required → satisfied — `TestInit_FreshScratchDir_FullIntegration` (real embedded FS, real migrated sqlite db, real filesystem, asserts through to `Resume`'s `readiness: clean`)
- `e2e`: optional → gathered anyway — manual dev-binary scratch-dir runs of the actual `init`/`init --refresh-docs` commands
- `manual-check`: required → satisfied — this Phase 2 review pass itself; zero unresolved 🔴 findings, one 🟠 found-and-fixed inline, remainder 🟡/💡 advisory
- `command-output`: required → satisfied — build/vet/gofmt/test all cited above with real output

All required cells satisfied with real evidence — no matrix-forced verdict override needed (unlike phase 1's `REQUEST_CHANGES`-by-matrix, this phase supplied its own unit+integration proof as expected).

## Next Action
- Gate clean. Ready to advance to `cli-stale-drift` per the roadmap's dependency order.
- Suggest (not auto-run): commit this phase's diff before starting `cli-stale-drift` — phase 1 + phase 2 are both still uncommitted, and piling a third phase's diff on top will make the next gate harder to scope.

---

scope:              on target
depth:              deep
artifact_alignment: ✅ aligned (1 documented, user-approved exception — CONTRACT.md)
gate:               ✅ pass: build, vet, gofmt, test (all packages)
review:             APPROVE with requests
blockers:           0 critical, 0 major (1 major found + fixed inline during this gate)
autofix:            0 safe_auto proposed, 0 gated_auto awaiting confirmation
verification:       go build/vet/gofmt/test ./... → pass; zharness audit --json → pointer_drift clean after check record; manual scratch-dir CLI testing → pass
harness_verdict:    zharness check record id 01KXTBJ0QAYZHKYMA32FNWCW8Q (APPROVE_WITH_REQUESTS); meta.latest_check_id updated via changeset; backlog item 01KXTBG4JZYTW528Y5XZQK8FEH recorded for out-of-scope current_phase/story-status gap
