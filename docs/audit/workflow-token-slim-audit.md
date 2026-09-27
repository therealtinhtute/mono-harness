# Workflow Token Slim Audit

**Date:** 2026-09-25 (revised 2026-09-27 for `hunt` and `think`)
**Scope:** the 10 `skills/workflow/*/SKILL.md` files, their references, `docs/WORKFLOW.md`, the playbooks in `docs/playbooks/`, the repository `AGENTS.md`, and the global rules in `rules/`.
**Method:** byte counts measured on `master` at `3632c91`; tokens ≈ bytes ÷ 4 (no tokenizer run, ±15%). Load paths traced from each skill's routing lines.
**Status:** report only. No source file was edited.
**Extends:** [`playbook-token-audit.md`](playbook-token-audit.md), which compressed wording inside each playbook. This audit targets a different waste: material a branch loads but never uses.

## 1. Conclusion

The largest savings come from loading only what the active branch uses, not from rewording. Four targets dominate:

- `git`: it reads its full `skills/workflow/git/references/workflow.md` on every commit.
- `check.md`: it is loaded on every `work full` phase and every `check` run.
- `AGENTS.md`: it sits in context on every turn.
- `hunt`: `rules/workflow-core.md` auto-triggers it on every error, its `SKILL.md` is the largest in the chain (~2,850), and a recurring bug pulls in the whole `failure-patterns.md` (~4,050) to read one table.

If each file is split by branch, with the same pointer pattern `work.md` → `work-full.md` already uses, the common paths lose an estimated 40–60% of their load. No rule, guard, proof requirement, or judge requirement is removed.

## 2. Baseline

### 2.1 Always in context (every turn)

| Source | Tokens | Note |
|---|---|---|
| `AGENTS.md` (this repo) | ~2,480 | The Project Structure tree is ~700 and had to be edited again when `hunt`/`think` landed; the Skill Pipeline prose repeats `skills/workflow/README.md` |
| `rules/*.md` (installed globally) | ~2,290 | `execution-discipline.md` "Always-on Discipline" (~230) restates SOUL and `karpathy-guidelines.md`; `workflow-core.md` "Workflow" (~170) restates `karpathy-guidelines.md` §4 and the `AGENTS.md` Harness block |
| 10 workflow skill descriptions | ~460 | Proportionate; `hunt` and `think` added ~100 |

### 2.2 Per invocation

| Invocation | Loads | Tokens | Frequency |
|---|---|---|---|
| `git` | SKILL.md (648) + `skills/workflow/git/references/workflow.md`, always read (2,123) | **~2,770** | Every commit |
| `work full`, per phase | SKILL (258) + `work.md` (771) + `work-full.md` (1,299) + `check.md` (2,533) + `check-validation.md` (637) | **~5,500** | Every phase |
| `hunt diagnose` / `fix` | SKILL (2,854) + `feedback-loops.md` (731), which step 1 always sends you to | **~3,590** | Every error (auto-triggered) |
| `hunt`, recurring or runtime-state symptom | the above + `failure-patterns.md` (4,053) | **~7,640** | Frequent on hard bugs |
| `hunt triage` | SKILL (2,854) + `triage.md` (1,614) | ~4,470 | Per queue review |
| `think decide` | SKILL (1,984) + `lenses.md` (1,260), which step 5 always sends you to | ~3,240 | Per design question (auto-triggered) |
| `check bounded` | SKILL (262) + `check.md` (2,533) | ~2,800 | Frequent |
| `brainstorm explore` | SKILL (285) + `brainstorm.md` (2,201) | ~2,490 | Frequent |
| `to-plan` | SKILL (235) + `to-plan.md` (1,243) | ~1,480 | Per initiative |
| `handoff` | SKILL (227) + `handoff.md` (1,240) | ~1,470 | Per session |
| `watzup` | SKILL (218) + `watzup.md` (742) | ~960 | Per session |
| `retro` | SKILL (160) + `retro.md` (1,522), plus one mode reference when fixing | ~1,680+ | Rare |

`hunt` and `think` are not projected playbooks: their logic lives in the skill directory, so neither `projection_parity` nor the playbook contract test in `cli/src/embedded.rs` pins them.

## 3. Findings

| # | Waste | Evidence |
|---|---|---|
| F1 | **Each branch loads the whole file.** | `check bounded` reads the preflight, owned state, and "What the Guards Cannot Check" (~1,600 tokens) that only durable modes use. `gate` reads the two-axis review and smell baseline (~470) that only step 5 of `full`/`bounded` uses. `brainstorm explore` reads the Plan Skeleton plus lock steps (~1,400) and the Grill section (~600). `hunt diagnose` stops at step 4 but reads steps 5–7 plus Bisect and Regression (~660). `hunt triage` reads the whole seven-step loop (~1,300) |
| F2 | **`git` always reads its full reference.** | A plain commit needs stage, secret scan, split, and commit. The `pr` and `merge` procedures, Error Handling, Anti-Patterns, and Command Reference ride along on every commit, and much of that is default model behaviour (no-ops) |
| F3 | **Boilerplate repeated across skills.** | The 6 thin triggers repeat "No binary runs the lifecycle. IF the playbook is absent…". All 10 carry both `version` and `metadata.version`, `compatibility`; `hunt` and `think` also end with a `Sources:` line (~60 each) read on every invocation |
| F4 | **A whole reference to read one table.** | `hunt` step 4 says "load `skills/workflow/hunt/references/failure-patterns.md` (gotchas table first)". The Gotchas table is ~245 tokens; the other 22 sections (~3,800) are independent patterns of which one usually applies. `think` step 5 loads all of `lenses.md` although it names the four core lenses (~320) as the set that "almost always applies" |
| F5 | **`AGENTS.md` caches the environment.** | The directory tree is one `ls` away and drifts: adding `hunt`/`think` required editing it. The pipeline already lives in `skills/workflow/README.md` |
| F6 | **Rules restate each other.** | See 2.1. `hunt`/`think` Hard Rules also restate globals ("No placeholders", "Fix the cause inside the authorized scope"), but those are skill-local emphasis and cost only on invocation |

The earlier F4 (dead `hunt`/`think` pointers in `rules/workflow-core.md`) is resolved: both skills landed in `3632c91`.

## 4. Recommendations

Ranked by frequency × savings.

| # | Change | Savings | Risk |
|---|---|---|---|
| R1 | **Slim `git`.** SKILL keeps the commit flow; `pr` and `merge` move to references loaded only for that subcommand; delete no-ops and SKILL/reference duplication | ~2,770 → **~1,100 per commit** | Low |
| R4 | **Slim `AGENTS.md`.** Replace the tree and duplicated pipeline prose with one pointer to `skills/workflow/README.md`; keep Gate Commands and the managed Harness block | ~2,480 → **~1,100 per turn** | Low |
| R7 | **Targeted read of `failure-patterns.md`.** Change the `hunt` pointer to: read `## Gotchas`, then only the `##` section it names (`grep -n '^## '` for offsets). Optionally move the Gotchas table into its own file | recurring bug ~7,640 → **~4,000** | Low: one sentence, no contract test |
| R5 | **Clean `rules/`.** Fold `execution-discipline.md` "Always-on Discipline" into the rules it restates; drop `workflow-core.md` "Workflow", which `karpathy-guidelines.md` §4 and the Harness block already cover | ~2,290 → ~1,900 per turn, every repo | Low |
| R2 | **Split `check.md` by branch.** `check.md`: modes, bounded path, output format. `check-durable.md`: preflight, gate steps, state sync. `check-review.md`: the two axes and baselines, loaded only at step 5. Move "What the Guards Cannot Check" into `check-validation.md` | bounded ~2,800 → ~1,200; gate ~500 less per phase | Medium: contract-test paths change |
| R3 | **Split `brainstorm.md` by branch.** Core: modes, lanes, explore, raw. `brainstorm-grill.md`: the Grill loop. `brainstorm-lock.md`: skeleton and lock steps 5–10 | explore ~2,490 → ~1,000 | Medium: as R2 |
| R8 | **Split `hunt` by mode.** SKILL keeps contract, modes, steps 1–4, hard rules, output. A new `fix.md` reference: steps 5–7, loaded by `fix`. A new `bisect-regression.md` reference: those two sections. `triage` loads `triage.md` without the loop | diagnose ~3,590 → ~2,930; triage ~4,470 → ~3,200 | Low–medium: the `fix` path must still load steps 5–7 |
| R9 | **Inline `think`'s core lenses.** Put the four core lens definitions in step 5; load `lenses.md` only for the Risk/Design/Evidence lenses or `lens <name>` | decide ~3,240 → ~2,300 | Low |
| R6 | Drop repeated boilerplate: the thin-trigger fallback line, duplicate `version`; move `hunt`/`think` `Sources:` into a skill-directory README (keep the MIT attribution in the tree) | ~50–60 per load | Low |

Estimated effect: about 1,780 fewer tokens per turn from `AGENTS.md` plus `rules/`, about 1,700 fewer per commit, about 3,600 fewer on a recurring `hunt`, about 660 fewer per `hunt diagnose`, about 940 fewer per `think decide`, and about 500 fewer per `work full` phase. Prompt caching makes the always-on material cheap to resend in money, so the per-turn gain is mostly a smaller, more legible context, not a proportional bill cut. The per-invocation gains are real cost: that material is new context every time a skill fires.

Suggested order: R1 + R4 + R5 + R7 + R9 first (low risk; every turn, every commit, and every hard bug benefit; R7 and R9 touch no contract test), then R2 + R3 together in a separate PR, then R8 once a real session digest shows `hunt` diagnose/triage volume.

Out of scope, noted: `hunt` and `think` set `model: opus` in frontmatter, which pins the model on every invocation. That is a cost lever larger than any byte cut here and deserves its own decision.

## 5. Safeguards

- **Split by the existing pattern.** `work.md` already routes with "read `docs/playbooks/work-full.md` now", pinned by the playbook contract test. Each new companion gets the same mandatory pointer at its branch point and a contract entry for it.
- **Move pinned phrases byte for byte.** Every phrase the contract test pins moves to its new file unchanged; guards, proof, and judge rules do not change.
- **Measure before and after.** Run `skills/workflow/retro/scripts/session-digest.py` on one or two real sessions per path (peak context, output tokens) alongside the full gate set.
- **Consumers.** R2 and R3 add projected playbooks: each new file must join the canonical playbook list and embed in `cli/src/embedded.rs`, get a contract entry, and a byte-identical copy under `docs/playbooks/` for `projection_parity`; `zharness update` then delivers it as a new managed file. R7–R9 change only skill-directory files, which ship through `npx skills add`, not `zharness`. Nothing existing breaks.
