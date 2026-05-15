# Git Skill Scripts

Automation scripts for common git workflows.

## Scripts

### commit-workflow.sh

Automated commit workflow with validation.

**Usage:**
```bash
./commit-workflow.sh "feat: add new feature"
./commit-workflow.sh "fix: resolve login bug" --dry-run
```

**Features:**
- Validates conventional commit format
- Auto-stages modified files
- Scans for secrets before commit
- Creates commit with signature
- Verifies commit success

**Options:**
- `-h, --help`: Show help
- `-v, --verbose`: Verbose output
- `-n, --dry-run`: Show what would be done

---

### create-pr.sh

Create pull request with standard template.

**Usage:**
```bash
./create-pr.sh "Add user authentication"
./create-pr.sh "Fix login bug" "reviewer1,reviewer2"
./create-pr.sh "WIP: New feature" --draft
```

**Features:**
- Pushes current branch to remote
- Generates PR description from commits
- Applies standard PR template
- Adds reviewers
- Supports draft PRs

**Options:**
- `-h, --help`: Show help
- `-b, --base`: Base branch (default: auto-detect)
- `-d, --draft`: Create as draft PR

**Requirements:**
- GitHub CLI (`gh`) must be installed

---

### safe-merge.sh

Safe merge with conflict detection and rollback.

**Usage:**
```bash
./safe-merge.sh feature-branch
./safe-merge.sh feature-branch --no-ff
./safe-merge.sh feature-branch --squash
```

**Features:**
- Fetches latest changes
- Detects conflicts before merge
- Performs merge with options
- Verifies merge success
- Automatic rollback on failure

**Options:**
- `-h, --help`: Show help
- `--no-ff`: Force merge commit
- `--squash`: Squash commits before merge

---

### branch-cleanup.sh

Clean up merged branches (local and remote).

**Usage:**
```bash
./branch-cleanup.sh
./branch-cleanup.sh --yes --remote
./branch-cleanup.sh --dry-run
```

**Features:**
- Lists merged branches
- Interactive confirmation
- Deletes local branches
- Optionally deletes remote branches
- Summary report

**Options:**
- `-h, --help`: Show help
- `-y, --yes`: Skip confirmation prompts
- `-r, --remote`: Also delete remote branches
- `-n, --dry-run`: Show what would be deleted

---

## Installation

All scripts are executable and can be run directly:

```bash
cd kit/skills/git/scripts
./commit-workflow.sh --help
```

## Testing

Run tests with:

```bash
bats tests/
```

## Contributing

Follow the script quality standards in `.kit/plans/20260416-week3-quality/phase-03-scripts.md`.

from therealTINHTUTE with love
