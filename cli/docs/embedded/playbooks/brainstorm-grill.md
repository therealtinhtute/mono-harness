# Playbook: brainstorm (grill loop)

Loaded by `docs/playbooks/brainstorm.md` for `grill` and `raw`, and with `docs/playbooks/brainstorm-lock.md` for `lock` and `refine`.

## Steps

4. **Clarify the boundary** — run the Grill loop below over whatever is missing (the whole Goal in `grill`; only the gaps in `lock`/`refine`). Require a concrete outcome, a measurable `success_signal:`, the affected actors, constraints, accepted requirements each with an `acceptance:` check, and non-goals. Stop rather than invent an unresolved product decision.

## Grill

Grill **relentlessly**. Map the request as a **design tree**: each decision branches into the decisions that depend on it. Work it in **rounds**.

- **Frontier** — every open decision whose prerequisites are settled. Ask the whole frontier in one round; a question that depends on another still open this round waits for a later round.
- **Format** — ask with the agent's own question tool (Claude Code `AskUserQuestion`, Codex `request_user_input`, Gemini CLI `ask_user`, omp `ask`): at most 4 per call, recommended option first, labelled `(Recommended)`; split a larger frontier across calls. Last-resort fallback, number each question: `❓ **Q1** — **<title>**: <question, with choices>` then `➡️ <recommended answer>`. Every question carries a recommendation.
- **Recompute** — after each round, settle what was answered and recompute the frontier; an answer that contradicts a settled decision reopens that branch.
- **Facts vs decisions** — facts are yours: read the repo, docs, and `docs/memory/`, or dispatch a sub-agent, and ask the rest of the frontier while it runs. Decisions are the user's: put each one to them and wait.
- **Stress-test** — pin each fuzzy or overloaded term to one canonical term (reuse `docs/PROJECT.md` vocabulary); probe boundaries with concrete edge-case scenarios; when the user's statement and the code disagree, put the contradiction to them as a question.
- **Absent owner** — a decision only someone else can make stays open as `open_question: <question> | owner: <who>`; keep grilling the other branches. An open question that blocks a requirement blocks the lock (step 4's stop rule).
- **Done** — the frontier is empty and `to-plan` could plan from the Goal without asking a single question: `outcome`, a checkable `success_signal`, `actors`, `authority`, each requirement with `acceptance:`, and `non-goals` are all concrete. Then summarize the settled tree, list any `open_question:`, mark as `adr candidate` each decision that is hard to reverse, surprising without context, and the result of a real trade-off, and ask: lock (or refine) now?
- **Grill on a plan** — read the plan first. Goal gaps → offer `refine`. Gaps inside `to-plan`'s phases are findings for the user, since phase definitions are immutable once planned.
- **`raw`** — at most 2 rounds, then stop and print three lists: decided, assumed, open. No lock offer.
