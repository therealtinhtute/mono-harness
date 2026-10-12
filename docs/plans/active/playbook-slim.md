---
status: active
lane: high-risk
---
# Playbook slim + multi-agent fit

## Goal
- outcome: the 12 embedded playbooks, WORKFLOW.md, the AGENTS block and the 6 stage shims state each invariant once, and they gain three single-line multi-agent rules taken from agent-SDLC practice current to Oct 2026 (Claude 5 generation). Every rule that remains is either enforced by a guard or justified in a rationale ledger that agents do not load.
- success_signal: a normal 1-phase cycle reads ≤ 30 KB of playbooks (53.3 KB today); `cd cli && cargo test` and `bash scripts/verify-doc-links.sh` pass; a dry-run cycle on a scratch plan behaves like it does today.
- actors: owner; contributor agents; consumers who run `zharness update`.
- authority: owner answers 2026-10-12 (scope = playbooks + workflow skills; research + plan before any edit; keep `check-validation.md` as its own file); audit https://github.com/therealtinhtute/mono-harness/pull/130; `docs/research/2026-10-12-agent-sdlc-practice.md`.
- requirements:
  - R1: each playbook is rewritten as routing + numbered steps. Exit Conditions sections and rationale prose are dropped, and each invariant has one owning playbook. | acceptance: per-file bytes ≤ the audit's targets (check-durable 2,300, check-validation 1,700, the rest as listed in PR 130); every required phrase in `cli/src/embedded.rs` still present | source: PR 130
  - R2: delegation uses one 5-field brief (pointers in, output path, check, time budget + stop rule, write scope). Parallel writers each get their own worktree or branch. | acceptance: the brief appears once in `work-full.md`; `check-review.md` and `brainstorm-grill.md` point to it; `rg -c "worktree" cli/docs/embedded/playbooks/work-full.md` ≥ 1 | source: research N1, S3, S2
  - R3: test ratchet: never delete, skip or weaken a test to go green. | acceptance: one line in `work-full.md` step 4 and in the `check-review.md` standards axis | source: research S1
  - R4: baseline check: before the first task of a phase, run the phase check once and record any pre-existing failure as `blocked`. | acceptance: one clause in `work-full.md` step 1 | source: research S1
  - R5: WORKFLOW.md, the AGENTS block and the question-tool line in the 6 shims are trimmed, shim versions are bumped, and the tier table gains an effort column (effort, not prose, controls thinking). | acceptance: WORKFLOW ≤ 1,500 B, AGENTS block ≤ 700 B, each stage shim ≤ 1,000 B; `bash scripts/validate-skill.sh` passes on the changed shims | source: PR 130, research N1, owner scope answer
  - R7: rationale ledger + retention test. `docs/playbook-rationale.md` (not embedded, so agents never load it) has one row per kept playbook rule: rule, owning playbook/step, the failure it prevents or the guard that enforces it, and an expiry condition. During the slim, a rule with no row is deleted. | acceptance: the ledger exists; every `escalate_when`, `BLOCKED_*` and ratchet/baseline rule in the slim set has a row; no rationale sentence remains in the playbooks | source: research N3, N4
  - R6: the projected copies match. | acceptance: `cd cli && cargo test --test projection_parity` passes | source: AGENTS.md
- non-goals:
  - NG1: no merge of `check-validation.md` into `check-durable.md` and no installer retire-list (owner: keep the two files).
  - NG2: no new roles, stages, verifier layers, eval suites or JSON plan format (research N3: a second reviewer cost 8%, and more than 4 handoffs fail).
  - NG3: no change to the step numbers cited by ADRs or by other playbooks; no edit to the frozen ZGUARD-CORE.
  - NG4: `git`, `hunt`, `think` and `retro` bodies are out of scope (already trimmed by prompt-token-diet); the `🥷` line stays.

## Phases and Verification
- approach: write the R7 ledger for the check path first; it decides what is kept. Then spike the check gate path, because it is the largest file set and is read twice per cycle. If it lands on target with tests green, slim the other stage paths, then the shared docs and shims, then run a dry-run cycle. Rejected: one rewrite pass over all files, because a behavior regression could not be traced to one file.
- constraints:
  - Keep every `required:` phrase in `cli/src/embedded.rs` verbatim, unless a task names the phrase edit and updates the test in the same commit.
  - Every edit lands in both `cli/docs/embedded/` and `docs/` (byte-identical).
- recovery: each phase is one commit; revert that commit to restore the previous playbook set.
- risks: over-compressed guard text changes agent behavior → keep every rule in `check-validation.md` and only compress its prose; R4 costs one extra test run per phase → the phase check is already defined, so the cost is a single run.
- seams: `cli/src/embedded.rs` contract tests; `cli/tests/projection_parity.rs`; `scripts/verify-doc-links.sh`; `scripts/validate-skill.sh`.
- phase_slug: `p1-check-spike`
  status: planned
  - goal: R7 (check path), R1 (check path), R3 (check-review), R6 | depends_on: none
  - surfaces: `check.md`, `check-durable.md`, `check-validation.md`, `check-review.md` (+ the `docs/playbooks/` copies) | avoided: other playbooks, `cli/src/`
  - escalate_when: the check gate path cannot reach ≤ 6.5 KB without dropping a pinned phrase or a guard rule
  - wave 1:
    - T0 create `docs/playbook-rationale.md` with rows for every check-path rule; mark the rows that have no failure and no guard — output: ledger — check: `bash scripts/verify-doc-links.sh` — stop_if: a rule's purpose cannot be found in git log or the ADRs (ask the owner)
    - T1 rewrite the 4 check playbooks, dropping the marked rules — output: check gate path ≤ 6.5 KB — check: `wc -c cli/docs/embedded/playbooks/check*.md` — stop_if: a pinned phrase would need to change
    - T2 copy to `docs/playbooks/` — output: parity — check: `cd cli && cargo test --test projection_parity` — stop_if: drift
  - phase check: `cd cli && cargo test && cd .. && bash scripts/verify-doc-links.sh`
- phase_slug: `p2-stage-paths`
  status: planned
  - goal: R7, R1, R2, R3, R4, R6 | depends_on: p1-check-spike
  - surfaces: `brainstorm*.md`, `to-plan.md`, `work.md`, `work-full.md`, `handoff.md`, `watzup.md` (+ copies) | avoided: check playbooks, `cli/src/`
  - escalate_when: an owner-of-invariant choice would move a status transition between stages
  - wave 1:
    - T2b extend the ledger to the stage-path rules — check: `bash scripts/verify-doc-links.sh`
    - T3 slim the brainstorm ×3 playbooks; point the grill sub-agent line to the R2 brief — output: ≤ audit targets — check: `wc -c` + `cargo test` — stop_if: pinned phrase change
    - T4 slim to-plan, work and work-full; add the R2 brief, the R3 ratchet and the R4 baseline — output: ≤ audit targets; R2-R4 acceptance met — check: `cd cli && cargo test` — stop_if: step renumbering
  - wave 2:
    - T5 slim handoff and watzup (handoff owns the final transitions) — output: ≤ audit targets — check: `cd cli && cargo test` — stop_if: step 6 citation breaks
    - T6 copy to `docs/playbooks/` — check: `cd cli && cargo test --test projection_parity`
  - phase check: `cd cli && cargo test && cd .. && bash scripts/verify-doc-links.sh`
- phase_slug: `p3-shared-and-shims`
  status: planned
  - goal: R5, R6 | depends_on: p1-check-spike
  - surfaces: `cli/docs/embedded/WORKFLOW.md`, `cli/docs/embedded/AGENTS.md`, `docs/WORKFLOW.md`, the 6 stage `skills/workflow/*/SKILL.md` | avoided: playbooks
  - escalate_when: the question-tool line cannot shrink without losing a host name
  - wave 1:
    - T7 trim WORKFLOW and the AGENTS block; add the effort column (deep=high, standard=medium, fast=low) — check: `wc -c` + `cd cli && cargo test`
    - T8 shorten the shared question-tool line in 6 shims; bump versions — check: `bash scripts/validate-skill.sh skills/workflow/<name>` for each shim
  - phase check: `cd cli && cargo test && cd .. && bash scripts/verify-doc-links.sh`
- phase_slug: `p4-dry-run`
  status: planned
  - goal: success_signal | depends_on: p2-stage-paths, p3-shared-and-shims
  - surfaces: a scratch plan in a throwaway branch (never committed) | avoided: everything else
  - escalate_when: the dry run diverges from today's behavior at any stage
  - wave 1:
    - T9 run brainstorm lock → to-plan → work → check gate → handoff on a scratch plan with the slim set; byte-count the reads — output: every stage completes; reads ≤ 30 KB — check: the transcript plus `wc -c` of the files read
    - T10 run all 4 gate commands from AGENTS.md — check: all exit 0
  - phase check: `cd cli && cargo fmt --check && cargo clippy --all-targets -- -D warnings && cargo test && cd .. && bash scripts/test-guards.sh && bash scripts/verify-doc-links.sh`

## Log

## Validation

## Current State and Next Action
- active_phase: p1-check-spike
- lifecycle_status: planned
- blockers: none
- open_items: merge PR 130 (the audit) or cite it by link only
- exact_next_action: work p1-check-spike
