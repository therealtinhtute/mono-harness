---
id: 01M0ZHGV3PLANQK8M2VHZ5E9TA
type: plan
intake_id: 01M0ZHGV3INTK8M2VHZ5E9TA
lane: high-risk
status: active
created: 2026-08-28
updated: 2026-08-28
---

# Plan: zharness guard-v3 — deferred hardening batch

## Outcome

- result: the two deferred guard hardenings (O2 verdict line-anchoring, O3 undated-entry visibility) land in the pre-commit hook core; two installer robustness gaps close (safePath injective mapping with legacy fallback; diffHunks bounded memory); the non-spine `git`/`interview` skills stop referencing binary commands deleted in v0.15.
- success_signals:
  - S1 [verdict anchoring]: a new Validation entry whose sub-bullet prose contains `verdict: REQUEST_CHANGES` before the real first-line verdict no longer loses its APPROVED — the guard re-executes its proofs.
  - S2 [undated visibility]: a newly added APPROVED entry whose first line lacks a timestamp is still split, seen, and re-executed by the guard.
  - S3 [S2/S3 parity holds]: fail-case (bad proof command) rejects, pass-case (good proof) accepts — the pre-v3 behavior is preserved.
  - S4 [installer safety]: uninstall restore still finds captured originals recorded by the v0.15.0 mapping (legacy fallback); no two distinct managed paths can collide in `.zharness/base/original/`.
  - S5 [bounded merge]: a managed file above the LCS cap merges via the conservative fallback — no quadratic blowup, result still correct (whole-side hunk).
  - S6 [skill truth]: `grep -rn "zharness preflight\|zharness query\|zharness init\|zharness db" skills/workflow/git skills/workflow/interview` returns 0 hits.

## Authority and Requirements

- authority:
  - `docs/plans/completed/zharness-v015-slim.md` — open_items this batch resolves (guard v3 O2/O3, safePath collision, diffHunks LCS cliff), plus the recorded guard v2 contract its changes must preserve.
  - `scripts/install-git-hooks.sh` — ZGUARD-CORE, the sole fail-closed layer; CI re-runs it (`.github/workflows/cli-ci.yml` hook-guard job).
  - `cli/internal/installer/installer.go`, `cli/internal/installer/threeway.go`, `cli/internal/installer/uninstall.go` — safePath/readOriginal/captureOriginal and diffHunks.
  - `skills/workflow/git/`, `skills/workflow/interview/` — non-spine skills referencing deleted commands (`preflight`, `query check`, `init`, `db`).
- requirements:
  - R1 [accepted]: `zharness_anchored_verdict` reads the verdict token from the entry's FIRST LINE only. Sub-bullet prose mentioning verdict tokens can never select or shadow the entry's verdict. | source: open_items O2
  - R2 [accepted]: `zharness_dump_entries` starts a new entry at any UNINDENTED `- ` line inside `## Validation` (timestamp optional); indented lines always continue the current entry. Undated entries become visible to R2 and R3. | source: open_items O3
  - R3 [accepted]: `safePath` is injective over managed paths (single-pass replacer: `_` → `__`, `/` → `_2F`); `readOriginal`/`captureOriginal` fall back to the legacy v0.15.0 mapping so captured originals recorded before this change still protect uninstall. | source: open_items safePath; R12 spirit (consumer bytes never destroyed)
  - R4 [accepted]: `diffHunks` refuses the quadratic LCS matrix when `(n+1)*(m+1)` exceeds 8,000,000 cells and falls back to a single whole-side hunk `{0,n} -> other` — bounded memory, conservative merge semantics. | source: open_items diffHunks
  - R5 [accepted]: `skills/workflow/git` and `skills/workflow/interview` contain zero references to binary commands deleted in v0.15 (`preflight`, `query`, `init`, `db`); their guidance becomes markdown/procedure-only. | source: p0/p3 sweep deferrals
  - R6 [accepted]: the guard change keeps v2 semantics — new-entry = full-text hash set difference; proofs re-executed under `timeout 300 sh -c`; same-session rule unchanged; CI extraction unchanged. | source: v2 hardening comments in ZGUARD-CORE

## Non-goals

- NG1: no new binary verbs, no changes to the three-verb surface.
- NG2: no spine-skill rewrites (their preflight degradation lines are a separate, accepted concern).
- NG3: no Myers-diff rewrite of threeway (bounded fallback suffices).

## Phases and Verification
<!-- Phase and task definitions are immutable after to-plan. Append-only Progress is the sole task execution-status source. -->

- planning_status: planned
- phases:
  - phase_slug: g1-hardening
    story_id: 01M0ZHGV3PHAS8M2VHZ5E9TA
    status: planned
    goal: Land the guard v3 hardenings, installer robustness fixes, and skills sweep with all gates green and an independent full review.
    depends_on: none
    surfaces_allowed: scripts/install-git-hooks.sh, scripts/test-guards.sh, .github/workflows/cli-ci.yml, cli/internal/installer/**, skills/workflow/git/**, skills/workflow/interview/**, docs/plans/**, this plan
    surfaces_avoided: any consumer repository; spine skills; the release pipeline
    requirements: R1, R2, R3, R4, R5, R6
    waves:
      - wave: 1
        goal: Guards and installer land with fixtures proving v2 semantics held.
        tasks:
          - task: R1+R2 in scripts/install-git-hooks.sh — first-line verdict anchoring; unindented entry starts.
            verify: new `scripts/test-guards.sh` fixtures — S1 prose-shadow, S2 undated, S3 fail/pass parity all behave as required.
          - task: R3 in cli/internal/installer — injective safePath + legacy fallback in captureOriginal/readOriginal, with collision test.
            verify: go test — distinct paths never share an original file name; a legacy-mapped .orig is still found and restored.
          - task: R4 in cli/internal/installer/threeway.go — LCS cell cap with whole-side fallback, cap test.
            verify: go test — oversized input terminates fast, result is the documented fallback.
      - wave: 2
        goal: Skills sweep and CI wiring.
        tasks:
          - task: R5 sweep skills/workflow/git + interview of deleted-command references.
            verify: S6 grep returns 0 hits; doc links OK.
          - task: Wire scripts/test-guards.sh into the cli-ci hook-guard job after core extraction.
            verify: CI yaml parses; local run of the script exits 0.
    checks:
      - `bash scripts/verify-doc-links.sh`
      - `cd cli && go build ./... && go vet ./... && go test ./...`
      - `bash scripts/test-guards.sh` -> all fixtures green
      - S4 rg zero (`grep -rniE 'sqlite|harness\.db' cli/` -> 0)
      - `zharness --help` still lists exactly install, update, uninstall
      - full check (Security, Performance, Architecture, Code Quality) with an independent judge — lane high-risk, so the gate verdict itself must be independent

## Progress

<!-- Append-only durable entries record timestamp, phase, wave, task, task_status, run_id, trace_id, exact verification/result, and changed surfaces or blocker. -->
- `2026-08-28T07:45:00Z` — g1-hardening W1, task hook O2+O3. task_status: `DONE`. verification: `zharness_anchored_verdict` reads only the entry's first line (R1); `zharness_dump_entries` starts entries at any unindented `- ` line, timestamp optional (R2). Fixture suite `scripts/test-guards.sh` — 11 fixtures green, including: prose `verdict: REQUEST_CHANGES` in a sub-bullet no longer shadows a first-line APPROVED; undated APPROVED entry now re-executed (fail-case rejects); v2 parity held (sabotaged proof rejects, clean proof accepts, same-session on high-risk rejects).
- `2026-08-28T07:45:00Z` — g1-hardening W2, task skills sweep. task_status: `DONE`. verification: git/interview skills carry zero deleted-command references (`grep -rn "zharness preflight\|zharness query\|zharness init\|zharness db" skills/workflow/git skills/workflow/interview` = 0); doc links OK.
- `2026-08-28T08:00:00Z` — g1-hardening W1, task safePath + diffHunks. task_status: `DONE`. verification: `safePath` injective ('_'→"__", '/'→"_2F"; prefix-free code) with `legacySafePath` fallback in `findOriginal` — collision test green, legacy original found and never overwritten (TestSafePath_Injective_AndLegacyFallback); `diffHunks` falls back to one whole-side hunk above `lcsCellCap` 8M cells (TestDiffHunks_CapFallsBackToWholeSide), normal path unchanged.
- `2026-08-28T08:00:00Z` — g1-hardening W2, task CI wiring. task_status: `DONE`. verification: `cli-ci.yml` hook-guard job now runs `bash scripts/test-guards.sh` after core extraction.
- `2026-08-28T08:05:00Z` — g1-hardening gates. task_status: `DONE`. verification: go build/vet/test green (cli); test-guards 11/11; doc-links 0 findings; S4 zero; `zharness --help` unchanged (verified at v0.15.0 release smoke). Remaining check: independent full review (lane high-risk).
- none

## Decisions

<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->

- `2026-08-28` — execution. decision: the guard's verdict regex now accepts both `verdict \`X\`` and `verdict: X` forms, read from the entry's first line only. rationale: the fixtures exposed that the v2 regex demanded a colon while the repository's real entry grammar is "verdict `APPROVED`" — meaning the R2/R3 guards had silently SKIPPED every real APPROVED entry so far (no proofs were ever re-executed by the hook on this repo's plans; authoring sessions ran proofs themselves). Guard v3 fixes both the grammar and the anchoring, and `scripts/test-guards.sh` locks the behavior in CI.
- none

## Validation

<!-- Append-only durable entries record timestamp, phase, exact command/result/output, run_id, check_id, verdict, and proof_gaps. -->

- none

## Current State and Next Action

- active_phase: g1-hardening (both waves executed, all local gates green; remaining check — the independent full review)
- lifecycle_status: in-progress
- latest_run_id: hand-append (guards run from committed scripts; no binary rows)
- latest_trace_ids: none
- latest_check_id: none — full check pending (independent judge, lane high-risk)
- latest_handoff_id: none
- blockers: none
- open_items: none
- exact_next_action: run the independent full check (Security/Performance/Architecture/Code Quality) on the guard-v3 diff, then record the gate entry, flip g1-hardening done, and move the plan to completed
