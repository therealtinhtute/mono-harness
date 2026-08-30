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
| Know what outside evidence says about documenting a repo for agents | [`docs/research/`](research) |
| See what is being built right now | the one plan under [`docs/plans/active/`](plans/active) — empty means nothing is in flight |
| See what was built before | [`docs/plans/completed/`](plans/completed), most recently [`absorb-encode-protocol.md`](plans/completed/absorb-encode-protocol.md) |
| Look up a CLI command, flag, or table | [`cli/docs/CONTRACT.md`](../cli/docs/CONTRACT.md) |

## Ownership

Three classes, and the class determines who is allowed to edit the file.

- **managed** — shipped in the installer's embedded doc set (`cli/docs/embedded/`), projected into `docs/` by the managed-set installer and hash-tracked under `.zharness/base/` for the updater's three-way merge. Edit the embedded source and cut a release; a local edit is staged as an update conflict, never silently overwritten.
- **authored** — written by hand, never embedded, never regenerated. The consumer owns the content; repository scripts (not a binary verdict) watch whether any authored markdown remains, and semantic correctness belongs to the author or external tooling.
- **scaffold-once** — written by the managed-set installer only when absent, then owned by the consumer. The binary never refreshes, overwrites, or deletes one. In a consumer repository the class covers `docs/README.md`, `docs/decisions/README.md`, and `docs/decisions/templates/decision.md`; in this repository those three paths were authored by hand before the scaffold existed, so they are listed below as authored.

| Path | Class | Notes |
|---|---|---|
| `docs/WORKFLOW.md` | managed | stage router; names the one playbook to read |
| `docs/playbooks/` | managed | 6 stage playbooks; edit `cli/docs/embedded/playbooks/` instead. `git` is absent by design — it owns no harness entity, so its procedure lives at `skills/workflow/git/references/workflow.md` |
| `docs/README.md` | authored | this page |
| `docs/ARCHITECTURE.md` | authored | how the system works |
| `docs/decisions/` | authored | numbered ADRs, an index, and `templates/decision.md` to copy for the next one |
| `docs/plans/` | authored | initiative records; `active/` is the live plan, `completed/` is history. Sessions append by hand per the stage playbooks; nothing generates these files |
| `docs/prompt-engineering-principles.md` | authored | required reading before editing any `SKILL.md` or rule |
| `docs/workflow-harness/` | authored | legacy-adoption guide |
| `docs/audit/` | authored | findings that requirements cite as authority |
| `docs/research/` | authored | external-literature evidence that requirements cite as authority; describes the outside world, not this repository |
| `docs/PROJECT.md` | authored | identity; scaffold-once in consumer repos |
| `docs/patterns/` | authored | how to encode an accepted rule as a native guard |
| `docs/templates/` | authored | copy-from templates (not the installer `templates/`) |
| `docs/memory/` | authored | optional session memory files |
| `docs/references/` | authored | frozen snapshots (including `zharness-v015/`) |

An existing path under `docs/` that is missing from this table is a defect in this table.

## Not under `docs/`

| Path | What it is |
|---|---|
| `cli/docs/` | the CLI's contract reference — authored; `cli/docs/embedded/` holds the installer's managed doc set |
| `skills/` | the installable agent skills; each has its own `SKILL.md` |
| `rules/` | source for the global rules installed into `~/.claude/rules/` |
| (legacy) a per-machine derived index | removed from the architecture in v0.15 — archive: v0.15 section of CHANGELOG.md |
| `.kit/` | per-machine scratch — cache, conflicts, logs; fully gitignored |
