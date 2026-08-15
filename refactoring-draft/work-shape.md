# Work Shape — Why A Small Task Costs So Much

**Added 2026-08-15**, after the owner reframed the root cause:

> "khi cái onedrive chạy 1 task nhỏ cũng sinh quá lâu, quá nhiều files quá phức tạp,
> file thì sinh ra nhiều mục thừa thải. harness chưa chặt chẽ, chưa tận dụng được.
> Làm sao kết hợp được nhiều thứ hơn, docs, tooling, context managements, cli"

The draft spec (M1–M7) treats **durability** — knowledge dying in gitignored paths.
That diagnosis holds, but it is not what hurts day to day. What hurts is that
**ceremony does not scale down with task size.** This file adds that as **M8**, and
answers the second question: how docs, tooling, context management, and CLI combine
instead of stacking.

---

## Evidence 1 — The harness records ceremony over knowledge, 23 : 1

All 39 changesets in onedrive-cloud, 78 entries, ~2 months of real work:

| Entity | Entries | What it is |
|---|---:|---|
| `story` | **23** | ceremony |
| `managed_doc` | 15 | CLI bookkeeping |
| `meta` | 12 | CLI bookkeeping |
| `trace` | 8 | ceremony |
| `handoff` | 7 | ceremony |
| `run` | 5 | ceremony |
| `check` | 5 | ceremony |
| `intake` | 2 | ceremony |
| `decision` | **1** | **the only thing that cannot be regenerated** |

Story outnumbers decision **23 to 1**. Of 78 entries, 27 are CLI self-bookkeeping and
50 are process records. **One** entry captures durable judgment — and it is in a
gitignored path (audit D4).

The harness is heavily *used*. It is not heavily *leveraged*. That is precisely
"chưa tận dụng được."

## Evidence 2 — One small task, 142 lines of prose

`fix: virtualizer scroll parent` — a scroll-container fix. Its footprint:

```
.kit/runs/work/20260718-1359-virtualizer-load-more-fix.md      103 lines
.kit/reports/check/20260718-1359-ui-virtualizer-fix.md          39 lines
+ story, trace, run, check changeset entries
```

The run doc's skeleton, emitted in full for a one-file fix:

```
# COOK RUN
## Preflight   ## Approach   ## Wave / Task Log
### Wave 1  → #### T1  #### T2  #### T3
## Summary
## Recommended commit split
### Commit A — chore(security): contain credential files …
### Commit B — perf(ui): virtualize folder views …
### Commit C — fix(auth): improve token refresh error logging
### Commit D — tests/misc only if still dirty        ← conditional, written unconditionally
## Next Recommended Action
```

And the check report's:

```
# CHECK REPORT
## Gate Evidence   ## Artifact Alignment
## Findings → ### Critical   ### Major   ### Minor / Suggestions   ← emitted empty or not
## Next Action
```

This is "nhiều mục thừa thải" made literal: `### Critical` is written whether or not a
critical finding exists, and `### Commit D` documents its own conditionality. A
one-file fix acquired a wave log, a four-commit split plan, and a three-severity
finding taxonomy.

## Evidence 3 — Upstream already forbids exactly this

`repository-harness/docs/plans/README.md`:

> **Do not split one task into story, design, trace, and validation records without an
> independent audience.**

zharness splits every task into **story + trace + run + check + handoff** (+ `intake`,
`intervention` when triggered) — seven record kinds, no independent audience. The
upstream principle that governs it:

> **Process follows work shape.** Bounded work stays bounded; coordinated or
> recoverable work gets one durable plan.

And their `AGENTS.md` makes it operational in three lines:

> - Answers, explanations, reviews, diagnoses, plans, and status reports are read-only.
> - For a bounded change, inspect affected behavior and proof, implement, and validate.
>   **No control-plane operation is required.**
> - Use one `docs/plans/active/` file when work spans sessions, coordinates
>   contributors, has dependencies, needs recovery, or cannot safely resume from its diff.

The escalation trigger is stated as a testable predicate, not a vibe:
**"cannot safely resume from its diff."**

## Counter-evidence — the plans themselves are fine

`.kit/plans/2026-08-15-walter-theme/plan.md` is 186 lines, 7 headings, and carries a
verified `bunx shadcn add` dry-run table plus a "Fragile assumption" section. That is a
*good* document. The weight is not in planning. It is in the record-keeping that
surrounds every task regardless of size.

So M8 must not touch plan quality. It gates **which records get created at all**.

---

## M8 — Work-shape gating

`zharness` decides ceremony from work shape, instead of applying one shape to
everything.

| Shape | Trigger | Records created |
|---|---|---|
| **read-only** | answer, review, diagnose, status | none |
| **bounded** | single session, resumable from its diff | trace only |
| **durable** | spans sessions, has dependencies, needs recovery, or cannot resume from its diff | one plan + trace + decisions |

Three rules:

1. **`preflight` returns the shape**, and the playbook path follows from it. The
   escalation predicate is *"can this resume from its diff?"* — mechanical enough to
   ask, and it is already the upstream test.
2. **A shape can escalate mid-task, never de-escalate.** Bounded work that turns out to
   span sessions promotes to durable and writes its plan then — cheap, because nothing
   was written before.
3. **Templates emit no empty sections.** `### Critical` appears only with a critical
   finding; a conditional commit is written only when its condition holds. This is a
   template-rendering change, not a policy change.

Retire or fold what the split produced: `story`, `run`, and `check` records collapse
into `trace` for bounded work. `story`'s 23 entries bought nothing a trace does not.

**Verification:** re-run the virtualizer fix's shape through the gate and assert it
produces **0 markdown files** — down from 142 lines. Assert the walter-theme migration
still classifies durable and keeps its plan intact.

---

## The integration question

> "Làm sao kết hợp được nhiều thứ hơn, docs, tooling, context managements, cli"

Today these are four stacked systems. They combine when each owns one link of a single
authority chain, and the CLI is the thing that walks it:

```
     AUTHORITY                WHERE IT LIVES              WHO ENFORCES
  ──────────────────────────────────────────────────────────────────────
  owner decision      →   docs/decisions/NNNN-*.md   →   human (ADR gate)
        ↓
  accepted rule       →   docs/product/, docs/patterns/  →  review
        ↓
  mechanical check    →   repo-native test/lint      →   CI          ← tooling
        ↓
  observed fact       →   docs/map/ (generated)      →   zharness wiki
        ↓
  the work            →   trace / plan by shape      →   zharness    ← work-shape gate
```

Each layer's authority comes from the one above it, and **nothing invents authority
from the layer below**. That is ADR 0028's rule, and it is the piece the current design
has no answer for:

> Code organization, repeated patterns, tests, tool defaults, configurable examples, and
> undocumented preferences show current behavior or convention; **they do not authorize a
> new invariant.**

What each of the four surfaces becomes:

- **CLI** — decides *shape* and serves *context*. Today it only records what happened.
  These are the two jobs that make it leveraged rather than merely used.
- **Context management** — `preflight` returns the smallest sufficient slice for the
  shape. R1–R9 already built the machinery (`query plan --section`, batched `trace add`);
  the shape gate is what finally makes it pay, because bounded work stops loading a plan
  it does not have.
- **Docs** — the only durable landing zone. Generated (`map/`) and written
  (`product/`, `decisions/`) split by decay rate, per D5/D9.
- **Tooling** — where an accepted rule stops being prose. The enforcement ladder
  (local → hook → CI → branch protection, *none proves another*) keeps `zharness audit`
  from overclaiming.

**M8 is the entry point to all of it.** Durability (M1–M7) makes knowledge survive;
work-shape gating makes the harness cheap enough to actually run on a small task. The
owner's complaint is the second one, so it should ship first.

---

## Consequence for the spec

`spec.md` currently phases P1 durability → P2 decisions → P3 docs → P4 map. M8 is not
in it. Recommended: **M8 becomes P0**, ahead of everything — it is the smallest change,
it is felt immediately, and it needs none of P1–P4 to land.

See [`open-questions.md`](open-questions.md) Q3.
