---
id: 01M0HARNESSEVALLOOP9K4
intake_id: 01M0HARNESSEVALINTK9K4
lane: normal
status: active
created: 2026-09-16
updated: 2026-09-16
---

# Plan: harness-eval-loop — reproducible observations and an optional ledger

> Promoted from draft on 2026-09-16 after audit-integrity-remediation closed and freed the
> active-plan slot. Identity, phases, requirements, and evidence obligations are unchanged from
> the reviewed draft. Promotion does not install hooks or change release policy.

## Outcome

- result: maintainers can reproduce routing observations, compare a future instruction change
  against frozen inputs, and review recurring failures through the existing optional ledger.
  Creating these documents does not create an automatic acceptance gate.
- success_signals:
  - S1: ten historical regression cases and four new prospective holdout variants have exact
    fixture bytes, requests, behavioral expectations, provenance, and immutable version IDs.
  - S2: a different session runs each case in a fresh disposable repository; retained evidence
    includes ordered tool calls, initial/final state, and scored observations.
  - S3: each run pins repository/fixture revisions, runtime version, model/settings, identities,
    exposure status, enforcement evidence, and portable redacted artifact paths/checksums.
  - S4: all 14 cases have completed baseline observations. Failures remain visible; a baseline
    measures current behavior and does not require a perfect score.
  - S5: three trials per side of the two-plan regression case compare `0dfb5b2^` with `0dfb5b2`.
    Reproduction, mixed results, and non-reproduction are valid observations. Only an observed
    difference supports a sensitivity claim.
  - S6: an accepted ADR authorizes the repository-local ledger; four evidenced incidents carry
    a taxonomy class, rationale, source, and coverage status.
  - S7: the ledger phase's own durable check names the ledger in its receipt and reviews every
    class with at least two rows. A disposable failing gate demonstrates the append behavior.
  - S8: paired ledger-present/absent trials meet the routing oracle, and all required evidence
    is complete before final closure.

## Authority and Requirements

- authority:
  - Owner's 2026-09-16 request to correct these drafts after review. This accepts proposal
    revisions; it does not start phases or enact a universal gate.
  - `AGENTS.md` and `docs/WORKFLOW.md`: scope, authority before policy, one active initiative,
    and evidence before completion.
  - `docs/audit/2026-09-16-better-harness-eval-audit.md`: findings and proposed repairs, not
    independent authority for mandatory policy.
  - `docs/audit/deepseek-harness-token-audit.md`: ten smoke observations, an earlier two-plan
    failure, interrupted attempts, and the explicit simulated-compaction limitation.
  - `docs/playbooks/check.md` steps 4/10: optional ledger read and durable finding append.
    These consumers do not automatically read routing cases or run logs.
  - `docs/playbooks/work-full.md` step 7: the existing failure taxonomy.
  - `docs/decisions/0003-durable-memory-not-wired-into-playbooks.md` and
    `docs/decisions/0004-docs-directory-deletion-655c6ac.md`: recurring cost and deletion history.
- requirements:
  - R1: score actions and state changes. Exact mode tokens may be asserted; matching or avoiding
    a playbook sentence is neither necessary nor sufficient proof of behavior.
  - R2: retain all ten historical cases as `regression`. They can guide future optimization,
    but cannot be relabeled as unseen holdout evidence for fixes they already informed.
  - R3: a separate evaluator authors four new `holdout` variants before future tuning starts:
    multiple-plan preflight, invalid-state stop, stale-summary reread, and explicit-review
    zero-write. These support prospective comparisons only.
  - R4: the maintainer explicitly requests a comparison for a routing instruction change.
    A reviewer other than the change author consumes the run log and artifacts. Equal scores
    without observed regressions mean `no-detected-regression`, not improvement. A behavioral
    improvement claim requires an observed targeted gain, no observed regression elsewhere,
    and the repeated comparison defined below.
  - R5: baseline/comparison executors differ from the instruction-change author. The holdout
    author/evaluator must not author candidate tuning. Record identities and exposure; a fresh
    session alone does not establish that holdout contents were unseen.
  - R6: report local validation, optional hooks, CI, and branch protection separately using
    `docs/patterns/encoding-invariants.md`. Per runtime and behavior, record hooks as
    `verified-present`, `verified-absent`, `unknown`, or `not-applicable` with evidence.
    Claude settings do not establish Codex enforcement; uninspected wrappers remain unknown.
  - R7: use only `MISSING_CONTEXT|WRONG_TOOL|BAD_OUTPUT|REPEATED_LOOP|UNSAFE_ACTION|LOST_DECISION|UNKNOWN`.
    Explain each classification and state `uncovered` when no relevant guard/eval exists.
  - R8: accept the ledger ADR before creating the real file. Consumer adoption remains optional.
  - R9: compare ledger presence in separate disposable fixtures. Both sides must pass the
    behavioral oracle; identical failing verdicts do not satisfy this requirement.
  - R10: no playbook, skill, guard, installer, or Go-test edits. Allow required active-plan
    bookkeeping and retained evidence explicitly in each phase's surfaces.

## Non-goals

- Changing releases. audit-integrity-remediation was independently validated and closed under
  its own authority before this promotion; reopening or amending it is out of scope here.
- A mandatory gate for every harness change, an automated runner/scorer, or a zharness verb.
- Cross-runtime conclusions, statistical significance, token-cost gains, or real compaction.
  The stale-summary case tests conflicting context only.
- Installing hooks, changing personal settings, or reading unrelated local transcripts.
- Pruning passing safety cases solely because they keep passing.

## Approach and Risks

Two independently useful phases: p1 ships case definitions, the manual protocol, and a measured
baseline together; p3 adds the optional ledger and demonstrates its consumers. Preserve existing
p1/p3 identities. Fold the old draft p2 baseline into p1 before locking; no phase has executed.
The ledger is not needed to use p1.

The minimal option is the manual suite. A runner could observe agent actions, but its maintenance
cost is not justified for this initial corpus. Sentence-contract tests remain useful alongside
behavioral observations.

### Manual evaluation contract

1. **Freeze definitions.** Write R01–R10 in smoke-table order in `docs/evals/routing.md`.
   An independent evaluator writes H01–H04 with new fixture/request combinations. Every case
   includes complete input bytes, request, expected reads/checks/writes, allowed changes, and
   provenance. Mark reconstructed historical fixtures as reconstructions. Freeze all 14 before
   the baseline; later edits create a new case-set version and preserve old results.
2. **Build isolated fixtures.** Each disposable git repository contains the tested playbooks
   and routed companions, minimal AGENTS/WORKFLOW entrypoints, PROJECT gate commands,
   `message.txt`, and the case's plans. Track `helo`, then leave a pending correction to `hello`.
   The tiny-lane gate command is `test "$(cat message.txt)" = hello`. Definitions include full
   plan bytes: verify phase, matching Current State, required lifecycle status, and empty
   Progress/Decisions/Validation. Omit plans for no-plan cases. Capture artifacts outside the fixture.
3. **Separate context.** The operator verifies an authenticated host is available and records
   version, model, effective settings, and exact invocation. Start a fresh session per case/trial;
   supply only its request and fixture. Exclude expected answers, the audit, this plan, other cases,
   prior transcripts, and personal memory. Do not expose prospective holdout contents or detailed
   failures to tuning authors. If exposure occurs, retain that case as regression evidence and
   author a replacement under a new ID before the next prospective comparison. Record the
   exposure without rewriting its original split. This is procedural separation, not a security
   sandbox claim.
4. **Retain evidence.** Use `docs/evals/evidence/<run-id>/` for redacted ordered tool traces,
   initial/final snapshots including untracked files, and scoring. Record hashes and repo-relative
   paths in `runs.md`; a machine-local transcript pointer alone is insufficient. Remove credentials
   and unrelated host data while preserving scoring events. Store historical fixture contents as
   `.txt` so they are not treated as live Markdown cross-references.
5. **Score actions.** Use `pass`, `fail`, `interrupted`, or `not-run` with reasons and trace
   pointers. Check intermediate writes and tool order as well as final diffs: a reverted write
   still violates zero-write. Stop cases cannot run gates or change state. A valid durable gate
   must read state first, append actual evidence, and synchronize both lifecycle statuses.
6. **Handle interruptions.** Usage limits, authentication/permission failures, and host outages
   produce `interrupted`. Preserve the attempt and retry in a new session with a new attempt ID
   after availability returns. Never count it as pass or agent failure. Missing sessions/artifacts
   block completion as `BLOCKED_CONTEXT` with a failure class; do not cherry-pick attempts.
7. **Measure the baseline.** One completed trial per case establishes a smoke baseline on the
   tested SHA, not automatically “v0.21.0.” Separately import the 2026-09-15 prose evidence as
   `historical-report`: retain ten reported post-fix passes and the earlier two-plan failure;
   mark missing traces/settings/identities `unrecorded` rather than inventing them.
8. **Compare the historical control.** Export check.md from `0dfb5b2^` and `0dfb5b2` into
   otherwise identical R07 fixtures; retain both texts and their diff. Run three fresh trials per
   side, alternating old/new, and log all six. `fcdae1f^` predates `check auto` and is unsuitable.
   An old-side pass does not invalidate the case. Record `not-reproduced` or mixed outcomes
   honestly; never alter inputs or delete cases just to produce a historical failure.
9. **Review future comparisons.** Pin baseline/candidate texts and use the same corpus, model,
   settings, and environment; runtime/model drift requires a new baseline. Compare both revisions
   on every case. An improvement claim requires three fresh trials per revision per case, raw
   counts, no newly failing predicates, and independent review. Mixed evidence is `inconclusive`.
   An observed regression blocks acceptance under this optional protocol. The maintainer records
   the decision/run ID in the requesting review or initiative. Existing playbooks do not invoke
   this procedure automatically.

The run log has one row per case/trial. Headers carry date/purpose, full repo SHA, case-set SHA,
target hashes, runtime/version/model/settings, instruction author, fixture author, executor,
reviewer, exposure status, enforcement evidence, artifact paths/hashes, and decision.

### Cost, dependencies, and recovery

- Requires git, Bash, the existing Go toolchain for repository gates, and an available authenticated
  agent host. No new account, API key, MCP server, runtime, or service is introduced.
- Minimum eval work: 14 baseline trials, six historical-control trials, six ledger-toggle trials,
  and one ledger-append fixture. Interrupted retries and normal phase reviews are additional.
- Added maintained surfaces: case/protocol document, run log with evidence, optional ledger,
  and its ADR/index entry. The maintainer owns curation and retention. Audit prose lacks exact
  fixtures; existing plan logs do not feed the ledger consumer. Removing the opt-in file is cheap,
  but deleting retained evidence needs an explicit retention decision.
- Fragile assumption: the historical failure reproduces on the selected model. If it does not,
  the baseline remains useful; sensitivity and improvement remain unproven. Retain safety cases
  even when saturated.
- Rollback reverts only this initiative's changes through normal review; preserve observations
  until their removal is accepted. No consumer data, release tags, or personal settings change.
- Verification failures follow `docs/playbooks/work-full.md` step 6. Do not relax expectations
  or erase observations to obtain a clean verdict.

## Phases and Verification

- planning_status: executed — phases locked at promotion; no phase has executed yet

- phase_slug: p1-routing-eval-set
  story_id: story-routing-eval-set-20260916
  status: checked
  goal: a reproducible manual suite and measured baseline usable without a ledger.
  depends_on: none
  surfaces_touched: docs/evals/routing.md, docs/evals/runs.md, docs/evals/evidence/**,
    this plan's lifecycle-owned sections
  surfaces_avoided: docs/playbooks/**, skills/**, cli/**, scripts/**, .github/**, personal settings
  waves:
    - wave: 1
      tasks:
        - task: p1.w1.t1 — write the protocol and R01–R10; arrange independent authorship of
            H01–H04 and freeze all definitions before the baseline.
          verify: `awk '/^### [RH][0-9][0-9] /{n++} /^- split: regression$/{r++} /^- split: holdout$/{h++} END{print n,r,h; exit !(n==14 && r==10 && h==4)}' docs/evals/routing.md`
            passes; a reviewer checks complete input bytes, requests, behavioral predicates,
            provenance, and prospective-only holdout claims for all 14 cases.
    - wave: 2
      tasks:
        - task: p1.w2.t1 — create the run schema, import historical-report provenance, and
            execute/retain all 14 baseline observations under the manual contract.
          verify: reconcile the 14 IDs with the frozen manifest; every case has a completed
            pass/fail, a fresh-session trace, initial/final state, and matching artifact hashes.
            Interrupted attempts remain additional rows. Record the manual checklist in Progress.
        - task: p1.w2.t2 — run all six historical-control trials; record old/new texts and
            the sensitivity conclusion permitted by their observed results.
          verify: `git show 0dfb5b2 -- docs/playbooks/check.md` identifies the intended preflight
            change; retained texts match their source revisions. Account for all six trials and
            confirm no case was removed following an old-side pass.
  expected_output: frozen suite, manual protocol, run log, and portable redacted evidence.
  check: `bash scripts/verify-doc-links.sh` and `bash scripts/test-guards.sh` pass.
    Inspect `git status --short` and `git diff --stat` for declared surfaces, including new files
    and required lifecycle writes. Cite the independent behavioral review at the normal-lane gate.

- phase_slug: p3-failure-ledger
  story_id: story-failure-ledger-20260916
  status: checked
  goal: an authorized optional ledger with demonstrated read and append behavior.
  depends_on: p1-routing-eval-set
  surfaces_touched: docs/decisions/0010-local-failure-ledger.md, docs/decisions/README.md,
    docs/evals/failures.md, docs/evals/runs.md, docs/evals/evidence/**, .claimignore
    (remove only the obsolete failures.md exception), this plan's lifecycle-owned sections
  surfaces_avoided: docs/playbooks/**, skills/**, cli/**, scripts/**, .github/**, personal settings
  waves:
    - wave: 1
      tasks:
        - task: p3.w1.t1 — write and accept ADR 0010 before creating the ledger; index it.
            Explain the changed scope from ADR 0004, ADR 0003's recurring cost, maintainer
            ownership, optional consumer adoption, and rollback by reverting the opt-in file.
          verify: `test -f docs/decisions/0010-local-failure-ledger.md` and
            `test "$(grep -c '0010-local-failure-ledger' docs/decisions/README.md)" -eq 1` pass.
            Record acceptance in Decisions before wave 2. If 0010 is occupied at promotion,
            resolve the draft's filename before locking, not during execution.
    - wave: 2
      tasks:
        - task: p3.w2.t1 — seed four incidents: two-plan routing miss; interrupted-install
            ownership recurrence; missing reversion proof; stale local pre-commit hook.
            Each row has incident date, class/rationale, surface, immutable source commit and
            path/section, coverage or `uncovered`. Use the routing audit and remediation plan
            recorded at `e75b977`; do not depend on its active path. Remove the obsolete
            failures.md exception from .claimignore.
          verify: `test -f docs/evals/failures.md` and `bash scripts/verify-doc-links.sh` pass;
            manually trace all four rows to their sources and confirm coverage claims.
        - task: p3.w2.t2 — run R03 (valid durable gate) three times with the ledger and three
            times without it in separate fresh fixtures. Require correct routing and synchronized
            state on both sides; ledger receipts and class-review output may differ.
          verify: review all six trials; every routing predicate passes. Any failure blocks this
            task. Retain the evidence and observation rows in runs.md.
        - task: p3.w2.t3 — use one disposable R03-derived fixture with two synthetic BAD_OUTPUT
            ledger incidents and a plan requiring a second output file absent from the candidate
            diff. Keep the ordinary message check passing; invoke a durable gate independently.
          verify: trace shows repeated-class review, `REQUEST_CHANGES` for the missing output,
            a new ledger row for that finding, and both statuses still `in-progress`. Label
            synthetic evidence; never copy its incident into the real ledger.
  expected_output: accepted ADR/index, real ledger, obsolete exception removed, toggle/append
    observations, evidence, and the real ledger-consumption receipt.
  check: `bash scripts/verify-doc-links.sh`, `bash scripts/test-guards.sh`, and
    `cd cli && CGO_ENABLED=0 go build ./... && go vet ./... && go test ./...` pass.
    This phase's own durable check reads failures.md, names it in the receipt, and records
    clean-of-class for repeated classes (or explicitly none). Keep the final phase `in-progress`
    for independent `check full`, then handoff. S7 is demonstrated here, never deferred.

## Current State and Next Action

- active_phase: p3-failure-ledger
- lifecycle_status: checked
- latest_run_id: run-p3-20260916T0700Z (p1 anchor: `run-p1-20260916T0220Z`; eval run
  `run-001-20260916`, extended with the p3 wave-2 ledger observations)
- latest_check_id: check-p3-20260916T0820Z (full, APPROVE_WITH_REQUESTS, judge independent —
  the initiative's one required independent review; prior anchors: `check-p3-20260916T0705Z`,
  gate, APPROVED, judge same-session, and `check-p1-20260916T0335Z`, gate,
  APPROVE_WITH_REQUESTS, judge same-session)
- latest_handoff_id: none
- completed: audit and draft corrected after review; draft promoted to active on 2026-09-16 at
  repo SHA `2013158` with identity preserved. `p1-routing-eval-set` waves 1 and 2 executed: the
  14-case suite is frozen in `docs/evals/routing.md` (v1, hash `270af72f…`), run 001 is a
  completed baseline in `docs/evals/runs.md` with all 14 observations plus six historical-control
  trials, and evidence is retained under `docs/evals/evidence/run-001/`. `p3-failure-ledger`
  waves 1 and 2 executed: ADR 0010 is written, accepted and indexed; `docs/evals/failures.md`
  exists with its four seeded incidents and the obsolete `.claimignore` exception is gone; and
  both of the ADR's claims have observed evidence in `docs/evals/runs.md` — six alternating R03
  trials (ledger present vs. absent, all six routing predicates passing on both sides) and one
  append trial in which a durable gate reviewed a repeated class, returned `REQUEST_CHANGES`, and
  appended a row. The phase gate has not run yet.
- blockers: none.
- open_items:
  - Promotion preconditions rechecked 2026-09-16: ADR slot `0010` is free (`docs/decisions/`
    holds 0001-0009); authenticated agent host available (`claude` 2.1.273, model
    `claude-opus-5`); repo SHA `2013158`.
  - Eval-host isolation is `--setting-sources project,local --strict-mcp-config`, which excludes
    the operator's personal `~/.claude` instructions, rules, memory, and MCP servers.
    `CLAUDE_CONFIG_DIR` isolation was tried first and is unusable: it produces an unauthenticated
    zero-token run. Record the effective settings per run under R6.
  - Two baseline failures are findings to carry, not defects to fix here: H04 rationalized an
    explicit `check review` into a durable gate and wrote plan state, and H01 announced its mode
    after reading plans, against `docs/playbooks/work.md:13`. `docs/playbooks/**` is in this
    phase's `surfaces_avoided`, so neither is repaired by this initiative; both are candidate
    ledger rows.
  - `docs/decisions/README.md`'s index is missing rows for ADRs 0007, 0008 and 0009 (the table
    jumps 0006 → 0010 after this phase's insert). Pre-existing, found while indexing 0010, and
    outside `p3.w1.t1`'s output; a backfill is a separate bounded change.
  - Carried forward from audit-integrity-remediation closure (both out of scope under R10,
    neither is a blocker here): the installed `.git/hooks/pre-commit` wrapper's call sequence can
    go stale relative to `scripts/install-git-hooks.sh` (the ZGUARD-CORE half is re-extracted per
    run and cannot); and `stashRestore` writes to `filepath.Join(root, e.rel)` with `e.rel` read
    back from `.zharness/update-stash/stash.tsv`, so a `../` component there escapes the repo root
    on `zharness update --abort`. Both are candidate ledger rows in p3.w2.t1, not fixes.
  - Requests from the independent `check full` of 2026-09-16T08:20Z, all non-blocking and none
    repaired by this initiative (each would edit an artifact this phase already closed, and F1/F2
    also name files under `surfaces_avoided` only as citation targets, not as edits):
    - F1 (major) — **repaired 2026-09-16T08:35Z, after that review entry.**
      `docs/playbooks/work-full.md:7` was a wrong line anchor for the seven-token taxonomy, which
      sits at line 28; corrected in all four places (`docs/evals/failures.md` Conventions and ADR
      0010's Context, Decision and Authority). The prose "(step 7)" was already correct. Proven by
      `test "$(sed -n 28p docs/playbooks/work-full.md | grep -c 'MISSING_CONTEXT|WRONG_TOOL')" -eq 1`.
    - F2 (major) — **repaired 2026-09-16T08:35Z, after that review entry.**
      `docs/evals/failures.md` row 1 cited `docs/playbooks/check.md:31` (step 2); step 2 and the
      sentence it quotes are at line 15, and line 31 is step 6. Corrected to `:15` and proven by
      `test "$(sed -n 15p docs/playbooks/check.md | grep -c 'list every candidate')" -eq 1`.
    - F1/F2 residual, and the reason this initiative is not closed: the class behind both is that
      `scripts/verify-doc-links.sh` checks only that a cited path exists and never validates the
      `:NN` suffix, so every line anchor in this repository is unverified — two of five were wrong
      at authoring time. The class predates this initiative (`docs/audit/` carries roughly twenty
      such anchors and `docs/plans/completed/` five more) and guarding it would touch `scripts/`
      and `docs/patterns/`, both in this phase's `surfaces_avoided`. It has no ADR and no guard,
      which is exactly what `handoff.md` step 6 refuses to leave behind.
    - F1/F2 shared cause: line-number anchors into live playbooks rot like the active-plan paths
      repaired in ADRs 0008/0009, and two of five were already wrong when written. A bounded
      follow-up should drop the `:NN` form in favor of the narrative "(step N)".
    - F3 (minor) — **repaired 2026-09-17, after that review entry.** ADR 0010's Status and
      Authority lines now name `harness-eval-loop.md` (then under `docs/plans/active/`), the
      form `docs/evals/failures.md` already uses; no follow-up edit is owed at closure.
    - F4 (minor): retained evidence embeds operator absolute paths — 81 occurrences of
      `/home/tinhpt/.claude/projects/...` in `trace.jsonl` files and 10 of
      `/home/tinhpt/Lab/mono-harness` in the retained generator scripts. No credential material
      accompanies them and the artifact paths `runs.md` references are all repo-relative, so S3
      holds; the raw traces are the residue.
- exact_next_action: **owner decision on 2026-09-16: do not close this initiative in that
  session.** The plan stays `active` and was never `git mv`'d. To resume, pick one of the two
  paths below, then run `handoff` again; nothing is owed to `work` or `check`.
  (a) Close cheaply: accept that the unverified-line-anchor class stays unguarded, clear the F4
  open item (F4) by decision rather than by fix, write `absorb: none` in `## Decisions`, and
  run `handoff` steps 5-7. This is blocked today only by `handoff.md` step 6's requirement that
  `open_items` read `none` and that a class-of-failure have an ADR or guard.
  (b) Close properly: first encode the invariant behind F1/F2 — `scripts/verify-doc-links.sh`
  validates only that a cited path exists and never checks the `:NN` suffix, so every line anchor
  in this repository is unverified. That is a bounded change under `scripts/` and
  `docs/patterns/encoding-invariants.md`, both of which are in this phase's `surfaces_avoided`, so
  it belongs to a separate initiative and not to a late edit here. Then close via (a).
  Either way, the independent `check full` of 2026-09-16T08:20Z already satisfies the final
  phase's one required complete review, so re-reviewing is not a precondition for closure.

## Progress
<!-- Execution evidence. -->
- [2026-09-16T02:20Z] p1-routing-eval-set / wave 1 / phase start / task_status=in-progress —
  plan promoted from draft to `docs/plans/active/harness-eval-loop.md` with identity preserved
  (`id: 01M0HARNESSEVALLOOP9K4`, `intake_id: 01M0HARNESSEVALINTK9K4`, phase slugs and story ids
  unchanged). Promotion preconditions rechecked and recorded in Current State. run anchor
  `run-p1-20260916T0220Z`. Surfaces: `docs/plans/active/harness-eval-loop.md`.
- [2026-09-16T02:35Z] p1-routing-eval-set / wave 1 / p1.w1.t1 / task_status=DONE — froze all 14
  case definitions in `docs/evals/routing.md` (case_set_version v1, case_set_hash
  `270af72fbe48caa3a03b044fc384ea75e7f6373f`). R01-R10 reconstruct the ten 2026-09-15 smoke rows
  from `docs/audit/deepseek-harness-token-audit.md` §4 S1 and are labelled reconstructions;
  H01-H04 were authored prospectively by isolated session
  `b5aa99dc-124e-4869-968c-05048b3926f1` (`claude-opus-5`, $1.238) which received only the eight
  playbooks, `AGENTS.md`, `docs/WORKFLOW.md`, `docs/patterns/encoding-invariants.md`, the fixture
  shape, and four category names — no regression case, audit, run log, or expected answer.
  verification: `awk '/^### [RH][0-9][0-9] /{n++} /^- split: regression$/{r++} /^- split: holdout$/{h++} END{print n,r,h; exit !(n==14 && r==10 && h==4)}' docs/evals/routing.md`
  printed `14 10 4`, rc=0. `bash scripts/verify-doc-links.sh` -> `doc links OK (0 findings)`.
  Surfaces: `docs/evals/routing.md`, `docs/evals/runs.md`, `docs/evals/evidence/**`.
- [2026-09-16T02:35Z] p1-routing-eval-set / wave 1 summary — one task, DONE. The frozen suite,
  the deterministic fixture generator, the trial runner, the holdout author's verbatim output and
  brief, and both historical-control `check.md` texts plus their diff are retained under
  `docs/evals/evidence/`. Baseline execution begins in wave 2; no baseline trial has been scored.
  Declared limit: the independent behavioral review of the 14 definitions required by p1.w1.t1's
  verify clause has not yet run; it is part of the p1 gate, not of this task's own command.

- [2026-09-16T03:05Z] p1-routing-eval-set / wave 2 / p1.w2.t1 / task_status=DONE — created the run
  schema and executed all 14 baseline observations in fresh isolated sessions, retained under
  `docs/evals/evidence/run-001/`. Verification ran: all 14 frozen case IDs reconcile 1:1 with a
  completed verdict row plus six retained artifacts each (`request.txt`, `state-before.txt`,
  `state-after.txt`, `tracked-diff.txt`, `summary.txt`, `trace.jsonl`); `sha256sum -c
  SHA256SUMS.txt` matched all 246 files. Result 12 pass / 2 fail (H01 mode-ordering, H04
  explicit-`review` upgraded to a durable gate with 2 845 bytes written). Nine invalid attempts
  retained as `not-run` under `evidence/run-001/invalid/` with their defect causes: three
  driver word-split build failures, two generator-v1 R02/R07 fixtures whose "unrelated" plan
  targeted `message.txt`, and four extractor-v1 holdout fixtures that swept the pending
  `message.txt` correction into the base commit. Both generator versions retained; no case
  definition changed. Surfaces: `docs/evals/runs.md`, `docs/evals/evidence/**`.
- [2026-09-16T03:20Z] p1-routing-eval-set / wave 2 / p1.w2.t2 / task_status=DONE — ran all six
  historical-control trials (R07 case, three per side, alternating). Verification ran:
  `git show 0dfb5b2 -- docs/playbooks/check.md` identifies the added total-plan-count preflight;
  both retained texts `diff`-match `0dfb5b2^` and `0dfb5b2` exactly; all six trials accounted for
  and no old-side trial passed, so no case was removed. Old side 0/3 (each resolved `mode: gate`
  and wrote `docs/plans/active/greeting.md`), new side 3/3 zero-write stops. Conclusion:
  `reproduced`. Surfaces: `docs/evals/runs.md`, `docs/evals/evidence/**`.
- [2026-09-16T03:20Z] p1-routing-eval-set / wave 2 / wave_summary — both wave 2 tasks DONE, no
  retries and no `BLOCKED_*`. Three defects found were in this initiative's own disposable
  tooling, never in a frozen case definition or in the harness under test; each was fixed in a
  new generator version, the superseded attempts retained as `not-run`, and the affected trials
  re-run. Run 001 decision recorded as **baseline** — not `no-detected-regression` (no prior
  comparable run) and not `improvement` (no instruction change was tested). Phase moves to its
  in-session gate.

- [2026-09-16T03:45Z] p3-failure-ledger / wave 1 / phase start / task_status=in-progress — phase
  set `in-progress` with Current State agreeing. run anchor `run-p3-20260916T0340Z`. ADR slot
  `0010` rechecked and still free. Surfaces: `docs/plans/active/harness-eval-loop.md`.
- [2026-09-16T03:50Z] p3-failure-ledger / wave 1 / p3.w1.t1 / task_status=DONE — wrote
  `docs/decisions/0010-local-failure-ledger.md` (Accepted 2026-09-16) and added its index row to
  `docs/decisions/README.md`. The ADR states the changed scope from ADR 0004 (that recovery left
  this path unrestored because nothing read it; `docs/playbooks/check.md:29,37` now do), ADR
  0003's recurring-cost argument and why optional-consumer wiring does not repeat it, maintainer
  ownership with no installer or template scaffolding, and rollback by deleting the opt-in file.
  Three alternatives are recorded as rejected: a derived index, a mandatory/commit-gating ledger,
  and reusing `docs/memory/`. Verification ran:
  `test -f docs/decisions/0010-local-failure-ledger.md` rc=0;
  `test "$(grep -c '0010-local-failure-ledger' docs/decisions/README.md)" -eq 1` rc=0;
  `bash scripts/verify-doc-links.sh` clean. Acceptance recorded in Decisions before wave 2, per
  the task. Surfaces: `docs/decisions/0010-local-failure-ledger.md`, `docs/decisions/README.md`.
- [2026-09-16T03:50Z] p3-failure-ledger / wave 1 / wave_summary — one task, DONE, no retries and
  no `BLOCKED_*`. One adjacent finding left unfixed on purpose: `docs/decisions/README.md`'s
  index is missing rows for 0007, 0008 and 0009, so the table now reads 0006 then 0010. The gap
  predates this phase and backfilling it is not this task's output; it is carried as an open item.
  Wave 2 proceeds to seed the ledger itself.
- [2026-09-16T06:30Z] p3-failure-ledger / wave 2 / p3.w2.t1 — DONE. Seeded
  `docs/evals/failures.md` with the four required incidents: the two-plan routing miss, the
  interrupted-install ownership recurrence, the missing reversion proof, and the stale local
  pre-commit hook. Each row carries incident date, class and rationale, surface, an immutable
  source commit with path/section, and a coverage claim or the literal `uncovered`. The routing
  audit and remediation plan are cited at `e75b977`, not at any active path. Removed the obsolete
  `docs/evals/failures.md` exception from `.claimignore`. Verification ran:
  `test -f docs/evals/failures.md` rc=0; `bash scripts/verify-doc-links.sh` clean after one
  targeted fix. Retries: 1 — the first run failed on
  `docs/evals/failures.md -> docs/plans/active/audit-integrity-remediation.md`, my own convention
  forbids citing an active-plan path and the verifier proved it; rewrote that citation as a
  non-link. All four rows were traced to their sources by hand and their coverage claims
  confirmed. Surfaces: `docs/evals/failures.md`, `.claimignore`.
- [2026-09-16T06:56Z] p3-failure-ledger / wave 2 / p3.w2.t2 — DONE. Ran R03 six times in six
  fresh disposable fixtures, alternating ledger-present and ledger-absent, one session each. The
  two sides were proved to differ only in the ledger:
  `diff -rq --exclude=.git --exclude=evals` over freshly built on/off fixtures exits 0 and both
  carry the identical pending ` M message.txt` (retained as
  `docs/evals/evidence/ledger-sides-identity-v1.txt`). All six trials pass all five R03
  `expect_pass` predicates, scored mechanically from the ordered tool trace and the fixture
  after-state, never from closing prose; neither `expect_fail` condition fired. The permitted
  difference appeared only where ADR 0010 predicts it: the on side named
  `docs/evals/failures.md` in `failure_ledger:` and reviewed repeated classes 3/3, the off side
  wrote the literal `absent` and reviewed none 3/3. Routing and lifecycle state were identical
  across both sides. Retries: 1 — `score-r03.py` v1 treated the first assistant text as the mode
  announcement and so mis-scored `led-off-1`, which narrated first; v2 takes the first text that
  declares a mode, both versions retained. The defect was in this initiative's disposable scorer,
  not in a frozen case definition and not in the harness. Observation rows and both tables are in
  `docs/evals/runs.md`; evidence under `docs/evals/evidence/run-001/led-{on,off}-{1,2,3}/`.
  Surfaces: `docs/evals/runs.md`, `docs/evals/evidence/**`.
- [2026-09-16T06:58Z] p3-failure-ledger / wave 2 / p3.w2.t3 — DONE. One disposable R03-derived
  fixture carrying two synthetic `BAD_OUTPUT` incidents, a plan declaring a second output
  `greeting.txt` absent from the candidate diff, and the ordinary `message.txt` check still
  passing; the durable gate was invoked in an independent session
  (`a6e9751f-d614-4fe0-a0db-281a68a66f93`). All four verification predicates pass, read from the
  trace and the after-state: the trace reads `failures.md` and the Validation entry names
  `BAD_OUTPUT (2026-09-10, 2026-09-12)` and states the diff is not clean of it; the verdict is
  `REQUEST_CHANGES` with finding C1 naming the missing `greeting.txt`; the ledger gained a row
  for that finding; and both the phase status and `lifecycle_status` remain `in-progress`. All
  fixture incidents are labelled `[SYNTHETIC]`/`[SYNTHETIC FIXTURE]` and none was copied into the
  real ledger. Retries: 1 — `runtrial-a.sh` v1's summariser aborted on an unbound `$PLANSET`
  under `set -u`; the trial had already completed so the summary was regenerated from the
  retained trace and snapshots with no model call repeated, and the lost shell exit code is
  recorded as unrecoverable. Surfaces: `docs/evals/runs.md`, `docs/evals/evidence/**`.
- [2026-09-16T07:00Z] p3-failure-ledger / wave 2 / wave_summary — three tasks, all DONE, no
  `BLOCKED_*`. Two retries, both against this initiative's own disposable tooling (the R03 scorer
  and the append trial's summariser); every superseded version is retained beside its replacement.
  Both of ADR 0010's claims now have observed evidence rather than assertion: the ledger is
  optional for consumers (6/6 routing predicates pass with it absent) and it is genuinely
  consumed when present (repeated-class review, a `REQUEST_CHANGES` traced to it, and an appended
  row). S7 is demonstrated here, not deferred. The phase gate is next, and this being the final
  phase it stays `in-progress` for an independent `check full`.
- [2026-09-16T08:35Z] p3-failure-ledger / post-review citation repair — DONE, and recorded
  **after** the independent `check full` entry of 2026-09-16T08:20Z. That entry therefore covers
  the pre-repair state on exactly two citation values, and nothing else in it is affected. Repaired
  its findings F1 and F2, both plain factual errors in my own artifacts:
  `docs/playbooks/work-full.md:7` -> `:28` (the seven-token taxonomy; four occurrences, in
  `docs/evals/failures.md` Conventions and ADR 0010's Context, Decision and Authority) and
  `docs/playbooks/check.md:31` -> `:15` in `docs/evals/failures.md` row 1 (the quoted
  one-active-plan preflight sentence; line 31 is step 6). Verification ran:
  `test "$(sed -n 28p docs/playbooks/work-full.md | grep -c 'MISSING_CONTEXT|WRONG_TOOL')" -eq 1`
  rc=0; `test "$(sed -n 15p docs/playbooks/check.md | grep -c 'list every candidate')" -eq 1`
  rc=0; `bash scripts/verify-doc-links.sh` clean. Only the numeric values changed; the narrative
  "(step N)" text was already correct everywhere and was left alone. F3 and F4 were deliberately
  not touched, and the underlying class stays unguarded — see open_items.
- [2026-09-17] p3-failure-ledger / post-review citation repair (F3) — DONE, also recorded **after**
  the independent `check full` entry of 2026-09-16T08:20Z, which is left unedited. ADR 0010 lines 6
  and 125 cited `docs/plans/active/harness-eval-loop.md`; both now read `harness-eval-loop.md`
  (then under `docs/plans/active/`). Verification ran:
  `test "$(grep -c 'docs/plans/active/harness-eval-loop.md' docs/decisions/0010-local-failure-ledger.md)" -eq 0`
  rc=0; `bash scripts/verify-doc-links.sh` clean. The reviewer's F3 proof
  (`grep -c 'docs/plans/active/'` = 2) still returns 2, because the historical note keeps the
  directory name, so it no longer evidences the defect. F4 (would change hashed evidence) and the
  unguarded `:NN` class (touches `scripts/`, in surfaces_avoided) remain open and untouched.

## Decisions

- [2026-09-16 (draft revision)] Owner requested corrections following review. Preserve initiative
  identity and keep the draft outside active/. Editing these documents does not start execution.
- [2026-09-16 (draft revision)] Fold the former p2-holdout-baseline into p1 so the first phase
  ships usable observations. Preserve p1/p3 identities; there is no executed p2 history.
- [2026-09-16 (draft revision)] Keep historical cases as regression evidence; new variants support
  future holdout comparisons only while exposure is controlled and recorded.
- [2026-09-16 (draft revision)] Keep lane normal: repository-local optional artifacts and existing
  conditional consumers, with independent behavioral evaluation and final independent full review.
  A mandatory cross-repository gate would require a separate policy decision.
- [2026-09-16 (draft revision)] Use manual runs, retain passing safety cases, allow honest
  historical non-reproduction, and demonstrate ledger consumption before closure.
- [2026-09-16 (draft revision)] The doc-link verifier excludes docs/plans/**; future paths may
  be named here. Live audit links still need to resolve. Lifecycle writes are allowed surfaces.

- [2026-09-16 / p1 / promotion] Promote the reviewed draft unchanged except for lifecycle
  fields, the stale slot-blocked text, and recorded promotion preconditions. Rationale: the draft
  was reviewed and approved as-is; promotion is a lifecycle transition, not a re-plan.
- [2026-09-16 / p1.w1.t1 / eval-host isolation] Isolate each tested session with
  `--setting-sources project,local --strict-mcp-config` rather than `CLAUDE_CONFIG_DIR`.
  Rationale: contract step 3 requires excluding personal memory, and the `CLAUDE_CONFIG_DIR`
  approach produces an unauthenticated zero-token run (no usable host). The chosen flags were
  verified empirically to remove the operator's global instructions, rules, memory, and MCP
  servers (probe session `f743e44e-6b4e-4e1b-98d2-6c5d97db2a9f`, reply `SOUL=NO TOOLS=0`).
- [2026-09-16 / p1.w1.t1 / fixture determinism] Specify fixture bytes by a retained deterministic
  generator plus the harness source SHA and a plan-set token, rather than by pasting every file
  into the case definitions. Rationale: the fixture is ~45 KB of verbatim playbook copies per
  case; the generator plus SHA reproduces those bytes exactly and keeps the case set readable.
- [2026-09-16 / p1.w1.t1 / permission mode] Run tested sessions with
  `--permission-mode bypassPermissions` in disposable fixtures. Rationale: contract step 6 makes a
  permission failure an `interrupted` result; allowing writes prevents mis-scoring a denied edit
  as an agent routing failure. The fixtures are throwaway repositories holding no real data.

- 2026-09-16 — p3-failure-ledger / p3.w1.t1 — ADR 0010 accepted: `docs/evals/failures.md` is
  created as a maintainer-owned, append-only local ledger, and no playbook is edited to require
  it. Rationale: `docs/playbooks/check.md:29` and `:37` already reference the path behind
  `IF … exists` guards, so the file's creation activates existing optional behavior instead of
  adding a mandatory step — which is the specific line ADR 0003 drew against wiring durable
  memory into the spine. Run 001's two failures are recurrences of classes with no guard, which
  is what step 4's "recorded two or more times" clause exists to surface and what it cannot do
  without a store. Rollback is a single `git rm`, both consumers being conditional on existence.
  R9's presence/absence comparison in wave 2 is what would falsify the value claim; the ADR is
  accepted before the file is created, per R8, not after the measurement.

## Validation

- 2026-09-16T03:35Z — phase `p1-routing-eval-set` — APPROVE_WITH_REQUESTS — mode: gate — in-session
  phase gate per `docs/playbooks/work-full.md` step 11 (`docs/playbooks/check.md` steps 1-4, 6-11).
  - `bash scripts/verify-doc-links.sh` — pass: `doc links OK (0 findings; 10 claim(s) under known-removed v0.15 surfaces)`. First run failed with 1 finding (a fixture path in the run log read as a live cross-reference); one targeted fix rewrote it as a bare filename with a parenthetical, matching the convention already used in `docs/evals/routing.md`. No retry after that.
  - `bash scripts/test-guards.sh` — pass: `guards: 40 passed, 0 failed`.
  - `awk '/^### [RH][0-9][0-9] /{n++} /^- split: regression$/{r++} /^- split: holdout$/{h++} END{print n,r,h; exit !(n==14 && r==10 && h==4)}' docs/evals/routing.md` — pass, prints `14 10 4`: the frozen case set is intact at 14 cases, 10 regression, 4 holdout.
  - `test "$(ls -d docs/evals/evidence/run-001/r001-[RH][0-9][0-9] | wc -l)" -eq 14` — pass: one retained artifact directory per frozen case.
  - `test "$(grep -c "^| .r001-[RH][0-9][0-9]. |" docs/evals/runs.md)" -eq 14` — pass: one completed verdict row per case, reconciled 1:1 with the artifact directories.
  - `test "$(ls -d docs/evals/evidence/run-001/hc-[on][le][dw]-[123] | wc -l)" -eq 6` — pass: all six historical-control trials retained, three per side.
  - `cd docs/evals/evidence/run-001 && sha256sum -c SHA256SUMS.txt --quiet` — pass, silent: all 246 retained artifact files match the manifest recorded at publication.
  - scope: on target. `git status --short` and `git diff --stat` show writes confined to
    `docs/evals/routing.md`, `docs/evals/runs.md`, `docs/evals/evidence/**`, and this plan's
    lifecycle-owned sections. Nothing under `docs/playbooks/**`, `skills/**`, `cli/**`,
    `scripts/**`, `.github/**`, or personal settings was touched, which is what
    `surfaces_avoided` requires. The unrelated in-flight edits to
    `docs/decisions/0008-*.md` and `docs/decisions/0009-*.md` and the completed
    `audit-integrity-remediation` move predate this phase and belong to the preserved F1 work.
  - requests (non-blocking, carried as open items, not fixed here): the two baseline failures are
    behavioral findings against `docs/playbooks/check.md` and `docs/playbooks/work.md`, both of
    which sit in this phase's `surfaces_avoided`. H04 rationalized an explicit `check review` into
    a durable gate and wrote 2 845 bytes of plan state; H01 announced its mode after reading
    plans, against `docs/playbooks/work.md:13`. They are recorded as candidate ledger rows for
    `p3-failure-ledger`, not repaired by this phase.
  - judge: same-session
  - judge_model: claude-opus-5
  - not_independently_verified: the phase author, the trial executor and the scorer are one
    identity. Case authorship is split — H01-H04 came from isolated session
    `b5aa99dc-124e-4869-968c-05048b3926f1`, which never saw a regression case, this plan, the
    audit, or any expected answer — so holdout scoring is independent of case design but not of
    execution, and R01-R10 scoring is independent of neither. The behavioral review that
    `expected_output` calls for is therefore recorded as an open request, not claimed here.
  - proof_gaps: the commands above verify the suite's structure, the artifacts' integrity and the
    run log's completeness. No command re-derives a verdict from a trace: the 14 verdicts and the
    `reproduced` conclusion rest on manual reading of the ordered `tool_use` streams and the
    before/after snapshots, recorded case by case in `docs/evals/runs.md`. That is the manual
    contract this phase set out to build, not a gap introduced by it.
  - receipt: context_sources: AGENTS.md, docs/WORKFLOW.md, docs/playbooks/work-full.md,
    docs/playbooks/check.md, docs/playbooks/check-validation.md, docs/evals/routing.md,
    docs/evals/runs.md, docs/evals/evidence/run-001/**, docs/audit/2026-09-16-better-harness-eval-audit.md /
    policy: p1-routing-eval-set expected_output and check / judge: same-session /
    judge_model: claude-opus-5 / retries: 1 (doc-links, one targeted fix) /
    rollback_point: 2013158 (all phase output is uncommitted; `git checkout -- docs/plans/active/harness-eval-loop.md` and removing the untracked `docs/evals/` tree reverts it) /
    failure_ledger: absent / enforcement: hook (`.git/hooks/pre-commit` is installed and
    re-extracts the ZGUARD-CORE proof re-executor per run, so these proofs re-run at commit
    time) / not_independently_verified: see the line above

- 2026-09-16T07:05Z — phase p3-failure-ledger — APPROVED — mode: gate
  - `bash scripts/verify-doc-links.sh` — rc=0, "doc links OK (0 findings; 10 claim(s) under known-removed v0.15 surfaces)".
  - `bash scripts/test-guards.sh` — rc=0, "guards: 40 passed, 0 failed".
  - `cd cli && CGO_ENABLED=0 go build ./... && go vet ./... && go test ./...` — rc=0, all four packages ok.
  - `test -f docs/evals/failures.md` — rc=0. The ledger this phase authored exists and was read during this gate.
  - `awk '/^- class:/{c[$3]++} END{n=0; for(k in c) if(c[k]>1) n++; print "classes_recorded_twice_or_more=" n; exit (n>0)}' docs/evals/failures.md` — rc=0, "classes_recorded_twice_or_more=0". Clean-of-class review per check.md step 4: **explicitly none**. The real ledger holds four incidents across four distinct classes (MISSING_CONTEXT, LOST_DECISION, BAD_OUTPUT, UNKNOWN), each recorded exactly once, so no class meets the two-or-more threshold and there is no repeated class for this diff to be clean or unclean of.
  - `test $(grep -c '^### ' docs/evals/failures.md) -ge 4` — rc=0. The four seeded incidents required by p3.w2.t1 are present.
  - `cd docs/evals/evidence/run-001 && sha256sum -c SHA256SUMS.txt --quiet && echo evidence_intact` — rc=0, "evidence_intact"; 326 files.
  - `awk '/^### [RH][0-9][0-9] /{n++} /^- split: regression$/{r++} /^- split: holdout$/{h++} END{print n,r,h; exit !(n==14 && r==10 && h==4)}' docs/evals/routing.md` — rc=0, "14 10 4". Case-set v1 is still frozen; this phase re-versioned no case.
  - `python3 -c "import json,glob,sys; f=sorted(glob.glob('docs/evals/evidence/run-001/led-*/oracle-score.json')); v=[json.load(open(p))['verdict'] for p in f]; print(len(f),'trials',v.count('pass'),'pass'); sys.exit(0 if len(f)==7 and v.count('pass')==7 else 1)"` — rc=0, "7 trials 7 pass". All six toggle trials and the append trial pass their respective oracles, scored from traces and after-states rather than prose.
  - scope: on target for the phase. The phase diff is confined to this phase's declared surfaces — `docs/decisions/0010-local-failure-ledger.md`, `docs/decisions/README.md`, `docs/evals/failures.md`, `docs/evals/runs.md`, `docs/evals/evidence/**`, `.claimignore` (only the obsolete failures.md exception removed), and this plan's lifecycle-owned sections. Nothing under `docs/playbooks/**`, `skills/**`, `cli/**`, `scripts/**`, `.github/**`, or personal settings was touched. The working tree additionally carries the already-closed F1 initiative (the audit document, the `audit-integrity-remediation` move to `docs/plans/completed/`, and the citation updates in ADRs 0008 and 0009); that is prior, separate work sitting uncommitted alongside this phase, not drift introduced here.
  - depth: standard. Blast radius is documentation and evidence only, but the phase makes two externally observable policy claims (ADR 0010's authority, and the eval record in runs.md), so plan alignment and evidence integrity were checked, not only the command gate.
  - plan alignment: p3.w1.t1, p3.w2.t1, p3.w2.t2 and p3.w2.t3 each have a Progress entry with its verification and its retries; the wave-2 summary is present. `expected_output` is met in full — accepted and indexed ADR, real ledger, obsolete exception removed, toggle and append observations in runs.md, evidence under `docs/evals/evidence/`, and this receipt naming the real ledger. S7 is demonstrated here rather than deferred.
  - required proof for lane `normal`: unit plus command output. Both classes captured — the Go unit suite via the third proof, command output via the rest.
  - retries: 2, both against this initiative's own disposable tooling and both retained beside their replacements — `score-r03.py` v1 mis-detected the mode announcement (`docs/evals/evidence/ledger-scorer-v1.txt`), and `runtrial-a.sh` v1's summariser aborted on an unbound variable under set -u (`docs/evals/evidence/ledger-append-runner-v1.txt`). No frozen case definition was changed and no model call was repeated to hide either defect.
  - enforcement: hook. `.git/hooks/pre-commit` is installed in this checkout and re-extracts its ZGUARD-CORE block per run, so the proofs above are re-executed at commit time and this verdict cannot be committed on stale evidence. `.github/workflows/docs-ci.yml` and `cli-ci.yml` re-run the doc-link, guard and Go commands in CI. Level `hook` on the ladder in `docs/patterns/encoding-invariants.md`, not merely authoring discipline.
  - receipt: context_sources: AGENTS.md, docs/WORKFLOW.md, docs/playbooks/work-full.md, docs/playbooks/check.md, docs/playbooks/check-validation.md, docs/plans/active/harness-eval-loop.md, docs/decisions/0010-local-failure-ledger.md, docs/evals/failures.md, docs/evals/routing.md, docs/evals/runs.md, docs/patterns/encoding-invariants.md / policy: docs/playbooks/check.md durable gate via docs/playbooks/work-full.md step 11 / judge: same-session / judge_model: claude-opus-5 / retries: 2 / rollback_point: 2013158 / failure_ledger: docs/evals/failures.md / enforcement: hook / not_independently_verified: see below
  - not_independently_verified: I authored every change this entry reviews, so the judge is `same-session` and this verdict is testimony about my own work. Three things specifically are not independently verified. (1) The R03 oracle scorer is mine; a second reader has not confirmed that `score-r03.py` v2 implements the five frozen `expect_pass` predicates faithfully, only that it runs and that its per-trial JSON matches what I read in the traces. (2) The judgement that the p3 diff stays inside this phase's declared surfaces is my own reading of my own diff. (3) The claim that the two seeded ledger rows drawn from the F1 closure (the stale pre-commit wrapper sequence and the `stashRestore` path escape) are accurately characterized rests on my earlier analysis of that initiative, which no one else has rechecked. The initiative's required independent `check full` has not run and is not claimed.
  - proof_gaps: none for lane `normal`. The complete Security, Performance, Architecture and Code Quality review is deliberately absent — check.md step 5 excludes it from `gate`, and work-full.md step 11 routes it to the one required independent `check full` on this final phase.
- 2026-09-16T08:20Z — phase p3-failure-ledger — APPROVE_WITH_REQUESTS — mode: full — the
  initiative's one required independent review (`docs/playbooks/check.md` steps 1-11, including
  step 5's complete Security, Performance, Architecture and Code Quality review that `gate` skips).
  - `bash scripts/verify-doc-links.sh` — rc=0, "doc links OK (0 findings; 10 claim(s) under known-removed v0.15 surfaces)".
  - `bash scripts/test-guards.sh` — rc=0, "guards: 40 passed, 0 failed".
  - `cd cli && CGO_ENABLED=0 go build ./... && go vet ./... && go test ./...` — rc=0; embedded, installer and interfaces packages ok.
  - `cd docs/evals/evidence/run-001 && sha256sum -c SHA256SUMS.txt --quiet && echo evidence_intact` — rc=0, "evidence_intact". Re-verified independently: 326 manifest lines and 326 files present under run-001, so no retained artifact is unlisted and none is missing.
  - `awk '/^- class:/{c[$3]++} END{n=0; for(k in c) if(c[k]>1) n++; print "repeated_classes=" n; exit (n>0)}' docs/evals/failures.md` — rc=0, "repeated_classes=0". Step 4 clean-of-class review: **explicitly none**. The real ledger holds four incidents in four distinct classes (MISSING_CONTEXT, LOST_DECISION, BAD_OUTPUT, UNKNOWN), each once, so no class reaches the two-or-more threshold.
  - `awk '/^### [RH][0-9][0-9] /{n++} /^- split: regression$/{r++} /^- split: holdout$/{h++} END{print n,r,h; exit !(n==14 && r==10 && h==4)}' docs/evals/routing.md` — rc=0, "14 10 4". Case-set v1 unchanged by this phase.
  - `test -f docs/decisions/0010-local-failure-ledger.md` — rc=0.
  - `test "$(grep -c '0010-local-failure-ledger' docs/decisions/README.md)" -eq 1` — rc=0; indexed exactly once.
  - `test -f docs/evals/failures.md` — rc=0.
  - `test -z "$(git status --porcelain -- docs/playbooks skills cli scripts .github)"` — rc=0. R10 holds: no playbook, skill, guard, installer, Go-test or workflow file is modified anywhere in the working tree.
  - `test -z "$(grep -rlEi 'sk-ant|ghp_|AKIA|BEGIN [A-Z ]*PRIVATE KEY' docs/evals/)"` — rc=0. Security sweep of the 5.2 MB of new evidence: no API key, token, or private key material. The 34 `apiKeySource` hits are CLI metadata, all valued "none"; no operator email appears.
  - `test "$(grep -n MISSING_CONTEXT docs/playbooks/work-full.md | cut -d: -f1)" = 28` — rc=0. Evidence for finding F1 below: the seven-token taxonomy lives at line 28, not the line 7 that four citations name.
  - `test "$(grep -n 'list every candidate' docs/playbooks/check.md | cut -d: -f1)" = 15` — rc=0. Evidence for finding F2 below: the total-plan-count rule and its quoted sentence live at line 15, not the line 31 the ledger's first row names.
  - `test "$(grep -c 'docs/plans/active/' docs/decisions/0010-local-failure-ledger.md)" -eq 2` — rc=0. Evidence for finding F3 below.
  - scope: on target. Writes stay inside `p3-failure-ledger`'s declared surfaces. The working tree also carries the already-closed F1 initiative (`docs/audit/2026-09-16-better-harness-eval-audit.md`, the `audit-integrity-remediation` move to `docs/plans/completed/`, and the matching citation repairs in ADRs 0008 and 0009); verified as prior separate work, not drift introduced here. `surfaces_avoided` is clean by command, not by reading.
  - depth: deep. Claims were re-derived from the raw retained artifacts rather than read from the run log's prose: all three `hc-old-*` `state-after.txt` snapshots show ` M docs/plans/active/greeting.md` and none of the three `hc-new-*` do, which is the observable basis for the `reproduced` conclusion; `led-append-1/after-greeting.md.txt` carries phase `status: in-progress` and `lifecycle_status: in-progress`; all seven `led-*/oracle-score.json` verdicts read `pass`; the appended rows in `led-append-1/after-failures.md.txt` are labelled `[SYNTHETIC FIXTURE]` and appear nowhere in the real ledger. Every `source:` field in `docs/evals/failures.md` was resolved: `90f5b84`, `b4c1ef0`, `e75b977` and `2013158` all exist and their cited paths resolve at those commits, so the ADR's immutable-source contract holds for all four rows.
  - findings (major, non-blocking — the requests of this verdict):
    - F1 — `docs/playbooks/work-full.md:7` is a wrong anchor and is used four times: `docs/evals/failures.md` Conventions, and ADR 0010's Context, Decision and Authority. Line 7 is a `## Owned Plan State` bullet; the taxonomy is at line 28 (step 7). The synthetic fixture ledger written for `p3.w2.t3` uses the correct, rot-proof form ("`docs/playbooks/work-full.md` step 7"), so the real artifacts are the ones that are wrong.
    - F2 — `docs/evals/failures.md` row 1 cites `docs/playbooks/check.md:31` (step 2) as the coverage for the two-plan routing miss. Line 31 is step 6; step 2 and the sentence the row quotes are at line 15. A reader following the anchor lands on unrelated text and cannot confirm the coverage claim.
    - F1 and F2 share one cause worth stating: line-number anchors into live, frequently edited playbooks rot exactly the way the active-plan paths repaired in ADRs 0008 and 0009 do, and two of five such anchors were already wrong at the moment of authoring. The narrative "(step N)" in each citation is the durable part; the ":NN" is not.
  - findings (minor):
    - F3 — ADR 0010 cites `docs/plans/active/harness-eval-loop.md` twice (Status, Authority). That is the citation form this same diff repairs in ADRs 0008 and 0009, and it guarantees a follow-up edit when this plan moves to `docs/plans/completed/`. Not a ledger row, so the ADR's own admissibility rule is not violated, but the rot is identical.
    - F4 — retained evidence embeds the operator's absolute paths: 81 occurrences of `/home/tinhpt/.claude/projects/...` inside `trace.jsonl` files and 10 of `/home/tinhpt/Lab/mono-harness` in the retained generator scripts. S3's "portable redacted artifact paths" is met for the paths `runs.md` references (all repo-relative), but the raw traces carry the operator's home layout into a repository distributed to consumers. No credential material accompanies them.
  - plan alignment: `expected_output` for `p3-failure-ledger` is met in full — ADR accepted and indexed, real ledger with four sourced incidents, obsolete `.claimignore` exception removed, toggle and append observations in `runs.md`, evidence retained, and this receipt naming the real ledger. S6, S7 and S8 are demonstrated here rather than deferred. R7, R8, R9 and R10 all hold. No finding above contradicts the plan.
  - required proof for lane `normal`: unit plus command output — both captured. Manual review is supplied by this entry itself.
  - ledger append: not performed. `docs/playbooks/check.md:37` conditions the append on `REQUEST_CHANGES`; this verdict is `APPROVE_WITH_REQUESTS`, so F1-F4 are carried as open items, consistent with the "Candidates not yet seeded" note the ledger already records for the same reason.
  - judge: independent
  - judge_model: claude-opus-5
  - retries: 0
  - enforcement: hook. `.git/hooks/pre-commit` is installed in this checkout and re-extracts its ZGUARD-CORE block per run, so the fourteen proofs above re-execute at commit time; `.github/workflows/docs-ci.yml` and `cli-ci.yml` re-run the doc-link, guard and Go commands in CI. Level `hook` on the ladder in `docs/patterns/encoding-invariants.md`.
  - proof_gaps: none for the required proof classes, but three limits are named rather than implied. (1) The 14 baseline verdicts and the six historical-control verdicts still rest on manual reading of ordered `tool_use` streams; this review re-derived the historical control's state deltas and all seven ledger-trial oracle scores from artifacts, not the ten regression verdicts one by one. (2) `judge: independent` here means a separate session that authored none of this diff and read only the repository; nothing external establishes that claim, per "What the Guards Cannot Check". (3) n=3 per side on the ledger toggle supports `no-detected-regression` only, which is what `runs.md` already claims and does not overstate.
  - receipt: context_sources: docs/playbooks/check.md, docs/playbooks/check-validation.md, docs/playbooks/work-full.md, docs/playbooks/work.md, docs/plans/active/harness-eval-loop.md, docs/decisions/0010-local-failure-ledger.md, docs/decisions/README.md, docs/evals/failures.md, docs/evals/routing.md, docs/evals/runs.md, docs/evals/evidence/run-001/**, .claimignore / policy: docs/playbooks/check.md mode full, the initiative's one required independent review per p3-failure-ledger check / judge: independent / judge_model: claude-opus-5 / retries: 0 / rollback_point: 2013158 / failure_ledger: docs/evals/failures.md / enforcement: hook / not_independently_verified: n/a for authorship — this session authored none of the diff under review; the residual limits are listed in proof_gaps above

<!-- Durable check evidence. -->
