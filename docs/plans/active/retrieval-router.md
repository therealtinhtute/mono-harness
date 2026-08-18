---
id: 01M0AP2DKHY99R9CXQNQBA8WR1
type: plan
intake_id: 01M0AP2JB9WVAJETZHV7MZ4KS8
lane: normal
status: active
created: 2026-08-18
updated: 2026-08-18
---

# Plan: Retrieval router for durable memory (P6)

## Outcome
- result: `zharness memory query` gains keyword/relevance ranking over the `memories` index P5 established, so agents can surface the most relevant memory entries instead of only exact type/scope/plan_id filters. Semantic/embedding search is explicitly deferred to a later, separate initiative — this phase ships the keyword half of the hybrid approach only.
- success_signals:
  - `zharness memory query` accepts a text-relevance mode that ranks matching entries by keyword match against stored content, returning results ordered by relevance score (not just `created_at DESC`).
  - Ranking is computed entirely in SQLite/Go — no new external dependency, no vector store, no embedding model call.
  - No spine playbook (`brainstorm`, `to-plan`, `work`, `check`, `git`, `handoff`, `watzup`) gains a new required step — retrieval stays CLI opt-in, matching P5's own boundary.
  - Retrieval reads only the `memories` table — no new read path touches `traces`, `decisions`, `checks`, or any other table as a retrieval source.
  - `cd cli && go test ./...` passes at the phase boundary.

## Authority and Requirements
- authority:
  - `docs/plans/completed/durable-memory.md` (P5) — the storage layer this phase builds retrieval on top of; P5's NG1 explicitly named "P6 (retrieval-router)" as the next initiative for ranking/search/routing, reserved rather than implemented there.
  - Owner decision, this session (AskUserQuestion, three questions): retrieval mechanism is hybrid — keyword/filter scoring now, embeddings deferred to a future separate initiative; actor surface is CLI opt-in only, no spine playbook auto-injection; routing scope is limited to the `memories` table, no cross-table sources.
- requirements:
  - R1 [accepted]: `zharness memory query` gains a keyword/text-relevance scoring mode that ranks matching entries by relevance against stored `type`/`summary`/body content, computed in SQLite/Go with no new external dependency. | source: owner decision (AskUserQuestion, P6 retrieval mechanism)
  - R2 [accepted]: Retrieval remains CLI opt-in only — no spine playbook gains a new required step calling memory retrieval; the P5 NG3 boundary ("no mandatory new step added to any spine playbook") continues to hold for P6. | source: owner decision (AskUserQuestion, P6 integration surface)
  - R3 [accepted]: Retrieval ranks/routes exclusively over the `memories` table — no new read path touches `traces`, `decisions`, `checks`, or other tables as a retrieval source. | source: owner decision (AskUserQuestion, P6 routing scope)
  - R4 [accepted]: No embeddings, vector store, or semantic-search dependency is introduced in this phase; that capability is explicitly deferred to a future, separate initiative if ever pursued. | source: owner decision (AskUserQuestion, P6 retrieval mechanism)
  - R5 [accepted]: `zharness memory query`'s existing exact-filter behavior (`--type`, `--scope`, `--plan-id`) is preserved unchanged; relevance ranking is an additive mode, not a replacement for exact filtering. | source: durable-memory.md R3 (direct read-back by ID/filter must keep working)

## Non-goals
- NG1: No semantic/embeddings-based search, vector store, or embedding-model dependency — reserved for a future, separate initiative, not this phase (R4).
- NG2: No spine playbook changes and no automatic memory retrieval injected into any workflow skill — retrieval stays CLI opt-in, matching P5's NG3 pattern (R2).
- NG3: No cross-table routing — retrieval does not treat `traces`, `decisions`, or `checks` as retrieval sources; scope stays the `memories` table alone (R3).
- NG4: No new external interface (HTTP/API) for retrieval — CLI only, consistent with P5's NG2.

## Approach and Risks
- approach: Add a `--keywords` flag to the existing `zharness memory query` command. When absent, behavior is byte-for-byte the current P5 behavior (R5): `--type` required, results ordered `created_at DESC, id DESC`. When `--keywords "<terms>"` is present, `--type` becomes optional and the command switches to a ranked mode: candidates are pulled via the same existing filtered SELECT (optionally narrowed by `--type`/`--scope`/`--plan-id`), each candidate's stored `type` field plus its markdown body (read via the same file-read path `MemoryGet` already uses) is scored in Go by case-insensitive keyword-token match count against the query terms, zero-score candidates are dropped, and the rest are ordered by score DESC then `created_at DESC` as a tiebreak. All scoring logic lives in a new `MemoryQueryRanked` function in `cli/internal/application/memory.go`, reusing `CreateMemory`/`MemoryGet`'s existing markdown-read plumbing rather than adding a second file-access path.
  - rejected: SQLite FTS5 virtual table (e.g. `memories_fts` as an external-content table over `memories`, ranked via `bm25()`). Verified available in the vendored `modernc.org/sqlite v1.54.0` driver (`CREATE VIRTUAL TABLE t USING fts5(body)` succeeds), so it would not add a new external dependency and would give higher-quality ranking than a Go substring/token count. Rejected for this phase because it requires a new migration, a shadow index kept in sync with `memories` (via triggers or a duplicated rebuild-path pass), and doubles the surface `db rebuild` must reconstruct — complexity not justified for "keyword now, embeddings later" (R1/R4) when a Go-side scoring pass satisfies R1 with zero schema change. Revisit only if keyword-mode ranking quality proves insufficient in practice.
  - rejected: a separate `zharness memory search` command instead of extending `memory query`. Rejected because it would fork the read path in two places for the same table, and R5 already frames ranking as "an additive mode" of the existing command, not a new one.
- constraints:
  - No schema migration — `MemoryQueryRanked` reads the existing `memories` table and `docs/memory/*.md` files only; no new column, table, or index.
  - `memory query --keywords` must stay a read-only command (uses `infrastructure.OpenReadOnly`, like today's `memory get`/`memory query`) and must not create WAL/SHM sidecars, matching the existing `TestInspectionCommandsDoNotCreateWALSidecars` contract.
  - `db rebuild` is unaffected — ranking is computed at query time from data `rebuildMemoriesFromMarkdown` already reconstructs; no new rebuild pass is needed.
- risks:
  - risk: reading every candidate's markdown body at query time (to score it) is an O(n) file-read per query, unbounded by index count. | mitigation: candidates are first narrowed by any supplied `--type`/`--scope`/`--plan-id` filters before body-scoring runs, matching R5's existing filter semantics; full-corpus unfiltered keyword scans are an accepted cost at this phase's expected memory-entry volumes (tens to low hundreds of entries), not thousands.
  - risk: naive substring/token scoring can rank a keyword appearing in noise (e.g. a common word inside an unrelated body) above a more relevant but differently-worded entry. | mitigation: explicitly out of scope for this phase — R4 defers semantic ranking to a future initiative; documented as a known limitation, not a bug, in `cli/docs/CONTRACT.md`.

## Phases and Verification
<!-- Phase and task definitions are immutable after to-plan. Do not add task status fields. Append-only Progress is the sole task execution-status source. Only each phase lifecycle status changes to mirror DB transitions: to-plan=planned; work after run create=in-progress; clean durable check=checked; closing handoff=done. Each planned phase records phase_slug, story_id, status, goal, depends_on, waves, tasks, and checks. -->
- planning_status: planned
- phases:
  - phase_slug: p6-retrieval-router
    story_id: 01M0APETYJAF2BPA6NBVT1B2VZ
    status: planned
    goal: Add keyword/relevance ranking to `zharness memory query` over the existing `memories` index, CLI-only, no new table or dependency.
    depends_on: none
    waves:
      - wave: 1
        title: Ranked query — application layer + CLI flag
        tasks:
          - task: Add `MemoryQueryRanked(db *sql.DB, keywords, memType, scope, planID string) ([]MemoryScoredView, error)` to `cli/internal/application/memory.go`. Reuse the existing filtered-candidate SELECT and `MemoryGet`'s file-read path; score each candidate by case-insensitive keyword-token match count against `type` + body; drop zero-score candidates; sort by score DESC, `created_at` DESC.
            touches: cli/internal/application/memory.go
            avoids: cli/internal/infrastructure/migrations.go (no schema change per constraints)
            expected_output: new exported function + a `MemoryScoredView` (or equivalent) type carrying the existing view fields plus a `Score` field
          - task: Wire a `--keywords` flag onto the `memory query` cobra command in `cli/internal/interfaces/memory.go`. When `--keywords` is non-empty, make `--type` optional and call `MemoryQueryRanked`; when empty, existing `runMemoryQuery`/`MemoryQuery` behavior is unchanged (R5).
            touches: cli/internal/interfaces/memory.go
            avoids: cli/internal/interfaces/repository_lock.go (read-only command, no exclusivity entry needed — matches existing `memory query`/`memory get` exclusion)
            expected_output: `zharness memory query --keywords "<terms>" --json` returns entries ordered by relevance
        checks:
          - cd cli && go test ./internal/application/... ./internal/interfaces/... -run Memory
      - wave: 2
        title: Coverage, read-only contract, and CONTRACT.md
        depends_on: wave 1
        tasks:
          - task: Add unit tests to `cli/internal/application/memory_test.go` covering ranked ordering (multiple matches score higher), zero-match exclusion, `created_at DESC` tiebreak on equal scores, and combined `--keywords` + `--type`/`--scope`/`--plan-id` filtering.
            touches: cli/internal/application/memory_test.go
            avoids: none
            expected_output: new `TestMemoryQueryRanked*` test functions, all passing
          - task: Add a `memory query --keywords` row to the command table in `cli/internal/interfaces/read_only_commands_test.go`'s `TestInspectionCommandsDoNotCreateWALSidecars`, proving the ranked path stays read-only (no WAL/SHM sidecars).
            touches: cli/internal/interfaces/read_only_commands_test.go
            avoids: none
            expected_output: new table row passes alongside existing `memory get`/`memory query` rows
          - task: Document `memory add`/`memory get`/`memory query` (including the new `--keywords` flag and its scoring-limitation note from the Risks section) in `cli/docs/CONTRACT.md`. Closes the gap identified during this phase's planning — P5 shipped these commands without a CONTRACT.md entry.
            touches: cli/docs/CONTRACT.md
            avoids: none
            expected_output: CONTRACT.md documents all three memory subcommands and their flags
        checks:
          - cd cli && go test ./...
          - bash scripts/verify-doc-links.sh (informational only — 16 pre-existing failures inherited from before this phase are not this phase's blocker; do not treat new failures introduced by this phase's own edits as acceptable)

## Progress
<!-- Append-only durable entries record timestamp, phase, wave, task, task_status, run_id, trace_id, exact verification/result, and changed surfaces or blocker. -->
- none

## Decisions
<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->
- none

## Validation
<!-- Append-only durable entries record timestamp, phase, exact command/result/output, run_id, check_id, verdict, and proof_gaps. -->
- none

## Current State and Next Action
- active_phase: p6-retrieval-router
- lifecycle_status: planned
- latest_run_id: none
- latest_trace_ids: []
- latest_check_id: none
- latest_handoff_id: none
- blockers: none
- open_items: [cli/docs/CONTRACT.md does not yet document memory add/get/query — closed by phase p6-retrieval-router wave 2]
- exact_next_action: work full phase p6-retrieval-router
