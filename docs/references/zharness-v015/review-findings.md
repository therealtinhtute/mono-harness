# Review findings: zharness v0.15 pre-review draft

> **Archive — a record of a read-only review performed 2026-08-26** against
> `docs/references/zharness-v015/v015-original-plan.md` and
> `docs/references/zharness-v015/v013-plan.md`, plus the 0.14.0 surface they describe.
>
> Line numbers in the Evidence column refer to those two archived files and to the
> repository **as of 2026-08-26**. They are historical measurements, not live
> cross-references — v0.15 deletes several of the files cited.
>
> Verdict at the time: **not to-plan-ready - 4 blockers.** All 4 blockers and the 6 majors
> were closed by the `/interview` pass recorded in
> `docs/references/zharness-v015/interview-spec.md`, and the merged result is
> `docs/plans/active/zharness-v015-slim.md`.
>
> Severity key: `B` blocker, `M` major, `m` minor.

---

## Findings

| # | Finding | Sev | Evidence |
|---|---|---|---|
| B1 | S7 (−30% chain cost) has **no requirement behind it**. The audit's own −31% comes 90% from the model-boundary fix; v013 owned it as P4. v015 R1–R10 never mention `model:` frontmatter or check-mode routing — and NG1 forbids touching the files that carry the pins. | blocker | `zharness-v015-slim.md:22` vs `docs/audit/sdlc-token-cache-audit.md:206-213`, `:18`, `:24`; v013 `plan.md:167-171`; pins live at `skills/workflow/check/SKILL.md:5`, `work/SKILL.md:3`, `watzup/SKILL.md:4` |
| B2 | S7's **measurement method is deleted by R1**. The cited method requires `zharness init` + a real `intake → story → run create → trace add` lifecycle measured in CLI output bytes. R1 deletes all five commands. No replacement method is specified. | blocker | `zharness-v015-slim.md:22` + `:36` vs `docs/audit/sdlc-token-cache-audit.md:5` |
| B3 | **NG1 contradicts S1 and S6.** The binary hard-stop lives in the *repo product* files, not only in global copies — all six say "Missing binary: print … and STOP." NG1 declares them untouched and "out of the product path"; S1 requires zero STOP with the binary absent. No v015 requirement covers rewriting them. v013 P0 did, explicitly. (A 7th stop lives in the skills README.) | blocker | `zharness-v015-slim.md:48` vs `:16`, `:21`; `skills/workflow/watzup/SKILL.md:14`, `work/SKILL.md:15`, `check/SKILL.md:16`, `brainstorm/SKILL.md:15`, `to-plan/SKILL.md:14`, `handoff/SKILL.md:14`, `skills/workflow/README.md:39,55`; v013 `plan.md:144-148` |
| B4 | The **second fail-closed guard has no implementing requirement**. S6 counts the independent-judge gate as fail-closed but relocates it "at the playbook layer" — playbook prose is discipline, not a gate. It is enforced today (`independent_judge_required`, lane resolved run→intake). R2 covers proof re-execution only: no lane read, no judge gate. v013 was explicit that the script reads `lane:` from plan frontmatter. | blocker | `zharness-v015-slim.md:21`, `:37` vs `cli/docs/CONTRACT.md:189`; v013 `plan.md:68` |
| M1 | R2's **"script pass marker" is undefined and forgeable**. Today verification is inside the writer — proofs re-execute before any DB write and a failure rejects the whole call. Under R2 the same agent runs the script *and* writes the entry; the hook checks a marker, not execution. The audit names this exact mechanism as what closed a real bypass. CI re-run demotes prevention to post-hoc detection. | major | `zharness-v015-slim.md:37` vs `cli/docs/CONTRACT.md:189`; `docs/audit/sdlc-token-cache-audit.md:201`; commit `c53fb76` |
| M2 | **R1's kill list is wrong against the 0.14.0 baseline.** Missing: `plan complete` / `plan abandon` (registered, real, but absent from CONTRACT — which is why the list missed them). Miscount: "memory×3" — four subcommands exist. Phantom: `status` and `doctor` don't exist in 0.14.0; they're *v013's proposed new verbs*. R1 was derived from v013's target surface, not the 0.14.0 codebase it claims to re-lock against. (`query×9` and `intervention` do check out.) | major | `zharness-v015-slim.md:36`, `:33` vs `cli/internal/interfaces/root.go:59`, `plan.go:16,21,30`, `memory.go:22,40,51,73`, `db.go:39`; v013 `plan.md:63,70` |
| M3 | **S5 depends on `docs/PROJECT.md`, which no requirement creates** — and which doesn't exist. v013 owned it via the K0–K5 taxonomy and the P2a greenfield scaffold with brainstorm-lock as the forcing step; v015 dropped the whole taxonomy without listing it as a rejection, and R1 additionally deletes `init` and `scaffold`. The existing elicitation form is a *different* file, `docs/ARCHITECTURE.md`; nothing reconciles the two. | major | `zharness-v015-slim.md:20` (sole mention), `:34`, `:36` vs v013 `plan.md:108,119-120,158-161`; `cli/docs/CONTRACT.md:219` |
| M4 | **Brownfield and greenfield have no mechanism.** R10 mandates "deterministic detection → drafted proposal → human approval" but v013 assigned detection to `doctor --adopt`, and NG7 bans `doctor`. R3's three-way merge is an *update* path; `.zharness/` does not exist here, and no requirement covers first-time scaffold of playbooks, templates, or the plan skeleton. | major | `zharness-v015-slim.md:45`, `:54`, `:38` vs v013 `plan.md:71,119,127` |
| M5 | **Playbook rewrite is in the Outcome but in no requirement.** The six playbooks carry 65 `zharness` invocations across 391 lines, nearly all naming commands R1 deletes. R5 covers only AGENTS.md / CLAUDE.md / codex. R8's gate (`go test`, `verify-doc-links.sh`) cannot catch stale command names in prose. | major | `zharness-v015-slim.md:14`, `:40`, `:43`; `docs/playbooks/work.md` (19 refs), `handoff.md` (16), `check.md` (11), `brainstorm.md` (8), `to-plan.md` (6), `watzup.md` (5); v013 `plan.md:144,147` |
| M6 | **R8's phase gate evaporates with the deletions it is meant to guard.** `go test ./...` covers 68 test files / 12,770 lines, 17 of them in `interfaces/` — the layer R1 deletes wholesale. The gate stays green by having nothing left to test. CI runs *only* `go test ./...`, so R2's "CI re-runs it" also has no existing surface. | major | `zharness-v015-slim.md:43`, `:37` vs `.github/workflows/cli-ci.yml:34` |
| m1 | Sole Authority anchor is a **gitignored file**. `verify-doc-links.sh` passes today only because the untracked copy is on this machine; on a clone or in CI, the sole source for decisions 1/2/4 does not exist. The plan file itself is also still untracked while declaring `status: active`. | minor | `zharness-v015-slim.md:26`, `:6` vs `.gitignore:6`; `git status`: `?? docs/plans/active/zharness-v015-slim.md` |
| m2 | `approach: not-planned`, `constraints: none`, `risks: none` on a `lane: high-risk` plan, while v013 carried 4 accepted risks. High-risk lane requires an independent judge at check time; an empty risk register is a missing to-plan input, not cosmetics. | minor | `zharness-v015-slim.md:57-61`, `:5` vs v013 `plan.md:173-178` |
| m3 | R4 is thinner than 0027 on the **pin mechanism itself** — it covers consumer bytes and CHANGELOG but never requires tagging/publishing 0.14.x as the archive release, and nothing stops the updater from carrying a pinned consumer past it (`MIN_ZHARNESS_VERSION` is a floor, not a ceiling). R7 specifies no memory export path for consumers with rows (this repo has 0, so no loss here). | minor | `zharness-v015-slim.md:39`, `:42` vs `skills/workflow/README.md:39`; v013 `plan.md:13` |
| m4 | Foreign decisions cited as bare numbers ("decision 0027", "decision 0020") against a repo with its own `docs/decisions/0001–0005`. v013 namespaced them. | minor | `zharness-v015-slim.md:39,45` vs v013 `plan.md:20,23`; `docs/decisions/` |
| m5 | The F3 residue v013 flagged is still live: `AGENTS.md:4` mandates `zharness --version`, which the skills README says was removed; `AGENTS.md:8` still states durable stages "require an initialized database" — the exact sentence S1 needs gone. R5 will fix it; noted so to-plan sequences it before the kill-switch test. | minor | `AGENTS.md:4,8`; `skills/workflow/README.md:39`; v013 `plan.md:38,98` |

## v013 → v015

| v013 decision | v015 | Rationale stated? | Assessment |
|---|---|---|---|
| D1 kill SQLite | **kept**, hardened (R1/S4/NG5) | yes — `:34` "index's only consumer is being deleted" | Sound. Strongest part of the plan. |
| D2 `record check` = sole write | **changed**: binary verb → `scripts/record-check.sh` + hook + CI | yes — `:34` "a binary command is not CI-enforceable" | Rationale is right; the **guarantee change is unanalyzed** (M1). Moving the check outside the writer makes it forgeable. |
| D4 EOL / pin / never delete `harness.db` | **kept** (R4) | yes — 0027 | Adequate on consumer bytes; **pin mechanism unspecified** (m3). |
| D5 global instruction merge (P3) | **dropped** (NG1/R6) | yes — owner decision `:33` | Legitimate as scope. But it silently removed the only vehicle that rewrote the 6 SKILL.md → **B3**. |
| 3 verbs `status`/`record check`/`doctor` | **changed** → installer-only | yes — `:34` | Defensible. But `status`/`doctor` owned scaffold, adopt-detection, and self-describing stages; only the 3-way merge survives into R3 → **M3, M4**. |
| K4 memory in SQLite | **changed** → `docs/memory/{id}.md` (R7) | yes — `:33,42`, loss accepted | Fine here (0 rows). Consumer export unspecified. Note it discards `supersede` shipped in `06dd6f5`. |
| K0–K5 knowledge taxonomy | **dropped** | **no** — not in the rejected-alternatives list `:34` | Silent drop. S5 survived the drop but its artifact did not → **M3**. |
| P4 model policy / re-measure | **dropped** | **no** — not listed as rejected | Silent drop while S7 was kept verbatim → **B1, B2**. The single largest cost lever left unowned. |
| P0 six SKILL.md STOP removal | **dropped** | **no** — NG1 asserts the opposite | → **B3**. Directly negates S1 and S6. |
| L0–L3 fail-open ladder | **dropped** | no | Largely moot: with an installer-only binary there is no CLI left to degrade, so L0–L2 collapse into L3. Acceptable, but "fail-open" is now a title with no defined contract. |
| Research anchors / audit evidence | **carried by reference** `:26` | yes | Reference-only carry to a gitignored path → **m1**. |

**Net:** everything v015 *kept* is well reasoned and better argued than v013. Everything it *silently dropped* is what three of its own success signals were standing on.

## Recommendation

**Keep the plan, do not send it to to-plan yet. One `brainstorm --mode refine` pass first.** Rewriting it wholesale would discard genuinely better reasoning at `:34`; running to-plan now would generate phases for S1/S5/S6/S7 out of requirements that cannot produce them, and phases are immutable after to-plan (`:64`).

Close in the refine pass, in this order:

1. **B3 first** — it decides everything else. Either NG1 narrows to "the *installed* `~/.claude/skills` copies" and a new R covers rewriting the 6 repo `skills/workflow/*/SKILL.md` + `README.md:39,55`, or S1 and S6 come out. Same edit resolves B1's blocker on the `model:` pins.
2. **B1/B2 together** — add an R owning the model-boundary fix (audit P2, `sdlc-token-cache-audit.md:174-180`) and a restated measurement method that survives R1, or drop S7 to a non-goal. As written it is unmeasurable *and* unachievable.
3. **B4 + M1** — one requirement specifying the record-check script's full contract: lane read from plan frontmatter, judge gate, and a marker bound to the commit/diff rather than a flag the author writes. If the marker can't be made unforgeable, say so and re-rate S6 as "1 fail-closed guard + 2 detections."
4. **M3/M4/M5** — three requirements: PROJECT.md creation + forcing step; who performs scaffold and adopt-detection now that `init`/`scaffold`/`doctor` are gone; playbook rewrite scope (65 call sites).
5. **M2** — re-derive R1 from `root.go`, not from `CONTRACT.md`. Add `plan complete`/`plan abandon`, correct memory to ×4, drop the phantom `status`/`doctor`.
6. **m1** — commit the plan, and copy v013's decisions/anchors table into v015 or into `docs/decisions/` so the Authority survives a clone.

M6 and m2 belong in to-plan's own output (replacement gate, risk register), not the refine pass.
