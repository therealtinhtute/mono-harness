# Agent SDLC practice (2025-26) vs the zharness workflow

Read-only research, 2026-10-12. It answers one question: what do current agent
harnesses and multi-agent coding setups do that this workflow lacks, and what
does it carry that they show is not needed? Input to
`docs/plans/active/playbook-slim.md`. It builds on the byte audit in
https://github.com/therealtinhtute/mono-harness/pull/130.

**Verdict:** the workflow already matches the shape the field converged on: an
external plan file, an append-only progress log, one task at a time,
generation kept apart from evaluation, and proof re-execution. The gap is
cost, not shape. Its playbooks restate their invariants about 6 times, and
three small multi-agent rules are missing: a fixed delegation brief, a test
ratchet, and a baseline check before new work. Add those three as single
lines, and get the bytes back by doing the slim from the audit.

## Sources

| # | Source | Takeaway used here |
|---|---|---|
| S1 | Anthropic, *Effective harnesses for long-running agents* (2025-11-26) | initializer + coding agent; feature list with `passes:false`; progress file + git log as the bridge between sessions; one feature per session; "it is unacceptable to remove or edit tests"; run a basic end-to-end test before starting a new feature |
| S2 | Addy Osmani, *Long-running Agents* (2026-04-28) | Google, Anthropic and Cursor converged on: plan file + progress file + structured handoff; planner / worker / judge split; "write the done-condition down before the agent starts"; treat context resets as first-class |
| S3 | *Natural-Language Agent Harnesses*, arXiv 2603.25723 | a harness = contracts + roles + state semantics + a failure taxonomy; delegated child agents take ~90% of the tokens; file-backed state helps most; **a verifier, multi-candidate search and dynamic orchestration add cost, not solve rate** |
| S4 | mgechev/skills-best-practices; Agent Skills spec (agentskills.io) | SKILL.md is for routing; numbered imperative steps; one term per concept; references one level deep and loaded just in time; "if the agent already does it reliably, delete the instruction" |
| S5 | Anthropic, *Effective context engineering for AI agents* (2025-09-29) | context is a finite budget; give sub-agents pointers, not copies |

## Pattern map

| Pattern (source) | zharness today | Gap | Action |
|---|---|---|---|
| External plan with a done-condition (S1, S2) | `## Goal` with `acceptance:` per requirement; to-plan step 8 traceability | none | keep |
| Progress file + git bridge (S1) | `## Log` (append-only, sole status source) + `## Current State` | none, but it is restated in 6 playbooks | one owner per invariant (slim R1) |
| One unit per session (S1) | one active plan; one phase per `work` run | none | keep |
| Generator / evaluator split (S2, S3) | `check full` and high-risk gates need an independent judge; proof re-execution in the hook | none | keep; S3 says not to add more verifier layers |
| Test ratchet (S1) | absent | an agent can delete or weaken a failing test to go green | R3: one line in `work-full.md` and `check-review.md` |
| Baseline check before new work (S1) | absent; red→green covers only the new test | a phase can start on an already-broken tree | R4: one line in `work-full.md` step 1 |
| Delegation contract: inputs, outputs, budget, stop rule (S3, S5) | ~550 B of prose at `work-full.md` step 4; the shape is restated ad hoc in `check-review.md`, `brainstorm-grill.md`, `think` | no fixed shape, so briefs drift and grow | R2: one 5-field brief, used in every place that delegates |
| Parallel workers isolated (S2: worktree per agent) | "parallelize only tasks the plan marks independent" | no isolation rule, so two writers can share a tree | part of R2: parallel writers each get their own worktree or branch |
| Failure taxonomy (S3) | `BLOCKED_*` codes, all in `work-full.md` | none; already one owner | keep; do not lift it into WORKFLOW (the audit rejected a mandatory kernel) |
| Context reset / compaction (S2) | re-read after compaction in 3 playbooks; handoff compacts the plan | restated 3× | one owner (slim) |
| Progressive disclosure (S4) | spine + companion playbooks; shims route to playbooks | spines carry rationale and Exit Conditions | slim R1 |
| Skill evals (S4) | `validate-skill.sh` (structure only) | no behavior eval | out of scope (the prompt-token-diet plan already set NG2: no eval suites); a dry-run cycle is the check |
| Initializer agent / `init.sh` (S1) | `brainstorm lock` + `docs/PROJECT.md` gate commands | none | keep |

## What the research says not to add

- **More roles or orchestration layers.** S3's ablations: a verifier, multi-candidate search and dynamic orchestration
  shift *which* tasks pass but do not raise the solve rate, and they add a lot of process cost.
  The existing single independent judge is enough.
- **A JSON feature list.** S1 uses JSON because models overwrite it less often than markdown. Here,
  the pre-commit guard already protects the plan sections it parses, and ADR 0001/0006 keeps markdown
  as the source of truth. Switching formats would cost more than the risk it removes.
- **A shared invariants kernel in WORKFLOW.md.** It would make WORKFLOW a mandatory read on every stage
  (the audit measured +~2 KB per stage).

## Workflow skills outside the playbooks

`skills/workflow/*/SKILL.md` total 30.5 KB. The six stage shims are ~1.2 KB each, and 302 B of that is the
same question-tool line (×6). `git`, `hunt` and `think` (6.5-7.9 KB) were already trimmed by the
prompt-token-diet plan (closed 2026-09-29, `git show 385cf56`). Only the shared question-tool line
and the `🥷` line pattern are left to cut, and NG1 of that plan keeps the `🥷` line.

## Risks

- Contract tests pin phrases, not behavior. A dry-run cycle on a scratch plan is the only behavior check
  (also flagged by the audit).
- Shim edits touch `version:` fields and run the skill validator.
- Sources S2 and S3 are secondary or preprint. The three additions rest mainly on S1, a primary
  source from the team that ships the host agent.
