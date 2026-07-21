# Context: write-boundary

Phase: write-boundary
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: medium
Expected Proof: unit, integration, command-output, manual-check

## Goal
Give the CLI commands to own every harness write the playbooks currently hand-author: run registration (+ `latest_run_id`) and the `latest_check_id` meta pointer. After this phase, no playbook instructs the LLM to write a `.changeset.jsonl` file.

## Scope Boundary
### Allowed Surfaces
- `cli/internal/interfaces/` — new `run create` command wiring; extend `check record` flags/behavior
- `cli/internal/application/` — run-create logic; check-record meta-pointer logic
- `cli/internal/infrastructure/changeset.go` — reuse existing `WriteChangeset`/`ApplyChangeset` (do not change format)
- `cli/docs/embedded/playbooks/work.md` — replace step-2 full-mode hand-authored changeset with `run create`
- `cli/docs/embedded/playbooks/check.md` — replace step-4 hand-authored meta changeset with `check record` default
- `cli/docs/CONTRACT.md` — document new command shapes
- tests under `cli/internal/**`

### Forbidden Surfaces
- `.kit/docs/playbooks/*` directly — those are generated; edit the embed, re-scaffold via `init` (projection enforcement is Phase 4)
- changeset JSONL on-disk format, ULID/fence/replay mechanics
- `score.go` / `audit.go` (Phases 2–3)
- migrations schema (Phase 2)

## Spec Hooks
- Requirement R1 (write-boundary), R5 (playbook updates)
- Constraint: simple-mode must keep skipping DB registration (FK on `runs.story_slug`)
- Acceptance: `grep` finds no hand-author-changeset instruction in playbooks; `run create` sets `latest_run_id` atomically; `check record` sets `latest_check_id`.

## Locked Decisions
- `run create` writes ONE changeset with two lines (create run + update `meta.latest_run_id`) and applies it in a single transaction — same semantics work.md did by hand, now owned by Go.
- `run create` is **full-mode only**; it requires an existing `story_slug`. Simple mode still writes no run row (documented carve-out stays).
- `check record` sets `meta.latest_check_id` by **default** (Open Question resolved toward default; a `--no-set-latest` escape hatch is optional, not required).

## Assumptions
- `WriteChangeset`/`ApplyChangeset`/`db changeset apply` already provide everything needed; this phase adds thin command wrappers, not new infra.
- The run artifact markdown file is still written by `work` (the CLI command only handles the DB/changeset side).

## Canonical Refs
- `.kit/planning/SPEC.md`, `.kit/planning/ROADMAP.md`
- `cli/internal/infrastructure/changeset.go` (WriteChangeset, ApplyChangeset, entityColumns)
- `.kit/docs/playbooks/work.md` step 2 (the hand-authored block being replaced)
- `.kit/docs/playbooks/check.md` step 4 (the meta changeset being replaced)
- existing `zharness story` / `trace add` commands as the pattern for a mutating command

## Rejected Options
- A generic `zharness meta set latest_run_id=...` — too low-level, re-exposes the footgun. Prefer intent-named `run create`.
- Making `run create` also write the markdown artifact — mixes concerns; the skill owns the human-readable file.

## Deferred Ideas
- A unified `zharness run|check` lifecycle command surface (later polish)
- `zharness next` front door (out of this initiative)

## Escalate If
- `run create`'s two-line semantics can't be reproduced without changing changeset format → route to `to-plan` (rescope) — do not silently alter the format.
- Simple-mode FK carve-out turns out to need a schema change → escalate; it's out of this phase's boundary.
