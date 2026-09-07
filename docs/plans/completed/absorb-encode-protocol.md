---
id: 01M0ABSORBENCODE9K3XJ7
type: plan
intake_id: 01M0ABSORBENCODEINTK9K3XJ7
lane: normal
status: completed
created: 2026-08-30
updated: 2026-08-30
---

# Plan: absorb-encode-protocol — handoff absorbs lessons; encode/improve skills; completed/ policy

## Outcome
- result: Closing an initiative asks three absorb questions before `git mv`. A class-of-failure or expensive-to-reverse decision cannot close until it lives in an ADR and/or a native guard. Completed plans are kept only when recovery still has an audience; otherwise they may be deleted after absorb. Two explicit skills (`encode-invariant`, `improve-harness`) exist in this repo. The Go installer and fail-closed hooks are unchanged in semantics.
- success_signals:
  - S1: `docs/playbooks/handoff.md` step 6 (and the embedded source) contains the absorb gate and the words `absorb: none`.
  - S2: A close with an unabsorbed invariant is specified as a stop, not a `git mv`.
  - S3: `docs/plans/completed/README.md` states retain-vs-delete in experimental language (recovery audience).
  - S4: `skills/workflow/encode-invariant/SKILL.md` exists, ≤30 lines trigger, defers to `docs/patterns/encoding-invariants.md`.
  - S5: `skills/workflow/improve-harness/SKILL.md` exists and requires a fresh-rerun before claiming improvement.
  - S6: `rg -n "zharness preflight|harness.db|zharness memory add" cli/docs/embedded/playbooks skills/workflow/encode-invariant skills/workflow/improve-harness` prints nothing.
  - S7: `bash scripts/verify-doc-links.sh` 0 findings; `cd cli && go test ./...` passes; `bash scripts/test-guards.sh` still 0 failed.

## Authority and Requirements
- authority:
  - Owner, 2026-08-30: implement the think spec “zharness protocol v0.16”; prefer the better design over preserving current ceremony; do not rewrite the Go installer; do not resurrect SQLite.
  - [repository-harness completed/README](https://raw.githubusercontent.com/hoangnb24/repository-harness/main/docs/plans/completed/README.md) — retain completed plans only when recovery has an audience.
  - [encoding-invariants.md](https://raw.githubusercontent.com/hoangnb24/repository-harness/main/docs/patterns/encoding-invariants.md) — authority gate, native validation, +/- proof.
  - [improve-harness SKILL](https://raw.githubusercontent.com/hoangnb24/repository-harness/main/.agents/skills/improve-harness/SKILL.md) — explicit, evidence, fresh rerun.
  - `docs/decisions/0003-durable-memory-not-wired-into-playbooks.md` — no mandatory memory write on handoff.
  - `docs/ARCHITECTURE.md` — three verbs; markdown is the record; hook guards.
  - `cli/docs/CONTRACT.md` — binary surface unchanged.
- requirements:
  - R1 [accepted]: `handoff` final-close (step 6), immediately before `git mv`, requires an absorb block in `## Decisions`: either `absorb: none` or named outputs (ADR path and/or guard path and/or memory id). | source: think spec absorb gate
  - R2 [accepted]: If the close answers that a class-of-failure or expensive-to-reverse decision exists and no ADR/guard yet records it, handoff **stops** and does not `git mv`. | source: think spec; experimental encode-invariant authority gate
  - R3 [accepted]: `absorb: none` is a valid close. No ADR, memory file, or guard is required when the run was only a log. | source: think spec; ADR 0003
  - R4 [accepted]: No handoff step writes `docs/memory/` unless work.md’s three triggers already fired. | source: ADR 0003
  - R5 [accepted]: `docs/plans/completed/README.md` tells agents to retain a completed plan only when recovery/transition still has an independent audience; otherwise rely on code, tests, ADRs, PRs, and git, and the plan **may** be deleted after R1 absorb. Default when unsure: keep. | source: experimental completed/README
  - R6 [accepted]: `docs/decisions/README.md` states that ADRs are the absorb path for expensive-to-reverse decisions; completed plans are not project knowledge. | source: think spec
  - R7 [accepted]: `docs/patterns/encoding-invariants.md` exists in this repo (adapted, not a verbatim copy of experimental paths/CLI). A thin `skills/workflow/encode-invariant/SKILL.md` routes to it. | source: experimental encoding-invariants.md
  - R8 [accepted]: Encode-invariant requires positive and negative proof against a native owner (`test-guards.sh`, `verify-doc-links.sh`, or `go test`). | source: experimental §4
  - R9 [accepted]: Thin `skills/workflow/improve-harness/SKILL.md`: one friction, one intervention, `Decision: pending fresh rerun` until a different session reruns; no claim of harness improvement without that. | source: experimental improve-harness
  - R10 [accepted]: No new `zharness` subcommand. No change to proof-reexec, high-risk judge, one-plan count, or H3 first-line full-judge semantics. | source: owner; CONTRACT.md
  - R11 [accepted]: Playbooks are edited in `cli/docs/embedded/playbooks/` then copied to `docs/playbooks/`. | source: ARCHITECTURE.md projection rule
  - R12 [accepted]: `skills/workflow/README.md` lists the two new skills as non-spine, never blocking a missing binary. | source: workflow README mapping table

## Non-goals
- NG1: No SQLite, preflight, trace add, memory CLI, or installer rewrite.
- NG2: No default deletion of all `docs/plans/completed/*.md` in this initiative (policy text only; no bulk delete of existing completed plans).
- NG3: No fail-closed hook that rejects a close because absorb was `none`.
- NG4: No auto-write to `docs/memory/` on handoff.
- NG5: No GRAPH, sandbox-in-binary, or evaluator platform.
- NG6: No edits to ADR 0001–0005 bodies.
- NG7: No copying experimental `scripts/bin/harness-cli` or Rust crates.
- NG8: Bulk absorb of historical completed plans in this repo is out of scope (policy going forward).

## Approach and Risks
- approach: Lock the protocol in playbooks first (P1 ships to every `zharness install` consumer). Then add this-repo skills + pattern (P2/P3). Skills are not in the Go managed set; consumers get the absorb gate via handoff.md alone. Rejected: new CLI verbs; delete-all completed plans this PR; hook that fails closed on missing absorb (would block every close).
- constraints:
  - Edit embedded playbooks, then `cp` to `docs/playbooks/`.
  - ADR 0003: absorb must not become a mandatory memory write.
  - Hook semantics frozen (R10).
- rejected alternatives:
  - Rewrite `cli/cmd/zharness` to add `absorb` — rejected: the gate is a handoff question, not a binary.
  - Delete-by-default for completed/ — rejected this initiative: existing audits/plans cite those files; R5 is policy for *future* closes, NG2 forbids bulk delete now.
  - Mandatory ADR on every close — rejected: most closes are logs (`absorb: none`).
- risks:
  - Agents skip the three questions. Mitigation: R1 requires a visible `absorb:` line in Decisions; still not a hook (R3/NG3).
  - encode-invariant skill invents policy. Mitigation: pattern §1 stop without accepted authority (copy experimental gate).
  - improve-harness claims success without fresh rerun. Mitigation: R9 forbids the claim; skill text is the check.
- recovery: revert the playbook/skill files. Hooks untouched. No migration.

## Phases and Verification
- planning_status: planned
- phases:
  - phase_slug: p1-absorb-gate
    story_id: 01M0ABSORBENCODEPH19K3XJ7
    status: done
    goal: Handoff cannot close without an absorb line; completed/README and decisions README state retain-vs-knowledge.
    depends_on: none
    surfaces_allowed: cli/docs/embedded/playbooks/handoff.md, docs/playbooks/handoff.md, docs/plans/completed/README.md, docs/decisions/README.md, docs/PROJECT.md
    surfaces_avoided: cli/cmd/**, cli/internal/**, scripts/install-git-hooks.sh, docs/decisions/0001*.md, docs/decisions/0002*.md, docs/decisions/0003*.md, docs/decisions/0004*.md, docs/decisions/0005*.md, docs/plans/completed/*.md
    requirements: R1, R2, R3, R4, R5, R6, R10, R11
    waves:
      - wave: 1
        goal: Write the gate and the two policy pages.
        tasks:
          - task: Insert the absorb gate into embedded handoff.md step 6 immediately before `git mv`; require `absorb: none` or named ADR/guard/memory in ## Decisions; stop if class-of-failure or expensive decision is unanswered by an artifact; project to docs/playbooks/handoff.md.
            verify: `rg -n "absorb: none" cli/docs/embedded/playbooks/handoff.md docs/playbooks/handoff.md` matches both; `diff -q cli/docs/embedded/playbooks/handoff.md docs/playbooks/handoff.md` is silent.
          - task: Add docs/plans/completed/README.md with retain-only-if-recovery-audience and may-delete-after-absorb; default when unsure is keep.
            verify: `rg -n "recovery|audience|absorb" docs/plans/completed/README.md` matches.
          - task: Add one sentence to docs/decisions/README.md that ADRs are the absorb path for expensive-to-reverse decisions and completed plans are not project knowledge.
            verify: `rg -n "absorb" docs/decisions/README.md` matches.
        stop_condition: the insert weakens full-check or git-mv identity rules — revert the hunk.
    checks:
      - `bash scripts/verify-doc-links.sh`
      - `rg -n "zharness preflight|harness.db" cli/docs/embedded/playbooks/handoff.md` prints nothing
    escalation: if consumers need completed/README in the installer managed set, stop and do not widen into cli/internal/installer this phase.

  - phase_slug: p2-encode-invariant
    story_id: 01M0ABSORBENCODEPH29K3XJ7
    status: done
    goal: This repo can turn an accepted rule into a native guard with +/- proof, without a new CLI.
    depends_on: p1-absorb-gate
    surfaces_allowed: docs/patterns/encoding-invariants.md, skills/workflow/encode-invariant/SKILL.md, skills/workflow/README.md
    surfaces_avoided: cli/**, scripts/install-git-hooks.sh, docs/playbooks/handoff.md
    requirements: R7, R8, R12
    waves:
      - wave: 1
        goal: Pattern + thin skill + README row.
        tasks:
          - task: Write docs/patterns/encoding-invariants.md (authority gate, boundary table, native owner, +/- proof, four enforcement levels). Do not copy experimental CLI paths or crate names.
            verify: `test -f docs/patterns/encoding-invariants.md` and `rg -n "harness-cli|crates/harness" docs/patterns/encoding-invariants.md` prints nothing.
          - task: Add skills/workflow/encode-invariant/SKILL.md as a thin trigger (≤30 body lines) routing to the pattern; never hard-stop on a missing zharness binary.
            verify: `wc -l skills/workflow/encode-invariant/SKILL.md` body after frontmatter is ≤30; `rg -n "docs/patterns/encoding-invariants.md" skills/workflow/encode-invariant/SKILL.md` matches.
          - task: Add encode-invariant to skills/workflow/README.md mapping as non-spine.
            verify: `rg -n "encode-invariant" skills/workflow/README.md` matches.
        stop_condition: skill body exceeds thin-trigger size — cut, do not add references/ yet.
    checks:
      - `bash scripts/verify-doc-links.sh`
    escalation: none

  - phase_slug: p3-improve-harness
    story_id: 01M0ABSORBENCODEPH39K3XJ7
    status: done
    goal: Explicit harness improvement cannot claim success without a fresh-session rerun.
    depends_on: p2-encode-invariant
    surfaces_allowed: skills/workflow/improve-harness/SKILL.md, skills/workflow/README.md, docs/templates/harness-improvement.md
    surfaces_avoided: cli/**, scripts/install-git-hooks.sh, docs/playbooks/**
    requirements: R9, R12
    waves:
      - wave: 1
        goal: Thin skill + optional template.
        tasks:
          - task: Add docs/templates/harness-improvement.md with fields baseline, gap owner, intervention, native proof, fresh-rerun, keep/revise/remove/pending.
            verify: `rg -n "pending fresh rerun" docs/templates/harness-improvement.md` matches.
          - task: Add skills/workflow/improve-harness/SKILL.md thin trigger; forbids claiming improvement without a different-session rerun; copy-from template into docs/plans/active/ only when the user invokes the skill.
            verify: `rg -n "pending fresh rerun|different session" skills/workflow/improve-harness/SKILL.md` matches.
          - task: List improve-harness as non-spine in skills/workflow/README.md.
            verify: `rg -n "improve-harness" skills/workflow/README.md` matches.
        stop_condition: skill tries to mutate hooks or installer — cut that instruction.
    checks:
      - `bash scripts/verify-doc-links.sh`
      - `cd cli && go test ./...`
      - `bash scripts/test-guards.sh`
      - `rg -n "zharness preflight|harness.db|zharness memory add" cli/docs/embedded/playbooks skills/workflow/encode-invariant skills/workflow/improve-harness` prints nothing
    escalation: none

## Progress
- [2026-08-30] (p1-absorb-gate, wave 1): task_status=in-progress. Phase start.
- [2026-08-30] (p1-absorb-gate, wave 1): task Insert absorb gate into handoff.md. task_status=DONE. Embedded step 6 has absorb: none / adr / guard / memory; stop without git mv if unabsorbed invariant. Projected. diff -q silent.
- [2026-08-30] (p1-absorb-gate, wave 1): task Add completed/README.md. task_status=DONE. Retain if recovery audience; may delete after absorb; default keep.
- [2026-08-30] (p1-absorb-gate, wave 1): task Decisions README absorb sentence. task_status=DONE. ADRs are absorb path; completed plans are run logs.
- [2026-08-30] wave-summary: p1-absorb-gate wave 1 done.
- [2026-08-30] (p2-encode-invariant, wave 1): task_status=in-progress. Phase start.
- [2026-08-30] (p2-encode-invariant, wave 1): task Write encoding-invariants.md. task_status=DONE. Pattern adapted; no harness-cli/crates/harness.
- [2026-08-30] (p2-encode-invariant, wave 1): task encode-invariant SKILL.md. task_status=DONE. Body 6 lines; routes to pattern; missing binary never stops.
- [2026-08-30] (p2-encode-invariant, wave 1): task README mapping. task_status=DONE. Non-spine row added.
- [2026-08-30] wave-summary: p2-encode-invariant wave 1 done.
- [2026-08-30] (p3-improve-harness, wave 1): task_status=in-progress. Phase start after p2 gate.
- [2026-08-30] (p3-improve-harness, wave 1): task harness-improvement.md template. task_status=DONE. Fields include pending fresh rerun.
- [2026-08-30] (p3-improve-harness, wave 1): task improve-harness SKILL.md. task_status=DONE. Forbids claim without different session; no hook/installer mutation.
- [2026-08-30] (p3-improve-harness, wave 1): task README mapping. task_status=DONE. Non-spine row added.
- [2026-08-30] wave-summary: p3-improve-harness wave 1 done.

## Decisions
- [2026-08-30] absorb: none

## Validation
- [2026-08-30 (gate)] mode `gate` verdict `APPROVED` — judge: `same-session` (lane: normal). p1-absorb-gate. Same-session: installer managed-set inclusion of completed/README not independently decided (escalation: do not widen).
  - `bash scripts/verify-doc-links.sh`
  receipt:
    context_sources: [handoff.md, completed/README.md, decisions/README.md]
    policy: docs/playbooks/check.md
    judge: same-session
    judge_model: grok-4.6
    retries: 0
    rollback_point: none
    failure_ledger: absent
    not_independently_verified: whether completed/README should join the installer managed set
- [2026-08-30 (gate)] mode `gate` verdict `APPROVED` — judge: `same-session` (lane: normal). p2-encode-invariant.
  - `bash scripts/verify-doc-links.sh`
  receipt:
    context_sources: [docs/patterns/encoding-invariants.md, skills/workflow/encode-invariant/SKILL.md, skills/workflow/README.md]
    policy: docs/playbooks/check.md
    judge: same-session
    judge_model: grok-4.6
    retries: 0
    rollback_point: none
    failure_ledger: absent
    not_independently_verified: none
- [2026-08-30 (gate)] mode `gate` verdict `APPROVED` — judge: `same-session` (lane: normal). p3-improve-harness. Final-phase complete review deferred to independent `check full` per work.md.
  - `bash scripts/verify-doc-links.sh`
  - `cd cli && go test ./...`
  - `bash scripts/test-guards.sh`
  receipt:
    context_sources: [skills/workflow/improve-harness/SKILL.md, docs/templates/harness-improvement.md, skills/workflow/README.md]
    policy: docs/playbooks/check.md
    judge: same-session
    judge_model: grok-4.6
    retries: 0
    rollback_point: none
    failure_ledger: absent
    not_independently_verified: full Security/Performance/Architecture/Code Quality review
- [2026-08-30 (full)] mode `full` verdict `APPROVED` — judge: `independent` (lane: normal). p3-improve-harness. Security/Performance/Architecture/Code Quality: no secrets, no hook weaken, 3-verb installer untouched, thin skills, absorb is playbook not fail-open hook. Independent reviewer FullJudge.
  - `bash scripts/verify-doc-links.sh`
  - `cd cli && go test ./...`
  - `bash scripts/test-guards.sh`
  receipt:
    context_sources: [handoff.md, completed/README.md, encoding-invariants.md, encode-invariant/SKILL.md, improve-harness/SKILL.md, harness-improvement.md]
    policy: docs/playbooks/check.md
    judge: independent
    judge_model: grok-4.6
    retries: 0
    rollback_point: none
    failure_ledger: absent
    not_independently_verified: none

## Current State and Next Action
- active_phase: none
- lifecycle_status: done
- blockers: none
- open_items: none
- exact_next_action: none (initiative completed; absorb: none)
