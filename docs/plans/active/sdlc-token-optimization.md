---
id: 01KZQT1CYB0R07QHVZ6CHXYTP3
type: plan
intake_id: pending — record via `zharness intake --type harness-improvement --lane normal --summary "SDLC token + cache optimization" --plan-path docs/plans/active/sdlc-token-optimization.md --plan-id 01KZQT1CYB0R07QHVZ6CHXYTP3 --json` on the executing machine, then replace this value
lane: normal
status: active
created: 2026-08-11
updated: 2026-08-11
---

# Plan: SDLC token + cache optimization (roadmap R1–R8, G1)

## Outcome
- result: the workflow chain costs ~36% less per phase with no loss of review depth, resumability, or verification rigor — implementing the roadmap locked in `docs/audit/sdlc-gap-analysis.md` §6.
- success_signals:
  - `query plan --section phase` returns `degraded: false` and only the requested phase block for a plan produced by `zharness scaffold plan` + `to-plan`
  - one `trace add` invocation can record a full wave of task entries; `query traces --phase {slug}` output is row-for-row identical to the per-call cadence
  - `preflight check --json` carries the same `context` packet shape as `preflight work`
  - re-running the cost model (`scratchpad` methodology re-created from the audit docs) shows `work` ≤ 32 turns and per-phase gate cost reduced ≥ 50%

## Authority and Requirements
- authority:
  - `docs/audit/sdlc-token-cache-audit.md` — measured findings F1–F5 and optimization spec P1–P4
  - `docs/audit/sdlc-gap-analysis.md` — five-criteria gap analysis, roadmap R1–R8/G1–G2, sequencing decision
  - owner instruction 2026-08-11: "tạo plan chi tiết thành file, làm sau" — plan now, execute later
- requirements:
  - R1 [accepted]: `extractPlanPhaseBlock` matches both the `### phase_slug: \`{slug}\`` heading form and the scaffold template's `- phase_slug: {slug}` list form; a list-form plan built from the literal `scaffold plan` output returns `degraded: false` with sibling phases excluded | source: audit F2 (measured: degraded path returns 5,863 B vs 1,290 B intended)
  - R2 [accepted]: `query traces`/`query decisions`/`query checks` with `--phase {slug}` where no story row has that slug fail with a distinct `unknown_phase` user error instead of returning `[]` exit 0 | source: gap analysis C4
  - R3 [accepted]: `skills/workflow/{work,check,brainstorm,handoff}/references/` are deleted; `git/` and `interview/` references are preserved; doc-link gate stays green | source: audit F5 (orphaned since logic moved to playbooks)
  - R4 [accepted]: the per-phase gate runs in-session via `check gate` (inheriting the work session's model and warm cache); the complete Security/Performance/Architecture/Code Quality review (`check full`) is required exactly once, on the final phase, before initiative closure — enforced as a `handoff` precondition, not a suggestion | source: audit F1 ($0.275/phase model-switch cost) + spec P2's load-bearing half
  - R5 [accepted]: `trace add` accepts a batched `--tasks '[{"task","task_status","summary"},...]'` form that writes one DB row and one Progress entry per element atomically; `work` flushes per wave, except entries with status `BLOCKED`/`NEEDS_CONTEXT`/`DONE_WITH_CONCERNS` which flush immediately | source: audit F3 (41.6% ceremony share; −16% modeled)
  - R6 [accepted]: `preflight check --json` returns the same `context` packet as `preflight work` (position, phases, latest IDs, drift), and `check.md` no longer mandates a separate `resume --json` call | source: gap analysis C3 (measured: check/to-plan preflights lack `context`)
  - R7 [accepted]: `db status`'s `context_cost_estimate` reflects the section-read path (post-R1) instead of the full-plan-read path, and its self-confessed drift note is removed | source: gap analysis §3
  - R8 [accepted]: every `zharness` invocation appends one JSONL line `{ts, argv, exit, ms, error_code}` to `.kit/log/zharness.jsonl` (gitignored), rotating at 1 MB, with zero change to stdout/stderr contracts | source: gap analysis G3 (no forensic record outside agent transcripts)
  - R9 [accepted]: `skills/workflow/README.md` explicitly declares deployment and production monitoring out of scope for the chain (or names the future `ship` skill if a deploy target exists by execution time) | source: gap analysis G1 — the gap is that the non-goal is undeclared

## Non-goals
- NG1: single-model chain — `brainstorm`/`to-plan` stay on opus; wrong decisions there are cheapest to make and most expensive to live with
- NG2: removing `check record`'s proof re-execution — it closed a real verification bypass (`c53fb76`); wall-clock cost is accepted
- NG3: trimming playbooks for size — they sit in the cached prefix; ~$0.002/phase is not worth capability
- NG4: subagent fan-out for parallel-safe tasks (G2) — deferred until this plan lands and the cost model is re-measured
- NG5: building a deploy/monitoring skill — R9 resolves G1 by declaration unless the owner names a real deploy target
- NG6: changing the per-task granularity of trace *rows* in the DB — batching changes call cadence only; resumability guarantees are about what the DB holds

## Approach and Risks
- approach: four phases ordered by risk, not size — a zero-risk CLI/docs commit first (p1), the single judgment-bearing routing change isolated second (p2), the round-trip-reduction CLI release third (p3), observability and coverage declaration last (p4). Each phase is independently shippable and independently revertible; p2 is the only phase that alters what review runs when, so it carries the strictest verification.
- constraints:
  - every CLI change must keep `cli/docs/CONTRACT.md` in sync in the same commit (projection-drift test enforces embedded docs; CONTRACT is hand-maintained)
  - the gate (`bash scripts/verify-doc-links.sh` && `cd cli && go test ./...`) must pass before every commit — no exceptions, including docs-only phases
  - playbook edits must change `cli/docs/embedded/playbooks/*` (source of truth) — `docs/playbooks/*` is scaffolded output; editing only the scaffold regresses on next `zharness init --refresh-docs`
  - harness state is per-machine: at execution start, run `zharness init` if needed, record the intake, and run `to-plan` to mint story rows for the phases below (definitions here are final; only story IDs are pending)
- risks:
  - risk: p2 silently degrades review depth if the final-phase `check full` precondition is written as guidance instead of a hard gate | mitigation: the precondition lands in `handoff.md` step 6 as a closure-blocking requirement AND `handoff record --close-phase` paths are tested to reject closure of the final phase without a recorded full-mode check; if CLI-side enforcement proves too invasive, the playbook gate alone ships and the CLI check is logged as a follow-up decision
  - risk: batched `trace add` loses un-flushed entries if a session dies mid-wave | mitigation: R5's immediate-flush rule for non-clean statuses; only uninterrupted clean progress is ever buffered, and a dead session's recovery path (`preflight` → `query traces`) already treats missing entries as not-done
  - risk: regex widening (R1) matches a stray `- phase_slug:` line in prose | mitigation: anchor to line start with optional list indentation, require the slug charset `[a-z0-9-]+`, and slice strictly between sibling markers; the heading form keeps precedence
  - risk: JSONL logging (R8) breaks the read-only guarantee of query commands on read-only filesystems | mitigation: log writes are best-effort — a failed log append never fails the command
- recovery: every phase is a plain git revert; no phase migrates DB schema; `.kit/log/` is gitignored so R8 leaves no repo residue

## Phases and Verification
<!-- Phase and task definitions are immutable after to-plan. Do not add task status fields. Append-only Progress is the sole task execution-status source. Only each phase lifecycle status changes to mirror DB transitions: to-plan=planned; work after run create=in-progress; clean durable check=checked; closing handoff=done. -->
- planning_status: planned
- phases:

### phase_slug: `p1-quick-wins`
- story_id: pending — mint via `zharness story --slug p1-quick-wins --goal "Zero-risk fixes: plan-section regex, unknown-phase error, orphaned references" --json`
- status: planned
- goal: land the three findings that need no judgment call — restore the 4.3× plan-section read (R1), make unknown-phase filters fail loudly (R2), delete the four orphaned references directories (R3)
- depends_on: none
- touched_surfaces: [cli/internal/application/plan_query.go, cli/internal/application/plan_query_test.go, cli/internal/interfaces/query_plan_test.go, cli/internal/application/query.go, cli/internal/interfaces/query_checks_test.go, cli/docs/CONTRACT.md, skills/workflow/work/references/**, skills/workflow/check/references/**, skills/workflow/brainstorm/references/**, skills/workflow/handoff/references/**]
- avoided_surfaces: [docs/playbooks/**, cli/docs/embedded/playbooks/**, skills/workflow/git/references/**, skills/workflow/interview/references/**, cli/internal/infrastructure/**]
- waves:
  - wave: 1
    tasks:
      - task: widen `extractPlanPhaseBlock` (and `planPhaseHeading` or a sibling list-form matcher) in `plan_query.go` to accept `^\s*- phase_slug: ([a-z0-9-]+)\s*$` list items, slicing to the next sibling `- phase_slug:` line at the same-or-lesser indent or the next `## `/`### phase_slug:` boundary; heading form keeps precedence when both exist
        outputs: [cli/internal/application/plan_query.go]
        check: "cd cli && go test ./internal/application/ -run 'PlanSection|PlanQuery' -v"
      - task: add test fixtures built from the *literal* output of `zharness scaffold plan --path` with phases filled in list form (copy the shape from `docs/audit/sdlc-token-cache-audit.md` §4), asserting `degraded == false`, sibling-phase exclusion, and CRLF preservation; keep all existing heading-form cases green
        outputs: [cli/internal/application/plan_query_test.go, cli/internal/interfaces/query_plan_test.go]
        check: "cd cli && go test ./... "
      - task: make `query traces|decisions|checks --phase {slug}` return user error `unknown_phase` (exit 1) when no story row carries the slug; `query plan --section phase --phase {slug}` keeps its documented degraded-fallback behavior (it has its own contract); update CONTRACT.md's error-code enumeration in the same commit
        outputs: [cli/internal/application/query.go, cli/internal/interfaces/query_checks_test.go, cli/docs/CONTRACT.md]
        check: "cd cli && go test ./internal/... -run 'Query' && grep -q unknown_phase cli/docs/CONTRACT.md"
  - wave: 2
    tasks:
      - task: delete `skills/workflow/{work,check,brainstorm,handoff}/references/` recursively; verify no SKILL.md, playbook, README, or doc references them; leave `git/` and `interview/` references untouched
        outputs: [deletion only]
        check: "bash scripts/verify-doc-links.sh && test ! -d skills/workflow/work/references && test -d skills/workflow/git/references"
- checks (phase gate): "bash scripts/verify-doc-links.sh && cd cli && go test ./..."

### phase_slug: `p2-check-routing`
- story_id: pending — mint via `zharness story --slug p2-check-routing --goal "Per-phase gate in-session, full review once at closure" --depends-on p1-quick-wins --json`
- status: planned
- goal: stop paying the opus cold-cache round trip every phase (R4): the per-phase gate becomes `check gate` executed in-session; the complete manual review runs exactly once, on the final phase, gated by handoff closure preconditions
- depends_on: ['p1-quick-wins']
- touched_surfaces: [cli/docs/embedded/playbooks/work.md, cli/docs/embedded/playbooks/check.md, cli/docs/embedded/playbooks/handoff.md, docs/playbooks/work.md, docs/playbooks/check.md, docs/playbooks/handoff.md, cli/internal/embedded/**, skills/workflow/check/SKILL.md]
- avoided_surfaces: [cli/internal/application/check_record.go — verdict/proof semantics unchanged, skills/workflow/work/SKILL.md — routing lives in the playbook]
- waves:
  - wave: 1
    tasks:
      - task: edit embedded `work.md` step 11 — route the per-phase gate to `check gate` (durable, in-session), stating explicitly that the complete manual review is deferred to the final phase; regenerate/scaffold `docs/playbooks/work.md` so both copies agree (projection-drift test guards this)
        outputs: [cli/docs/embedded/playbooks/work.md, docs/playbooks/work.md]
        check: "cd cli && go test ./internal/embedded/ -run Drift"
      - task: edit embedded `check.md` — document `gate` as the standard per-phase path and `full` as the closure review; no mode-semantics change (gate already skips the manual review per §5)
        outputs: [cli/docs/embedded/playbooks/check.md, docs/playbooks/check.md]
        check: "cd cli && go test ./internal/embedded/ -run Drift"
      - task: edit embedded `handoff.md` steps 5–6 — closing the FINAL phase additionally requires the latest check for that phase to be mode `full` with a clean verdict; a gate-mode check is sufficient for non-final phase closure only; word it as a closure-blocking precondition in the same register as the existing "no alternative or early-completion condition" clause
        outputs: [cli/docs/embedded/playbooks/handoff.md, docs/playbooks/handoff.md]
        check: "cd cli && go test ./internal/embedded/ -run Drift && grep -qi 'full' docs/playbooks/handoff.md"
  - wave: 2
    tasks:
      - task: evaluate CLI-side enforcement — whether `check record` should persist mode (gate|full) so `handoff record --close-phase` can reject final-phase closure without a full-mode check; if the schema/changeset cost is disproportionate, record the decision (playbook-gate-only) via `zharness decision add` and stop — do NOT half-implement
        outputs: [either cli/internal/application/{check_record,handoff}.go + migration + tests, or a recorded decision]
        check: "cd cli && go test ./... (if implemented) — otherwise `zharness query decisions --phase p2-check-routing --json` shows the recorded decision"
- checks (phase gate): "bash scripts/verify-doc-links.sh && cd cli && go test ./..."

### phase_slug: `p3-fewer-round-trips`
- story_id: pending — mint via `zharness story --slug p3-fewer-round-trips --goal "Batch trace add, stage-shaped preflight for check, honest cost gauge" --depends-on p1-quick-wins --json`
- status: planned
- goal: remove the mandated round trips that return ≤150 bytes each — batched wave traces (R5), a `context` packet on `preflight check` (R6), and a cost gauge that measures the current read path (R7)
- depends_on: ['p1-quick-wins']
- touched_surfaces: [cli/internal/interfaces/trace*.go, cli/internal/application/trace.go, cli/internal/application/preflight.go, cli/internal/interfaces/preflight.go, cli/internal/application/db_status.go, cli/internal/application/context.go, cli/docs/CONTRACT.md, cli/docs/embedded/playbooks/work.md, cli/docs/embedded/playbooks/check.md, docs/playbooks/work.md, docs/playbooks/check.md]
- avoided_surfaces: [cli/internal/infrastructure/changeset.go — one changeset entry per trace row is preserved, docs/plans/** ]
- waves:
  - wave: 1
    tasks:
      - task: add `trace add --tasks '[{"task","task_status","summary"},...]'` accepting 1–20 entries, mutually exclusive with the single-task flags; write one DB row + one changeset entry + one appended Progress line per element, atomically under the repository lock; response is `[{"id":...},...]` in input order; single-task form unchanged
        outputs: [cli/internal/interfaces (trace command), cli/internal/application/trace.go, cli/docs/CONTRACT.md]
        check: "cd cli && go test ./... -run 'Trace'"
      - task: add a replay-equivalence test — a batched write followed by `query traces --phase {slug}` yields rows byte-identical (minus IDs/timestamps) to the same entries written per-call; DB rebuilt from changesets must agree
        outputs: [cli/internal/infrastructure/repository_replay_test.go or sibling]
        check: "cd cli && go test ./internal/infrastructure/ -run Replay"
      - task: extend `preflight check --mode {gate|full} --json` to include the same `context` packet as `preflight work` (position, phases, latest run/check/handoff IDs, drift); update CONTRACT.md
        outputs: [cli/internal/application/preflight.go + call sites, cli/docs/CONTRACT.md]
        check: "cd cli && go test ./... -run 'Preflight'"
      - task: update `db status` `context_cost_estimate` to model the section-read path (current-state slice + one phase block) for work/handoff/watzup instead of full-plan bytes; delete the drift-confession note; keep the bytes/4 heuristic and its attribution
        outputs: [cli/internal/application/db_status.go]
        check: "cd cli && go test ./... -run 'DbStatus|Status'"
  - wave: 2
    tasks:
      - task: edit embedded `work.md` steps 7/9 — task statuses are collected as tasks complete and flushed once per wave via `--tasks`; entries with status `BLOCKED`/`NEEDS_CONTEXT`/`DONE_WITH_CONCERNS` flush immediately in their own call; wave-summary trace (step 9) unchanged
        outputs: [cli/docs/embedded/playbooks/work.md, docs/playbooks/work.md]
        check: "cd cli && go test ./internal/embedded/ -run Drift"
      - task: edit embedded `check.md` step 1 — read lifecycle position from the preflight `context` packet; delete the separate `zharness resume --json` mandate (keep `resume` itself — other callers use it)
        outputs: [cli/docs/embedded/playbooks/check.md, docs/playbooks/check.md]
        check: "cd cli && go test ./internal/embedded/ -run Drift"
- checks (phase gate): "bash scripts/verify-doc-links.sh && cd cli && go test ./..."

### phase_slug: `p4-observability-and-scope`
- story_id: pending — mint via `zharness story --slug p4-observability-and-scope --goal "JSONL invocation log; declare post-merge lifecycle scope" --depends-on p3-fewer-round-trips --json`
- status: planned
- goal: give failed lifecycles a forensic record outside agent transcripts (R8) and make the missing deploy/monitoring stages a declared non-goal instead of an ambient gap (R9)
- depends_on: ['p3-fewer-round-trips']
- touched_surfaces: [cli/cmd/zharness/main.go, cli/internal/interfaces/root.go, .gitignore (\.kit/ already covered — verify), skills/workflow/README.md]
- avoided_surfaces: [stdout/stderr shapes — CONTRACT.md must not change for R8, cli/internal/application/**]
- waves:
  - wave: 1
    tasks:
      - task: append one JSONL line `{"ts","argv","exit","ms","error_code"}` to `.kit/log/zharness.jsonl` on every invocation, wired at the root-command exit path; best-effort (append/mkdir failure never alters the exit code or output); rotate by rename to `.1` at 1 MB; redact nothing (argv carries no secrets by contract)
        outputs: [cli/cmd/zharness/main.go or cli/internal/interfaces/root.go]
        check: "cd cli && go test ./... && cd /tmp && rm -rf zlogtest && mkdir zlogtest && cd zlogtest && git init -q . && zharness preflight watzup --json >/dev/null 2>&1; test -s .kit/log/zharness.jsonl"
      - task: add a scope paragraph to `skills/workflow/README.md`: the chain covers plan→code→verify→commit/PR; deployment, release management, and production monitoring are explicitly out of scope (revisit if a deploy target materializes — G2/ship-skill pointer to the gap analysis)
        outputs: [skills/workflow/README.md]
        check: "bash scripts/verify-doc-links.sh && grep -qi 'out of scope' skills/workflow/README.md"
- checks (phase gate): "bash scripts/verify-doc-links.sh && cd cli && go test ./..."

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
- lifecycle_status: planned (definitions final; DB rows pending — harness state is per-machine)
- latest_run_id: none
- latest_trace_ids: []
- latest_check_id: none
- latest_handoff_id: none
- blockers: none
- open_items: [record intake and replace frontmatter intake_id, mint story rows for the four phases (`to-plan` resume path: definitions exist, only IDs are pending), p2 wave-2 CLI-enforcement decision]
- exact_next_action: on the executing machine — `zharness init` (if needed) → record intake → `to-plan full` to mint story IDs → `work full phase p1-quick-wins`
