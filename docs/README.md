# Documentation

Start here. Every document in this repository is reachable from this page, and every path under `docs/` has exactly one owner.

## Where to go

| You want to | Read |
|---|---|
| Understand how the harness works | [`docs/ARCHITECTURE.md`](ARCHITECTURE.md) |
| Know why something is built the way it is | [`docs/decisions/README.md`](decisions/README.md) |
| Run a workflow stage | [`docs/WORKFLOW.md`](WORKFLOW.md), then the one playbook it names |
| Adopt the harness on an existing project | [`docs/workflow-harness/migration.md`](workflow-harness/migration.md) |
| Write or edit a skill or a rule | [`docs/prompt-engineering-principles.md`](prompt-engineering-principles.md) |
| See what is being built right now | the one plan under [`docs/plans/active/`](plans/active) — empty means nothing is in flight |
| See what was built before | [`docs/plans/completed/`](plans/completed), most recently [`docs-architecture.md`](plans/completed/docs-architecture.md) |
| Look up a CLI command, flag, or table | [`cli/docs/CONTRACT.md`](../cli/docs/CONTRACT.md), [`cli/docs/SCHEMA.md`](../cli/docs/SCHEMA.md), [`cli/docs/STATE.md`](../cli/docs/STATE.md) |

## Ownership

Three classes, and the class determines who is allowed to edit the file.

- **managed** — present in the binary's embedded doc set (`cli/docs/embedded/`), projected into `docs/` by `zharness init`, and hash-tracked in the `managed_docs` table. Edit the embedded source and cut a release; a local edit is staged as a conflict under `.kit/conflicts/`, never silently overwritten.
- **authored** — written by hand, never embedded, never regenerated. The binary neither creates nor touches these.
- **scaffold-once** — written by `zharness init` only when absent, then owned by the consumer. The binary never refreshes, overwrites, or deletes one. In a consumer repository the class covers `docs/README.md`, `docs/decisions/README.md`, and `docs/decisions/templates/decision.md`; in this repository those three paths were authored by hand before the scaffold existed, so they are listed below as authored.

| Path | Class | Notes |
|---|---|---|
| `docs/WORKFLOW.md` | managed | stage router; names the one playbook to read |
| `docs/playbooks/` | managed | 6 stage playbooks; edit `cli/docs/embedded/playbooks/` instead. `git` is absent by design — it owns no harness entity, so its procedure lives at `skills/workflow/git/references/workflow.md` |
| `docs/README.md` | authored | this page |
| `docs/ARCHITECTURE.md` | authored | how the system works |
| `docs/decisions/` | authored | numbered ADRs, an index, and `templates/decision.md` to copy for the next one |
| `docs/plans/` | authored | initiative records; `active/` is the live plan, `completed/` is history. The CLI appends to these but never creates them from an embedded source |
| `docs/prompt-engineering-principles.md` | authored | required reading before editing any `SKILL.md` or rule |
| `docs/workflow-harness/` | authored | legacy-adoption guide |
| `docs/audit/` | authored | findings that requirements cite as authority |

An existing path under `docs/` that is missing from this table is a defect in this table.

## Not under `docs/`

| Path | What it is |
|---|---|
| `cli/docs/` | the CLI's own contract, schema, and state reference — authored, and the source of the embedded set under `cli/docs/embedded/` |
| `skills/` | the installable agent skills; each has its own `SKILL.md` |
| `rules/` | source for the global rules installed into `~/.claude/rules/` |
| `harness.db` | derived index, gitignored, rebuilt with `zharness db rebuild --yes` |
| `.kit/` | per-machine scratch — cache, conflicts, logs; fully gitignored |
