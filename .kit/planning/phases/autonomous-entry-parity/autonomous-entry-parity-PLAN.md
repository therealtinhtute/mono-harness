# Plan: autonomous-entry-parity

Phase: autonomous-entry-parity
Status: ready
Wave Count: 3
Execution Owner: work
Updated At: 2026-07-19

## Goal
Resolve #39 and #40: explicit unambiguous end-to-end execution intent must cross brainstorm's procedural gates without coaching, and cold agents must have an exact CLI command to mint valid RUN/changeset ULIDs. Ship both in cli/v0.4.0 only after an isolated dev dry pilot reaches `valid:true`.

## Inputs
- `autonomous-entry-parity-CONTEXT.md`
- GitHub #39, #40
- `cli/docs/embedded/playbooks/{brainstorm,work}.md`
- `cli/docs/CONTRACT.md`
- `cli/internal/interfaces/root.go`
- `cli/internal/embedded/embedded_test.go`

## Wave 1 — DONE before plan refresh
### T1 — Add explicit-intent carve-out to brainstorm.md
- status: DONE
- type: docs
- touches: `cli/docs/embedded/playbooks/brainstorm.md`
- output: Step 2/9 and review/exit wording now continue on qualifying explicit execution intent while preserving ambiguity/replacement/product/destructive/outward-facing gates
- verification: targeted grep of every confirm/approval/ask/wait occurrence → internally consistent

### T2 — Add embedded autonomy content-contract test
- status: DONE
- type: test
- touches: `cli/internal/embedded/embedded_test.go`
- output: test asserts autonomy phrases + preserved ambiguity gates + obsolete unconditional sentence absent
- verification: `go test ./internal/embedded/... -v` → pass (6/6)

## Wave 2
### T3 — Add non-mutating `zharness id`
- type: implementation
- inputs: existing `ulid.Make().String()` conventions in application writers
- touches:
  - `cli/internal/interfaces/root.go`
  - new `cli/internal/interfaces/id.go`
  - focused interface/command test file
- avoid: DB/application mutation, migrations, entity rows
- steps:
  1. Add root subcommand `id`, no positional args.
  2. Mint with `ulid.Make().String()`.
  3. Plain output: ULID + newline. JSON: `{"id":"..."}` via existing global `--json` behavior.
  4. Tests assert both output modes, parse/shape, no args, and two calls differ.
- expected outputs: exact cross-platform ID primitive available to all playbooks/agents
- verification: `cd cli && go test ./internal/interfaces/... -run TestID -v`
- stop if: command touches DB/files or needs new dependency
- escalate to: to-plan phase autonomous-entry-parity

### T4 — Document exact ID flow
- type: docs
- inputs: T3 command contract
- touches:
  - `cli/docs/CONTRACT.md`
  - `cli/docs/embedded/playbooks/work.md`
  - `cli/internal/embedded/embedded_test.go`
- avoid: unrelated playbooks
- steps:
  1. Add `id` to CONTRACT.md as a non-mutating Core helper with plain/JSON shapes.
  2. In work.md Step 2, run `zharness id --json` before writing the RUN and use the returned id in frontmatter/registration.
  3. Before every manually-authored changeset filename, run `zharness id --json` again; never reuse the RUN id as the changeset id.
  4. Add embedded content assertion that work.md names `zharness id --json` and distinguishes RUN vs changeset IDs.
- expected outputs: cold agent has no placeholder/generator ambiguity
- verification: `go test ./internal/embedded/... -v`; targeted grep/read
- stop if: docs imply ID command mutates state
- escalate to: to-plan phase autonomous-entry-parity

### T5 — Full verification + repeated isolated dev dry pilot
- type: test
- inputs: T1-T4
- touches: fresh scratch target + isolated temporary HOME only
- avoid: this repo's live `.kit/` during pilot
- steps:
  1. `go build ./...`, `go test ./...`, `go vet ./...`, `gofmt -l .`.
  2. Build `dev` binary; scaffold a brand-new target.
  3. Run Codex with auth-only HOME, `--ignore-user-config --ignore-rules --ephemeral`; prompt/transcript outside target; prepend dev binary to PATH.
  4. Independently verify task tests, transcript zero SKILL reads/questions, `validate:true`, resume no drift, audit no pointer drift/unlinked proofs.
- expected outputs: complete uncontaminated dry pass
- verification: verbatim outputs + transcript grep
- stop if: any conjunct fails — file/route finding, do not release
- escalate to: to-plan phase autonomous-entry-parity

## Wave 3
### T6 — Release cli/v0.4.0
- type: implementation
- inputs: T5 clean
- touches: git tag, GitHub release
- avoid: CI edits unless directly required by release failure
- steps:
  1. Commit/push code/docs + planning/run evidence.
  2. Push `cli/v0.4.0`; watch goreleaser to completion.
  3. Verify release assets and fresh installer resolution.
- expected outputs: published v0.4.0 matrix
- verification: `gh release view v0.4.0`; `install-zharness.sh`; `zharness --version`
- stop if: pipeline needs structural rework
- escalate to: to-plan phase

### T7 — Bump MIN_ZHARNESS_VERSION to 0.4.0
- type: docs
- inputs: T6 published
- touches: `skills/workflow/README.md`, 6 spine `SKILL.md` files
- avoid: `interview/SKILL.md`, `git/**`
- steps:
  1. Replace 0.3.0 with 0.4.0 in README gate text and six spine literals.
  2. Verify seven intended occurrences; interview unchanged.
- expected outputs: spine rejects CLIs lacking #39/#40 fixes
- verification: targeted `grep -rn MIN_ZHARNESS_VERSION`
- stop if: any non-spine skill would change
- escalate to: none

## Risks / Watch-fors
- IDs are unique data, not semantic aliases: RUN id and changeset filename id must be minted separately.
- The dry pilot is the release gate because both defects are cross-agent interpretation failures, not just code syntax.
