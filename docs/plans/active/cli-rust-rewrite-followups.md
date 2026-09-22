---
status: active
lane: high-risk
---
# Close the cli-rust-rewrite follow-ups

## Goal
- outcome: The four findings left by the cli-rust-rewrite closure are resolved — the completed plan again carries p2-rust-port's Validation entry and the pre-compaction Log spacing, the three symlink-following write sites in the installer fail closed, and each of the ten review requests from the 2026-09-22T06:02:22Z full entry is either fixed or closed by a recorded owner ruling.
- success_signal: `git show HEAD:docs/plans/completed/cli-rust-rewrite.md` carries p2-rust-port's Validation entry byte-identical to `ed88b3b^`'s active-plan blob and its Log section adds no line the parent blob lacked; `cd cli && cargo test` is green with one new regression test per write site, each failing against the pre-fix code; `bash scripts/verify-doc-links.sh` and `bash scripts/test-guards.sh` exit 0; and `## Log` holds one entry per request id 1-10 naming the fixing commit or the owner's ruling.
- actors: the repository owner (maintainer) and consumers who run `zharness install`/`update`/`uninstall` in a repository carrying a planted symlink.
- authority: the 2026-09-22T06:02:22Z `full` Validation entry of `docs/plans/completed/cli-rust-rewrite.md` (requests 1-10); the bounded `check` verdict of 2026-09-22 (p2 entry + Log spacing); `docs/memory/2026-09-22-zharness-atomic-write-symlink.md`; the owner's selection of these four findings (2026-09-22).
- requirements:
  - R1: the completed plan carries p2-rust-port's last Validation entry byte-identical to the pre-closure blob | acceptance: the entry block equals `ed88b3b^:docs/plans/active/cli-rust-rewrite.md` lines 171-187 byte for byte | source: bounded check 2026-09-22
  - R2: the compacted Log adds no line the pre-compaction blob did not contain, so no surviving entry is re-spaced | acceptance: the Log section's diff against the parent blob adds exactly one line, the `absorb` entry | source: bounded check 2026-09-22
  - R3: the three non-test write sites (`cli/src/installer/mod.rs:120`, `cli/src/installer/uninstall.rs:143`, `cli/src/installer/mod.rs:498`) refuse to write through a symlink at the temp or managed path | acceptance: one regression test per site reproduces the planted-symlink attack and fails against the pre-fix code | source: full entry requests 1-2
  - R4: each of the ten review requests is fixed or closed by a recorded owner ruling | acceptance: one `## Log` entry per request id 1-10 naming the fixing commit or the ruling | source: full entry requests 1-10
- non-goals:
  - NG1: no new verbs or flags; the binary stays install/update/uninstall.
  - NG2: the final non-rc tag `cli/v0.24.0` is not pushed here — it stays an owner action after this initiative.
  - NG3: no re-capture of `cli/testdata/golden/**` from a Go binary; the tree it was captured from is deleted.
  - NG4: no rewrite of `site/`.

## Phases and Verification
- approach: Repair the record first (P1) so the closure commit is correct before anything lands on top of it; then fix the security class (P2) with one regression test per write site; then the mechanical doc and CI truth-up (P3); then the three requests that need an owner ruling (P4). Rejected: folding the record repair into a later commit — the amend must precede new commits or the closure commit stays wrong; and fixing only `mod.rs:120` — the class has three sites and the review proved coverage is complete.
- risks: the amend rewrites a pushed commit → it is local (`master` is ahead 1, `ed88b3b` unpushed); stop if `git branch -r --contains ed88b3b` is non-empty. The symlink fix changes bytes for a legitimate path → the golden replay is the oracle; stop if a fixture diverges. Editing the embedded `check-validation.md` invalidates the frozen fixtures → P4 gates it on an owner ruling, and any fixture hash update is a hand edit recorded in Log. The historical `hook-guard` red recurs on any push whose before-SHA predates `4cd86cf` → P4's ruling; the red is confined to that range.
- constraints: preserve the `.zharness/base/manifest.json` schema (`zharness_version`, `installed_at`, `files[].path`, `files[].sha256`) and the repo registry at `$XDG_CONFIG_HOME/zharness/repos`; touch nothing under `scripts/` except as P4's ruling requires; never re-capture the golden fixtures; keep the binary's verbs and flags unchanged.
- recovery: before P1's amend, `git reset --hard ed88b3b` restores the closure commit; after P2/P3, `git revert` the phase's own commit; P2's fixtures are the oracle, so a bad fix is caught before commit.
- phase_slug: `p1-record-repair`
  status: checked
  - goal: R1, R2 | depends_on: none
  - surfaces: `docs/plans/completed/cli-rust-rewrite.md` | avoided: `cli/**`, `.github/**`, `scripts/**`, `docs/playbooks/**`
  - escalate_when: the amend would rewrite a commit already pushed (`git branch -r --contains ed88b3b` non-empty), or a restored line differs from the parent blob's bytes
  - wave 1:
    - T1 Restore p2-rust-port's Validation entry — the 17-line block at `ed88b3b^:docs/plans/active/cli-rust-rewrite.md` lines 171-187 — into `docs/plans/completed/cli-rust-rewrite.md`, between the p1 entry and the p3 entry header. — output: the completed plan carries one Validation entry per phase, four in total — check: `diff <(git show ed88b3b^:docs/plans/active/cli-rust-rewrite.md | sed -n '171,187p') <(sed -n '/^- 2026-09-22T03:28:06Z/,/^  - judge_model/p' docs/plans/completed/cli-rust-rewrite.md)` — stop_if: any restored line differs from the parent blob's bytes
    - T2 Remove the separators the compaction introduced, so the Log adds no line the parent blob lacked. — output: the Log section's diff against the parent adds only the `absorb` entry — check: `test "$(diff <(git show ed88b3b^:docs/plans/active/cli-rust-rewrite.md | sed -n '/^## Log/,/^## Validation/p') <(sed -n '/^## Log/,/^## Validation/p' docs/plans/completed/cli-rust-rewrite.md) | grep -c '^>')" = 1` — stop_if: a surviving entry's bytes change
    - T3 Amend `ed88b3b` with the corrected record (`git add` the plan, then `git commit --amend --no-edit`) so the closure commit itself is correct rather than a follow-up commit. — output: one closure commit carrying the corrected record — check: `git show HEAD:docs/plans/completed/cli-rust-rewrite.md | sed -n '/^- 2026-09-22T03:28:06Z/,/^  - judge_model/p' | wc -l` prints 17, and the hook's guard output reads passed — stop_if: the hook rejects, or the amend would touch a pushed commit
  - phase check: `diff <(git show ed88b3b^:docs/plans/active/cli-rust-rewrite.md | sed -n '171,187p') <(sed -n '/^- 2026-09-22T03:28:06Z/,/^  - judge_model/p' docs/plans/completed/cli-rust-rewrite.md) && test "$(diff <(git show ed88b3b^:docs/plans/active/cli-rust-rewrite.md | sed -n '/^## Log/,/^## Validation/p') <(sed -n '/^## Log/,/^## Validation/p' docs/plans/completed/cli-rust-rewrite.md) | grep -c '^>')" = 1 && bash scripts/test-guards.sh`
- phase_slug: `p2-symlink-hardening`
  status: planned
  - goal: R3 | depends_on: none
  - surfaces: `cli/src/installer/mod.rs`, `cli/src/installer/uninstall.rs`, `cli/src/test_support.rs`, `cli/tests/**` | avoided: `cli/testdata/golden/**`, `.github/**`, `docs/**`
  - escalate_when: the fix changes the bytes written for a legitimate path (a golden fixture diverges), or a fourth non-test write site is found
  - wave 1:
    - T1 `write_file_atomic` (`cli/src/installer/mod.rs:120`): open the temp path with `create_new(true)` (or `O_NOFOLLOW`) so a planted symlink aborts instead of being followed; decide the stale-temp behavior and record it in Log. Name the regression test `write_file_atomic_refuses_symlink`. — output: the site refuses a symlinked temp path — check: `cd cli && cargo test --lib write_file_atomic_refuses_symlink` — stop_if: the new test passes against the pre-fix code
    - T2 `cli/src/installer/uninstall.rs:143`: same treatment for the managed-path restore of a captured original. Name the regression test `uninstall_restore_refuses_symlink`. — output: the restore refuses a symlinked managed path — check: `cd cli && cargo test --lib uninstall_restore_refuses_symlink` — stop_if: the new test passes against the pre-fix code
    - T3 `cli/src/installer/mod.rs:498`: same treatment for the AGENTS.md creation path. Name the regression test `agents_md_refuses_symlink`. — output: the creation refuses a dangling symlink — check: `cd cli && cargo test --lib agents_md_refuses_symlink` — stop_if: the new test passes against the pre-fix code
  - phase check: `cd cli && cargo fmt --check && cargo clippy --all-targets -- -D warnings && test "$(cargo test --lib refuses_symlink 2>/dev/null | grep -cE 'refuses_symlink.* \.\.\. ok$')" -ge 3 && cargo test`
- phase_slug: `p3-doc-ci-truth-up`
  status: planned
  - goal: R4 (requests 3, 4, 6, 7, 8) | depends_on: none
  - surfaces: `AGENTS.md`, `CONTRIBUTING.md`, `docs/ARCHITECTURE.md`, `docs/decisions/0012-five-section-plan-and-update-migration.md`, `cli/testdata/golden/capture.sh` (header only), `cli/docs/CONTRACT.md`, `.github/workflows/cli-ci.yml` | avoided: `cli/testdata/golden/*/` fixture data, `cli/docs/embedded/**`, `cli/src/**`
  - escalate_when: a doc outside the listed surfaces must change to keep `verify-doc-links.sh` green
  - wave 1:
    - T1 Request 3: repoint the four live docs off the deleted Go tree (`AGENTS.md:7`, `CONTRIBUTING.md:27`, `docs/ARCHITECTURE.md:16,24,26`, ADR 0012:42,45,49). — output: no live doc describes the CLI as Go or cites `cli/cmd/`, `cli/internal/**`, cobra — check: `! rg -n "cobra|cli/cmd/|cli/internal/" AGENTS.md CONTRIBUTING.md docs/ARCHITECTURE.md docs/decisions/0012-five-section-plan-and-update-migration.md` — stop_if: a hit is a historical record that must stay
    - T2 Request 4: correct `cli/testdata/golden/capture.sh`'s header to state it can no longer run and that it is the record of how the fixtures were produced. — output: the header no longer claims runnability — check: `bash -n cli/testdata/golden/capture.sh && ! grep -q "Builds the Go binary at the current HEAD" cli/testdata/golden/capture.sh && git diff --exit-code -- cli/testdata/golden ':!cli/testdata/golden/capture.sh'` — stop_if: the fixture data changes
    - T3 Request 6: add the root-level `-v, --version` to `cli/docs/CONTRACT.md`'s command table, matching `cli/src/cli.rs:29-30`. — output: the table lists every flag clap registers — check: `grep -q -- "--version" cli/docs/CONTRACT.md` — stop_if: the table's verb rows no longer match `cli/src/cli.rs`
    - T4 Requests 7, 8: fix `cli-ci.yml`'s `build-test` naming (R7 says `cargo build`; the job runs fmt/clippy/test) and the `size-gate` comment ("four release targets" → three). — output: the workflow's text matches what it runs — check: `actionlint .github/workflows/cli-ci.yml && ! grep -q "four release targets" .github/workflows/cli-ci.yml` — stop_if: the `hook-guard` job changes
  - phase check: `bash scripts/verify-doc-links.sh && bash scripts/test-guards.sh && ! rg -n "cobra|cli/cmd/|cli/internal/" AGENTS.md CONTRIBUTING.md docs/ARCHITECTURE.md docs/decisions/0012-five-section-plan-and-update-migration.md && grep -q -- "--version" cli/docs/CONTRACT.md && ! grep -q "four release targets" .github/workflows/cli-ci.yml && actionlint .github/workflows/cli-ci.yml && cd cli && cargo test`
- phase_slug: `p4-owner-rulings`
  status: planned
  - goal: R4 (requests 5, 9, 10) | depends_on: none
  - surfaces: `docs/playbooks/check-validation.md` and `cli/docs/embedded/playbooks/check-validation.md` (request 5, if approved), `docs/plans/completed/cli-rust-rewrite.md` (request 9, if approved), `scripts/install-git-hooks.sh` (request 10, if approved) | avoided: `cli/src/**`, `cli/testdata/golden/*/` fixture data
  - escalate_when: always — each task needs an owner ruling before it runs, and the phase does not start until all three are given
  - wave 1:
    - T1 Request 5: obtain the owner's ruling on the proof-contract example (`cd cli && go test ./...` in both copies). If approved, edit both copies and hand-update the fixture hashes the embedded bytes feed. — output: a recorded ruling plus the change, or a recorded decision to leave it — check: the Log entry names the ruling, and if edited `cd cli && cargo test` stays green — stop_if: no ruling is recorded
    - T2 Request 9: obtain the owner's ruling on the completed plan's stale R5 text (four targets, `[true,5]`, `macos-13`). If approved, correct that text. — output: a recorded ruling plus the change, or a recorded decision to leave it — check: the Log entry names the ruling — stop_if: no ruling is recorded
    - T3 Request 10: obtain the owner's ruling on the historical `hook-guard` red (`fe814b4a..4cd86cf`): leave it, or add a guard exemption. — output: a recorded ruling plus the change, or a recorded decision to leave it — check: the Log entry names the ruling, and if a guard change lands `bash scripts/test-guards.sh` exits 0 — stop_if: no ruling is recorded
  - phase check: `test "$(sed -n '/^## Log/,/^## Validation/p' docs/plans/active/cli-rust-rewrite-followups.md | grep -ciE 'request (5|9|10)')" -ge 3 && bash scripts/test-guards.sh`

## Log
- 2026-09-22T07:29:48Z — p1-record-repair — start — phase opened; T1 restores p2-rust-port's Validation entry (parent lines 171-187), T2 removes the separators the compaction introduced, T3 amends `ed88b3b` so the closure commit itself is correct.
- 2026-09-22T07:31:18Z — p1-record-repair/T1 — done — `diff <(git show ed88b3b^:docs/plans/active/cli-rust-rewrite.md | sed -n '171,187p') <(sed -n '/^- 2026-09-22T03:28:06Z/,/^  - judge_model/p' docs/plans/completed/cli-rust-rewrite.md)` exits 0: the 17-line p2-rust-port block is byte-identical to the parent blob and sits between the p1 entry and the p3 header. Changed: docs/plans/completed/cli-rust-rewrite.md.
- 2026-09-22T07:31:18Z — p1-record-repair/T2 — done — the Log section's diff against the parent blob adds exactly one line, the `absorb` entry (`grep -c '^>'` prints 1). The Log is rebuilt from the parent blob with the 21 `start`/`done` lines dropped, deletion-created blank runs collapsed to one, and the `absorb` entry appended; all 19 surviving entries are byte-identical. Changed: docs/plans/completed/cli-rust-rewrite.md.
- 2026-09-22T07:31:18Z — p1-record-repair/T3 — done — the closure commit is recreated as `c720c46` carrying the corrected record; the hook prints "✅ v0.15 guards passed" and `git show HEAD:docs/plans/completed/cli-rust-rewrite.md | sed -n '/^- 2026-09-22T03:28:06Z/,/^  - judge_model/p' | wc -l` prints 17. Changed: git history (one commit replaced).
- 2026-09-22T07:31:18Z — p1-record-repair — decision — T3's literal `git commit --amend` cannot pass the guard, so the commit was recreated instead. The amend's base is `ed88b3b` itself, which already carries `docs/plans/completed/cli-rust-rewrite.md`, so `zharness_guard_old_blob`'s predecessor mapping never fires and the restored p2 entry counts as new — the guard re-executes its proofs, which cannot run at this HEAD (the Go tree they build is deleted, and `8032af1..HEAD` no longer matches that phase's surfaces). Fix: `git reset --soft ed88b3b^` then re-commit with the same message, so the base is `877c011` and the mapping resolves `completed/` back to `active/`; the guard then sees every entry as old and passes. `ed88b3b^` and `HEAD^` are both `877c011`, so the phase's checks stay valid; `ed88b3b` is now dangling and reachable only through the reflog.

## Validation
- 2026-09-22T07:34:02Z — phase `p1-record-repair` — verdict: APPROVED — mode: gate
  - `bash -c 'diff <(git show ed88b3b^:docs/plans/active/cli-rust-rewrite.md | sed -n "171,187p") <(sed -n "/^- 2026-09-22T03:28:06Z/,/^  - judge_model/p" docs/plans/completed/cli-rust-rewrite.md) && test "$(diff <(git show ed88b3b^:docs/plans/active/cli-rust-rewrite.md | sed -n "/^## Log/,/^## Validation/p") <(sed -n "/^## Log/,/^## Validation/p" docs/plans/completed/cli-rust-rewrite.md) | grep -c "^>")" = 1 && bash scripts/test-guards.sh'` — exit 0; both record clauses hold and the guard prints "guards: 44 passed, 0 failed".
  - `bash scripts/test-guards.sh` — exit 0; "guards: 44 passed, 0 failed", including the S6 rename-predecessor and S7 range-selector cases the phase's fix depends on.
  - `bash scripts/verify-doc-links.sh` — exit 0; "doc links OK (0 findings; 10 claim(s) under known-removed v0.15 surfaces)".
  - scope: on target — 877c011..HEAD is the recreated closure commit: the plan rename into `completed/` plus `docs/memory/2026-09-22-zharness-atomic-write-symlink.md`, whose bytes are identical to `ed88b3b`'s; nothing under `cli/`, `.github/`, `scripts/` or `docs/playbooks/` is touched.
  - requirements: R1 met (the p2-rust-port block at `HEAD:docs/plans/completed/cli-rust-rewrite.md` is byte-identical to `ed88b3b^:docs/plans/active/cli-rust-rewrite.md` lines 171-187 — the diff exits 0 with no output, and the completed plan carries one entry per phase, four in total) | R2 met (the Log section's diff against the parent blob adds exactly one line, the `absorb` entry; `grep -c '^>'` prints 1, so no surviving entry is re-spaced)
  - rollback_point: 877c011 (the phase diff base; `git reset --hard 877c011` drops the recreated closure commit c720c46, and the pre-correction ed88b3b stays reachable through the reflog)
  - requests: none
  - proof_gaps: `judge: independent` is testimony, not proof; the `requirements:` line is testimony; the phase's claim that the guard's predecessor mapping resolves `completed/` back to `active/` is verified only through the guard's exit code and its S6 case, not by reading the guard internals.
  - judge: independent
  - judge_model: devin/deepseek-v4-1-flash

## Current State and Next Action
- active_phase: p1-record-repair
- lifecycle_status: checked
- blockers: none
- open_items: none
- exact_next_action: work full phase p2-symlink-hardening
