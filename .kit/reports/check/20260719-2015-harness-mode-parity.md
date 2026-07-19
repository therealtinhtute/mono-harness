---
id: 01KXX2Z2TPJCJW6X1CVH2VPYTJ
type: check
phase: harness-mode-parity
lane: high-risk
run_id: 01KXX1RG84Z8JSW2EG3Y5G6D80
proof_links: [{"command":"zharness --version","output_ref":"0.3.0","artifact_path":""}, {"command":"go build ./... && go test ./...","output_ref":"all packages pass, incl. internal/embedded embed-integrity tests","artifact_path":""}, {"command":"go test ./internal/application/... -run TestValidate -v","output_ref":"9/9 pass (4 new mode-aware tests + 5 pre-existing regression tests unchanged)","artifact_path":"cli/internal/application/validate_test.go"}, {"command":"scratch-dir integration proof: simple-mode chain → validate --json valid:true; full-mode regression → unchanged; negative control (unregistered full-mode run) → correctly valid:false","output_ref":"valid:true (simple), valid:true (full, unchanged), valid:false (negative control)","artifact_path":""}, {"command":"docs_version/stale_docs drift check across a version bump (0.2.0 binary docs vs 0.3.0 binary)","output_ref":"stale_docs fires with correct init --refresh-docs recovery","artifact_path":""}, {"command":"zharness audit --json cross-check: audit.ContractViolations is composed directly from Validate()","output_ref":"confirmed via cli/internal/application/audit.go:56 — no duplicated logic, mode carve-out applies to both consumers automatically","artifact_path":"cli/internal/application/audit.go"}, {"command":"go vet ./... && gofmt -l .","output_ref":"clean, no findings","artifact_path":""}, {"command":"secrets scan grep over full phase diff (24bdc2f..HEAD)","output_ref":"no hits","artifact_path":""}, {"command":"gh release view v0.3.0 --json tagName,assets","output_ref":"5 platform assets + checksums.txt published","artifact_path":""}, {"command":"bash scripts/install-zharness.sh (fresh scratch dir)","output_ref":"resolves v0.3.0, zharness --version confirms","artifact_path":""}, {"command":"zharness init --refresh-docs; zharness resume --json","output_ref":"stale_docs drift cleared on this repo's own .kit/docs/","artifact_path":""}]
created: 2026-07-19
updated: 2026-07-19
---

# CHECK REPORT

Run ID: check-20260719-2015-harness-mode-parity
Scope: full
Artifact Alignment: aligned
Review Verdict: APPROVED
Phase: harness-mode-parity
Spec: .kit/planning/SPEC.md
Plan: .kit/planning/phases/harness-mode-parity/harness-mode-parity-PLAN.md
Cook Run: .kit/runs/work/20260719-1821-harness-mode-parity.md
Created At: 2026-07-19 20:15

## Gate Evidence
- tests: `go test ./...` (full suite) → pass; `go test ./internal/application/... -run TestValidate -v` → 9/9 pass
- types: `go build ./...` → pass (Go build is the type-check gate)
- lint: `go vet ./...` + `gofmt -l .` → clean
- build: `go build ./cmd/zharness` (scratch binary used for integration proof) → pass; release build via goreleaser → pass (`gh run view` all green)

## Artifact Alignment
- status: aligned
- notes:
  - spec coverage: R9's acceptance criterion (`zharness validate --json` passes on the produced chain) is the direct target; scratch-dir proof shows `valid:true` on a simple-mode chain for the first time in this repo's history
  - boundary compliance: diff (24bdc2f..HEAD, 28 files, 968 insertions/25 deletions) stays entirely within `harness-mode-parity-CONTEXT.md`'s Allowed Surfaces (`cli/internal/application/validate.go`+test, `cli/docs/embedded/playbooks/{work,check}.md`, `cli/docs/CONTRACT.md`, release surface, `skills/workflow/README.md` + the 6 spine `SKILL.md` files — the last expanded from the original Forbidden-Surfaces assumption via a recorded correction, see below); `interview/SKILL.md`, `git/**`, and any DB schema migration were correctly left untouched
  - proof trail status: all 3 trace_ids on the RUN artifact score `detailed` tier (not `minimal`) — sufficient evidence class for `unit`/`integration`/`manual-check` cells below
  - one locked-decision correction recorded mid-execution: `harness-mode-parity-CONTEXT.md`'s Forbidden Surfaces assumed the 6 spine `SKILL.md` files symbolically reference `README.md`'s `MIN_ZHARNESS_VERSION` constant; they hardcode it literally per file. Corrected inline in CONTEXT.md's Assumptions (append, not silent rewrite) before editing — same discipline `work.md`'s Output Rules require.

## Findings
### Critical
- none

### Major
- none

### Minor / Suggestions
- `~/.claude/skills/*` (globally installed skill copies used by this session's actual live invocations) have not been resynced to the `0.3.0`-gated `SKILL.md` content — they resync via the documented `npx skills add ... -g -y` installer (`CLAUDE.md` Development Commands), a user action outside this phase's file scope. Visible evidence: this very `check full` re-invocation's preamble still quoted the stale `0.2.0` gate text from before the resync.
- Pre-existing, out-of-phase observation (not introduced by this diff, not actioned here): `.gitignore` ignores `.kit/docs/`, `cli/dist/`, and `**/.validation-report.json` in addition to the `harness.db`/`.kit/cache/` pair literally named in SPEC.md's R2 — a wording gap in R2's text vs. the actual (reasonable — scaffolded/generated content) scaffolding behavior, predating this phase, outside its Allowed Surfaces.
- `zharness audit --json` still surfaces several pre-existing `contract_violations`/`pointer_drift` entries (plan_id nulls on 4 earlier-phase RUN artifacts — GitHub #36; thin-triggers-smoketest's ad-hoc check registration gap — already backlogged `01KXWH4YNC9RRFR1VPE6DK8P14`; HANDOFF.md null run_id/check_id). None reference `harness-mode-parity`'s own artifacts — confirmed by name in the audit output — so none are new findings from this diff; restated here only for continuity with the initiative's running backlog.

## Next Action
- ready for PR / proceed to Phase 8 (`agent-pilot-rerun`) — re-run the second-agent pilot against the now-released `cli/v0.3.0` to confirm R9's literal bar (`validate --json` → `valid:true`) on a genuinely cold simple-mode chain
