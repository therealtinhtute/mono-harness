# Spine Playbook Token Audit

Scope: `docs/WORKFLOW.md`, the 6 spine playbooks (`docs/playbooks/*.md`), and their 6 thin `SKILL.md` triggers. Report only: no source file was edited. The optimized set is at [`playbook-token-audit/optimized/`](playbook-token-audit/optimized/).

Token method: bytes ÷ 4 (no tokenizer run). Backtick-dense text tokenizes worse than prose, so treat absolute numbers as ±15%; the percentages are more reliable.

## 1. Executive summary

| Metric | Value |
|---|---|
| Original (13 files) | 54,616 B ≈ **13,650 tokens** |
| Optimized (13 files) | 42,763 B ≈ **10,690 tokens** |
| Reduction | **≈ 21.7%** (≈ 2,960 tokens) |
| Contract proof | `TestOnePlan_PlaybookContract` passes against the optimized playbooks (all ~75 pinned phrases kept, no retired/forbidden string added) |
| Behavioral risk | **LOW** overall; 4 items rated MEDIUM (§7) |

What actually loads per invocation (SKILL + playbook; `WORKFLOW.md` loads only "if routing is unclear"):

| Invocation | Original | Optimized | Δ |
|---|---|---|---|
| `work full` (includes `check.md`, run in-session at step 11) | ≈ 6,650 | ≈ 5,010 | −25% |
| `check` | ≈ 4,200 | ≈ 3,110 | −26% |
| `handoff` | ≈ 1,840 | ≈ 1,470 | −20% |
| `brainstorm` | ≈ 1,810 | ≈ 1,500 | −17% |
| `to-plan` | ≈ 1,410 | ≈ 1,100 | −22% |
| `watzup` | ≈ 1,200 | ≈ 940 | −22% |

`check.md` is the single highest-leverage file: every `work full` phase loads it, and every `check` run loads it.

Bigger savings (≈550–1,100 more tokens on response-only `check` and `bounded` `work` paths) need test or installer changes and are listed as MOVE/TEST in §5, not applied.

## 2. Hard constraint found

`cli/internal/embedded/embedded_test.go` (`TestOnePlan_PlaybookContract`) pins about 75 exact phrases across the 6 playbooks and forbids ~50 retired strings. It shaped the audit:

- Any compression that touches a pinned phrase has to keep it byte for byte. The optimized set does.
- `"Append-only `## Progress` is the sole task execution-status source"` is required in **all 6** playbooks. It is the obvious MERGE-to-`WORKFLOW.md` candidate, and the test blocks it (→ TEST).
- `check.md` forbids `"Before reading any plan, print the resolved mode"`: `check` announces its mode *after* preflight, and `work` announces *before*. That difference is deliberate, so the two steps were **not** harmonized.
- `"check record"` is a retired string. Rewrites avoid it ("clean check entry recorded").

Verification run (scratchpad copy of `cli/`, optimized playbooks + WORKFLOW dropped into `docs/embedded/`):

```text
$ cd <scratch>/cli && CGO_ENABLED=0 go test ./internal/embedded/... -run TestOnePlan_PlaybookContract -v
--- PASS: TestOnePlan_PlaybookContract (0.00s)
    --- PASS: .../brainstorm_locks_honest_bootstrap_state
    --- PASS: .../to-plan_defines_phases_as_markdown_truth
    --- PASS: .../work_appends_durable_markdown_progress
    --- PASS: .../check_preserves_review_intent_and_records_evidence
    --- PASS: .../handoff_closes_phases_before_initiatives
    --- PASS: .../watzup_recaps_without_writing
ok  github.com/therealtinhtute/skills/cli/internal/embedded
```

`TestProjectionParity` was not run in the copy. It fails by design until `cli/docs/embedded/` and `docs/` hold identical bytes (see §9).

## 3. Detailed audit

Legend: K = KEEP, C = COMPRESS, Mg = MERGE, Mv = MOVE, R = REMOVE, T = TEST. Savings are approximate tokens.

### WORKFLOW.md (≈440 → 420)

| Item | Class | Reason | Δ |
|---|---|---|---|
| Authority paragraphs | C | Same rules, fewer words | −15 |
| Stage table, `git`/`interview` line | K | Routing data | 0 |
| Execution boundary + `escalate_when` | K | Shared authority. Not made the only copy, because SKILL.md loads WORKFLOW only conditionally | 0 |

### SKILL.md triggers ×6 (≈1,710 → 1,475 total, ~40 each)

| Item | Class | Reason | Δ |
|---|---|---|---|
| "The lifecycle needs no binary: `zharness` only installs…" (×6) | C | Already in WORKFLOW.md. Cut to "No binary runs the lifecycle.", which keeps the anti-hallucination signal | −25 ea |
| "it holds this stage's operating logic" / "Read WORKFLOW first if…" | C | Parenthetical + IF form | −8 ea |
| "If the playbook is absent…" fallback | K | The one thing a trigger must carry alone | 0 |
| `watzup` "This stage stays read-only." | K | Needed when the playbook is absent | 0 |
| `handoff` "there is no database row to write" | T | Leftover from the DB era, low value. Kept because an agent trained on pre-v0.15 docs might otherwise look for a DB | 0 |
| Argument lines | C | Duplicate `argument-hint`; kept only the default value | −5 ea |
| `Defer to:` lines | K | Routing value | 0 |
| `version` + `metadata.version` | T | Probably skills.sh format compat; frontmatter left untouched | 0 |
| `🥷` prefix | T | Conflicts with the user-global `👾 · ` prefix rule (§5) | 0 |

### watzup.md (≈950 → 720)

| Item | Class | Reason | Δ |
|---|---|---|---|
| Step 1 prose around the shell block | C | The block encodes the fallback chain. Kept one clause for each non-obvious *why* (PR-only clone → exit 128, unborn branch), so nobody "simplifies" the script | −110 |
| Shell block | K | Must stay inline: consumers get no `scripts/` | 0 |
| Output Shape + Exit Conditions | Mg | The exit restated the 5 output items | −45 |
| "the optional block above never runs its own writes" | R | Dangling reference: no optional block exists (leftover from the removed index-sync) | −15 |
| Remaining steps 2–5 | C | Filler cut; all pinned phrases kept | −60 |

### brainstorm.md (≈1,490 → 1,215)

| Item | Class | Reason | Δ |
|---|---|---|---|
| Purpose's owned-section list | Mg | Owned Plan Sections already lists them | −25 |
| Preconditions step 1 + Modes table | Mg | One sentence above the table | −15 |
| `docs/PROJECT.md` unanswered-question rule (step 6, step 9, Exit: 3×) | Mg | Stated once normatively (step 6), once as a check (step 9); Exit references steps 6–9 | −80 |
| Step 6 consumer-repo explanation | C | "(consumer repo)" carries the reason | −30 |
| Exit Conditions | C | Lock exit references steps 6–9 instead of re-listing them | −70 |
| Pinned phrases (13) | K | Contract | 0 |

### to-plan.md (≈1,135 → 865)

| Item | Class | Reason | Δ |
|---|---|---|---|
| Immutability (Owned §, step 3, step 4, step 7, Rules ×2: 6×) | Mg | One invariant bullet | −90 |
| "Sole status source" (step 5 + Rules) | Mg | One invariant bullet (the pinned text stays) | −25 |
| Planning Rules section | Mg | Folded into an Invariants list under Owned Sections | −60 |
| "Waves expose executable coordination; tasks expose exact proof." | R | Motivational; steps 5–6 already require it | −12 |
| Step 9 Handoff | Mg | Duplicated the Exit's next-action clause | −20 |
| Steps 1–8 wording | C | — | −60 |

### work.md (≈2,450 → 1,900)

| Item | Class | Reason | Δ |
|---|---|---|---|
| Modes + plan-count precondition + bounded-rejection paragraph | Mg | One mode list; rejection criteria live on `bounded` | −70 |
| Compaction re-read | C | Kept locally (see §4, G2) | −10 |
| Step 11 prompt-cache rationale | C | Rule kept whole; *why* cut to one clause + audit reference | −110 |
| Step 7 Progress field list | Mg | Owned Plan Sections already defines the fields | −30 |
| Status Routing preamble ("Referenced by steps 6-7…") | R | Step 6 now points to the table | −12 |
| Memory conventions | C | Three triggers + redaction kept whole | −70 |
| Memory conventions location | Mv/T | Only needed when a trigger fires (~180 tok every `work` load). Moving it needs installer support for a new file | (−180 if moved) |
| Exit Conditions parenthetical about `full` | R | Step 11 already says it | −35 |
| `escalate_when` section (4 lines) | C | One line, identical to WORKFLOW.md. Kept local because WORKFLOW is conditional | −20 |
| Awk fallback block | K | Consumers get no scripts | 0 |

### check.md (≈3,900 → 2,850)

| Item | Class | Reason | Δ |
|---|---|---|---|
| Purpose's F1 cache rationale | R | Duplicate of `work.md` step 11, which is the actor that needs it | −70 |
| `full` ⇒ `judge: independent` (Modes, step 7, step 9 hook, Guards, Exit: 5×) | Mg | Normative once (step 7); enforcement once (step 9); limitation once (Guards) | −80 |
| Mode bullets `gate`/`full` re-explaining work/handoff | C | Purpose owns those cross-refs | −60 |
| Validation format mega-paragraph | C | Split into 2 bullets (verdict position, proof bullet shape). Every guard-parsing rule kept | −150 |
| Draft-reshape rationale (old-side hashes) | C | Cut to a parenthetical | −40 |
| Proof re-execution contract | C | "no inherited shell state… no activated virtualenv, no prior `cd`…" said the same thing twice | −60 |
| Steps 8 + 9 "never self-certify / cite real output / REQUEST_CHANGES may cite failing" (also in step 3: 3×) | Mg | Once in step 8; step 3's "Never self-certify" removed | −60 |
| Step 7 hook sentence + step 9 `Optional hook` ladder (2×) | Mg | One enforcement-level statement in step 9 | −90 |
| Step 1 "require a started phase" | R | Preflight (step 2) enforces it for `gate`/`full`, and `work.md` step 11 enforces it on the in-session path | −10 |
| "Response-Only Review and Bounded Gate" section | Mg | Folded into the Zero-write rule; "narrowest checks" and "same fields in the response" kept | −70 |
| What the Guards Cannot Check | C | Bullet 2 repeated the verdict-position rule from the Validation format | −130 |
| Exit Conditions | C | Gate exit references steps 1-4 and 6-11 | −60 |
| Output Format block | K | Output contract, byte-identical | 0 |
| Pinned phrases (27) + forbidden strings | K | Contract | 0 |

### handoff.md (≈1,575 → 1,240)

| Item | Class | Reason | Δ |
|---|---|---|---|
| Precondition 2 rationale ("a summarized turn cannot be assumed…") | R | The rule stands alone | −25 |
| Precondition 1 preserve list + Owned-state preserve list | Mg | One preserve sentence | −25 |
| Anti-Patterns (6 items) | Mg | 5 map 1:1 to steps 5–7. "second continuity markdown" and "continue work is not an exact next action" were folded into step 7's checklist | −110 |
| Step 6 absorb sub-bullets | C | One sentence; all three `absorb:` forms and the stop rule kept | −40 |
| Exit Conditions | C | — | −60 |

## 4. Semantic duplicate groups

| # | Duplicated behavior | Copies (original) | Resolution in optimized set | Why not merged further |
|---|---|---|---|---|
| G1 | "Append-only `## Progress` is the sole task execution-status source" | 6 playbooks + to-plan ×2 | Down to 1 per playbook | Test pins it in all 6 → **T** |
| G2 | Re-read the plan after compaction/summary | work, handoff, check | Kept in each, 1 line | MOVE to WORKFLOW would drop it from context: SKILL.md loads WORKFLOW only "if routing is unclear" → **T** |
| G3 | `escalate_when` | WORKFLOW, work | Identical one-liner in both | Same conditional-load reason → **T** |
| G4 | Exactly one active plan / name every candidate | all 6 | Kept per stage | Each stage's failure action differs (stop / route / list), so they are not true duplicates → **K** |
| G5 | `full` ⇒ `judge: independent` | check ×5, handoff, work | check ×3 roles (rule / enforcement / limitation), handoff ×1 | Handoff needs it as a closure precondition → **K** |
| G6 | Prompt-cache F1 rationale | work, check | work only | — |
| G7 | "zharness plays no part in running a stage" | WORKFLOW + 6 SKILL.md | WORKFLOW full, SKILL.md short form | — |
| G8 | Phase/task definitions immutable | to-plan ×6, brainstorm, work, check, handoff | 1 per playbook | Every writer stage needs its own guard → **K** |
| G9 | Never self-certify / cite real output | check ×3 | check ×1 | — |
| G10 | Plan read by section, never whole file | watzup, work | Kept both | Pinned in both → **K** |

## 5. Context architecture

| Layer | Holds today | Recommendation |
|---|---|---|
| Always loaded (CLAUDE.md / AGENTS.md block) | Read WORKFLOW, read-only vs change, evidence rule | No change. Playbook content must not leak here |
| Shared workflow (`WORKFLOW.md`) | Authority, routing table, execution boundary, escalation | Good home for G1–G3, but **only if** SKILL.md loads it unconditionally (+420 tok per invocation, which cancels most G1–G3 savings). Net ≈ 0, so not worth doing |
| Trigger (`SKILL.md`) | Tone, mode resolution, fallback, defer-to | Correctly thin. The `🥷` prefix conflicts with the user-global `👾 · ` rule; decide which layer wins and state it in one place |
| Stage (`playbooks/*.md`) | Operating logic | Correct layer. Two sub-blocks are over-scoped (below) |
| Dynamic / on demand | Nothing yet | Candidates below |

Over-scoped, stage-layer content (**Mv**, needs installer and test changes, not applied):

1. **`check.md` Validation format + proof re-execution contract + Guards section (≈600 tok optimized).** Only durable `gate`/`full` step 8–9 needs it, yet `review` and `bounded` runs load it every time. Split it into a separate check-validation playbook, read at step 8. Saves ≈550 tok per response-only check. Blockers: `embedded_test.go` asserts `playbook count = 6`, and 2 pinned phrases live in that block.
2. **`work.md` full-mode execution + Status Routing + Memory (≈1,160 tok) during `bounded` runs.** Bounded work needs only Modes + Zero-write + Exit. Splitting full mode out saves ≈1,100 tok per bounded run. Same blockers.
3. **`work.md` Memory section (≈180 tok).** Needed only on a trigger → reference file.

Must stay static and inline: the `watzup` base-branch script and the `work` awk slicer. `zharness install` ships no `scripts/`, so a MOVE-to-script breaks consumers.

## 6. Change diff (optimized vs original)

```diff
WORKFLOW.md
~ Authority/Context/Execution boundary: wording compressed, IF→ form for tooling-vs-playbook conflict

SKILL.md ×6
~ "The lifecycle needs no binary: `zharness` only installs…" → "No binary runs the lifecycle."
~ "Follow X — it holds… Read WORKFLOW first if…" → "Follow X (…); read WORKFLOW first IF…"
~ Argument lines: dropped restated hint prose, kept defaults
= frontmatter, tone line, fallback, Defer-to

watzup.md
~ step 1 prose → 1 sentence + kept shell block
- Exit: "optional block above never runs its own writes" (dangling)
⇄ Output Shape + Exit Conditions → "Output and Exit"

brainstorm.md
⇄ Preconditions folded into Modes; Purpose section list folded into Owned Sections
⇄ PROJECT.md unanswered-question rule 3× → step 6 (rule) + step 9 (check)
~ Exit "Lock" → references steps 6–9

to-plan.md
⇄ Planning Rules + scattered immutability/status rules → one "Invariants" list
- "Waves expose executable coordination; tasks expose exact proof."
⇄ step 9 Handoff → Exit Conditions

work.md
⇄ plan-count precondition + bounded-rejection paragraph → Modes list
~ step 11 cache rationale → one clause + audit ref
~ step 7 Progress fields → reference Owned Plan Sections
- Status Routing preamble; Exit parenthetical on `full`
~ Memory conventions compressed (moved below Status Routing)
~ escalate_when section → one line

check.md
- Purpose F1 cache rationale (kept in work.md)
~ gate/full mode bullets shortened (cross-refs live in Purpose)
~ Validation-format paragraph → intro + 2 bullets; re-execution contract de-duplicated
⇄ step 7 hook sentence + step 9 ladder → step 9 "Declare enforcement honestly"
⇄ steps 3/8/9 self-certify rules → step 8
⇄ "Response-Only Review and Bounded Gate" → Zero-write rule
- step 1 "require a started phase" (enforced by preflight / work step 11)
~ Guards section: dropped restatement of verdict-position rule
= Output Format block byte-identical

handoff.md
- Precondition 2 rationale sentence
⇄ Anti-Patterns → step 5/6 requirements + step 7 checklist
~ absorb sub-bullets → one sentence
```

Why each is safe: every removed sentence either (a) is restated elsewhere in the **same** file, (b) is rationale with no imperative, or (c) refers to something that no longer exists. No cross-file removal depends on `WORKFLOW.md` being loaded. All MUST/NEVER strength is kept (several softened "do not" became NEVER, never the other way). No new behavior: every IF→ rewrite restates an existing condition.

## 7. Risk review

| Change | Risk | Why | Mitigation |
|---|---|---|---|
| handoff Anti-Patterns folded into steps | MEDIUM | Negative lists catch failure modes that positive checklists sometimes miss in LLM compliance | A/B one closure run per `docs/audit/wave-session-ab-protocol.md`; restore the list if the absorb/`git mv` order slips |
| brainstorm and check Exit Conditions by step reference | MEDIUM | An agent may treat "steps 6–9 hold" as done without re-checking each item | Re-expand if a lock ships with an unanswered `docs/PROJECT.md` |
| check Guards section shortened | MEDIUM | "Silence can mean unparsed" is the main anti-overclaim cue; it survives, but with less emphasis | Watch for `proof_gaps: none` on entries the hook did not parse |
| work step 11 rationale shortened | MEDIUM | The rationale is what stops "helpful" dispatch to `/check`; a thinner why may weaken it | The NEVER imperative is kept; A/B if `/check` dispatch reappears |
| check step 1 "require a started phase" removed | LOW | Enforced by preflight (explicit/auto) and `work.md` step 11 (in-session) | — |
| SKILL zharness sentence shortened | LOW | Signal kept | — |
| watzup dangling sentence removed | LOW | Its referent did not exist | — |
| Everything else | LOW | Same-file restatement or pure wording | Contract test passes |

## 8. Pre-existing issues found (not fixed; out of scope)

1. **Contradiction, `work.md` step 11 vs `check.md` step 10.** `work` runs check steps 6–11 in-session, and step 10 sets the phase to `checked`, yet `work` says "Do not mark the phase checked or done". Both were preserved verbatim in meaning. Needs an owner decision → **T**.
2. **`watzup.md` step 2** prefers `scripts/plan-slice.sh` but has no awk fallback, while `work.md` does. Consumers lack the script. Not added, since adding it would be new behavior.
3. **`brainstorm.md` step 7**: "Confirm at most one non-empty plan exists… if one exists, stop" effectively means *zero* must exist. The wording is pinned by the test.
4. **Prefix conflict**: the skills demand `🥷`, the user-global rules demand `👾 · `.
5. **`lifecycle rows`** (DB-era wording) is pinned by the test in brainstorm/work.

## 9. How to apply (if approved)

1. Copy `optimized/playbooks/*.md` → `cli/docs/embedded/playbooks/`, and `optimized/WORKFLOW.md` → `cli/docs/embedded/WORKFLOW.md`.
2. Copy the same bytes to `docs/playbooks/` and `docs/WORKFLOW.md`.
3. Replace each `skills/workflow/<stage>/SKILL.md` body with `optimized/skills/<stage>.SKILL.md` (frontmatter is unchanged).
4. Gate: `cd cli && CGO_ENABLED=0 go build ./... && go vet ./... && go test ./...`, then `bash scripts/verify-doc-links.sh`, then `bash scripts/validate-skill.sh` on the 6 skills.

## Applied

Applied on branch `chore/playbook-token-optimize`, with these deviations from the `optimized/` snapshot, which is left as it was:

- work step 11 and its Exit condition: the in-session gate sets a clean non-final phase `checked`. The final phase stays `in-progress` for an independent `check full`.
- watzup step 2 gains the inline `awk` fallback. brainstorm step 7 reads "Confirm no non-empty plan exists".
- "lifecycle rows" is removed from the zero-write rules and added to the contract test's `retired` list.
- Deep split: full mode moves to `playbooks/work-full.md` and the Validation entry format moves to `playbooks/check-validation.md`, for 8 playbooks. "What the Guards Cannot Check" stays in `check.md`.
- Open: a standalone `check gate` on the final phase still sets `checked`, and `check`'s preflight then rejects `full`.
