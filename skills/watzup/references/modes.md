# Watzup Execution Modes

## Fast Mode (Default)

**Purpose:** Quick session summary for routine wrap-ups.

**Scope:**
- Last 10 commits
- Console output only
- No file generation

**Commands:**
```bash
git log --oneline --graph --decorate -10
git diff HEAD~10..HEAD --stat
git diff HEAD~10..HEAD --shortstat
```

**Output:**
```
📊 Session Review — {branch-name}

✓ {commit-count} commits analyzed
✓ {file-count} files changed (+{additions}/-{deletions})
✓ Quality: {score}/10

Key changes:
  • {change-1}
  • {change-2}
  • {change-3}

⚠️  Risks: {risk-count}
  • {risk-1}

Next: {primary-action}
```

---

## Deep Mode

**Purpose:** Comprehensive review for PR preparation or milestone wrap-up.

**Scope:**
- Last 30-50 commits (adjust based on branch age)
- Console output + detailed report file
- Save to `.kit/reports/watzup/{YYYYMMDD}-{branch}.md`

**Commands:**
```bash
git log --oneline --graph --decorate -50
git diff HEAD~50..HEAD --stat
git diff HEAD~50..HEAD --shortstat
```

**Output:**
Console summary (same as fast mode) plus detailed report file.

**Report Format:**
```markdown
---
title: Session Review — {branch-name}
branch: {branch-name}
commits: {count}
files: {count}
quality-score: {score}
created: YYYY-MM-DD
tags: [watzup, review, session]
---

## Session Summary — {branch-name}

### Changes Overview
- **Commits**: {count} ({types breakdown})
- **Files**: {modified} modified, {added} added, {removed} removed
- **Lines**: +{additions} -{deletions}

### Key Changes
1. {change-1} — {impact}
2. {change-2} — {impact}
3. {change-3} — {impact}

### Quality Assessment
- **Test Coverage**: {increased/decreased/unchanged}
- **Documentation**: {updated/missing}
- **Breaking Changes**: {yes/no}

### Risks & Blockers
- {risk-1}
- {risk-2}

### Next Steps
1. {action-1}
2. {action-2}
```

---

## Mode Selection

**Usage:**
```bash
/watzup                    # fast mode (default)
/watzup feature/my-branch  # fast mode on specific branch
/watzup deep               # deep mode on current branch
/watzup feature/my-branch deep  # deep mode on specific branch
```

**When to use fast:**
- Daily session wrap-ups
- Quick status checks
- Before short breaks

**When to use deep:**
- Before creating PR
- After completing major milestone
- Weekly/sprint reviews
- When detailed documentation needed
