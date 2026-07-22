---
id: 01KY44BHA4WJ6HQBV7M3BJNZW2
type: run
phase: scoring-removal
lane: high-risk
mode: full
plan_id: none
trace_ids: [01KY44F8KZ9MJ32SCPH6QT9Y5E, 01KY44MX7RXCQ1V8QB62654YSR]
created: 2026-07-22
updated: 2026-07-22
---

# COOK RUN

Run ID: work-20260722-1300-scoring-removal
Mode: full
Status: complete
Spec: .kit/planning/SPEC.md
Roadmap: .kit/planning/ROADMAP.md
Phase: scoring-removal
Plan: .kit/planning/phases/scoring-removal/scoring-removal-PLAN.md
Started At: 2026-07-22 13:00

## Preflight
- scope drift: no — working tree clean at phase start (`git status --short --branch` clean, in sync except 2 local commits ahead of origin from dead-surface-removal, not pushed)
- working tree note: none
- required artifacts present: yes — `.kit/planning/phases/scoring-removal/{scoring-removal-CONTEXT.md,scoring-removal-PLAN.md}` both exist, `status: ready`, no placeholder markers
- selected phase: scoring-removal (explicit `zharness next full phase scoring-removal --json` → `{"mode":"full","active_phase":"scoring-removal"}`, no stop)
- pre-verify: `grep -rn "score-trace\|entropy_score\|ScoreTrace\|EntropyScore" cli/docs/CONTRACT.md` confirms only 2 CONTRACT.md sections + 1 mapping-table row reference `score-trace`/`audit`'s `entropy_score` shape — matches plan's Wave 2 scope. Real project `.kit/harness.db` is still at `schema_version 2` (migration `0003_drop_dead_surface` from dead-surface-removal not yet applied via `zharness migrate` against this project's own db) — noted as a pre-existing gap from the prior phase, out of scope here since it doesn't block scoring-removal's own build/test/CLI-surface work.

## Wave / Task Log
### Wave 1
#### T1 — Remove scoring from the CLI
- status: DONE
- changed files:
  - deleted: `cli/internal/interfaces/score.go` (`newScoreTraceCmd`/`runScoreTrace`, whole-file — only content was the `score-trace` command)
  - deleted: `cli/internal/application/score.go` (`ScoreTrace`, `TraceScore`, `loadTrace`, `countTracesForRun` — all four were `score-trace`-only, confirmed unused by a clean `go build ./...` post-delete, no other caller)
  - deleted: `cli/internal/application/score_test.go` (all 4 tests were `ScoreTrace`-only)
  - edited: `cli/internal/interfaces/root.go` — removed `root.AddCommand(newScoreTraceCmd())`
  - edited: `cli/internal/application/audit.go` — removed `EntropyScore` field from `AuditReport`, its doc-comment sentence, the `entropyScore(...)` call in `Audit`, and the `entropyScore` function itself
  - edited: `cli/internal/interfaces/audit.go` — updated `Short` text (dropped "and an entropy score"), removed `entropy_score=` from the human-readable output line
  - edited: `cli/internal/application/audit_test.go` — removed all `EntropyScore`/`entropy_score` assertions (3 spots: `TestAuditCleanState`, `TestAuditUnlinkedProofFixture` baseline + after), adjusted one stale doc-comment ("moves the entropy score" → "lists the finding")
  - edited: `cli/cmd/zharness/lifecycle_test.go` — removed `EntropyScore` field from the anonymous `auditResp` struct and its assertion
- verification:
  - `go build ./...` → pass (exit 0, no output)
  - `go vet ./...` → pass
  - `gofmt -l .` → empty (no formatting drift)
  - `go test ./...` → all 7 packages pass (`cmd/zharness`, `internal/application`, `internal/domain`, `internal/embedded`, `internal/infrastructure`, `internal/interfaces` pass; `docs/embedded` no-test)
  - `zharness --help | grep -i score` (built fresh to `/tmp/zharness-test`) → empty, confirms `score-trace` fully gone from the command surface
  - `zharness audit --json` on a fresh scratch `init` → `{"pointer_drift":[],"contract_violations":[...one missing_key...],"unlinked_proofs":[]}` — no `entropy_score` key, matches plan's expected output
- notes:
  - done directly (small, focused removal across a known set of files), not delegated
  - `TraceScore`'s removal did not break any unrelated caller — `go build ./...` passing after the delete is the negative-space proof (plan's "stop if: removing TraceScore breaks an unrelated caller" did not trigger)

### Wave 2
#### T2 — Update check playbook + contract docs, prove gate unchanged
- status: DONE
- changed files:
  - edited: `cli/docs/embedded/playbooks/check.md` — Step 4: deleted bullet 3 (the `score-trace` loop), renumbered 4→3, 5→4, 6→5; bullet 2 (now unchanged wording except) dropped "and `entropy_score`" from the `unlinked_proofs is informational` sentence; Command Reference: removed the `zharness score-trace {id} --json` line. Validation Matrix table (lines ~112-127) left byte-for-byte unchanged — confirmed via `git diff` (see verification).
  - edited: `cli/docs/CONTRACT.md` — removed `### score-trace <trace-id>` section entirely; `audit`'s `--json` shape line dropped `"entropy_score": N`; Workflow-Step → CLI-Action Mapping table's `check: gate evaluation` row dropped `+ score-trace <id> --json`
  - edited: `cli/docs/SCHEMA.md` — cross-check paragraph: removed `score-trace` from the read-only command list; command-count line updated `14`→`13` commands, read-only count `5`→`4` (preserving the same pre-existing baseline methodology/undercount noted in dead-surface-removal's check report, not silently fixing it — only decrementing for this phase's actual removal)
  - re-scaffolded `.kit/docs/` from the embed via `zharness init --refresh-docs --json` (run from the repo root, `/home/tinhpt/Lab/skills`) — **side note**: this incidentally applied the pending `0003_drop_dead_surface` migration (schema_version 2→3) to the real project `.kit/harness.db`, resolving the pre-existing gap flagged in this run's own Preflight section (init's migrate-on-open behavior, not something scoring-removal set out to do, but a beneficial side effect of using the standard re-scaffold command)
- verification:
  - `go build ./... && go vet ./... && gofmt -l . && go test ./...` (from `cli/`) → all pass, all 7 packages ok (internal/embedded's drift-check test confirms `.kit/docs/playbooks/check.md` now matches the updated embed byte-for-byte)
  - `grep -n "score-trace\|entropy_score" cli/docs/embedded/playbooks/check.md cli/docs/CONTRACT.md .kit/docs/playbooks/check.md` → empty across all three
  - `git diff --unified=1 -- cli/docs/embedded/playbooks/check.md` → confirms only the score-trace bullet + Command Reference line were touched; the Validation Matrix table itself is absent from the diff, proving it's untouched
  - `zharness audit --json` (both a fresh scratch `init` and the real project db via `/tmp/zharness-test`) → no `entropy_score` key in either; real-project run additionally surfaced one expected `pointer_drift: out_of_order` (latest_check stale vs. this phase's new `latest_run_id` — resolves via this phase's own `check record` call, same pattern as dead-surface-removal's gate) plus a long tail of pre-existing unrelated debt (older checks' unlinked proofs, non-ULID legacy ids) — none touch this diff's files
  - **Gate-unchanged proof** (manual matrix walkthrough, per plan's T2 verification — the matrix is prose evaluated by an agent per Step 4, not Go code, so there is no automated fixture to run): for lane `high-risk`, the Validation Matrix requires `unit`, `integration`, `manual-check`, `command-output` (all `required`; only `e2e` is `optional`) — unchanged table, confirmed above. Walkthrough: if this session's gathered proof omitted `integration` (e.g., only `unit`+`command-output`+`manual-check` were gathered), Step 4 bullet 3 ("Evaluate the Validation Matrix... A required cell with no matching evidence ⇒ gate FAIL, name the exact missing evidence class, and stop") still fires exactly as before this phase's edit — the removed bullet (old #3, the `score-trace` loop) was never part of the matrix lookup or its required/optional cell values, only fed the (now-removed) trace-evidence-tier gate for score-carrying cells, which no cell in the table ever referenced directly. This session's own actual proof set (`unit`: `go test ./...`; `integration`: `TestRunCreateReplaySafety`/`db`-backed tests already exercise real sqlite; `command-output`: build/vet/gofmt/grep output above; `manual-check`: this review pass) satisfies all 4 required cells for `high-risk` — gate does not FAIL for this phase's own diff.
- notes:
  - done directly, not delegated
  - stop condition "the matrix stops FAILing correctly without score-trace" did not trigger — the matrix's cell values and required/optional/n-a shape are identical before and after this phase's edit

## Summary
- Both waves DONE. `score-trace` command + `ScoreTrace`/`TraceScore` application logic + `entropy_score`/`EntropyScore` field fully removed from CLI, docs (`CONTRACT.md`, `SCHEMA.md`), and the embedded `check.md` playbook (re-scaffolded into `.kit/docs/`). Lane×proof Validation Matrix left byte-for-byte unchanged — confirmed via diff and a manual gate-unchanged walkthrough. Incidental side effect: `zharness init --refresh-docs` applied the pending dead-surface-removal migration (schema_version 2→3) to the real project db.
- Phase gate: `.kit/reports/check/20260722-1330-scoring-removal.md` — **APPROVED**, 0 critical/major findings, 2 minor/suggestion notes (pre-existing audit debt tail, incidental schema migration side effect). `zharness check record` ran (id `01KY44QD3ZKDNG4P3H5Z7Y3AM9`), `meta.latest_check_id` updated atomically, the one new `pointer_drift: out_of_order` finding confirmed resolved by a follow-up `zharness audit --json`.

## Next Recommended Action
- `git`: single commit — this phase's diff is one concern (scoring-removal refactor + its own docs/bookkeeping), no unrelated bookkeeping mixed in this time (unlike dead-surface-removal's 2-commit split).
- After commit: continue to Harness Subtraction Pass Phase 4 (`single-source-playbooks`, the final phase) if the user wants to keep going, or `handoff` to close out the session.
