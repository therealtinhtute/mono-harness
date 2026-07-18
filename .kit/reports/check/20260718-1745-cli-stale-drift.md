---
id: 01KXTNYEY80SPB3QZ6NX4D173J
type: check
phase: cli-stale-drift
lane: high-risk
run_id: 01KXTBQXH36QHNEBC2CXRDKQ4G
proof_links: [{"command":"cd cli && go build ./... && go vet ./... && gofmt -l .","output_ref":"all clean, zero output/errors","artifact_path":"cli/"},{"command":"cd cli && go test ./... -v","output_ref":"pass, all packages - includes 6 TestResumeStaleDocs* cases, TestClearingSemantics_RefreshDocsResolvesStaleDocsDrift, TestLifecycle_ScratchDirFullChain (real built binary via os/exec), plus all pre-existing tests with zero regressions","artifact_path":"cli/internal/application/resume_stale_docs_test.go"},{"command":"cd cli && go test ./cmd/zharness/... -run TestLifecycle -v -timeout 120s","output_ref":"pass - real zharness binary driven through init/intake/story/run-registration/trace add/check record/handoff record/resume/validate/audit on a scratch dir; validate and audit both fully clean modulo the one known not_yet_implemented gap","artifact_path":"cli/cmd/zharness/lifecycle_test.go"},{"command":"zharness audit --json (installed v0.1.0, read-only, against real repo db)","output_ref":"1 pointer_drift out_of_order (self-resolved after this check record landed); 5 pre-existing/expected contract_violations, none introduced by this phase","artifact_path":".kit/harness.db"},{"command":"zharness score-trace (both wave traces)","output_ref":"both scored tier=detailed","artifact_path":".kit/runs/work/20260718-1718-cli-stale-drift.md"}]
created: 2026-07-18
updated: 2026-07-18
---

# CHECK REPORT

Run ID: check-20260718-1745-cli-stale-drift
Scope: full
Artifact Alignment: aligned (one documented, user-approved exception — CONTRACT.md; see notes)
Review Verdict: APPROVE with requests
Phase: cli-stale-drift
Spec: .kit/planning/SPEC.md
Plan: .kit/planning/phases/cli-stale-drift/cli-stale-drift-PLAN.md
Cook Run: .kit/runs/work/20260718-1718-cli-stale-drift.md
Created At: 2026-07-18 17:45

## Gate Evidence
- tests: `cd cli && go test ./... -v` → pass (all 6 packages)
- types/vet: `cd cli && go vet ./...` → pass
- lint: `cd cli && gofmt -l .` → pass (clean)
- build: `cd cli && go build ./...` → pass

## Artifact Alignment
- status: aligned
- notes:
  - **scoping note**: the working tree still carries phase 1 + phase 2's uncommitted output (both already independently gated). This review isolates cli-stale-drift's own changes using the run artifact's per-task "changed files" log as the source of truth, since `git diff HEAD` mixes all three phases' edits together in shared files (resume.go, audit.go, root.go, CONTRACT.md, STATE.md all now carry cumulative diffs). Phase 1/2's own content was not re-reviewed here — already gated on its own merits.
  - spec coverage: R3 (complete) — stale_docs drift, named recovery, clearing semantics all implemented and tested; matches CONTEXT.md's Spec Hooks
  - Locked Decisions cross-checked against code: firing rule (exists ∧ differs ∧ neither `dev`) ✅ tested 6 ways including a pre-migration-schema edge case found during design, not in the original plan; missing-stamp → no drift (chose "omit" over an informational field, per CONTEXT.md's own explicit either/or) ✅; single-source recovery string ✅ byte-identical grep-verified; additive to drift array, no new readiness state ✅
  - boundary compliance: all changed files fall inside Allowed Surfaces (`cli/internal/**`, `cli/docs/STATE.md`, Go test files) with **one exception**: `cli/docs/CONTRACT.md` (added `stale_docs` to the `resume --json` drift-type enum) — same documented-exception pattern already established and user-approved during cli-embed-scaffold's gate; applied here without re-asking since it's the same low-stakes category of decision, already resolved once this initiative
  - proof trail: intact — `trace_ids` populated during execution this time (not backfilled after the fact, unlike cli-embed-scaffold), both scored tier=detailed

## Findings
### Critical
- none

### Major
- **[scope adapted, user-confirmed via AskUserQuestion mid-execution, not silently reduced]** T4's acceptance criteria ("readiness transitions clean → in-progress → checked") is not achievable through the current CLI surface. Investigation (documented in the run artifact) found the gap is larger than initially scoped: `meta.current_phase` is never written by any documented command — not even `work`'s own run-registration changeset — so `resume`'s readiness never leaves `"clean"` through the whole lifecycle chain tested (confirmed empirically: the test's first draft assumed `"in-progress"` and failed against the real binary). `story.status` also never transitions past `"planned"`. Both are the same class of gap already recorded as backlog item `01KXTBG4JZYTW528Y5XZQK8FEH` during cli-embed-scaffold's own gate — confirmed and detailed further here, not newly introduced, not this phase's fix (CONTEXT.md's Goal is drift detection + lifecycle testing, not building missing lifecycle-management commands, which would be a separate, unplanned, initiative-sized change touching the locked 20-command surface). T4 delivered a genuine, real integration test of the lifecycle as it actually behaves, including `validate --json`/`audit --json` fully clean per PLAN.md's literal ask — only the readiness-transition assertion was scoped down to match reality.

### Minor / Suggestions
- none new this phase (T1-T3 landed clean on first or second design pass; process discipline from cli-embed-scaffold's gate — wave traces recorded immediately, not backfilled — was carried forward and held)

## Harness Gate Flow (Validation Matrix, lane=high-risk)
- `unit`: required → satisfied — 6 `TestResumeStaleDocs*` cases (match/differ/dev-written/dev-cli/missing/pre-migration-schema)
- `integration`: required → satisfied — `TestClearingSemantics_RefreshDocsResolvesStaleDocsDrift` (crosses ScaffoldDocs+Resume+real sqlite db); `TestLifecycle_ScratchDirFullChain` (crosses the real compiled binary boundary via `os/exec`)
- `e2e`: optional → satisfied anyway — the lifecycle test is a genuine CLI-level smoke test, reusable as-is by `cli-release`
- `manual-check`: required → satisfied — this Phase 2 review pass; zero unresolved 🔴, one 🟠 (scope-adapted with user sign-off, not a defect)
- `command-output`: required → satisfied — build/vet/gofmt/test all cited above with real output

All required cells satisfied with real evidence.

## Next Action
- Gate clean (with one documented, user-approved scope adaptation). Ready to advance to `cli-release` per the roadmap's dependency order — **stop for explicit user confirmation before the real `git push` + tagged GitHub release**, per this initiative's own carried-forward execution boundary.
- Suggest (not auto-run): commit — three phases (playbook-authoring, cli-embed-scaffold, cli-stale-drift) are now stacked uncommitted. The scoping cost of isolating this phase's diff from the shared working tree (see Artifact Alignment note above) is a concrete, realized instance of the risk flagged at cli-embed-scaffold's own gate — committing before `cli-release` is stronger advice now than it was then.
- Backlog `01KXTBG4JZYTW528Y5XZQK8FEH` (current_phase/story.status transitions) remains open and un-scoped into any of the remaining 3 phases (cli-release, thin-triggers, agent-pilot) — worth a deliberate decision (fix now, defer to a future initiative, or accept permanently) before `agent-pilot` tries to prove a second agent can drive the full lifecycle, since a second agent would hit the exact same "readiness never leaves clean" surprise this phase's T4 did.

---

scope:              on target
depth:              deep
artifact_alignment: ✅ aligned (1 documented, user-approved exception — CONTRACT.md; scoping note re: mixed working tree)
gate:               ✅ pass: build, vet, gofmt, test (all packages, incl. real-binary lifecycle test)
review:             APPROVE with requests
blockers:           0 critical, 1 major (scope-adapted with user sign-off, not a defect)
autofix:            0 safe_auto proposed, 0 gated_auto awaiting confirmation
verification:       go build/vet/gofmt/test ./... → pass; real zharness binary lifecycle test → pass; zharness audit --json → pointer_drift clean after check record
harness_verdict:    zharness check record id 01KXTNYEY80SPB3QZ6NX4D173J (APPROVE_WITH_REQUESTS); meta.latest_check_id updated via changeset
