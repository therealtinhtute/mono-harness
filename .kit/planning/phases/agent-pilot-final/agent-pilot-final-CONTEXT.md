# Context: agent-pilot-final

Phase: agent-pilot-final
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: low
Expected Proof: e2e, manual-check

## Goal
Run the final genuinely cold Codex pilot against the autonomous-entry-parity release and close R9 only if all conjunctive conditions pass: zero SKILL.md reads, zero harness-mechanics coaching/questions, task tests pass, lifecycle artifacts/DB links are complete, and `zharness validate --json` returns `valid:true`.

## Scope Boundary
### Allowed Surfaces
- Fresh scratch target outside this repo
- Isolated temporary HOME/CODEX_HOME containing authentication only
- Protocol/transcript files outside the target repo
- `docs/workflow-harness/pilot-evidence/{date}-agent-pilot-final.md`
- Read-only everywhere else

### Forbidden Surfaces
- `cli/**`, `skills/workflow/**`, and this repo's live `.kit/`
- Continuing a stopped session with procedural approval — any such need is a pilot FAIL
- Reusing any prior attempt's target directory

## Spec Hooks
- R9 verbatim: second-agent lifecycle pass, no SKILL.md reads, `zharness validate --json` passes

## Locked Decisions
- Runtime: Codex CLI, isolated with fresh HOME/CODEX_HOME (auth.json only), `--ignore-user-config --ignore-rules --ephemeral`.
- Target contains only fresh `git init` + `zharness init` output before execution. Prompt and transcript live outside it.
- Protocol remains root `AGENTS.md` pointer + a real product task + "Work autonomously end-to-end; ask only for a product decision." No harness command, mode, threshold, or phase hint.
- The task explicitly requires six product/test files: `inventory/{__init__,models,store,report}.py`, `test_store.py`, `test_report.py`. This naturally exceeds work.md's ≤5-file simple-mode guard, so the docs themselves must route the agent through `to-plan` + full mode; the prompt never explains why six files matter.
- Inspect transcript for any `/SKILL.md`, `.agents/skills`, `.claude/skills`, or `.codex/skills` access before accepting it.
- Grade literally; passing code alone is insufficient. The final DB/artifact trail must include an intake row, at least one story, registered full-mode RUN, at least one linked trace, registered CHECK, zero pointer drift/unlinked proofs, and `validate:true`.

## Assumptions
- Account authentication may still affect tone/model defaults, but isolated filesystem controls remove local skills/config/rules/memories from the runtime's reachable HOME.
- Root AGENTS.md path bug #37 may be hit and self-recovered; this is known and does not require coaching.

## Canonical Refs
- `docs/workflow-harness/pilot-evidence/2026-07-19-second-agent-pilot.md`
- GitHub #38, #39, and #40
- autonomous-entry-parity release proof (`cli/v0.4.1`)

## Rejected Options
- Reusing Phase 8 attempt 1/2: invalid because global SKILL.md files were read.
- Accepting Phase 8 isolated attempt 3: invalid because it stopped for procedural confirmation before producing artifacts.

## Deferred Ideas
- Third-runtime pilot (Cursor) — R9 requires one independent runtime, not multiple.

## Escalate If
- Any SKILL.md read, procedural coaching request, failed task verification, or `valid:false` result → NO-GO and route a new finding without hotfixing this phase.
