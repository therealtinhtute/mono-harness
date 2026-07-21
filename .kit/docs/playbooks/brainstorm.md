# Playbook: brainstorm

## Purpose

Turn a raw idea, notes, a trade-off question, or reference files into either a recommendation (explore) or a locked `.kit/planning/SPEC.md` (lock modes). Acts as a brainstorming partner that challenges assumptions, surfaces trade-offs, and recommends the simplest viable path — never accepts the first idea uncritically. Operates by YAGNI (remove speculative scope), KISS (prefer the simpler approach), DRY (deduplicate only when proven painful).

Defer elsewhere: use `to-plan` for roadmap/phase generation from an already-locked spec; use `check` for quality gates after implementation. `brainstorm` never generates implementation phases, task breakdowns, or wave plans.

## Preconditions

- **Version gate**: run `zharness --version` before anything else. A `dev` build (unreleased local build) always satisfies the gate. Otherwise, if the binary is missing or reports a version below `0.1.0` (`MIN_ZHARNESS_VERSION`), print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and stop. Do not proceed without a passing gate.
- Classify the request per `AUTHORITY.md` — locking a spec and firing `intake` is a change-class action.

## Modes

Mode is a hint from input shape, not a commitment. Ask the user only when mode or scope is genuinely ambiguous. A request that explicitly asks to implement, complete, or work autonomously end-to-end **and** already supplies bounded scope plus a checkable success criterion carries **explicit execution intent**: select the obvious mode and proceed without a second procedural confirmation. This intent does not authorize replacing an existing SPEC, deciding an unresolved product choice, or taking a destructive or outward-facing action — ask before any of those.

| Input shape | Mode | Ask before | Artifact |
|---|---|---|---|
| Vague trade-off question, no lock intent (e.g. "REST or GraphQL?") | `explore` | Generating options (intent is ambiguous) | `.kit/reports/brainstorm/{YYYYMMDD}-{slug}.md` |
| Raw idea, notes, or partial draft | `lock-from-idea` | Writing IDEA.md or SPEC.md, unless the request carries explicit execution intent | `.kit/planning/SPEC.md` (+ `.kit/planning/IDEA.md`) |
| `@file:` references to RFC / PRD / README / markdown | `lock-from-files` | Extracting and clarifying, unless the request carries explicit execution intent | `.kit/planning/SPEC.md` |
| Existing `.kit/planning/SPEC.md` present, user wants to revise/update | `refine` | Editing SPEC.md; always ask before replacing/discarding it | `.kit/planning/SPEC.md` (edited in place) |

When multiple shapes co-occur (idea plus a file reference), default to `lock-from-files`; ask only if that choice changes the user's intended scope.

**Ambiguous cases** — resolve by asking the user before producing output:
- User pastes a file but asks "what do you think?" → could be `explore` or `lock-from-files`; ask which.
- User gives an idea but `.kit/planning/SPEC.md` already exists → could be `refine` or a new `lock-from-idea`; ask whether to revise or replace.
- Trade-off question touching an existing locked spec → run `explore` first, then offer `refine` if exploration changes scope.

**Mode transitions** (never silent — always surface to the user):
- `explore` → `lock-from-idea`: user says "lock this" or equivalent; re-confirm scope, then transition.
- `lock-*` → `explore`: user wants to think about alternatives; surface options, then return to lock with the chosen path.
- `refine` → `lock-from-idea`: only if the existing SPEC.md will be discarded — warn the user before overwriting.

## Anti-Pattern: "Too Simple to Brainstorm"

Every input passes through option exploration — todo lists, config tweaks, single-function utilities included. "Simple" is exactly where unexamined assumptions cause the most wasted work later. A 30-second exploration counts; a skipped one does not.

## Scope Guardrail

Discussion clarifies WHAT to build, never adds new capabilities mid-session. Allowed: "How should errors surface?", "What's the empty state?", "Mobile-first or desktop-first?". Not allowed: "Should we also add comments?", "What about search/filtering?" — those are new scope; capture them in `Deferred Ideas` and continue.

## HARD-GATE (mandatory, every session)

Every session must include option exploration before any output. In `lock-from-files` mode, name 1-2 alternatives the source implicitly rejected and why. In every other mode, generate and compare 2-3 viable options. Never produce a SPEC.md or a recommendation without articulating what was *not* chosen and why.

## Steps

1. **Detect mode** from input shape (hint only, see Modes table).
2. **Resolve intent** — when the request carries explicit execution intent and mode/scope are unambiguous, state the resolved mode briefly and proceed. Otherwise ask a short structured question to verify mode and scope. Prefer 1-2 questions per turn; batch up to 4 only when finalizing scope.
3. **Classify the work item** (mandatory for lock modes, best-effort for explore) — declare:
   - input type: `new-spec` | `spec-slice` | `change-request` | `new-initiative` | `maintenance` | `harness-improvement`
   - lane: `tiny` | `normal` | `high-risk`
   - risk flags (only what applies): `auth`, `authorization`, `data-model`, `audit-security`, `external-systems`, `public-contract`, `cross-platform`, `existing-behavior`, `weak-proof`, `multi-domain`
   - affected surfaces: `api`, `browser`, `mobile`, `desktop`, `worker`, `db`, `provider`, `docs`
4. **Gather evidence** — read referenced files; minimum needed.
5. **Generate options & evaluate trade-offs** (mandatory, see HARD-GATE) — 2-3 viable paths in `explore`/`lock-from-idea`/`refine`; in `lock-from-files`, name 1-2 alternatives the source rejected. Use the Decision Frameworks below.
6. **Clarify gaps** (lock modes only) — apply the Clarification Rubric below until goal, scope, constraints, acceptance are lockable.
7. **Recommend or lock**:
   - `explore`: pick one option with rationale and rejected alternatives, then write the Explore report below to `.kit/reports/brainstorm/{YYYYMMDD}-{slug}.md`.
   - lock modes: run `zharness id --json` first; save the returned `id` as the SPEC's own frontmatter `id` (distinct from `intake_id`) and never invent a placeholder. Emit the SPEC skeleton via `zharness scaffold spec --path .kit/planning/SPEC.md --json` and fill it (see Artifacts below); capture rejected alternatives in `Key Decisions`; include classification metadata (step 3) in the header. Immediately after writing SPEC.md: run `zharness init` if no db exists yet (idempotent — `--json` reports `status: "exists"` when already initialized), then `zharness intake --type {input type} --summary "{one-line summary}" --lane {lane} --json` using the step-3 classification; write that separate returned `id` into SPEC.md's frontmatter as `intake_id`.
8. **Self-review** (lock modes only) — apply the Lock Checklist below. Fix issues inline.
9. **User review gate** (lock modes only) — show the SPEC.md path and a concise decision summary. If the original request carries explicit execution intent, self-review found no unresolved product decision, and the next step is neither destructive nor outward-facing, that original intent satisfies this gate: continue to the declared downstream stage without a second procedural response. Otherwise ask the user to approve before suggesting `to-plan`; if changes are requested, edit and re-run step 8.
10. **Hand off** — proceed to `to-plan` after an approved lock (including qualifying explicit execution intent); suggest `refine` if exploration changed scope; suggest `work simple` only when the scoped change is intentionally direct and planning overhead is unnecessary.

## Clarification Rubric (lock modes, apply before locking)

Primary dimensions:
1. **Goal clarity** — is the target outcome explicit? Is the problem statement concrete?
2. **Actor clarity** — is it clear who uses this or benefits from it? Is the affected surface known?
3. **Boundary clarity** — is in-scope work explicit? Is out-of-scope work explicit?
4. **Constraint clarity** — are technical/product limits known? Are timing/dependency constraints known?
5. **Acceptance clarity** — can a later planner tell when the spec is good enough?

If clarity is weak, ask short questions: What is the smallest useful outcome? What will we explicitly not build here? Who is the primary actor? What hard constraints already exist? What result would make the user say "yes, this spec is right"?

**Lock rule**: do not finalize casually. A spec is ready when goal is explicit, scope is bounded, constraints are visible, and acceptance criteria are concrete enough for planning. If not, still write the spec only if unresolved gaps are called out clearly in `Open Questions` and `Ambiguity Report`.

## Decision Frameworks

Pros/Cons table per option:

| Criterion | Option A | Option B | Option C |
|---|---|---|---|
| Complexity | Low | Medium | High |
| Effort | 2h | 4h | 8h |
| Risk | Low | Medium | High |
| Maintainability | High | Medium | Low |

Trade-off questions: Speed vs Quality (can we iterate later?), Simplicity vs Features (what's MVP?), Short-term vs Long-term (technical debt implications?).

Risk assessment: What could go wrong? How likely is failure? What's the impact? How can we mitigate?

Effort sizing: Small (<2h, e.g. bug fix/config change), Medium (2-8h, e.g. new component/refactor), Large (8-24h, e.g. new feature/architecture), XL (>24h, e.g. major rewrite/migration).

YAGNI checklist: is this required now or "just in case"? Will it be used in the first release? Is there a simpler approach? Can it be added later if needed?

KISS checklist: can a junior dev understand this? Are there fewer moving parts possible? Could this be config instead of code? Is abstraction premature?

## Lock Checklist (self-review before showing the user)

Run on `.kit/planning/SPEC.md` (and `.kit/planning/IDEA.md` if present) before the user review gate. Fix issues inline — no need to re-review after fixing.

1. **Placeholder scan** — no `TBD`, `TODO`, "details to be determined", "similar to above", or empty sections; every requirement is concrete and falsifiable; every acceptance criterion is checkable without re-asking the user.
2. **Internal consistency** — goal and acceptance criteria align; In Scope and Out of Scope don't contradict; constraints don't conflict with requirements; architecture (if mentioned) matches feature descriptions.
3. **Scope check** — one coherent feature or module, not multiple independent subsystems (if detected, surface to the user before locking and suggest decomposition into separate specs); `Deferred Ideas` captures anything pulled out of scope.
4. **Ambiguity check** — every requirement has one interpretation, not two; every actor is named (not "the user" without specifying which); every external dependency is named, not implied.
5. **HARD-GATE compliance** — at least 1-2 alternatives explicitly named; rejection reason captured for each; `Key Decisions` documents trade-offs, not just outcomes.

After self-review passes, report: "SPEC written to `.kit/planning/SPEC.md`" plus the chosen mode and key decisions. If the request carries qualifying explicit execution intent (Step 9), continue to its declared downstream stage and include the SPEC summary in the eventual final response. Otherwise ask the user to review and wait for an explicit response; if changes are requested, edit and re-run this checklist. Never infer approval from silence or from a vague request.

Anti-patterns: skipping the placeholder scan because "the spec looks complete" (placeholders hide in formatted text); treating ambiguous or merely conversational input as explicit execution intent; replacing an existing SPEC or deciding a product trade-off without confirmation; re-running this checklist after every minor edit (once per lock cycle is enough).

## Artifacts

### SPEC.md — `.kit/planning/SPEC.md`

Emit the skeleton with `zharness scaffold spec --path .kit/planning/SPEC.md --json`, then fill it — the CLI carries the full template so it no longer lives in this playbook. Frontmatter: `id` (this SPEC's own ULID), `type: spec`, `phase: none`, `lane`, `intake_id`, `created`/`updated`. Body sections: header block (Status, Input Type, Lane, Risk Flags, Affected Surfaces, Downstream, Updated At), Source Mode, Source Inputs, Scenario, Goal, Users / Actors, Requirements, Boundaries (In/Out of Scope), Constraints, Acceptance Criteria, Validation Expectations, Dependencies / Assumptions, Key Decisions, Open Questions, Deferred Ideas, Ambiguity Report.

Rules: requirements numbered and falsifiable; boundaries explicit; acceptance criteria concrete enough for planning; open questions visible, not hidden in prose; risk flags describe why the work is sensitive, not repeat the whole spec; downstream names the next intended step; frontmatter `id` is this SPEC's ULID; `phase` is always `none` (a SPEC precedes phase decomposition — `to-plan` assigns phases); frontmatter `lane` mirrors the body's `Lane:` field, keep both in sync; frontmatter `intake_id` is the ULID `zharness intake` returns at the SPEC-lock step, absent only if the version gate blocked this playbook.

### IDEA.md — `.kit/planning/IDEA.md` (lock-from-idea mode only)

Captures the raw idea verbatim, preserved for future reference alongside the derived SPEC.md.

### Explore report — `.kit/reports/brainstorm/{YYYYMMDD}-{slug}.md`

```yaml
---
title: Brainstorm - {slug}
description: {one-line summary}
status: draft | active | completed
created: YYYY-MM-DD
tags: [brainstorm, {slug}]
---
```

Body order: recommendation first, problem statement, evaluated approaches with pros/cons, rationale, risks, next steps.

## Command Reference

- `zharness --version` — version gate
- `zharness id --json` — mint the SPEC artifact's own ULID before writing frontmatter; do not reuse `intake_id`
- `zharness scaffold spec --path .kit/planning/SPEC.md --json` — emit the SPEC skeleton to fill
- `zharness init` — idempotent; run before the first `intake` if no db exists
- `zharness intake --type {input type} --summary "{one-line summary}" --lane {lane} --json` — fired once, at the SPEC-lock step, after self-review passes; returned `id` goes into SPEC.md's frontmatter `intake_id`

## Exit / Handoff Conditions

- `explore` done: one recommendation with rationale and rejected alternatives, plus a best-effort input-type/lane suggestion when possible, and the Explore report written to `.kit/reports/brainstorm/{YYYYMMDD}-{slug}.md`.
- Lock modes done: `SPEC.md` exists with boundaries, acceptance criteria, classification metadata, `intake_id` recorded, and the review gate is satisfied by either an explicit response or qualifying explicit execution intent from the original request.
- Next handoff must be obvious: `to-plan` after an approved lock, `refine` if scope shifted mid-session, `work simple` only when intentionally warranted.
- Never produce both `.kit/planning/SPEC.md` and a `.kit/reports/brainstorm/` artifact in the same session unless the user explicitly asked for both.
