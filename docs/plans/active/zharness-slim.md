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
  - `zharness update --check --root ~/Lab/Ligaturizer` exits 1 and names its stale playbooks, writing
    nothing there; `--check --all` over a fixture registry exits 1 and names the drifted fixture.
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
  - R10 [accepted]: a `## Validation` entry's first line carries `verdict: <VERDICT>` (the only form the
    guard's R2 re-executes; an unanchored verdict skips R2) and cites each command with its result in at
    most 3 output lines. | source: owner, P3; guard audit 2026-09-19
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
  - R16 [accepted]: `scripts/install-git-hooks.sh` replaces an existing pre-commit hook it installed when
    that hook's bytes differ from the current template, and exits non-zero (no silent `|| true`) when the
    existing hook is foreign; the guard core is untouched. | source: owner, 2026-09-19 stale-hook finding

## Non-goals
- NG1: serving playbooks from the binary or a global directory (rejected: breaks fresh clones without
  the binary and couples repo plan format to a machine-wide version).
- NG2: shipping the ZGUARD hook or CI guard to consumer repositories.
- NG3: any edit inside the guard core, or re-validating historical completed plans.
- NG4: guarding `:NN` line anchors (accepted debt from harness-eval-loop).
- NG5: fixing OpenCode's duplicate skill listing.

## Approach and Risks
- approach: four sequential, independently mergeable phases on branch `feat/zharness-slim`, one
  commit group per phase, each closed by `check full` with an independent (subagent) judge because
  the lane is high-risk. P1 adds the read-only drift check first so P2–P4 changes are observable on
  real repositories. P2 removes the merge machinery before P3 so plan migration lands in the smaller
  `update.go`. P3 changes the plan format and migrates this plan itself as its live proof. P4 is
  prose-only plus one owner-confirmed home-directory step.
- constraints:
  - Guard core (between `# ZGUARD-CORE-BEGIN` / `# ZGUARD-CORE-END`) byte-identical to `master`.
  - `docs/playbooks/**` and `cli/docs/embedded/playbooks/**` stay byte-identical copies; same for
    `docs/WORKFLOW.md` and `cli/docs/embedded/WORKFLOW.md`.
  - No write into any consumer repository; real-repo checks use `--check` only.
  - Home-directory writes limited to `~/.config/zharness/repos` (R1) and the R15 symlink after confirmation.
  - Deletions use `trash`.
- dependencies: Go toolchain for `cli/`; `gh` only for the existing release script; no new modules.
- rejected:
  - Global playbook serving (NG1).
  - Registry via `os.UserConfigDir()` — resolves to `~/Library/Application Support` on macOS,
    against R1; use `$XDG_CONFIG_HOME` else `$HOME/.config`.
  - Keeping `.zharness/base/upstream/` for a lighter AGENTS diff — one sha256 in the manifest is enough
    to detect a hand edit (R4).
  - A separate migration command — owner chose automatic migration inside `update` (R12).
- risks:
  - P2 downgrade: an older binary expects `upstream/` blobs. Mitigation: ADR 0011 states that downgrade
    needs `zharness uninstall` then `install`; ownership ledger keeps uninstall safe.
  - P3 migration corrupts a user plan. Mitigation: migrate only a plan whose `## ` headings exactly match
    the known 9-section set; any other heading set leaves the file untouched with a notice; assert
    Validation sha256 equality before writing; write via temp file plus rename.
  - P3 hides guard coverage: hook-parsed headings or fields renamed. Mitigation: R8; `test-guards.sh`
    and the negative R2 probe run in the P3 gate.
  - P1 registry write failure (read-only HOME) blocks install. Mitigation: registry errors warn only.
- recovery: each phase reverts with `git revert` of its commits; P2 consumer state recovers with the
  previous release binary plus `uninstall`/`install`. Stop and ask the owner if a gate returns
  `REQUEST_CHANGES` twice on one phase, or if any change would touch the guard core.

## Phases and Verification
- planning_status: planned
- phases:
  - phase_slug: `p1-drift-check`
    - story_id: `p1-drift-check-20260919T1300Z`
    - status: checked
    - goal: R1, R2, R3, R16.
    - depends_on: none
    - surfaces: `cli/internal/installer/{registry,check}.go` (+ tests), `installer.go`, `update.go`,
      `uninstall.go`, `cli/internal/interfaces/manage.go`, `scripts/install-zharness.sh`,
      `scripts/install-git-hooks.sh` (outside core), `scripts/test-guards.sh`
    - avoided: guard core, playbooks, templates, consumer repos
    - wave 1:
      - T1 registry: `Register(root)`, `Unregister(root)`, `Registered()` over `$XDG_CONFIG_HOME` or
        `$HOME/.config` + `/zharness/repos`, one absolute path per line, deduplicated; `Install` and
        `RunUpdate` register, `Uninstall` unregisters; errors print a warning and never fail the verb.
        check: `cd cli && go test ./internal/installer/ -run Registry` (temp HOME: add twice → 1 line;
        uninstall → 0 lines; read-only dir → verb still exits 0).
      - T2 drift check: `Check(root) ([]string, error)` compares each fresh-overwrite target and the
        rendered AGENTS block with the on-disk bytes, and PROJECT.md `## ` headings with the embedded
        identity template; `update --check [--all]` in `manage.go`, exit 1 on drift, missing registered
        roots reported as `missing:` without failing. check: `cd cli && go test ./internal/installer/
        ./internal/interfaces/ -run 'Check'` (fixture clean → 0; edited playbook → 1 naming it;
        dropped PROJECT heading → 1; `--all` with one clean + one drifted + one missing root → 1).
    - wave 2:
      - T3 upgrade hook: `scripts/install-zharness.sh` runs `zharness update --check --all` after
        `--version`, printing the result and never failing the install. check: `bash -n
        scripts/install-zharness.sh && bash scripts/test-install-zharness.sh`.
      - T4 R16: `create_pre_commit_hook` compares the existing hook with the template; identical →
        "up to date"; ours (first two lines match) but different → overwrite; foreign → error exit;
        the `|| true` in `main` is removed. check: `bash scripts/test-guards.sh` with a new case for
        stale-ours → replaced and foreign → non-zero.
    - phase check: gate commands from `docs/PROJECT.md`; `bash scripts/test-guards.sh`;
      `go build -o "$TMPDIR/zh" ./cmd/zharness && "$TMPDIR/zh" update --check --root ~/Lab/Ligaturizer;
      test $? -eq 1 && git -C ~/Lab/Ligaturizer status --short` (exit 1, no new changes there).
  - phase_slug: `p2-drop-threeway`
    - story_id: `p2-drop-threeway-20260919T1300Z`
    - status: planned
    - goal: R4, R5, R6, R7.
    - depends_on: `p1-drift-check`
    - surfaces: `cli/internal/installer/*.go` (+ tests), `cli/internal/interfaces/manage.go`,
      `cli/internal/embedded/embedded.go` (manifest field), `docs/decisions/0011-*.md`,
      `docs/PROJECT.md` ("where state lives"), `scripts/test-install-zharness.sh`
    - avoided: `ownership.go` behavior, `ownership_test.go`, guard core, playbooks
    - wave 1:
      - T1 ADR 0011: supersedes ADR 0007 `Merge: true` and ADR 0008 decision 5; states hash guard,
        write-once PROJECT.md, downgrade path. check: `bash scripts/verify-doc-links.sh`.
      - T2 AGENTS hash guard: manifest gains `agents_block_sha256` written on every block write;
        `update` compares the on-disk block with it before any write, and on mismatch prints a unified
        diff and returns an error unless `--force`; a manifest without the field records the current
        block hash (first run after upgrade, no refusal). check: `go test ./internal/installer/ -run
        Agents` (untouched → replaced; hand-edited → error and tree sha unchanged; `--force` →
        replaced; legacy manifest → accepted).
      - T3 PROJECT.md write-once: target loses `Merge`; `Install`/`update` write it only when absent.
        check: `go test ./internal/installer/ -run Project` (existing file with custom text byte-identical
        after update).
    - wave 2:
      - T4 removal: `trash` `threeway.go`, `stash_test.go`; delete stash, conflict, base-draft and
        upstream code from `update.go`; drop `--continue`/`--abort`; `update` removes a leftover
        `.zharness/base/upstream/` and keeps `ownership.tsv` and `manifest.json`. check: `test ! -e
        cli/internal/installer/threeway.go`; `rg -n 'stash|threeWay|conflictOpenTag|upstream/'
        cli/internal` returns 0 lines; `git diff master -- cli/internal/installer/ownership_test.go`
        empty; non-test LOC ≤ 1,394.
    - phase check: gate commands; `bash scripts/test-install-zharness.sh`; temp-repo run of
      install → edit AGENTS block → `update` (non-zero, `find . -type f | sort | xargs shasum` identical)
      → `update --force` (block restored) → `uninstall` (managed set gone, user files kept).
  - phase_slug: `p3-slim-plan`
    - story_id: `p3-slim-plan-20260919T1300Z`
    - status: planned
    - goal: R8, R9, R10, R11, R12, R13.
    - depends_on: `p2-drop-threeway`
    - surfaces: `docs/playbooks/*.md` and embedded copies, `docs/WORKFLOW.md` and embedded copy,
      `cli/docs/embedded/templates/`, `cli/docs/embedded/embed.go`, `cli/internal/installer/migrate.go`
      (+ test), `update.go`, `skills/workflow/README.md`, `site/docs/workflow.html`,
      `skills/craft/create-skill/references/skill-anatomy-and-requirements.md`, this plan
    - avoided: guard core, `docs/plans/completed/**`
    - wave 1:
      - T1 format: `brainstorm.md` holds the 5-section skeleton (`## Goal` = outcome, requirements,
        non-goals; `## Phases and Verification` opens with `approach:`/`risks:`), lane scaling, no
        input-type taxonomy, frontmatter `lane`, `status` only; `to-plan.md` owns `## Phases and
        Verification` and stops minting `story_id`; `work.md`/`work-full.md` append to `## Log`;
        `check.md`/`check-validation.md` expose `bounded|gate|full` and the R10 entry shape;
        `handoff.md` adds close-time compaction (R11); `WORKFLOW.md` matches. check: `diff -r
        docs/playbooks cli/docs/embedded/playbooks && diff docs/WORKFLOW.md cli/docs/embedded/WORKFLOW.md`;
        `rg -n 'story_id|intake_id|## Progress|## Decisions|mode: review' docs/playbooks docs/WORKFLOW.md`
        returns 0 lines; `wc -w` of playbooks + WORKFLOW below 5,427.
      - T2 dead templates: `trash` `check.md handoff.md spec.md plan.md` under
        `cli/docs/embedded/templates/`; fix the `embed.go` comment; update README, site page, and the
        skill-anatomy reference. check: `ls cli/docs/embedded/templates` shows only
        `project.identity.md`; `bash scripts/verify-doc-links.sh`.
    - wave 2:
      - T3 migration: `MigratePlan([]byte) ([]byte, bool, error)` maps Outcome + Authority and
        Requirements + Non-goals → Goal, Approach and Risks → head of Phases and Verification, Progress +
        Decisions → Log, drops `intake_id`/`story_id` lines; any other heading set → unchanged; refuses
        when Validation sha256 would differ; `RunUpdate` applies it to the single active plan.
        check: `go test ./internal/installer/ -run Migrate` (9-section fixture → 5 sections, Validation
        sha equal; second run no-op; unknown heading → untouched).
    - wave 3:
      - T4 live migration: built binary migrates this plan; commit through the hook. check: Validation
        sha256 before = after; `bash .git/hooks/pre-commit` exit 0; negative R2 probe (anchored
        `verdict: APPROVED` + `false`) exit 1, then restored.
      - T5 compaction measure: apply R11 to a scratch copy of
        `docs/plans/completed/harness-eval-loop.md`; report `wc -w` before/after; the completed file is
        not modified. check: `git diff --quiet master -- docs/plans/completed/`.
    - phase check: gate commands; `bash scripts/test-guards.sh`; `diff <(git show master:scripts/install-git-hooks.sh | awk '$0=="# ZGUARD-CORE-BEGIN"{on=1;next} $0=="# ZGUARD-CORE-END"{on=0} on') <(awk '$0=="# ZGUARD-CORE-BEGIN"{on=1;next} $0=="# ZGUARD-CORE-END"{on=0} on' scripts/install-git-hooks.sh)` empty.
  - phase_slug: `p4-portability`
    - story_id: `p4-portability-20260919T1300Z`
    - status: planned
    - goal: R14, R15.
    - depends_on: `p3-slim-plan`
    - surfaces: playbooks + embedded copies, WORKFLOW.md + embedded copy, `rules/workflow-core.md`,
      `skills/workflow/README.md`; home directory only for T3 after confirmation
    - avoided: Go code, guard core
    - wave 1:
      - T1 stage names: replace slash-command syntax with stage names; neutral F1 wording; model pins →
        tiers (`deep`, `standard`, `fast`) with one host mapping table. check: `rg -n
        '`/(check|work|to-plan|brainstorm|handoff|watzup)\b' cli/docs/embedded docs/playbooks
        docs/WORKFLOW.md` 0 lines; `rg -n 'opus|sonnet|haiku' cli/docs/embedded` hits only the mapping table.
      - T2 rule fix: `rules/workflow-core.md` routes `check` at phase end and before a PR, not before
        every commit. check: `rg -n 'Before any commit' rules/workflow-core.md` 0 lines.
    - wave 2:
      - T3 single skill source: inspect `~/.agents/skills` and `~/.claude/skills`, show the owner the
        exact `trash`/`ln -s` commands, run them only on confirmation. check: `readlink ~/.claude/skills`
        prints the `~/.agents/skills` path; `ls ~/.claude/skills | wc -l` equals the source count.
    - phase check: gate commands; `bash scripts/verify-doc-links.sh`.

## Progress
- 2026-09-19T13:10Z — p1-drift-check — wave 1 — T1 — task_status=in-progress — phase start (run anchor 2026-09-19T13:10Z)
- 2026-09-19T13:40Z — p1-drift-check — wave 1 — T1 — task_status=DONE — `go test ./internal/installer/ -run Registry` pass (3 tests; TestMain isolates XDG_CONFIG_HOME) — cli/internal/installer/registry.go, registry_test.go, installer.go, update.go, uninstall.go
- 2026-09-19T13:40Z — p1-drift-check — wave 1 — T2 — task_status=DONE — `go test ./internal/installer/ ./internal/interfaces/ -run Check` pass (6 tests) — cli/internal/installer/check.go, check_test.go, cli/internal/interfaces/manage.go, manage_test.go
- 2026-09-19T13:40Z — p1-drift-check — wave 1 — summary — T1, T2 DONE
- 2026-09-19T13:55Z — p1-drift-check — wave 2 — T3 — task_status=DONE — `bash -n scripts/install-zharness.sh` ok; `bash scripts/test-install-zharness.sh` 4 passed — scripts/install-zharness.sh
- 2026-09-19T13:55Z — p1-drift-check — wave 2 — T4 — task_status=DONE — `bash scripts/test-guards.sh` 46 passed, 0 failed (4 new R16 cases); guard core diff vs HEAD empty — scripts/install-git-hooks.sh, scripts/test-guards.sh
- 2026-09-19T13:55Z — p1-drift-check — wave 2 — summary — T3, T4 DONE; phase checks: go test/vet/build/gofmt clean, doc links OK, `zh update --check --root ~/Lab/Ligaturizer` exit 1 with 11 drift lines and its `git status` unchanged
- 2026-09-19T08:31Z — p1-drift-check — gate — requests — task_status=DONE — (1) `canonicalRoot` (EvalSymlinks, raw-path fallback) in register/unregister + `TestRegistry_SymlinkedPathIsOneEntry`; (2) `--all` prints `error    <root>: <err>`, continues, exits non-zero + `TestRunCheck_AllContinuesPastUnreadableRoot`; (4) `mv … || return 1`; `go test ./...`, `go vet`, gofmt clean, test-guards 46 passed — cli/internal/installer/registry.go, registry_test.go, check.go, check_test.go, scripts/install-git-hooks.sh

## Decisions
- 2026-09-19 — lock — the installed `.git/hooks/pre-commit` was stale (3-argument call into the
  4-argument core since b90f354), so R2 compared against `/old.md` and failed open; reinstalled with
  `--force`, range re-check `master..HEAD` passed (R2 rc=0, PHASE-DONE rc=0) and an anchored negative
  probe was rejected (rc=1). Owner added R16.
- 2026-09-19 — lock — R2 skips entries whose first line has no anchored `verdict:`; the closed
  harness-eval-loop entries use that unanchored form. Not a guard change (NG3); R10 now requires the
  anchored form for new entries.
- 2026-09-19 — p1-drift-check/T4 — ownership of an existing hook is detected by the
  `ZHARNESS_HOOK_SOURCE=` line, not by matching the first two lines as planned: a future header edit
  would otherwise make every older hook of ours look foreign.
- 2026-09-19 — p1-drift-check — every phase gate runs as an independent subagent judge instead of
  in-session (`work-full.md` step 11), because guard R3 rejects `judge: same-session` on this
  high-risk plan.
- 2026-09-19 — p1-drift-check — Progress timestamps 13:10Z–13:55Z were estimated, not read from the
  clock; the real UTC time at the gate was ~08:29Z. Entries are left as written; from 08:31Z on,
  times come from `date -u`.
- 2026-09-19 — p1-drift-check/gate — request (3), `--check` combined with `--continue`/`--abort`, is not
  fixed here: P2 removes both flags (R6), which removes the combination.
- 2026-09-19 — p2-drop-threeway — planned deviations: the recorded base becomes the manifest's
  per-file sha256 (`loadBase` → path→sha; `seedPath` and `removeManagedFile` compare hashes), and
  the AGENTS hash guard reuses the manifest's existing `AGENTS.md` entry (already the canonical block
  sha) instead of a new `agents_block_sha256` field.

## Validation
- 2026-09-19T08:29Z — phase `p1-drift-check` — verdict: APPROVE_WITH_REQUESTS — mode: gate
  - `cd cli && go test ./... -count=1` — ok docs/embedded, internal/embedded, internal/installer, internal/interfaces
  - `cd cli && go vet ./...` — no findings
  - `test -z "$(gofmt -l cli)"` — exit 0, no unformatted files
  - `bash scripts/test-guards.sh` — guards: 46 passed, 0 failed (R16 fresh/current/stale-ours/foreign cases ok)
  - `bash scripts/verify-doc-links.sh` — doc links OK (0 findings)
  - `bash scripts/test-install-zharness.sh` — summary: 4 passed, 0 failed
  - `bash -c 'cmp <(git show master:scripts/install-git-hooks.sh | sed -n "11,347p") <(sed -n "11,347p" scripts/install-git-hooks.sh)'` — guard core (lines 11–347) byte-identical to master
  - `sh -c 'cd cli && go run ./cmd/zharness update --check --root "$HOME/Lab/Ligaturizer" >/dev/null 2>&1; test $? -eq 1'` — drift reported (11 lines), exit 1; Ligaturizer git status and managed-file hashes unchanged
  - scope: on target — R1, R2, R3 and R16 are implemented within the planned surfaces; guard core, playbooks, templates and consumer repos untouched
  - requests: (1) major — registry.go:65,81 / manage.go:17-26 unresolved symlink paths → duplicate registry entry, stale after uninstall; (2) minor — check.go:117 one unreadable root aborts --all; (3) minor — manage.go:71 --check ignores --continue/--abort; (4) nit — install-git-hooks.sh:498 mv failure not propagated
  - judge: independent
  - judge_model: claude-opus-5

## Current State and Next Action
- active_phase: p1-drift-check
- lifecycle_status: checked
- blockers: none
- open_items: none
- exact_next_action: work full p2-drop-threeway
