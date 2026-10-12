# Agent SDLC practice (to Oct 2026) vs the zharness workflow

Read-only research, 2026-10-12, weighted to sources from Aug-Oct 2026 and the
current model generation (Claude Opus 5.5, released 2026-09-22; Claude 5 family;
GPT-5.6). It answers two questions: what current agent harnesses do that this
workflow lacks, and what it carries that current models no longer need. It feeds
`docs/plans/completed/playbook-slim.md` and builds on the byte audit in
https://github.com/therealtinhtute/mono-harness/pull/130.

**Verdict:** the workflow already has the shape the field converged on in 2026.
Durable state lives in files (plan + `## Log`), work is one unit at a time, a
single independent judge sits apart from the generator, and rules are enforced
by executable guards rather than prose. What has changed is the cost side.
Current models need fewer and shorter instructions, and every rule left in
prose costs reasoning tokens. So the slim is the right move, with one new
discipline taken from the Aug-Sep 2026 sources: **a rule stays only if a guard
enforces it or a recorded failure justifies it, and its "why" lives outside
the loaded text.** Three one-line multi-agent rules still fill real gaps.

## Sources

Q3-2026 sources come first; the older ones are kept only where nothing newer replaced them.

| # | Source (date) | Takeaway used here |
|---|---|---|
| N1 | Anthropic, *Prompting Claude Opus 5.5* (platform docs, Sep 2026) | At default `medium` effort it matches Opus 5 at `high` with fewer tokens. **Lower effort reduces thinking more reliably than prompt instructions do.** Remove instructions that stood in for thinking. For unattended runs: keep the task's parts in a checklist file, treat a text-only end of turn as a report rather than proof of done, and stop after 2-3 automatic continuations. For multi-agent: give subagents an elapsed-time budget. |
| N2 | Anthropic, *Introducing dynamic workflows in Claude Code* (GA, Jul-Oct 2026) | Fan-out to many parallel subagents; findings are verified independently before they are folded in; progress is saved so an interrupted run resumes. Usage costs much more, so start on a scoped task. |
| N3 | Marmelab, *The State of AI Harness Engineering 2026* (2026-09-24): 246 repos, 57 publications | Same model, different harness: 68%→88%. Keep the harness to the strict minimum and add rules only in reaction to an observed failure. Prose rules are rarely enforced (4-16%), so back them with scripts and hooks. Test guards both ways. Give every control an expiry condition and a recorded reason. **A second reviewer agent lowered success by 8%; a 4-role team beat one agent by only 0.4 pt; past 4 handoffs work "almost always failed".** |
| N4 | *Why Does CLAUDE.md Keep Growing?* arXiv 2608.11095 (Aug 2026) | 247,694 instruction lifetimes: agent prompts grow +226% over their life because a rule's rationale is lost and nobody dares delete it. Recording the *why* next to each rule halts the growth (+211%→+1.4%) and improves instruction-following by up to 23%. |
| N5 | ETH Zurich, *Evaluating AGENTS.md* (Feb 2026, via N3) | Human-written context files help about 4%; machine-generated ones hurt. Either way they cost 14-22% more reasoning tokens. |
| S1 | Anthropic, *Effective harnesses for long-running agents* (Nov 2025) | Progress file + git as the bridge between sessions; one feature at a time; "it is unacceptable to remove or edit tests"; run a basic test before new work. |
| S2 | Addy Osmani, *Long-running Agents* (Apr 2026) | Google, Anthropic and Cursor converged on planner / worker / judge, a done-condition written first, and a worktree per parallel agent. |
| S3 | *Natural-Language Agent Harnesses*, arXiv 2603.25723 | A harness = contracts + roles + state + failure taxonomy. File-backed state helps most; a verifier and multi-candidate search add cost, not solve rate. |
| S4 | mgechev/skills-best-practices; agentskills.io spec | SKILL.md routes; numbered imperative steps; one term per concept; "if the agent already does it reliably, delete the instruction". |

## Pattern map

| Pattern (source) | zharness today | Gap | Action |
|---|---|---|---|
| State in files, not context (N1, N2, S1) | `## Log` + `## Current State` + one active plan | none | keep |
| Done-condition written first (N1, S2) | `## Goal` with `acceptance:`; to-plan traceability | none | keep |
| A text-only end of turn is not done (N1) | proof re-execution guard; `check` gate | none | keep |
| Cap on automatic retries (N1: 2-3) | `escalate_when`: same command failed twice | none; it already matches | keep |
| One independent judge (N3, S2) | judge only on high-risk and `full`; `check-review` sub-agent is opt-in | none | keep; **do not add reviewer layers** (N3: −8%) |
| Few handoffs (N3: >4 fails) | brainstorm → to-plan → work → check → handoff | at the limit | no new stage or role; `tiny`/`bounded` lanes skip stages |
| Rules backed by scripts (N3) | ZGUARD-CORE hook; `test-guards.sh` covers both pass and reject cases | none | keep; a prose rule a guard already enforces shrinks to a pointer |
| Each rule's why recorded outside the loaded text (N3, N4) | rationale sits inline in the playbooks; ADRs cover only some rules | **gap**: the slim cannot tell which rules are safe to delete | **R7**: rationale ledger + retention test |
| Less prompting for current models (N1, N5, S4) | spines carry rationale, Exit Conditions and restated invariants | **gap**: ~45% of the bytes | **R1** (slim per PR 130) |
| Effort as the control, not prose (N1) | WORKFLOW maps tiers to model names only | small | **R5**: add an effort column to the tier table |
| Delegation brief: inputs, outputs, budget, stop (N1, S3) | ~550 B of prose at `work-full.md` step 4 | no fixed shape; no time budget | **R2**: 5-field brief incl. budget |
| Parallel workers isolated (S2, N2) | "parallelize only independent tasks" | no isolation rule | part of **R2**: a worktree or branch per parallel writer |
| Test ratchet (S1) | absent | an agent can weaken a test to get green | **R3** |
| Baseline check before new work (S1) | absent | a phase can start on a red tree | **R4** |
| Usage telemetry / evals (N3) | none | out of scope | NG: a dry-run cycle is the check this round |

## What not to add

- **More reviewer or verifier agents, or more stages.** N3 and S3 both measure no gain or a loss. Dynamic
  workflows (N2) already provide fan-out plus verification when a user asks for it, so the playbooks
  should not reinvent that.
- **A JSON feature list.** The guard already protects the parsed plan fields, and ADR 0001/0006 keeps markdown.
- **"Think carefully" style instructions.** N1: effort is the control, so such lines only add latency.
- **A shared invariants kernel in WORKFLOW.md.** It would be a mandatory +2 KB per stage (audit).

## Workflow skills outside the playbooks

`skills/workflow/*/SKILL.md` total 30.5 KB. Each stage shim repeats a 302 B question-tool line. `git`, `hunt`,
`think` and `retro` were trimmed by the closed prompt-token-diet plan (`git show 385cf56`) and stay out of scope.

## Risks

- Contract tests pin phrases, not behavior. The only behavior check this round is a dry-run cycle.
- Several N-sources are secondary (N3 aggregates others; N4 and S3 are preprints). R1 and R7 rest on N1
  (primary, current model) and N4's measured effect; R2-R4 rest on S1 and N1.
- Guidance tied to a specific model rots (N3: Anthropic removed context resets once Opus 4.5 no longer needed
  them). The R7 ledger's expiry column exists for exactly this.
