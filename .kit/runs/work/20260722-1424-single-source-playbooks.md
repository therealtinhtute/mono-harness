---
id: 01KY4BCT00MYGZFJAW9JP650JE
type: run
phase: single-source-playbooks
lane: low
mode: full
plan_id: none
trace_ids: [01KY4BMQT1KXV6AXBGDWP9D3FS]
created: 2026-07-22
updated: 2026-07-22
---

# COOK RUN

Run ID: work-20260722-1424-single-source-playbooks
Mode: full
Status: passed
Spec: .kit/planning/SPEC.md
Roadmap: .kit/planning/ROADMAP.md
Phase: single-source-playbooks
Plan: .kit/planning/phases/single-source-playbooks/single-source-playbooks-PLAN.md
Started At: 2026-07-22 14:24

## Preflight
- scope drift: no — working tree has only unrelated pre-existing uncommitted handoff bookkeeping (`.kit/HANDOFF.md`, `.kit/harness.db`, 2 changeset files from the prior handoff/run-create calls), none inside this phase's Allowed Surfaces
- working tree note: none blocking
- required artifacts present: yes — `.kit/planning/phases/single-source-playbooks/{single-source-playbooks-CONTEXT.md,single-source-playbooks-PLAN.md}` both exist, `status: ready`, no placeholders
- selected phase: single-source-playbooks (explicit `zharness next full phase single-source-playbooks --json` → `{"mode":"full","active_phase":"single-source-playbooks"}`, no stop)
- dependency check: `zharness query phases --json` — `depends_on: scoring-removal`, which is complete (committed `2d6e2fc`, gated APPROVED); Escalate-If condition ("Phases 1 or 3 aren't finalized") is cleared
- known gap (unrelated, pre-existing, not this phase's job): `stories.status` for write-boundary/dead-surface-removal/scoring-removal all still show `planned` in the DB (no story-status-update command exists) — `current_phase` mismatch treated as bookkeeping drift, not truth, per work.md Step 1

## Wave / Task Log
### Wave 1
#### T1 — Drift-guard test
- status: DONE
- changed files:
  - new: `cli/internal/embedded/projection_drift_test.go` — `TestProjectionDrift_KitDocsMatchesEmbed` walks the embed manifest and asserts each path's content under `.kit/docs/` (this repo's real, git-tracked projection — relative path `../../../.kit/docs` from the test's package dir) is byte-identical to the embed. Fails naming the drifted path + the fix command if not.
- verification:
  - `cd cli && go test ./internal/embedded/ -run Projection -v -count=1` → PASS on the current tree
  - negative-case proof (per plan): appended a line to `.kit/docs/playbooks/check.md`, re-ran with `-count=1` → FAIL, error names the exact drifted path and the fix (`zharness init --refresh-docs`); restored the original content via a saved copy, re-ran → PASS again; `diff` against the pre-corruption copy confirmed byte-identical restore
  - `go build ./... && go vet ./... && gofmt -l . && go test ./... -count=1` → all pass, 7 packages ok (1 no-test-files)
- notes:
  - done directly (single new file, one concept), not delegated
  - **deviation from plan's suggested location**: plan/CONTEXT.md named `embedded_test.go` as "already exists as a home"; put the test in a new file `projection_drift_test.go` in the same package instead — same location in spirit (package `embedded`), just a separate file to keep the new drift-guard concept isolated from the existing manifest/content tests. No import-cycle risk since it only uses `os`/`filepath` + package-local `BuildManifest`/`FS`, never imports `application` (which already imports `embedded` — importing it back would cycle).
  - stop condition "scaffold and embed can't be compared deterministically (line endings, trailing newline)" did not trigger — direct `string(got) != string(want)` byte comparison was sufficient, no normalization needed (embed and `.kit/docs` are both plain UTF-8 markdown with LF endings)

#### T2 — Document the contract
- status: DONE
- changed files:
  - edited: `cli/docs/CONTRACT.md` — added one line under `init`'s side-effects bullet naming the embed as canonical and pointing at the new drift-guard test
  - edited: `README.md` — appended one sentence to the `init` usage note
  - edited: `docs/workflow-harness/migration.md` — appended one sentence to the `init` explainer paragraph
- verification:
  - `grep -rn "edited in .*embed" cli/docs/CONTRACT.md README.md docs/workflow-harness/` → present in all 3 files
- notes:
  - done directly (3 one-line doc edits), not delegated
  - **flagged, not fixed** (out of scope for this task): `docs/workflow-harness/migration.md` line 37 and the root `.gitignore` still say `.kit/docs/` should be *ignored* ("committing it just invites drift"), but `git ls-files .kit/docs` shows it is actually tracked (commit `77ed8bb` "chore(kit): track .kit/docs projection + harness.db cache", predating this phase). This is a real, pre-existing contradiction between stated guidance and repo state — it's exactly why T1's drift-guard test has teeth (a tracked, hand-editable copy is a real drift vector, not hypothetical) — but resolving the gitignore/tracking inconsistency itself is outside this task's Allowed Surfaces and the plan's stated scope; noting it for the check gate and a future cleanup, not silently fixing it here.

## Summary
- Both tasks DONE. `TestProjectionDrift_KitDocsMatchesEmbed` added (`cli/internal/embedded/projection_drift_test.go`), proven to PASS on the current tree and FAIL on an injected drift (then restored). One-line "edit the embed only" rule added to `CONTRACT.md`, `README.md`, `docs/workflow-harness/migration.md`.
- No blocked tasks.
- Unresolved concern (flagged, not fixed, out of this task's scope): `.gitignore` + `migration.md` still say `.kit/docs/` should be ignored, but it is actually git-tracked (`77ed8bb`) — a pre-existing contradiction the new test's real teeth depend on; worth a future small cleanup pass reconciling the two.

## Next Recommended Action
- Phase gate: `.kit/reports/check/20260722-1445-single-source-playbooks.md` — **APPROVED**, 0 critical/major findings, 4 minor/suggestion notes (test-location deviation explained, pre-existing gitignore/migration.md contradiction flagged, expected pointer_drift self-resolved, pre-existing audit debt tail). `zharness check record` ran (id `01KY4BSGY48T93BYNHVGNF3PED`), `meta.latest_check_id` updated atomically, confirmed `pointer_drift` empty on a follow-up `audit --json`.
- `git`: this closes the Harness Subtraction Pass (all 4 phases now done: write-boundary, dead-surface-removal, scoring-removal, single-source-playbooks) — suggest committing, then `handoff`.
