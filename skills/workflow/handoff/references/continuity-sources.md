# Continuity Sources for Handoff

Use these in priority order when the repo follows the harness flow.

## Canonical vs. Narrative

The `handoffs` entity (written by `zharness handoff record`) is canonical — it is what `resume`/`watzup` read to reconstruct state on another machine. `.kit/HANDOFF.md` is the narrative layer: human-readable prose carrying the same anchors (entity `id`, `run_id`, `check_id`, open items) for a person reading cold. Never let the markdown diverge from the entity — write the entity first, then mirror its fields into the markdown frontmatter.

## Primary Sources

1. `.kit/workflow-state.yml`
   - determine the current phase pointer and latest canonical artifact paths quickly
   - treat it as the index, then verify the pointed files before trusting details

2. `.kit/planning/ROADMAP.md`
   - determine active phase order
   - identify whether work is mid-phase, blocked, or ready for next phase

3. `.kit/planning/phases/{slug}/{slug}-CONTEXT.md`
   - capture locked decisions, assumptions, and rejected options that the next session must not rediscover

4. `.kit/planning/phases/{slug}/{slug}-PLAN.md`
   - extract current wave/task state, expected verification, and remaining work

5. Latest `.kit/runs/work/*.md`
   - capture the real execution trail
   - pull task statuses, blockers, concerns, and proof commands

6. Latest `.kit/reports/check/*.md`
   - capture whether the phase gate passed, drifted, or failed from a canonical check report

## What to Synthesize

### Continuity Summary
- active phase
- last meaningful action completed
- highest-priority unresolved blocker or concern
- exact next action to resume
- whether `.kit/workflow-state.yml` was current, stale, or missing

### Do Not Lose
- blocker taxonomy from `work` (`BLOCKED_CONTEXT`, `BLOCKED_SCOPE`, `BLOCKED_VERIFICATION`, `BLOCKED_CONTRACT_DRIFT`)
- artifact drift or proof-gap findings from `check`
- any plan boundary the next session must stay inside

## Fallback Behavior

If harness artifacts are missing:
- say `continuity_mode: standard`
- fall back to git state + recent commits + working tree only

If artifacts partially exist:
- say `continuity_mode: partial-harness`
- name which source is missing instead of implying continuity is complete
