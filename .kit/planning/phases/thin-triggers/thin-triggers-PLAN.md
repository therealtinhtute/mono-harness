# Plan: thin-triggers

Phase: thin-triggers
Status: ready
Wave Count: 3
Execution Owner: work
Updated At: 2026-07-18

## Goal
6 spine SKILL.md files become ≤30-line thin triggers on MIN_ZHARNESS_VERSION 0.2.0; references pruned; repo docs updated; Claude-chain parity proven.

## Inputs
- Installed `zharness` 0.2.0 (cli-release done); `.kit/docs/**` scaffolded on this repo
- `cli/docs/embedded/**` ↔ references diff notes from playbook-authoring

## Wave 1
### T1 — MIN bump + shared trigger template
- type: docs
- inputs: released 0.2.0
- touches: `skills/workflow/README.md`
- avoid: skill bodies (next tasks)
- steps:
  1. Bump MIN_ZHARNESS_VERSION to 0.2.0 in README; update the 4-layer description: logic in embedded playbooks, skills as triggers
  2. Write the canonical thin-trigger template (from CONTEXT locked shape) into the README for contributors
- expected outputs: updated workflow README
- verification: inspection; grep confirms no other file still claims 0.1.0 as MIN
- stop if: —
- escalate to: check

### T2 — Rewrite watzup + handoff (safest pair)
- type: refactor
- inputs: T1 template, playbooks present in `.kit/docs/`
- touches: `skills/workflow/{watzup,handoff}/SKILL.md`, their references/
- avoid: git/interview, changing frontmatter name/description semantics
- steps:
  1. Rewrite each SKILL.md per template; delete references verifiably absorbed into the playbook (diff notes), keep any leftovers and file a finding
  2. Resync `~/.claude/skills/{watzup,handoff}`
  3. Live-verify each: run the skill on this repo; output honors the playbook contract (watzup: output contract; handoff: HANDOFF.md + `handoff record`)
- expected outputs: 2 thin triggers ≤30 lines, live-verified
- verification: `wc -l` ≤ 30 each; live run transcripts in the run artifact
- stop if: playbook insufficiency (agent needs deleted reference content)
- escalate to: finding vs playbook-authoring; pause skill

## Wave 2
### T3 — Rewrite check + work
- type: refactor
- inputs: T2 pattern proven
- touches: `skills/workflow/{check,work}/SKILL.md`, references/
- avoid: same as T2
- steps: same distill-delete-resync-live-verify loop; work verified by executing a trivial task in simple mode; check verified by gating that run
- expected outputs: 2 thin triggers, live-verified
- verification: `wc -l` ≤ 30; live run transcripts
- stop if: same as T2
- escalate to: same as T2

### T4 — Rewrite to-plan + brainstorm
- type: refactor
- inputs: T3 done
- touches: `skills/workflow/{to-plan,brainstorm}/SKILL.md`, references/
- avoid: same as T2
- steps: same loop; live-verify with a throwaway scratch spec (explore mode for brainstorm; phase-refresh for to-plan) — do NOT disturb this initiative's live SPEC/ROADMAP
- expected outputs: 2 thin triggers, live-verified
- verification: `wc -l` ≤ 30; transcripts
- stop if: same as T2
- escalate to: same as T2

## Wave 3
### T5 — Repo docs sweep
- type: docs
- inputs: all 6 rewritten
- touches: root `README.md`, `CLAUDE.md`, `docs/workflow-harness/migration.md`
- avoid: cli/docs (frozen)
- steps:
  1. Update the architecture narrative everywhere: playbooks embedded, `.kit/docs/`, thin triggers, init scaffolds
  2. migration.md: new-adopter path is now just `install → zharness init` (scaffold included), legacy path unchanged
- expected outputs: consistent repo docs
- verification: grep for stale claims ("references/", "logic lives in SKILL.md", MIN 0.1.0)
- stop if: —
- escalate to: check

### T6 — Claude-chain parity run (R10)
- type: test
- inputs: full rewritten chain
- touches: `.kit/` artifacts only (a real mini-cycle)
- avoid: faking artifacts
- steps:
  1. Run a small real task through the full chain on this repo: brainstorm (lock a tiny follow-up) → to-plan → work → check → git → handoff, all via the thin triggers
  2. `zharness validate --json` + `audit --json`: zero pointer_drift on the produced chain
- expected outputs: R10 acceptance evidence in the run artifact
- verification: audit/validate JSON captured; artifact chain complete (SPEC→…→HANDOFF)
- stop if: any stage needs content the playbook lacks
- escalate to: check (findings decide: playbook fix next cycle vs blocker)

## Risks / Watch-fors
- Daily-driver risk is here: never leave a skill half-rewritten across a session boundary — each of T2–T4 lands atomically per skill
- Installed-copy drift: resync `~/.claude/skills` immediately per skill, or the live verify tests the old copy
