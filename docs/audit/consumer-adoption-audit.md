# Consumer Adoption Audit — the harness as actually run

**Date:** 2026-08-17
**Scope:** the `zharness` harness (`cli`, `skills/workflow`, `docs/playbooks`) as adopted by one real consumer repository, `onedrive-cloud` (private, read-only for this audit — cited as evidence, never edited).
**Method:** live measurement against the consumer's own committed state and its `harness.db` at HEAD `0b32adb`. Every `zharness` invocation was run in that repository and its stdout measured in bytes. No session transcript was replayed; where a number is inferred rather than observed, it is labelled so.
**Primary metric:** tokens resident in context at the moment a stage begins work — the "cold entry cost" of a stage. Ceremony (operation count) is not re-measured here; `docs/audit/workflow-harness-ceremony-audit.md` owns it.
**Relationship to prior audits.** Extends `docs/audit/sdlc-token-cache-audit.md` with the dimension it could not see: what happens to a harness after months of real adoption, once lifecycle state has accumulated and an invariant has silently broken. That audit measured a clean throwaway repository. This one measures a dirty one.

### Caveats

Byte counts are exact. Token counts use 4 chars/token and carry roughly 15% error; every conclusion rests on ratios computed with the same divisor, so the error largely cancels. Phase and trace counts are the consumer's actual DB rows as of the measurement date and will keep growing — the D2 finding is specifically about that growth, so a re-measurement later will show larger, not different, numbers.

---

## 1. Headline

The architecture passes. The adoption does not.

Measured against the four necessary-and-sufficient conditions for an agent harness (agent loop, tool interface, context management, control mechanisms), this harness satisfies all four at runtime, and its control layer is stronger than typical — `check record` re-executes every cited proof command and requires exit 0 before recording an APPROVED verdict, which is a direct defence against the reported-success-without-verification failure that motivates the whole category.

What fails is the gap between what the harness assumes and what it enforces. One CLI command depends on an invariant ("exactly one active plan") that nothing in the system checks, warns about, or repairs. In the consumer repository that invariant has been broken since 2026-08-16, and the resulting cost is roughly 26,000 tokens per `work` invocation.

| # | Finding | Measured cost | Nature | Fix risk |
|---|---|---:|---|---|
| **D1** | `query plan` hard-errors on 2 active plans; no playbook recovery path exists for an error (only for `degraded`) | **~26,400 tok/invocation** (inferred) | Unenforced invariant + missing recovery branch | Low |
| **D2** | `context.phases` in the preflight packet is unbounded — no cap, no status filter, no `Omitted` declaration | 1,075 tok/call today, **grows without limit** | Policy applied to `traces` but not `phases` | Low |
| **D3** | Lifecycle integrity broken: 5 phases `in-progress` simultaneously; `out_of_order` drift unresolved | forces per-stage drift reconciliation | Consumer state, not harness code | None (state repair) |
| **D4** | Consumer `CLAUDE.md` is 349 lines / 3,169 tok, mostly restating what the repository already shows | 3,169 tok **per turn**, not per stage | Consumer context discipline | None |

D1 and D2 are harness defects and belong in this repository's backlog. D3 and D4 are consumer hygiene, recorded here because they are what made D1 and D2 visible.

---

## 2. What the harness gets right

Recorded first, because the defect list below is longer and would otherwise misrepresent the system.

**Progressive disclosure is real, not claimed.** The six spine `SKILL.md` files measure 276–692 tokens each; the playbooks they route to measure 1,106–2,746. The expensive half loads only when a stage actually runs. This is the structure Anthropic's own guidance recommends and most hand-built harnesses skip.

| Spine skill | SKILL.md | Playbook | Loaded together only when the stage runs |
|---|---:|---:|---|
| `watzup` | 283 | 1,106 | |
| `brainstorm` | 349 | 1,414 | |
| `to-plan` | 302 | 1,329 | |
| `work` | 344 | 2,746 | |
| `check` | 335 | 2,558 | |
| `handoff` | 276 | 2,019 | |

**Proof verification is enforced in the CLI, not requested in a prompt.** `check record` re-runs each `--proof-links` command and rejects the record with `proof_verification_failed` if any exits non-zero. A model cannot talk its way past this, which is the correct place to put the boundary.

**The bounded-packet policy exists and is well specified.** `cli/internal/application/context.go` defines `contextTraceTail = 30` and an `OmittedField{Field, Reason, Fetch}` contract requiring any bounded packet to declare what it cut and how to retrieve it. The design is right; D2 is that it was applied to one field and not its neighbour.

**The prompt-cache finding from the prior audit shipped correctly.** `docs/playbooks/work.md` step 11 now performs the phase gate in-session rather than dispatching to the `check` skill, with the reasoning recorded inline: caches are model-scoped, and `check` pins `model: opus`, so a per-phase dispatch paid a cold cache every phase. This is the audit-to-fix loop working as intended.

---

## 3. Cold entry cost of one `work` invocation

Consumer repository, `zharness` 0.9.1, before a single line of product code is read.

| Layer | Tokens | Frequency | Status |
|---|---:|---|---|
| Global `CLAUDE.md` + 5 files in `rules` + memory index | 2,305 | every turn | expected |
| Consumer `CLAUDE.md` (349 lines) | 3,169 | every turn | **D4** |
| `skills/workflow/work/SKILL.md` | 344 | per stage | correct by design |
| `zharness preflight work --json` | 2,595 | per stage | **D2** (1,075 of it) |
| `docs/playbooks/work.md` | 2,746 | per stage | acceptable |
| Whole-plan fallback read after `query plan` fails | 26,365 | per stage | **D1** (inferred) |
| **Total** | **~37,500** | | |

Add `docs/playbooks/check.md` when step 11's in-session gate runs and the figure reaches roughly 40,000 tokens of resident context before productive work begins.

Recoverable without touching playbook content: approximately 29,500 tokens, or 78%.

### Preflight payload by stage

Measured directly, same repository, same moment:

| Stage | Bytes | Tokens | Carries a context packet |
|---|---:|---:|---|
| `brainstorm` | 148 | 37 | no |
| `to-plan` | 142 | 35 | no |
| `watzup` | 6,254 | 1,563 | yes, without `phases` |
| `work` | 10,381 | 2,595 | yes |
| `check` | 10,383 | 2,595 | yes |
| `handoff` | 10,387 | 2,596 | yes |

Packet composition for `work`:

| Field | Bytes | Tokens | Entries |
|---|---:|---:|---:|
| `phases` | 4,300 | 1,075 | 18 |
| `traces` | 5,697 | 1,424 | 11 |
| `drift` | 284 | 71 | 1 |
| position + latest IDs + readiness | 278 | 69 | — |

`traces` is bounded at 30 and correctly declares nothing omitted at 11 entries. `phases` is not bounded at all.

---

## 4. D1 — `query plan` is broken by an unenforced invariant

### Observed

```
$ zharness query plan --section phase --phase polish-sweep --json
{"error":{"code":"ambiguous_active_plan","message":"query plan: 2 active plans found; this command requires exactly one"}}
```

Two files under the consumer's active-plan directory carry `status: active`:

| Plan | Lines | Tokens | Frontmatter | Reality |
|---|---:|---:|---|---|
| `onedrive-cloud/docs/plans/active/ui-ux-audit-remediation.md` | 1,621 | 26,365 | `status: active`, `updated: 2026-08-16` | commit `0b32adb` states its ten phases were applied |
| `onedrive-cloud/docs/plans/active/check-review-remediation.md` | 410 | 7,442 | `status: active`, `updated: 2026-07-30` | untouched for 18 days |

### The two defects behind it

**First: the invariant is assumed but never enforced.** `QueryPlanSection` in `cli/internal/application/plan_query.go` documents its own precondition — it "resolves the single active plan" — and raises `ambiguous_active_plan` when that fails. Nothing upstream protects it:

- `preflight` reports `docs: "ready"` with two active plans present. Its docs-status check does not count them.
- `brainstorm` has no guard against locking a second plan while one is already active.
- `resume` surfaces `out_of_order` drift but has no drift type for multiple active plans.

An invariant that one command depends on, that no other component checks, is a latent failure waiting for the first user who starts a second initiative before closing the first. That is ordinary usage, not misuse.

**Second: the playbook defines recovery for the wrong failure.** `docs/playbooks/work.md` step 1 instructs:

> If `query plan` reports `degraded: true`, read the plan file directly for this phase's definition.

`degraded: true` is the *graceful* path — the command succeeded and returned the whole file because a heading did not match. An `error` response is a different case, and the playbook is silent on it. With no defined branch, the only remaining source for the phase definition is the plan file itself.

### Cost

**Inferred, not observed.** No session transcript was replayed for this audit. The 26,365-token figure is the size of the larger active plan, which is what a full-file fallback would cost. What is certain is that the command fails and that the playbook prescribes no bounded alternative; the exact recovery an agent improvises may vary.

The fix reframes the cost: a silent 26,000-token fallback becomes an explicit ~50-token stop naming the broken invariant and its repair.

---

## 5. D2 — the context packet's `phases` field is unbounded

`cli/internal/application/context.go` builds the packet. Its two list fields are treated differently:

```go
if contextPhasesStages[stage] {
    phases, err := QueryPhases(db)   // no cap, no status filter, no Omitted
    pkg.Phases = phases
}

if resumeView.Position.CurrentPhase != nil {
    traces, err := QueryTracesByPhase(db, phaseSlug, contextTraceTail)  // capped at 30
    pkg.Traces = traces
    if total > contextTraceTail {
        pkg.Omitted = append(pkg.Omitted, OmittedField{ /* ... */ })    // declares the cut
    }
}
```

`QueryPhases` in `cli/internal/application/query.go` runs `SELECT ... FROM stories ORDER BY created_at, slug` with no `LIMIT` and no status predicate.

The file's own doc comment states the policy this violates: bounding exists to prevent the packet growing with a long-running initiative's history, and R5 requires any bounded packet to declare what it omitted. `traces` honours both. `phases` honours neither.

### Current consumer state

18 phases, of which 10 are terminal:

| Status | Count | Actionable |
|---|---:|---|
| `checked` | 6 | no |
| `done` | 4 | no |
| `in-progress` | 5 | yes |
| `planned` | 3 | yes |

Every `work`, `check`, and `handoff` preflight serialises all 18 in full. The 10 terminal ones — 597 of the field's 1,075 tokens — describe work that is finished and cannot be selected. This cost is paid on every stage entry, forever, and rises with every phase the repository ever completes.

---

## 6. D3 — lifecycle integrity in the consumer

Not a harness defect. Recorded because it is the state that exposed D1 and D2, and because the harness detected part of it and could have detected the rest.

**Five phases are `in-progress` at once:**

```
2026-07-30  in-progress  credential-incident-remediation
2026-08-16  in-progress  copy-and-truncation
2026-08-16  in-progress  list-view-semantics
2026-08-16  in-progress  polish-sweep
2026-08-16  in-progress  status-color-tokens
```

The first has been open for 18 days and belongs to the stale plan in D1.

**Drift is detected but unresolved.** `zharness resume --json` returns `readiness: "drifted"` with one `out_of_order` finding: the latest check belongs to run `01M059KMWVKF87DC865PX9ZP14` while `latest_run_id` is `01M05FMSTHNPXJ1PTYQA3YNSYR`. The recovery string is present and correct. It has not been run.

Because `docs/playbooks/work.md` step 1 treats DB-versus-plan disagreement as "a stop requiring reconciliation", every stage entry now begins with an investigation the harness already diagnosed and offered to fix.

**Stale state file.** `.kit/workflow-state.yml` remains in the consumer with `last_updated: 2026-07-18`, pointing at a `.kit/planning/phases/...` layout the harness no longer uses. This repository's own `CLAUDE.md` states state is "not a hand-edited `workflow-state.yml` pointer file". The file is gitignored and inert, but it is a second, contradictory answer to "where is lifecycle state" for any agent that reads it.

---

## 7. D4 — consumer `CLAUDE.md` weight

349 lines, 12,676 bytes, ~3,169 tokens, resident on **every turn** — the highest-frequency layer in the stack, and therefore the highest-leverage one.

Section headings, verbatim: Development Commands, Architecture Overview, Key Directories, API Routes (App Router), State Management (Zustand), Authentication & Security, Component Architecture, Component Patterns, API Route Best Practices, Configuration Files, Configuration Management, Code Style & Conventions, Styling Best Practices, File Handling Features, Development Workflow, Testing (When Implemented), Important Notes, Project-Specific Patterns, Design System Deviations.

Roughly half describe what the repository already discloses to anyone with `Read` and `Glob`: commands live in `package.json`, directories are visible, App Router routes are the directory tree. `docs/prompt-engineering-principles.md` in this repository already states the rule this violates — spend the budget on gotchas, not on what the filesystem shows.

The genuinely non-derivable sections — Design System Deviations, Project-Specific Patterns, Important Notes — are the ones worth keeping and the ones a reader has to scroll past everything else to reach.

---

## 8. Residue from the previous audit

`skills/workflow/to-plan/references/` is an empty directory. F5 in `docs/audit/sdlc-token-cache-audit.md` removed the orphaned reference files; the directory itself survived. Zero token cost, but it makes the skill tree misreport which skills carry references.

---

## 9. Where each finding belongs

| Finding | Owner | Blocks | Depends on |
|---|---|---|---|
| D3 (state repair) | consumer repository | D1's live symptom | nothing |
| D1 (invariant + recovery) | this repository | — | nothing; D3 clears the symptom, D1 prevents recurrence |
| D2 (bound `phases`) | this repository | — | nothing |
| D4 (`CLAUDE.md` diet) | consumer repository | — | nothing |

All four are independently actionable. D3 is a state repair with no code change and the largest immediate effect; D1 and D2 are the durable fixes that stop the same state from costing the same again.

The most fragile assumption in this audit: that the whole-plan fallback is what actually happens on `ambiguous_active_plan`. If a session transcript shows the agent stopping instead of reading, D1's cost figure drops sharply — but D1 remains a real defect, because an unenforced invariant and an undefined recovery branch are wrong regardless of which way the agent improvises.
