# Output Template

## File Location

Save to: `.kit/plans/{date}-{slug}/interview-notes.md`

Example: `.kit/plans/20260201-user-auth/interview-notes.md`

## Template

```markdown
---
title: Interview Notes: [Topic]
description: Interview summary capturing requirements, decisions, and open questions.
status: draft
created: [YYYY-MM-DD]
tags: [interview, planning]
input: [file path | "inline"]
duration: [X minutes]
---

# Interview Notes: [Topic]

## Key Insights

### Core Requirements
- [Insight 1]
- [Insight 2]

### User Flows
- [Flow description]

### Technical Decisions
- [Decision]: [Rationale]

### Edge Cases Identified
- [Edge case]: [Expected behavior]

---

## Decisions Made

| Decision | Rationale | Priority |
|----------|-----------|----------|
| [Decision 1] | [Why] | MUST |
| [Decision 2] | [Why] | SHOULD |

---

## MoSCoW Prioritization

### Must Have
- [ ] [Requirement]

### Should Have
- [ ] [Requirement]

### Could Have
- [ ] [Requirement]

### Won't Have (this phase)
- [ ] [Requirement]

---

## Open Questions

- [ ] [Unresolved question 1]
- [ ] [Unresolved question 2]

---

## Next Steps

1. Create plan with `/plan` workflow
2. Or start implementation with `/code`

---

## Related

- Input: [link to original file if any]
- Plan: [link to generated plan if created]
```

## Quick Version (Minimal)

For simple interviews, use condensed format:

```markdown
---
title: Interview: [Topic]
description: Condensed interview summary.
status: draft
created: [YYYY-MM-DD]
tags: [interview]
---

# Interview: [Topic]

## Insights
- [Key point 1]
- [Key point 2]

## Decisions
- [Decision]: [Why]

## Next: /plan or /code
```
