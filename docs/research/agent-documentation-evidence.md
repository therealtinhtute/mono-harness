# Evidence: what repository documentation actually does for coding agents

Authored. External-literature evidence gathered 2026-08-20 to settle one question: should the harness *generate* documentation for consumer repositories, or *guard* documentation those repositories already have?

Requirements in `docs/plans/active/consumer-doc-drift-gate.md` cite this page as authority. Nothing here describes this repository's own behavior — for that, read [`docs/ARCHITECTURE.md`](../ARCHITECTURE.md).

## Method and limits

Six web searches via Exa on 2026-08-20, covering three angles: measured effect of repository context files on agent task success; how doc-staleness detection is built in shipping products; and whether code-shape doc generators cover languages beyond TypeScript. Sources are a mix of peer-reviewed-style preprints (arXiv), vendor engineering blogs, and one practitioner benchmark.

Limits that matter when citing this page:

- The vendor posts (doccupine, Doc Bridge, baz.ai, Falconer, Dosu) each sell the thing they describe. Their *mechanisms* are checkable and are what this page extracts; their effect claims are not independent.
- The two agent-benchmark preprints disagree in emphasis and use different agents, task sets, and injection strategies. Where they agree (F1, F2) the finding is treated as solid; where only one reports an effect it is labelled as such.
- No finding here was reproduced locally. This page is literature, not measurement. A local measurement against a real consumer repository would be stronger and does not exist yet.

## F1 — Generated context files are net-negative on repositories that already have docs

*Evaluating AGENTS.md: Are Repository-Level Context Files Helpful for Coding Agents?* (Gloaguen, Mündler, Müller, Raychev, Vechev; arXiv 2602.11988, 2026-02-12) evaluated four coding agents across SWE-bench Lite and a purpose-built AGENTbench.

| Condition | Task success | Inference cost |
|---|---|---|
| LLM-generated context file | −0.5% (SWE-bench Lite), −2% (AGENTbench) | +20% / +23% |
| Developer-written context file | +4% average vs. no context | up to +19% |
| LLM-generated, **after deleting every `.md` and `docs/`** | **+2.7%**, beating developer-written docs | not reported |

The third row is the load-bearing one. Generated context earns its keep exactly when there is nothing else to read, and costs performance when it duplicates documentation the repository already carries. The authors' explanation is redundancy: generated overviews restate what the repository already discloses, and those tokens compete for attention with task-relevant ones.

A second result from the same paper is narrower and more surprising: *"context files, even developer-provided ones, are not effective at providing a repository overview."* Measured by steps-until-the-agent-touches-a-file-in-the-gold-patch, context files did not speed up file discovery on either benchmark.

**Implication for this repository.** A consumer repository that has run `zharness init` has a populated `docs/`. Generating a wiki on top of it lands in the degraded regime, not the +2.7% one. This is the strongest single argument against building a generator into the CLI, and it is why `docs/plans/completed/docs-architecture.md` NG5 (no AST-derived documentation) survives contact with outside evidence rather than resting on taste.

## F2 — On-demand retrieval saves cost, not correctness

*Do Context Files Help Coding Agents? A Two-Agent Ablation Study on Real Repositories* (arXiv 2607.27250) ran 291 agent runs across two frontier agents and 17 real tasks, comparing three injection strategies: `none`, `always_on` (context file resident every turn), and `selective` (a short retrieval hint plus topic-organized `wiki/*.md` read on demand).

- **Correctness: null across all three strategies.** For two repositories the `selective` wiki was roughly 10× and 18× the word count of the `AGENTS.md` it replaced, and the larger corpus still did not raise pass rates.
- **Cache-creation tokens: `selective` lower than `none` on 11/11 tasks** (p = 0.001, Holm-corrected p = 0.012). The authors read this mechanically — it follows from delivery mechanics, not from the agent becoming more capable.
- **One dose-dependent behavioral effect.** On the single repository whose context file warns that the full test suite takes over 20 minutes, blind full-suite invocations per cell fell monotonically: `none` 3.67 → `always_on` 2.44 → `selective` 1.67.

The authors' own summary of the failure mode: real tasks fail on implementation skill, not on missing repository knowledge a context file could supply. Across both agents the real context file never converted a near-miss into a pass.

**Implication.** Two corrections to claims made earlier in this initiative's discussion. First, token-reduction figures quoted for wiki tooling are cost-side savings, not capability gains — they should never be presented as "the agent does better." Second, the one thing that measurably changed agent behavior was a *gotcha* (this suite is slow), which is exactly the principle `docs/audit/consumer-adoption-audit.md` D4 states — "spend the budget on gotchas, not on what the filesystem shows" — now with an external number attached.

## F3 — Documentation's measurable value is rationale, not code shape

*Code-QA-Bench: Separating Code Reasoning from Documentation Memorization in Repository-Level QA* (arXiv 2605.29277) evaluated four frontier models over 528 code-derivable and 100 doc-dependent tasks across 10 repositories.

| Comparison | Mean delta |
|---|---|
| code access vs. closed-book | **+0.231** |
| documented vs. code-only, on doc-dependent tasks | +0.071 (p < 0.003) |
| documented vs. code-only, on code-derivable tasks | **+0.007** (≈ zero) |

Where documentation helps, it helps on the completeness axis, supplying *"design rationale, deprecation warnings, edge case caveats"* — the details that cannot be inferred from code patterns.

**Implication.** Documentation that restates code shape scores approximately zero, because the agent recovers that from the code. Decision records score. This is direct external support for the ownership split already in `docs/README.md` and for `docs/decisions/` being the highest-value authored surface in the tree.

## F4 — Structural forensics is the one thing automation does better than a human brief

The sourcebook benchmark (2026-03-28) compared no-context, handwritten brief, full-repo dump (Repomix), and generated structural context across tasks in cal.com, hono, and pydantic.

- Full-repo dump performed worse than a targeted handwritten brief, and sometimes worse than no context at all — an independent replication of F1's direction.
- Handwritten briefs won on *workflow recipes* ("when the project needs i18n, use this exact pattern").
- Generated structural analysis won on facts no human recalls: hub files (`types.ts` imported by 183 files), 14 generated files that must not be edited, an 88% co-change coupling between `auth/provider` and `middleware`, circular dependencies. Output for a 10,453-file repository: 858 tokens.

Their conclusion: *"Repo dumps show the code. Handoffs explain the project."*

**Implication.** If code-derived documentation is ever revisited for this harness, the defensible slice is git-and-static-analysis forensics (co-change, hub files, generated-file traps), not prose overviews. That slice is explicitly out of scope for the current initiative and is recorded here so a future initiative does not have to rediscover it.

## F5 — Four independent products converge on the same four-piece drift mechanism

**doccupine** (2026-07-24) states the requirement set directly: exact drift detection needs a link from page to source, a **pinned commit SHA** the page was written or verified against, a change signal, and a size measurement. On the piece people skip:

> A link without a baseline tells you the file has commits; it cannot tell you which of them happened after your page was written. You get an alert on every push forever, which trains everyone to ignore alerts.

And on ordering: *"Any system that auto-publishes documentation needs hand-edit detection before it needs anything else"* — a content hash of what the generator last wrote, compared against what is on disk now.

**Doc Bridge** (`ak-docs gate run index-freshness`) implements the CI half and names a trap worth quoting:

> If CI rebuilds the index before checking it, the generated files in the branch are no longer the evidence being tested. The job proves that a fresh index can be generated in the runner. It does not prove that reviewers saw or committed the changed routing context.

Its gate fails closed with `expected: b099695d… / actual: 359355e5…` and exit code 1. Its author is explicit that freshness is not correctness: a fresh index can faithfully encode a bad decision.

**baz.ai Skills Maintainer** (2026-08-10) targets instruction files specifically, and names why they fail differently from prose docs: *"A developer reading bad documentation notices and works around it. An agent reading a stale rule just follows it, on every pull request that touches the path."*

**freemansoft** (2026-07-13) contributes the discrimination rule — the directory decides which documents owe the code the truth:

```
docs/*.md       ← must match the code. The gate checks these.
docs/specs/     ← point-in-time records.
docs/plans/     ← frozen once the work ships.
docs/reviews/   ← the gate ignores them.
```

Rewriting a superseded design document to match today's code *falsifies the record*, so those directories must be exempt by construction rather than by an agent's judgment. Paired with a severity rule: *"Grade the severity, or the gate gets tuned out. A gate everyone skips is worse than no gate, because it looks like coverage."*

**Implication.** This repository already holds three of the four pieces and the directory discrimination, without having built any of it for this purpose:

| Mechanism piece | Status here |
|---|---|
| Hand-edit detection (content hash) | present — three-hash conflict staging in `managed_docs`, **managed docs only** |
| Directory discrimination | present — `docs/plans/` and `docs/decisions/` are point-in-time by construction; `docs/ARCHITECTURE.md` is the top-level doc that owes the code truth |
| Link from page to source | present in content — `docs/ARCHITECTURE.md` carries 21 distinct `path/to/file.go:NN` citations (verified 2026-08-20), e.g. `cli/internal/application/plan_write.go:36`, `cli/internal/application/managed_docs.go:107` |
| **Pinned baseline SHA** | **absent** |
| Change signal + size measurement | absent (git provides both; nothing reads them) |

The gap is narrow and specific: authored documentation has zero integrity checking, while managed documentation has a full three-hash mechanism. `docs/decisions/0004-docs-directory-deletion-655c6ac.md` is the incident where that asymmetry cost 26 files, and its closing line — *"This ADR records the incident; it does not prevent the next one"* — names the same gap from the inside.

## F6 — Wiki generation as a dependency, not a build

**DeepWiki** (Cognition) auto-generates architecture diagrams, source-linked summaries, and Q&A for a repository; steerable via `.devin/wiki.json` (`repo_notes` and an explicit `pages` array, max 30 pages / 80 enterprise); exposed over a free no-auth MCP server at `https://mcp.deepwiki.com/mcp` with three tools (`read_wiki_structure`, `read_wiki_contents`, `ask_question`); auto-refreshes repositories carrying a DeepWiki badge.

**Disqualifier for the case at hand:** the free tier covers **public repositories only**. Private consumer repositories require a paid Devin account.

**codesight** compiles an AST-derived wiki locally with no LLM call, but its AST support is TypeScript/JavaScript only (Express, Hono, Fastify, Koa, Elysia, NestJS, tRPC, Drizzle, TypeORM, React), falling back to regex elsewhere.

**CodeWiki** (ACL 2026) generates hierarchical documentation via LLM multi-agent decomposition, scoring 68.79% on its quality metric, and explicitly does **not** address staleness.

**Implication.** Every option in this category is either language-restricted, paywalled for the private case, or has no answer for the problem F5 shows the whole market is organized around. The generation half is crowded and the guarding half is not — and guarding is the half that matches what this CLI already is.

## What this evidence rules out

- **Building a documentation generator into `zharness`** — F1 says it degrades performance on repositories that already have `docs/`, which is every repository that has run `zharness init`.
- **Justifying such a generator by token savings** — F2 shows the savings are cache-side with no correctness gain.
- **Adopting DeepWiki as the consumer-repository answer** — F6, private repositories are paywalled.
- **Prose overviews as the generated artifact, if generation is ever revisited** — F3 and F4 both put the value in rationale and structural forensics, not overview prose.

## Sources

| Source | Type | URL |
|---|---|---|
| Evaluating AGENTS.md (arXiv 2602.11988) | preprint | https://arxiv.org/html/2602.11988v1 |
| Two-Agent Ablation Study (arXiv 2607.27250) | preprint | https://arxiv.org/html/2607.27250 |
| Code-QA-Bench (arXiv 2605.29277) | preprint | https://www.alphaxiv.org/abs/2605.29277 |
| sourcebook context-file benchmark | vendor benchmark | https://sourcebook.run/blog/we-benchmarked-ai-context-files |
| doccupine, Documentation Drift | vendor blog | https://doccupine.com/blog/documentation-drift |
| Doc Bridge, stale context failing CI | vendor blog | https://dev.to/agentskit/i-made-stale-coding-agent-context-fail-ci-instead-of-failing-silently-3434 |
| baz.ai Skills Maintainer | vendor blog | https://baz.ai/resources/blog/skills-maintainer-keeping-your-instruction-files-true |
| Wiring a Documentation Gate Into Code Review | practitioner blog | https://joe.blog.freemansoft.com/2026/07/wiring-documentation-gate-into-code.html |
| DeepWiki | product docs | https://docs.devin.ai/work-with-devin/deepwiki |
