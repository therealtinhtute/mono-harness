# Playbook: brainstorm (grill loop)

Loaded by `docs/playbooks/brainstorm.md` for `grill` and `raw`, and with `brainstorm-lock.md` for `lock`/`refine`.

## Steps

4. **Clarify the boundary** — grill whatever is missing (the whole Goal in `grill`; gaps only in `lock`/`refine`). Stop rather than invent an unresolved product decision.

## Grill

Grill until the **Done** condition below holds. Map decisions as a tree and work in rounds.

- **Frontier** — every open decision whose prerequisites are settled; ask it all in one round (question tool, ≤4 per call, recommended option first). Recompute after each answer; a contradiction reopens its branch.
- **Facts vs decisions** — find facts yourself (repo, docs, `docs/memory/`, or a sub-agent briefed per `work-full.md` step 4); put decisions to the user.
- **Stress-test** — one canonical term per concept (reuse `docs/PROJECT.md`); probe edge cases; raise code-vs-statement contradictions as questions.
- **Absent owner** — `open_question: <question> | owner: <who>`; it blocks the lock only if it blocks a requirement.
- **Done** — `to-plan` could plan without asking: concrete `outcome`, checkable `success_signal`, `actors`, `authority`, requirements each with `acceptance:`, `non-goals`. Summarize the tree, list open questions, mark hard-to-reverse trade-offs `adr candidate`, and ask: lock (or refine) now?
- **Grill on a plan** — Goal gaps → offer `refine`; gaps in planned phases are findings (phases are immutable).
- **`raw`** — at most 2 rounds, then print decided / assumed / open. No lock offer.
