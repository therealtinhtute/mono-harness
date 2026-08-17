# Reference Models

Three external sources read for structure, grouped by lineage. **None is adopted as a
dependency** — the mechanisms are extracted, the code is not.

- `harness-experimental` (local copy) + `hoangnb24/repository-harness` (upstream, ahead)
  — same lineage
- `Houseofmvps/codesight` — generated knowledge base

---

## harness-experimental (local copy + upstream)

Same lineage, two snapshots: the local copy is behind, the public upstream is ahead.

### Local copy — `/Users/tinhtute/Lab/harness-experimental`

Rust, `harness-v0.1.10`.

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

### Resident cost (local copy)

`AGENTS.md` 1.6KB + `docs/WORKFLOW.md` 5KB = **~1.6k tokens**.
Roughly 10x lighter than onedrive-cloud's ~6.6k.

### Upstream (ahead) — `hoangnb24/repository-harness`

`github.com/hoangnb24/repository-harness` — Rust, public, `pushed_at 2026-08-13`.
"Turn any repo into an agent-ready workspace." Read on 2026-08-15; five mechanisms
the local copy did not show.

#### A — The ADR template is 10 lines, and it carries `Status`

`docs/templates/decision.md` is **331 bytes**:

```
# NNNN Decision Title      Date: YYYY-MM-DD
## Status        Proposed | Accepted | Superseded | Rejected
## Context       What problem, constraint, or ambiguity forced this decision?
## Decision      ## Alternatives Considered   ## Consequences (Positive / Tradeoffs)
## Follow-Up
```

`Status` is the mechanism that makes supersession work without deletion. The draft
spec's P2 template omitted both `Status` and `Follow-Up`. Take them.

#### B — Promotion is filtered, not blanket

`docs/decisions/README.md` states the trigger set outright:

> **Add A Decision When** — a lasting product or architecture choice changes; public
> compatibility or data ownership changes; security or recovery policy changes;
> validation is materially added, removed, or weakened; or the source-of-truth
> hierarchy changes.

And the exclusion, in one line: **"Task-local choices stay in the active plan."**

This is a direct correction to draft P2, which gates *completion* on promotion and
would therefore flood `docs/decisions/` with task-local noise. See
[`open-questions.md`](open-questions.md) Q2.

Also: *"Installed consumers begin with an empty decision index and add only real
consumer choices."* The scaffold ships the index and the criteria — never upstream's
own ADRs.

#### C — Deliberate absence is a retrieval feature

Both `docs/README.md` and `docs/decisions/README.md` carry a `## History` section
whose entire job is to explain what is **missing** and why:

> The former SQLite control plane, protocol v1, story packets, migration evidence, and
> compatibility documentation are preserved by Git history and immutable
> `harness-cli-v*` tags. They are intentionally absent from the current tree so search
> and agent retrieval return current product authority.

Removal is the point; git history and immutable tags are the provenance. Stronger than
the draft's "plans stay disposable" — this makes *pruning* a named, explained act
rather than a silent one. The draft spec has no equivalent.

#### D — A third knowledge kind: enforced invariants

`docs/patterns/encoding-invariants.md` (4.4KB) + ADR `0028-authoritative-invariant-encoding`
define a category the draft's two-way synthesis missed — a rule compiled into
repository-native validation. It does not rot, because CI fails.

The discipline is what makes it usable:

| Step | Rule |
|---|---|
| **Authority gate** | An accepted document or explicit owner decision must state the rule. *"Code organization, repeated patterns, tests, tool defaults, configurable examples, and undocumented preferences show current behavior or convention; they do not authorize a new invariant."* Stop and ask if two boundaries fit the words. |
| **Boundary table** | Authority · Scope · Allowed · Forbidden · Exceptions · Diagnostic — written before choosing a tool |
| **Native owner** | Integrate into the repo's existing test/lint/validation command. Never add a framework to enforce one rule. |
| **Both directions** | Positive proof (allowed case passes) *and* negative proof (forbidden fixture fails with the intended diagnostic). *"A passing repository with no exercised violation does not prove that the guard can detect recurrence."* |
| **Enforcement ladder** | local validation → optional hook → checked-in CI → branch protection. **None proves another.** *"A checked-in CI job does not prove that it ran on the current revision. A green job does not prove branch protection requires it."* |

The diagnostic standard is concrete:

```text
public/orders imports internal/storage: public packages must not import internal
packages (docs/architecture.md). Depend on the public storage interface instead.
```

Violating item · broken rule · **authority pointer** · next action — not `validation failed`.

Directly applicable to `zharness audit`, whose findings today make no distinction
between "a check exists" and "a check ran and passed."

#### E — `CLAUDE.md` gets a managed block, not a one-time emit

```markdown
<!-- HARNESS:BEGIN -->
## Harness
Claude Code does not auto-load `AGENTS.md`. Import that single canonical
project instruction source. Keep this bare `@` line outside backticks so the
import remains active.

@AGENTS.md
<!-- HARNESS:END -->
```

258 bytes. It self-heals on update, where the draft's "emit when absent" does not.
The backtick gotcha is encoded in the block itself.

Note also `.agents/skills/` rather than `.claude/skills/`, each skill carrying an
`agents/openai.yaml` adapter — vendor-neutral by layout.

#### F — Measured resident cost

| File | Bytes | ~Tokens |
|---|---|---|
| `AGENTS.md` | 1,613 | 403 |
| `docs/README.md` | 1,622 | 405 |
| `docs/WORKFLOW.md` | 5,564 | 1,391 |
| **Comparable to the draft's gate** | **8,799** | **~2,199** |
| `docs/HARNESS.md` (on-demand, principles) | 1,900 | 475 |

**The draft spec's ≤1,000-token gate is 2.2x tighter than the best-in-class reference
it cites.** See [`open-questions.md`](open-questions.md) Q1 — this is the sharpest
unresolved item in the draft.

Worth noting the split: `WORKFLOW.md` is procedure and resident; `HARNESS.md` is
principles and read on demand. Separating them is what keeps the resident path at
~2.2k instead of ~2.7k.

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
