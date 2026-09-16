# Better-Harness Eval Audit — zharness v0.21.0

**Date:** 2026-09-16; revised after owner-requested review.
**Lens:** [LangChain's Better-Harness recipe](https://www.langchain.com/blog/better-harness-a-recipe-for-harness-hill-climbing-with-evals):
behavioral evals, prospective holdout, traces, and human review. The original audit also drew on
the QoderAI better-harness configured-versus-demonstrated framing; no pinned source artifact was
retained for that prompt, so it is not authority for repository requirements.
**Scope:** repository instructions, playbooks, workflow skills, ZGUARD-CORE, CI, and recorded
validation/learning evidence. Installer product behavior was not independently reviewed here.
**Method:** configured means a mechanism exists in source; demonstrated means cited observations
show it ran. A snapshot, declaration, or missing record is not proof of enforcement or causality.
**Status:** findings remain open. This revision corrects the proposal; it does not close the active
initiative, run agent evals, or establish a harness improvement.

Remediation-plan evidence below comes from the plan recorded at `e75b977`:
`git show e75b977:docs/plans/active/audit-integrity-remediation.md`. This is an immutable evidence
reference, so moving the live plan at closure does not break it. Current lifecycle state must
still be read from the live plan before any operational action.

No baseline/change/rerun loop has been demonstrated by this audit. Manual fresh-session evals
are feasible; an automated runner is not a prerequisite. The existing `improve-harness` skill
requires a different-session rerun before a gain is claimed. The occupied active-plan slot blocks
a second durable initiative, but does not block read-only analysis or unrelated bounded changes.

## 1. Scorecard

| Dimension | Verdict | Configured | Demonstrated and limits |
|---|---|---|---|
| Task understanding | Supported in the sampled plan | `AGENTS.md` and `docs/WORKFLOW.md` define authority and request scope; planning names success criteria. | The remediation plan records nine success signals and explains why its findings collapse into three mechanisms. This is evidence about that plan, not a general agent success rate. |
| Controlled execution | Partial | Playbooks specify routing, read-only preflight, and zero-write modes. `setup/settings.json` contains hook templates. | Ten routing scenarios have reported passing observations in `docs/audit/deepseek-harness-token-audit.md`. They use one runtime/model; the stale-summary fixture did not exercise actual compaction. Effective hook behavior remains a separate question. |
| Change validation | Partial | Guard fixtures, projection parity, and sentence-contract tests cover deterministic rules. | The remediation plan records guard and reversion proofs; the routing audit records smoke observations. No prospective held-out comparison demonstrates generalization of the routing fixes. |
| Reliable delivery | Closure evidence missing | High-risk/full validation requires an independent judge; handoff requires final full review. | The remediation plan remains in-progress with empty Validation while later releases exist. This establishes a closure gap, not violation of an existing release-blocking policy. |
| Learning capture | Optional path unused; benefit unproven | `check.md` conditionally consumes a ledger; `work-full.md` requires classes on `BLOCKED_*` execution lines. | The ledger is absent and incidents remain scattered. No classified execution entries were found, but the audit did not establish a qualifying unclassified `BLOCKED_*` event. |

The original audit reported 40 passing guard fixtures on 2026-09-16. The remediation plan also
records 40/40 on 2026-09-08. Those dated results are not fresh agent-routing evidence.

## 2. Findings

### F1 — The remediation initiative lacks independent closure evidence

- **Evidence:** the high-risk remediation plan records all four phases as `in-progress` and
  leaves Validation empty. Its Current State explicitly names the missing independent review.
  Release commits `6e7427c`, `3cf8f0f`, and `2013158` followed the implementation.
  `scripts/install-git-hooks.sh` selects plan paths for plan-entry validation; it does not
  establish a general “no release while a plan is open” rule.
- **Impact:** the occupied slot prevents a second durable initiative from opening. The later
  routing changes lack entries in this plan, but that does not prove bypass: `docs/playbooks/check.md`
  explicitly allows bounded changes outside an initiative. “Gate pressure caused bypass” is an
  untested hypothesis and is not a finding.
- **Repair:** independently validate the four phases under their existing authority, with phase
  gates and a final independent full review. The original implementation range
  `a451278..b4c1ef0` is a review starting point; inspect subsequent relevant drift and pin the
  actual tested SHA. Synchronize phase/Current State for each gate, then close phases in dependency
  order through `docs/playbooks/handoff.md`. Do not fabricate same-session independence.
- **Acceptance:** matching independent entries and real proof for each phase; final full review;
  handoff's absorb/closure requirements satisfied; doc links pass after the move. Update any
  remaining live citations in the same closure scope. Historical commit references stay unchanged.
- **Boundary:** a release-blocking guard would be a new policy requiring its own accepted
  authority. This audit does not propose it. Existing tags are not changed by operational closure.
- **Status:** open; independent phase validation has not been performed by this revision.

### F2 — Routing observations are not yet a reproducible prospective eval protocol

- **Evidence:** `cli/internal/embedded/embedded_test.go` explicitly distinguishes sentence
  assertions from agent behavior. The ten-case table in `docs/audit/deepseek-harness-token-audit.md`
  records Codex CLI 0.154.0, configured model `gpt-6-astra`, an earlier two-plan routing failure,
  and passing observations for all ten scenarios after corrections/reruns. Some initial attempts
  were interrupted by usage limits. Exact fixture bytes and portable traces are not retained there.
  The stale-summary case simulates conflicting context; real compaction was not exercised.
- **Impact:** sentence tests remain useful but cannot establish routing behavior. The smoke
  observations cannot establish generalization of fixes already tuned against those inputs.
  Reassigning four historical cases to “holdout” would not repair that limitation.
- **Repair:** preserve all ten cases as regression evidence, reconstruct exact fixtures with
  explicit provenance, and have a separate evaluator author four new prospective holdout variants.
  Freeze them before future tuning; isolate tested sessions from expected answers and prior
  evidence. Record exposure honestly. A published corpus is not a hard isolation boundary.
- **Manual consumer:** the maintainer requests a comparison when reviewing a routing instruction
  change; a reviewer other than the author reads the run log and artifacts. Existing check steps
  consume only the optional failure ledger, not the routing suite. No automatic gate is added.
- **Acceptance:** every case has exact inputs, behavioral predicates, a completed baseline
  observation, ordered tool calls, before/after state including untracked files, identities,
  revision/model/settings pins, and redacted portable evidence. Interrupted runs stay visible.
  A baseline may contain failures. Equal scores without regressions support only
  `no-detected-regression`; an improvement claim needs an observed targeted gain and a repeated
  baseline/candidate comparison without observed regressions.
- **Historical control:** compare the two-plan regression case against `0dfb5b2^` and `0dfb5b2`
  in otherwise identical fixtures, three fresh trials per side. Inspection of that commit's diff
  shows the explicit total-plan-count preflight. `fcdae1f^` predates `check auto` and confounds
  the intended comparison. A passing old-side trial does not falsify the case. Log mixed results
  or `not-reproduced` and leave sensitivity unproven when appropriate; do not delete safety cases.
- **Status:** open. The manual baseline and retained evidence do not yet exist.

### F3 — The optional ledger is absent; classification use is not demonstrated

- **Evidence:** `docs/playbooks/check.md` steps 4/10 conditionally read and append to
  docs/evals/failures.md. ADR 0004 records its deletion in `655c6ac`; it remains absent.
  `docs/playbooks/work-full.md` step 7 requires a taxonomy class on `BLOCKED_*` execution
  entries. Searching plans found no recorded class values, but did not establish a qualifying
  execution entry violating that rule. Text mentioning the rule is not an execution event.
- **Impact:** optional ledger-based recurrence review is inactive. The routing miss, interrupted
  ownership recurrence, missing reversion proof, and stale local hook each have separate evidence
  but no shared incident index. This supports an opt-in experiment, not a finding that memory
  must be mandatory or that `failure_class` is broken.
- **Repair:** accept a repository-local ADR explaining the changed scope since ADR 0004 and
  ADR 0003's recurring-cost concern, then seed those four incidents. Rows name incident date,
  existing class and rationale, surface, immutable source, and actual coverage or `uncovered`.
  Class repetition triggers the existing clean-of-class review; creating a new eval/guard is a
  separate scoped action, not an automatic taxonomy rule.
- **Acceptance:** the ledger phase's own durable check reads the new file, names it in its
  receipt, and reviews classes with at least two rows. A disposable fixture with a missing
  required output demonstrates `REQUEST_CHANGES` and one appended finding row. Synthetic fixture
  incidents do not enter the real ledger. Three paired ledger-present/absent trials must each
  meet the routing oracle; two equally failing sides do not establish success.
- **Cost and recovery:** the maintainer owns curation and retention; adopting repositories pay
  the extra read on applicable checks. Consumers remain opt-in. Remove the obsolete failures.md
  exception from `.claimignore` when restoring the file. Reverting the opt-in ledger restores
  the previous absence behavior without changing consumer installations.
- **Status:** open. No ledger/ADR was created and no consumer was exercised by this revision.

### F4 — Effective hook enforcement is unverified per runtime

- **Evidence:** `setup/settings.json` contains privacy, removal-command, post-compaction, and
  question-reminder hook templates. The original audit reported that local Claude settings route
  events through an external wrapper rather than those template entries. It retained neither a
  redacted settings snapshot nor evidence of the wrapper's effective behavior.
- **Limit:** missing template entries does not prove equivalent enforcement is absent. A wrapper
  may delegate. Configured hooks also do not prove successful runtime enforcement. Claude settings
  cannot establish the behavior of the Codex runtime used in the routing observations.
- **Repair:** each new run records the runtime and effective settings. Mark each relevant hook
  behavior `verified-present`, `verified-absent`, `unknown`, or `not-applicable`, with evidence.
  Report local validation, optional hooks, CI, and branch protection separately using
  `docs/patterns/encoding-invariants.md`. An unknown state remains a declared comparison limit.
- **Acceptance:** actual run metadata and supporting observations exist; an empty schema field
  alone does not close the finding. Do not dump credentials or unrelated settings into artifacts.
- **Boundary:** no personal-settings edit or hook installation is part of this proposal.
- **Status:** open; effective local hook behavior was not reverified during this correction.

## 3. Evidence limits

- No prompt-layer baseline/change/rerun experiment or cross-runtime comparison was performed.
- No actual compaction event, branch-protection enforcement, or consumer installation was tested.
- The original audit reported 20 local Claude transcripts, but supplied no inventory or mapping
  to scored runs. Their count is not validation evidence; this revision did not inspect them.
- `docs/audit/wave-session-ab-protocol.md` defines a separate A/B protocol; no result is claimed.
- Token-cost analysis remains in the existing token audits and is outside this proposal.

## 4. Recommended sequence

1. Close F1 under the remediation initiative's own authority before promoting another durable plan.
2. Ship the manual suite and measured baseline together. Preserve historical regression cases,
   add prospective holdout variants, and record runtime enforcement as part of every actual run.
3. Accept the ledger ADR, create the optional file, and demonstrate read/append behavior during
   that phase's own validation. Do not defer required acceptance evidence past initiative closure.

Operational closure uses lifecycle proof. Routing comparisons use case observations. Ledger
adoption uses paired routing and consumer evidence. Each step has a matching verification method;
none is reported as a harness gain merely because new files exist.
