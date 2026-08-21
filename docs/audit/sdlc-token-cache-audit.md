# SDLC Token + Prompt-Cache Audit — and the Optimization Spec

**Date:** 2026-08-11
**Scope:** the `watzup → brainstorm → to-plan → work → check → git → handoff` chain in `skills/workflow`, the playbooks in `docs/playbooks/`, and the `zharness` CLI read/write surface.
**Method:** `zharness` built from source (`cli/`, Go), `zharness init` in a throwaway git repository, then a real `intake → story → run create → trace add` lifecycle driven against a filled 3-phase plan. Every CLI output was measured in bytes. Turn counts come from the playbooks' own mandated steps.
**Supersedes nothing.** This extends `workflow-harness-ceremony-audit.md` with the dimension it did not model: **prompt caching**.

### Caveats

Byte counts are exact (measured). Token counts use ~3.5 chars/token and carry ~15% error; every conclusion rests on ratios measured with the same divisor, so the error cancels. Dollar figures use Claude API list prices as of 2026-08-11 — Opus 5 at $5/$25 per MTok, Sonnet 5 at its introductory $2/$10 (reverts to $3/$15 after 2026-08-31), Haiku 4.5 at $1/$5 — with cache reads at 0.10× input and 5-minute cache writes at 1.25× input.

---

## 1. Headline

The previous audit concluded the CLI is the cheapest part of the system. That still holds. This audit adds two findings it could not see:

1. **The chain's per-stage `model:` assignments are the single largest cost lever, because prompt caches are model-scoped.** Every stage boundary in the pipeline is a model switch, and every model switch is a full cache rebuild. `work` (sonnet) invokes `check` (opus) once per phase, so the most expensive stage in the chain pays a cold cache every single phase.

2. **The `query plan --section phase` optimization shipped in the ceremony audit's P3 proposal does not work.** It silently degrades to a whole-file read on every plan the documented tooling produces, and the degraded response is *larger* than the file it was meant to replace.

| # | Finding | Measured cost | Fix risk |
|---|---|---|---|
| **F1** | `check` runs on opus and is invoked per phase; cache is model-scoped, so it never warms | **$0.275/phase, 63% of `check`'s cost** | Medium — splits one skill into two modes |
| **F2** | `query plan --section phase` always returns `degraded: true` on scaffold-shaped plans | 4.3× larger read; 5% larger than just reading the file | **None** — one regex |
| **F3** | `trace add` is mandated per task: a 36-byte return for a full context round trip | 41.6% of `work`'s cost is ceremony, not engineering | Low |
| **F4** | `work` generates 3–7× the 20-block cache lookback window per phase | breakpoint refreshes, not a hard failure | Low |
| **F5** | `skills/workflow/{work,check,brainstorm,handoff}/references/` are orphaned | ~27 KB dead weight, 0 tokens | None |

---

## 2. Measured cost, cache-aware

One phase of a normal-lane initiative: 5 tasks across 2 waves, 6 changed files at review.

| Stage | Model | Cache-read tok | Cache-write tok | Output tok | Cost (cached) | Cost (uncached) | Cache saves |
|---|---|---:|---:|---:|---:|---:|---:|
| `watzup` | haiku | 25,045 | 7,356 | 549 | $0.0144 | $0.0351 | 59% |
| `brainstorm` | opus | 51,024 | 11,150 | 1,786 | $0.1398 | $0.3555 | 61% |
| `to-plan` | opus | 38,499 | 7,808 | 1,423 | $0.1036 | $0.2671 | 61% |
| `work` | sonnet | 580,170 | 23,801 | 4,014 | $0.2157 | $1.2481 | 83% |
| **`check`** | **opus** | 380,820 | 26,580 | 3,094 | **$0.4339** | $2.1144 | 79% |
| `git` | sonnet | 30,185 | 11,296 | 277 | $0.0370 | $0.0857 | 57% |
| `handoff` | sonnet | 72,855 | 9,596 | 960 | $0.0482 | $0.1745 | 72% |
| **Total** | | | | | **$0.9927** | $4.2804 | 77% |

Two things to read off this table:

**Caching already absorbs 77% of the raw cost.** Any optimization proposal that ignores caching will mis-rank its own priorities — which is exactly what a raw-token count does here. By raw tokens `work` is the biggest stage (48% of the chain); by actual cost it is second at 22%, because its long stable prefix caches unusually well (83%).

**`check` is the most expensive stage at 44% of chain cost**, despite having 23 turns to `work`'s 37. Two multipliers stack: it runs on opus (2.5× sonnet's input price at current intro pricing), and its cache never warms.

---

## 3. F1 — model-scoped caches make the pipeline pay a cold prefix at every stage boundary

Prompt caches are keyed per model. Switching the model invalidates the tools, system, and messages tiers together — there is no partial survival and no escape hatch (unlike system-prompt edits, which have one).

The chain's declared models, from each `SKILL.md` frontmatter:

```
watzup=haiku → brainstorm=opus → to-plan=opus → work=sonnet → check=opus → git=sonnet → handoff=sonnet
```

Six model switches per initiative, and `work.md` step 11 routes to `check full` **once per phase**, so the sonnet→opus→sonnet round trip repeats for every phase in the plan.

Measured, for the `check` stage:

| Configuration | Cost/phase |
|---|---:|
| `check full` on opus, cold cache (current) | $0.4339 |
| `check full` on sonnet, cold cache | $0.1736 |
| `check full` on sonnet, warm from `work` | **$0.1588** |

**$0.275 per phase, 63% of the stage's cost**, is paid for the model switch alone. On an 8-phase initiative that is $2.20 — more than twice the entire rest of the chain's per-phase cost.

The fix is *not* "put everything on sonnet." `check.md` defines four modes, and they do not have the same intelligence requirement:

- `gate` — run the automated checks, audit lifecycle alignment, record the verdict. Mechanical. Explicitly **does not** perform the manual review (`check.md` §5).
- `full` — gate plus the complete Security/Performance/Architecture/Code Quality review. Genuinely intelligence-sensitive.

`work.md` step 11 currently routes to `check full` on every phase, so every phase pays opus for a stage that is mostly gate work. Splitting the model assignment along the mode boundary that the playbook already defines is the highest-ROI change in this audit.

---

## 4. F2 — the P3 plan-slicing optimization is dead on arrival

`work.md` step 1 instructs the agent to call `query plan --section phase` instead of reading the whole plan, citing P3 in the ceremony audit as the saving. Measured against a filled 3-phase plan:

| Read path | Bytes returned |
|---|---:|
| `query plan --section phase` — plan from `zharness scaffold plan` | **5,863** (`degraded: true`) |
| Reading the plan file directly | 5,580 |
| `query plan --section phase` — plan in the format the regex expects | **1,290** |

The optimization is worth 4.3× when it works. It never works. `cli/internal/application/plan_query.go:34`:

```go
var planPhaseHeading = regexp.MustCompile("(?m)^### phase_slug: `([^`\r\n]+)`[ \t]*\r?$")
```

The regex requires a `### phase_slug: \`{slug}\`` markdown heading. But `cli/docs/embedded/templates/plan.md:36-37` — the template `zharness scaffold plan` emits — uses a YAML list instead:

```
- planning_status: not-planned
- phases: none
```

and `to-plan.md` step 3 never instructs the agent to write `###` headings. Grepping the repository, the heading form exists in exactly four places: the regex, its own tests, `cli/docs/CONTRACT.md`, and `docs/plans/completed/harness-memory-ceremony-convergence.md` — the hand-written plan that *commissioned* the feature. The regex was written to match its one inspiring sample.

Every plan produced through the documented path therefore falls into the `degraded` branch, which returns the whole file JSON-escaped — 5% *larger* than the file it replaced. The optimization runs backwards.

Under caching the per-phase cost is small ($0.0124, ~6% of `work`), because the wasted bytes land in a cached prefix. It is still a pure loss, it is one regex, and the same bug will keep the intended 4.3× saving out of reach for as long as it stands.

`--section current-state` is unaffected — `## Current State and Next Action` is a `## ` heading that the template does emit — so `handoff` and `watzup` are clean.

---

## 5. F3 — the per-task ceremony loop

`work.md` steps 5–7 mandate, per task: read target → read neighbour → edit → verify → `trace add`. Splitting those turns into ceremony (bookkeeping) and productive (engineering):

```
ceremony    16 turns   244K raw tok   41.6% of the work stage
productive  21 turns   345K raw tok
```

The dominant single item is `trace add`, mandated "immediately" after each task (step 7). Its CLI response is **36 bytes**. The round trip to obtain those 36 bytes replays the entire conversation context.

Batching to one `trace add` per wave instead of per task:

| Cadence | Turns | Raw tokens |
|---|---:|---:|
| per-task (current) | 37 | 590,583 |
| per-wave (batched) | 32 | 496,406 (**−16%**) |

Per-task granularity was a deliberate choice (G1 in the ceremony audit) and buys resumability at task rather than wave granularity. The proposal below keeps that guarantee without keeping the round trips.

---

## 6. F4 — the 20-block cache lookback window

Each cache breakpoint walks backward at most 20 content blocks looking for a prior entry. A tool round trip contributes roughly two blocks (assistant `tool_use` + `tool_result`).

| Phase size | Turns | ≈ content blocks | Lookback windows spanned |
|---|---:|---:|---:|
| 3 tasks | 27 | 54 | 2 |
| 5 tasks | 37 | 74 | 3 |
| 8 tasks | 52 | 104 | 5 |
| 12 tasks | 72 | 144 | 7 |

Claude Code manages breakpoint placement itself, so this is not a hard failure. It is corroborating evidence for F3: the per-task loop generates 3–7× the window per phase, and the raw per-task cost climbs from 120K tokens (3-task phase) to 151K (12-task phase) — **+26% per unit of work as the phase grows.** Smaller phases are cheaper per task, which is an argument for `to-plan.md` step 3's existing split-early bias, not against it.

---

## 7. F5 — orphaned reference directories

`skills/workflow/{work,check,brainstorm,handoff}/references/` (~27 KB) are referenced by no `SKILL.md` and no playbook — the operating logic moved into `docs/playbooks/` when the harness landed. They cost zero tokens (nothing loads them) but they are stale documentation that will drift. `git/` and `interview/` still use theirs and must be kept.

---

## 8. Optimization spec

Four phases, ordered by ROI-per-unit-risk. Each is independently shippable.

### P1 — Fix the plan-section regex

**Change:** make `planPhaseHeading` in `cli/internal/application/plan_query.go` accept the YAML list form the scaffold template actually emits (`^\s*- phase_slug: {slug}$`, terminating at the next sibling `- phase_slug:` or the next `## ` heading), in addition to the existing `###` heading form. Keep both: `harness-memory-ceremony-convergence.md` uses headings and must not regress.

**Verification:** a new `plan_query_test.go` case built from the literal output of `scaffold plan --path` with phases filled in list form, asserting `degraded == false` and that the returned content excludes sibling phases. Plus `cd cli && go test ./...`.

**Value:** restores the intended 4.3× on `work`'s hot read path. **Risk:** none — pure widening, existing tests pin the old form.

### P2 — Split `check` by mode, not by stage

**Change:** `check gate` runs on sonnet; `check full` stays on opus. Since `model:` is per-skill frontmatter and not per-mode, this means either (a) splitting `check` into two skills, or (b) `work.md` step 11 routing to `check gate` per phase and reserving `check full` for the final phase before `handoff`. Option (b) is the smaller change and needs no new skill.

**Verification:** `check.md` §5 already states that `gate` does not perform the complete manual review, so (b) does not weaken any guarantee the playbook makes — but the final-phase `check full` must be made mandatory in `handoff.md` step 6's preconditions, or the complete review can be skipped entirely on a plan whose last phase closes cleanly.

**Value:** $0.275/phase (63% of the stage). **Risk:** medium — this is the one change that alters what review runs when. It must not become "the deep review never happens"; the `handoff` precondition is the load-bearing half of the change.

### P3 — Batch `trace add` to wave granularity, preserve task resumability

**Change:** allow `zharness trace add` to accept multiple task entries in one invocation (the same shape `decision add --decisions '[...]'` already uses). `work.md` step 7 becomes: record task status locally as each task completes, then flush the wave's entries in one call at step 9.

**Verification:** `query traces --phase {slug}` must return the same per-task rows as today — the resumability guarantee is about what the DB holds, not how many calls wrote it. An existing `repository_replay_test.go`-style assertion over a batched write proves it.

**Value:** −16% on `work`. **Risk:** low, but note the one real regression: a session that dies mid-wave loses the un-flushed task entries. Mitigate by flushing on `BLOCKED`/`NEEDS_CONTEXT` immediately rather than deferring, so only clean-progress entries are ever buffered.

### P4 — Delete orphaned references

**Change:** remove `skills/workflow/{work,check,brainstorm,handoff}/references/`. Keep `git/` and `interview/`.

**Verification:** `bash scripts/verify-doc-links.sh` must stay green — if anything links into those directories, that link is the actual finding.

**Value:** 0 tokens; removes drift surface. **Risk:** none.

### Explicitly not proposed

- **Putting the whole chain on one model.** It would maximize cache reuse, but `brainstorm` and `to-plan` are where a wrong decision is cheapest to make and most expensive to live with. The saving does not justify it.
- **Dropping `check record`'s proof re-execution.** It re-runs every cited proof command before recording an APPROVED verdict (`check.md` step 9). That doubles gate wall-clock — 1.7s in this repo — and it is the mechanism that closed a real verification bypass in `c53fb76`. Keep it.
- **Trimming the playbooks.** At ~9 KB each they are the largest fixed load, but they sit in the cached prefix and are read once per stage. Cutting them trades a real capability for ~$0.002/phase.

### Expected result

| | Current | After P1–P3 |
|---|---:|---:|
| `work` | $0.2157 | ~$0.180 |
| `check` (per phase) | $0.4339 | ~$0.159 |
| Chain total, one phase | $0.9927 | ~$0.68 (**−31%**) |

The saving concentrates where the audit says it should: 90% of it comes from P2, the model-boundary fix, and P2 is the only change in the set that needs a judgment call about review depth rather than a mechanical edit.
