---
id: 01M0EG9ZS3VVCF05EEMAW5S685
type: plan
intake_id: 01M0EG9T31545JJFK8HBJJB0XA
lane: normal
status: completed
created: 2026-08-20
updated: 2026-08-20
---

# Plan: close the eight review findings on PR #60

## Outcome
- result: every finding raised in the review of PR #60 (`docs/docs-architecture-plan` → `master`) is either fixed with executable proof or explicitly rejected in writing. No consumer repository can lose a file to `zharness init` or `db migrate-layout`, and no document in this repository describes behaviour the same PR changed.
- success_signals:
  - `bash scripts/verify-doc-links.sh` exits zero and `cd cli && go test ./...` passes all packages.
  - A failed `db migrate-layout` leaves the working tree byte-identical to its pre-migration state, proven by a test that fails against the current `snapshotManagedTargets`.
  - No code path can replace an entire consumer-owned `CLAUDE.md` with a three-line import block.
  - `docs/workflow-harness/migration.md` names every file `zharness init` writes or mutates, including `CLAUDE.md` and the three scaffold-once docs.
  - Every `path:line` citation added or touched by PR #60 resolves to the line it claims, honouring `docs/decisions/README.md`'s own "resolves at merge time" promise.

## Authority and Requirements
- authority:
  - Review of PR #60, 2026-08-20. Gates were run independently at review time and both passed — `bash scripts/verify-doc-links.sh` reported 0 findings, `go test -C cli ./...` passed 6/6 packages. Embedded ↔ projected playbook parity was checked byte-for-byte (6/6 identical) and the `AGENTS.md` managed block matches its embedded source. **These findings are therefore not gate failures; every one of them is invisible to the current gates.** That is itself the reason they need a plan rather than a re-run.
  - Verified 2026-08-20 by reading source, not by trusting the review:
    - `cli/internal/application/layout_migration.go:297` — `snapshotManagedTargets` seeds `paths` with exactly `AGENTS.md` and `.gitignore`, then walks `docsFS` adding only `docs/`-relative entries. `CLAUDE.md` and the three `scaffoldOnceDocs` paths are absent.
    - `cli/internal/application/layout_migration.go:135` — `ScaffoldDocs` is called *after* the snapshot is taken, and it now writes `CLAUDE.md` (`init.go:56`) and the scaffold-once set (`init.go:62`). Six later failure points each call `restoreFileSnapshots`: `:143` WAL checkpoint, `:149` `tempDB.Close`, `:154` `legacyDB.Close`, `:159`/`:164`/`:169` legacy reopen and checkpoint, `:175` backup rename, `:179` activation rename, `:184` backup removal.
    - `cli/internal/application/layout_migration.go:330-346` — `restoreFileSnapshots` already handles the absent-before case by removing the path, so extending the snapshot list is sufficient; no new restore logic is required.
    - `cli/internal/application/init.go:86` — the legacy probe reads `filepath.Join(root, ".kit", "docs", relPath)`, parameterised by `relPath`, so it now fires for `CLAUDE.md` as well as `AGENTS.md`.
    - `cli/internal/application/init.go:102-103` — when `legacyMatches` is true the branch sets `updated = block + "\n"`, discarding the entire prior file. For `CLAUDE.md` the block body is `@AGENTS.md` (`init.go:18`), three lines total.
    - `cli/internal/application/init.go:60` — `result.AgentsShimWritten = agentsWritten || claudeWritten`, a single boolean for two distinct files.
    - `cli/internal/interfaces/init.go:95-96` — the sole consumer prints the fixed string `updated AGENTS.md managed block`.
    - `docs/README.md:14` links `plans/active`; `docs/plans/active/` does not exist in the tree (`ls` reports no such directory) because `zharness plan complete` moved the last plan out and git does not track empty directories.
    - `scripts/verify-doc-links.sh:64` matches only backticked paths ending `.md|.sh|.go|.json|.yml|.toml|.py`. Directory citations and markdown link targets are structurally invisible to the gate, which is why finding 5 passed CI.
    - `skills/workflow/git/references/workflow.md:1` reads `# Playbook: git`, while `skills/workflow/git/SKILL.md:13` states it is not a harness-projected playbook and `docs/README.md:31` states `git` is absent from `docs/playbooks/` by design.
    - `cli/internal/interfaces/preflight.go:23-31` — `preflightPlaybooks` declares six stages and no longer contains `git`.
    - `docs/ARCHITECTURE.md:56` cites `cli/internal/application/init.go:33`; the `SyncManagedDocs` call opens at `:32` (`:33` is the bare `db,` argument). `docs/ARCHITECTURE.md:74` cites `cli/internal/interfaces/preflight.go:30`; the `preflightPlaybooks` declaration is at `:23` and `:30` is now the closing brace, shifted by this PR's removal of the `git` entry. `docs/workflow-harness/migration.md:29` repeats the same `init.go:33`.
  - `docs/plans/completed/docs-architecture.md` R13 — the git.md deprojection deliberately adds no prune path, and argues the orphaned projection is "inert once nothing routes to it". That reasoning is accepted and is not reopened here; what is missing is only the consumer-facing instruction to delete the stale copy.
  - `docs/decisions/README.md` — "every structural claim carries a `path:line` citation that resolves at merge time". This is the repository's own standard and it is what makes finding 8 a defect rather than a nit.
  - Existing test surface to extend rather than invent: `cli/internal/application/layout_migration_test.go:83` (`TestMigrateLayoutDocsConflictRollsBackActivation`) already drives a failing migration and asserts rollback; `cli/internal/application/init_test.go:134,154` already cover the `CLAUDE.md` import paths.
- requirements:
  - R1 [accepted]: `snapshotManagedTargets` snapshots every path `ScaffoldDocs` can write — adding `CLAUDE.md` and the three entries of `scaffoldOnceDocs`. The scaffold-once paths are read from the `scaffoldOnceDocs` variable, not re-listed as string literals, so a future fourth entry cannot silently fall out of the rollback set. | source: `layout_migration.go:297`, `init.go:119-123`
  - R2 [accepted]: a test proves the rollback gap. It must fail against the current `snapshotManagedTargets` and pass after R1: run a migration that fails after `ScaffoldDocs` succeeds, then assert the tree is byte-identical to its pre-migration state — including that a `CLAUDE.md` absent before is absent after, and that `docs/README.md` absent before is absent after. A test that only asserts the new code path exists does not satisfy this. | source: `layout_migration_test.go:83`
  - R3 [accepted]: the legacy `.kit/docs/` probe is pinned to `AGENTS.md` and cannot fire for any other file. The whole-file-replace branch is unreachable for `CLAUDE.md`. | source: `init.go:86,102-103`
  - R4 [accepted]: a test proves R3: a repository holding a consumer-owned `.kit/docs/CLAUDE.md` byte-identical to its root `CLAUDE.md` keeps the root file's content after `ScaffoldDocs`, with the managed block appended rather than substituted. | source: `init.go:102-103`
  - R5 [accepted]: `zharness init` reports the file it actually changed. A run that writes `CLAUDE.md` and leaves `AGENTS.md` untouched must not print `updated AGENTS.md managed block`. | source: `init.go:60`, `interfaces/init.go:95-96`
  - R6 [accepted]: `docs/workflow-harness/migration.md` lists every mutation `zharness init` performs on a consumer repository — the projected `docs/` set, the `AGENTS.md` managed block, the `CLAUDE.md` managed block, and the three scaffold-once files — and states that the scaffold-once files are written only when absent and never refreshed. | source: `init.go:29-72`
  - R7 [accepted]: `docs/workflow-harness/migration.md` tells an already-initialized consumer to delete the now-orphaned `docs/playbooks/git.md`, states that the harness will not remove it, and names `skills/workflow/git/references/workflow.md` as where the procedure now lives. | source: `managed_docs.go:107`, `docs/plans/completed/docs-architecture.md` R13
  - R8 [accepted]: the `docs/README.md` link to the active-plan directory resolves on GitHub. | source: `docs/README.md:14`
  - R9 [accepted]: `skills/workflow/git/references/workflow.md` no longer titles itself a playbook, and says in its own opening lines that it is the git skill's procedure and not a harness-projected playbook — so an agent that opens the file directly cannot mistake it for one. | source: `skills/workflow/git/SKILL.md:13`, `docs/README.md:31`
  - R10 [accepted]: every `path:line` citation in `docs/ARCHITECTURE.md` and `docs/workflow-harness/migration.md` resolves to the construct it names, verified by reading the cited line after all code changes in this plan have landed. | source: `docs/decisions/README.md`
  - R11 [accepted]: both gates pass at the end of every phase — `bash scripts/verify-doc-links.sh` exits zero, `cd cli && go test ./...` passes. | source: `CLAUDE.md` gate-commands section

## Non-goals
- NG1: no prune path is added to `managed_docs.go`. `docs-architecture.md` R13 decided deliberately that `planManagedDocActions` walks the embedded FS and never visits an orphaned row, and proved the stale file inert. This plan documents the consequence; it does not reverse the decision.
- NG2: `scripts/verify-doc-links.sh` is not extended to check directory citations or markdown link targets. Widening the gate is a real initiative with its own false-positive budget across every file in the repository; finding 5 is fixed at the citation, not at the checker. Recorded here so the gap is a known deferral rather than an oversight.
- NG3: no change to `preflightPlaybooks`, `contextEligibleStages`, or the `git` stage's `reduced` preflight mode. The review confirmed `preflight git` resolves correctly.
- NG4: no re-litigation of the git.md deprojection, the scaffold-once class, or the `CLAUDE.md` import mechanism. Those were decided in `docs-architecture.md` and shipped; this plan fixes their edges only.
- NG5: no new release is cut as part of this plan. Whether these changes ship as `cli/v*` is a separate call made at `git` time.
- NG6: no reformatting, restructuring, or "while I'm here" editing of `migration.md`, `ARCHITECTURE.md`, or `docs/README.md`. Only the sentences named in the requirements change.

## Approach and Risks
- approach: fix the Go defects first with tests that fail before the fix, then repair the documents — in that order, because two of the citations this plan must correct point into the very files the Go phase edits. Fixing docs first would mean re-verifying every line number twice. Each finding maps to exactly one requirement and one task; nothing is bundled.
- why preferred: the three Go findings are the only ones that can destroy consumer data, and each has an existing test file to extend rather than a new harness to build. The five doc findings are single-sentence edits whose only difficulty is ordering. Splitting on that dependency is the smallest split that is still honest.
- rejected alternatives:
  - One phase for all eight. Rejected: the ARCHITECTURE.md citations cannot be verified until `init.go` stops moving, so a single phase would need an internal ordering rule anyway — which is what a phase boundary already is.
  - One phase per finding (eight phases). Rejected: the five doc edits share one verification command and no dependencies between them; separate phases would add ceremony without adding proof.
  - Fix only the medium/high findings and file the rest. Rejected: the user asked for all review comments, and the three low doc findings are each a one-line edit.
  - Widen `verify-doc-links.sh` so finding 5 is caught structurally. Rejected into NG2 — the gate change is larger than everything else in this plan combined.
- risks:
  - **R-A — the rollback test passes for the wrong reason.** A test that constructs its failure before `ScaffoldDocs` runs would pass against the unfixed code and prove nothing. Mitigation: mandatory red-then-green — run the new test against unmodified `snapshotManagedTargets` and record the failure output in `## Progress` before writing the fix. If it does not fail first, the test is wrong, not the code.
  - **R-B — pinning the legacy probe breaks legacy `AGENTS.md` adoption.** `AGENTS.md` genuinely relies on the `.kit/docs/` probe. Mitigation: change only the path the probe reads, never the branch logic, and confirm the existing `AGENTS.md` tests in `init_test.go` still pass unchanged.
  - **R-C — splitting the init report changes JSON output.** `interfaces/init.go` has a `--json` branch; altering `ScaffoldResult` could change a consumer-visible key. Mitigation: read the JSON branch before editing and keep its shape unless the requirement forces otherwise; state the decision in `## Progress`.
  - **R-D — `.gitkeep` in `docs/plans/active/` collides with plan resolution.** Mitigation: `preflightActivePlanGlob` is `docs/plans/active/*.md` (`preflight.go:20`), so a dotfile cannot match; confirm with `zharness preflight to-plan --json` and `zharness query plan --json` after adding it, and check `ResolveActivePlan`'s six call sites are glob-based before committing.
  - **R-E — corrected citations drift again on the next edit.** Unavoidable in principle. Mitigation: cite the declaration line of a named construct, never a line inside a call's argument list — `init.go:33` broke precisely because it pointed at a bare `db,`.
- stop conditions: stop and ask if the rollback test cannot be made to fail against the current code (means the finding is wrong and the fix would be speculative); if pinning the legacy probe breaks any existing `AGENTS.md` test (means R-B is real and the probe carries load not yet understood); or if `ScaffoldResult` is consumed anywhere outside `interfaces/init.go` and `layout_migration.go` (means R-C is wider than a print statement).
- recovery: every change is confined to the working tree until `check` passes. Revert the phase's commits and re-enter at the failed task; no harness state is mutated by these edits beyond the plan's own lifecycle rows.

## Phases and Verification

### Phase 1 — `pr60-go-correctness`
- story: `01M0EG9ZS5XTQJV3J5J2CZ689B`
- status: done
- depends on: none
- goal: close the three Go defects — incomplete migration rollback, the whole-file-replace legacy probe reaching `CLAUDE.md`, and the mislabelled init report line.
- covers: R1, R2, R3, R4, R5, R11
- surfaces allowed: `cli/internal/application/layout_migration.go`, `cli/internal/application/init.go`, `cli/internal/interfaces/init.go`, `cli/internal/application/layout_migration_test.go`, `cli/internal/application/init_test.go`
- surfaces avoided: `cli/docs/embedded/**` (projected content, needs a release), `cli/internal/application/managed_docs.go` (NG1), `cli/internal/interfaces/preflight.go` (NG3), everything under `docs/` (Phase 2 owns it)

**Wave 1 — prove the two data-loss defects fail today**

Both tasks are independent; each writes a test and records its red output before any production code changes.

- T1.1 — write the migration rollback test in `layout_migration_test.go`, modelled on `TestMigrateLayoutDocsConflictRollsBackActivation:83`. Arrange a repo with no `CLAUDE.md` and no `docs/README.md`, force a failure at a point *after* `ScaffoldDocs` returns, then assert both paths are absent afterwards and that `AGENTS.md` and `.gitignore` are byte-identical to their pre-migration bytes.
  - expected output: a failing test naming the leftover `CLAUDE.md` (and/or `docs/README.md`).
  - verify: `cd cli && go test ./internal/application/ -run TestMigrateLayout -v` — **must fail**, and the failure text goes into `## Progress`. A pass here means the test does not reach the gap; fix the test, not the code.
  - stop: if no failure point after `ScaffoldDocs` can be triggered from a test, stop and report — the finding's scenario may be unreachable.
- T1.2 — write the legacy-probe test in `init_test.go`: root `CLAUDE.md` with real consumer content, plus `.kit/docs/CLAUDE.md` byte-identical to it. Assert after `ScaffoldDocs` that the consumer content is still present and the managed block was appended.
  - expected output: a failing test showing the root file reduced to the three-line managed block.
  - verify: `cd cli && go test ./internal/application/ -run TestScaffoldDocs -v` — **must fail** on the new test only; every pre-existing `TestScaffoldDocs*` still passes.
  - stop: if the new test passes unfixed, the branch is unreachable — report before changing `init.go`.

**Wave 2 — fix, gated on Wave 1 being red**

- T2.1 — extend `snapshotManagedTargets` (`layout_migration.go:297`) to include `CLAUDE.md` and every path in `scaffoldOnceDocs`, iterating the variable rather than repeating literals (R1).
  - verify: `cd cli && go test ./internal/application/ -run TestMigrateLayout -v` — T1.1 now passes, `TestMigrateLayoutDryRunThenApply` and `TestMigrateLayoutDocsConflictRollsBackActivation` unchanged.
- T2.2 — pin the legacy probe at `init.go:86` to `AGENTS.md`, leaving the branch logic at `:102-103` untouched (R3).
  - verify: `cd cli && go test ./internal/application/ -run TestScaffoldDocs -v` — T1.2 passes; every pre-existing scaffold test passes with no edits to it. Any pre-existing test needing modification means R-B fired: stop and report.
- T2.3 — split the report so `zharness init` names the file it changed (R5). Read the `--json` branch in `cli/internal/interfaces/init.go` first; if `ScaffoldResult` is consumed anywhere beyond `interfaces/init.go` and `layout_migration.go`, stop and report (R-C).
  - verify: `cd cli && go build ./... && go test ./...`; then in a scratch directory outside this repository, run `zharness init` on a fixture with a current `AGENTS.md` and no `CLAUDE.md` and confirm the printed line names `CLAUDE.md`.

**Wave 3 — phase gate**

- T3.1 — run both repository gates and route to `check`.
  - verify: `cd cli && go test ./...` all packages ok; `bash scripts/verify-doc-links.sh` exits zero.
  - escalation: any failure returns to the owning wave; do not proceed to Phase 2 with a red gate.

### Phase 2 — `pr60-doc-truth`
- story: `01M0EG9ZSERMRZVPE860XRCJPR`
- status: done
- depends on: `pr60-go-correctness` — R10 verifies citations into `init.go`, whose line numbers Phase 1 may shift.
- goal: make every document claim that PR #60 rendered stale true again.
- covers: R6, R7, R8, R9, R10, R11
- surfaces allowed: `docs/workflow-harness/migration.md`, `docs/README.md`, `docs/ARCHITECTURE.md`, `skills/workflow/git/references/workflow.md`, `docs/plans/active/.gitkeep`
- surfaces avoided: all `.go` files (Phase 1 owns them), `cli/docs/embedded/**`, `scripts/verify-doc-links.sh` (NG2), `skills/workflow/git/SKILL.md` (already correct)

**Wave 1 — the two consumer-facing corrections**

Independent; both edit `migration.md` but in different paragraphs.

- T1.1 — rewrite the `migration.md:29` sentence to name every mutation `init` performs: the projected `docs/` set, the `AGENTS.md` block, the `CLAUDE.md` block, and the three scaffold-once files with their write-only-when-absent semantics (R6).
  - verify: `grep -n "CLAUDE.md" docs/workflow-harness/migration.md` returns the new sentence; read `cli/internal/application/init.go:29-72` alongside it and confirm no write it performs is unlisted.
- T1.2 — add the git.md deprojection note to `migration.md`: an already-initialized consumer keeps a stale `docs/playbooks/git.md` and its `managed_docs` row after upgrading, the harness will not remove either, delete the file by hand, and the procedure now lives at `skills/workflow/git/references/workflow.md` (R7).
  - verify: `grep -n "playbooks/git.md" docs/workflow-harness/migration.md` returns the note; `bash scripts/verify-doc-links.sh` exits zero.

**Wave 2 — the three routing and citation repairs**

All independent of Wave 1 and of each other.

- T2.1 — resolve the `docs/README.md:14` active-plan link: add `docs/plans/active/.gitkeep` so the directory exists in the tree (R8, R-D).
  - verify: `git add -f docs/plans/active/.gitkeep && git status --short` shows it staged; `zharness preflight to-plan --json` still resolves this plan; `zharness query plan --json` returns it unchanged.
  - stop: if either command's behaviour changes, remove the file and drop the link instead.
- T2.2 — retitle `skills/workflow/git/references/workflow.md:1` away from `# Playbook: git` and state in its opening lines that it is the git skill's own procedure, not a harness-projected playbook (R9).
  - verify: `head -5 skills/workflow/git/references/workflow.md` no longer reads `Playbook`; the file's claim agrees with `skills/workflow/git/SKILL.md:13` and `docs/README.md:31`.
- T2.3 — correct the three unresolved citations, pointing each at its construct's declaration line (R10, R-E): `docs/ARCHITECTURE.md:56` and `docs/workflow-harness/migration.md:29` → the `SyncManagedDocs` call site; `docs/ARCHITECTURE.md:74` → the `preflightPlaybooks` declaration.
  - verify: for each corrected citation, read the cited `path:line` after Phase 1's edits have landed and confirm it shows the named construct. Then re-check every other `path:line` in `docs/ARCHITECTURE.md` that points into `cli/internal/application/init.go` or `cli/internal/interfaces/preflight.go`, since Phase 1 may have shifted them too.
  - stop: if more than three citations are stale, list them all in `## Progress` before editing — a wider drift means the ADR promise needs a mechanical check, which is NG2 territory.

**Wave 3 — phase gate**

- T3.1 — run both gates and route to `check`.
  - verify: `bash scripts/verify-doc-links.sh` exits zero; `cd cli && go test ./...` all packages ok.
  - final acceptance: all eight findings resolved — 1 (T2.1/Phase 1), 2 (T2.2/Phase 1), 3 (T2.3/Phase 1), 4 (T1.1/Phase 2), 5 (T2.1/Phase 2), 6 (T2.2/Phase 2), 7 (T1.2/Phase 2), 8 (T2.3/Phase 2).

## Current State and Next Action
- active phase: none — initiative complete
- lifecycle_status: done
- latest run: `01M0EGZP9F5R9TKSGVZHJABG88`
- latest check: `01M0EHD7N7NC54RGGV5SB0EZNA` (APPROVED, judge same-session — gate for Phase 2; the final-phase `full` review ran at `2026-08-20T03:04Z`, also APPROVED, with its evidence appended to `## Validation` against this same ID)
- latest trace: `01M0EHM93GPSS4VCQV2SB52WCS` (final-phase full review)
- latest handoff: `01M0EM6DXHZ3G3CFTQBEEP9TZ2` (closed `pr60-go-correctness`), `01M0EM6JP08Y8418NJK082H2EY` (closed `pr60-doc-truth`)
- blockers: none
- completed work: all eight findings from the review of PR #60 closed — three Go defects with red-then-green tests (`pr60-go-correctness`), five document-truth corrections (`pr60-doc-truth`). Shipped to `master` as `9f96ff5`, `798f0a5`, `7f02463` (rebase-merged via PR #60).
- open items: none blocking. Carried forward as known context, not work:
  - Harness gap (reported, not fixed here): `handoff.md` step 6 requires a `full` check on the final phase, but `work`'s in-session gate already moves the story to `checked`, and `zharness check record` rejects a non-`in-progress` story with `story_not_checkable`. The demanded full check therefore cannot record a durable row after the gate that precedes it. Evidence for the full review is appended to `## Validation` against the gate's check ID instead.
  - Resolved (was an open Phase 1 concern): the rollback test's dependence on POSIX `rename`-onto-non-empty-directory semantics is not a portability gap — `cli/.goreleaser.yaml:11-13` ships `darwin` and `linux` only, and `.github/workflows/cli-ci.yml:16` runs `ubuntu-latest`. Windows is not a target.
  - Not independently verified (same-session judge, Phase 1): `db migrate-layout` was exercised only against the `fstest.MapFS` fixture, never against a real legacy consumer repository on disk.
  - Non-blocking observation: a migration that fails after `SyncManagedDocs` staged conflicts leaves files under `.kit/conflicts/`, which `snapshotManagedTargets` does not cover. That directory is gitignored per-machine scratch rebuilt by `zharness init`, so this is residue, not data loss.
  - Not independently verified (same-session judge, Phase 2): that GitHub actually renders the `plans/active` link as a browsable directory once only `.gitkeep` is tracked — verified structurally (file tracked, both plan globs are `*.md`), not by loading the rendered page. And no automated check asserts doc↔code agreement for `path:line` citations; all 28 were resolved by hand, and `verify-doc-links.sh` does not read line numbers (NG2).
  - The dead `ScaffoldResult.AgentsShimNoticePath` field (`init.go:24`) is referenced nowhere. Pre-existing, out of scope, deliberately left alone.
  - Plan-parser gap (out of scope, reported): this plan's phase headings read `### Phase N — \`{slug}\``, but `extractPlanPhaseHeadingBlock` (`cli/internal/application/plan_query.go:178`) expects `### phase_slug: \`{slug}\``, so `zharness query plan --section phase` returns `available phases: (none defined)` with `degraded: true`. The template `cli/docs/embedded/templates/plan.md` does carry `## Decisions`, `## Validation`, and `## Current State and Next Action`; this plan was authored without them and they were hand-added mid-run.
- next action: none — initiative closed. The two harness gaps above (plan-heading parser, and `check record` rejecting the `full` check that `handoff.md` step 6 demands) are unfiled; open them as issues if they should be fixed rather than remembered.

## Validation
- `2026-08-20T02:43:01Z` — check. verdict: `APPROVED`. check: `01M0EGSV5EQKZ0S3B88GMNWP83`. run: `01M0EGHVGAN9TK70M66XRWFEKX`. phase: `pr60-go-correctness`. judge: `same-session` (claude-opus-5).
  - `go test -C cli ./...` → Validation entry 2026-08-20T02:45Z: 6 packages ok, including TestMigrateLayoutRollsBackEveryScaffoldedFile and TestScaffoldDocsKeepsClaudeMdWithLegacyKitCopy, both red before Wave 2
  - `bash scripts/verify-doc-links.sh` → Validation entry 2026-08-20T02:45Z: doc links OK (0 findings), exit 0
- `2026-08-20T02:53:36Z` — check. verdict: `APPROVED`. check: `01M0EHD7N7NC54RGGV5SB0EZNA`. run: `01M0EGZP9F5R9TKSGVZHJABG88`. phase: `pr60-doc-truth`. judge: `same-session` (claude-opus-5).
  - `bash scripts/verify-doc-links.sh` → Validation entry 2026-08-20T02:56Z: doc links OK (0 findings), exit 0
  - `go test -C cli ./...` → Validation entry 2026-08-20T02:56Z: all 6 packages ok, exit 0
- `2026-08-20T03:04Z` — check, mode `full` (final-phase review, `handoff.md` step 6). verdict: `APPROVED`. check: `01M0EHD7N7NC54RGGV5SB0EZNA` (no new row — see the harness gap in Current State; the story was already `checked`, so `check record` returned `story_not_checkable`, and the reviewed tree is byte-identical to the tree that check gated). run: `01M0EGZP9F5R9TKSGVZHJABG88`. judge: `same-session` (claude-opus-5).
  - `go vet -C cli ./...` → exit 0, no findings
  - `go test -C cli ./...` → 6/6 packages ok
  - `bash scripts/verify-doc-links.sh` → doc links OK (0 findings), exit 0
  - Security: no secrets, network, or credential surface. The `relPath == "AGENTS.md"` pin at `cli/internal/application/init.go:91` removes a whole-file-replace branch that was reachable for consumer-owned files — net reduction in blast radius. `snapshotManagedTargets` derives paths from the compile-time `scaffoldOnceDocs` list, not from input, so the added `filepath.FromSlash` join carries no traversal risk. Verified `restoreFileSnapshots` (`layout_migration.go:333-347`) removes a path only when `!snapshot.Exists`, so a pre-existing consumer `CLAUDE.md` is restored by content and mode, never deleted.
  - Class-of-bug coverage: `writeManagedBlock` has exactly two callers (`init.go:51,57`) and the `.kit/docs/` probe appears once (`init.go:92`); no sibling instance of the parameterized-probe bug exists. `snapshotManagedTargets` is the sole snapshot builder and all ten failure points in `MigrateLayout` route through `restoreFileSnapshots`.
  - Performance: snapshot base list grows 2 → 6 small markdown reads on a once-per-repository command. Not a hot path.
  - Architecture: iterating the `scaffoldOnceDocs` variable keeps one source of truth for the scaffold-once set (R1). `ScaffoldResult` is package-internal with two consumers, so splitting the boolean changes no external contract.
  - Non-goals: NG1–NG6 all held. No `managed_docs.go` prune path, `scripts/verify-doc-links.sh` untouched, `preflightPlaybooks` unchanged, no re-litigation, no release, no drive-by reformatting.
  - Required proof for lane `normal` (unit + command output): satisfied. Both new tests were proven red before their fix.

## Decisions
- `2026-08-20T02:42:13Z` — Inject the post-ScaffoldDocs migration failure by placing a non-empty directory at the legacy db backup path, making os.Rename at layout_migration.go:175 fail. (phase: `pr60-go-correctness`), task: T1.1. rationale: The rollback gap only appears at failure points reached after ScaffoldDocs returns successfully. The pre-existing conflict test fails inside SyncManagedDocs, before CLAUDE.md or the scaffold-once docs are ever written, which is exactly why it passed against the bug. The backup-rename blocker needs no seam, no interface change and no test hook, is deterministic on Linux (rename onto a non-empty directory fails), and sits outside every MigrateLayout precondition check..
- `2026-08-20T02:42:13Z` — Split ScaffoldResult.AgentsShimWritten into two booleans instead of rewording the single print line. (phase: `pr60-go-correctness`), task: T2.3. rationale: One boolean for two files cannot report which file changed, so any single message is wrong in some run. Two fields keep each print line true independently, and every existing assertion on AgentsShimWritten still holds because a fresh root writes both files - no test needed editing. The --json branch emits no scaffold field, so R-C carried no output-compatibility cost..
- `2026-08-20T02:47:58Z` — Added a permanent .claimignore entry for docs/playbooks/git.md rather than rewording the migration note to avoid a backticked path. (phase: `pr60-doc-truth`), task: T1.2. rationale: Naming the exact path is the whole point of the note - a consumer cannot delete a file the doc refuses to spell. The gate flagged it correctly: the file genuinely does not exist in this repository, because docs-architecture R13 deprojected it. This is the case .claimignore exists for, and the entry states it is permanent rather than a deferral, per docs-architecture R12s requirement that deferral entries say so. Plan gap: Phase 2 assumed the doc edits were gate-neutral, but T1.2 necessarily introduces a reference to a deliberately absent file..
- `2026-08-20T03:42:57Z` — plan completed. rationale: every phase_slug is a done story.

## Progress
- `2026-08-20T02:39:01Z` — wave 1. run: `01M0EGHVGAN9TK70M66XRWFEKX`. summary: Phase pr60-go-correctness started (run 01M0EGHVGAN9TK70M66XRWFEKX). Wave 1 writes both data-loss tests red before any production edit..
- `2026-08-20T02:40:30Z` — wave 1, task T1.1. task_status: `DONE`. run: `01M0EGHVGAN9TK70M66XRWFEKX`. summary: Added TestMigrateLayoutRollsBackEveryScaffoldedFile (cli/internal/application/layout_migration_test.go); blocks os.Rename at layout_migration.go:175 with a non-empty dir at the backup path, the earliest failure point after ScaffoldDocs succeeds. RED as required: go test ./internal/application/ -run TestMigrateLayout -v reports CLAUDE.md, docs/README.md, docs/decisions/README.md and docs/decisions/templates/decision.md all survived rollback; AGENTS.md correctly rolled back, isolating the gap to the 4 unsnapshotted paths. Both pre-existing TestMigrateLayout tests still pass..
- `2026-08-20T02:40:30Z` — wave 1, task T1.2. task_status: `DONE`. run: `01M0EGHVGAN9TK70M66XRWFEKX`. summary: Added TestScaffoldDocsKeepsClaudeMdWithLegacyKitCopy (cli/internal/application/init_test.go); seeds a consumer CLAUDE.md plus a byte-identical .kit/docs/CLAUDE.md. RED as required: the root file is reduced to the three-line managed block, confirming the whole-file-replace branch at init.go:102-103. All 9 pre-existing TestScaffoldDocs tests pass unchanged..
- `2026-08-20T02:40:30Z` — wave 1. run: `01M0EGHVGAN9TK70M66XRWFEKX`. summary: Wave 1 complete: both data-loss defects proven red before any production edit, each failing with exactly the predicted output. Gate to Wave 2 satisfied..
- `2026-08-20T02:41:59Z` — wave 2, task T2.1. task_status: `DONE`. run: `01M0EGHVGAN9TK70M66XRWFEKX`. summary: snapshotManagedTargets (layout_migration.go:297) now seeds CLAUDE.md and iterates the scaffoldOnceDocs variable rather than repeating literals, so a future fourth entry cannot fall out of the rollback set (R1). Verify: go test ./internal/application/ -run TestMigrateLayout -v => all 3 PASS, both pre-existing tests unchanged..
- `2026-08-20T02:41:59Z` — wave 2, task T2.2. task_status: `DONE`. run: `01M0EGHVGAN9TK70M66XRWFEKX`. summary: Pinned the legacy .kit/docs/ probe to AGENTS.md (init.go:85-91); branch logic at :102-103 untouched (R3). Verify: go test ./internal/application/ -run TestScaffoldDocs -v => all 10 PASS. R-B did not fire: zero pre-existing tests needed modification..
- `2026-08-20T02:41:59Z` — wave 2, task T2.3. task_status: `DONE`. run: `01M0EGHVGAN9TK70M66XRWFEKX`. summary: R-C checked first: ScaffoldResult is consumed only by cli/internal/interfaces/init.go:95 and two tests, and the --json branch emits no scaffold field, so no consumer-visible JSON key changes. Split AgentsShimWritten into AgentsShimWritten + ClaudeShimWritten and added the matching print line (R5). Observable proof in a scratch fixture: a second init with AGENTS.md current and CLAUDE.md removed prints only "updated CLAUDE.md managed block" - the exact scenario the finding named..
- `2026-08-20T02:42:13Z` — wave 2. run: `01M0EGHVGAN9TK70M66XRWFEKX`. summary: Wave 2 complete: all three Go defects fixed, both red tests now green, full suite ok (6 packages) and doc-link gate 0 findings. Diff confined to the 5 allowed surfaces..
- `2026-08-20T02:47:58Z` — wave 1, task T1.1. task_status: `DONE`. run: `01M0EGZP9F5R9TKSGVZHJABG88`. summary: Rewrote the migration.md init paragraph into a three-row ownership table naming every write: projected (docs/WORKFLOW.md, docs/playbooks/*.md), managed block (root AGENTS.md init.go:51 and root CLAUDE.md init.go:57, with the @AGENTS.md import rationale), and scaffold-once (docs/README.md, docs/decisions/README.md, docs/decisions/templates/decision.md at init.go:64,126, written only when absent and never refreshed). Added the explicit promise that an existing CLAUDE.md keeps its content byte-for-byte. Verify: grep -n CLAUDE.md docs/workflow-harness/migration.md returns the new rows; read against init.go:30-72, no write is unlisted (R6)..
- `2026-08-20T02:47:58Z` — wave 1, task T1.2. task_status: `DONE`. run: `01M0EGZP9F5R9TKSGVZHJABG88`. summary: Added an Upgrading section to migration.md covering the git deprojection: preflight git returns no playbook, the procedure lives at skills/workflow/git/references/workflow.md, init will NOT clean up the stale docs/playbooks/git.md because planManagedDocActions walks the embedded FS and never visits a departed row (managed_docs.go:107, citation verified), with an explicit trash/rm command and a note that the orphaned managed_docs row is inert and clears on db rebuild (R7)..
- `2026-08-20T02:47:58Z` — wave 1. run: `01M0EGZP9F5R9TKSGVZHJABG88`. summary: Wave 1 complete: both consumer-facing corrections landed in migration.md; doc-link gate green after adding the justified .claimignore entry..
- `2026-08-20T02:52:40Z` — wave 2, task T2.1. task_status: `DONE`. run: `01M0EGZP9F5R9TKSGVZHJABG88`. summary: Added docs/plans/active/.gitkeep (git add -f) so the plans/active link in docs/README.md resolves on GitHub (R8). R-D cleared: both plan globs are *.md (interfaces/preflight.go:20, plan_resolve.go:39), so .gitkeep can never be read as a plan..
- `2026-08-20T02:52:40Z` — wave 2, task T2.2. task_status: `DONE`. run: `01M0EGZP9F5R9TKSGVZHJABG88`. summary: Retitled skills/workflow/git/references/workflow.md from 'Playbook: git' to 'git skill: workflow reference' plus a note stating it is not projected, init never writes it, preflight git returns no playbook, and a docs/playbooks/git.md on disk is a stale pre-deprojection projection (R9)..
- `2026-08-20T02:52:40Z` — wave 2, task T2.3. task_status: `DONE`. run: `01M0EGZP9F5R9TKSGVZHJABG88`. summary: Re-resolved all 28 path:line citations in docs/ARCHITECTURE.md and docs/workflow-harness/migration.md. Only one was stale: ARCHITECTURE.md:74 cited interfaces/preflight.go:30 (a closing brace); corrected to :23, the preflightPlaybooks declaration (R10, R-E). Stop condition (>3 stale) not triggered..
- `2026-08-20T02:52:44Z` — wave 2. run: `01M0EGZP9F5R9TKSGVZHJABG88`. summary: Wave 2 complete: link/citation truth. .gitkeep restores the plans/active link, the git reference no longer calls itself a playbook, and the single stale citation (preflight.go:30 -> :23) is corrected. Next: wave 3 gates..
- `2026-08-20T02:54:24Z` — wave 3. run: `01M0EGZP9F5R9TKSGVZHJABG88`. summary: Wave 3 gate: both gates green (verify-doc-links.sh 0 findings; go test -C cli ./... 6 packages ok). Check 01M0EHD7N7NC54RGGV5SB0EZNA APPROVED, judge same-session. Phase pr60-doc-truth synced to checked in plan and DB. Gate 1 initially failed on the git reference's assertion that no cli/docs/embedded/playbooks/git.md exists; rephrased to cite the directory instead of the absent file rather than spend a second .claimignore exception..
- `2026-08-20T02:57:27Z` — wave 3. run: `01M0EGZP9F5R9TKSGVZHJABG88`. summary: Final-phase full review (handoff.md step 6): Security/Performance/Architecture/Code Quality all clean, NG1-NG6 held, class-of-bug coverage complete (writeManagedBlock has 2 callers, probe appears once). go vet clean. Could not record a durable row: story already checked, check record returns story_not_checkable - reported as a harness gap; evidence appended to Validation against check 01M0EHD7N7NC54RGGV5SB0EZNA. Windows-portability concern closed: goreleaser targets darwin+linux only..
- `2026-08-20T03:42:19Z` — handoff recorded. handoff: `01M0EM6DXHZ3G3CFTQBEEP9TZ2`. run: `01M0EGHVGAN9TK70M66XRWFEKX`. check: `01M0EGSV5EQKZ0S3B88GMNWP83`. phase closed.
- `2026-08-20T03:42:24Z` — handoff recorded. handoff: `01M0EM6JP08Y8418NJK082H2EY`. run: `01M0EGZP9F5R9TKSGVZHJABG88`. check: `01M0EHD7N7NC54RGGV5SB0EZNA`. phase closed.
