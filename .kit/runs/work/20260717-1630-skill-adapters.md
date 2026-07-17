# COOK RUN

Run ID: work-20260717-1630-skill-adapters
Mode: full
Status: passed
Spec: .kit/planning/SPEC.md
Roadmap: .kit/planning/ROADMAP.md
Workflow State: .kit/workflow-state.yml
Phase: skill-adapters
Plan: .kit/planning/phases/skill-adapters/skill-adapters-PLAN.md
Started At: 2026-07-17 16:30

## Preflight
- scope drift: no — `skills/workflow/brainstorm`, `skills/workflow/to-plan`, `skills/workflow/work` working trees only carry pre-existing modifications inside Allowed Surfaces (`brainstorm/references/spec-template.md`, `work/references/run-artifact-template.md` — both already modified before this phase started, both inside this phase's own Allowed Surfaces)
- working tree note: pre-existing out-of-scope leftovers from harness-concept phase (README.md, skills/workflow/README.md, docs/workflow-harness/) remain untouched, same as noted in cli-core/cli-domain's run artifacts
- required artifacts present: yes — skill-adapters-CONTEXT.md and skill-adapters-PLAN.md both read in full; `cli/docs/CONTRACT.md` (canonical ref) confirmed exists and read in full
- selected phase: skill-adapters (2 waves, 4 tasks: T1 brainstorm rewrite, T2 to-plan rewrite, T3 work rewrite, T4 sample chain run + validate)

## Wave / Task Log
### Wave 1
#### T1 — Rewrite brainstorm SKILL.md
- status: DONE
- changed files:
  - skills/workflow/brainstorm/SKILL.md (added `<version-gate>` block after `<security>`; step 7 now fires `zharness init`/`zharness intake` and writes `intake_id` into SPEC frontmatter; Output Rules updated)
  - skills/workflow/brainstorm/references/spec-template.md (added `intake_id` frontmatter field + Rules bullet)
- verification:
  - `bash scripts/validate-skill.sh skills/workflow/brainstorm/SKILL.md` → PASS (frontmatter/structure/security/anti-pattern checks all ✅)
  - `grep -c 'zharness' skills/workflow/brainstorm/SKILL.md` → 3 (≥2 required)
  - `grep -n 'workflow-state' skills/workflow/brainstorm/SKILL.md` → empty, as required
- notes:
  - plan's literal verification command (`bash scripts/validate-skill.sh skills/workflow/brainstorm`, directory form) doesn't match the script's actual argument (`$1` is the SKILL.md file path, confirmed by reading the script) — used the file-path form instead, consistent with how CONTEXT.md's own Locked Decisions line documents this same script having already been verified against the file path
  - real gap found and resolved: CONTRACT.md's `intake` command hard-requires an existing db (`db_unreadable` if absent) and CONTEXT.md's Locked Decisions only assign "`zharness init` if no db" to `to-plan`, but brainstorm runs chronologically *before* to-plan — on a fresh project there is no db yet when SPEC lock fires. Resolved by having brainstorm's own lock step call `zharness init` (idempotent) immediately before `zharness intake`, same pattern `to-plan` uses — logged in `.kit/implementation-notes.md`
  - `MIN_ZHARNESS_VERSION` didn't exist anywhere yet; set to `0.1.0` (matches cli-core's deferred tag) and the gate treats local `dev` builds as always passing, since no tagged release exists yet during this initiative's own dogfooding — logged in `.kit/implementation-notes.md`

#### T2 — Rewrite to-plan SKILL.md
- status: DONE
- changed files:
  - skills/workflow/to-plan/SKILL.md (added `<version-gate>` block after `<security>`; Step 2 now runs `zharness init`/`zharness story` per phase instead of initializing workflow-state.yml; Step 5 renamed "State integrity + handoff guidance" and now verifies via `zharness query state --json` / `zharness query phases --json`; Output Format section's two workflow-state.yml mentions removed; References list's `workflow-state-template.yml` bullet removed)
  - skills/workflow/to-plan/references/roadmap-template.md (not in T2's declared `touches` list, edited anyway — see notes)
- verification:
  - `grep -n 'workflow-state.yml' skills/workflow/to-plan/SKILL.md` → empty, as required
  - `grep -c 'zharness' skills/workflow/to-plan/SKILL.md` → 4 (≥3 required)
  - `bash scripts/validate-skill.sh skills/workflow/to-plan/SKILL.md` → PASS (all checks ✅, matching T1's lint convention even though not explicitly listed in T2's verification line)
- notes:
  - scope extension beyond declared `touches`: `to-plan/references/roadmap-template.md` had two of its own literal "workflow-state" mentions ("also initialize `.kit/workflow-state.yml` using `workflow-state-template.yml`" and "set `entry_phase` and `current_phase` in `.kit/workflow-state.yml`") that directly contradicted the just-rewritten SKILL.md if left as-is — a template instructing the harness-CLI-first skill to still write the retired yml file. Edited both to reference `zharness init`/`zharness story` and `zharness query state` instead. Judged this as necessary contradiction-avoidance, not scope creep, since the alternative (leaving it untouched) would ship a self-contradicting skill — logged in `.kit/implementation-notes.md`
  - `workflow-state-template.yml` itself (the reference file) was NOT deleted — CONTEXT.md's Scope Boundary explicitly defers that deletion to Phase 8; to-plan/SKILL.md's References list no longer points to it, so it's now an orphaned-but-present file until Phase 8 removes it

#### T3 — Rewrite work SKILL.md
- status: DONE
- changed files:
  - skills/workflow/work/SKILL.md (added `<version-gate>` block after `<security>`; Execution Loop step 1 now uses `zharness query state`/`query phases`; step 2 registers the run in the harness via a hand-authored "run" changeset + `db changeset apply`; new step 9 fires `zharness trace add` at wave completion; old step 9 "Workflow-state update" replaced with step 10 "State check" via `zharness query state`; Output Rules' workflow-state.yml mention replaced)
  - skills/workflow/work/references/run-artifact-template.md (dropped `Workflow State:` header line; rule bullet rewritten from "update workflow-state.yml" to "run `zharness trace add` per wave, append id to `trace_ids`")
- verification:
  - `grep -n 'trace add' skills/workflow/work/SKILL.md` → matches step 9
  - `grep 'workflow-state' skills/workflow/work/SKILL.md` → empty, as required
  - `bash scripts/validate-skill.sh skills/workflow/work/SKILL.md` → PASS WITH WARNINGS (1 warning: 156/150 lines, token-efficiency soft target; not a T3 verification criterion, left as-is rather than trimming necessary instructions)
- notes:
  - real blocking gap found and resolved: CONTRACT.md's ported command set has no CLI action that creates a "runs" table row for a live (non-legacy-import) project — `trace add --run-id` only validates an *existing* run row (`unknown_run_id` if given but not found), it never creates one; `check record --run-id` likewise requires a pre-existing row. Confirmed via reading `cli/internal/application/trace.go`'s `CreateTrace` (no insert path) and `cli/internal/interfaces/validate.go`/`validate.go`'s `rowExists(db, "runs", id)` check, which would flag every freshly created RUN artifact as `stale_pointer` (failing `zharness validate` exit-0, a hard requirement for T4). Resolved by having `work` author a one-line changeset (`entity: "run", op: "create"`) itself and apply it via the already-shipped, generic `db changeset apply <path>` command (Phase 3/cli-core, not a cli/** change) — confirmed this is the actually-intended mechanism by finding `cli/docs/SCHEMA.md` line 121 already names `work` (alongside `trace add`) as the producer of `run`-entity changesets, so this is implementing a documented-but-unwired responsibility, not inventing new behavior. No cli/** files touched — stays inside skill-adapters' Allowed Surfaces. Logged in `.kit/implementation-notes.md`.
  - did not touch the pre-existing RUN-frontmatter `plan_id` gap (PLAN artifacts carry no id at all, so `plan_id` can only ever be a syntactically-valid ULID with nothing real to link to) — already flagged in cli-domain's own implementation-notes entries and CONTRACT.md's Known Gap; out of scope for T3, left exactly as the existing template already had it
  - **amended during T4** (see below): step 2's run-registration changeset grew a second JSONL line (`meta.latest_run_id` update) after T4's dry-run surfaced that gap too

### Wave 2
#### T4 — Sample chain run + validate
- status: DONE
- changed files:
  - skills/workflow/to-plan/SKILL.md (Step 2 — added the `current_phase`/`entry_phase` meta-changeset instruction; see notes)
  - skills/workflow/work/SKILL.md (Step 2 — run-registration changeset now carries a second line updating `meta.latest_run_id`; see notes)
  - no repo `.kit/planning/` or `cli/**` files touched — all dry-run artifacts live under a scratch project outside the repo (`/home/tinhpt/.claude/jobs/a0eb0eaa/tmp/scratch-project/`)
- verification:
  - built `zharness` from `cli/` (`go build -o {scratch}/zharness ./cmd/zharness`) → clean build; `{scratch}/zharness --version` → `zharness version dev` (confirms the version-gate's dev-passes clause)
  - ran the dry-run chain manually (SPEC.md written per spec-template.md → `zharness init`/`intake`/`story` → `zharness db changeset apply` for the meta current_phase/entry_phase fix → RUN artifact written per run-artifact-template.md → `zharness db changeset apply` for the run-registration two-line changeset → dummy T1 task (`echo > dryrun-proof.txt`, `test -f` → pass) → `zharness trace add --wave 1 ... --run-id {run id} --json` → trace id appended to RUN frontmatter)
  - `zharness validate --json` → `{"valid":true,"findings":[{"link":"SPEC->PLAN","issue":"not_yet_implemented",...}]}`, exit 0 — the one finding is the pre-existing, already-documented SPEC->PLAN gap (harness-contracts/T2), not a new one
  - `zharness query state --json` → `{"current_phase":"dry-run-phase","entry_phase":"dry-run-phase","latest_run_id":"{run id}","latest_check_id":null,...}` — all pointers correctly populated
  - `zharness resume --json` → `{"readiness":"in-progress","drift":[],...}`
  - `find . -name workflow-state.yml` (scratch project root) → empty, confirming zero yml writes anywhere in the simulated chain
- notes:
  - **Gap #2 found and resolved (to-plan)**: same root cause as T3's run-row gap — no live command sets `meta.current_phase`/`entry_phase` either (only legacy `import` does); `zharness story` alone left `query state` reporting nulls. Resolved by having `to-plan`'s Step 2 author a one-line meta-update changeset once all stories are created, applied via the same generic `db changeset apply`. Logged in `.kit/implementation-notes.md`.
  - **Gap #3 found and resolved (work)**: `meta.latest_run_id` also never gets set live — and cli-domain's own T3 notes had already flagged this exact pointer-maintenance gap and pre-assigned it to `skill-adapters`/`work`. Confirmed the changeset engine applies every JSONL line in a file within one transaction before advancing the fence, so folded a second line (`update meta.latest_run_id`) into `work`'s existing run-registration changeset rather than issuing a separate file — one atomic write, no lag between a run existing and being pointed at. Logged in `.kit/implementation-notes.md`.
  - re-ran `bash scripts/validate-skill.sh skills/workflow/work/SKILL.md` after the Gap #3 edit → still PASS WITH WARNINGS (156/150 lines, same pre-existing soft warning, no new failures)
  - did not touch the still-open, pre-existing `plan_id` cross-link gap or the deferred `workflow-state-template.yml` file deletion (both already logged, both explicitly out of scope for skill-adapters)

## Summary
- passed tasks: T1, T2, T3, T4 — all 4 tasks in skill-adapters DONE
- blocked tasks: none
- unresolved concerns: `plan_id` cross-link gap (pre-existing, flagged for a future phase); `workflow-state-template.yml` file deletion deferred to Phase 8 per CONTEXT.md's own Scope Boundary; doc-debt items carried forward from prior phases (Security note in SCHEMA.md/CONTRACT.md, `query phases`/`query artifacts --json` shape lock, Error Codes table amendments)

## Next Recommended Action
- `check full` — gate skill-adapters' diff before advancing to `validation-gate`

