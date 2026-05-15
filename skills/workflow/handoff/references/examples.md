# Handoff Skill Examples

## Example 1: Standard Session Handoff
**Scenario**: End of session after implementing authentication feature

**Input**:
```bash
/handoff
```

**Output**:
```
📝 Handoff captured

✓ Git state: feature/add-auth (3 commits)
✓ Continuity mode: standard
✓ Progress: 2 completed, 1 in progress
✓ Blockers: 0
✓ Next steps: 3 actions

Handoff written to: .kit/HANDOFF.md
```

**Explanation**: The skill captures git state, identifies completed work (login, logout) and in-progress work (session management), and documents next steps (add tests, create PR).

---

## Example 2: Handoff with Blockers
**Scenario**: Session blocked by missing API credentials

**Input**:
```bash
/handoff waiting for API keys from DevOps
```

**Output**:
```
📝 Handoff captured

✓ Git state: feature/payment-integration (2 commits)
✓ Progress: 1 completed, 1 blocked
✓ Blockers: 1 (API credentials)
✓ Next steps: 2 actions

Handoff written to: .kit/HANDOFF.md
```

**HANDOFF.md excerpt**:
```markdown
## Blockers & Issues

### Missing Stripe API Credentials
- **Issue**: Cannot test payment integration without production API keys
- **Needed**: Stripe API keys from DevOps team
- **Next**: Once keys received, add to .env and test payment flow
```

**Explanation**: The skill captures the blocker with context from user argument, documents what's needed to unblock, and suggests next action.

---

## Example 3: Mid-Feature Handoff
**Scenario**: Context switch needed before feature completion

**Input**:
```bash
/handoff
```

**Output**:
```
📝 Handoff captured

✓ Git state: feature/search-filters (1 commit)
✓ Progress: 1 completed, 2 in progress
✓ Blockers: 0
✓ Next steps: 4 actions

Handoff written to: .kit/HANDOFF.md
```

**HANDOFF.md excerpt**:
```markdown
## Progress This Session

### Completed ✓
- Added search input component with debouncing

### In Progress ⏳
- Implementing filter dropdown (UI done, logic pending)
- Adding sort options (not started)

## Next Steps

1. **Complete filter dropdown logic** — Wire up filter state to search query, test with mock data
2. **Add sort options** — Implement sort dropdown with options: relevance, date, price
3. **Connect to API** — Replace mock data with actual search endpoint
4. **Add tests** — Unit tests for search logic, integration tests for filters
```

**Explanation**: The skill captures mid-feature state with clear breakdown of completed vs in-progress work, enabling seamless continuation later.

---

## Example 4: Harness Phase Handoff
**Scenario**: Phase finished in `cook`, but `check` requested proof follow-up

**Input**:
```bash
/handoff phase complete, waiting on proof-gap cleanup
```

**Output**:
```
📝 Handoff captured

✓ Git state: feature/inbox-ui (5 commits)
✓ Continuity mode: full-harness
✓ Active phase: inbox-ui
✓ Latest cook run: .kit/runs/cook/20240901-1010-inbox-ui.md
✓ Latest check verdict: APPROVE with requests
✓ Next steps: 2 actions

Handoff written to: .kit/HANDOFF.md
```

**HANDOFF.md excerpt**:
```markdown
---
status: ready-for-review
continuity-mode: full-harness
active-phase: inbox-ui
---

## Continuity Anchors

**Latest Cook Run**: `.kit/runs/cook/20240901-1010-inbox-ui.md`
**Latest Check Verdict**: approve-with-requests
**Proof / Drift Notes**:
- missing verification output for task T5 in cook run log

## Next Steps

1. **→ START HERE: append proof for T5** — update the cook run artifact or rerun `/cook phase inbox-ui` so `check` can pass cleanly
2. **Re-run /check full** — confirm artifact alignment is now clean
```

**Explanation**: The skill preserves the exact phase, run artifact, and gate verdict so the next session resumes from the real blocker instead of re-reading everything.

---

## Example 5: Continuation from Previous Handoff
**Scenario**: Starting new session, reading previous handoff

**Input**:
```bash
cat .kit/HANDOFF.md
```

**Output**:
```markdown
---
session-date: 2026-05-01
branch: feature/add-auth
status: in-progress
last-updated: 2026-05-01 18:30
---

# Session Handoff — feature/add-auth

## Current State
...

## Next Steps

1. **Add session management** — Implement session store with Redis, handle session expiry
2. **Add integration tests** — Test full auth flow from login to protected route access
3. **Update API docs** — Document auth endpoints in OpenAPI spec
```

**Explanation**: Previous handoff provides complete context for continuation, enabling immediate resumption of work without context loss.
