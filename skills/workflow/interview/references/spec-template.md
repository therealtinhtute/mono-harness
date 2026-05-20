# Validated Spec Template

```markdown
---
title: {Plan Title}
status: validated
interviewed: YYYY-MM-DD
---

## Outcome
{Single truth that must be real when this work is done}

## Success Condition
{Measurable proof — verifiable by someone who wasn't in the interview}

## Scope
**May change:** {files, modules, areas}
**Must not change:** {contracts, schemas, behaviors}

## Context to Read First
- {file, doc, service needed to start cold}

## Key Decisions
1. **{Decision}**: {Choice} — because {rationale}. Depends on: {related decision, if any}.

## Validation Loop
**During work:** {cheap, fast checks after each change}
**Final proof:** {comprehensive checks before claiming done}

## Stop / Pause
**Done when:** {unambiguous criteria}
**Pause if:** {conditions requiring human input}
```

All six hard-gate fields must be filled. No TBD, TODO, or placeholders.
Present the draft for acceptance before writing to file.
