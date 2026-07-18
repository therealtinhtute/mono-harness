# Context: agent-pilot

Phase: agent-pilot
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: low (evidence-gathering; no production surface changes)
Expected Proof: e2e (live second-agent lifecycle pass), inspection (evidence doc)

## Goal
Prove agent-agnosticism (R9): a non-Claude agent completes `intake → story → trace add → check record` on a sample task using only the written docs (`.kit/docs/**`, AGENTS.md) + the installed CLI — zero SKILL.md reads — and the evidence is published.

## Scope Boundary
### Allowed Surfaces
- A scratch target project (fresh dir or a dedicated test repo) for the pilot run
- `docs/workflow-harness/pilot-evidence/` (evidence doc)
- Read-only on everything else

### Forbidden Surfaces
- `cli/**`, `skills/workflow/**` — pilot failures are findings, never hotfixes (roadmap rule)
- This repo's live `.kit/` (the pilot runs on a scratch target, not on Lab/skills' own state)

## Spec Hooks
- R9 (pilot + evidence), acceptance: "completed without reading any SKILL.md, `zharness validate --json` passes on the produced chain"
- SPEC open question resolved at phase start: Codex CLI vs Cursor, by availability

## Locked Decisions
- Pilot target: a scratch project scaffolded by `zharness init` — not this repo — so the pilot exercises the fresh-project path (R2) and cannot contaminate live state
- Sample task: small but real (e.g. a tiny script/module with a test), enough to give trace/check content meaning; not a hello-world with empty proof
- Protocol: the second agent receives only a pointer to `AGENTS.md` and the task statement — no coaching about the lifecycle in the prompt; mid-run human answers are allowed only for the task's product questions, never for harness mechanics
- Every deviation/confusion is logged as a finding in the evidence doc with a severity and a routing (playbook fix / CLI gap / authority-doc gap), mirroring the 2026-07-17 pilot-evidence format

## Assumptions
- The repo owner has a working second-agent runtime available; if neither Codex CLI nor Cursor is usable at phase start, the phase is blocked (not skipped) — R9 is an acceptance criterion, not optional

## Canonical Refs
- `docs/workflow-harness/pilot-evidence/2026-07-17-lab-skills-import.md` (evidence format precedent)
- `.kit/docs/**` as written by the released binary (the actual artifact under test)

## Rejected Options
- Piloting on this repo's live state — contaminates the daily driver and tests the legacy-import path instead of the fresh path
- Simulating the second agent with Claude pretending — self-defeating; the point is a different runtime's cold read

## Deferred Ideas
- Multi-agent matrix (Codex AND Cursor AND others); scripted conformance harness for playbook walks

## Escalate If
- The second agent cannot complete the lifecycle due to a docs gap that a reasonable fix cannot wait on → to-plan phase (playbook-fix mini-phase) — this is the one case where a fix cycle precedes closing the initiative
- The pilot passes but only with harness-mechanics coaching → treat as FAIL for R9; route findings and re-run
