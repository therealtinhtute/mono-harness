# AGENTS — Entrypoint

You are operating this repository's workflow harness via the `zharness` CLI. This file is the entrypoint; it does not contain workflow logic itself — it points you at the docs that do.

## First, every time

1. Run `zharness --version`. A `dev` build always passes. Otherwise, if the binary is missing or reports a version below `0.1.0` (this doc set's minimum, `MIN_ZHARNESS_VERSION`), stop and tell the user: `zharness not found or out of date — run: bash scripts/install-zharness.sh`. A future CLI release may raise this floor and stamp it into each doc's `docs_version` instead of this fixed number — if `zharness resume` reports `stale_docs` drift, trust that over this static value.
2. Classify the request as read-only or change per `AUTHORITY.md`, before running any command beyond `--version`/`--help`.
3. Look up your stage in `CONTEXT_RULES.md` to see exactly which docs to read — no more, no fewer.

## Lifecycle

Six stages, in order, each with one playbook under `playbooks/`:

1. **brainstorm** — turn a raw idea or request into a locked `SPEC.md`. → `playbooks/brainstorm.md`
2. **to-plan** — derive `ROADMAP.md` and per-phase plans from the locked spec. → `playbooks/to-plan.md`
3. **work** — execute the active phase wave-by-wave, verifying each task. → `playbooks/work.md`
4. **check** — gate a phase's diff with a deterministic verdict. → `playbooks/check.md`
5. **handoff** — close out a session, recording a resumable state snapshot. → `playbooks/handoff.md`
6. **watzup** — session-start recap of position, drift, and next action (read-only). → `playbooks/watzup.md`

Enter the chain at `watzup` when resuming existing work, or at `brainstorm` when starting something new. `work` routes into `check` as its own phase gate; `check` and `handoff` do not invoke each other.

Two commands sit outside this spine and carry no dedicated playbook: a plain commit/push request, and a pre-planning clarification pass — handle both with judgment and the general read-only/change rules in `AUTHORITY.md`.

## Reference docs

- `AUTHORITY.md` — read-only vs change request classes; which commands each may run
- `CONTEXT_RULES.md` — per-stage doc scope, so you never over-read
- `playbooks/*.md` — one per stage: purpose, preconditions, exact commands, artifacts, exit conditions

Every command in every playbook is a real `zharness` subcommand, verified against the shipped CLI. If a playbook step and a live `--help` output disagree, trust `--help` and report the mismatch — do not guess.
