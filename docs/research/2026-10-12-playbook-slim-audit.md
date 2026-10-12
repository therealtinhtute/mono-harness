# suhuynh-pbslim — playbook token audit

Read-only audit of `cli/docs/embedded/` (worktree `quaqua/suhuynh-pbslim` @ c8a8429). No edits, no commits.
Sizes from `wc -c`; pinned-phrase bytes computed from `cli/src/embedded.rs:239-509`.

**Verdict:** a ≥40% cut per stage is reachable with no lifecycle change. Most of the bulk is not spine-vs-companion duplication. It comes from three kinds of prose: restated Exit Conditions, rationale and explainer sentences, and cross-stage invariants that every playbook repeats. Keep progressive disclosure, because the companion split is what keeps `bounded` paths cheap. Shrink the spines, merge only `check-validation.md` into `check-durable.md`, and keep every step number stable.

## Changes

None. Read-only. Deliverable: this report.

## Test command + result

No code changed, so the gate was not run. Evidence commands:

- `wc -c cli/docs/embedded/*.md cli/docs/embedded/playbooks/*.md skills/workflow/*/SKILL.md`: embedded total **48,090 B** (12 playbooks 44,793 + WORKFLOW 2,187 + AGENTS 1,110).
- Pinned-bytes script over `embedded.rs` `required:` arrays: per-file floor in the table below.
- Duplication: a `grep -rlF` sweep over the embedded set (counts below).

## Challenge

Premise: "one full cycle ~36 KB". Measured, a `normal` single-phase cycle reads **53.3 KB**. The breakdown is lock 9.7, to-plan 5.0, work full 8.7, same-session gate 11.5 (check + durable + validation, `work-full.md` step 8), independent `check full` 13.5 (fresh agent, cold re-read) and handoff 5.0. The gate path costs more than the brief assumed, because `check` files are read twice per cycle: in-session and by the independent judge. Shims add 7.4 KB, and WORKFLOW adds 2.2 KB when it is read.

## Per-file table

| File | Now B | Pinned B | Target B | What to cut (file:line) |
|---|---|---|---|---|
| AGENTS.md (block) | 1,110 | 0 | 700 | `AGENTS.md:20-21` binary/no-DB history (in every consumer turn, so the highest cost per byte); `:3-5` overlaps `WORKFLOW.md:7` |
| WORKFLOW.md | 2,187 | 0 (parity only) | 1,500 | `:13` restates `:3`; `:26-30` tier table duplicates shim `model:` frontmatter, so keep one row; `:36` cite-by-step rationale → 1 clause |
| brainstorm.md | 3,158 | 246 | 1,800 | Exit `:40-46` (461 B) restates the mode-table "Durable effect" column; `:5` purpose restates the table; `:23-31` lanes: keep the table, drop `:25` |
| brainstorm-grill.md | 3,009 | 175 | 1,700 | `:14` question-tool format (~400 B) duplicates all 6 shims `:13`, so make it 1 line; `:11` design-tree prose; `:17` stress-test → 1 line; `:19` Done restates `:7` field list |
| brainstorm-lock.md | 3,531 | 447 | 2,100 | `:3` header after pinned prefix; `:41` section-ownership paragraph (~520 B), so keep 1 line; `:49` self-review restates grill Done + `:46`; `:46` PROJECT.md fallback chain → 1 line |
| to-plan.md | 4,971 | 518 | 2,800 | Exit `:59-61` (310 B) restates step 8; `:46` step 1 field enumeration → "read `## Goal`"; `:49-53` tracer-bullet 4 sub-bullets (~600 B) → 1 line; `:47-48` rationale clauses; `:9` preconditions dup of one-plan rule |
| work.md | 3,082 | 167 | 1,500 | Exit `:32-37` (573 B): full exit duplicates `work-full.md:34`, and `:37` escalate_when duplicates `WORKFLOW.md:38` verbatim; Memory `:22-30` (~800 B) → 2 lines (ADR 0003: opt-in, so keep only the triggers); `:5` purpose restates modes |
| work-full.md | 5,640 | 299 | 3,000 | Status Routing table `:36-43` (~550 B) duplicates `:31-32`; `:30` sub-agent delegation (~550 B) → 1 line; `:34` step 8 (~1,050 B): prompt-cache rationale duplicates `WORKFLOW.md:32`, and the status-transition prose duplicates `check-durable.md:19-20` and `handoff.md:26-31`; `:15-19` plan-slice.sh/awk dual path (also `watzup.md:26`) → awk only; `:25-28` red/green → 2 lines |
| check.md | 4,770 | 709 | 2,500 | Exit `:46-50` (380 B) restates steps; `:5` purpose: keep only its 2 pinned clauses; `:23` record-check fallback prose → 1 line; `:21` step 1 overlaps durable preflight |
| check-durable.md + check-validation.md → **merged check-durable.md** | 3,680 + 3,001 = 6,681 | 902 + 441 | 3,700 | Merge, because a durable gate always loads both (`check.md:14` → `check-durable.md:16`), so the split only adds 2 headers and a hop. `check-durable.md:17` enforcement (~550 B) → 1 line keeping the pinned phrase; `:11` owned-state dup; `check-validation.md:19` restates the template at `:7-17`; `:24` compaction exception duplicates `handoff.md:34`; `:30` guard-limits → bullets |
| check-review.md | 2,021 | 141 | 1,200 | `:5` sub-agent opt-in decision tree (~650 B) → 1 line; `:8` 12-smell list → "Fowler smells" (labels only, judgement already stated) |
| handoff.md | 4,961 | 604 | 2,800 | Exit `:39-43` (403 B) restates `:25-35`; `:12-18` Owned Plan State duplicates steps 5/6; `:36` "completed plan is a run log…" (~250 B); `:9-10` preconditions dup step 2; `:37` verify list overlaps `:28` |
| watzup.md | 2,969 | 254 | 1,700 | `:14` base-branch explainer prose (~350 B): keep the sh at `:17-22`, drop the narration; Output `:31-39` restates steps 1-5 → merge; `:26` plan-slice dual path |
| **Total embedded** | **48,090** | — | **~27,000 (−44%)** | |

The pinned floor stays under 25% of every target, so the cuts fit around contract phrases.

## Top duplications (measured with `grep -rlF`)

1. "`## Log` is the sole task execution-status source": 6 files (to-plan, work-full, check-durable, brainstorm-lock, handoff, watzup). The phrase is **pinned in all 6** (`embedded.rs`). Keep it as the one-liner it is and drop the surrounding re-explanation, for example `brainstorm-lock.md:41` and `to-plan.md:41-42`.
2. One-active-plan / "non-empty" / list-candidates-and-stop: 6 files (`to-plan.md:9`, `work.md:11`, `check-durable.md:7`, `watzup.md:9`, `brainstorm-lock.md:47`, `work-full.md:21`).
3. Exit Conditions restating Steps: 6 files, 2,127 B total (`brainstorm.md:40`, `check.md:46`, `handoff.md:39`, `to-plan.md:59`, `work.md:32`, plus `work-full` routing table `:36`).
4. Phase-status transitions (in-progress → checked → done) narrated in 3 places: `work-full.md:34`, `check-durable.md:18-20`, `handoff.md:26-31`. Only the owner of each transition needs to state it.
5. Plan slicing (plan-slice.sh else awk): `work-full.md:15-19`, `watzup.md:26`.
6. Re-read-after-compaction: `work.md:14`, `handoff.md:10`, `check-durable.md:7`.
7. Mode print + zero-write rule: `work.md:13,16`, `check.md:15,17`, `brainstorm.md:21`.
8. Verbatim duplicates: escalate_when (`WORKFLOW.md:38` = `work.md:37`) and the prompt-cache rationale (`WORKFLOW.md:32` ≈ `work-full.md:34`).
9. Question-tool format: `brainstorm-grill.md:14` plus all 6 stage shims at `:13` (302 B each, 1.8 KB). The shim copy serves Claude; the playbook copy serves other agents. Cut both to 1 short line.
10. "Loaded by …" headers: 6 companions. Four are pinned by prefix only (`embedded.rs`). Keep the pinned prefix and drop the rest of each line.

## Proposed structure

- **Keep the spine + companion split.** It is what keeps the cheap paths cheap: work bounded never loads work-full, and check bounded never loads durable. Folding companions into spines would raise the cost of `bounded`/`explore`, the most frequent calls.
- **One merge:** `check-validation.md` → `check-durable.md`. A durable gate always loads both.
- **Spine = routing + checklist:** Purpose ≤1 line, a mode/lane table, numbered steps (imperative, ≤2 lines each) and the Output block where one exists. Drop the Exit Conditions section; the last step is the exit check.
- **Each invariant has one owner.** A transition or section rule is stated only in the playbook that writes it. Other playbooks get a ≤1-line pointer, or nothing.
- **Rationale out.** "Why" sentences (prompt cache, guard internals, run-log philosophy) move to the ADRs they already cite.
- **Keep step numbers stable.** ADR 0012 cites `handoff.md` step 6, and the spines cite `check-durable.md` step 8 and `work-full.md` step 8. Renumbering would break those citations.
- **Rejected: a shared "plan invariants" kernel in WORKFLOW.md.** It would make WORKFLOW a mandatory read: +~2 KB on every stage, more than the per-file dup it removes (~150-300 B each).

## Per-stage before → after (playbook bytes only)

| Stage path | Before | After | Δ |
|---|---|---|---|
| brainstorm explore | 3,158 | 1,800 | −43% |
| brainstorm grill/raw | 6,167 | 3,500 | −43% |
| brainstorm lock/refine | 9,698 | 5,600 | −42% |
| to-plan | 4,971 | 2,800 | −44% |
| work bounded | 3,082 | 1,500 | −51% |
| work full | 8,722 | 4,500 | −48% |
| check bounded (no diff / diff) | 4,770 / 6,791 | 2,500 / 3,700 | −48% / −46% |
| check gate | 11,451 | 6,200 | −46% |
| check full | 13,472 | 7,400 | −45% |
| handoff | 4,961 | 2,800 | −44% |
| watzup | 2,969 | 1,700 | −43% |
| normal 1-phase cycle | 53,275 | ~29,300 | −45% |

## Contract-test impact (`cli/src/embedded.rs`)

- `:125` `playbook_count_is_twelve` → 11 (merge).
- The `check-validation.md` contract (10 phrases) moves under the `check-durable.md` path. Drop `"read \`docs/playbooks/check-validation.md\`"` from the check-durable required list. Drop or reword `"Every Validation entry must include timestamp, stable phase slug, exact command/result and concise output"`, which restates the template.
- Every other required phrase (5+14+5+13+3+8+16+19+5+14+7 = 109) **can stay verbatim**. The pinned floor is ≤25% of each target, so the cut works around them with 0-2 more phrase edits per file.
- `RETIRED` and `forbidden` lists: no impact (cuts only remove text).
- `cli/tests/projection_parity.rs`: copy the bytes to `docs/playbooks/` and **delete `docs/playbooks/check-validation.md`**. Otherwise parity fails with "projected file has no embedded counterpart" (`projection_parity.rs:74`).
- Installer tests pin only the `work.md`/`watzup.md` names (`installer/mod.rs:611,696`; `update.rs:262`), so no impact.
- Other refs to `check-validation.md`: `CHANGELOG.md` (historical, leave as is) and both `check-durable.md` copies.

## Effort

~3-4 h for one agent: 12 rewrites + parity copy + 1 test edit + `cd cli && cargo test` + `bash scripts/verify-doc-links.sh`. Recommended order: start with the **check gate path**, the largest file set and read twice per cycle, as a ≤30 min spike. If it lands at ≤6.2 KB with tests green, do the rest.

## Risks

- **Orphan file in consumer repos.** `zharness update` never deletes a playbook it stopped shipping (no remove path in `installer/update.rs`), so consumers keep a stale `docs/playbooks/check-validation.md`. That file is harmless but misleading. Either add a retire-list removal (a new installer behavior that needs an owner decision) or skip the merge and keep the 2 files (cost: ~300 B on gate/full only).
- **Behavior drift from over-compressed guard text.** `check-validation.md:21-26` (anchored `verdict:`, backtick proof sub-bullets, re-exec contract) feeds the frozen pre-commit guard. Compress only the prose; keep every rule.
- **Contract tests pin text, not behavior.** Green tests will not prove agents still behave the same. A dry-run cycle on a scratch plan after the rewrite is the real check.
- **Shim edits** (question-tool line) touch `skills/workflow/*` versions and run the skill validator hook.

## Lessons

- Exit Conditions sections in playbooks restate their Steps (~2.1 KB across 6 files); a stage doc's last step should be its exit check.
- `zharness update` cannot retire a shipped playbook: any file merge/rename leaves orphans in consumers.
