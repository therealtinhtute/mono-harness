# Watzup Skill Examples

## Example 1: Standard Session Review
**Scenario**: Review current branch after completing feature work

**Input**:
```bash
/watzup
```

**Output**:
```
📊 Session Review — feature/add-auth

✓ 5 commits analyzed
✓ 12 files changed (+450/-120)
✓ Quality: 8/10

Key changes:
  • Added authentication system (login, logout, session)
  • Implemented JWT token handling
  • Added auth middleware and guards

⚠️  Risks: 1
  • Missing integration tests for auth flow

Next: Add integration tests, then create PR
```

**Explanation**: The skill analyzes recent commits, identifies key changes, assesses quality, and provides actionable next steps. Quality score is based on test coverage, documentation, and change scope.

---

## Example 2: Pre-PR Review
**Scenario**: Review before creating pull request to ensure readiness

**Input**:
```bash
/watzup feature/add-auth
```

**Output**:
```
📊 Session Review — feature/add-auth

✓ 8 commits analyzed
✓ 18 files changed (+680/-95)
✓ Quality: 9/10

Key changes:
  • Authentication system (login, logout, session, JWT)
  • Auth middleware and route guards
  • Integration tests for auth flows
  • API documentation updates

✓ PR Ready
  • All tests passing
  • Documentation updated
  • No uncommitted changes

Next: Create PR to main
```

**Explanation**: The skill confirms PR readiness by checking tests, documentation, and uncommitted changes. High quality score indicates the branch is ready for review.

---

## Example 3: Session with Risks
**Scenario**: Review session with potential issues flagged

**Input**:
```bash
/watzup
```

**Output**:
```
📊 Session Review — feature/refactor-db

✓ 3 commits analyzed
✓ 25 files changed (+1200/-800)
✓ Quality: 6/10

Key changes:
  • Database schema migration (users, sessions, tokens)
  • ORM model updates
  • Migration scripts

⚠️  Risks: 3
  • Large schema changes (breaking)
  • Missing rollback migration
  • No integration tests for migration

Next: Add rollback migration, test on staging before PR
```

**Explanation**: The skill identifies high-risk changes (large schema migration) and flags missing safety measures (rollback, tests). Lower quality score reflects these risks.

---

## Example 4: Clean State Review
**Scenario**: Review when no recent changes exist

**Input**:
```bash
/watzup
```

**Output**:
```
📊 Session Review — main

✓ No new commits since last review
✓ Working tree clean
✓ Up to date with origin/main

Next: Start new feature branch or pull latest changes
```

**Explanation**: The skill reports clean state when no changes are detected, suggesting next actions for starting new work.

---

## Example 5: Uncommitted Changes Warning
**Scenario**: Review with uncommitted changes blocking PR

**Input**:
```bash
/watzup
```

**Output**:
```
📊 Session Review — feature/add-auth

✓ 5 commits analyzed
✓ 12 files changed (+450/-120)

⚠️  Blockers: 1
  • 3 uncommitted files (src/auth/*, tests/auth.test.ts)

Next: Commit or stash changes before creating PR
```

**Explanation**: The skill flags uncommitted changes as blockers for PR creation, suggesting commit or stash operations.
