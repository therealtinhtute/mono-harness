# Workflow Harness — Gap Matrix

Current-state gap inventory for the workflow-harness initiative. Each row names a concrete gap in today's `skills/workflow/` chain, which skill owns closing it, the artifact that will carry the fix, the risk of leaving it open, and the roadmap phase that closes it. See `skills/workflow/README.md` for the target model and `.kit/planning/SPEC.md` for the full requirement set.

## Matrix

| Gap | Owner-skill | Artifact | Risk | Phase |
| :--- | :--- | :--- | :--- | :--- |
| Request classification (risk lane) lives only in `brainstorm` prose — no ID, no lane is persisted anywhere `check` can read it back | `brainstorm` | `SPEC.md` frontmatter `intake_id`/lane (SPEC R15, R18) | cao — `check`'s deterministic verdict matrix (R19) keys off intake lane; without a persisted lane the gate has no reliable input | skill-adapters |
| `workflow-state.yml` is a single hand-edited pointer file — no history, no schema, unsafe under concurrent sessions, and every skill both reads and writes it | `to-plan`, `work` | SQLite workflow entities: phases/runs/checks/handoffs (SPEC R13, R14) | cao — R14 requires no skill touch the yml after migration; a half-migrated state is worse than today's file | harness-contracts |
| `work` run logs are markdown-only (`.kit/runs/work/*.md`) — nothing queryable links a wave to its verification output or lets `check` pull trace evidence | `work` | `trace` entity via `zharness trace add`, linked to the run (SPEC R6, R18) | vừa — a doc-review problem today; becomes a correctness gap once `check` depends on trace data (R19) | skill-adapters |
| `check` verdicts are subjective prose — no deterministic proof matrix, no `validate` command, no machine-readable findings on failure | `check` | `zharness validate` + verdict matrix (SPEC R11, R19) | cao — inconsistent gates are the whole reason this initiative exists; this is the core correctness surface | validation-gate |
| `watzup`/`handoff` are git+markdown only — no `resume` command, no unified readiness states, no cross-machine continuity snapshot | `watzup`, `handoff` | `zharness resume`; unified states `clean\|in-progress\|drifted\|no-harness` (SPEC R12, R20) | vừa — today's HANDOFF.md blocker (this repo: `.kit/` is gitignored, planning is local-only) is exactly this gap | continuity |
| `zharness` binary does not exist — no command surface, no release pipeline, no install path; every "harness command" referenced above is currently aspirational | all 8 skills | `cli/` Go module, ported command surface (SPEC R5, R6) | cao — every other row in this matrix is blocked until the binary exists and is installable | cli-core |

Zero empty cells above; every row carries all 5 fields.

## Story ↔ Phase Mapping Decision

**Decision: one `zharness` story per `to-plan` phase; story slug = phase slug** (e.g., story id `harness-concept` for phase `harness-concept`). Confirmed, not overturned.

**Rationale:** The 8-phase roadmap already decomposes this initiative into blast-radius-bounded, independently-verifiable units — each phase carries its own `-CONTEXT.md`, `-PLAN.md`, and verification commands. That is exactly what a harness "story" tracks (a story-sized unit of work with proof requirements). A 1:1 mapping means `to-plan`'s existing phase boundary is the single unit `zharness story` records, so `check record` and `resume` only ever need one story ID per phase — no second, competing definition of "done" to reconcile at the gate.

**Rejected alternative: one story per plan task (wave-item).** Rejected because tasks already get per-task verification inside the run artifact's Wave/Task Log, and `trace add` records proof at wave granularity. Promoting every task to a harness story would multiply bookkeeping (20+ stories for this initiative alone) without adding proof granularity the trace layer doesn't already provide.
