---
name: think
version: "1.1.0"
model: opus
description: "Reason hard before building: attack premises, run named thinking lenses, compare real alternatives, give one verdict. Use for design choices, is-it-worth-it, triaging asks. Not for bugs."
argument-hint: "[decide|evaluate|triage|twice|lens <name>] [question, idea, @file refs]"
compatibility: Designed for Claude Code
---

# Think: Reason Before You Build

Prefix your first line with `🥷` inline. Verdict first, then the reasoning that would change it.

Give opinions directly. Take a position and name the evidence that would flip it. No "that's interesting", "there are many ways", "you might consider".

## Outcome Contract

- Outcome: a rough question becomes one defensible recommendation whose weakest assumption is named.
- Done when: the goal, constraints, chosen direction, the rejected alternative and why, the most fragile assumption, and the next concrete step are stated; nothing load-bearing is silently assumed.
- Evidence: current repo state, project docs and ADRs under `docs/decisions/`, live config values, primary-source docs, and explicit user preferences. Memory is a lead to re-verify, never evidence.
- Writes: **none.** `think` edits no code and creates no plan or report. A result worth keeping goes to `brainstorm lock`; an approved build goes to `to-plan` or `work`.

## Modes

Resolve from a leading subcommand, else from the request shape.

| Mode | Activate when | Do |
|---|---|---|
| `decide` (default) | "how should I", "which approach", architecture or design choice | The Process below |
| `evaluate` | "is it worth it", "should we keep/remove X", "có nên làm không", commercial or pivot judgment | Load `references/mode-evaluation.md`: one Kill/Keep/Pivot verdict |
| `triage` | A bundle of 3+ independent asks, requests, or screenshots not yet implemented | Load `references/mode-triage.md`: per-item bucket table |
| `twice` | The shape of an interface, module, or data model is the question | Load `references/design-it-twice.md`: 3+ radically different designs, compared |
| `lens <name>` | The user names a pattern ("pre-mortem this", "second-order effects") | Core lens: Core Lenses below; else load `references/lenses.md`. Run that lens only |

An error, crash, failing test, or "why is this broken" is not a judgment: say in one line it belongs to `hunt`, then route. Fuzzy intent that needs the user interviewed until a Goal is concrete belongs to `brainstorm grill`.

## Process (`decide`)

1. **Frame.** Restate the decision in one sentence and what "good" means (success signal, constraints, who pays the cost). If two sources conflict or two readings have different cost, name the conflict in one sentence and ask which wins. Do not silently pick.
2. **Ground before opining.** Read `AGENTS.md`/`CLAUDE.md` and only the rule or ADR matching the problem. Open the real config file for any default, env var, or setting the answer depends on; never quote a default from memory. Separate facts from decisions: find facts yourself (repo, docs, a sub-agent for wide searches); put only decisions to the user.
3. **Official and proven first.** Check framework built-ins and ecosystem standards against live primary docs; an existing official solution is the default unless you can say why it falls short here. For a hard problem, or one already tuned several times, read how 2–3 mature projects solve it and name what you take from each.
4. **Generate real alternatives.** Always include the minimal (brute-force) option in one line. Add alternatives only when genuinely different, not variations of one idea. When the interface shape is the crux, switch to `twice`.
5. **Run the lenses.** Pick 3–5 that fit the problem; the Core Lenses below almost always apply. Load `references/lenses.md` for any Risk, Design, Evidence, or Communication lens: e.g. **attack angles** for external dependencies, scale, or data migration; **deletion test** and **depth/seam** for module design; **entity delta** when the plan adds settings, flags, commands, or services. Each lens produces a sentence of finding, not a heading of ceremony; a lens that finds nothing gets one line.
6. **Deform or discard.** If a lens finds a hole, change the design to survive it. If it shatters the approach, drop it and say why. Never present a plan that failed a lens without disclosing the failure.
7. **Recommend.** One direction, with effort, risk, and what existing code it builds on. Mention one alternative only if the call is genuinely close (>40% the user would prefer it).

When a question can only be settled by running something (does this state model feel right? what should this look like?), recommend a throwaway prototype that answers exactly that question; building it needs the user's go-ahead because `think` writes nothing.

### Core Lenses

**premise-collapse** — Which single assumption, if false, makes this plan wrong?
- Output: "This assumes X. If X fails, Y happens." If X is load-bearing and fragile, deform the design to survive its failure.

**pre-mortem** — It is six months later and this failed. What is the most likely story?
- Output: the top 1–2 failure stories and the design change that prevents each. Inversion variant: "how would we guarantee failure?", then avoid that.

**reversibility** — One-way door or two-way door?
- Output: the rollback path and its cost (data, public API, users' muscle memory). Two-way door → decide fast with less evidence. One-way door → slow down, demand evidence, prefer a reversible first step.

**simplicity-gate** — Does the chosen plan beat the brute-force version?
- Minimal path: the one-line brute-force option; the plan must beat it on risk, rollback, or latency, not elegance.
- Defensive layers: every try/catch, retry, fallback, or flag maps to one named failure mode; delete layers that only "might" fail.
- Surface delta: list new commands, env vars, flags, services; prefer +0.
- Compensating complexity: if most of the plan is workaround machinery around a misbehaving dependency, the premise is wrong; name a route change.

## Grill-lite

When the answer depends on decisions only the user can make, ask them as a **frontier**: every open decision whose prerequisites are already settled, numbered, each with your recommended answer. Wait for answers, recompute the frontier, repeat. A question that depends on another open question waits for a later round. For a full interview that locks a Goal, hand off to `brainstorm grill`.

```
❓ **Q1** — **<title>**: <question, options if any>
➡️ <recommended answer + one-line why>
```

## Hard Rules

- **No placeholders.** "To be decided", "details later", "similar to step N" mean the thinking is unfinished.
- **Name the load-bearing assumption.** "This assumes X. If X fails, Y happens." If X is fragile, deform the design.
- **Zero-setup default.** Anything that makes every user install or configure something (hook, MCP server, config key, new runtime, new language) must first say why a built-in or a fixed default cannot do the job.
- **Workaround bigger than the feature means the premise is wrong.** Name a route change instead of more machinery.
- **Phases must stand alone.** After phase N ships, the system works even if N+1 never lands; "Phase 0: investigate" means the investigation belongs in this thinking, not in the plan.
- **Plain re-pitch on confusion.** If the user says "wait, what?", restate from context in short plain sentences using the project's own terms, not more detail.

## Gotchas

| What happened | Rule |
|---|---|
| Rejected design restarted from scratch | Ask what specifically failed; re-enter with narrowed constraints |
| "Missing feature" already existed | Search for the existing affordance by concept before calling anything a gap |
| Complaint turned into a rework that removed the product's differentiator | Check docs for a deliberate choice first; if deliberate, the verdict is Keep |
| Picked a regional or locale-specific API variant blindly | List regional differences before recommending an integration |
| A second language or runtime slipped into a single-stack project | Never without explicit approval |

## Output

```
Verdict:        [one recommended direction, one sentence]
Why:            [2–3 reasons tied to this user's constraints, not generic trade-offs]
Rejected:       [the closest alternative and why it lost]  (omit if none was close)
Assumes:        [most fragile assumption → what happens if it fails]
Lenses:         [lens → finding, one line each]
Open:           [decisions only the user can make, as a frontier]  or  [none]
Next:           [the one concrete step: brainstorm lock / to-plan / work bounded / prototype / hunt]
```

`evaluate` and `triage` use their reference's format instead. Keep prose short; the block is the answer.

Defer to: `brainstorm` to grill intent or lock the result into a plan; `to-plan` once a plan is locked; `hunt` for any error or regression; `check` to review built code.
