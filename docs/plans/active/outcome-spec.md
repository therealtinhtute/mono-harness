---
status: active
lane: high-risk
---
# Outcome-spec for normal plans

## Goal
- outcome: a `normal` plan is the Goal plus one phase that brainstorm lock writes itself. `to-plan` runs only for `high-risk`. A normal cycle is lock → work → check → handoff.
- success_signal: a normal-lane dry run commits lock → work → check gate → check full → handoff → move to completed through the hook with no to-plan step and no task list; all 4 gate commands pass; a normal cycle reads ≤ 18 KB of playbooks (22.3 KB after the slim, 39.8 KB before).
- actors: owner; contributor agents; consumers who receive the playbooks via `zharness update`.
- authority: owner answers 2026-10-12: normal = Goal + 1 auto phase; skip to-plan for normal; keep the 3 lanes; run as high-risk with a new ADR. Research: `docs/research/2026-10-12-agent-sdlc-practice.md` (option B).
- requirements:
  - R1: ADR 0014 records the decision and amends ADR 0012 decision 2 (a normal plan's shape); existing plans with task lists stay valid, with no migration. | acceptance: `test -f docs/decisions/0014-outcome-spec-normal-plans.md` and ADR 0012 links to it | source: owner
  - R2: `brainstorm-lock.md` writes, for `normal`, a phase `main` whose `phase check:` lists the Goal's acceptance commands, and sets `exact_next_action: work full phase main`. The `high-risk` skeleton is unchanged (`approach: not-planned`, next action `to-plan`). | acceptance: new contract phrases in `cli/src/embedded.rs` pass (red first) | source: owner Q1
  - R3: `to-plan` is for `high-risk` only; for a `normal` plan it stops and routes to `work`. `brainstorm.md`, `WORKFLOW.md` and the `to-plan`/`brainstorm` shims agree. | acceptance: contract phrases pass; `bash scripts/validate-skill.sh` passes on the changed shims | source: owner Q2
  - R4: `work-full.md` runs a phase that has no task list by working each requirement in turn; each `done` Log entry names the requirement and its acceptance check. | acceptance: contract phrase passes; the normal dry run's Log shows `done` per requirement | source: owner Q1
  - R5: tiny and high-risk behave as today. | acceptance: every existing contract phrase still passes; the high-risk dry run from the slim plan still commits through the hook | source: owner Q3
- non-goals:
  - NG1: no change to the guard (ZGUARD-CORE), the hook-parsed headings, or `update` migration.
  - NG2: no change to the tiny or high-risk flow; no new stages, roles, or files besides ADR 0014.

## Phases and Verification
- approach: keep the five sections and one phase per normal plan, so the guard, check preflight and handoff need no change; only lock writes the phase and to-plan steps aside. Contract tests go red first. Rejected: a phaseless normal plan (check preflight, handoff and the guard's completed check would all need a new path).
- risks: consumers' agents read an old to-plan habit → the to-plan playbook routes normal plans to work explicitly; the auto phase misses a check → its phase check is the Goal's acceptance commands, so check-durable step 4 still judges each requirement.
- seams: `cli/src/embedded.rs` contract tests; `cli/tests/projection_parity.rs`; the pre-commit hook (dry runs).
- constraints:
  - Every edited playbook lands byte-identical in `docs/playbooks/`.
  - No existing `required:` phrase is removed; new phrases are added.
- recovery: each phase is its own commit range; revert it to restore the slim playbooks.
- phase_slug: `p1-adr`
  status: checked
  - goal: R1 | depends_on: none
  - surfaces: `docs/decisions/0014-outcome-spec-normal-plans.md`, `docs/decisions/0012-*.md` (one link line), `docs/decisions/README.md` | avoided: playbooks, `cli/`
  - escalate_when: the ADR would need to change a guard-parsed field
  - wave 1:
    - T1 write ADR 0014 and link it from ADR 0012 and the index — output: ADR — check: `bash scripts/verify-doc-links.sh` — stop_if: authority unclear
  - phase check: `bash scripts/verify-doc-links.sh`
- phase_slug: `p2-playbooks`
  status: checked
  - goal: R2, R3, R4, R5 | depends_on: p1-adr
  - surfaces: `cli/src/embedded.rs` (contract arrays only), `brainstorm.md`, `brainstorm-lock.md`, `to-plan.md`, `work-full.md` (+ `docs/playbooks/` copies), `cli/docs/embedded/WORKFLOW.md` + `docs/WORKFLOW.md`, `skills/workflow/{to-plan,brainstorm}/SKILL.md` | avoided: guard, installer, check/handoff playbooks
  - escalate_when: a change needs an existing pinned phrase removed or a guard edit
  - wave 1:
    - T2 add the new contract phrases for R2-R4 and run them red — output: failing test naming each phrase — check: `cd cli && cargo test --lib embedded` — stop_if: fails for another reason
  - wave 2:
    - T3 edit the playbooks, WORKFLOW and shims until green — output: green — check: `cd cli && cargo test` — stop_if: a pinned phrase must go
    - T4 copy to `docs/` — check: `cd cli && cargo test --test projection_parity`
  - phase check: `cd cli && cargo fmt --check && cargo clippy --all-targets -- -D warnings && cargo test`
- phase_slug: `p3-dry-run`
  status: in-progress
  - goal: R4, R5, success_signal | depends_on: p2-playbooks
  - surfaces: a throwaway worktree and branch (never pushed) | avoided: everything else
  - escalate_when: the hook rejects a normal-lane step the playbooks prescribe
  - wave 1:
    - T5 normal-lane dry run: lock (with auto phase) → work → check gate → check full → handoff → move to completed — output: every commit passes the hook — check: transcript plus `git log` of the scratch branch
    - T6 high-risk regression dry run (to-plan path) — output: same — check: same
    - T7 measure playbook bytes for a normal cycle — output: ≤ 18 KB — check: `wc -c` of the files read
  - phase check: `cd cli && cargo fmt --check && cargo clippy --all-targets -- -D warnings && cargo test && cd .. && bash scripts/test-guards.sh && bash scripts/verify-doc-links.sh`

## Log
- 2026-10-12T23:35Z — p1-adr — start — baseline `bash scripts/verify-doc-links.sh` clean
- 2026-10-12T23:35Z — p1-adr/T1 — done — `bash scripts/verify-doc-links.sh` 0 findings; ADR 0014 added, linked from ADR 0012 Status and the index
- 2026-10-12T23:36Z — p2-playbooks — start — baseline `cd cli && cargo test` green
- 2026-10-12T23:38Z — p2-playbooks/T2 — done — `cd cli && cargo test --lib embedded` red first: missing "`normal`: Goal + one phase `main`, no `to-plan`"; 5 new contract phrases added
- 2026-10-12T23:38Z — p2-playbooks/T3 — done — `cd cli && cargo test` green (49+2); brainstorm, brainstorm-lock (normal phase block), to-plan (high-risk only), work-full (no-task phase), work, WORKFLOW, 3 shims (versions bumped)
- 2026-10-12T23:38Z — p2-playbooks/T4 — done — `cd cli && cargo test --test projection_parity` pass
- 2026-10-12T23:38Z — p2-playbooks — decision — also edited skills/workflow/README.md (outside listed surfaces) so the spine description matches ADR 0014; lane names un-backticked in shim Defer lines because validate-skill.sh reads backticks there as skill names
- 2026-10-12T23:39Z — p2-playbooks — decision — judge transcript reformatted to the guard format (it used `verdict: pass`); requests 1-2 fixed right after the gate (lock-written phase immutable; grill Done names lock for normal); request 3: `bash scripts/validate-skill.sh` passed on to-plan, brainstorm, work; final check full covers these edits
- 2026-10-12T23:39Z — p3-dry-run — start — baseline `cd cli && cargo test` green
- 2026-10-12T23:40Z — p3-dry-run/T5 — done — normal dry run (throwaway worktree, never pushed): lock with auto phase `main` → work (`done` per R1) → check full (proof re-executed) → handoff close → git mv, every commit passed the hook; no to-plan, no task list
- 2026-10-12T23:40Z — p3-dry-run/T6 — done — high-risk regression: lock → to-plan → work → check full committed; same-session judge rejected (R3 JUDGE GUARD)
- 2026-10-12T23:40Z — p3-dry-run/T7 — done — normal cycle reads 30,596 B (lock + work + same-session gate + independent full + handoff), 23,114 B without the separate gate; target 18 KB missed: the 18 KB estimate left out the cold check-full read, which every initiative close requires
- 2026-10-12T23:40Z — p3-dry-run — decision — success_signal byte target missed; asking the owner whether a single-phase plan may skip the same-session gate and go straight to the independent check full (would land at 23.1 KB)

## Validation
- 2026-10-12T23:36Z — phase `p1-adr` — verdict: APPROVED — mode: gate
  - `bash scripts/verify-doc-links.sh` — doc links OK (0 findings)
  - `test -f docs/decisions/0014-outcome-spec-normal-plans.md` — exit 0
  - scope: on target — ADR 0014, ADR 0012 Status line, index rows
  - requirements: R1 met (0014 exists; 0012:5 links it; 0014 keeps task-list plans valid, no migration; five headings present)
  - rollback_point: 4a74a34
  - requests: 1. 0014:9 attributes the task list/to-plan to 0012:31, which says only "one flat phase" (non-blocking); 2. 0014:28 cites :289, heading match is at :290 (non-blocking)
  - judge: independent
  - judge_model: claude-opus-5-5
- 2026-10-12T23:39Z — phase `p2-playbooks` — verdict: APPROVE_WITH_REQUESTS — mode: gate
  - `cd cli && cargo fmt --check && cargo clippy --all-targets -- -D warnings && cargo test` — 49 + 2 passed
  - `bash scripts/verify-doc-links.sh` — 0 findings
  - `diff -r docs/playbooks cli/docs/embedded/playbooks` — identical
  - scope: on target — embedded.rs adds 6 lines only; WORKFLOW copies identical
  - requirements: R2 met (brainstorm-lock normal block; phrases pinned) | R3 met (brainstorm.md, to-plan.md, WORKFLOW, shims agree) | R4 partial (→ p3-dry-run) | R5 partial (→ p3-dry-run)
  - rollback_point: 6b897e6
  - requests: 1. brainstorm-lock.md:53 immutability names only to-plan, not the lock-written `main` phase; 2. brainstorm-grill.md:17 Done criterion names only to-plan; 3. record validate-skill.sh evidence for R3
  - judge: independent
  - judge_model: claude-opus-5-5

## Current State and Next Action
- active_phase: p3-dry-run
- lifecycle_status: in-progress
- blockers: owner decision on the byte miss (accept 30.6 KB, or skip the duplicate gate for single-phase plans → 23.1 KB)
- open_items: commit 4a74a34 (slim close) rides this branch and lands with this work
- exact_next_action: owner decision, then independent check full
