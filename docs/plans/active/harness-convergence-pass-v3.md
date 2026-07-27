---
id: 01KYF4NZQ49M3SK6EF68CY1SEA
type: plan
intake_id: 01KYF4PHXC6GWRMBASYR06Q5B5
lane: high-risk
status: active
created: 2026-07-26
updated: 2026-07-27
---

# Plan: Harness Convergence Pass v3 — docs-first, one DB, mandatory CLI guard

## Outcome
- result: Repository docs, code, tests, and observable runtime behavior are authoritative while `zharness` provides mandatory stage preflight, one replayable root database, typed durable-lifecycle state, and mechanical rail guards without parallel lifecycle markdown.
- actors: repository owner, agents invoking the active workflow stages, and consumer repositories initialized by the released CLI.
- success_signals:
  - Every active workflow stage uses deterministic CLI preflight and honest reduced/blocked routing.
  - A fresh or migrated repository uses root `harness.db`, tracked `.kit/changesets`, safe root managed docs, and one evolving plan per durable initiative.
  - Durable execution reaches `planned -> in-progress -> checked -> done` through public commands, while bounded work performs zero lifecycle writes.
  - Resume, validation, audit, recap, and Git workflows operate without dedicated lifecycle markdown evidence files.

## Authority and Requirements
- authority:
  - Locked Harness Convergence Pass v3 approved on 2026-07-26.
  - Repository instructions and public CLI contracts.
  - Existing code, tests, changeset replay, and observable command behavior.
  - Upstream repository-harness semantics audited at commit `0a79bbebf02d70a281d01fd060d0fd0af125e027`, adopted selectively without weakening the local mandatory CLI guard.
- requirements:
  - R1 [accepted]: Keep `.kit` and exactly one writable root `harness.db`; gitignore its DB/WAL/SHM files and retain replayable tracked deltas at `.kit/changesets/*.changeset.jsonl`.
  - R2 [accepted]: Treat repository docs, code, tests, and runtime behavior as intent/proof authority; SQLite remains an execution ledger and recovery index.
  - R3 [accepted]: Provide read-only `zharness preflight <stage> [--mode <mode>] --json` routing for all active workflow stages.
  - R4 [accepted]: Keep the CLI mandatory, permit reduced docs-first work when appropriate, and require initialized DB state for durable mutations.
  - R5 [accepted]: Use `docs/plans/active/{slug}.md`, moved to `docs/plans/completed/{slug}.md` after final validation, as the only durable initiative markdown.
  - R6 [accepted]: Retain typed meta/intake/story/run/trace/check/handoff/intervention rows without requiring separate lifecycle artifact paths.
  - R7 [accepted]: Public run creation, clean check recording, and closing handoff must advance stories `planned -> in-progress -> checked -> done`.
  - R8 [accepted]: Project root managed docs safely, preserve consumer-owned AGENTS bytes, store installed hashes/version in `managed_docs`, and stage conflicts instead of overwriting local edits.
  - R9 [accepted]: Provide replay/parity/rollback-safe `zharness migrate layout --to v2 [--dry-run]` without rewriting historical changesets.
  - R10 [accepted]: Summarize legacy workflow history before removing obsolete tracked lifecycle markdown with `trash`; Git history remains raw provenance.
  - R11 [accepted]: Encode authority-before-edit, independent work-shape/human-judgment decisions, configurable-default limits, behavior-appropriate evidence, honest proof gaps, and durable decisions only when future work needs them.
  - R12 [accepted]: Keep the managed entrypoint plus workflow map within 1,000 words and keep stage playbooks stage-specific.
  - R13 [accepted]: Bounded/simple work invokes preflight but creates no intake, story, run, trace, check, handoff, plan, report, changeset, or markdown artifact.
  - R14 [accepted]: Avoid speculative infrastructure and preserve public behavior outside approved contract changes.

## Non-goals
- NG1: No `.harness/`, second database, baseline snapshot database, generic event-store conversion, or changeset compaction.
- NG2: No full lifecycle state for read-only or bounded work.
- NG3: No new workflow skill, service, runtime, external integration, credential, or MCP dependency.
- NG4: No CI workflow modification, self-updater, automatic three-way merge system, or invented consumer application runbook.
- NG5: No typed lifecycle table collapse or rewriting of existing changesets.
- NG6: Phase 3 does not perform the Phase 4 global docs-first mental-model rewrite.

## Approach and Risks
- approach: Deliver four independently usable phases in dependency order: universal preflight; root layout and managed docs; one-plan lifecycle; concise docs-first contract. Keep all durable mutations changeset-first and use existing public lifecycle commands rather than new state paths.
- chosen decisions:
  - Root DB plus `.kit/changesets` plus root docs, not a new namespace.
  - Exactly one database, not dual state generations.
  - Mandatory CLI with conditional DB use, not universal ceremony or optional rails.
  - One evolving plan with typed DB rows, not DB-only work or parallel markdown records.
  - Managed-doc hashes in DB, not another manifest file.
- constraints:
  - Existing changesets replay in ULID order without modification.
  - Migration failure leaves active DB/docs unchanged.
  - Consumer-owned content outside managed blocks is byte-preserved.
  - Root docs refresh does not overwrite local changes without explicit force.
  - Each phase leaves a usable, testable harness if later phases never ship.
  - Legacy deletion uses `trash` only after coverage and replay proof.
- risks:
  - RISK-1: Read-only or bounded routes accidentally mutate lifecycle state | mitigation: checksum/count tests | recovery: stop and restore zero-write routing before proceeding.
  - RISK-2: Managed docs overwrite consumer edits | mitigation: per-file hashes and conflict staging | recovery: leave active files unchanged and surface recovery.
  - RISK-3: Lifecycle readers retain hidden file dependencies | mitigation: real-binary no-artifact fixture plus validate/audit proof | recovery: block legacy removal until DB-only readers pass.
  - RISK-4: One plan loses current requirements or proof | mitigation: preserve the nine-section contract and append-only Progress/Decisions/Validation | recovery: stop consolidation when coverage cannot be demonstrated.
  - RISK-5: Phase 3 drifts into the Phase 4 rewrite | mitigation: change only stage-specific one-plan procedures/tests | recovery: defer global authority wording and public mental-model cleanup to `docs-first-contract`.

## Phases and Verification
<!-- Phase and task definitions are immutable. Only phase lifecycle status changes to mirror DB transitions. Append-only Progress is the sole task execution-status source; task definitions contain no status fields. -->
### Phase 1: universal-preflight
- phase_slug: universal-preflight
- story_id: 01KYF4Y6B2AS5ZWPWV5Q4G3199
- status: done
- goal: Give every active workflow stage deterministic read-only CLI routing while the prior layout remains operational.
- depends_on: none
- waves:
  - wave: 1-3
    tasks:
      - task: implement the preflight matrix, zero-write proof, and stage adapters
    checks:
      - command: go -C cli test ./...
        expected: all packages pass with preflight routing and no-write coverage
- closure: approved by check `01KYF71N4RMWYGTSJTD4BWC740`.

### Phase 2: root-layout
- phase_slug: root-layout
- story_id: 01KYF4Y6B9PRQ1TYHN382ANKZT
- status: done
- goal: Activate one root database and safe root managed docs while preserving replayable changesets.
- depends_on: universal-preflight
- waves:
  - wave: 1-4
    tasks:
      - task: root DB, managed docs, migration/parity/rollback, and dogfood activation
    checks:
      - command: go -C cli test ./...
        expected: all packages pass with migration and managed-doc proof
- closure: approved by check `01KYF9NC8RXVCNYM8M56YQX0RF`.

### Phase 3: one-plan-lifecycle
- phase_slug: one-plan-lifecycle
- story_id: 01KYF4Y6BGNWEZVY0B05KYE3DV
- status: done
- goal: Replace parallel lifecycle markdown with one evolving plan while retaining typed DB lifecycle guards.
- depends_on: root-layout
- waves:
  - wave: 1
    parallel: false
    tasks:
      - task: T1 add one-plan schema and scaffold contract
    checks:
      - command: go -C cli test ./... -run 'Plan|Scaffold|Replay|Migration'
        expected: pass
  - wave: 2
    parallel: false
    tasks:
      - task: T2 remove markdown dependencies from lifecycle commands
    checks:
      - command: go -C cli test ./... -run 'Run|Check|Handoff|Status|Resume|Validate|Bounded'
        expected: pass
  - wave: 3
    parallel: false
    tasks:
      - task: T3 rewrite six stage playbooks around one plan and add proof
        touches: [six embedded playbooks, byte-identical root mirrors, plan template, lifecycle/projection/scaffold tests, this active plan]
        avoids: [production schema/domain/application behavior, global Phase 4 rewrite, legacy deletion]
    checks:
      - command: go -C cli test ./... -run 'Lifecycle|Projection|OnePlan'
        expected: pass
      - command: git diff --check
        expected: pass
  - wave: 4
    parallel: false
    tasks:
      - task: T4 consolidate legacy history and remove obsolete lifecycle markdown after proof
    checks:
      - command: git ls-files '.kit/**' && test "$(git ls-files '.kit/**' | grep -vc '^.kit/changesets/')" -eq 0 && go -C cli test ./...
        expected: tracked `.kit` content is changesets only and all tests pass

### Phase 4: docs-first-contract
- phase_slug: docs-first-contract
- story_id: 01KYF4Y6BPVC5585GZ56GFTB6D
- status: planned
- goal: Finalize concise docs-first authority and workflow contracts after one-plan lifecycle closure.
- depends_on: one-plan-lifecycle
- waves:
  - wave: pending planning refresh after Phase 3 gate
    tasks:
      - task: rewrite managed entrypoint/workflow/public contracts and run final verification
    checks:
      - command: go -C cli test ./...
        expected: all packages pass with live-command and word-budget proof

## Progress
- 2026-07-26T12:40:22Z phase=universal-preflight wave=1-3 task=T1-T3 task_status=DONE run_id=01KYF4Z95MQJG7PEZPF00JKC9H trace_id=01KYF6CPWHRYHF1EAAJMTRF10F verification=go -C cli test ./... -> pass result=universal preflight implemented and approved by check 01KYF71N4RMWYGTSJTD4BWC740
- 2026-07-26T13:26:05Z phase=root-layout wave=1-4 task=T1-T4 task_status=DONE run_id=01KYF75AMDJ8VB806YXW2H8ZPQ trace_id=01KYF9113K0MY35JC0B8BSBDQB verification=go -C cli test ./... -> pass result=root DB, managed docs, migration parity, rollback, and dogfood activation approved by check 01KYF9NC8RXVCNYM8M56YQX0RF
- 2026-07-27T02:48:00Z phase=one-plan-lifecycle wave=1 task=T1 task_status=DONE run_id=01KYF9TECM63NSPBEKNTNYFJXE trace_id=01KYFACVFSJHSANYWKGR4PPQT7 verification=go -C cli test ./... -run 'Plan|Scaffold|Replay|Migration' -> pass result=one-plan schema/scaffold contract complete
- 2026-07-27T02:48:00Z phase=one-plan-lifecycle wave=2 task=T2 task_status=DONE run_id=01KYF9TECM63NSPBEKNTNYFJXE trace_id=01KYGP9S4X3QSBWHXN12VPDTD4 verification=go -C cli test ./... -> pass result=DB-only lifecycle readers and public status transitions approved by check 01KYGQK2NVCY4FT34AF7AWTVHA
- 2026-07-27T03:08:39Z phase=one-plan-lifecycle wave=3 task=T3 task_status=in-progress run_id=01KYGRQGV0ZS281GCB7ZP2EXXY trace_id=none verification=go -C cli test ./... -run 'Lifecycle|Projection|OnePlan' -> not-run result=implementation active; Wave 3 verification and gate remain pending
- 2026-07-27T04:10:00Z phase=one-plan-lifecycle wave=3 task=T3 task_status=DONE run_id=01KYGRQGV0ZS281GCB7ZP2EXXY trace_id=01KYGWA9M8GCHBAD0EB8XT0ZPQ verification=go -C cli test ./... -> pass result=six one-plan stage playbooks, projection identity, bounded/review zero-write behavior, and two-dependent-phase lifecycle proof approved by semantic review
- 2026-07-27T04:19:00Z phase=one-plan-lifecycle wave=4 task=T4 task_status=in-progress run_id=01KYGRQGV0ZS281GCB7ZP2EXXY trace_id=none verification=73 changesets replayed; query state and validate -> pass result=history summary coverage APPROVED and scratch root DB reconstructed without legacy lifecycle markdown; expected pre-gate out_of_order drift remains until the latest run is checked
- 2026-07-27T04:24:00Z phase=one-plan-lifecycle wave=4 task=T4 task_status=DONE run_id=01KYGRQGV0ZS281GCB7ZP2EXXY trace_id=01KYGX2X47TV7ZTG252JXVNCSG verification=tracked `.kit` tree changesets-only; go -C cli test ./... -> pass result=legacy history consolidated; approved planning/plans/runs/reports/HANDOFF/docs projections moved to Trash; 74 changesets and five provenance docs preserved
- 2026-07-27T05:21:09Z phase=one-plan-lifecycle wave=gate-remediation task=F1-F4 task_status=DONE run_id=01KYGRQGV0ZS281GCB7ZP2EXXY trace_id=01KYH0A4NJVETX0N7KYQHWVA6W changed_surfaces=active-plan routing, read-only inspection commands, lifecycle transition guards, strict ULID validation verification=targeted application/CLI/replay/migration suites and independent reviews -> pass result=all four findings from check 01KYGXK5BJ5YASGBRK8CGDNBX1 resolved; full Phase 3 re-gate pending
- 2026-07-27T07:48:17Z phase=one-plan-lifecycle wave=gate-remediation-2 task=F5-F8 task_status=DONE run_id=01KYGRQGV0ZS281GCB7ZP2EXXY trace_id=01KYH8QHPCGK1VFPWB1EBH6MW2 changed_surfaces=audit fixtures, persisted enum integrity, WAL-safe shared readers, serialized mutations and tracked recovery verification=full Go suite, race tests, vet, Linux/Darwin builds, 76-change scratch replay parity, and independent reviews -> pass result=all remaining Phase 3 gate blockers resolved; durable final gate pending

## Decisions
- 2026-07-26T12:03:31Z phase_task=initiative/lock decision=Use one root DB, tracked `.kit/changesets`, root managed docs, and one evolving plan rationale=This preserves replay and repository legibility while removing duplicate lifecycle state.
- 2026-07-26T13:10:34Z phase_task=root-layout/T3 decision=Keep legacy lifecycle markdown until Phase 3 DB-only validation and history coverage pass rationale=Early deletion would remove the current recovery/proof source before its replacement is demonstrated.
- 2026-07-27T02:48:00Z phase_task=one-plan-lifecycle/T2 decision=Preserve public typed lifecycle commands and make artifact paths optional/deprecated rationale=Status transitions and queryability remain useful while dedicated markdown records are redundant.
- 2026-07-27T03:08:39Z phase_task=one-plan-lifecycle/T3 decision=Limit Wave 3 to six stage procedures, mirrors, template, tests, and this bootstrap plan rationale=The global authority/work-shape rewrite belongs to Phase 4 and legacy removal belongs to Wave 4.

## Validation
- 2026-07-26T12:40:22Z phase=universal-preflight command=go -C cli test ./... result=pass output=all packages passed after routing fixes run_id=01KYF4Z95MQJG7PEZPF00JKC9H check_id=01KYF71N4RMWYGTSJTD4BWC740 verdict=APPROVED proof_gaps=none
- 2026-07-26T13:26:05Z phase=root-layout command=go -C cli test ./... result=pass output=all packages passed with migration and managed-doc coverage run_id=01KYF75AMDJ8VB806YXW2H8ZPQ check_id=01KYF9NC8RXVCNYM8M56YQX0RF verdict=APPROVED proof_gaps=none
- 2026-07-27T02:48:00Z phase=one-plan-lifecycle command=go -C cli test ./... result=pass output=all packages passed after Waves 1-2 review fixes run_id=01KYF9TECM63NSPBEKNTNYFJXE check_id=01KYGQK2NVCY4FT34AF7AWTVHA verdict=APPROVED proof_gaps=none for Waves 1-2
- 2026-07-27T03:08:39Z phase=one-plan-lifecycle command=go -C cli test ./... -run 'Lifecycle|Projection|OnePlan' result=not-run output=Wave 3 implementation/verification pending run_id=01KYGRQGV0ZS281GCB7ZP2EXXY check_id=none verdict=not-recorded proof_gaps=Wave 3 unit/integration/real-binary and projection evidence
- 2026-07-27T04:10:00Z phase=one-plan-lifecycle command=go -C cli test ./... result=pass output=targeted expanded suite and full suite passed; semantic review APPROVED run_id=01KYGRQGV0ZS281GCB7ZP2EXXY check_id=none verdict=wave-complete proof_gaps=full Phase 3 gate remains after Wave 4
- 2026-07-27T04:19:00Z phase=one-plan-lifecycle command=scratch replay/query state/resume/validate result=pass output=73 changesets reconstructed schema 5; validate valid=true with zero findings; resume preserved current phase/run and reported only expected pre-gate latest-check drift run_id=01KYGRQGV0ZS281GCB7ZP2EXXY check_id=none verdict=pre-deletion-proof proof_gaps=post-trash tracked-tree and full-suite verification remain
- 2026-07-27T04:24:00Z phase=one-plan-lifecycle command=tracked-tree assertion plus go -C cli test ./... result=pass output=tracked `.kit` entries are changesets only; active/completed plans and five provenance docs retained; all Go packages passed after cleanup run_id=01KYGRQGV0ZS281GCB7ZP2EXXY check_id=none verdict=wave-complete proof_gaps=full Phase 3 gate and latest-run check remain
- 2026-07-27T04:31:00Z phase=one-plan-lifecycle command=full Security/Performance/Architecture/Code Quality review result=fail output=four major findings: legacy planning routing, writable reader opens, non-monotonic/stale-run lifecycle closure, and incomplete strict ULID coverage run_id=01KYGRQGV0ZS281GCB7ZP2EXXY check_id=01KYGXK5BJ5YASGBRK8CGDNBX1 verdict=REQUEST_CHANGES proof_gaps=regression tests for active-plan routing, read-only journal behavior, lifecycle monotonic/latest-run guards, and retained-entity/plan-id ULIDs
- 2026-07-27T07:52:31.148Z phase=one-plan-lifecycle command=full Go suite, race suite, vet, Linux/Darwin builds, diff checks, 78-change scratch replay/resume/validate/audit, and full Security/Performance/Architecture/Code Quality review result=pass output=all packages and race coverage passed; root and scratch lifecycle state clean with pending=0, valid=true, audit empty; tracked recovery identity blocker fixed and independently APPROVED run_id=01KYGRQGV0ZS281GCB7ZP2EXXY check_id=01KYH8Z9NCBWNNQ2WCJ9AXS8P3 verdict=APPROVED proof_gaps=none

## Current State and Next Action
- active_phase: one-plan-lifecycle
- lifecycle_status: done
- latest_run_id: 01KYGRQGV0ZS281GCB7ZP2EXXY
- latest_trace_ids: [01KYGWA9M8GCHBAD0EB8XT0ZPQ, 01KYGX2X47TV7ZTG252JXVNCSG, 01KYH0A4NJVETX0N7KYQHWVA6W, 01KYH8QHPCGK1VFPWB1EBH6MW2]
- latest_check_id: 01KYH8Z9NCBWNNQ2WCJ9AXS8P3
- latest_handoff_id: 01KYH91W0BE3D66PZYN7J4GW91
- completed: Phase 1 universal-preflight, Phase 2 root-layout, and Phase 3 one-plan-lifecycle are done; Phase 3 closed with clean check `01KYH8Z9NCBWNNQ2WCJ9AXS8P3` and handoff `01KYH91W0BE3D66PZYN7J4GW91` after full replay, audit, concurrency, and deep-review proof.
- blockers: []
- open_items: [Phase 4 docs-first-contract remains planned and has not started]
- exact_next_action: Run `work full phase docs-first-contract` when Phase 4 execution is explicitly started; do not perform Phase 4 work as part of this Phase 3 closure.
