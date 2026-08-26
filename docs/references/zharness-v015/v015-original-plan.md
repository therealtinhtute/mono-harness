# Archive: zharness v0.15 plan, pre-review draft (superseded)

> **Archive — not a live plan.** This is `docs/plans/active/zharness-v015-slim.md` as it
> stood on 2026-08-26 before the read-only review, the `/interview` pass, and the merge
> with v0.13. It carried 7 success signals, 10 requirements, 7 non-goals,
> `approach: not-planned`, and `phases: none`.
>
> The review found 4 blockers against this draft — see `docs/references/zharness-v015/review-findings.md`.
> The 7 decisions that closed them are in `docs/references/zharness-v015/interview-spec.md`.
> The live plan is `docs/plans/active/zharness-v015-slim.md`.
>
> Kept because the review's evidence column cites this file by line number; without it those
> citations point at content that no longer exists.

---

---
id: 01M0Z674XY7Y2CTSYVKYVYPV79
type: plan
intake_id: 01M0Z679N4DKH6XWC1RSTPW3AG
lane: high-risk
status: active
created: 2026-08-26
updated: 2026-08-26
---

# Plan: zharness v0.15 "slim" — Installer-only binary, markdown-only state, fail-open

## Outcome
- result: zharness becomes an installer/updater binary only (repository-harness model). State is git-committed markdown alone. The full lifecycle runs fail-open from repo-local instructions (AGENTS.md block v2 + docs/playbooks + scripts) with zero CLI dependency, and enforcement moves to repo scripts + git hooks + CI. SQLite and the entire lifecycle command surface are deleted from the source tree as a breaking release.
- success_signals:
  - S1 [kill-switch]: with the binary absent from PATH, a real task completes from repo-local instructions alone — zero STOP, correct markdown bookkeeping.
  - S2: `scripts/record-check.sh` re-executes every proof command before an APPROVED Validation entry is written; the pre-commit hook refuses commits carrying an APPROVED entry that never passed the script; CI re-runs it.
  - S3: the whole lifecycle (brainstorm → to-plan → work → check → handoff) completes via markdown + scripts; `zharness --help` shows only install/update/uninstall.
  - S4: `rg -i "sqlite|harness\.db" cli/` ≈ 0 except the EOL note in CHANGELOG.
  - S5 [identity test]: on a freshly scaffolded consumer repo, a new unprimed session answers "what is this / how is it architected / what is in progress" correctly from `docs/PROJECT.md` + plans alone.
  - S6: exactly 2 fail-closed guards remain in the whole system (proof verification via script+hooks, independent judge for high-risk lanes at the playbook layer) plus 1 pause-point (material product ambiguity); no other fail-closed gate exists.
  - S7: chain cost −30% vs the 8/2026 baseline and single-session cache-read ratio >80%, measured by the method in `docs/audit/sdlc-token-cache-audit.md`.

## Authority and Requirements
- authority:
  - `.kit/plans/2026-08-26-zharness-v013-slim/plan.md` — approved 2026-08-26; its decisions 1, 2, 4, its research anchors, and its audit evidence table carry over (baseline numbers refreshed to 0.14.0).
  - `docs/plans/completed/harness-markdown-truth.md` — markdown is already the sole source of truth; `db rebuild` reconstructs the index from committed markdown alone, proving the DB is disposable.
  - `cli/docs/CONTRACT.md` — current fail-closed surface, `check record` proof-verification contract (§189 area), `preflight` blocking behavior (§44 area).
  - hoangnb24/repository-harness: AGENTS.md (read live — ~20-line block, "no task database", work-shape routing), decision 0027 (EOL playbook: pin last release, consumer-owned bytes, no `legacy/` dir), README (the `harness` binary is installer + safe updater only).
  - sellke/writ — repo scripts (`eval.sh`) + GitHub Actions enforcing markdown schema; kf-rahman/Agentic-Workflow — git hooks enforcing commit-time rules.
  - OpenAI Codex config reference — `project_doc_fallback_filenames` is a real key (array, "additional filenames to try when AGENTS.md is missing"); opencode docs — global `~/.config/opencode/AGENTS.md` + project `AGENTS.md` load natively.
  - `docs/audit/consumer-adoption-audit.md` (D1–D4), `docs/audit/sdlc-token-cache-audit.md` (token/cache baseline).
  - Owner decisions, brainstorm session 2026-08-26: installer-only binary + repo scripts; skip all global skill-tree merge/scanning (leave the 6 global SKILL.md hard-stops untouched and out of the product path); drop the global instruction merge, replace with repo-local instruction layer + one codex config line; accept memory-feature loss with a filename convention; re-lock as v0.15 against the 0.14.0 codebase.
  - Alternatives rejected: keeping 3 CLI verbs (status/record check/doctor) — rejected because a binary command is not CI-enforceable and stays a parallel control plane; keeping proof verification in the binary — same reason; merging the two global skill trees — owner said skip; keeping SQLite as derived index — rejected because the index's only consumer (the CLI lifecycle) is being deleted, and repo-harness 0027 shows the cost of carrying dead protocol surfaces.
- requirements:
  - R1 [accepted]: The binary exposes exactly install / update / uninstall. Every lifecycle command (preflight, resume, query×9 views, audit, init, migrate, import, db rebuild/status, intake, story, run create, trace add, decision add, handoff record, validate, intervention, id, memory×3, scaffold, status, record check, doctor) is deleted from source, not hidden or deprecated. | source: owner decision; 0027 "no legacy/ dir"
  - R2 [accepted]: `scripts/record-check.sh` re-executes every proposed proof command (exit 0 required, 5-minute timeout each — same semantics as current `check record`) before an APPROVED/APPROVE_WITH_REQUESTS Validation entry is appended; the pre-commit hook exits non-zero for any committed APPROVED entry lacking a script pass marker; CI re-runs the script. | source: CONTRACT.md check record; writ eval.sh pattern; owner
  - R3 [accepted]: The updater uses a three-way merge on `.zharness/base/` (BASE/LOCAL/UPSTREAM); conflicts stop for human resolution (`--continue`/`--abort`); activation is transactional; no consumer-owned file outside the managed set is ever deleted. | source: repository-harness updater
  - R4 [accepted]: EOL playbook: the installer/updater never deletes a consumer's `harness.db` or sidecars (consumer-owned bytes); CHANGELOG carries the v0.15 breaking note; consumers pinning 0.14.x keep a working product. | source: decision 0027; owner
  - R5 [accepted]: AGENTS.md block v2 is ≤ ~20 lines, zero-CLI-required, work-shape routing, with the closing sentence "no task database, no parallel control-plane state"; CLAUDE.md remains an `@AGENTS.md` import bridge; codex gets exactly one config line: `project_doc_fallback_filenames = ["CLAUDE.md"]`. | source: repository-harness AGENTS.md; codex config reference; owner
  - R6 [accepted]: No file under `~/.claude`, `~/.codex`, `~/.agents`, or `~/.config/opencode` is scanned, merged, or modified by this initiative except the single codex config line in R5. | source: owner decision (skip global merge)
  - R7 [accepted]: `docs/memory/{id}.md` replaces SQLite memory; supersede lineage and diacritic-fold retrieval are consciously dropped (filename carries id; agents grep directly). | source: owner decision
  - R8 [accepted]: Each phase is independently mergeable; `cd cli && go test ./...` and `bash scripts/verify-doc-links.sh` pass at every phase boundary. | source: harness-markdown-truth R12 pattern
  - R9 [accepted]: The fresh-session identity test (S5) runs with zero priming; answers must come only from the scaffolded repo. | source: original plan P2a verify; owner
  - R10 [accepted]: No fabricated history backfill; brownfield onboarding is read-only-first (deterministic detection → drafted proposal → human approval before writes). | source: decision 0020; issue #25; owner

## Non-goals
- NG1: No merge, sync, or edit of the existing global instruction layer (two skill trees, global rules, codex AGENTS.md) beyond the single codex config line in R5. The 6 global SKILL.md hard-stops stay untouched; the slim product path does not depend on them.
- NG2: No hidden or deprecated lifecycle commands in the binary — deleted from source entirely, history preserved only in git tags.
- NG3: No edits to consumer repositories beyond what the installer/updater manages; no fabricated backfill of consumer history.
- NG4: No application runtime, credentials, schema validation, or product policy shipped — "no fabricated application truth".
- NG5: No SQLite read-compatibility shims; old consumer databases are consumer-owned bytes.
- NG6: The six-stage pipeline shape (brainstorm → to-plan → work → check → handoff, plus watzup) is unchanged; only the enforcement mechanism moves from CLI to scripts/discipline.
- NG7: No `status`/`doctor` binary commands in this release; a read-aggregator script (`scripts/status.sh`) is an optional convenience, not a workflow dependency.

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