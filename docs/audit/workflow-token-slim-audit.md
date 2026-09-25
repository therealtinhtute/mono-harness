# Workflow Token Slim Audit

**Date:** 2026-09-25
**Scope:** the 8 `skills/workflow/*/SKILL.md` files, their references, `docs/WORKFLOW.md`, the playbooks in `docs/playbooks/`, the repository `AGENTS.md`, and the global rules in `rules/`.
**Method:** byte counts measured on `master` at `977fd8b`; tokens ≈ bytes ÷ 4 (no tokenizer run, ±15%). Load paths traced from each skill's routing lines.
**Status:** report only. No source file was edited.
**Extends:** [`playbook-token-audit.md`](playbook-token-audit.md), which compressed wording inside each playbook. This audit targets a different waste: material a branch loads but never uses.

## 1. Conclusion

The largest savings come from loading only what the active branch uses, not from rewording. Three targets dominate:

- `git`: it reads its full `references/workflow.md` on every commit.
- `check.md`: it is loaded on every `work full` phase and every `check` run.
- `AGENTS.md`: it sits in context on every turn.

If each file is split by branch, with the same pointer pattern `work.md` → `work-full.md` already uses, the common paths lose an estimated 40–60% of their load. No rule, guard, proof requirement, or judge requirement is removed.

## 2. Baseline

### 2.1 Always in context (every turn)

| Source | Tokens | Note |
|---|---|---|
| `AGENTS.md` (this repo) | ~2,340 | The Project Structure tree is ~500; the Skill Pipeline prose repeats `skills/workflow/README.md` |
| `rules/*.md` (installed globally) | ~2,260 | `workflow-core.md` points at `hunt` and `think`, which do not exist; its "Always-on Discipline" overlaps `karpathy-guidelines.md` and `execution-discipline.md` |
| 8 skill descriptions | ~330 | Proportionate |

### 2.2 Per invocation

| Invocation | Loads | Tokens | Frequency |
|---|---|---|---|
| `git` | SKILL.md (644) + `references/workflow.md`, always read (2,112) | **~2,760** | Every commit |
| `work full`, per phase | SKILL (256) + `work.md` (765) + `work-full.md` (1,288) + `check.md` (2,513) + `check-validation.md` (632) | **~5,450** | Every phase |
| `check bounded` | SKILL (260) + `check.md` (2,513) | ~2,770 | Frequent |
| `brainstorm explore` | SKILL (284) + `brainstorm.md` (2,184) | ~2,470 | Frequent |
| `to-plan` | SKILL (234) + `to-plan.md` (1,234) | ~1,470 | Per initiative |
| `handoff` | SKILL (226) + `handoff.md` (1,234) | ~1,460 | Per session |
| `watzup` | SKILL (216) + `watzup.md` (738) | ~950 | Per session |
| `retro` | SKILL (158) + `retro.md` (1,510), plus one mode reference when fixing | ~1,670+ | Rare |

## 3. Findings

| # | Waste | Evidence |
|---|---|---|
| F1 | **Each branch loads the whole file.** | `check bounded` reads the preflight, owned state, and "What the Guards Cannot Check" (~1,600 tokens) that only durable modes use. `gate` reads the two-axis review and smell baseline (~470) that only step 5 of `full`/`bounded` uses. `brainstorm explore` reads the Plan Skeleton plus lock steps (~1,400) and the Grill section (~600) |
| F2 | **`git` always reads its full reference.** | A plain commit needs stage, secret scan, split, and commit. The `pr` and `merge` procedures, Error Handling, Anti-Patterns, and Command Reference ride along on every commit, and much of that is default model behaviour (no-ops) |
| F3 | **Boilerplate repeated in the 6 thin triggers.** | "No binary runs the lifecycle. IF the playbook is absent…", plus both `version` and `metadata.version`, `license`, `compatibility` |
| F4 | **Dead pointers.** | `rules/workflow-core.md` routes to `hunt` and `think`. The cost is not tokens but an agent looking for a skill that is not installed |
| F5 | **`AGENTS.md` caches the environment.** | The directory tree is one `ls` away; the pipeline already lives in `skills/workflow/README.md` |

## 4. Recommendations

Ranked by frequency × savings.

| # | Change | Savings | Risk |
|---|---|---|---|
| R1 | **Slim `git`.** SKILL keeps the commit flow; `pr` and `merge` move to references loaded only for that subcommand; delete no-ops and SKILL/reference duplication | ~2,760 → **~1,100 per commit** | Low |
| R2 | **Split `check.md` by branch.** `check.md`: modes, bounded path, output format. `check-durable.md`: preflight, gate steps, state sync. `check-review.md`: the two axes and baselines, loaded only at step 5. Move "What the Guards Cannot Check" into `check-validation.md` | bounded ~2,770 → ~1,200; gate ~500 less per phase | Medium: contract-test paths change |
| R3 | **Split `brainstorm.md` by branch.** Core: modes, lanes, explore, raw. `brainstorm-grill.md`: the Grill loop. `brainstorm-lock.md`: skeleton and lock steps 5–10 | explore ~2,470 → ~1,000 | Medium: as R2 |
| R4 | **Slim `AGENTS.md`.** Replace the tree and duplicated pipeline prose with one pointer to `skills/workflow/README.md`; keep Gate Commands and the managed Harness block | ~2,340 → **~1,100 per turn** | Low |
| R5 | **Clean `rules/`.** Fix the `hunt`/`think` pointers; fold "Always-on Discipline" into one file | ~2,260 → ~1,600 per turn, every repo | Low |
| R6 | Drop repeated boilerplate from the 6 thin triggers | ~50 per load | Low |

Estimated effect: about 500 fewer tokens per `work full` phase, about 1,700 fewer per commit, and about 1,900 fewer per turn from `AGENTS.md` plus `rules/`. Prompt caching makes the always-on material cheap to resend in money, so the per-turn gain is mostly a smaller, more legible context, not a proportional bill cut.

Suggested order: R1 + R4 + R5 first (low risk; every turn and every commit benefit), then R2 + R3 together in a separate PR.

## 5. Safeguards

- **Split by the existing pattern.** `work.md` already routes with "read `docs/playbooks/work-full.md` now", pinned by the playbook contract test. Each new companion gets the same mandatory pointer at its branch point and a contract entry for it.
- **Move pinned phrases byte for byte.** Every phrase the contract test pins moves to its new file unchanged; guards, proof, and judge rules do not change.
- **Measure before and after.** Run `skills/workflow/retro/scripts/session-digest.py` on one or two real sessions per path (peak context, output tokens) alongside the full gate set.
- **Consumers.** R2 and R3 add embedded files; `zharness update` delivers them as new managed files. Nothing existing breaks.
