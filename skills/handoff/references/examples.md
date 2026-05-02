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

## Example 4: Ready for Review Handoff
**Scenario**: Feature complete, ready for PR

**Input**:
```bash
/handoff feature complete, ready for review
```

**Output**:
```
📝 Handoff captured

✓ Git state: feature/add-auth (5 commits)
✓ Progress: 5 completed, 0 in progress
✓ Blockers: 0
✓ Next steps: 2 actions

Handoff written to: .kit/HANDOFF.md
```

**HANDOFF.md excerpt**:
```markdown
---
status: ready-for-review
---

## Progress This Session

### Completed ✓
- Authentication system (login, logout, session)
- JWT token handling and refresh
- Auth middleware and route guards
- Integration tests for auth flows
- API documentation updates

## Next Steps

1. **Run final checks** — Verify all tests pass, lint clean, no console errors
2. **Create PR** — Use /git pr to create pull request to main
```

**Explanation**: The skill marks status as ready-for-review and documents final steps before PR creation.

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
