---
id: 01KZQT1CYB0R07QHVZ6CHXYTP3
type: plan
intake_id: 01KZQVGGFSR9GZH6VCQ2WJFDC1
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
- story_id: 01KZQVGMWQ2PG1RJ47015YP4TP
- status: done
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
- story_id: 01KZQVGMWZADNY0AN10ESTEYQC
- status: done
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
- story_id: 01KZQVGMX7WYAS6S6Z1J6FWTY0
- status: done
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
- story_id: 01KZQVGMXF4NNZH727CR534YYF
- status: checked
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
- `2026-08-11T07:28:55Z` — wave 1. run: `01KZQVJ1X37NPRKF6Z1YDY85KF`. summary: phase p1-quick-wins started.
- `2026-08-11T07:31:52Z` — wave 1, task widen extractPlanPhaseBlock regex. task_status: `DONE`. run: `01KZQVJ1X37NPRKF6Z1YDY85KF`. summary: list-form matcher added, heading precedence kept; go test ./internal/application/ -run 'PlanSection|PlanQuery' -> PASS.
- `2026-08-11T07:31:52Z` — wave 1, task list-form test fixtures + CLI round-trip test. task_status: `DONE`. run: `01KZQVJ1X37NPRKF6Z1YDY85KF`. summary: 9 new tests (app + interfaces layers); go test ./... -> PASS.
- `2026-08-11T07:34:26Z` — wave 1, task unknown_phase error for query traces|decisions|checks --phase. task_status: `DONE`. run: `01KZQVJ1X37NPRKF6Z1YDY85KF`. summary: RequireKnownPhase in application/query.go, wired via checkKnownPhase; CONTRACT.md updated; go test ./internal/... -run Query -> PASS; grep unknown_phase CONTRACT.md -> 3 hits.
- `2026-08-11T07:34:26Z` — wave 1. run: `01KZQVJ1X37NPRKF6Z1YDY85KF`. summary: wave 1 complete: R1 regex widening + tests, R2 unknown_phase error + tests, CONTRACT.md updated.
- `2026-08-11T07:35:15Z` — wave 2, task delete orphaned references/ dirs. task_status: `DONE`. run: `01KZQVJ1X37NPRKF6Z1YDY85KF`. summary: deleted work/check/brainstorm/handoff references/ (6 files); git/ and interview/ kept; 2 historical mentions added to .claimignore following P5w4 precedent; verify-doc-links.sh -> OK.
- `2026-08-11T07:35:15Z` — wave 2. run: `01KZQVJ1X37NPRKF6Z1YDY85KF`. summary: wave 2 complete: F5 orphaned references deleted.
- `2026-08-11T09:24:43.517Z` — handoff recorded. handoff: `01KZR26XBXY8NVEY6QP3VSJD6R`. run: `01KZQVJ1X37NPRKF6Z1YDY85KF`. check: `01KZQW0SW1FY3CQA9HSJ9T1F6V`. phase closed.
- `2026-08-11T09:26:10Z` — wave 1. run: `01KZR291BC9QYFN70AJ82JJHME`. summary: phase p2-check-routing started.
- `2026-08-11T09:30:08Z` — wave 1, task route per-phase gate to in-session check gate. task_status: `DONE`. run: `01KZR291BC9QYFN70AJ82JJHME`. summary: work.md step 11 + Exit Conditions: explicit in-session gate steps, explicit no-dispatch-to-/check-skill (opus frontmatter would defeat the point); go test ./internal/embedded/ -run Drift -> PASS.
- `2026-08-11T09:30:08Z` — wave 1, task document gate as standard per-phase path, full as closure review. task_status: `DONE`. run: `01KZR291BC9QYFN70AJ82JJHME`. summary: check.md Purpose + mode list + Output Format mode line; go test ./internal/embedded/ -run Drift -> PASS.
- `2026-08-11T09:30:08Z` — wave 1, task final-phase check-full closure precondition in handoff.md. task_status: `DONE`. run: `01KZR291BC9QYFN70AJ82JJHME`. summary: step 5 clarifies gate suffices for non-final; step 6 requires clean check full for final phase, notes mode not yet DB-persisted (wave 2 scope); go test ./internal/embedded/ -run Drift -> PASS; grep full docs/playbooks/handoff.md -> hit.
- `2026-08-11T09:30:08Z` — wave 1. run: `01KZR291BC9QYFN70AJ82JJHME`. summary: wave 1 complete: R4 playbook routing shipped across work/check/handoff, all three embedded<->scaffolded in sync, go test ./... -> PASS.
- `2026-08-11T09:31:14Z` — wave 2, task evaluate CLI-side enforcement. task_status: `DONE`. run: `01KZR291BC9QYFN70AJ82JJHME`. summary: evaluated: needs checks.mode migration + a final-phase concept absent from the schema (single-parent depends_on can't answer DAG-terminal queries). Disproportionate per plan's own mitigation clause; recorded decision 01KZR2JG6PRPRJN2H69AXYAG0E, shipping playbook-only precondition instead.
- `2026-08-11T09:31:14Z` — wave 2. run: `01KZR291BC9QYFN70AJ82JJHME`. summary: wave 2 complete: CLI-enforcement evaluated and declined with recorded rationale; playbook-only precondition stands.
- `2026-08-11T12:19:43Z` — wave 1. run: `01KZRC6GR8DZHW7M2ZFT9Q9R3R`. summary: phase p3-fewer-round-trips started.
- `2026-08-11T12:27:40Z` — wave 1, task add trace add --tasks batched flag. task_status: `DONE`. run: `01KZRC6GR8DZHW7M2ZFT9Q9R3R`. summary: batched trace add --tasks implemented (domain.TraceTask, application.CreateTraces, CLI --tasks flag, CONTRACT.md updated); go test ./... -run Trace green.
- `2026-08-11T12:29:40Z` — wave 1, task replay-equivalence test for batched trace add. task_status: `DONE`. run: `01KZRC6GR8DZHW7M2ZFT9Q9R3R`. summary: cli/internal/infrastructure/trace_replay_test.go: batch vs per-call field equivalence + DB-rebuilt-from-changesets agreement; go test ./internal/infrastructure/ -run Replay green.
- `2026-08-11T12:32:42Z` — wave 1, task extend preflight check context packet. task_status: `DONE`. run: `01KZRC6GR8DZHW7M2ZFT9Q9R3R`. summary: check gate/full now receive the same context packet as work/handoff (phases+position+latest IDs+drift); review/bounded stay packet-free; supersedes ceremony-audit NG2; CONTRACT.md updated; go test ./... -run Preflight green.
- `2026-08-11T12:35:47Z` — wave 1, task fix db status context_cost_estimate to model section-read path. task_status: `DONE`. run: `01KZRC6GR8DZHW7M2ZFT9Q9R3R`. summary: watzup/work/handoff now cost Outcome+CurrentState / phase-block / CurrentState respectively instead of full-plan bytes; brainstorm/to-plan/check unchanged; drift-confession note replaced; go test ./... -run 'DbStatus|Status' green.
- `2026-08-11T12:35:47Z` — wave 1. run: `01KZRC6GR8DZHW7M2ZFT9Q9R3R`. summary: wave 1 complete: batched trace add --tasks, replay-equivalence test, preflight check context packet, honest context_cost_estimate — all 4 tasks DONE, full go test ./... and verify-doc-links.sh green.
- `2026-08-11T12:39:04Z` — wave 2, task edit work.md steps 7/9 for batched-flush description. task_status: `DONE`. run: `01KZRC6GR8DZHW7M2ZFT9Q9R3R`. summary: step 7 collects task entries into a pending batch (DONE stays pending, BLOCKED/NEEDS_CONTEXT/DONE_WITH_CONCERNS flushes immediately); step 9 flushes remaining pending entries then the unchanged wave-summary call; Command Reference updated.
- `2026-08-11T12:39:04Z` — wave 2, task edit check.md step 1 to read from preflight context packet. task_status: `DONE`. run: `01KZRC6GR8DZHW7M2ZFT9Q9R3R`. summary: step 1 reads lifecycle position from preflight check context packet instead of a separate zharness resume --json call; Command Reference updated; resume itself untouched for other callers.
- `2026-08-11T12:39:08Z` — wave 2. run: `01KZRC6GR8DZHW7M2ZFT9Q9R3R`. summary: wave 2 complete: work.md/check.md playbook edits for batched flush and context-packet-based lifecycle read; binary rebuilt and docs re-scaffolded from correct repo root after a misdirected first attempt (stray files in cli/ cleaned up, no data lost); go test ./internal/embedded/ -run Drift green, full go test ./... green, doc-links green.
- `2026-08-11T12:41:41.933Z` — handoff recorded. handoff: `01KZRDFJSD77QK47M9EF1EMJ0H`. run: `01KZR291BC9QYFN70AJ82JJHME`. check: `01KZR2NDBVP1VB11H727827DXN`. phase closed. next action: close p3-fewer-round-trips next, then begin p4-observability-and-scope.
- `2026-08-11T12:41:49.417Z` — handoff recorded. handoff: `01KZRDFT39YVSG1BARN8QTAQA8`. run: `01KZRC6GR8DZHW7M2ZFT9Q9R3R`. check: `01KZRDDGB8VZXF0G5ACQZSG5FA`. phase closed. next action: begin p4-observability-and-scope (JSONL invocation logging, skills/workflow/README.md scope declaration) — this is the initiative's final phase and its closure requires a clean check full verdict per handoff.md step 6.
- `2026-08-11T12:44:23Z` — wave 1. run: `01KZRDKNV8SSNFGCQBJ1HKY2DF`. summary: phase p4-observability-and-scope started.
- `2026-08-11T12:46:26Z` — wave 1, task JSONL invocation logging. task_status: `DONE`. run: `01KZRDKNV8SSNFGCQBJ1HKY2DF`. summary: logInvocation wired into Execute (root.go); appends {ts,argv,exit,ms,error_code} to .kit/log/zharness.jsonl, best-effort, rotates to .1 at 1MB; stdout/stderr contracts untouched; go test ./... and smoke test green.
- `2026-08-11T12:47:11Z` — wave 1, task README scope declaration. task_status: `DONE`. run: `01KZRDKNV8SSNFGCQBJ1HKY2DF`. summary: added SDLC Stage Coverage section to skills/workflow/README.md declaring deployment/release/monitoring out of scope per G1, pointing to a future ship skill (G2) if a deploy target materializes; doc-links + grep check green.
- `2026-08-11T12:47:11Z` — wave 1. run: `01KZRDKNV8SSNFGCQBJ1HKY2DF`. summary: wave 1 (only wave) complete: JSONL invocation logging (R8) and README scope declaration (R9) both DONE; go test ./..., go vet ./..., verify-doc-links.sh, and both task-specific smoke checks all green.

## Decisions
<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->
- `2026-08-11T09:31:03Z` — Ship the final-phase check-full requirement as a playbook-only precondition (handoff.md step 6); do not add CLI-side enforcement (a checks.mode column plus handoff record --close-phase validation) in this phase. (phase: `p2-check-routing`), task: evaluate CLI-side enforcement. rationale: CLI-side enforcement needs two things the codebase does not have today: a schema migration adding mode to checks (plus a --mode flag on check record and backfill semantics for pre-migration rows with no mode), and a notion of the initiative's final phase, which does not exist anywhere in the data model -- stories.depends_on is single-parent only, so determining finality means computing DAG-terminal nodes (stories nothing else depends_on) at close time, a case the current schema and handoff.record code path were never designed to answer, and a multi-leaf DAG (independent phase chains) has no defined single final phase to even check against. That is real cross-cutting schema work, not a quick add, and the plans own risk mitigation for this exact task explicitly authorizes stopping here rather than half-implementing. The playbook precondition already written (handoff.md step 6, requiring a check full verdict named explicitly, not inferred from verdict alone) gives the practical protection this requirement needed, at the same enforcement level most of this codebase's cross-phase ordering already relies on (dependency ordering itself is playbook-followed, not hard-blocked by depends_on). Revisit only if the playbook-level discipline proves insufficient in practice -- that would be its own initiative, not a wave-2 add-on to this one..

## Validation
<!-- Append-only durable entries record timestamp, phase, exact command/result/output, run_id, check_id, verdict, and proof_gaps. -->
- `2026-08-11T07:36:31.873Z` — check. verdict: `APPROVED`. check: `01KZQW0SW1FY3CQA9HSJ9T1F6V`. run: `01KZQVJ1X37NPRKF6Z1YDY85KF`. phase: `p1-quick-wins`. judge: `same-session` (claude-sonnet-5).
  - `bash scripts/verify-doc-links.sh` → doc links OK (0 findings)
  - `cd cli && go test ./...` → ok x6 packages, 0 failures
- `2026-08-11T09:32:38.651Z` — check. verdict: `APPROVED`. check: `01KZR2NDBVP1VB11H727827DXN`. run: `01KZR291BC9QYFN70AJ82JJHME`. phase: `p2-check-routing`. judge: `same-session` (claude-sonnet-5).
  - `bash scripts/verify-doc-links.sh` → doc links OK (0 findings)
  - `cd cli && go test ./...` → ok x6 packages, 0 failures
- `2026-08-11T12:40:33.896Z` — check. verdict: `APPROVED`. check: `01KZRDDGB8VZXF0G5ACQZSG5FA`. run: `01KZRC6GR8DZHW7M2ZFT9Q9R3R`. phase: `p3-fewer-round-trips`. judge: `same-session` (claude-fable-5).
  - `cd cli && CGO_ENABLED=0 go build ./...` → gate proof: build clean
  - `cd cli && go vet ./...` → gate proof: vet clean
  - `cd cli && go test ./...` → gate proof: all 6 packages pass
  - `bash scripts/verify-doc-links.sh` → gate proof: doc links OK (0 findings)
  - `cd cli && go test ./internal/embedded/ -run Drift` → gate proof: embedded playbooks match scaffolded docs
- `2026-08-11T12:48:30.724Z` — check. verdict: `APPROVED`. check: `01KZRDW204N3EPF4AH8APAZK3K`. run: `01KZRDKNV8SSNFGCQBJ1HKY2DF`. phase: `p4-observability-and-scope`. judge: `same-session` (claude-fable-5).
  - `cd cli && CGO_ENABLED=0 go build ./...` → full-mode proof: build clean
  - `cd cli && go vet ./...` → full-mode proof: vet clean
  - `cd cli && go test ./...` → full-mode proof: all 7 packages pass (incl. new invocation_log_test.go)
  - `bash scripts/verify-doc-links.sh` → full-mode proof: doc links OK (0 findings)
  - `grep -qi "out of scope" skills/workflow/README.md` → full-mode proof: R9 scope paragraph present

## Current State and Next Action
- active_phase: p4-observability-and-scope (checked via `check full` — same-session judge, APPROVED; ready for closing handoff)
- lifecycle_status: checked
- latest_run_id: 01KZRDKNV8SSNFGCQBJ1HKY2DF
- latest_trace_ids: [01KZRDMGTHRPA3KE8XXSZZC50B, 01KZRDR87G0KA22E2D1F1PGAQC, 01KZRDSMDJZY56P3HRWP8V8JQ6, 01KZRDSMDW7W1R1KFCZVVHJF73]
- latest_check_id: 01KZRDW204N3EPF4AH8APAZK3K
- latest_handoff_id: 01KZRDFT39YVSG1BARN8QTAQA8
- blockers: none
- open_items: none
- exact_next_action: run closing `handoff record --close-phase` for p4-observability-and-scope (run 01KZRDKNV8SSNFGCQBJ1HKY2DF, check 01KZRDW204N3EPF4AH8APAZK3K, a recorded `full`-mode APPROVED verdict) — this closes the whole initiative; move this plan to `docs/plans/completed/` after
