---
id: zharness-slim-20260919T1200Z
intake_id: zharness-slim-intake-20260919T1200Z
lane: high-risk
status: active
created: 2026-09-19
updated: 2026-09-19
---

# Plan: zharness-slim — consistent playbooks across repos, a smaller installer, a lighter plan

## Outcome
- result: after a `zharness` upgrade the owner can list every installed repository whose managed
  docs lag the binary, `update` no longer carries a three-way merge, a new plan is a 5-section file
  that scales with its lane, and the playbooks name stages instead of Claude slash commands.
- success_signals:
  - `zharness update --check --all` exits 1 and names Ligaturizer (known stale set) on this machine.
  - `cli/internal/installer/threeway.go` is gone and non-test Go lines in `cli/internal/installer/` drop by at least 400
    from 1,794 (`wc -l` excluding `*_test.go`).
  - `scripts/test-guards.sh` passes with zero edits between `# ZGUARD-CORE-BEGIN` and `# ZGUARD-CORE-END`.
  - The plan template has 5 `## ` sections; a `tiny` request creates no plan file.
  - `rg -n '`/(check|work|to-plan|brainstorm|handoff|watzup)\b' cli/docs/embedded` returns 0 lines
    (pattern: a backtick followed by a slash command name).

## Authority and Requirements
- authority:
  - Owner approval in the 2026-09-19 `/think` session (design, P1–P4 order, registry variant of P1,
    auto-migration of the active plan).
  - ADR 0007 (fresh overwrite) and ADR 0008 (recorded ownership; stash transaction), partly superseded here.
  - Guard core in `scripts/install-git-hooks.sh` (headings and fields it parses).
- requirements:
  - R1 [accepted]: `install` and `update` record the repository root once (deduplicated) in
    `~/.config/zharness/repos`; `uninstall` removes it. | source: owner, P1 registry
  - R2 [accepted]: `zharness update --check` is read-only and exits 1 when any fresh-overwrite target
    or the AGENTS block differs from the embedded bytes, or when `docs/PROJECT.md` lacks a heading the
    embedded identity template has; it exits 0 otherwise and names each drifted path. | source: owner, P1
  - R3 [accepted]: `--check --all` runs R2 over every registered root, reports missing roots without
    failing on them, and exits 1 if any root drifted. `scripts/install-zharness.sh` runs it after an
    upgrade. | source: owner, P1
  - R4 [accepted]: `update` replaces the ZHARNESS block between its markers; if the block's sha256
    differs from the hash recorded at the last write, `update` writes nothing, prints the diff and
    exits non-zero unless `--force` is passed. Prose outside the markers is never changed. | source: owner, P2
  - R5 [accepted]: `docs/PROJECT.md` is written only when absent and never merged or overwritten. | source: owner, P2
  - R6 [accepted]: three-way merge, `.zharness/base/upstream/`, the update stash, and
    `--continue`/`--abort` are removed; the ownership ledger `.zharness/base/ownership.tsv` and every
    ADR 0008 decision 1–4 behavior are preserved (existing uninstall tests pass unchanged). A new ADR
    0011 records the supersession of ADR 0007's `Merge: true` and ADR 0008 decision 5. | source: owner, ADR 0008
  - R7 [accepted]: all checks in R4 run before the first write, so a refused `update` changes no file. | source: ADR 0008 intent
  - R8 [accepted]: a new plan has exactly these sections: `## Goal`, `## Phases and Verification`,
    `## Log`, `## Validation`, `## Current State and Next Action`; the three hook-parsed headings keep
    their exact text and the guard core is not edited. | source: owner, P3; guard audit
  - R9 [accepted]: lane scaling — `tiny` creates no plan (bounded path); `normal` has one phase with
    `phase_slug:` and `status:` and a flat task list; `high-risk` keeps phases, waves and tasks. Plan
    frontmatter drops `intake_id` and `story_id` (no guard or CLI reads them). | source: owner, P3
  - R10 [accepted]: a `## Validation` entry cites commands with exit code and at most 3 output lines. | source: owner, P3
  - R11 [accepted]: at final close `handoff` drops `done` entries from `## Log`, keeps `decision`
    entries and the last Validation entry per phase byte-identical, and the hook passes on that commit. | source: owner, P3
  - R12 [accepted]: `update` migrates the active plan from the 9-section format: section merges only,
    `## Validation` bytes identical (sha256 before = after), idempotent on an already-migrated plan. | source: owner, P3
  - R13 [accepted]: `check.md`, `handoff.md`, `spec.md`, `plan.md` are removed from
    `cli/docs/embedded/templates/` (no Go code reads them; only `project.identity.md` is read) and the
    stale `zharness scaffold` comment in `cli/docs/embedded/embed.go` is corrected; the plan skeleton
    lives only in the `brainstorm` playbook; `check` exposes `bounded`, `gate`, `full` (`review` folds
    into `bounded`); `brainstorm` drops the input-type taxonomy. | source: owner, P3
  - R14 [accepted]: playbooks and WORKFLOW.md name stages, not slash syntax; the F1 rationale is
    host-neutral ("do not switch model or agent mid-phase; prompt cache is per model"); model pins
    become tiers in WORKFLOW.md; `rules/workflow-core.md` no longer routes every commit through `check`. | source: owner, P4
  - R15 [accepted]: the owner's machine keeps one skill source; the symlink between `~/.agents/skills`
    and `~/.claude/skills` is created only after explicit owner confirmation at that step. | source: owner, P4

## Non-goals
- NG1: serving playbooks from the binary or a global directory (rejected: breaks fresh clones without
  the binary and couples repo plan format to a machine-wide version).
- NG2: shipping the ZGUARD hook or CI guard to consumer repositories.
- NG3: any edit inside the guard core, or re-validating historical completed plans.
- NG4: guarding `:NN` line anchors (accepted debt from harness-eval-loop).
- NG5: fixing OpenCode's duplicate skill listing.

## Approach and Risks
- approach: not-planned
- constraints:
  - none
- risks:
  - none

## Phases and Verification
- planning_status: not-planned
- phases: none

## Progress
- none

## Decisions
- none

## Validation
- none

## Current State and Next Action
- active_phase: none
- lifecycle_status: not-planned
- blockers: none
- open_items: [to-plan must define stable phases, waves, tasks, and checks]
- exact_next_action: to-plan
