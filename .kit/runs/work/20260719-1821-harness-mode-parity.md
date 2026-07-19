---
id: 01KXX1RG84Z8JSW2EG3Y5G6D80
type: run
phase: harness-mode-parity
lane: high-risk
plan_id: 01KXX1PPNZV87JQSS6JAF2QEJR
trace_ids: [01KXX20P82K47DTYCRDF369W8Y, 01KXX24V8KK1T2915X1H9QV0CE]
created: 2026-07-19
updated: 2026-07-19
---

# COOK RUN

Run ID: work-20260719-1821-harness-mode-parity
Mode: full
Status: running
Spec: .kit/planning/SPEC.md
Roadmap: .kit/planning/ROADMAP.md
Phase: harness-mode-parity
Plan: .kit/planning/phases/harness-mode-parity/harness-mode-parity-PLAN.md
Started At: 2026-07-19 18:21

## Preflight
- scope drift: no
- working tree note: only `.kit/planning/` (ROADMAP.md amendment + new phase dirs) and pilot-evidence doc from the prior `to-plan`/`agent-pilot` work are staged; nothing in `cli/**` touched yet
- required artifacts present: yes (SPEC.md locked, ROADMAP.md amended, CONTEXT.md + PLAN.md written this session)
- selected phase / source prompt: harness-mode-parity, per user's explicit "Full fix" authorization

## Wave / Task Log
### Wave 1
#### T1 — validate.go mode-awareness
- status: DONE
- changed files:
  - cli/internal/application/validate.go
  - cli/internal/application/validate_test.go
  - cli/testdata/chain-simple-mode/{planning/SPEC.md, runs/work/20260101-1200-sample-task.md, reports/check/20260101-1300-sample-task.md}
- verification:
  - `go build ./...` → pass
  - `go test ./internal/application/... -run TestValidate -v` → pass (9/9, including 4 new mode-aware tests + all 5 pre-existing regression tests unchanged)
- notes:
  - simple-mode RUN/CHECK skip phase-existence, plan_id-ULID, and DB stale_pointer checks; `id` ULID-shape still enforced unconditionally

#### T2 — work.md mode-aware run registration
- status: DONE
- changed files:
  - cli/docs/embedded/playbooks/work.md
- verification:
  - `grep -n "mode:\|story_slug" work.md` → simple-mode branch present, full-mode branch unchanged, header contradiction resolved
- notes:
  - RUN frontmatter template gains `mode: {full|simple}`; Step 2 explicitly branches, simple mode skips changeset registration entirely

#### T3 — check.md mode-aware check registration
- status: DONE
- changed files:
  - cli/docs/embedded/playbooks/check.md
- verification:
  - `grep -n "mode:\|check record" check.md` → simple-mode skip branch present, full-mode branch unchanged
- notes:
  - persisted report frontmatter gains `mode: {full|simple}` (inherited from gated RUN); Step 4 skips `check record` for simple-mode-gated runs; also resolves check-side twin (backlog `01KXWH4YNC9RRFR1VPE6DK8P14`, GitHub #30 root cause)

### Wave 2
#### T4 — CONTRACT.md documentation
- status: DONE
- changed files:
  - cli/docs/CONTRACT.md
- verification:
  - `grep -n "Issue:" cli/internal/application/validate.go` cross-checked against the documented enum — every emitted issue string (`missing_key`, `broken_link`, `stale_pointer`, `not_yet_implemented`) now appears in CONTRACT.md's `validate` entry
- notes:
  - fixed pre-existing doc/code drift (`not_yet_implemented` was already emitted, never documented) plus the new mode-aware carve-out paragraph

#### T5 — Scratch-dir integration proof
- status: DONE
- changed files:
  - none (scratch dir outside this repo, per phase boundary — not committed)
- verification:
  - Built `zharness dev` binary; in a fresh scratch dir: simple-mode RUN+CHECK (hand-written per work.md/check.md's new simple-mode branch, no DB registration) → `zharness validate --json` returns `{"valid":true,...}` — first time ever for a simple-mode chain
  - Full-mode RUN+CHECK (real `story`, changeset-registered run, `check record`-registered check) coexisting in the same scratch dir → `valid:true`, unchanged from pre-phase behavior
  - Negative control: full-mode RUN with no DB registration → `valid:false` with `stale_pointer`, confirming the mode gate is precise, not a blanket loosening
  - `zharness audit --json` on the final scratch state → `pointer_drift: []`, `entropy_score: 13`
- notes:
  - all four assertions from the PLAN's T5 steps captured verbatim above
  - additional check (advisor-prompted): docs_version is literally the CLI's own release version string (`-X main.version` via goreleaser), not a separate counter — verified by building two binaries at `0.2.0`/`0.3.0`, confirming `resume --json` reports `stale_docs` drift with the correct `zharness init --refresh-docs` recovery when the newer binary meets docs stamped by the older one. Shipping any new, distinct version tag in T6 is sufficient; no separate docs_version bump needed.
  - full suite re-run per PLAN T6 step 1: `go build ./... && go test ./...` → all packages pass, including `internal/embedded` (embed-integrity tests, confirms no accidental drift in the embedded doc set)

### Wave 3
#### T6 — Release cli/v0.3.0
- status: DONE
- changed files:
  - none (git tag + GitHub release; local commits pushed)
- verification:
  - user confirmed via AskUserQuestion: push now, version v0.3.0 (minor — changes documented CONTRACT.md behavior + embedded playbook content, not a pure internal bugfix)
  - `git push origin master` (12 commits, phases 1-7 — none had been pushed before this phase) then `git tag cli/v0.3.0 && git push origin cli/v0.3.0`
  - GitHub Actions `cli-release` workflow run 29685388973 triggered; tracked via `gh run watch`
- notes:
  - master was 12 commits behind origin before this push — all prior phases' work (1-6) shipped to the remote for the first time as part of this push, not just this phase's changes

#### T7 — MIN_ZHARNESS_VERSION bump
- status: DONE
- changed files:
  - skills/workflow/README.md
  - skills/workflow/{watzup,brainstorm,to-plan,work,check,handoff}/SKILL.md
- verification:
  - `grep -rn "MIN_ZHARNESS_VERSION" skills/workflow/README.md skills/workflow/*/SKILL.md` → 6 spine skills + README.md now read `0.3.0`; `interview/SKILL.md` correctly untouched (stale `0.1.0`, pre-existing, out of scope per SPEC R8, already backlogged)
  - `cd cli && go build ./... && go test ./...` → clean after these doc-only changes (no code touched)
- notes:
  - **Assumption correction during execution**: `harness-mode-parity-CONTEXT.md`'s Forbidden Surfaces assumed the 6 spine SKILL.md files symbolically reference README.md's constant. On inspection they hardcode the literal version string per file — bumping only README.md's prose would have left every skill's actual gate check still passing a buggy `0.2.0` binary. Corrected inline in CONTEXT.md's Assumptions section before editing; the mechanical fix (`0.2.0`→`0.3.0`, 6 files, same pattern `thin-triggers` used for `0.1.0`→`0.2.0`) is a one-line-per-file string replacement, not new scope.
  - `~/.claude/skills/*` (globally installed copies outside this repo) were deliberately NOT hand-edited — they resync via the documented `npx skills add ... -g -y` installer (`CLAUDE.md` Development Commands), not ad-hoc patching of untracked files. Flagged to the user as a follow-up.

