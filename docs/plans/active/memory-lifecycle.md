---
id: 01M0WMQV8GFKNVY31TH8QRG74N
type: plan
intake_id: 01M0WMQV956AZRH913R7RZKF75
lane: normal
status: active
created: 2026-08-25
updated: 2026-08-25
---

# Plan: Memory lifecycle — supersede lineage, diacritic-fold retrieval, write-path gating

## Outcome
- result: the memory subsystem supports a correctable lifecycle — corrections supersede instead of duplicate, ranked retrieval survives Vietnamese diacritic variation, and writes are gated against near-duplicates — with markdown remaining the sole source of truth and zero new external dependencies.
- success_signals:
  - a superseded entry is invisible to default `memory query`/ranked views, restorable via `--include-superseded`, and its `superseded_by` chain is queryable end-to-end
  - the golden retrieval eval passes 10/10 fixed queries returning their target fixture inside the top 3, including mixed-diacritic pairs (e.g. "kiem tra" matching "kiểm tra")
  - `memory add` without `--force` refuses a near-duplicate of an existing entry, naming it (fixture-proven)
  - `zharness db rebuild` reproduces every supersede chain from committed markdown alone
  - the work playbook states when to record a memory and the redaction rule (never credentials/keys/tokens in bodies)

## Authority and Requirements
- authority:
  - this repository's 7-dimension agentic-memory audit (session 2026-08-25), grounded in `cli/internal/application/memory.go`, `cli/internal/domain/memory.go`, `cli/internal/infrastructure/migrations.go`
  - researched prior art: Zep/Graphiti non-lossy bi-temporal invalidation (arXiv 2501.13956 — invalidate by timestamp, preserve the edge so timelines stay reconstructable); mem0 additive-by-default writes with pre-insert dedup context lookup (docs.mem0.ai/how-it-works)
  - completed plans `docs/plans/completed/durable-memory.md` (P5, md-first ordering R1) and `docs/plans/completed/retrieval-router.md` (P6, keyword ranking scope)
- requirements:
  - R1 [accepted]: `zharness memory supersede --old-id {id} --new-id {id}` writes `superseded_by` + `superseded_at` into the old entry's frontmatter and mirrors a `superseded_by` column in the `memories` index; it refuses unknown IDs, self-supersession, and re-superseding an already-superseded entry with exit-1 user errors | source: Zep non-lossy invalidation pattern
  - R2 [accepted]: superseded entries are excluded from `memory query`, ranked query, and the preflight context packet by default; `--include-superseded` restores them; direct `memory get --id` still resolves any ID and reports its status | source: audit dim 5
  - R3 [accepted]: `memory add` runs a ranked match of the new summary against existing entries before writing; a best-hit score of ≥4 folded-token matches refuses with `similar_memory` listing the IDs unless `--force` is passed | source: mem0 pre-insert dedup gate
  - R4 [accepted]: the ranked-query tokenizer case-folds and strips Vietnamese diacritics on both query keywords and body tokens, so "kiem tra" matches "kiểm tra"; a golden fixture test proves ≥10 fixed queries hit their target within the top 3 including mixed-diacritic pairs | source: audit M3
  - R5 [accepted]: `db rebuild` reconstructs `superseded_by`/`superseded_at` from markdown frontmatter alone; the schema migration adding these columns bumps schema version once | source: P3 markdown truth
  - R6 [accepted]: `docs/playbooks/work.md` gains the trigger convention (record on: fact correction, durable lesson, owner preference) and the redaction rule (never store credentials, keys, or token values in memory bodies); the same two rules appear as comments in the scaffolded decision-template sibling `docs/memory/` README if one exists, else in work.md only | source: audit dims 1+6

## Non-goals
- NG1: vector/embedding storage and semantic search — deferred until corpus size justifies it
- NG2: BM25 scoring, graph storage, or graph traversal
- NG3: automatic write triggers or event hooks — triggers stay documented conventions invoked by agents
- NG4: retention/expiry policies and time-based or relevance decay
- NG5: cross-repository memory sharing or sync — memory remains per-repo under `docs/memory/`
- NG6: multi-user tenancy or a hosted API

## Approach and Risks
- approach: extend the existing markdown-first memory subsystem in place — one schema migration (0013) adds lineage columns to the derived `memories` index; a new `memory supersede` command mutates only markdown frontmatter plus that index; ranked-query tokenization gains a Vietnamese diacritic fold inside the existing Go scorer (hand-rolled precomposed-character map, no new module dependency); `CreateMemory` gains a pre-insert dedup check calling the existing ranked query; `db rebuild`'s frontmatter parser reads the new fields; work.md gains the conventions section.
- constraints:
  - zero new external module dependencies (NG1 posture); the diacritic fold must be a stdlib-only map
  - markdown stays sole source of truth: every index column must round-trip through `db rebuild`
  - the working tree currently carries the uncommitted missing-DB hint change-set in package `interfaces`; phase waves must start from a clean or committed tree so diffs never mix — commit that work first
- rejected alternatives:
  - embedding/vector backend (sqlite-vec) for retrieval — new dependency, unjustified at current corpus size (NG1)
  - Graphiti-style graph rebuild with LLM-driven invalidation — massive scope, needs an LLM in the write path (NG2)
  - soft-delete flag instead of lineage pointer — loses the correction chain the audit's dim 2 metadata requirement asks for
- risks:
  - dedup gate false positives block legitimate writes → mitigation: `--force` escape hatch with distinct error code naming the similar IDs
  - fold map too narrow (missed Vietnamese forms) → mitigation: golden eval fixtures include both precomposed and combining-mark forms; failures are visible in CI
  - supersede chain cycles (A→B→A) → mitigation: command refuses any old-id already carrying `superseded_by`; chains are strictly forward by construction
  - migration drift between index and markdown → mitigation: rebuild round-trip test is a required check in every phase wave touching the index

## Phases and Verification
<!-- Phase and task definitions are immutable after to-plan. Do not add task status fields. Append-only Progress is the sole task execution-status source. Only each phase lifecycle status changes to mirror DB transitions: to-plan=planned; work after run create=in-progress; clean durable check=checked; closing handoff=done. Each planned phase records phase_slug, story_id, status, goal, depends_on, waves, tasks, and checks. -->
- planning_status: planned
- phases:
  - phase_slug: supersede-lineage
    story_id: 01M0WPP9ASKQNEHJY4F22H6ENP
    status: planned
    goal: memory supersede command with lineage columns, default exclusion filtering, and rebuild-safe frontmatter round-trip
    depends_on: null
    requirements: [R1, R2, R5]
    allowed_surfaces: [cli/internal/application/memory.go, cli/internal/application/context.go, cli/internal/domain/, cli/internal/infrastructure/migrations.go, cli/internal/interfaces/memory.go]
    avoided_surfaces: [cli/internal/application/init.go scaffold bodies, embedded playbook sources]
    waves:
      - wave: 1
        tasks:
          - "Failing regression tests first: supersede happy path writes frontmatter+index and refuses unknown/self/re-supersede IDs (exit 1 codes unknown_memory_id / supersede_self / already_superseded); default query+ranked+context packet exclude superseded entries; --include-superseded restores them; db rebuild reproduces a fixture chain from markdown alone"
        verification: "cd cli && go test ./internal/application/ -run 'Supersede|MemoryRebuild' fails on unmodified code"
      - wave: 2
        tasks:
          - "Migration 0013: ALTER memories ADD superseded_by TEXT NULL, superseded_at TEXT NULL; bump schema version"
          - "Implement application.SupersedeMemory (md rewrite then index update, md-first ordering) + interfaces wiring behind `zharness memory supersede --old-id --new-id --json`"
          - "Filter superseded rows out of MemoryQuery/MemoryQueryRanked/BuildContextPacket unless IncludeSuperseded; MemoryGet reports status field"
          - "Extend rebuild's frontmatter parser to read superseded_by/superseded_at"
        verification: "focused suite green: cd cli && go test ./internal/application/ ./internal/interfaces/ ./internal/infrastructure/"
      - wave: 3
        tasks:
          - "Live CLI proof: add two memories, supersede one, show default query hides it, --include-superseded shows chain, zharness db rebuild --yes, re-query chain intact"
          - "Full gate: cd cli && go test ./... && bash scripts/verify-doc-links.sh"
        verification: "all commands exit 0; rebuild output lists memories count unchanged"
    checks:
      - "go test focused suites exit 0"
      - "live supersede + rebuild round-trip proof transcript recorded in Progress"
      - "verify-doc-links.sh 0 findings"
  - phase_slug: retrieval-quality
    story_id: 01M0WPP9B9C1N0PN4WQQN2BKKY
    status: planned
    goal: diacritic-fold tokenizer with golden hit@3 eval plus pre-insert dedup gate on memory add
    depends_on: supersede-lineage
    requirements: [R3, R4]
    allowed_surfaces: [cli/internal/application/memory.go, cli/internal/interfaces/memory.go]
    avoided_surfaces: [go.mod (no new modules), infrastructure/]
    waves:
      - wave: 1
        tasks:
          - "Golden eval fixture: ≥10 fixed queries over seeded memories incl. mixed-diacritic pairs (kiem tra↔kiểm tra, dong bo↔đồng bộ), asserting target inside top 3 — failing first"
          - "Failing dedup-gate tests: add without --force near-duplicate refuses similar_memory naming IDs; --force succeeds"
        verification: "cd cli && go test ./internal/application/ -run 'RankingEval|DedupGate' fails on unmodified code"
      - wave: 2
        tasks:
          - "Implement foldFold tokenizer normalization (case-fold + Vietnamese precomposed/combining-mark strip) applied to keywords and body tokens in MemoryQueryRanked scoring"
          - "Dedup gate in interfaces add path: rank summary tokens, best score ≥4 → refuse similar_memory unless --force"
        verification: "focused suites green including golden eval 10/10"
      - wave: 3
        tasks:
          - "Full gate: cd cli && go test ./... && bash scripts/verify-doc-links.sh"
        verification: "exit 0"
    checks:
      - "golden eval asserts hit@3 on all fixture queries, runs in go test"
      - "dedup refusal + --force override proven live via CLI transcript in Progress"
  - phase_slug: memory-conventions-docs
    story_id: 01M0WPP9BTW45Z0PW17HFFNY9E
    status: planned
    goal: work playbook documents memory trigger conventions and the body redaction rule
    depends_on: null
    requirements: [R6]
    allowed_surfaces: [docs/playbooks/work.md]
    avoided_surfaces: [cli/ entirely, docs/embedded/ (managed source syncs separately)]
    waves:
      - wave: 1
        tasks:
          - "Add 'Memory conventions' section to docs/playbooks/work.md: record on fact correction (via supersede), durable cross-session lesson, owner preference; redaction rule — never credentials/keys/token values in bodies; pointer to memory add/supersede/query usage"
        verification: "bash scripts/verify-doc-links.sh exits 0 with the new section's references resolving"
      - wave: 2
        tasks:
          - "Full gate: cd cli && go test ./... (unchanged-code sanity)"
        verification: "exit 0"
    checks:
      - "doc-link gate passes; section present in rendered playbook"

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
- active_phase: none
- lifecycle_status: not-planned
- latest_run_id: none
- latest_trace_ids: []
- latest_check_id: none
- latest_handoff_id: none
- blockers: none
- open_items: [commit the pending missing-DB hint change-set before run create so phase diffs stay clean]
- exact_next_action: work full phase supersede-lineage
