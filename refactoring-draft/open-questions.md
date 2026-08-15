# Open Questions — Resolved

Items that research after the interview put back in play, each contradicting something
already written in [`spec.md`](spec.md) or [`decisions.md`](decisions.md).

**All four are resolved as of 2026-08-15** and folded into `spec.md`. This file is kept
as the reasoning record — the *why* behind each correction, so none of it gets
re-litigated. Resolutions are logged as D13–D15 in [`decisions.md`](decisions.md).

| | Question | Resolution |
|---|---|---|
| Q1 | ≤1,000-token resident gate unachievable | **Split into two benchmarked gates** — D13 |
| Q2 | ADR promotion gate would flood `docs/decisions/` | **Filter by trigger criteria** — D14 |
| Q3 | M8 work-shape gating missing from the spec | **M8 becomes P0** — D15 |
| Q4 | Two upstream mechanisms with no equivalent | **Folded into P1 and P3**, not deferred |

---

## Q1 — The ≤1,000-token resident gate is probably unachievable

> **Resolved: split into two gates.** See D13.

**Locked in D12 / spec Success Condition:** `AGENTS.md` managed block +
`.harness/WORKFLOW.md` + `docs/README.md` ≤ **1,000 tokens** combined.

**What research found:** `repository-harness`, the model this spec cites as
best-in-class, measures **~2,199 tokens** across the same three surfaces:

| File | Bytes | ~Tokens |
|---|---:|---:|
| `AGENTS.md` | 1,613 | 403 |
| `docs/README.md` | 1,622 | 405 |
| `docs/WORKFLOW.md` | 5,564 | 1,391 |

The gate is **2.2x tighter than the reference**. The spec's own Pause clause
anticipated this — *"the resident budget forces the router below usefulness"* — but a
gate that is expected to trip is not a gate.

**Resolution — two gates, both benchmarked against real numbers:**

| Gate | Budget | Rationale |
|---|---:|---|
| Entrypoint pair (`AGENTS.md` block + `docs/README.md`) | ≤ 900 | upstream: 808 |
| Full resident path (+ `.harness/WORKFLOW.md`) | ≤ 2,400 | upstream: 2,199 |

Still ~2.7x lighter than onedrive-cloud's measured 6.6k. The split is more diagnostic
than one number: it separates "the map got bloated" from "the procedure got bloated."

Also taken: upstream keeps `docs/HARNESS.md` (principles, ~475 tok)
**out of the resident path**, loaded on demand. Splitting principles from procedure is
what holds `WORKFLOW.md` at 1,391.

---

## Q2 — ADR promotion needs a filter, not a completion gate

> **Resolved: filter by trigger criteria; the gate asserts promotion was considered.**
> See D14.

**Written in spec P2:** *"`handoff` playbook gate: promotion required before a plan
moves to `completed/`."*

**Problem:** that promotes everything. Given the measured ratio — 78 changeset entries,
1 decision — most recorded choices are task-local and would become ADR noise, which is
the failure mode `docs/decisions/` exists to prevent.

**Upstream states both halves explicitly.** Trigger set:

> a lasting product or architecture choice changes; public compatibility or data
> ownership changes; security or recovery policy changes; validation is materially
> added, removed, or weakened; or the source-of-truth hierarchy changes

Exclusion, one line: **"Task-local choices stay in the active plan."**

**Resolution:** `zharness decision promote` presents candidates against those five
triggers and the human selects. The handoff gate asserts promotion was **considered**
(non-empty candidate review), not that everything was promoted. `decision add --durable`
lets the call site mark intent when it is already known.

Two smaller corrections in the same area, both taken:

- The ADR template gains **`## Status`** (Proposed | Accepted | Superseded | Rejected)
  and **`## Follow-Up`**. `Status` is what makes supersession work without deletion —
  the draft template had neither.
- `zharness init` scaffolds an **empty** decision index plus the trigger criteria.
  Never mono-harness's own ADRs.

---

## Q3 — M8 (work-shape gating) should lead the program

> **Resolved: M8 becomes P0, ahead of P1–P4.** See D15.

**Not in the spec at all.** Raised by the owner after it was drafted:

> "1 task nhỏ cũng sinh quá lâu, quá nhiều files quá phức tạp … harness chưa chặt chẽ,
> chưa tận dụng được"

Evidence and design in [`work-shape.md`](work-shape.md). Short version: a scroll-parent
one-file fix produced 142 lines of markdown across a run doc and a check report, with
empty severity headings and a four-commit split plan. Ceremony does not scale down.

**Resolution: M8 becomes P0**, ahead of P1–P4. It is the smallest change, it needs
nothing else to land, and it is the pain the owner actually reported. Durability
(M1–M7) is the diagnosis; ceremony weight is the symptom being felt.

If P0 lands and the harness becomes cheap to run on small work, P1–P4 get exercised far
more than they would otherwise.

---

## Q4 — Two mechanisms found upstream that the spec had no equivalent for

> **Resolved: both folded into existing phases rather than deferred or given a phase of
> their own.** Neither was blocking; both are cheap enough that carrying them as a
> backlog would have cost more than landing them.

**Deliberate absence as a retrieval feature. → folded into P3.** Upstream's
`docs/README.md` and `docs/decisions/README.md` each carry a `## History` section
explaining what was *removed* and why — superseded material is pruned from the tree so
retrieval returns current authority, with git history and immutable tags as provenance.
The spec treated plans as disposable but never made pruning a named, explained act.

**Enforced invariants as a third knowledge kind. → folded into P1 as `audit` hardening,
plus spec decision 9 and the validation loop.** `docs/patterns/encoding-invariants.md`
plus ADR 0028 define knowledge compiled into repository-native validation — it does not
rot because CI fails. The draft's synthesis table had only two kinds (derived map,
durable memory). The transferable disciplines:

- **Authority gate** — conventions, tests, and tool defaults show behavior; they do not
  authorize a rule. Stop and ask when two boundaries fit the words.
  *(→ spec Key Decision 9, and the authority chain in `work-shape.md`.)*
- **Both-directions proof** — *"a passing repository with no exercised violation does
  not prove that the guard can detect recurrence."*
  *(→ spec Validation Loop: every new gate ships with a fixture that trips it.)*
- **Enforcement ladder** — local → optional hook → checked-in CI → branch protection,
  and **none proves another**. Directly applicable to `zharness audit`, whose findings
  today do not distinguish "a check exists" from "a check ran and passed."
  *(→ P1 `audit` hardening.)*
- **Diagnostic standard** — violating item, broken rule, **authority pointer**, next
  action. Never `validation failed`.
  *(→ P1 `audit` hardening.)*
