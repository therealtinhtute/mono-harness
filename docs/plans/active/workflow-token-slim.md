---
status: active
lane: high-risk
---
# Workflow Token Slim

## Goal
- outcome: every recommendation R1–R9 of `docs/audit/workflow-token-slim-audit.md` is applied, so each common path loads only the material its branch uses, with no rule, guard, proof requirement, or judge requirement removed.
- success_signal: every byte budget below holds (`wc -c`, tokens ≈ bytes ÷ 4), every phrase the playbook contract test pinned at `2befb37` is still pinned, and the full gate set in `AGENTS.md` passes.
- actors: agents running the workflow skills in this repo and in consumer repos (via `npx skills add` and `zharness update`); the repo owner reviewing the PRs.
- authority: `docs/audit/workflow-token-slim-audit.md` (merged in `2befb37`); owner decisions on 2026-09-27: all of R1–R9 in three phases, lane `high-risk`, `model:` pins are a non-goal, work branches from `master` after the audit merged.
- requirements:
  - R1: `git` commit path loads only `skills/workflow/git/SKILL.md`, holding the commit flow (stage, secret scan, split, commit, push) moved byte for byte from `references/workflow.md`; `pr` and `merge` stay in `references/workflow.md`, loaded only for those subcommands; the dead `review` defer-pointer becomes `check` | acceptance: `wc -c < skills/workflow/git/SKILL.md` ≤ 6000 (from 11,083 on the commit path) and SKILL.md no longer says to always read `references/workflow.md` | source: audit R1, F2
  - R2: `check.md` splits by branch into `check.md` (modes, bounded path, output), `check-durable.md` (preflight, owned state, steps 4 and 8–11), and `check-review.md` (two axes and baselines, loaded only at step 5); "What the Guards Cannot Check" moves into `check-validation.md` | acceptance: `check.md` ≤ 5000 bytes; `check.md` + `check-durable.md` + `check-validation.md` ≤ 10,700 bytes (from 12,677); contract and parity tests pass | source: audit R2, F1
  - R3: `brainstorm.md` splits into core (modes, lanes, explore steps 1–3, exit), `brainstorm-grill.md` (Grill loop), and `brainstorm-lock.md` (Plan Skeleton and steps 4–10) | acceptance: `brainstorm.md` ≤ 3400 bytes (from 8,805); contract and parity tests pass | source: audit R3, F1
  - R4: `AGENTS.md` drops the Project Structure tree and the Skill Pipeline prose for one pointer to `skills/workflow/README.md`, keeping Gate Commands, the playbook-edit parity rule, the Prompt Engineering pointer, Architecture Notes, and the managed Harness block unchanged | acceptance: `wc -c < AGENTS.md` ≤ 5200 (from 9,938) and the `ZHARNESS:BEGIN`…`END` block is byte-identical | source: audit R4, F5
  - R5: `rules/` drops the restated "Always-on Discipline" from `execution-discipline.md` (folding any item with no other home into `karpathy-guidelines.md`) and the "Workflow" section from `workflow-core.md` | acceptance: `cat rules/*.md | wc -c` ≤ 7700 (from 9,164) | source: audit R5, F6
  - R6: skill boilerplate shrinks: one `version:` per SKILL frontmatter (drop the duplicate `metadata.version`), the thin-trigger fallback line loses "No binary runs the lifecycle." but keeps the IF-absent behavior, and the `hunt`/`think` `Sources:` lines move to a `NOTICE.md` in each skill directory | acceptance: `grep -L` finds no `metadata:` block and no `Sources:` line in the 10 workflow SKILL.md files; every one still passes `scripts/validate-skill.sh` | source: audit R6, F3
  - R7: `hunt` step 4 reads only `## Gotchas` of `references/failure-patterns.md`, then only the `##` section it names | acceptance: `hunt/SKILL.md` names the targeted read (`## Gotchas` then one section) and no longer says to load the whole file | source: audit R7, F4
  - R8: `hunt` splits by mode: steps 5–7 move to `references/fix.md` (loaded by `fix`), Bisect and Regression to `references/bisect-regression.md`, and `triage` loads `references/triage.md` without reading the loop | acceptance: `wc -c < skills/workflow/hunt/SKILL.md` ≤ 9000 (from 11,414) and each moved section has a mandatory load pointer at its branch point | source: audit R8, F1
  - R9: `think` step 5 carries the four core lens definitions inline; `references/lenses.md` loads only for Risk/Design/Evidence lenses or `lens <name>` | acceptance: `think decide` path (SKILL.md alone) ≤ 9300 bytes (from 12,974) | source: audit R9, F4
- non-goals:
  - NG1: `model:` frontmatter pins (`hunt`, `think`, `to-plan`, `check`, `git`); a separate cost decision.
  - NG2: any change to the pre-commit guards, proof re-execution, judge rules, or `scripts/test-guards.sh` fixtures.
  - NG3: rewording retained text; moved text moves byte for byte (wording is `playbook-token-audit.md` territory).
  - NG4: new `zharness` verbs, new scripts, or a release cut.
  - NG5: `site/`, which is hand-authored and allowed to drift.

## Phases and Verification
- approach: three PR-sized phases ordered by risk. Phase 1 changes only skill-directory files, `AGENTS.md`, and `rules/`, none of which a contract test pins. Phase 2 splits the two projected playbooks with the pointer pattern `work.md` → `work-full.md` already uses, moving each pinned phrase byte for byte and re-pointing its contract entry. Phase 3 splits `hunt` by mode. Rejected: one PR for all nine (too wide a diff for an independent judge, and phase 1 would wait on the riskiest change); rewording in place (the audit shows the waste is unused branches, not verbose sentences).
- risks: a moved pinned phrase lands in a file the contract does not read → the contract test fails loudly; mitigation: re-point the entry in the same task and run the phrase-preservation check. A split drops a pointer so an agent never loads the moved rules → mitigation: every moved block gets a mandatory "read X now" pointer at its branch point, checked by `grep`. Budgets are estimates (±15%) → a miss inside 15% is recorded, not a blocker.
- constraints: phrases pinned at `2befb37` stay pinned; `ZHARNESS` block in `AGENTS.md` byte-identical; `cli/docs/embedded/playbooks/*` and `docs/playbooks/*` byte-identical; the release binary stays ≤ 1,850,000 bytes.
- recovery: each phase is one PR; revert that PR to restore the prior load. Phase 2 consumers who already ran `zharness update` keep working on the old single files until they update again.
- seams: `cd cli && cargo test` (`one_plan_playbook_contract`, `projection_parity`, `playbook_count`); `bash scripts/validate-skill.sh <SKILL.md>`; `bash scripts/verify-doc-links.sh`; byte budgets via `wc -c`; phrase preservation via `bash -c 'comm -23 <(git show 2befb37:cli/src/embedded.rs | grep -oE "^ +\"[^\"]+\",\$" | sed "s/^ *//" | sort -u) <(grep -oE "^ +\"[^\"]+\",\$" cli/src/embedded.rs | sed "s/^ *//" | sort -u) | (! grep .)'` (preservation check). All exist; no new seam.
- phase_slug: `slim-always-on-and-git`
  status: planned
  - goal: R1, R4, R5, R6, R7, R9 | depends_on: none
  - surfaces: `skills/workflow/*/SKILL.md`, `skills/workflow/git/references/workflow.md`, `skills/workflow/{hunt,think}/NOTICE.md`, `AGENTS.md`, `rules/*.md`, `CHANGELOG.md` | avoided: `cli/`, `docs/playbooks/`, `scripts/`, `site/`
  - escalate_when: a budget misses by more than 15%, or folding "Always-on Discipline" would drop an item that has no other home in `rules/`
  - wave 1:
    - T1 slim `git`: move Steps 1–5 of `references/workflow.md` into SKILL.md byte for byte; SKILL loads `references/workflow.md` only for `pr`/`merge`; drop role/context no-ops; `review` → `check` — output: commit path is SKILL.md alone — check: `sh -c 'test $(wc -c < skills/workflow/git/SKILL.md) -le 6000 && grep -q "Step 1: Stage + analyze" skills/workflow/git/SKILL.md && ! grep -q "Step 1: Stage" skills/workflow/git/references/workflow.md && grep -q "symbolic-ref" skills/workflow/git/references/workflow.md && bash scripts/validate-skill.sh skills/workflow/git/SKILL.md'` — stop_if: a moved step needs rewording to fit
    - T2 slim `AGENTS.md`: drop the tree and pipeline prose, keep the playbook-edit parity sentence, add one pointer to `skills/workflow/README.md` — output: smaller `AGENTS.md` with the Harness block untouched — check: `bash -c 'test $(wc -c < AGENTS.md) -le 5200 && cmp <(git show 2befb37:AGENTS.md | sed -n "/ZHARNESS:BEGIN/,/ZHARNESS:END/p") <(sed -n "/ZHARNESS:BEGIN/,/ZHARNESS:END/p" AGENTS.md)'` — stop_if: a Gate Commands line would change
    - T3 clean `rules/`: remove `execution-discipline.md` "Always-on Discipline" and `workflow-core.md` "Workflow", folding any orphan item into `karpathy-guidelines.md` — output: no restated section in `rules/` — check: `sh -c 'test $(cat rules/*.md | wc -c) -le 7700 && ! grep -q "Always-on Discipline" rules/*.md'` — stop_if: an item has no equivalent elsewhere and no natural home
    - T4 `hunt` targeted read: rewrite the step 4 pointer to read `## Gotchas`, then only the named section — output: recurring-bug path ≈ 4,000 tokens instead of 7,640 — check: `sh -c 'grep -q "## Gotchas" skills/workflow/hunt/SKILL.md && ! grep -q "gotchas table first" skills/workflow/hunt/SKILL.md && bash scripts/validate-skill.sh skills/workflow/hunt/SKILL.md'` — stop_if: a Gotchas row names no section
    - T5 `think` core lenses inline: move `## Core (almost always)` of `references/lenses.md` into step 5; load `lenses.md` only for other lenses — output: decide path is SKILL.md alone — check: `sh -c 'test $(wc -c < skills/workflow/think/SKILL.md) -le 9300 && grep -q "premise-collapse" skills/workflow/think/SKILL.md && grep -q "simplicity-gate" skills/workflow/think/SKILL.md && ! grep -q "^## Core" skills/workflow/think/references/lenses.md && bash scripts/validate-skill.sh skills/workflow/think/SKILL.md'` — stop_if: core lenses need the rest of `lenses.md` to be usable
  - wave 2:
    - T6 boilerplate across the 10 workflow SKILL.md files: single `version:`, shorter fallback line, `Sources:` → `NOTICE.md`; add one `CHANGELOG.md` Unreleased entry for the phase — output: leaner frontmatter, MIT attribution kept in-tree — check: `sh -c 'for f in skills/workflow/*/SKILL.md; do ! grep -q "^metadata:" $f && ! grep -q "^Sources:" $f && bash scripts/validate-skill.sh $f >/dev/null || exit 1; done; test -f skills/workflow/hunt/NOTICE.md && test -f skills/workflow/think/NOTICE.md'` — stop_if: `metadata.version` turns out to be read by `npx skills` or any script
  - phase check: `bash scripts/verify-doc-links.sh && cd cli && cargo fmt --check && cargo clippy --all-targets -- -D warnings && cargo test`
- phase_slug: `split-check-brainstorm`
  status: planned
  - goal: R2, R3 | depends_on: none
  - surfaces: `cli/docs/embedded/playbooks/{check,check-durable,check-review,check-validation,brainstorm,brainstorm-grill,brainstorm-lock}.md` and their `docs/playbooks/` copies, `cli/src/embedded.rs` (contract entries, `playbook_count`, doc comment), `docs/ARCHITECTURE.md`, `docs/README.md`, `CHANGELOG.md` | avoided: `scripts/install-git-hooks.sh`, `cli/src/installer/`, `site/`
  - escalate_when: a pinned phrase cannot move without changing its meaning, a guard parses text that would move, or `bounded`/`explore` still needs a moved block to finish
  - wave 1:
    - T1 split `check.md` into `check.md` + `check-durable.md` + `check-review.md`, move "What the Guards Cannot Check" into `check-validation.md`, add mandatory pointers at each branch point, re-point contract entries, bump `playbook_count`, copy to `docs/playbooks/` — output: bounded loads `check.md` only — check: `bash -c 'test $(wc -c < cli/docs/embedded/playbooks/check.md) -le 5000 && test $(cat cli/docs/embedded/playbooks/{check,check-durable,check-validation}.md | wc -c) -le 10700 && cd cli && cargo test'` — stop_if: a pinned phrase must be edited to fit its new file
  - wave 2:
    - T2 split `brainstorm.md` into core + `brainstorm-grill.md` + `brainstorm-lock.md` the same way — output: explore loads `brainstorm.md` only — check: `bash -c 'test $(wc -c < cli/docs/embedded/playbooks/brainstorm.md) -le 3400 && cd cli && cargo test'` — stop_if: as T1
  - wave 3:
    - T3 sync docs: the playbook count and companion list in `docs/ARCHITECTURE.md`, `docs/README.md`, and the `embedded.rs` doc comment; one `CHANGELOG.md` Unreleased entry noting the new managed files — output: docs name 12 playbooks — check: `sh -c '! grep -q "eight playbooks" docs/ARCHITECTURE.md && bash scripts/verify-doc-links.sh'` — stop_if: none
  - phase check: `bash -c 'comm -23 <(git show 2befb37:cli/src/embedded.rs | grep -oE "^ +\"[^\"]+\",\$" | sed "s/^ *//" | sort -u) <(grep -oE "^ +\"[^\"]+\",\$" cli/src/embedded.rs | sed "s/^ *//" | sort -u) | (! grep .)' && bash scripts/verify-doc-links.sh && bash scripts/test-guards.sh && cd cli && cargo fmt --check && cargo clippy --all-targets -- -D warnings && cargo test && cargo build --release --target x86_64-unknown-linux-musl && test "$(wc -c < target/x86_64-unknown-linux-musl/release/zharness)" -le 1850000`
- phase_slug: `split-hunt-modes`
  status: planned
  - goal: R8 | depends_on: slim-always-on-and-git
  - surfaces: `skills/workflow/hunt/SKILL.md`, `skills/workflow/hunt/references/{fix,bisect-regression}.md`, `skills/workflow/README.md` (hunt row), `CHANGELOG.md` | avoided: other `hunt` references, `cli/`, `docs/playbooks/`
  - escalate_when: the `fix` path would load more than it does today, or a Hard Rule only makes sense next to steps 5–7
  - wave 1:
    - T1 move steps 5–7 to `references/fix.md` with a mandatory pointer from the `fix` mode row and the end of step 4 — output: `diagnose` stops loading steps 5–7 — check: `sh -c 'grep -q "references/fix.md" skills/workflow/hunt/SKILL.md && test -f skills/workflow/hunt/references/fix.md && bash scripts/validate-skill.sh skills/workflow/hunt/SKILL.md'` — stop_if: step 7 cleanup is needed by `diagnose`
  - wave 2:
    - T2 move Bisect and Regression to `references/bisect-regression.md`; route `triage` straight to `references/triage.md` without the loop; update the `hunt` row in `skills/workflow/README.md`; one `CHANGELOG.md` entry — output: every mode loads only its branch — check: `sh -c 'test $(wc -c < skills/workflow/hunt/SKILL.md) -le 9000 && bash scripts/verify-doc-links.sh'` — stop_if: `triage.md` relies on loop steps it does not restate
  - phase check: `sh -c 'bash scripts/validate-skill.sh skills/workflow/hunt/SKILL.md && bash scripts/verify-doc-links.sh'`

## Log

## Validation

## Current State and Next Action
- active_phase: none
- lifecycle_status: planned
- blockers: none
- open_items: none
- exact_next_action: work full phase slim-always-on-and-git
