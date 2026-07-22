---
id: 01KY4BR73A4ZFJXQ4MQP6AE5M1
type: check
phase: single-source-playbooks
lane: high-risk
mode: full
run_id: 01KY4BCT00MYGZFJAW9JP650JE
proof_links: [{"command":"go build ./...","output_ref":"exit 0, no output","artifact_path":"cli/"},{"command":"go vet ./...","output_ref":"exit 0, no output","artifact_path":"cli/"},{"command":"gofmt -l .","output_ref":"empty, no formatting drift","artifact_path":"cli/"},{"command":"go test ./... -count=1","output_ref":"ok across cmd/zharness, internal/application, internal/domain, internal/embedded, internal/infrastructure, internal/interfaces (docs/embedded no-test)","artifact_path":"cli/"},{"command":"go test ./internal/embedded/ -run Projection -v -count=1","output_ref":"PASS on current tree; FAIL when .kit/docs/playbooks/check.md was deliberately drifted (error named the exact path + fix command); PASS again after restore, diff confirmed byte-identical","artifact_path":"cli/internal/embedded/projection_drift_test.go"},{"command":"go test ./internal/application/ -run TestInit_FreshScratchDir_FullIntegration -v","output_ref":"PASS — real db + real fs scaffold, satisfies integration proof class","artifact_path":"cli/internal/application/scaffold_integration_test.go"},{"command":"grep -rn \"edited in .*embed\" cli/docs/CONTRACT.md README.md docs/workflow-harness/","output_ref":"present in CONTRACT.md, README.md, migration.md","artifact_path":"cli/docs/CONTRACT.md"},{"command":"zharness audit --json","output_ref":"one expected pointer_drift: out_of_order (latest_run_id now this phase's run, latest_check stale — resolves via this report's check record); long pre-existing contract_violations/unlinked_proofs tail from earlier phases, none touching this diff's files","artifact_path":"cli/internal/application/audit.go"},{"command":"git ls-files .kit/docs; cat .gitignore","output_ref":".kit/docs/ is git-tracked (commit 77ed8bb) despite .gitignore and migration.md line 37 saying it should be ignored — pre-existing contradiction, flagged not fixed","artifact_path":".gitignore"}]
created: 2026-07-22
updated: 2026-07-22
---

# CHECK REPORT

Run ID: check-20260722-1445-single-source-playbooks
Scope: full
Artifact Alignment: aligned
Review Verdict: APPROVED
Phase: single-source-playbooks
Spec: .kit/planning/SPEC.md
Plan: .kit/planning/phases/single-source-playbooks/single-source-playbooks-PLAN.md
Cook Run: .kit/runs/work/20260722-1424-single-source-playbooks.md
Created At: 2026-07-22 14:45

## Gate Evidence
- tests: `go test ./... -count=1` → pass (6 packages ok, 1 no-test-files); `-run Projection` negative-case proof (drift injected → FAIL, restored → PASS) also captured
- types: n/a (Go — covered by build)
- lint: `go vet ./...` → pass; `gofmt -l .` → pass
- build: `go build ./...` → pass

## Artifact Alignment
- status: aligned
- notes:
  - **Spec coverage**: implements Requirement R4 (single-source playbooks) — `.kit/docs/playbooks/*` is now guarded against divergence from `cli/internal/embedded/`'s Go embed by `TestProjectionDrift_KitDocsMatchesEmbed`; the "edit the embed only" rule is now stated in `CONTRACT.md`, `README.md`, and `docs/workflow-harness/migration.md`. Acceptance criterion "a test asserts `.kit/docs/playbooks/*` == embed byte-for-byte" is met literally.
  - **Boundary compliance**: changed files — `cli/internal/embedded/projection_drift_test.go` (new), `cli/docs/CONTRACT.md`, `README.md`, `docs/workflow-harness/migration.md` — all inside Allowed Surfaces (test lives in `cli/internal/embedded/`; the three docs are named explicitly). Forbidden Surfaces respected: no playbook *content* edited, no changeset format/schema/scoring/entity code touched. `cli/internal/application/init.go` (also an Allowed Surface) was left untouched — it already implements the projection correctly; Allowed Surfaces is a ceiling, not a requirement to touch every listed file.
  - **Execution Proof Alignment** (high-risk lane, all 4 required cells covered): `unit` — `go test ./...`; `integration` — `TestProjectionDrift_KitDocsMatchesEmbed` reads the real on-disk `.kit/docs/` tree (not a fixture), and `TestInit_FreshScratchDir_FullIntegration` exercises a real db + real fs scaffold; `manual-check` — this review pass, 0 critical/major findings; `command-output` — build/vet/gofmt/test output above. `e2e` (optional for this lane) not gathered, not required.
  - **Decision/Context Alignment**: matches Locked Decisions in `single-source-playbooks-CONTEXT.md` — embed is canonical, `.kit/docs/` never hand-edited, a Go test (not a new CLI command, not a pre-commit hook) enforces it. Neither rejected option ("commit `.kit/docs` and treat as canonical", "a git pre-commit hook instead of a test") was reintroduced.
  - Uncommitted diff also carries this run's own bookkeeping (3 new `.kit/changesets/*.jsonl` from `run create`+`trace add`(x1 so far)+prior handoff record, `.kit/runs/work/20260722-1424-single-source-playbooks.md`, `.kit/harness.db`) plus leftover unrelated handoff bookkeeping from the prior session turn (`.kit/HANDOFF.md` update, one earlier changeset) — none of it conflicts with this phase's scope.

## Findings
### Critical
- none

### Major
- none

### Minor / Suggestions
- 💡 Test file location deviates from `single-source-playbooks-CONTEXT.md`'s suggested home (`embedded_test.go` "already exists as a home") — placed in a new file `projection_drift_test.go` in the same package instead, to avoid an import cycle (`application` already imports `embedded`; the plan's alternative framing of comparing via `ScaffoldDocs` would require `embedded` to import `application` back). Same package, same spirit, documented in the run artifact.
- 💡 Pre-existing, unrelated to this diff: `.gitignore` and `docs/workflow-harness/migration.md` (pre-this-phase text) say `.kit/docs/` should be *ignored* ("committing it just invites drift"), but `git ls-files .kit/docs` shows it has been git-tracked since commit `77ed8bb`, predating this phase. Not fixed here (outside this task's stated scope), but it's exactly the reason the new drift-guard test has real teeth — a tracked, hand-editable copy is a live drift vector, not hypothetical. Worth a future small cleanup reconciling the gitignore entry with actual practice.
- 💡 Expected `pointer_drift: out_of_order` in `zharness audit --json` (latest_run_id now points to this phase's run, latest_check still scoring-removal's) — self-resolves once this report's `check record` call runs, same pattern observed in every prior phase's gate.
- 💡 Long-standing `audit --json` `contract_violations`/`unlinked_proofs` tail from phases predating this initiative (slim-playbooks, cli-release, etc.) — none touch this diff's files, not this phase's responsibility.

## Next Action
- ready for PR — suggest `git` then `handoff` (this closes the Harness Subtraction Pass — all 4 phases now done)
