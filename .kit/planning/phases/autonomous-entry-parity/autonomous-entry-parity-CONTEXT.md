# Context: autonomous-entry-parity

Phase: autonomous-entry-parity
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: medium
Expected Proof: unit, integration, e2e, command-output

## Goal
Resolve GitHub #39 and #40. A fully isolated Codex run first stopped because `brainstorm.md` mandates unconditional procedural confirmations; after that docs fix, the required dry pilot proceeded autonomously (zero SKILL.md reads, task tests passed) but fabricated an invalid RUN ULID because `work.md` requires agents to mint IDs without providing an exact command. Make explicit execution intent durable across unambiguous transitions and add a non-mutating `zharness id` helper so cold agents can produce valid artifact/change-set IDs without language-specific dependencies.

## Scope Boundary
### Allowed Surfaces
- `cli/docs/embedded/playbooks/brainstorm.md`, `cli/docs/embedded/playbooks/work.md`
- `cli/docs/CONTRACT.md`
- `cli/internal/embedded/embedded_test.go` (focused content-contract tests)
- `cli/internal/interfaces/` root registration + focused `id` command/test files
- CLI release surface (tag → goreleaser, install verification)
- `skills/workflow/README.md` and the 6 spine `SKILL.md` version literals

### Forbidden Surfaces
- Other embedded playbooks/docs unless a test proves a direct contradiction
- DB schema or mutating application behavior (`zharness id` is pure/non-mutating)
- `skills/workflow/interview/**` and `skills/workflow/git/**`
- GitHub #37's root-shim path defect — separate issue, not blocking because cold agents self-recover

## Spec Hooks
- R4: each playbook must be self-sufficient for a competent agent
- R9: a non-Claude agent completes the lifecycle using only written docs + CLI, with no SKILL.md reads and passing validation
- Constraint: preserve agent-agnostic plain Markdown, no runtime-specific syntax

## Locked Decisions
- Explicit execution intent is durable authorization only when the prompt clearly requests implementation/completion/end-to-end execution and already defines scope + success criteria.
- Step 2 asks only for genuinely ambiguous mode or scope. A fully specified change request with explicit autonomous execution intent resolves without a procedural confirmation.
- Step 9 may continue past SPEC self-review without another response only when the original request grants end-to-end execution, no unresolved product choice remains, and no destructive/outward-facing action is implied. The agent still reports the SPEC path and decision summary.
- Preserve confirmation for interactive brainstorm requests, ambiguous intent, replacement of an existing SPEC, unresolved product decisions, destructive actions, and outward-facing effects.
- Add focused embedded-doc tests asserting both halves: autonomy carve-out present/ambiguity gate preserved, and `work.md` names the exact ID command. Do not rely on prose review alone.
- Add non-mutating `zharness id`: plain output is the ULID plus newline; `--json` is `{"id":"..."}`. It uses the same `ulid.Make().String()` implementation as existing entity writers, touches no DB/files, and is registered on the root command.
- `work.md` must call `zharness id --json` before writing a RUN artifact and again for each manually-authored changeset filename; never suggest a language-specific generator.
- `CONTRACT.md` documents `id` as a non-mutating Core helper with no arguments and no domain side effects.
- Release as `cli/v0.4.0` (documented playbook/CLI contract change, same minor-bump reasoning used for v0.3.0) and bump MIN_ZHARNESS_VERSION to 0.4.0 across the 6 spine triggers + README.

## Assumptions
- The isolated Codex attempt's stop is entirely caused by `brainstorm.md` lines 16/52/65/112; it had already found `.kit/docs`, read only embedded docs, and asked exactly the prohibited procedural question.
- The Phase 9 dry pilot proves the autonomy wording works (no question, zero SKILL reads, implementation/tests complete) and isolates #40: its only hard validation failure is the fabricated RUN ID `01K0A1RUNDEF23456789GHJKMN`.
- The unchanged final protocol contains explicit autonomous end-to-end intent, so both fixes are exercised naturally without harness-mechanics coaching.
- **Pattern-completeness correction (gate pass):** CONTRACT.md already names `brainstorm`, `to-plan`, `work`, and `check` as consumers of the ID helper, but the first implementation documented only `work.md`. A pre-approval sweep found the same manual-ID shape in SPEC frontmatter (`brainstorm.md`) and meta changeset filenames (`to-plan.md`, `check.md`). These three direct contradictions are now in scope under Allowed Surfaces' explicit exception; this is the same #40 pattern, not a new feature.

## Canonical Refs
- GitHub #39, #40
- `cli/docs/embedded/playbooks/brainstorm.md`, `work.md`
- `cli/docs/CONTRACT.md`
- `cli/internal/embedded/embedded_test.go`, `cli/internal/interfaces/root.go`
- Phase 8 isolated transcript: `/private/tmp/.../scratchpad/agent-pilot-rerun-cold-3-transcript.log`
- Phase 9 dry transcript: `/private/tmp/.../scratchpad/autonomous-entry-dry-pilot-transcript.log`

## Rejected Options
- Coach the pilot with "approve/continue": rejected — Phase 8 treats harness-mechanics coaching as FAIL.
- Remove all brainstorm confirmation gates: rejected — would permit silent scope/replacement decisions and violate user-control safeguards.
- Change the pilot prompt to mention brainstorm mode: rejected — contaminates the test instead of fixing the docs.
- Embed a Python/Node/shell ULID generator in `work.md`: rejected — adds runtime/dependency assumptions and contradicts the agent-agnostic exact-CLI-command goal; the Go binary already owns ULID generation.
- Reuse a story/intake ID as a RUN ID: rejected — conflates entity identity and still gives simple mode no reliable source.

## Deferred Ideas
- Fix root `AGENTS.md` relative links (#37) in its own scoped cycle.

## Escalate If
- The carve-out cannot be expressed without changing AUTHORITY.md semantics → stop and refresh this plan rather than expanding silently.
- `zharness id` would require DB state or mutating behavior → stop; it must remain a pure helper.
- Repeated dev dry pilot still asks a procedural question, reads SKILL.md, fabricates an ID, or returns `valid:false` → route a new finding; do not release.
