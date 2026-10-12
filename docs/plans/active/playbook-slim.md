---
status: active
lane: normal
---
# Playbook slim

## Goal
- outcome: the playbooks, WORKFLOW.md and the 6 stage shims say each rule once and say nothing else.
- success_signal: a normal 1-phase cycle reads ≤ 30 KB of playbooks (53.3 KB today); all gate commands pass; a dry-run cycle behaves like today.
- actors: owner; contributor agents; consumers who run `zharness update`.
- authority: owner, 2026-10-12: scope = playbooks + workflow shims, keep `check-validation.md` as its own file, "simple is the best"; audit https://github.com/therealtinhtute/mono-harness/pull/130; `docs/research/2026-10-12-agent-sdlc-practice.md`.
- requirements:
  - R1: cut every sentence that is not a step, a routing rule, or a rule a guard/test enforces or a known failure needs (the "why" goes in the commit message). Drop Exit Conditions sections; each invariant has one owning playbook. | acceptance: bytes ≤ the PR 130 targets; every `required:` phrase in `cli/src/embedded.rs` still present | source: PR 130, research N1/N3/N4
  - R2: add three lines to `work-full.md`: never weaken a test to go green; run the phase check once before the first task; a sub-agent brief = pointers, output path, check, budget, write scope (one worktree per parallel writer). | acceptance: the three lines exist; `check-review.md` and `brainstorm-grill.md` point to the brief | source: research S1, N1, S2
  - R3: trim WORKFLOW.md, the AGENTS block and the shared question-tool line in the 6 shims. | acceptance: WORKFLOW ≤ 1,500 B; each stage shim ≤ 1,000 B; `bash scripts/validate-skill.sh` passes | source: PR 130
- non-goals:
  - NG1: no new files, stages, roles, reviewers or installer changes.
  - NG2: no step renumbering; no edit to ZGUARD-CORE; `git`/`hunt`/`think`/`retro` untouched.

## Phases and Verification
- approach: check playbooks first (biggest, read twice per cycle); if they hit target with tests green, do the rest, then a dry run. One commit per task, so any regression reverts cleanly. Rejected: a rationale ledger file (adds a file to maintain; the commit message carries the why).
- risks: a cut changes behavior → the dry run (T4) is the check; stop and restore that commit if a stage diverges.
- seams: `cli/src/embedded.rs` contract tests; `cli/tests/projection_parity.rs`; `scripts/validate-skill.sh`.
- phase_slug: `slim`
  status: in-progress
  - goal: R1, R2, R3 | depends_on: none
  - surfaces: `cli/docs/embedded/` + `docs/` copies, the 6 stage shims | avoided: `cli/src/`, other skills
  - escalate_when: a target needs a pinned phrase or a guard rule removed
  - T1 slim the 4 check playbooks (+ copies) — check: `cd cli && cargo test`
  - T2 slim the other 8 playbooks; add the R2 lines (+ copies) — check: `cd cli && cargo test`
  - T3 trim WORKFLOW, the AGENTS block, the 6 shims — check: `bash scripts/validate-skill.sh` on each shim
  - T4 dry run brainstorm → to-plan → work → check → handoff on a scratch branch; count bytes read — check: every stage completes, ≤ 30 KB
  - phase check: `cd cli && cargo fmt --check && cargo clippy --all-targets -- -D warnings && cargo test && cd .. && bash scripts/test-guards.sh && bash scripts/verify-doc-links.sh`

## Log
- 2026-10-12T15:36Z — slim — start — owner chose option A (slim, keep lanes); T1 check playbooks first
- 2026-10-12T15:37Z — slim/T1 — done — `cd cli && cargo test` pass (49+2); check path 14,472 → 9,415 B; gate path 11,451 → 7,059 B (target 6,500 missed: pinned phrases ~1.6 KB); surfaces: 4 check playbooks + copies

## Validation

## Current State and Next Action
- active_phase: slim
- lifecycle_status: in-progress
- blockers: none
- open_items: none
- exact_next_action: work slim (T2)
