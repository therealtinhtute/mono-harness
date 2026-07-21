# AUTHORITY — Request-Class Authority

Before touching `zharness`, classify the incoming request. The class determines which commands you may run. Do not infer mutation permission from a read-only request — if unsure which class a request falls into, treat it as read-only until the user affirmatively asks for the change to be executed.

## Read-only class

**Triggers:** answer a question, recap status, review or explain existing state/artifacts, inspect history. No request to create, change, execute, or close a phase.

**Allowed commands:**
- `zharness resume [--json]`
- `zharness query <state|phases|artifacts|check> [--phase <slug>] [--latest] [--json]`
- `zharness audit [--json]`
- `zharness validate [--json]`
- `zharness score-trace <trace-id> [--json]`
- `zharness score-context <trace-id> [--json]`
- `zharness --help`, `zharness --version`

**Forbidden:** `init`, `intake`, `story`, `trace add`, `db changeset apply`, `check record`, `handoff record`, `backlog`, `decision`, `intervention`, `tool`, `import`. None of these may run under a read-only request, even if the result "would help."

**Done-definition:** the request is answered using only the output of allowed commands plus file reads. Zero DB rows written, zero changesets created, zero files under `.kit/` created or modified.

## Change class

**Triggers:** the user asks to plan, execute, gate, or close out a phase — any of the six lifecycle stages (brainstorm, to-plan, work, check, handoff) — or explicitly asks for a mutating record (backlog item, decision, tool usage, human intervention).

**Allowed commands:** whatever the active stage playbook specifies. A change request is still scoped to the playbook currently in effect — `work` does not run `check record`, `check` does not run `story`. See `CONTEXT_RULES.md` for the per-stage doc-and-command scope.

**Done-definition:** the Exit / handoff conditions section of the stage playbook that handled the request. Mutation is expected and required; a change request that produced no DB row or changeset did not complete.

## Stage-to-class mapping

Source of truth: `STATE.md`'s Writer/Reader Ownership table, restated here as authority classes.

| Stage | Class | Writes | Reads |
|---|---|---|---|
| watzup | read-only | nothing | everything (`resume`) |
| git (`query check --latest`) | read-only | nothing | `check` (latest only) |
| brainstorm | change | `intake` | `spec`, prior `story` rows |
| to-plan | change | `story` rows, `meta.current_phase` | `spec`, prior `story` rows |
| work | change | `run` rows, `trace` rows | `story`, latest run/check |
| check | change | `check` rows | `run`, `story` |
| handoff | change | `handoff` rows | `run`, `check`, `story` |

## Straddling requests

"Review this, then fix what's wrong" is two requests, not one: handle the review under read-only class first, present findings, and only proceed to the change class once the user confirms the fix should be executed. Do not chain a mutating command onto a review's output without that confirmation.
