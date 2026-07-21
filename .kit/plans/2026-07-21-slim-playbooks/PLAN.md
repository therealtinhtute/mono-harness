# PLAN: Slim Playbooks (Track A) — move fat to CLI, cut per-run token load

**Date:** 2026-07-21
**Origin:** Architecture audit §5/§6/§11 (deferred in Harness Subtraction Pass SPEC) + user feedback 2026-07-21 (sessions too long / too many tokens). Un-defers audit #5.
**Lane:** high-risk (public CLI contract, embedded-playbook rewrites, multi-domain). Proof matrix = unit + integration + command-output + manual-check per phase.

## Locked decisions (interview 2026-07-21)

| # | Fat block | Decision |
|---|---|---|
| 1 | Artifact templates (run/check/handoff/spec skeletons, ~200+ lines across playbooks) | **Move to CLI scaffold** |
| 2 | watzup Output Contract + Examples (~184 lines) | **Render in CLI** (`zharness resume` emits formatted recap) |
| 3 | work mode-resolution + detection tables (~44 lines) | **Move to CLI** (`zharness next`) |
| 4 | Simple mode + FK carve-out prose | **Keep mode, compress carve-out** |

## Goal & success signal

Cut per-run playbook prose ~30–40% (audit §11 estimate) by moving deterministic templates/format/routing out of markdown-loaded-every-invocation into the Go binary (loaded zero tokens, emitted on demand). Target lengths: `check` <200, `work` <180, `watzup` <160. Every guarantee preserved: replay, traceability, lane×proof gate, the recap format.

## CLI additions (exact shapes — no placeholders)

### 1. `zharness scaffold <kind> --path <file> [--slug X] [--lane L] [--mode full|simple] --json`
- `kind ∈ {run, check, handoff, spec}`. Writes the artifact **skeleton** (frontmatter + section headers + one-line field hints as comments) to `--path`; the agent then fills it. Idempotent-safe: refuses to overwrite a non-empty existing file (returns `file_exists`).
- Fold `run` into existing `run create`: `run create` already registers the row + sets `latest_run_id`; extend it to also emit the run skeleton to `--artifact-path` in the same call (so one command registers **and** scaffolds). `check`/`handoff`/`spec` are new `scaffold` subcommands (no DB row on scaffold; the row is still written by `check record` / `handoff record` / `intake` as today).
- Template source moves from playbook prose → Go embed (`cli/docs/embedded/templates/*.md`).

### 2. `zharness resume` gains text rendering
- Default human output = the Recap format (Title / Trạng thái / Context / Thay đổi / Risks table / Next, Vietnamese, ≤25 lines, forbidden-phrase-safe **by construction**). `--json` keeps the current machine shape.
- Renders the **harness-state** block (readiness, drift→recovery, handoff, phase/run/check chain) fully. The **git/WIP** block (branch position, uncommitted-diff analysis, change themes) stays agent-gathered — resume has no git access. watzup keeps git+WIP gathering, drops the ~102-line Output Contract for the harness block + all ~82 lines of Examples. Realistic cut: ~130–150 lines.

### 3. `zharness next --json`
- Reads state (folds current `query state`/`query phases` logic) → returns `{mode, active_phase, stop?: {code, message, recovery}}`. Encodes work.md's Mode-Resolution + Full-Mode-Detection tables + stop-message shapes in Go. work.md drops ~54 lines; agent runs `zharness next` and follows its output.

### 4. Simple-mode carve-out compression
- The FK-constraint explanation (~15 lines in work step 2 + check branches) is already **enforced** by the CLI (`run create` rejects simple; `check record` skips simple). Prose collapses to one line + a `CONTRACT.md` ref. `zharness next` returns `mode: simple` with the carve-out handled.

## Phases (each independently mergeable — ships a slimmer, working state)

### Phase S1 — scaffold (biggest win)
- Build `zharness scaffold` + extend `run create` to emit skeleton. Move templates → `cli/docs/embedded/templates/`.
- Strip templates from `work`, `check`, `handoff`, `brainstorm` playbooks → replace with a `zharness scaffold ...` call + a 3-line "fill these sections" note.
- Verify: unit (scaffold emits valid frontmatter, refuses overwrite), integration (scaffold→fill→`validate` passes), command-output (`go build/test`), manual-check (review pass).

### Phase S2 — `zharness next` + work slim
- Build `zharness next`; strip work.md routing tables + compress simple-mode carve-out.
- Verify: unit (next resolves each state row in the old table correctly — port the table into test cases), integration (next against a real .kit fixture), command-output, manual-check. **Invariant:** every routing outcome the prose produced, `next` must reproduce (table→test parity).

### Phase S3 — resume text render + watzup slim
- Add resume text rendering; strip watzup Output Contract (harness block) + Examples.
- Verify: unit (render matches format contract: 25-line cap, risk table columns, no forbidden `git ` substring), integration (render against fixtures for each readiness state), command-output, manual-check.

Order by impact: **S1 → S2 → S3** (independent; any can ship alone).

## Sequencing with the in-flight Harness Subtraction Pass (recommendation)

Prerequisite for **all** of the above — deploy Phase 1 first (from the 2026-07-21 state-readiness check):
1. Replay the 23 pending changesets into local `harness.db`.
2. Rebuild + reinstall `zharness` so the installed binary has `run create` (playbooks about to gain *more* CLI calls need a current binary).

Then recommended global order:
`Phase 2 dead-surface-removal` → `Phase 3 scoring-removal` (edits check.md — do before S1's check.md strip to avoid double-edit) → **`S1 → S2 → S3`** → `Phase 4 single-source-playbooks` **last** (it projects the *final* slim text + adds the drift-guard test).

## Fragile assumption (premise collapse)

This plan assumes CLI-emitted output is what the agent needs **verbatim**. If a context needs a non-standard artifact shape, a rigid template is less flexible than prose. Mitigation: `scaffold` emits a **skeleton the agent still fills and may extend**, not a locked final output — flexibility preserved.

## Rollback

Every phase = playbook-text + one CLI command, revertable independently via git. No schema change, no data migration. `scaffold`/`next`/`resume-text` are additive commands; reverting a playbook to the embedded template is a text-only change.

## Non-scope

- Dropping SQLite / in-memory fold (audit §5 Option B) — still deferred.
- Memory unification (audit §4) — separate initiative.
- `interview`→`brainstorm --grill` (audit §15) — not here.
- references/ (1449 lines) pruning — lower per-run impact (on-demand); default-prune recommended but tracked separately, not in these phases.
