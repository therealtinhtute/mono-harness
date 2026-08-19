# Architecture

How the harness actually works, and why it is shaped this way. Decisions that are expensive to reverse have their own records under `docs/decisions/`; this document describes the running system.

## The one idea

Committed markdown is the truth. Everything else is derived from it and can be thrown away.

That single constraint explains most of the design. The SQLite database is an index, not a store. The projected playbooks are a copy, not a source. The `.kit/` directory is scratch, not state. If you deleted every derived artifact and re-ran the CLI, you would get the same system back, because the only irreplaceable bytes are the ones git is already tracking.

## The four surfaces

```
skills/workflow/*/SKILL.md   thin trigger  ->  runs `zharness preflight <stage> --json`
        |
        v
docs/playbooks/<stage>.md    the procedure ->  projected from the binary, never hand-edited
        |
        v
zharness (Go binary)         the guardrail ->  writes markdown, derives the DB row
        |
        v
docs/plans/active/*.md       the truth     ->  committed; harness.db is rebuilt from it
```

A spine skill contains no operating logic. It version-gates the binary, calls `preflight`, and reads whatever playbook path comes back. The logic lives in the playbook, which is a file — so any agent that can read a file and run a CLI executes the same lifecycle, with no dependence on a particular vendor's skill format.

## Markdown first, then the row

Every lifecycle write goes to markdown before it goes to the database, and the ordering is deliberate rather than incidental.

`cli/internal/application/plan_write.go:36` states the rule directly: markdown is the write target and the DB row is derived from it. A SQL transaction cannot span a file write, so the guarantee is one-directional by construction — a failed markdown write leaves zero rows behind it (`cli/internal/application/trace.go:65`), while a DB failure after a successful markdown write leaves the markdown line as the durable fact, recoverable by rebuild. The error text says exactly that: *"plan markdown recorded, but db write failed"*.

The consequence worth internalizing: **a field that lands only in the database and never in markdown disappears on the next rebuild.** That is the test any new mutating command has to pass.

## `harness.db` is an index

`cli/internal/application/rebuild.go:73` reconstructs the entire database from committed plan markdown alone. The file is gitignored, and losing it costs nothing.

Rebuild is honest about what it cannot recover, which is the useful part of reading its doc comment. Trace and decision entries come back with freshly minted IDs — content preserved, identity not — because nothing else in the schema references them. A run is reconstructed only when a Validation entry backreferences it, since that is the one entry shape carrying the run's story slug, and `runs.story_slug` is `NOT NULL`. A run mentioned only by a trace and never checked has no recoverable slug and is dropped, with dependent rows getting a `NULL` run_id rather than a dangling one.

Three derived indexes share one column shape — path, SHA256, updated_at:

| Table | Indexes | Migration |
|---|---|---|
| `managed_docs` | the projected doc set under `docs/` | `cli/internal/infrastructure/migrations.go:131` |
| `plan_index` | active plans under `docs/plans/` | `cli/internal/infrastructure/migrations.go:218` |
| `memories` | memory entries under `docs/memory/` | `cli/internal/infrastructure/migrations.go:237` |

Staleness is always a hash comparison, never a timestamp guess.

## Managed docs: projection with conflict staging

The binary carries two embedded filesystems, and the split is the whole mechanism.

`cli/docs/embedded/embed.go:9` declares `FS` — `AGENTS.md`, `WORKFLOW.md`, and `playbooks/`. This set is projected into the consumer repo's `docs/` at `init` (`cli/internal/application/init.go:33`) and hash-tracked. `cli/docs/embedded/embed.go:18` declares `Templates` separately, on purpose: templates are never projected, only emitted when `zharness scaffold` asks for one.

`SyncManagedDocs` (`cli/internal/application/managed_docs.go:43`) compares three values per file — the recorded hash, the hash on disk, and the hash of the embedded content. Unchanged files are refreshed silently. A file the consumer edited locally is **not** overwritten: it is staged as a conflict under `.kit/conflicts/` and the sync reports it (`cli/internal/application/managed_docs.go:56`), unless `--force-docs` is passed.

One property matters for anything that removes a file from the embedded set. `planManagedDocActions` (`cli/internal/application/managed_docs.go:107`) drives its iteration by walking the embedded filesystem, not by reading the `managed_docs` table. A row whose path is no longer embedded is simply never visited — it does not error, and no local file is deleted. There is no prune path, and nothing in this package removes anything: deprojecting a doc leaves an inert orphan in already-initialized repos rather than breaking them.

## Exactly one active plan

`ResolveActivePlan` (`cli/internal/application/plan_resolve.go:73`) is the single entry point for obtaining the active plan. Six call sites use it and nothing bypasses it: `cli/internal/application/plan_query.go:68`, `cli/internal/application/plan_write.go:22`, `cli/internal/application/resume.go:156`, `cli/internal/application/plan_lifecycle.go:53`, `cli/internal/application/plan_lifecycle.go:94`, `cli/internal/application/validate.go:446`.

It never returns a bare error for the two interesting cases. Zero plans yields `Stop{Code: "none"}` with the recovery `brainstorm lock`. Two or more yields `Stop{Code: "ambiguous"}` with a bounded candidate list built from frontmatter previews — plan bodies are never read to disambiguate, because reading them is the unbounded cost the contract exists to prevent.

The enforced invariant is *at most* one active plan. Zero is a legitimate idle state and produces no validation finding. See [`docs/decisions/0002-single-active-plan-resolver.md`](decisions/0002-single-active-plan-resolver.md).

## Preflight is the only entry point

Every spine stage begins with `zharness preflight <stage> --json`. It resolves readiness without mutating anything and returns the playbook path, the lifecycle position, and any drift.

When the database or the projected docs are missing, preflight returns a `stop` whose recovery is `zharness init` (`cli/internal/application/preflight.go:76`). Skills do not self-scaffold; they run the recovery they are handed. The stage-to-playbook mapping lives in one table, `preflightPlaybooks` at `cli/internal/interfaces/preflight.go:30` — stages absent from it own no harness entity and get no playbook back.

## What lives where

| Path | Nature |
|---|---|
| `docs/plans/` | authoritative — the initiative record |
| `docs/decisions/` | authoritative — ADRs |
| `docs/playbooks/`, `docs/WORKFLOW.md` | projected from the binary; edit `cli/docs/embedded/` instead |
| `harness.db` | derived; gitignored; rebuild with `zharness db rebuild --yes` |
| `.kit/` | per-machine scratch — cache, conflicts, logs; fully gitignored |

`docs/README.md` carries the complete ownership table.
