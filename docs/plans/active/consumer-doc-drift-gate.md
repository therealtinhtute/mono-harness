---
id: 01M0FAVR17J6RGK6S40HRNA7F3
type: plan
intake_id: 01M0FAY5EZSM3CPBCAXF6RTCQB
lane: normal
status: active
created: 2026-08-20
updated: 2026-08-20
---

# Plan: consumer-repo documentation — guard the authored half, generate nothing

## Outcome
- result: authored documentation in a consumer repository gains an integrity signal of the kind managed documentation already has. `zharness audit` can tell an agent whether the repository still has authored documentation at all, and whether the top-level architecture document still describes the code it cites. The harness continues to generate no documentation and to scaffold no new file.
- success_signals:
  - `zharness audit` reports a finding when the database holds `managed_docs` rows but every markdown file under `docs/` is a managed path — the signature of the `655c6ac` deletion, which today produces a fully green repository.
  - `zharness audit` reports a finding when a pinned authored document's cited source files have moved past the commit it was pinned to, naming the document, the citations that moved, and the size of the change.
  - A file under `docs/plans/` or `docs/decisions/` never produces a drift finding, no matter how far the code has moved, because those are point-in-time records.
  - A repository whose authored document carries no pin produces no drift finding and no error — pinning is opt-in, and its absence is not a defect.
  - The scaffold-once set stays at exactly three files; `docs/ARCHITECTURE.md` is still never created by the CLI.
  - `bash scripts/verify-doc-links.sh` and `cd cli && go test ./...` both pass.
  - Phase A merges and stays useful even if Phase B never ships.

## Authority and Requirements
- authority:
  - `docs/research/agent-documentation-evidence.md` — external evidence gathered 2026-08-20. F1: LLM-generated context files cost 0.5–2% task success and over 20% inference cost on repositories that already carry documentation, and only help (+2.7%) once every `.md` and `docs/` is deleted. F2: on-demand wiki retrieval is null on correctness and saves cache tokens only. F3: documentation beats code-only by +0.071 on rationale questions and by +0.007 — effectively zero — on anything derivable from code. F5: four independent products converge on the same four-piece drift mechanism. F6: DeepWiki's free tier is public-repository only; codesight's AST support is TypeScript/JavaScript only; CodeWiki has no staleness answer.
  - `docs/plans/completed/docs-architecture.md` — NG4 ("no empty `ARCHITECTURE.md` skeleton is scaffolded into consumer repos") and NG5 ("no auto-generated, AST-derived documentation — the `codesight` model — reviewed and rejected; the docs here describe decisions, not code shape"). Both were owner decisions before this research existed; F1, F3, and F6 now support them from outside.
  - `docs/decisions/0004-docs-directory-deletion-655c6ac.md` — the incident. Commit `655c6ac`, authored through the GitHub web UI, deleted 26 files and 4,285 lines under `docs/`. Its closing sentence states the unclosed gap: "This ADR records the incident; it does not prevent the next one."
  - `docs/audit/consumer-adoption-audit.md` D4 — the consumer `CLAUDE.md` measured at 349 lines / 12,676 bytes / ~3,169 tokens resident every turn, roughly half of it restating what `Read` and `Glob` already disclose: "spend the budget on gotchas, not on what the filesystem shows."
  - `docs/ARCHITECTURE.md` — carries 21 distinct `path/to/file.go:NN` citations (verified 2026-08-20). These are already the page-to-source link that F5 names as mechanism piece one; only the pinned baseline is missing.
  - `cli/internal/application/managed_docs.go:43,56,107` — `SyncManagedDocs` already implements three-hash comparison (recorded, on-disk, embedded) with conflict staging under `.kit/conflicts/`. It covers managed docs only; no equivalent exists for authored docs.
  - `cli/internal/application/audit.go:17-49` — verified 2026-08-20: `Audit(db, cliVersion)` composes `Resume` and `Validate` only. It receives no repository root and touches no filesystem, and `AuditReport`'s three-array JSON shape is a documented public contract.
  - `docs/README.md` — the ownership table; every path under `docs/` must appear in it, and an unlisted path is declared a defect in that table.
  - Owner instruction, this session: research the consumer-repository documentation question against outside practice before choosing a direction, then record the result as durable spec rather than as conversation.
- requirements:
  - R1 [accepted]: The CLI generates no documentation content — no wiki, no overview, no AST-derived or LLM-derived prose — in this initiative or as a consequence of it. | source: `docs/research/agent-documentation-evidence.md` F1; `docs/plans/completed/docs-architecture.md` NG5
  - R2 [accepted]: `zharness audit` gains a finding that fires when the database holds at least one `managed_docs` row and no markdown file under `docs/` lies outside the managed path set — the condition that held after `655c6ac` while every gate stayed green. | source: `docs/decisions/0004-docs-directory-deletion-655c6ac.md`; `docs/research/agent-documentation-evidence.md` F5
  - R3 [accepted]: `zharness audit` gains a finding that fires when an authored document declares a pinned commit and at least one source path it cites has commits after that pin. The finding names the document, each moved citation, and lines added/removed since the pin. | source: `docs/research/agent-documentation-evidence.md` F5 (doccupine's four pieces: link, pinned SHA, change signal, size measurement)
  - R4 [accepted]: Drift checking is scoped by directory, not by judgment. Only top-level `docs/*.md` is eligible; `docs/plans/`, `docs/decisions/`, and `docs/audit/` are exempt by construction because rewriting a point-in-time record to match today's code falsifies the record. | source: `docs/research/agent-documentation-evidence.md` F5 (freemansoft's directory rule); `docs/README.md` ownership table
  - R5 [accepted]: Pinning is opt-in and additive. A document with no pin is not a defect, produces no finding, and produces no error; a repository that never pins anything sees byte-identical `audit` output to today's apart from R2. | source: `docs/plans/completed/docs-architecture.md` NG4 (a doc ships with real content or is not created — the same principle forbids demanding metadata a repo never chose to add)
  - R6 [accepted]: The two findings are severity-graded and distinguishable in `audit --json`, so a consumer can act on one and ignore the other. | source: `docs/research/agent-documentation-evidence.md` F5 ("grade the severity, or the gate gets tuned out")
  - R7 [accepted]: `audit` stays read-only. It creates no WAL/SHM sidecar, writes no file, and repairs nothing it checks — a drift finding is reported, never auto-fixed and never auto-repinned. | source: `cli/internal/application/audit.go` (`OpenReadOnly`); `docs/research/agent-documentation-evidence.md` F5 (Doc Bridge: "CI should not repair the evidence it is checking")
  - R8 [accepted]: A decision record fixes the boundary between what the harness guards and what it delegates to external tooling, and `docs/README.md` names the ownership class for documentation the harness neither writes nor guards. | source: owner instruction, this session; `docs/README.md` ("an existing path under `docs/` that is missing from this table is a defect in this table")
  - R9 [accepted]: Freshness is reported as freshness, never as correctness. Finding text must not imply a pinned, unmoved document is accurate. | source: `docs/research/agent-documentation-evidence.md` F5 (Doc Bridge: "a fresh index can faithfully encode a bad ownership decision")

## Non-goals
- NG1: No documentation generator of any kind — not AST-derived, not LLM-derived, not a repository wiki. F1 measures this as net-negative on repositories that already have `docs/`, which is every repository that has run `zharness init` (R1). This also holds `docs/plans/completed/docs-architecture.md` NG5 in place rather than reopening it.
- NG2: No new scaffolded file. `docs/ARCHITECTURE.md` is still never created by `zharness init`, and the scaffold-once set stays at three files. `docs/plans/completed/docs-architecture.md` NG4 forbids shipping a skeleton, and a link to a file that may not exist would also fail `scripts/verify-doc-links.sh`.
- NG3: No dependency on DeepWiki, codesight, CodeWiki, or any external documentation service — no MCP server call, no network access, no vendored parser. F6 rules each out on its own terms: paywalled for private repositories, TypeScript/JavaScript only, or no staleness answer.
- NG4: No auto-repair, no auto-repin, and no generated pull request. Reporting only (R7). The "detect, then let a human decide" split is deliberate; auto-publishing is what makes drift invisible.
- NG5: No semantic or content comparison between a document and the code it cites. Drift here is a commit-range fact computed from git, not an LLM judgment about whether prose is still true.
- NG6: No change to any spine playbook (`watzup`, `brainstorm`, `to-plan`, `work`, `check`, `handoff`) and no new mandatory step. This holds the boundary `docs/decisions/0003-durable-memory-not-wired-into-playbooks.md` set: "every mandatory playbook step is paid on every invocation, forever, by every consumer repo."
- NG7: No edits to any consumer or reference repository. `onedrive-cloud`, `harness-experimental`, and `codesight` are cited as evidence only, per `docs/plans/completed/harness-markdown-truth.md` NG1.
- NG8: No structural-forensics analysis (hub files, co-change coupling, generated-file detection). F4 identifies it as the one place automation beats a human brief, and it is recorded in the evidence page so a future initiative can pick it up — it is not in this one.
- NG9: No CI workflow, no git hook, and no blocking gate in this initiative. `audit` reports; wiring a report into a blocking check is a separate decision with its own cost.

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
- `2026-08-20T10:21:54Z` — Guard authored consumer-repo documentation instead of generating it; lock the initiative as consumer-doc-drift-gate.. rationale: External evidence (docs/research/agent-documentation-evidence.md F1) measures LLM-generated context files at -0.5% to -2% task success and +20% cost on repositories that already carry docs, and only positive (+2.7%) once every .md and docs/ is deleted -- which is the opposite of a zharness-initialized repo. F3 puts documentation value at +0.007 (zero) on code-derivable questions and +0.071 on rationale. F6 disqualifies every external generator: DeepWiki is public-repo-only on the free tier, codesight is TS/JS-only, CodeWiki has no staleness answer. F5 shows four independent products converging on a four-piece drift mechanism of which this repo already holds three, missing only the pinned baseline SHA. Generating is crowded and net-negative here; guarding matches what the CLI already is..

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
