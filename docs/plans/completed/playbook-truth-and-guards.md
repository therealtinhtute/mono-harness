---
id: 01M0PLAYBOOKTRUTH9K3XJ7
type: plan
intake_id: 01M0PLAYBOOKTRUTHINTK9K3XJ7
lane: normal
status: completed
created: 2026-08-30
updated: 2026-08-30
---

# Plan: playbook-truth-and-guards — map matches v0.15; two invariants dual-encoded

## Outcome
- result: live playbooks and the workflow README describe the v0.15 system (markdown + hook, no lifecycle CLI); “at most one active plan” is enforced by the pre-commit/CI guard; durable Validation no longer asks for a ULID check id and carries a compact git-native receipt. A stronger model following the map cannot reconstruct a deleted control plane.
- success_signals:
  - S1: `rg -n "lifecycle ledger|DB-mirroring|mirrored check row|check_id: ULID|latest_run_id" cli/docs/embedded/playbooks skills/workflow/README.md` prints nothing.
  - S2: `docs/playbooks/*.md` is byte-identical to `cli/docs/embedded/playbooks/*.md`.
  - S3: `bash scripts/test-guards.sh` reports all passed, including a new fixture that rejects two non-empty files under `docs/plans/active/` and accepts zero or one.
  - S4: `cd cli && go test ./...` passes; `embedded_test.go` forbids the S1 substrings.
  - S5: `bash scripts/verify-doc-links.sh` reports 0 findings.
  - S6: ADR `docs/decisions/0006-v015-authority.md` exists and the decisions index lists it; ADR 0001–0003 and 0005 bodies are not rewritten.

## Authority and Requirements
- authority:
  - Owner selection, 2026-08-30: close `macos-guard-portability`, then lock one narrow initiative covering H1 + H2 + H4 only. Not a Level-3 platform rewrite.
  - `docs/audit/harness-engineering-gap-audit.md` — H1 (map residue), H2 (one-plan prose-only), H4 (receipt / ULID ghost).
  - `docs/ARCHITECTURE.md`, `cli/docs/CONTRACT.md`, `docs/PROJECT.md` — v0.15 slim: three verbs, markdown is the record, two fail-closed hook guards.
  - `docs/decisions/0002-single-active-plan-resolver.md` — historical: the invariant used to be `ResolveActivePlan`; that file is gone.
  - rari, *Harness Engineering* (X article, 2026-08-29): encode important rules twice; map not a giant manual; change receipt without hiding a broken process.
- requirements:
  - R1 [accepted]: Embedded playbooks under `cli/docs/embedded/playbooks/` contain none of: `lifecycle ledger`, `DB-mirroring`, `harness.db`, `latest_run_id`, `mirrored check row`, `check_id: ULID`. Projected `docs/playbooks/` matches byte-for-byte. | source: audit H1; `cli/internal/embedded/embedded_test.go` kill-list pattern
  - R2 [accepted]: `skills/workflow/README.md` no longer describes a “harness-backed runtime”, no longer says the change is “a durable, machine-recorded, replayable trail instead of relying on markdown pointers alone”, and does not present `zharness init && zharness import && zharness query state` as current GO evidence. | source: audit H1; `docs/ARCHITECTURE.md`
  - R3 [accepted]: New ADR 0006 states that v0.15 deleted the derived index and the lifecycle CLI; ADRs 0001–0003 and 0005 remain historical records; current authority is `docs/ARCHITECTURE.md` + `cli/docs/CONTRACT.md`. Those four ADR bodies are not rewritten. The decisions index gains one row. | source: audit H1 action 3; `docs/decisions/README.md`
  - R4 [accepted]: `cli/internal/embedded/embedded_test.go` forbids the R1 substrings (and `zharness memory`). `cd cli && go test ./...` passes. | source: audit H1 action 4
  - R5 [accepted]: The ZGUARD-CORE block (or a sibling function extracted the same way) rejects a commit/CI push when more than one non-empty `docs/plans/active/*.md` exists; zero or one is accepted; the failure names both paths and says to `git mv` the finished file to `docs/plans/completed/`. `scripts/test-guards.sh` covers reject-two and accept-one. Existing proof-reexec and high-risk-judge semantics are unchanged (macos-guard-portability NG1). | source: audit H2; `scripts/install-git-hooks.sh`
  - R6 [accepted]: Durable `check` playbook schema has no ULID issuer (`check_id: ULID`, “record the returned check ID if one was issued”, “mirrored check row”). A durable Validation entry is specified to include a grep-able `receipt:` block with `context_sources`, `policy`, `judge`, `judge_model`, `retries`, `rollback_point`, `not_independently_verified`. No `cost_usd` field. Git commit SHA remains the artifact id. | source: audit H4; article receipt schema

## Non-goals
- NG1: No lifecycle CLI verbs (`preflight`, `query`, `trace add`, `memory`, `audit`, `validate`, `check record`, SQLite).
- NG2: No sandbox, no host tool-permission table enforcement, no scanning `~/.claude` / `~/.codex`.
- NG3: No GRAPH/coordination layer; no `cost_usd` in Validation.
- NG4: No change to the high-risk independent-judge rule or proof re-execution semantics.
- NG5: H3 (final-phase `full` cannot be `same-session`), H5/H9 (failure ledger / class), H6 (named gap on BLOCKED), H7 (`plan-slice.sh`), H8 (ARCHITECTURE permission paragraph) are deferred — not this initiative.
- NG6: Do not rewrite `docs/audit/*.md` bodies; they are dated records. Cite `harness-engineering-gap-audit.md`, do not edit it to “stay current.”
- NG7: Do not rewrite ADR 0001–0005 bodies.
- NG8: Do not touch consumer repositories.

## Approach and Risks
- approach: edit the live map first (embedded playbooks, then project; workflow README; CONTRACT until-then lie; plan template ghost IDs), then lock the kill-list and ADR 0006 so residue cannot return, then add the one-plan count to ZGUARD-CORE with fixtures and a CI call. No new CLI verbs. Rejected: rewriting dated audits/ADRs; resurrecting ResolveActivePlan; a JSONL control plane.
- constraints:
  - Edit `cli/docs/embedded/playbooks/` as source; `docs/playbooks/` must stay a byte copy.
  - ZGUARD-CORE proof-reexec and high-risk-judge semantics unchanged (NG4).
  - ADR 0001–0005 bodies untouched (NG7).
- risks:
  - Kill-list substrings that appear in historical sentences inside playbooks. Mitigation: grep the six playbooks before expanding retired[]; stop and narrow the string rather than rewrite history into the playbook.
  - One-plan guard false-positive on an empty `docs/plans/active/*.md` glob (zero files). Mitigation: zero is valid idle; only `n > 1` rejects.
- recovery: revert the hook commit independently; playbook/README/ADR 0006 are markdown-only reverts.

## Phases and Verification
<!-- Phase and task definitions are immutable after to-plan. Do not add task status fields. Append-only Progress is the sole task execution-status source. -->
- planning_status: planned
- phases:
  - phase_slug: p1-playbook-truth
    story_id: 01M0PLAYBOOKTRUTHPH19K3XJ7
    status: done
    goal: Live playbooks, workflow README, CONTRACT, and the plan template describe v0.15; Validation uses a git-native receipt, not a ULID issuer.
    depends_on: none
    surfaces_allowed: cli/docs/embedded/playbooks/**, docs/playbooks/**, skills/workflow/README.md, cli/docs/CONTRACT.md, cli/docs/embedded/templates/plan.md
    surfaces_avoided: docs/audit/**, docs/decisions/0001*.md, docs/decisions/0002*.md, docs/decisions/0003*.md, docs/decisions/0004*.md, docs/decisions/0005*.md, cli/cmd/**, cli/internal/installer/**
    requirements: R1, R2, R6
    waves:
      - wave: 1
        goal: Strip 0.14 ghosts and add the receipt block.
        tasks:
          - task: Strip residue from embedded playbooks (watzup, work, to-plan, check) and project copies; add receipt fields to check.md.
            verify: `rg -n "lifecycle ledger|DB-mirroring|mirrored check row|check_id: ULID|latest_run_id" cli/docs/embedded/playbooks docs/playbooks` prints nothing.
          - task: Rewrite `skills/workflow/README.md` to the v0.15 model; delete the 2026-07-17 GO block that cites deleted CLI verbs.
            verify: `rg -n "harness-backed runtime|zharness init && zharness import|machine-recorded, replayable trail" skills/workflow/README.md` prints nothing.
          - task: Fix `cli/docs/CONTRACT.md` so it no longer says the binary registers no subcommands; drop ULID ghost fields from `cli/docs/embedded/templates/plan.md`.
            verify: `rg -n "registers no subcommands|latest_run_id|latest_check_id" cli/docs/CONTRACT.md cli/docs/embedded/templates/plan.md` prints nothing.
        stop_condition: a required playbook phrase in `embedded_test.go` disappears; stop and restore it rather than weakening the test.
    checks:
      - `diff -q cli/docs/embedded/playbooks docs/playbooks` is silent (byte-identical).
      - `bash scripts/verify-doc-links.sh` reports 0 findings.
    escalation: if a consumer-facing file outside surfaces_allowed still lies (e.g. assets SVG), record it in Decisions and leave it; do not widen this phase.

  - phase_slug: p2-authority-kill-list
    story_id: 01M0PLAYBOOKTRUTHPH29K3XJ7
    status: done
    goal: ADR 0006 is the current-authority overlay; embedded_test forbids the residue strings.
    depends_on: p1-playbook-truth
    surfaces_allowed: docs/decisions/0006-v015-authority.md, docs/decisions/README.md, cli/internal/embedded/embedded_test.go
    surfaces_avoided: docs/decisions/0001*.md, docs/decisions/0002*.md, docs/decisions/0003*.md, docs/decisions/0004*.md, docs/decisions/0005*.md, cli/docs/embedded/playbooks/**
    requirements: R3, R4
    waves:
      - wave: 1
        goal: Record authority and lock the kill list.
        tasks:
          - task: Add ADR 0006 and a row in the decisions index; do not rewrite 0001–0005.
            verify: `test -f docs/decisions/0006-v015-authority.md` and `rg -n "0006-v015-authority" docs/decisions/README.md` matches.
          - task: Extend `retired` in `embedded_test.go` with the R1 substrings.
            verify: `cd cli && go test ./...`
        stop_condition: go test fails because a playbook still contains a new retired string — return to p1, do not delete the assertion.
    checks:
      - `cd cli && go test ./...` passes.
    escalation: none; this phase is markdown plus one test file.

  - phase_slug: p3-one-plan-guard
    story_id: 01M0PLAYBOOKTRUTHPH39K3XJ7
    status: done
    goal: At most one non-empty active plan is a hook/CI check. Proof and judge guards stay unchanged.
    depends_on: p2-authority-kill-list
    surfaces_allowed: scripts/install-git-hooks.sh, scripts/test-guards.sh, .github/workflows/cli-ci.yml
    surfaces_avoided: cli/cmd/**, cli/internal/installer/**, docs/playbooks/**
    requirements: R5
    waves:
      - wave: 1
        goal: Count active plans in ZGUARD-CORE, fixture it, call it from pre-commit and CI.
        tasks:
          - task: Add `zharness_guard_at_most_one_active_plan` inside ZGUARD-CORE; reject n>1 naming paths; accept 0 or 1; wire pre-commit and cli-ci.yml.
            verify: `bash scripts/test-guards.sh` reports all passed including reject-two and accept-one/zero.
          - task: Mutation-check that proof-reexec and high-risk-judge fixtures still pass (NG4).
            verify: `bash scripts/test-guards.sh` still includes S1–S5 ok lines.
        stop_condition: any existing S1–S5 fixture fails; revert the count function rather than rewrite proof semantics.
    checks:
      - `bash scripts/test-guards.sh` all passed, 0 failed.
    escalation: if macOS bash 3.2 cannot count files without `local -A`, use a file list like the hash-set fix; do not introduce bash 4 syntax.

## Progress
- [2026-08-30] (p1-playbook-truth, wave 1): task_status=DONE. Stripped 0.14 ghosts from embedded playbooks (work/watzup/to-plan/check), projected to docs/playbooks, rewrote skills/workflow/README.md, fixed CONTRACT.md until-then lie, dropped ULID fields from plan template. `rg` on residue strings: empty. `diff -q` playbooks: identical.
- [2026-08-30] (p2-authority-kill-list, wave 1): task_status=DONE. Added ADR 0006 and index row. Extended embedded_test.go retired[] with R1 substrings. Updated one required to-plan phrase in the test to the new v0.15 wording (the old phrase *was* the residue). `cd cli && go test ./...` pass.
- [2026-08-30] (p3-one-plan-guard, wave 1): task_status=DONE. Added zharness_guard_at_most_one_active_plan in ZGUARD-CORE; wired pre-commit and cli-ci.yml; R5 fixtures (zero/one/two/empty). `bash scripts/test-guards.sh`: 19 passed, 0 failed. S1–S5 still ok.
- [2026-08-30] wave-summary: p1+p2+p3 executed this session. Leftover live lie outside scope: assets/spec-plan-workflow.svg still shows `zharness init --json`.
- [2026-08-30] (p1, owner override): task_status=DONE. Owner asked to fix remaining live lies. Updated `assets/spec-plan-workflow.svg` and `assets/workflow-usage-flow.html` to `zharness install`; ARCHITECTURE now lists the third guard (at-most-one active plan). `rg zharness init assets/`: empty.

## Decisions
- [2026-08-30] (p2): stop_condition said restore a missing required test phrase; the missing phrase was itself 0.14 residue (`mirror their DB transitions while the ledger exists`). Replaced the required string with the new playbook sentence instead of putting the lie back. R1 wins over a test that encoded the ghost.
- [2026-08-30] (p1): assets SVG / workflow-usage-flow.html still say `zharness init`. Left per p1 escalation (outside surfaces_allowed).
- [2026-08-30] (handoff): owner overrode p1 escalation and required the SVG/HTML/ARCHITECTURE leftovers fixed before close. Remaining audit items H3/H5–H9 stay out of this plan (NG5); they need a new lock after this file moves to completed.

## Validation
- [2026-08-30 (gate)] mode `gate` verdict `APPROVED` — judge: `same-session` (lane: normal). Combined p1–p3 diff. Same-session: Security of the new count guard vs a malicious empty-file pair was not independently reviewed. Architecture: no new CLI verbs. Performance: glob of docs/plans/active/*.md. Code quality: fixture suite includes reject-two.
  - `bash scripts/test-guards.sh`
  - `bash scripts/verify-doc-links.sh`
  - `cd cli && go test ./...`
  receipt:
    context_sources: [plan p1-p3, playbooks, ZGUARD-CORE]
    policy: docs/playbooks/check.md
    judge: same-session
    judge_model: grok-4.6
    retries: 1
    rollback_point: none
    not_independently_verified: empty-file pair as a bypass of R5
- [2026-08-30 (initiative close)] mode `full` verdict `APPROVED` — judge: `same-session` (lane: normal). Security: hook still runs `sh -c` on cited proofs (unchanged threat model, not independently re-audited); one-plan guard is a glob of non-empty markdown, no secrets. Performance: no new hot path. Architecture: three verbs, markdown record, three ZGUARD-CORE checks; no lifecycle CLI returned. Code Quality: 19 guard fixtures, `go test ./...` green, doc-link gate 0 findings. SPA complete for close.
  - `bash scripts/test-guards.sh`
  - `bash scripts/verify-doc-links.sh`
  - `cd cli && go test ./...`
  receipt:
    context_sources: [plan p1-p3, ARCHITECTURE, assets]
    policy: docs/playbooks/check.md
    judge: same-session
    judge_model: grok-4.6
    retries: 0
    rollback_point: none
    not_independently_verified: sh -c proof execution threat model

## Current State and Next Action
- active_phase: none
- lifecycle_status: done
- blockers: none
- open_items: none
- exact_next_action: none (completed; remaining audit H3/H5–H9 need a new lock)
