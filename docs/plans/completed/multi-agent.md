---
status: completed
lane: normal
---
# Multi-agent fan-out: swarm, arena and orchestrate-lite

## Goal
- outcome: Agents in this repo run several sub-agents through one written procedure adapted from pstack: a `fanout` mono-mode playbook (swarm partition/race plus an orchestrate-lite section), a separate `arena` skill, and a shared `work-fanout.md` core that `work full` step 4 also loads.
- success_signal: `cd cli && cargo test` passes with contracts pinning the new files; `bash scripts/validate-skill.sh` passes for `skills/router/mono-mode` and `skills/reason/arena`; `bash scripts/verify-doc-links.sh` reports 0 findings.
- actors: the owner (ní); the main agent that fans out; the sub-agents it spawns.
- authority: owner decisions in brainstorm on 2026-10-23 — adopt pstack's swarm + separation, arena and orchestrate-lite; model defaults as a table in skill text; a new mono-mode playbook as the entry point; arena as a separate skill like pstack; orchestrate state in the plan Log only; a different-family verifier recommended, not required; own branch off master. Reference: github.com/michael-denyer/pstack-claude (`swarm`, `arena`, `poteto-mode` Subagents, `playbooks/orchestrate.md`, `principle-separate-before-serializing-shared-state`).
- requirements:
  - R1: `skills/stages/work/references/work-fanout.md` holds the shared core: shared-file separation before spawning (one owner per file; tooling moves land first and the hook is reinstalled), a role → model table where the owner's named model wins, the GOAL/SCOPE/CONTEXT/ACCEPTANCE/VERIFY/TIMEBOX/FORBIDDEN/REPORT brief, aggregate rules (reject a report without tested source and output; one fresh retry; then `blocked`, never a pass), integrate rules (merge one at a time, held shared files edited after, full check on the combined tree), and a different-family verifier recommended; `work-full.md` step 4 loads it. | acceptance: an `embedded.rs` PlaybookContract pins those lines and `skill_reference_resolution` copies the file; `cd cli && cargo test` passes | source: owner decision (swarm + separation, model table)
  - R2: `skills/router/mono-mode/playbooks/fanout.md` exists with numbered steps (frame → separate → spawn → aggregate → integrate → check), each naming its skill or `work-fanout.md`; the mono-mode SKILL.md table routes parallel/sub-agent requests to it and still names the other five playbooks. | acceptance: `bash scripts/validate-skill.sh skills/router/mono-mode/SKILL.md` passes; `grep -q 'playbooks/fanout.md' skills/router/mono-mode/SKILL.md` | source: owner decision (new mono-mode playbook)
  - R3: `fanout.md` has an orchestrate-lite section: standing orders kept in a plan section and pasted into every spawn and retry, a rolling window of at most 10 in-flight agents, landing as each unit verifies, liveness checked from branches and `git status` (never by resuming an agent), and all state as plan `## Log` entries naming model, branch and SHA, with no store outside the plan. | acceptance: `grep -niE 'standing orders|rolling window|## Log' skills/router/mono-mode/playbooks/fanout.md` shows each; `grep -c orch skills/router/mono-mode/playbooks/fanout.md` finds no store path | source: owner decision (orchestrate-lite, plan Log only)
  - R4: `skills/reason/arena/SKILL.md` is a skills.sh skill: frame (artifact, 3–6 point rubric, runners and models, one output location each) → fan out (one message, background) → cross-judge (a different family recommended) → pick a base by rubric → graft by hand → verify; divergent candidates trigger a reframe, not an average; it briefs runners per `work-fanout.md`. | acceptance: `bash scripts/validate-skill.sh skills/reason/arena/SKILL.md` passes; README's skill list names `arena` | source: owner decision (arena as a separate skill)
- non-goals:
  - NG1: an `orch` CLI, store, or any control-plane state outside the plan.
  - NG2: a model config file or zharness config for models.
  - NG3: a mandatory cross-family judge; hooks; installing pstack itself.
  - NG4: changes to the repo-restructure work on `refactor/repo-restructure`.

## Phases and Verification
- approach: pstack's procedures adapted as markdown in our stage/router/reason layout, sharing one core reference; rejected installing pstack (its routing hook competes with mono-mode) and its `orch` store (AGENTS.md forbids parallel control-plane state).
- phase_slug: `main`
  status: done
  - goal: R1, R2, R3, R4 | depends_on: none
  - surfaces: `skills/stages/work/references/work-fanout.md`, `work-full.md`, `skills/router/mono-mode/**`, `skills/reason/arena/**`, `cli/src/embedded.rs`, `cli/tests/skill_reference_resolution.rs`, `README.md` | avoided: `cli/src/installer/**`, hooks, `docs/WORKFLOW.md`
  - wave 1:
    - T1 Finish `work-fanout.md` (draft already present) and the step 4 pointer in `work-full.md`; contract + resolution entries (owner: main agent) — output: R1 — check: `cd cli && cargo fmt --check && cargo test`
  - wave 2:
    - T2 Write `playbooks/fanout.md` with the swarm steps and the orchestrate-lite section; add its row to mono-mode SKILL.md (owner: main agent) — output: R2, R3 — check: `bash scripts/validate-skill.sh skills/router/mono-mode/SKILL.md` and the R3 greps
    - T3 Write `skills/reason/arena/SKILL.md`; add `arena` to README's skill list (owner: main agent) — output: R4 — check: `bash scripts/validate-skill.sh skills/reason/arena/SKILL.md`
  - phase check: `cd cli && cargo fmt --check && cargo clippy --all-targets -- -D warnings && cargo test && cd .. && bash scripts/validate-skill.sh skills/router/mono-mode/SKILL.md && bash scripts/validate-skill.sh skills/reason/arena/SKILL.md && bash scripts/verify-doc-links.sh`

## Log
- 2026-10-23T03:04Z — main — decision — T2/T3 run inline by the main agent, not as a 2-agent fanout: two small files, spawn cost exceeds the work
- 2026-10-23T03:04Z — main — decision — corrected work-stage state: phase and lifecycle set back to in-progress (final phase awaits independent check full; only closing handoff marks done); placeholder T00:00Z timestamps replaced with the time of correction
- 2026-10-23T03:13Z — main — decision — check full requests 1-2 fixed: contract pins 7 more work-fanout.md lines (tooling first, owner model wins, GOAL/FORBIDDEN/REPORT fields, second miss → blocked, merge one at a time), proven by breaking one line (cargo test fails) and restoring; R3 grep now -niE
- 2026-10-23T03:13Z — main — decision — absorb: none

## Validation
- 2026-10-23T03:05Z — phase `main` — verdict: APPROVE_WITH_REQUESTS — mode: full
  - `cd cli && cargo fmt --check && cargo clippy --all-targets -- -D warnings && cargo test` — exit 0; tests 80/3/2 pass, 0 failed
  - `bash scripts/validate-skill.sh skills/router/mono-mode/SKILL.md` — exit 0; PASS WITH WARNINGS (8, baseline)
  - `bash scripts/validate-skill.sh skills/reason/arena/SKILL.md` — exit 0; PASS WITH WARNINGS (8, baseline)
  - `bash scripts/verify-doc-links.sh` — exit 0; doc links OK (0 findings)
  - `grep -q playbooks/fanout.md skills/router/mono-mode/SKILL.md` — exit 0; table has 6 playbook rows
  - `grep -nE 'standing orders|rolling window|## Log' skills/router/mono-mode/playbooks/fanout.md` — exit 0; matches only line 22 (## Log), see request 2
  - scope: on target — diff stays in the phase surfaces; avoided surfaces untouched; diff clean of the repeated BAD_OUTPUT class
  - requirements: R1 met (work-fanout.md holds every listed rule; work-full.md:19 loads it; contract + resolution entry; cargo test pass) | R2 met (fanout.md steps 1-6 name skills; SKILL.md:20 row; validate-skill pass) | R3 met (fanout.md:18-22 by content; no orch store) | R4 met (arena SKILL.md frame→verify, reframe rule, briefs per work-fanout.md; README.md:173) | success_signal met
  - rollback_point: 7a1880e
  - requests: 1) cli/src/embedded.rs:346-358 pins 7 lines but not R1's brief fields, retry-then-blocked, owner-model-wins, tooling-lands-first or merge-one-at-a-time (work-fanout.md:17,31,37-46,58,63); add markers if R1 means all of them. 2) R3 acceptance grep is case-sensitive and misses "Standing orders"/"Rolling window" (fanout.md:18-19); use -i or match the text.
  - judge: independent
  - judge_model: claude-opus-5-5

## Current State and Next Action
- active_phase: none
- lifecycle_status: completed
- blockers: none
- open_items: none
- exact_next_action: owner decides commit, PR and release of feat/multi-agent
