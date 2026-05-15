# Git Skill Examples

### Example 1: Standard Commit Workflow
**Scenario**: Commit changes to multiple files with proper conventional commit format

**Input**:
```bash
/git cm
```

**Output**:
```
✓ staged: 3 files (+45/-12 lines)
  - src/auth/login.ts
  - src/auth/logout.ts
  - tests/auth.test.ts
✓ security: passed
✓ commit: a1b2c3d feat(auth): add logout functionality

from therealTINHTUTE with love
```

**Explanation**: The skill automatically stages all changes, scans for secrets, and creates a commit with conventional format. The type `feat` is chosen because new functionality was added, and scope `auth` groups related authentication changes.

---

### Example 2: Create Pull Request
**Scenario**: Create a PR from feature branch to main with description

**Input**:
```bash
/git pr main feature/add-auth
```

**Output**:
```
✓ pushed: feature/add-auth → origin
✓ PR created: #123
  Title: feat(auth): add authentication system
  URL: https://github.com/user/repo/pull/123
  
Report saved: .kit/reports/git/20260416-pr-123.md
```

**Explanation**: The skill pushes the feature branch, generates a PR title from commits, and creates the PR using gh CLI. A detailed report is saved for reference.

---

### Example 3: Safe Merge with Conflict Detection
**Scenario**: Merge feature branch into main, detecting conflicts before merge

**Input**:
```bash
/git merge main feature/add-auth
```

**Output**:
```
✓ fetched: origin/main
✓ conflict check: passed
✓ merged: feature/add-auth → main
✓ commits: 3 new commits
  - a1b2c3d feat(auth): add login
  - b2c3d4e feat(auth): add logout
  - c3d4e5f test(auth): add auth tests
```

**Explanation**: The skill fetches latest changes, checks for conflicts before merging, and provides a summary of merged commits. If conflicts were detected, it would stop and suggest manual resolution.

---

### Example 4: Revert Bad Commit
**Scenario**: Safely revert a commit that broke tests without losing history

**Input**:
```bash
git revert a1b2c3d
```

**Output**:
```
✓ reverted: a1b2c3d feat(auth): add logout
✓ commit: d4e5f6g revert: feat(auth): add logout

This reverts commit a1b2c3d which broke authentication tests.
```

**Explanation**: Using `git revert` creates a new commit that undoes changes, preserving history. This is safer than `git reset --hard` which would lose the commit entirely.

---

### Example 5: Commit with Secret Detection
**Scenario**: Attempt to commit code containing API keys, blocked by security scan

**Input**:
```bash
/git cm
```

**Output**:
```
✓ staged: 2 files (+30/-5 lines)
❌ security: FAILED
  
Secrets detected in:
  - src/config.ts:12 (API_KEY)
  - src/config.ts:15 (SECRET_TOKEN)

⚠️  Commit blocked. Add these to .gitignore or use environment variables.
```

**Explanation**: The security scan detects potential secrets before committing. The commit is blocked to prevent credential exposure. User should move secrets to `.env` files and add to `.gitignore`.
