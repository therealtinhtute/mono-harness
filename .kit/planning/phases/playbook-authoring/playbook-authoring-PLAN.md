# Plan: playbook-authoring

Phase: playbook-authoring
Status: complete
Wave Count: 3
Execution Owner: work
Updated At: 2026-07-18

## Goal
Author `cli/docs/embedded/`: 6 self-sufficient stage playbooks + AGENTS.md shim + CONTEXT_RULES.md + AUTHORITY.md, command-verified against the real CLI.

## Inputs
- `skills/workflow/{brainstorm,to-plan,work,check,handoff,watzup}/SKILL.md` + their `references/`
- `cli/internal/interfaces/*.go` (authoritative command surface)
- `cli/docs/STATE.md` (state contract, drift/recovery vocabulary)

## Wave 1
### T1 — Extract the command surface inventory
- type: docs
- inputs: `cli/internal/interfaces/*.go`
- touches: `cli/docs/embedded/` (scratch notes only, committed in T2+)
- avoid: Go source edits
- steps:
  1. Enumerate every command/flag from the cobra definitions (`Use:`/`Short:`/flag registrations)
  2. Cross-check against commands quoted in the 6 SKILL.md files; list mismatches (aspirational or renamed commands)
- expected outputs: verified command inventory (table in the PR description or a scratch section later deleted)
- verification: each inventoried command runs with `--help` without error: `zharness {cmd} --help`
- stop if: a SKILL.md-quoted command has no CLI implementation and no documented changeset workaround
- escalate to: to-plan phase (the workaround belongs in a CLI phase, not a doc)

### T2 — AUTHORITY.md + AGENTS.md shim
- type: docs
- inputs: SPEC R6, upstream request-class pattern (concept only)
- touches: `cli/docs/embedded/AUTHORITY.md`, `cli/docs/embedded/AGENTS.md`
- avoid: agent-runtime-specific instructions (no Claude tool names)
- steps:
  1. Write AUTHORITY.md: read-only vs change request classes, which `zharness` commands each class may run, done-definitions per class
  2. Write AGENTS.md shim: entrypoint, lifecycle overview, links to playbooks + CONTEXT_RULES + AUTHORITY, version-gate instruction
- expected outputs: two docs, shim ≤ ~60 lines
- verification: inspection — no mutating command listed as read-only-allowed; shim links resolve to files that will exist in the set
- stop if: an authority rule contradicts an existing skill contract (e.g. watzup read-only vs a mutating step)
- escalate to: brainstorm refine

## Wave 2
### T3 — Spine playbooks: brainstorm, to-plan, work
- type: docs
- inputs: T1 inventory, respective SKILL.md + references
- touches: `cli/docs/embedded/playbooks/{brainstorm,to-plan,work}.md`
- avoid: inventing new lifecycle steps; changing any command semantics
- steps:
  1. Distill each SKILL.md + references into one playbook per the locked structure (Purpose → Preconditions → Steps → Artifacts → Exit)
  2. Inline the artifact templates (spec-template, roadmap/phase templates, run-artifact-template) into their playbook
  3. Replace Claude-specific mechanics (AskUserQuestion, Skill invocation) with agent-neutral equivalents ("ask the user structured questions", "proceed to the {stage} playbook")
- expected outputs: 3 playbooks
- verification: side-by-side checklist — every numbered workflow step and every zharness call in the source SKILL.md is present or consciously dropped with a note; all commands appear in T1 inventory
- stop if: semantic loss cannot be avoided for a step
- escalate to: user clarification

### T4 — Spine playbooks: check, handoff, watzup
- type: docs
- inputs: T1 inventory, respective SKILL.md + references
- touches: `cli/docs/embedded/playbooks/{check,handoff,watzup}.md`
- avoid: weakening check's verdict matrix or watzup's output contract
- steps:
  1. Same distillation as T3; check keeps its severity/verdict tables verbatim; watzup keeps the full output contract (forbidden phrases, layout, self-check)
  2. Handoff playbook inlines the handoff template + anchor-resolution steps
- expected outputs: 3 playbooks
- verification: same side-by-side checklist as T3
- stop if: same as T3
- escalate to: user clarification

## Wave 3
### T5 — CONTEXT_RULES.md + coherence pass
- type: docs
- inputs: all 8 docs from T2–T4
- touches: `cli/docs/embedded/CONTEXT_RULES.md`, minor edits across the set
- avoid: adding new rules not derivable from the playbooks
- steps:
  1. Write CONTEXT_RULES.md: per-stage table — docs to read, docs NOT to read, plus git's single `query check --latest` step (R5)
  2. Coherence pass over the whole set: uniform structure, consistent terminology (phase/story/run/check/handoff), no dangling links
  3. Verify R4: for each playbook, walk it as a cold agent — can every step be executed with only this doc + CLI help?
- expected outputs: CONTEXT_RULES.md + final coherent doc set
- verification: link check (`grep -o '\[.*\](.*)'` targets exist); cold-walk notes recorded in the run artifact
- stop if: a playbook still requires reading a SKILL.md to execute
- escalate to: check (gate the phase)

## Risks / Watch-fors
- The watzup/check contracts contain exact-string requirements (recovery strings, forbidden phrases) — copy verbatim, do not paraphrase; #24 was exactly a paraphrase drift
- Command inventory (T1) is the anchor for everything — do not start T3/T4 before it is complete
