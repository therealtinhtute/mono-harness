# IDEA: Agent-agnostic workflow chain — docs-first inversion

Date: 2026-07-18
Source: session analysis comparing `skills/workflow/` + `zharness` against
`hoangnb24/repository-harness` (the upstream concept this harness mirrors).

## Raw idea

Today the workflow chain's logic lives inside 8 Claude Code `SKILL.md` files
(plus their `references/`), and `zharness` is just the tool those skills call.
That couples the whole lifecycle to one agent runtime. Invert it:

- Canonical per-stage **playbook docs** become the source of truth, embedded in
  the `zharness` Go binary and written out by `zharness init` (which also
  scaffolds `.kit/` and an `AGENTS.md` shim).
- `SKILL.md` files shrink to ~20-line **thin triggers**: version gate → read
  the playbook → follow it.
- Borrow the docs-first patterns proven by `repository-harness`:
  - `AGENTS.md` stable shim as the canonical agent entrypoint
  - `CONTEXT_RULES` (which stage reads which doc — no over-reading)
  - request-class authority (read-only requests must not mutate harness state)

Goal: any agent (Codex, Cursor, Claude Code) can operate the chain by reading
docs + calling `zharness` — Claude Code skills become one convenient trigger
layer among many, not the only way in.

## Decisions made during brainstorm

1. **Docs canonical home: embedded in the binary** — version-locked with the
   CLI; one source of truth; no #24-style doc/code drift. Rejected: installer
   copy from the skills repo (N drifting copies), shim-pointing-at-repo
   (breaks offline/private).
2. **Scope: 6 spine skills** (brainstorm, to-plan, work, check, handoff,
   watzup). `git`/`interview` stay minimal-integration as the entity mapping
   already designates. Rejected: all 8 (embedding commit standards / gh guides
   in a Go binary is scope bloat).
3. **Staleness: version-stamp + drift detection** — docs written with a
   `docs_version`; `resume` reports `stale_docs` drift when behind the CLI,
   recovery `zharness init --refresh-docs`. Rejected: write-once-at-init
   (silent drift on CLI upgrade — the exact #24 lesson).
4. **Proof: pilot with a second agent** — at least one lifecycle pass
   (intake → trace → check) driven by a non-Claude agent reading only docs +
   CLI. Rejected: structural review only (weak proof).
5. **Lane: high-risk** — rewrites the daily-driver chain and its public
   contract (others install these skills).

Rejected architecture alternative: docs-first without thinning the skills
(skills keep full content marked "derived") — two sources of truth, guaranteed
drift, worst of both worlds.
