---
id: 01M0AKX48PFDAMVQ9WQWVP45RX
type: plan
intake_id: 01M0AKX8NZ58GZ9RBQ1KX0TZYR
lane: normal
status: completed
created: 2026-08-18
updated: 2026-08-18
---

# Plan: Durable cross-session agent memory (P5)

## Outcome
- result: zharness gains a durable, queryable memory layer so agents running the workflow pipeline (brainstorm/to-plan/work/check/git/handoff, watzup) retain institutional context — decisions, gotchas, prior outcomes — across sessions instead of relosing it each run. Storage follows this repo's own just-established architecture: markdown is the write target and source of truth, harness.db holds a derived, rebuildable index — never the reverse.
- success_signals:
  - `zharness memory add` writes a markdown entry under `docs/memory/` first; the DB row is derived from the written file, matching the P2 write-ordering pattern (markdown-first, DB derived).
  - `zharness memory query`/`zharness memory get` reads memory back through the CLI without any spine skill needing to grep `docs/memory/` by hand.
  - `zharness db rebuild` reconstructs the memory index from committed `docs/memory/*.md` content alone, consistent with R9's plan_index precedent — no memory content lives only in harness.db.
  - The six spine playbooks (`brainstorm`, `to-plan`, `work`, `check`, `git`, `handoff`) gain no new required step for this phase alone — memory writes are opt-in via explicit CLI calls, not a mandatory ceremony addition.
  - `cd cli && go test ./...` and `bash scripts/verify-doc-links.sh` pass at the phase boundary.

## Authority and Requirements
- authority:
  - `docs/plans/completed/harness-markdown-truth.md` — the just-completed initiative this plan follows architecturally: markdown-first writes, SQLite as a rebuildable derived index (R8, R9, R10 of that plan).
  - `cli/internal/infrastructure/migrations.go` — the `managed_docs`/`plan_index` SHA256 freshness pattern, the precedent this phase's memory index reuses rather than inventing a second staleness mechanism.
  - Owner decision, this session (AskUserQuestion, three questions): Outcome is cross-session memory on harness.db; actors are the zharness CLI and workflow skills only, no new external interface; this phase is storage/write-path only, retrieval/ranking is a separate later initiative (P6, retrieval-router).
  - Owner decision, this session: this plan is a new, independent initiative — not phases appended to `harness-markdown-truth.md` — because durable-memory's scope does not fit that plan's own Outcome or Non-goals (NG6 there explicitly excludes changes to the pipeline shape beyond version bumps).
- requirements:
  - R1 [accepted]: Memory entries are committed markdown files under `docs/memory/`, one file per entry with frontmatter (`id`, `type`, `scope`, `created`); harness.db stores a derived `memories` index (path, sha256, type, scope, created_at) mirroring `plan_index`'s column shape. No memory content is ever written to harness.db only. | source: harness-markdown-truth R8/R9, owner decision
  - R2 [accepted]: Only `zharness memory {add,get,query}` (new CLI subcommands) and the six spine skills that call them read or write memory; no HTTP endpoint, no new external query surface. | source: owner decision (AskUserQuestion, P5 actors)
  - R3 [accepted]: This phase implements the durable write path (`zharness memory add`) and a direct read-back by ID or path (`zharness memory get`); it does not implement ranking, semantic search, or routing across multiple memory entries — that is exclusively P6 (retrieval-router)'s scope, a separate initiative locked after this one. | source: owner decision (AskUserQuestion, P5 non-goals)
  - R4 [accepted]: `zharness db rebuild` reconstructs the `memories` index from committed `docs/memory/*.md` content alone, with no read of any non-committed state, consistent with `harness-markdown-truth` R10. | source: harness-markdown-truth R10
  - R5 [accepted]: Memory entries carry a `scope` field distinguishing plan-scoped (`plan_id` set, tied to one initiative) from repo-global (`plan_id` empty) — to-plan resolves the exact write/query CLI flag shape; this requirement only fixes that the distinction must exist, not its final syntax. | source: owner decision, this session (memory tied to specific initiatives vs. durable across all of them are both real use cases named in the original P5 progress notes)

## Non-goals
- NG1: No ranking, semantic search, embeddings, or multi-source routing — reserved for P6 (retrieval-router), a separate plan locked after this one completes.
- NG2: No new external interface (HTTP/API) for memory access — CLI and the six spine skills only (R2).
- NG3: No mandatory new step added to any spine playbook — memory writes are opt-in this phase; whether/how playbooks should call it is a decision for a later phase or initiative, not this one.
- NG4: No migration of the global `~/.claude` auto-memory system (user's personal cross-project memory) into this repo-scoped store — they remain separate, unrelated systems.

## Approach and Risks
- approach: Implement the memory layer using the same markdown-first, DB-derived pattern P2/P3 of `harness-markdown-truth` established for traces/decisions/checks — a new `docs/memory/` directory of git-committed entry files, a `memories` migration reusing the `plan_index`/`managed_docs` SHA256 shape, and three new CLI subcommands (`memory add`, `memory get`, `memory query`) that write markdown first and derive the index row from the written file. One phase, three waves, ordered by real dependency: schema+write path, then the read path (depends on the schema existing), then rebuild integration and command registration (depends on both).
- rejected alternatives:
  - Storing memory content directly in harness.db with no markdown backing — rejected: reintroduces the dual-source-of-truth problem `harness-markdown-truth` P2/P3 just eliminated for traces/decisions/checks (R1).
  - Extending the existing `decisions` table with a `memory` type flag instead of a new table — rejected: `## Decisions` is scoped to phase/task rationale by its own doc comment ("record timestamp, phase/task, decision, and rationale"); conflating memory with it would break that meaning for every existing consumer.
- constraints:
  - Memory writes are opt-in CLI calls this phase; no playbook is required to call them (NG3).
  - `docs/memory/` entries are git-committed markdown, same as plan files — never gitignored.
- risks:
  - R-A: A new top-level `docs/memory/` directory could interact with `scripts/verify-doc-links.sh`'s allowlist scanning (first path segment must be in `skills docs rules cli setup references`). Mitigation: confirm `docs/` is already covered before adding any cross-references inside memory entries, and re-run the script in wave 3's check.
  - R-B: `zharness memory add`/`query`/`get` need correct registration in `repository_lock.go` — `add` takes the write lock, `get`/`query` stay read-only — matching how `plan complete`/`plan abandon` were registered in P0 wave 4. Mitigation: add all three explicitly and cover with a `read_only_commands_test.go`-style assertion in wave 3.

## Phases and Verification
<!-- Phase and task definitions are immutable after to-plan. Do not add task status fields. Append-only Progress is the sole task execution-status source. Only each phase lifecycle status changes to mirror DB transitions: to-plan=planned; work after run create=in-progress; clean durable check=checked; closing handoff=done. Each planned phase records phase_slug, story_id, status, goal, depends_on, waves, tasks, and checks. -->
- planning_status: planned
- phases:
  - phase_slug: p5-durable-memory
    story_id: 01M0AMA2R0TCFDR4BBT8WJK7GJ
    status: done
    goal: Add durable, cross-session agent memory — a markdown-first docs/memory/*.md write path with a derived memories index in harness.db, reconstructible via db rebuild.
    depends_on: none
    waves:
      - wave: 1 — schema and write path
        tasks:
          - task: Add a `memories` migration to `cli/internal/infrastructure/migrations.go` with `id TEXT PRIMARY KEY`, `path TEXT UNIQUE NOT NULL`, `type TEXT NOT NULL`, `scope TEXT NOT NULL`, `plan_id TEXT`, `sha256 TEXT NOT NULL`, `created_at TEXT NOT NULL`, mirroring the `plan_index` column shape (R1).
          - task: Implement `zharness memory add --type {type} --scope {plan|global} [--plan-id {id}] --summary "..." --json`, writing `docs/memory/{id}.md` first (frontmatter: id/type/scope/plan_id/created) and deriving the `memories` DB row from the written file — markdown first, DB derived, matching P2's write-ordering pattern (R1).
        checks:
          - check: `cd cli && go test ./internal/application/... -run Memory`
      - wave: 2 — read path
        tasks:
          - task: Implement `zharness memory get --id {id} --json`, returning the parsed markdown entry's frontmatter plus body.
          - task: Implement `zharness memory query --type {type} [--scope {plan|global}] [--plan-id {id}] --json`, listing matching entries (path plus metadata only) from the `memories` index — an index/list operation, not ranking or cross-source routing, which stays exclusively P6 (retrieval-router)'s scope (R3, NG1).
        checks:
          - check: `cd cli && go test ./internal/application/... -run Memory`
      - wave: 3 — rebuild integration and registration
        tasks:
          - task: Wire `zharness db rebuild` to reconstruct the `memories` index from committed `docs/memory/*.md` content alone, with no read of non-committed state (R4).
          - task: Add a test wiping `harness.db`, rebuilding from a fixture `docs/memory/` directory, and asserting the index matches the pre-wipe state.
          - task: Register `memory add`, `memory get`, `memory query` in `cli/internal/interfaces/root.go`; add `memory add` to the write-lock set and `memory get`/`memory query` to the read-only set in `repository_lock.go` (R-B).
        checks:
          - check: `cd cli && go test ./... && cd .. && bash scripts/verify-doc-links.sh`

## Progress
<!-- Append-only durable entries record timestamp, phase, wave, task, task_status, run_id, trace_id, exact verification/result, and changed surfaces or blocker. -->
- `2026-08-18T14:31:08Z` — wave 1. run: `01M0AMGMDF93MWSDYQEXEK4RTY`. summary: Phase p5-durable-memory started (in-progress).
- `2026-08-18T14:39:18Z` — wave 1, task Add memories migration to migrations.go (Version 12, plan_index column shape). task_status: `DONE`. run: `01M0AMGMDF93MWSDYQEXEK4RTY`. summary: 0012_memories added; migrations_test.go and SCHEMA.md updated to match (schemaVersion 12, memories table + memory changeset entity row)..
- `2026-08-18T14:39:18Z` — wave 1, task Implement zharness memory add write path (application layer). task_status: `DONE`. run: `01M0AMGMDF93MWSDYQEXEK4RTY`. summary: domain.Memory + Validate(), application.CreateMemory writes docs/memory/{id}.md via writeManagedFile before inserting the derived memories row (managedDocSHA256 for sha256), mirroring CreateTrace ordering. CLI wiring (interfaces/memory.go, root.go, repository_lock.go) deferred to wave 3 per the plan approach note (rebuild integration AND command registration together)..
- `2026-08-18T14:39:22Z` — wave 1. run: `01M0AMGMDF93MWSDYQEXEK4RTY`. summary: Wave 1 (schema and write path) complete: memories migration (0012), domain.Memory, application.CreateMemory markdown-first write path. cd cli && go test ./internal/application/... -run Memory passes (5/5); full cd cli && go test ./... passes..
- `2026-08-18T14:40:46Z` — wave 2, task Implement zharness memory get --id (application layer). task_status: `DONE`. run: `01M0AMGMDF93MWSDYQEXEK4RTY`. summary: application.MemoryGet resolves memories.path by id, reads docs/memory/{id}.md, and parses frontmatter (type/scope/plan_id/created) plus body via frontmatterPreview/frontmatterPreviewField (plan_resolve.go) — reads the markdown file itself, not the DB row, per R1 markdown-as-truth. unknown_memory_id ValidationError on a missing id..
- `2026-08-18T14:40:46Z` — wave 2, task Implement zharness memory query --type [--scope] [--plan-id] (application layer). task_status: `DONE`. run: `01M0AMGMDF93MWSDYQEXEK4RTY`. summary: application.MemoryQuery filters the memories index by required type plus optional scope/plan_id, returns path+metadata only (MemoryListView), no body/ranking — R3/NG1 boundary respected..
- `2026-08-18T14:40:50Z` — wave 2. run: `01M0AMGMDF93MWSDYQEXEK4RTY`. summary: Wave 2 (read path) complete: application.MemoryGet + application.MemoryQuery, 9/9 memory application tests pass (cd cli && go test ./internal/application/... -run Memory), full cd cli && go test ./... passes..
- `2026-08-18T14:47:42Z` — wave 3. run: `01M0AMGMDF93MWSDYQEXEK4RTY`. summary: wave 3 (rebuild integration and registration): rebuildMemoriesFromMarkdown wired into RebuildFromMarkdown, memory add/get/query registered in root.go, memory add added to exclusiveMutationCommandPaths, TestRebuildMemoriesFromMarkdownRoundTrip proves index round-trips after wipe+rebuild, read_only_commands_test.go extended with memory get/query cases proving no WAL sidecars on read paths. go build ./... and go test ./... pass; scripts/verify-doc-links.sh's 16 failures confirmed pre-existing (identical on stashed clean base, unrelated to this work). Phase p5-durable-memory (all 3 waves) complete..
- `2026-08-18T14:50:14Z` — handoff recorded. handoff: `01M0ANKZD520QM3TYT1DWTE67Y`. run: `01M0AMGMDF93MWSDYQEXEK4RTY`. check: `01M0ANJ9M8202M2M9RRBJZS8EP`. phase closed.

## Decisions
<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->
- `2026-08-18T14:50:30Z` — plan completed. rationale: every phase_slug is a done story.

## Validation
<!-- Append-only durable entries record timestamp, phase, exact command/result/output, run_id, check_id, verdict, and proof_gaps. -->
- `2026-08-18T14:49:19Z` — check. verdict: `APPROVE_WITH_REQUESTS`. check: `01M0ANJ9M8202M2M9RRBJZS8EP`. run: `01M0AMGMDF93MWSDYQEXEK4RTY`. phase: `p5-durable-memory`. judge: `same-session` (claude-sonnet-5).
  - `cd cli && go test ./...` → Validation entry 2026-08-18T14:52:00Z: all packages ok

## Current State and Next Action
- active_phase: none
- lifecycle_status: done
- latest_run_id: 01M0AMGMDF93MWSDYQEXEK4RTY
- latest_trace_ids: [01M0ANFAXXTDHX7PB6TPKKHPCG]
- latest_check_id: 01M0ANJ9M8202M2M9RRBJZS8EP
- latest_handoff_id: 01M0ANKZD520QM3TYT1DWTE67Y
- blockers: none
- open_items: none
- exact_next_action: zharness plan complete --json (final phase closure)
- exact_next_action: handoff record --close-phase p5-durable-memory, then zharness plan complete
