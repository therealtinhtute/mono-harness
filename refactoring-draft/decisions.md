# Interview Decision Log

Session 2026-08-15. Decisions locked with the owner, in order asked.

---

## D1 — Harness generation: keep zharness, delete Gen-1

**Chosen:** Keep `zharness 0.9.1`. Delete the legacy generation (`workflow-state.yml`,
`.kit/planning/`, `.kit/runs/`, `.kit/reports/`), rescuing real product authority first.

**Rejected:** Migrating to harness-experimental's repository-as-record model — it means
abandoning the owner's own CLI and losing resume/drift detection.

**Rejected:** Keeping both — that is the current state, and it produced two state
pointers disagreeing by a month.

---

## D2 — `.kit/` is absorbed into `.harness/`, kept slim

**Owner instruction:** "remove .kit or merge it into harness keep it slim."

One harness directory. Tracked workflow docs at the top, machine state beneath:

```
.harness/
├── WORKFLOW.md          ✓ tracked
├── playbooks/           ✓ tracked (CLI-managed)
├── plans/
│   ├── active/          ✓ tracked
│   └── completed/       ✓ tracked
└── state/               ✗ gitignored
    ├── changesets/
    ├── log/
    └── conflicts/
```

Also removes the root-level `harness.db` clutter.

> **Refined by D23** — `harness.db` sits at `.harness/harness.db`, not inside
> `state/`. The tree above is otherwise unchanged.

---

## D3 — Ship as a zharness release with migration

**Chosen:** Change `cli/internal/interfaces/paths.go`, add a `.kit`→`.harness` entry to
`layout_migration.go`. Consumers migrate on next init.

**Why it is cheap:** `MigrateLayout` already exists and already performed
`.kit/harness.db` → `harness.db` with snapshot, parity check, and rollback.
`legacyDBPath` sets the precedent. `SyncManagedDocs(db, changesetDir, docsRoot, ...)`
takes its root as a **parameter** — only the `interfaces` constants pin the path.

**Rejected:** A onedrive-cloud-local hand-rolled fix — the split would silently break on
the next `zharness init`.

---

## D4 — Tracked/ignored line: `.harness/state/` is the only ignored part

Workflow docs are shared and reviewable; machine state is per-machine.
One gitignore line. Fixes D3 in the audit — product authority no longer sits in a
wholly gitignored directory.

**Rejected:** Ignoring playbooks too — cloning without zharness would leave no readable
process.

**Rejected:** Committing changesets — noisy diffs on every stage transition and
near-certain merge conflicts on parallel branches.

---

## D5 — Delete `KNOWNS.md` and everything tied to it

**Owner instruction:** "xoá hết KNOWNS & những thứ liên quan của nó, đó là code cũ bị
outdate rồi."

Verified orphaned: nothing references it, its mandated marker comments exist nowhere,
and `.gitignore:47` carries a dead `.knowns/` entry.

---

## D6 — `docs/` is durable memory, not prose

**Owner correction:** "docs cũng không phải là handwritten mà long-term memory, durable
memory docs của projects là xương sống, là map, là knowledge base của project."

This reframed the design axis. The question is not "hand-written vs generated" — it is
what makes a knowledge base retrievable and true. `docs/` is the project's backbone, map,
and knowledge base.

---

## D7 — Reference codesight's structure and method, not codesight itself

**Owner instruction:** "tham khảo codesight cấu trúc, cách tạo project wiki data không
phải lấy hết codesight vào."

Take the tiered router/article shape, freshness stamps, epistemic boundary, negative
space, source paths on every fact, and blast-radius artifact. Build it into zharness.
Do not add codesight as a dependency.

---

## D8 — `zharness wiki` is deterministic

**Chosen:** File walk plus regex. No AST, no LLM, no tokens, identical output every run.
Stack detection stays adapter-shaped, Next.js App Router first.

**Verified tractable** against a real repo — routes, env vars, coverage, import graph,
and subsystems are all derivable without a TypeScript parser.

**Rejected:** Agent-written articles during a workflow stage — costs tokens every run,
non-deterministic, can hallucinate. Exactly what codesight avoids by refusing an LLM.

**Rejected:** Deterministic skeleton plus agent semantic layer — the halves drift and
regeneration would overwrite agent prose.

---

## D9 — Generated map lives in `docs/map/`

The map **is** project knowledge, so it belongs in `docs/` under the owner's own rule.
Namespaced as `docs/map/` with a DO-NOT-EDIT banner and freshness stamp on every file.
`docs/README.md` routes to both halves, so `AGENTS.md` points at exactly one place.

```
docs/                 ← ALL project knowledge
├── README.md         ← router to both halves
├── map/              GENERATED · do not edit
├── product/          WRITTEN · intent
└── decisions/        WRITTEN · ADRs
```

**Rejected:** A separate top-level `.map/` — adds a third location and forces `AGENTS.md`
to point at two places for project knowledge.

**Rejected:** Putting it in `.harness/` — contradicts the rule; the map is project
knowledge, not workflow tooling.

---

## D10 — ADRs promoted at plan completion

```
work → check → handoff
                │
                ├─ promote decisions
                │     ↓
                │  docs/decisions/0001-*.md
                │
                └─ .harness/plans/active/x.md
                        ↓
                   plans/completed/x.md
```

The knowledge base grows; plans stay disposable. Matches harness-experimental's rule that
completed plans may be removed once decisions, code, tests, and history preserve the
lasting result.

**Rejected:** Keeping decisions inside plan files — buried in 400-line documents mixed
with task checklists, so the same question gets re-litigated.

**Rejected:** Writing ADRs at decision time rather than harvesting at handoff — adds
mid-flow friction and produces ADRs for choices that later get reversed.

---

## D11 — Scope: all seven improvements, phased

M1 layout · M2 wiki · M3 scaffold · M4 ADRs · M5 drift · M6 legacy detection · M7 thinning.

None overlap the already-shipped R1–R9 token optimization
(`docs/plans/completed/sdlc-token-optimization.md`, 2026-08-11).

---

## D12 — Success measured on durability and resident cost

1. Zero durable knowledge in gitignored paths, proven by a fresh-clone test.
2. Consumer resident context ≤ 1,000 tokens.

**Rejected:** Reusing the R1–R9 per-stage cost model as the primary metric — this program
is about durability and navigation, so that number may barely move even on full success.

---

---

# Post-Research Corrections

Decided 2026-08-15, after research surfaced four contradictions with the drafted spec.
Reasoning in [`open-questions.md`](open-questions.md).

---

## D13 — Resident cost is two gates, not one — supersedes D12.2

**Chosen:** Entrypoint pair (`AGENTS.md` block + `docs/README.md`) ≤ **900** tokens;
full resident path (+ `.harness/WORKFLOW.md`) ≤ **2,400**. Both benchmarked against
`repository-harness`'s measured 808 / 2,199.

**Why:** D12's ≤1,000 combined is 2.2x tighter than the reference model the spec cites as
best-in-class. A gate that is expected to trip is not a gate. Two numbers are also more
diagnostic than one — they separate "the map got bloated" from "the procedure got
bloated." Still ~2.7x lighter than onedrive-cloud's measured 6.6k.

**Also taken:** principles stay **out** of the resident path, loaded on demand. That
separation is what holds upstream's procedure doc at 1,391 tokens.

**Rejected:** Holding ≤1,000 as a forcing function — it would ship a router too thin to
navigate with, and the spec's own Pause clause already predicted the trip.

**Rejected:** A single ≤2,400 gate — simpler to check, but loses the diagnostic split.

---

## D14 — ADR promotion filters against triggers — refines D10

**Chosen:** `zharness decision promote` presents candidates against five durability
triggers and the human selects. The `handoff` gate asserts promotion was **considered**
(non-empty candidate review), not that everything was promoted.

Triggers: lasting product or architecture choice changes; public compatibility or data
ownership changes; security or recovery policy changes; validation materially added,
removed, or weakened; source-of-truth hierarchy changes.
Exclusion: **task-local choices stay in the active plan.**

**Why:** 78 changeset entries yielded 1 decision. A promote-everything gate would fill
`docs/decisions/` with task-local noise — the exact failure mode ADRs exist to prevent.

**Also taken:** `decision add --durable` marks intent at the call site. The ADR template
gains `## Status` (Proposed | Accepted | Superseded | Rejected) and `## Follow-Up` —
`Status` is what makes supersession work without deletion. `zharness init` scaffolds an
**empty** decision index plus the trigger criteria, never mono-harness's own ADRs.

**Rejected:** The completion gate as drafted — guarantees nothing is lost, at the cost of
an ADR directory nobody can retrieve from.

**Rejected:** Fully manual promotion with no handoff gate — lowest friction, but durable
decisions would simply never get harvested. The measured ratio is the evidence.

---

## D15 — M8 work-shape gating becomes P0 — reorders D11

**Owner reframe:** "khi cái onedrive chạy 1 task nhỏ cũng sinh quá lâu, quá nhiều files
quá phức tạp … harness chưa chặt chẽ, chưa tận dụng được."

**Chosen:** M8 ships **first**, ahead of P1–P4. `preflight` classifies work as
read-only / bounded / durable and the playbook path follows; shapes escalate but never
de-escalate; templates emit no empty sections.

**Why:** it is the smallest change, it needs none of P1–P4 to land, and it is the pain
actually reported. Durability (M1–M7) is the diagnosis; ceremony weight is the symptom
being felt. Once the harness is cheap to run on small work, P1–P4 get exercised far more
than they otherwise would.

**Boundary:** M8 gates *which records get created*, never plan quality. The 186-line
walter-theme plan is good work — the weight is in the record-keeping that surrounds every
task regardless of size.

**Rejected:** M8 as P5 after durability — preserves the drafted order but leaves the
reported pain unaddressed longest.

**Rejected:** M8 as a separate program — it shares `preflight`, the templates, and the
record model with P1–P4; splitting it would duplicate all three.

---

## D16 — Q4's two mechanisms fold into P1 and P3

**Chosen:** Pruning-as-a-named-act (`## History` sections recording what was removed and
why) folds into P3. The enforcement ladder and diagnostic standard fold into P1 as
`audit` hardening; the authority gate becomes spec Key Decision 9 and both-directions
proof becomes a validation-loop rule.

**Why:** neither is blocking, both are cheap, and carrying them as backlog would cost more
than landing them inside phases that already touch the same code.

**Rejected:** A fifth phase for enforced invariants — it is four disciplines, not a build.

---

---

# Acceptance Interview

Conducted 2026-08-15 in Vietnamese, after the spec was accepted, to verify owner and
agent share the same reading. Six ambiguities surfaced; all six resolved.

---

## D17 — Bounded work leaves no file, only a DB trace

**Chosen:** Zero markdown. The trace row still lands in `harness.db` so `recap` / `watzup`
can see the task happened in a later session; the agent still summarizes in chat. The real
record of the work is **the git diff and the commit message**.

**Rejected:** No trace at all — cheapest, but `watzup` would show a gap where work
actually happened.

**Rejected:** One minimal ~10-line file per bounded task — visible without the CLI, but
100 small tasks means 100 files, which is the problem restated at smaller scale.

---

## D18 — The agent declares the shape; the CLI validates and records it

**Chosen:** `preflight` returns the question and the criteria; the agent reads the request
and declares — `zharness preflight --shape bounded`. The CLI validates the declaration
against what it can see and records it. The owner is not asked, but can override in one
sentence.

**Why:** the CLI cannot read intent, and the escalation predicate ("can this resume from
its diff?") is about work that has not happened yet — there is no diff to measure at
declaration time.

**Rejected:** Asking the owner every task — most accurate, but adds a question-and-answer
round to exactly the small tasks P0 exists to make cheap.

**Rejected:** A CLI heuristic over `git status` / file counts — fully mechanical, but it
measures the past to classify the future.

**Rejected:** Default-bounded with silent auto-escalation — cheapest, but the plan always
arrives late.

**Resolves** the spec's P0 Pause clause: shape classification is not mechanical, and it
was never going to be. Declared-and-validated is the answer, not a better predicate.

---

## D19 — Bounded work skips the check gate as well as the check report

**Owner decision:** for a bounded task, drop **both** — no `check` report file, and no
`go test` / `verify-doc-links` run.

**Concern raised, and overruled by the owner.** The measured enforcement ladder:

| Gate | Local | CI |
|---|---|---|
| `go test ./...` | manual | ✅ `cli-ci.yml` — but only on `cli/**` paths |
| `verify-doc-links.sh` | manual | ❌ none |

So a bounded change under `cli/` is still caught by CI, while a bounded change under
`docs/`, `skills/`, or `rules/` has **nothing** checking it once the local gate is gone.

**The owner accepts this risk and declined all four compensations** — no CI addition, no
pre-commit hook, no path-filter change. Rationale: a broken doc link is not a disaster,
and it is cheap to fix once noticed.

**Consequence, to be carried out in P0:** `CLAUDE.md`'s gate rule currently reads *"Both
must pass before any commit"* with no exception. It must be scoped to durable work, or it
contradicts this decision on its face.

---

## D20 — onedrive-cloud stays read-only; its cleanup is separate work

**Chosen:** split into two pieces of work.

1. **This plan** — builds the CLI inside mono-harness. Strictly read-only toward
   onedrive-cloud, exactly as the Standing Constraint says.
2. **A separate, separately-approved piece of work** — runs the real migration and legacy
   cleanup on onedrive-cloud, once the CLI exists and its rollback is proven.

**Why:** D1 said *"delete the legacy generation, rescuing real product authority first"*,
but every path it named lives in onedrive-cloud, which the Standing Constraint declares
untouchable. The two could not both be true. D1's deletion is now a **requirement on the
CLI** (detect, report, rescue before removing), not an action this plan performs.

---

## D21 — mono-harness migrates its own layout inside P1

**Chosen:** P1 moves this repo too — `docs/playbooks/` → `.harness/playbooks/`,
root `harness.db` + `.kit/changesets/` → `.harness/state/`.

**Why:** dogfooding. `MigrateLayout` gets proven on a real repository before it reaches
onedrive-cloud, and if it has a bug the owner absorbs it rather than a repo doing real
work. Pairs with D20: mono-harness is the migration's test subject precisely so
onedrive-cloud does not have to be.

**Rejected:** Code-only P1 proven by fixtures — smaller and easier to review, but ships an
unexercised migration.

---

## D22 — `refactoring-draft/` stays where it is

**Chosen:** leave it at the repository root. It is brainstorm evidence and reasoning, not
an executable plan. `to-plan` reads it and generates a separate plan under
`docs/plans/active/`.

**Why:** the two artifacts have different lifecycles — the plan is disposable, this
evidence is the record of why the plan looks the way it does.

---

## D23 — `harness.db` sits at `.harness/harness.db` — refines D2

**Owner decision:** the database is the artifact everything reads through, so it belongs at
the root of `.harness/`, not buried inside `state/` alongside logs and staged conflicts.

**Layout:**

```
.harness/
├── WORKFLOW.md          ✓ tracked
├── playbooks/           ✓ tracked
├── plans/               ✓ tracked
├── harness.db           ✗ ignored  (+ -wal, -shm sidecars)
└── state/               ✗ ignored
    ├── changesets/
    ├── conflicts/
    └── log/
```

**Context:** today `dbPath = "harness.db"` sits at the repository root, so moving it into
`.harness/state/` would demote it two levels at once. `.harness/harness.db` is a one-level
move that keeps it visible.

**Two costs, both accepted, both mitigated rather than ignored:**

1. **The gitignore is two rules, not one.** `store.go:24` sets
   `PRAGMA journal_mode=WAL`, so SQLite writes `harness.db-wal` and `harness.db-shm`
   beside the database. Both are per-machine state:

   ```gitignore
   .harness/state/
   .harness/harness.db*
   ```

   The glob now sits in a directory where everything else is tracked, so `audit` gains a
   check: **fail if any `harness.db*` file is tracked by git.** A binary committed here
   would produce merge conflicts on every stage transition.

2. **Gate #1 can pass falsely.** The durability test's precondition was "delete
   `.harness/state/`" — one directory, impossible to half-do. With the database outside
   it, the precondition becomes two deletions, and forgetting the second leaves the test
   green because the database was never actually gone.

   **Mitigation, mandatory:** the fresh-clone test asserts `.harness/harness.db` and both
   sidecars **do not exist** before it calls `db rebuild`. The assertion is the test, not
   a convenience — without it the gate cannot fail for the reason it exists.

**Rejected:** `.harness/state/harness.db` — one gitignore line and a clean single-directory
precondition, but buries the primary artifact and inverts the dependency (the database is
*derived* from `state/changesets/`, so it would sit beneath its own source).

**Rejected:** `.harness/db/` as its own directory — keeps both rules directory-shaped with
no glob, but Gate #1 still needs two deletions and it adds a directory holding one file.

---

## D24 — P0 is a routing fix, not a shape system — re-scopes D15

**Found while brainstorming the process itself, 2026-08-15.** The mechanism M8 proposed to
build **already exists in the repository**:

- `docs/playbooks/work.md:15` — *"**Zero-write rule:** bounded/simple mode creates no
  lifecycle rows, plans, reports, changesets, or markdown artifacts. It does not edit an
  existing active plan. The Git diff plus captured executable/observable proof are its
  durable evidence."* That is D17, verbatim, already shipped.
- `work.md:17` — mechanical rejection criteria already stated: over five files, roughly 100
  changed lines, unclear scope, unfamiliar subsystem, multi-phase, or no verification path.
- `work.md:12` — the agent already declares `--mode {full|bounded}`. That is D18, already
  the status quo — which is why D18 alone fixes nothing.

**The actual defect is routing.** `cli/internal/application/next.go:56-93`:

```go
if mode == "simple" { return NextView{Mode: "simple"}, nil }
plans, _ := findActivePlans()
if len(plans) == 0 {
    if mode == "auto" { return NextView{Mode: "simple"}, nil }
}
return resolveFullMode(db, plans[0], explicitPhase)
```

A bare `zharness next` parses to `auto`, and `auto` picks the light path **only when no
plan is open**. Work size never enters the decision, so while any plan is open every task
inherits full-mode ceremony regardless of size.

Compounding it, **two resolution paths disagree**: `work.md` step 1 decides by work size,
`zharness next` decides by plan existence, and the CLI carries more apparent authority than
the playbook.

**Chosen:** shrink P0 from "build a three-shape classification system" to four targeted
fixes — auto-routing weighs work size; one resolution path instead of two; add the
genuinely-new `read-only` shape; templates stop emitting empty sections. Same outcome, a
fraction of the build.

**Rejected:** keeping M8 as drafted and retiring `full`/`bounded` for the new vocabulary —
conceptually cleaner and gives one vocabulary, but rewrites something that already works
correctly and forces a migration of every playbook.

**Rejected:** measuring first before deciding — the routing code is unambiguous on reading;
a probe would confirm what `next.go:56-93` already states outright.

**Limit of this finding:** the mechanism demonstrably produces this symptom. It is *not*
proven to be the cause of the 2026-07-18 virtualizer run specifically — onedrive-cloud is
read-only and out of reach from this environment.

---

## D25 — Each shape gets a stage chain, not just a record set

**Gap found in the same pass:** the spec said which *records* each shape writes but never
which *stages* it runs. With three shapes there are three chains, and none was drawn.

| Shape | Stage chain | Records | Gates |
|---|---|---|---|
| read-only | none — answer in place; `watzup` only when the state *is* the question | none | none |
| bounded | `work` (bounded) → `git` | DB trace only | none |
| durable | `watzup → brainstorm → to-plan → work → check → git → handoff` | plan + trace + decisions | full `check` |

Consistent with `work.md:17`, which already routes rejected-bounded work "through
`brainstorm` and `to-plan`" — the durable chain.

---

## D26 — D19 removes an existing behavior, not only a proposed one

**Recorded for honesty, not re-decided.** D19 was taken believing bounded work was being
designed. It was not: bounded mode exists, and its zero-write rule ends *"The Git diff plus
**captured executable/observable proof** are its durable evidence."* Today's bounded mode
already skips the artifacts **but still captures proof**.

D19 drops the proof capture as well. So its true effect is a **weakening of shipped
behavior**, which is a materially different decision from the one presented at the time.

The owner has been told and the decision stands unless they revisit it. Flagged here so a
later reader does not mistake D19 for a formalization of the status quo.

---

## Standing Constraint

`onedrive-cloud`, `harness-experimental`, and `codesight` are **read-only reference
sources**. Cite them as evidence; never edit them. Every file this plan touches lives
under `/Users/tinhtute/Lab/mono-harness`.
