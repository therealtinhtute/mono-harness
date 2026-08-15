# Reference Models

Two external repos read for structure. **Neither is adopted as a dependency** — the
mechanisms are extracted, the code is not.

---

## harness-experimental

`/Users/tinhtute/Lab/harness-experimental` — Rust, `harness-v0.1.10`.

Notable: it **retired** the SQLite control plane (decision 0027, "End Protocol V1 And
Focus The Repository Protocol") and moved to repository-as-system-of-record. mono-harness
runs the generation they deprecated. Worth knowing before investing in optimizing it —
though owning the CLI is exactly what makes keeping it the right call here.

### The split, stated explicitly in `docs/README.md`

```
Consumer-Owned Truth   -> README, product/, ARCHITECTURE, decisions/, code, tests, CI
Harness-Owned Process  -> WORKFLOW.md, plans/, templates/, patterns/
```

### Three structural moves

1. **`docs/README.md` is a 40-line documentation map.** The router. Nothing else resident.
2. **`CLAUDE.md` is 11 lines: `@AGENTS.md`.** One canonical instruction file — no
   shim-vs-canon confusion of the kind `KNOWNS.md` created.
3. **`docs/decisions/` absorbs what plans would otherwise carry forever.** Plans get
   *deleted* after completion because the lasting result lives in an ADR plus tests plus
   git history. Their `docs/plans/active/` currently reads "No durable work is currently
   active" — an empty active directory is the healthy state.

### Rules worth quoting

> Completed plans may be removed from the current tree when decisions, code, tests, and
> Git history preserve their lasting result. This keeps current retrieval focused
> without deleting provenance.

> Do not split one task into story, design, trace, and validation records without an
> independent audience.

> Start with the smallest authoritative surface.

### Resident cost

`AGENTS.md` 1.6KB + `docs/WORKFLOW.md` 5KB = **~1.6k tokens**.
Roughly 10x lighter than onedrive-cloud's ~6.6k.

---

## codesight

`github.com/Houseofmvps/codesight` — TypeScript, 1,359 stars.
"Universal AI context generator. Saves thousands of tokens per conversation."

### Structure

| Tier | File | Size | Read when |
|---|---|---|---|
| Router | `.codesight/wiki/index.md` | 2.0KB (~500 tok) | New session — orientation only |
| Articles | `.codesight/wiki/*.md` (8) | 0.5–6KB (125–1500 tok ea) | Domain question, then read the source it names |
| Facets | `.codesight/{routes,config,graph,libs,events,coverage,middleware,cicd}.md` | 0.15–13KB | Single-concern lookup |
| Full dump | `.codesight/CODESIGHT.md` | 20.9KB (~5.8k tok) | Complete context |

`docs/` holds exactly **one** file (`wasm-plugins.md`). `README.md` (51.8KB) is the
product doc. The generated map is deliberately kept out of `docs/`.

### Seven mechanisms worth taking

**1. The knowledge base states its own ROI.**
```
> Token savings: this file is ~5,800 tokens. Without it, AI exploration would cost
> ~35,900 tokens. Saves ~30,200 tokens per conversation.
```

**2. Freshness stamp + regeneration command in every file.**
```
> Last scanned: 2026-07-27 13:30 — re-run after significant changes
_Generated 2026-07-27 — re-run `npx codesight --wiki` if the codebase has changed._
```

**3. Epistemic boundary, stated up front.** The map declares it is not the territory:
> **How to use safely:** These articles tell you WHERE things live and WHAT exists. They
> do not show full implementation logic. Always read the actual source files before
> implementing new features or making changes. Never infer how a function works from the
> wiki alone.

**4. Explicit negative space.** A `## What the Wiki Does Not Cover` section listing eight
concrete blind spots — dynamic routes, npm-internal routes, WebSocket/SSE handlers, raw
SQL tables, computed fields, TS types that aren't columns, `[inferred]` low-precision
entries, partial gRPC/tRPC/GraphQL capture. Plus inline confidence markers: `[inferred]`
= regex-detected, `✓` = test-covered.

**5. Routing by question type, priced.**
```
- New session:            read index.md for orientation — WHERE things are
- Architecture question:  read overview.md (~500 tokens)
- Domain question:        read the article, then read those source files
- Before implementing:    read the source files listed in the article
- Full source context:    read .codesight/CODESIGHT.md
```

**6. Every fact carries its source path.** `middleware — src/detectors/middleware.ts`.
An index *into* source, never a substitute for it.

**7. Blast radius as a first-class artifact.** `graph.md`:
```
## Most Imported Files (change these carefully)
- `src/types.ts`   — imported by **51** files
- `src/scanner.ts` — imported by **17** files
```
Knowledge no prose document ever contains.

### The critical constraint

`.codesight/` is **100% AST-derived — deterministic, no LLM, 200ms.** It holds zero
decisions and zero reasoning. That is deliberate, and it is why it never goes stale.

It also generates the AI-config shims. From `.gitignore`:
```
# Generated AI-config outputs written by the test suite
tests/fixtures/**/CLAUDE.md
tests/fixtures/**/AGENTS.md
tests/fixtures/**/codex.md
tests/fixtures/**/.cursorrules
tests/fixtures/**/copilot-instructions.md
```
The "compatibility shim" pattern `KNOWNS.md` claims but never implements.

---

## The Synthesis

Two kinds of knowledge, different decay rates, different homes:

| | **Derived map** | **Durable memory** |
|---|---|---|
| Holds | routes, env vars, import graph, coverage, high-impact files | why we bounded caches, what we rejected, product intent |
| Source | mechanical derivation from code | human judgment |
| Decays | every commit | rarely |
| When stale | regenerate (~200ms) | **cannot regenerate — must be written** |
| codesight | `.codesight/` ✓ | not its job |
| harness-experimental | not its job | `docs/decisions/`, `docs/product/` ✓ |
| onedrive-cloud today | stale hand-written `codebase-summary.md` pretending to be this | scattered across 3 roadmaps and 2 specs, or gitignored |

**The diagnosis:** onedrive-cloud hand-maintains what should be generated, and gitignores
what can only be written. mono-harness ships neither half, so consumers invent both badly.
