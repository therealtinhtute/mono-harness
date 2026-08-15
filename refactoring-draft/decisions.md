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
    ├── harness.db
    ├── changesets/
    ├── log/
    └── conflicts/
```

Also removes the root-level `harness.db` clutter.

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

## Standing Constraint

`onedrive-cloud`, `harness-experimental`, and `codesight` are **read-only reference
sources**. Cite them as evidence; never edit them. Every file this plan touches lives
under `/Users/tinhtute/Lab/mono-harness`.
