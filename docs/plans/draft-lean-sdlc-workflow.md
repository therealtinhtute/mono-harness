---
status: draft
lane: high-risk
---
# Lean SDLC workflow with measured model autonomy

## Goal

- outcome: reduce unnecessary instruction loading, routing ceremony and pauses while preserving scope, recoverable state, independent review and observable proof. Let capable models choose implementation methods inside explicit authority; use inexpensive agents for narrow research. Keep the harness portable, markdown-first and independent of a lifecycle runtime.
- success_signal: the candidate passes repository gates and the behavioral comparison defined below with no observed correctness/authority/proof regression; it improves at least one predeclared workflow-cost metric on the tested corpus. Report exact model/runtime coverage and uncertainty. If evidence is equal, inconclusive or unavailable, do not claim optimization or complete the performance requirement.
- actors: owner, lead coding/review agents, inexpensive research agents, contributors on different skill-capable hosts, and recipients of completed tasks/PRs.
- authority: owner's 2026-10-13 request for a lightweight, more reliable SDLC that benefits from strong current models, followed by the request to prepare a complete migration and future implementation plan. This document specifies the follow-up for later authorized execution; it does not start it now.
- prerequisite: finish and archive `playbooks-to-skills`; installed skill procedures and portable routing are the migration's responsibility. Do not begin this initiative while any non-empty active plan exists.
- requirements:
  - R1: preserve all durable and authority invariants listed below, including ADR 0014. | acceptance: invariant matrix maps each rule to current authority and executable/observable proof; guards and existing contracts pass | source: owner goal and repository ADRs
  - R2: classify existing tiny/normal/high-risk lanes from uncertainty, impact, dependency and recoverability, then resolve work bounded/full from the actual request's durable intent; never use arbitrary file/line counts. | acceptance: large mechanical known change stays bounded; a small risky/ambiguous change takes the appropriate durable/clarification path; active-plan existence alone never selects full | source: reviewed simplification
  - R3: make skills selective routers with precise descriptions and contextual reads; remove duplicated or unsupported general prompting claims. | acceptance: installed task traces load only relevant stage/companions; claim inventory resolves numeric/model claims; local references and skill validators pass | source: owner and primary-source guidance
  - R4: permit autonomous execution and targeted verification inside approved scope, with meaningful stop boundaries. | acceptance: coding/resume cases complete without redundant permission loops or test repetition; behavior changes still have credible proof and high-risk/full review remains independent | source: reviewed simplification
  - R5: express model use as host-level capability guidance rather than mandatory vendor/model/stage locks. | acceptance: no forced named-model metadata on six core skills; documented roles and transfer brief; lead retains decisions; candidate promotion requires two recorded strong-model configurations evaluated | source: owner and Luna research preference
  - R6: strengthen task readiness, integration and handover using current plan/Git surfaces. | acceptance: normal remains taskless, high-risk tasks remain traceable, combined-result verification and stale-evidence rejection work, delivery distinguishes implementation/PR/merge/deploy states | source: Matt/Paseo synthesis and existing authority
  - R7: compare A/B/C with isolated fixtures and independently authored fresh holdouts, using retained evidence and bounded trial counts. | acceptance: complete run matrix, independent scoring, measured targeted gain with no observed predicate regressions; unavailable data remains unavailable | source: existing manual eval protocol
  - R8: publish coherent repository guidance and a model-upgrade reevaluation recipe; preserve rollback and one source of state. | acceptance: live docs/site agree, model changes require fresh baseline, rejected candidates can be reverted to B without undoing migration | source: owner requirement for a self-contained handoff
- non-goals:
  - NG1: no new CLI command, runtime service, task database, derived index, heartbeat loop, agent registry, mandatory MCP or external issue tracker.
  - NG2: no new lane or plan/guard schema; no mandatory task list for normal; no blanket removal of tests or independent judges.
  - NG3: no implicit commit/push/PR/merge/deploy permission; no expanded production access. Delivery evidence is distinct from authority to publish.
  - NG4: no universal model ranking, unsupported token-saving percentage, fixed instruction budget or claim that instructions deterministically enforce agent behavior.
  - NG5: no automatic fleet-wide change to personal/global skills, host settings or real consumer repositories; no full Paseo integration or eval platform.

### What the next agent must know

This draft is a complete dependent specification, not a second active initiative. Keep five top-level plan sections. After migration closes, move this file rather than copy it, change frontmatter to active, set active_phase to f1-contract, and record the actual B SHA in a decision Log entry. Check current source against the facts below; do not assume the earlier master checkout is current.

Prerequisite record: `docs/plans/completed/playbooks-to-skills.md`, with final clean independent Validation, closed phases and the source commit named in its completion evidence. If later archival pruning removes the file, recover it through the recorded completion commit with git show; absence without a verifiable completion record blocks promotion. Do not infer completion from PR state alone.

| Source | Relevant authority/fact |
|---|---|
| PR #133 head `156bc96b15715ddd8b43e69bbeaa4bcd0fff8c1a` | Pre-migration source baseline A; includes the shallow-clone fixture fix |
| `docs/decisions/0014-outcome-spec-normal-plans.md` | Normal = Goal + phase main; lock writes phase; to-plan high-risk only; single-phase work skips duplicate gate and goes to independent check full |
| `docs/decisions/0012-five-section-plan-and-update-migration.md` | Plan shape and durable ownership |
| `scripts/install-git-hooks.sh`, `scripts/test-guards.sh` | Executable guard boundaries, not proof of actual model compliance |
| `cli/src/embedded.rs` | Current contract assertions; after migration, their procedure inputs live in skills |
| `docs/evals/README.md` | Manual independent comparison, three fresh trials per side/case, evidence retention, no cherry-picking |
| `docs/prompt-engineering-principles.md` | Existing principles include unreferenced numeric and model-family generalizations; review as claims, not established laws |
| `skills/workflow/git/` | Git procedure owns outward delivery actions; it owns no plan state |

External evidence is design input, not repository authority:

- [Matt task slicing](https://github.com/mattpocock/skills/blob/49dd158d1076134a641b33efb035946536778336/skills/engineering/to-tickets/SKILL.md): vertical, independently verifiable slices with dependency edges.
- [Matt implement-spec](https://github.com/mattpocock/skills/blob/49dd158d1076134a641b33efb035946536778336/skills/engineering/implement-spec/SKILL.md): already has parallel ready work, separate worktrees, merger and branch review. Do not repeat Paseo README's outdated contrast that Matt is only sequential.
- [Matt PR evidence](https://github.com/mattpocock/skills/blob/49dd158d1076134a641b33efb035946536778336/skills/engineering/pr/SKILL.md): compact before/after and impact evidence.
- [Paseo merge/review](https://github.com/hanh9898/matt-with-paseo/blob/18ed60b34215acd44bba3db16c5d50d40e7e94f5/plugins/matt-with-paseo/skills/matt-with-paseo/SKILL.md#L360-L406): compare reports to artifacts and verify combined results/seams. Borrow those invariants, not the control plane.
- [OpenAI skill/prompt guidance](https://developers.openai.com/blog/rethinking-skills-and-prompts-for-gpt-6-astra): concise descriptions, selective loading, fewer prescribed recipes, context-sensitive boundaries. Recheck official guidance when implementing; this is not a benchmark of this repository.

### Invariants to preserve

| Invariant | Required treatment |
|---|---|
| Read-only intent changes nothing | Preserve; explicit request to write an artifact authorizes that artifact only |
| One active plan; five sections; stable parsed fields | Preserve exact parser-facing grammar and append-only history |
| Normal flow already simplified | Preserve Goal + main phase, no normal task list/to-plan, single final independent check |
| High-risk dependency plan | Keep phases/waves/tasks and requirement-to-check mapping; autonomy does not invent product rules or remove acceptance |
| Proof and independent review | Preserve guard-required command/verdict layout and high-risk/full judge requirement; independence declaration remains testimony |
| Authorization | Preserve owner boundaries, especially commits/push/PR/merge/deploy; earlier explicit permission persists only within its scope |
| Recovery | Log is execution history, Current State identifies next action, Git/artifacts establish current reality; do not add another status store |
| Skill resolution | Preserve migration's installed-path rules and unsupported-host boundary |

### Exact candidate behavior

1. **Routing.** Delete the numeric five-file/~100-line rejection from work. Bounded requires clear requested outcome, known verification and a reversible local change without unresolved cross-subsystem decisions or durable coordination needs. Normal is one reviewable durable change. High-risk remains the existing migration/security/public-contract/dependency lane. File count is information, never sole routing authority. Unclear product behavior still needs clarification before dependent edits.
2. **Selective context.** Core SKILL.md files contain applicability, mode routing, outcome and reference pointers. Keep existing thin wrappers; do not merge all procedures into one always-read file. Remove duplicate prose only after mapping its invariant to the surviving authority. References are read for the active branch of the workflow. No new numeric line/token cap is an acceptance substitute.
3. **Execution latitude.** Preserve approved outcomes, surfaces, dependencies and checks. Within them the agent chooses inspection sequence, implementation mechanics and independent read-only research. High-risk write parallelism remains explicitly dependency/ownership-controlled and isolated. Normal work does not gain a task graph merely to use research agents.
4. **Verification scope.** Behavior changes need a test/repro that can fail for the bug or missing behavior, then pass on the fix, or an explicitly justified observable equivalent. Docs/mechanical changes use appropriate existing checks. Run required repository/phase checks; after a pass rerun only after relevant changes, failed integration or unresolved evidence. Do not add tests that merely mirror low-impact implementation. Never weaken tests to get green.
5. **Stopping and clarification.** Ask when a product choice, changed acceptance, destructive operation or authority boundary genuinely depends on the owner. Continue already-authorized independent work. Replace arbitrary second-failure stopping with a no-progress rule: repeating the same failing verification after one targeted fix without new evidence blocks; materially new evidence may support a new scoped hypothesis, recorded in Log for durable work. Never retry the same failed action indefinitely. Infrastructure/access failure is reported separately from product failure.
6. **Models.** Remove mandatory model names from six core workflow skill frontmatter and the absolute mid-phase model-switch ban. Document capability roles instead: strong lead for decisions/difficult implementation/final synthesis; inexpensive agents such as Luna for scoped exploration; capable independent judge for required reviews. Use the caller's host configuration; do not build a model resolver. A transfer carries goal/scope, evidence pointers, source identity and next action. Cache cost is a measurement, not a reason to prohibit escalation. Recheck any actual host-required metadata before removal and record compatibility evidence.
7. **Task readiness.** For high-risk tasks, put observable output, acceptance check, dependency and ownership in existing task fields/surfaces. A slice should be a thin end-to-end outcome. Owner assignment lives in task prose or an append-only decision Log entry, not a new parsed field. For normal plans use requirements and their acceptance directly; no added ticket file/tracker.
8. **Integration.** A worker report is a claim. Verify its source identity/diff and acceptance. Record tested commit SHA plus a hash of tracked dirty diff and relevant untracked inputs as ordinary Log/evidence prose. Compare that snapshot with the current combined tree before reusing proof; on a relevant change, invalidate and rerun affected acceptance checks. After combining changes, verify the combined tree and review shared interfaces; tests on separate branches do not prove that result. Record in existing Log/Validation, no new parsed fields or custom freshness engine.
9. **Resume.** Reconcile plan with actual branch/HEAD, dirty tree, pending worker changes and evidence. A valid logged done item is not reimplemented; a changed artifact can require revalidation. A conflicting state blocks with a concrete explanation instead of inventing completion or silently overwriting user edits.
10. **Delivery.** Final summary/PR evidence identifies outcome, artifact/branch or PR if one exists, acceptance results, limitations, and next receiving action. Distinguish implemented, locally verified, committed, PR-created, merged and deployed. A handoff preserves session continuity; it does not imply business acceptance or deployment. Keep git actions explicitly scoped and authorized.
11. **Prompt claims.** Replace unsupported universal instruction budgets, exponential/linear degradation, fixed Vietnamese token penalty and universal negative-prompt causality with sourced, qualified observations or remove them. Retain precise positive guidance and test/remove iteration. Model-specific guidance is dated and linked; latest is not hardcoded product policy.

These choices preserve the product boundaries; they revise prompt-level method constraints. Record a new ADR using the next free decision number (expected 0016 after migration's 0015), explicitly preserving ADR 0014. If a supposedly prompt-only edit changes guard grammar or an ADR decision, stop and amend authority before implementation rather than dropping the associated assertion.

## Phases and Verification

- approach: freeze authority/invariants and a development evaluation contract, implement a lean candidate, evaluate controlled comparisons, then update live guidance and hand over. Each phase leaves an usable repository; candidate prompts are not deployed to real consumers before acceptance.
- alternatives rejected: import both external workflows wholesale (duplicate state/runtime); remove all process and rely on model intelligence (loses recovery/proof); add more prompts without measuring (cannot isolate improvement); mix migration with tuning (confounds attribution and rollback).
- seams: migrated stage references and wrappers, current Rust contract tests, existing guard fixtures, doc links, isolated installed-skill fixtures and observable agent traces. No new provider API dependency; use existing host interfaces manually unless a tiny adapter is needed for evidence capture.
- constraints: preserve all invariants; scope references below resolve after migration; no global host changes; evaluator independence and hidden-case separation; no weakening scoring after seeing candidate results.
- recovery: pin accepted migration source as B before tuning. Revert candidate semantic changes to B if quality/authority regresses; retain failed trials and decision records. Reverting candidate must not undo safe retirement. A promoted plan is closed only after final evidence, not because the trial budget is exhausted.

### Artifact and evidence design

- Use `docs/evals/lean-workflow.md` for the public development protocol and cases; `docs/evals/runs.md` for indexed trial records; `docs/evals/evidence/<run-id>/` for redacted retained artifacts. Reuse existing conventions rather than another database or status ledger.
- A small Bash adapter/scorer may be added under scripts only if existing host exports cannot capture needed events reproducibly. No new general orchestrator. Capture the actual invocation/options in the run manifest; do not invent a currently executable runner. The archived trial-runner-v1.txt is a reference, not a live command.
- Never put host-dependent paid agent invocations in guard-replayed proof bullets. Store runtime outputs as observable evidence; use a deterministic artifact-completeness/scoring check as the replayable seam.
- New deterministic completeness/scoring entrypoint: `bash scripts/verify-lean-eval.sh <run-dir>`. The implementation must reject missing/inconsistent metadata, incomplete required matrices and mismatched hashes; test it against a complete synthetic record and negative records before trusting it. This checks evidence structure, not semantic model correctness; an independent reviewer scores behavior.
- Required per trial: arm and instruction hashes/source SHA, fixture SHA/hash and request bytes, host/runtime/model/effort settings, author/executor/reviewer identities and exposure status, ordered tool trace, before/after tracked and untracked snapshots, diff, command exits, interruptions, outcome predicates, available latency/token/cost counters. Unavailable counters are null with reason, never inferred from character count.
- Protect secrets; fixtures have no production access. Expected outcomes and hidden tests are outside the agent-visible fixture. Do not load the candidate authoring plan, comparison rubrics, prior trial reports or another arm's instructions into the evaluated agent.

### Controlled comparison

| Arm | Frozen source |
|---|---|
| A | PR #133 head 156bc96b15715ddd8b43e69bbeaa4bcd0fff8c1a, before migration; its applicable instructions are exposed in the old layout |
| B | Actual accepted migration completion SHA, captured when this plan is promoted; migrated instructions without semantic tuning |
| C | Candidate semantic revision, captured after f2; no moving ref such as master/latest |

Use the same task-code snapshot, tools, permissions, environment and pinned model/settings for all arms. Overlay only each arm's instruction distribution and record unavoidable path/layout differences. Do not compare the repository's evolving application code as the task fixture. A→B measures migration effects; B→C measures semantic tuning. Changing model and instructions simultaneously cannot establish which caused a gain.

Fix the host binary/version and native skill loader for the first comparison. All arms expose the same six skill names through that loader: A uses its thin wrappers plus old consumer docs/playbooks, B/C use their installed references; each receives its own pinned AGENTS/WORKFLOW and matching companion files. No arm sees another arm's directories, global custom skills or stale instructions. A trace predicate verifies the intended wrapper/procedure was actually read. For the second configuration prefer the same host; if the provider requires a different host, report a separate model-plus-host stratum and limit the conclusion accordingly.

D02 uses a checkpoint controlled by the evaluator: stop after the first requirement's successful acceptance command and corresponding done Log entry are visible, before edits for the second requirement. Capture the whole dirty tree, untracked files, plan and trace; restart a fresh session against that exact snapshot with only a neutral request to resume the named initiative. Do not supply the old reasoning transcript or scorer advice. If the host cannot reliably intercept that boundary, mark this runtime case unavailable; a constructed snapshot may be reported separately as a resume fixture, never as real interruption evidence. Score the checkpoint as well as the resumed outcome.

Scoring identities are separate from instruction authorship. Give the semantic reviewer randomized arm labels and remove candidate-identifying commentary where possible without altering tool events; retain the mapping with the evaluator. Record any unavoidable unblinding (paths can reveal an arm). Blind scoring mitigates bias, but never claim hidden instructions or secure isolation beyond what the host actually provides.

Development cases (complete fixture/request/oracle authored in f1 before candidate tuning):

| Case | Workload and observable predicates |
|---|---|
| D01 | Known reversible mechanical change exceeding old numeric cap plus existing targeted checks: correct result, bounded routing, no durable plan mutation, no unrelated edits |
| D02 | Normal behavior change with two requirements and existing test seam, interrupted after first verified requirement: Goal/main flow, credible red/green, correct resume without repeating implementation, independent final review and accurate delivery |
| D03 | High-risk change with two independent owned slices and a shared integration seam: correct isolation, source-bound reports, combined-tree verification, rejection of stale proof, preserved judge boundary |
| D04 | Ambiguous product requirement and read-only review variants: no unauthorized edits, targeted clarification only for unresolved decision, continued authorized inspection, no fabricated PR/merge/deploy or bypass of owner permission |

Add exactly two fresh prospective cases from an independent fixture author/evaluator. Give that author only the public workload domains and invariant contract, not C, tuning outcomes or candidate-author conversation. Do not store hidden case inputs/oracles in the tuning worktree. Existing R/H routing cases are historical regression for this initiative; they have been inspected and cannot be claimed as fresh holdouts.

1. Run one pilot each for A/B/C on D01 and D02 (six trials) to validate capture and runtime access. Pilots are setup evidence, not improvement evidence. Stop for infrastructure or serious authority failures before multiplying trials.
2. Freeze A/B/C and six cases; run three fresh sessions per arm per case on one available strong model (54 trials). Record all attempts, including blocked/interrupted ones. Order arms to avoid all baseline trials preceding all candidate trials; keep fixture resets identical.
3. Once C has no newly failing correctness/authority/proof predicates and all first-model evidence is complete, confirm B/C on the same six cases with a second available strong-model configuration, three fresh sessions per arm/case (36 trials). A neutral cost result still proceeds for two-configuration coverage; a critical regression stops for correction. The first model plus this second model gives two tested configurations; do not require a third model. Re-baseline if runtime/settings drift.
4. Total planned execution is six pilots plus ninety scored trials. Before launching a batch, report provider/model identities and estimated runtime/cost from pilots; if runtime access or authorized spending is unavailable, mark execution blocked. Do not silently downgrade trial counts, invent model IDs or start an unlimited eval loop.
5. Correctness/authority/proof predicates are primary: no new failing predicate or incomplete required trial is allowed for acceptance. A faster wrong answer is rejected. False success, unauthorized mutation and evidence fabrication reject the candidate and require targeted correction.
6. Predeclare targeted cost metrics in f1: unnecessary clarification count on answerable work, instruction bytes actually read, redundant check invocations, and elapsed time. Report token/cost counters when available. A targeted gain requires improved median on at least one declared metric in both model comparisons, no increased critical failures and no worsened completion predicates; report all metrics and raw distributions, including tradeoffs. This is a tested-corpus result, not a population-level significance claim.
7. Equal results = no-detected-regression, mixed results = inconclusive, missing runtime/evidence = BLOCKED_CONTEXT. Quality-passing but no gain does not satisfy this plan's optimization success signal. Retain results; revise only a named candidate behavior and rerun affected comparisons plus full final suite. Exposed prospective cases become regression and need new independently authored replacements. Do not keep tuning against disclosed holdouts.

If the finite batch ends with no gain, inconclusive evidence or rejection, stop tuning and provide an honest no-promotion result, the B rollback instructions and the exact unmet requirements. Keep f3 uncompleted and request owner disposition (retain B/close as cancelled, or authorize a specifically scoped new experiment); do not reset history, claim optimization or launch another batch automatically. Only observed improvement with complete evidence proceeds to f4 and normal completion.

### Repository checks

- R: `cd cli && cargo fmt --check && cargo clippy --all-targets -- -D warnings && cargo test`
- S: `cd cli && cargo build --release --target x86_64-unknown-linux-musl && test "$(wc -c < target/x86_64-unknown-linux-musl/release/zharness)" -le 1850000`
- G: `bash scripts/test-guards.sh`
- D: `bash scripts/verify-doc-links.sh`
- V: validate each changed SKILL.md with `bash scripts/validate-skill.sh` and its file path; omit generated validation reports from commits.
- E: `bash scripts/verify-lean-eval.sh <run-dir>` after f1 creates it; actual Validation uses the concrete retained run directory, not this metavariable.

### Execution phases

- phase_slug: `f1-contract`
  status: planned
  - goal: R1, R3, R7 | depends_on: completed migration and draft promotion
  - surfaces: next ADR/index, public eval specification, minimal evidence-check script/tests, source/claim inventory in this plan's Log | avoided: candidate skill changes, private holdout content, production access
  - escalate_when: a proposed simplification changes plan grammar, independent-review policy, outward permission or unsupported-host scope
  - wave 1:
    - F1 inventory current rules against invariants and proposed behavior; record ADR authority and the exact A/B source pins — output: keep/change/remove map with no unsupported enforcement claim — check: independent map review, D — stop_if: a live rule has conflicting authority
  - wave 2:
    - F2 define D01-D04 fixtures and preregister predicates/metrics; specify evidence records; implement the minimal deterministic completeness verifier with valid/invalid synthetic records — output: public protocol and red/green verifier tests — check: `bash scripts/test-lean-eval.sh` and D — stop_if: scorer passes absent trace, mismatched hash or incomplete trial matrix
    - F3 independent evaluator authors/seals two new cases outside the tuning workspace and records identity/exposure/hash in evaluator-owned metadata — output: sealed case inventory, not contents for tuner — check: independent separation audit — stop_if: tuner sees hidden inputs/oracles
  - phase check: D, verifier tests and independent contract/evaluation review; no behavioral improvement claim

- phase_slug: `f2-candidate`
  status: planned
  - goal: R1, R2, R3, R4, R5, R6 | depends_on: f1-contract
  - surfaces: migrated six stage skills/references, WORKFLOW source/projection, git delivery wording only, prompt principles, relevant contract tests | avoided: installer semantics, guard core/schema, implicit delivery authorization
  - escalate_when: task acceptance or parser-facing fields must change, or runtime ownership cannot stay with host
  - wave 1:
    - F4 implement routing/selective context/claim cleanup; preserve normal flow; update phrase tests only where an explicitly approved semantic rule changes — output: lean routing and documented contract mapping — check: R, D, V and D01/D04 development smoke — stop_if: retaining a phrase is confused with proving behavior
  - wave 2:
    - F5 implement autonomy, verification, model-role, integration/resume and delivery wording from exact candidate behavior — output: coherent end-to-end instructions with existing durable state — check: R, G, D, V and D02/D03 development smoke — stop_if: work reports completion without current proof or widens Git authority
  - phase check: R, G, D, V and independent instruction review; freeze C source/hash before measured comparisons

- phase_slug: `f3-evaluate`
  status: planned
  - goal: R1, R2, R3, R4, R5, R6, R7 | depends_on: f2-candidate
  - surfaces: disposable fixtures/isolated host sessions, eval run records and redacted evidence | avoided: tuning against hidden expectations, global configuration and paid replay hooks
  - escalate_when: runtime/account/budget is unavailable, hidden exposure occurs, candidate violates authority, or required data is missing
  - wave 1:
    - F6 run six capture pilots, verify runtime isolation and artifact completeness, record batch estimates — output: valid pilot manifest — check: E plus independent trace inspection — stop_if: access/capture/scoring failure
  - wave 2:
    - F7 execute 54 A/B/C trials then 36 B/C confirmation trials when first comparison passes — output: complete controlled run matrices — check: E per run and independent correctness/authority scoring — stop_if: critical regression; keep all completed/blocked attempts
  - wave 3:
    - F8 independently review metrics, raw outcomes, identities and exposure; accept only the declared predicates/gain or report exact failure/inconclusive status — output: decision with runtime limits and traceable evidence — check: E and manual comparison contract review — stop_if: narrative conclusion exceeds observed evidence
  - phase check: R, G, D plus E and independent empirical verdict; no automatic completion on missing runtime or equal scores

- phase_slug: `f4-delivery`
  status: planned
  - goal: R8, R1, R7 | depends_on: f3-evaluate
  - surfaces: live README/workflow docs, relevant hand-authored site pages, ADR consequences, eval summary, plan stage-owned sections | avoided: automatic release/global install/remote publishing
  - escalate_when: publication is requested without authority or docs claim untested model/host behavior
  - wave 1:
    - F9 reconcile live guidance and site with accepted result; publish capability-role and model-upgrade reevaluation recipe; keep lessons concise and findings scoped — output: coherent docs, installed-path examples and honest performance statement — check: D, inspect changed site pages locally if layout changes, independent claim/source review — stop_if: sources or results do not support a claim
  - wave 2:
    - F10 run R/S/G/D, independent check full and closing handoff; provide outcome/artifact/proof/limitations/next receiving action — output: reviewed deliverable and recovery point — check: four gate exits, R1-R8 matrix, empirical evidence and independent final review — stop_if: an acceptance remains unmet
  - phase check: R, S, G, D and independent final review; archive only through closing handoff

### Handoff prompt for an implementing agent

Read AGENTS, WORKFLOW, this plan and the named prerequisite completion record. Confirm exactly one active plan and actual B source identity. Preserve ADR 0014 and parser-facing invariants. Execute the selected phase only, inspect current Log/Current State before resuming, and keep implementation/evaluation evidence distinct. When native delegation is supported, prefer available inexpensive researchers such as Luna for narrow repository questions with explicit ownership; otherwise the lead performs that research. The lead makes decisions; required reviews still need an independent capable judge or separate reviewer session, and cannot be silently replaced by self-review. Do not access hidden cases as the tuning author. Stop and report missing model access/evidence rather than substituting a prose simulation. Do not commit, push, merge or release unless explicitly authorized in the execution session.

## Log

## Validation

## Current State and Next Action

- active_phase: none
- lifecycle_status: planned
- blockers: migration is not yet implemented/validated/archived; this draft is not executable as a second active plan
- open_items: runtime model identities, source B/C hashes and costs are execution evidence to capture by the specified protocol, not unresolved product choices
- exact_next_action: complete playbooks-to-skills; then promote this draft in the authorized follow-up and run work full phase f1-contract
