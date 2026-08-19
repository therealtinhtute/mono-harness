---
id: 01M0C5EMT7HMP1RGZH9MV5GKET
type: plan
intake_id: 01M0C5EQETZ60HJDWXG3JJHVTA
lane: normal
status: active
created: 2026-08-19
updated: 2026-08-19
---

# Plan: docs architecture — authored project docs, then consumer scaffold

## Outcome
- result: every file under `docs/` has an unambiguous owner (managed, authored, or scaffold-once), and this repository gains real authored documentation reachable from one entrypoint — first by hand for this repo, then by `zharness init` for consumer repos.
- success_signals:
  - A fresh agent session answers "what is this repo's architecture and which decisions are locked" by reading `docs/README.md` plus one linked file, with zero `grep` over source.
  - Every path under `docs/` appears in the ownership table in `docs/README.md`; an unlisted path is a defect.
  - `bash scripts/verify-doc-links.sh` and `cd cli && go test ./...` both pass.
  - Phase A merges and stays useful even if Phase B never ships.

## Authority and Requirements
- authority:
  - `cli/internal/application/managed_docs.go:107-112` — membership in the embedded FS is the sole trigger for hash tracking; `AGENTS.md` is the only exclusion. A file outside that tree can never be tracked or conflict-staged.
  - Verified 2026-08-19: `grep -n "Remove\|Delete\|prune\|trash" cli/internal/application/managed_docs.go` returns empty — no code path deletes a local file, so non-embedded files survive `init --refresh-docs` untouched.
  - `cli/internal/application/init.go:59-104` — `writeAgentsManagedBlock`'s absent-file branch (`os.IsNotExist` → write, else leave alone) is the existing shape for a write-once-if-absent scaffold.
  - `cli/internal/application/init_test.go:12,28,70` — `docsVersion` is a per-call parameter against a `fstest.MapFS` fixture, not a global constant; bumping the real docs version does not break this test.
  - `docs/audit/consumer-adoption-audit.md:135` — the 26,365-token figure is explicitly "Inferred, not observed"; it is a file size, attributed to the ambiguous-active-plan invariant, which commit `7a4195f` already closed.
  - `docs/plans/completed/durable-memory.md:38` (NG3) — wiring `zharness memory` into playbooks is deferred to a separate initiative.
  - Measured 2026-08-19 on a clean tree at `1838a7b`: `bash scripts/verify-doc-links.sh` reports **16 broken doc cross-references** and exits non-zero. `CLAUDE.md` links two docs that do not exist (`docs/prompt-engineering-principles.md`, `docs/workflow-harness/migration.md`); four playbooks link three missing audit files; `skills/workflow/README.md` links six missing files. The repository's own `check` gate is therefore red before any change in this initiative, which is direct observed evidence — unlike the inferred token figure — that authored documentation is missing rather than merely unrouted.
  - Owner decision, 2026-08-19: split the work so markdown-only lands first and binary changes are a separate, optional second phase.
  - `hoangnb24/repository-harness` — upstream docs taxonomy (`docs/README.md` router, `ARCHITECTURE.md`, `decisions/NNNN-*.md` with an index, `templates/`), and its stated principle "start with the smallest authoritative surface".
- requirements:
  - R1 [accepted]: `docs/README.md` exists in this repo and contains an ownership table naming, for every existing path under `docs/`, exactly one class — `managed` (present in `cli/docs/embedded/`), `authored` (human-written, never embedded), or `scaffold-once` (written by `init` only when absent). | source: `managed_docs.go:107-112`
  - R2 [accepted]: `docs/README.md` is the single entrypoint; every authored doc is reachable by link from it. No authored doc is discoverable only by directory listing. | source: upstream `repository-harness` docs router
  - R3 [accepted]: no file created by this initiative is added to `cli/docs/embedded/`. Verification: `find cli/docs/embedded -name README.md -o -name ARCHITECTURE.md -o -name decision.md` returns nothing. | source: `managed_docs.go:112`
  - R4 [accepted]: `docs/ARCHITECTURE.md` states the harness's own design in prose, and every structural claim cites a real `path:line` that exists at merge time — at minimum: markdown as source of truth, `harness.db` as a derived rebuildable index, managed-docs hash tracking with conflict staging, and the single `ResolveActivePlan` contract. | source: `docs/plans/completed/harness-markdown-truth.md`, `cli/internal/application/plan_resolve.go:73`
  - R5 [accepted]: `docs/decisions/` holds numbered ADRs with an index README, and the decisions already made and verified — the D1 single-resolver contract, markdown-as-truth, and the deliberate deferral of memory-playbook wiring — are each recorded as an ADR rather than left only in commit messages or session transcripts. | source: `docs/plans/completed/*.md`, commit `7a4195f`
  - R6 [accepted]: Phase A touches only new files plus `docs/README.md`-adjacent authored paths; it changes no Go source and requires no release. Verification: the Phase A diff contains no `.go` file and no path under `cli/docs/embedded/`. | source: owner decision 2026-08-19
  - R7 [accepted]: Phase B adds a scaffold-once class to `zharness init` that writes `docs/README.md`, `docs/decisions/README.md`, and a decision template only when each is absent, records no `managed_docs` row, and leaves consumer edits byte-identical across `init --refresh-docs` with zero new `.kit/conflicts/` entries. | source: `init.go:59-104`, verified absence of any deletion path
  - R8 [accepted]: Phase B adds routing by convention only — one sentence in `cli/docs/embedded/WORKFLOW.md` instructing the agent to read `docs/README.md` once per session when it exists, treating its absence as a non-error — and introduces no new `preflight` field and no per-repo configuration. | source: existing `WORKFLOW.md` stage table precedent
  - R9 [accepted]: Phase B ships behind a `cli/v*` release tag; the bare `vX.Y.Z` tag is never pushed. | source: existing release convention in `.github/workflows`
  - R10 [accepted]: after Phase A, `bash scripts/verify-doc-links.sh` exits zero. Each of the 16 broken references is resolved by exactly one of: writing the missing doc, retargeting the link to a doc that exists, or adding a `.claimignore` entry carrying a `# reason`. Deleting a link to hide a genuinely missing doc is not an accepted resolution. | source: measured gate failure at `1838a7b`; `CLAUDE.md` gate-commands section

## Non-goals
- NG1: no JSON, YAML, or frontmatter routing configuration. The reader is always an LLM, and the flat markdown table already in `WORKFLOW.md` proves markdown suffices.
- NG2: no edits to any file inside `cli/docs/embedded/` during Phase A, and no edits to the projected `docs/WORKFLOW.md` or `docs/playbooks/*.md` at any point in this initiative except the single `WORKFLOW.md` sentence in Phase B made at the embedded source. Projected copies are hash-tracked; local edits are conflict-staged to gitignored `.kit/conflicts/` and lost.
- NG3: no wiring of `zharness memory` into any spine playbook. `durable-memory.md:38` deferred that decision to a later initiative; bundling it here reopens closed scope.
- NG4: no placeholder or filler content. A doc ships with real content or is not created. Specifically, no empty `ARCHITECTURE.md` skeleton is scaffolded into consumer repos in Phase B.
- NG5: no auto-generated, AST-derived documentation (the `codesight` model). Reviewed and rejected for this repo — the docs here describe decisions, not code shape.
- NG6: the 26,365-token figure is not used as justification for any part of this work. It is inferred rather than measured and its stated cause was already fixed. This initiative is justified by ownership ambiguity and the absence of authored project docs, not by that number.
- NG7: no move or rename of `cli/docs/{CONTRACT,SCHEMA,STATE}.md`. They stay at their current paths and are indexed from `docs/README.md`.

## Approach and Risks
- approach: not-planned
- constraints:
  - none
- risks:
  - none

## Phases and Verification
<!-- Phase and task definitions are immutable after to-plan. Do not add task status fields. Append-only Progress is the sole task execution-status source. Only each phase lifecycle status changes to mirror DB transitions: to-plan=planned; work after run create=in-progress; clean durable check=checked; closing handoff=done. Each planned phase records phase_slug, story_id, status, goal, depends_on, waves, tasks, and checks. -->
- planning_status: not-planned
- phases: none

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
- open_items: [to-plan must define stable phases, stories, waves, tasks, and checks]
- exact_next_action: to-plan
