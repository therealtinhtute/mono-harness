# GitHub CLI Command Patterns

Known-good `gh` command templates for GitHub code research. These patterns are tested
and reliable for the librarian workflow.

## Variable Convention

Set variables when useful for readability:
```bash
REPO='owner/repo'
REF='branch-or-sha'
DIR='src'
FILE='path/to/file.ts'
```

## Pattern 0: Resolve Default Branch

When REF is unknown, resolve the default branch:
```bash
gh repo view "$REPO" --json defaultBranchRef --jq '.defaultBranchRef.name'
```

Example:
```bash
REPO='cli/cli'
REF=$(gh repo view "$REPO" --json defaultBranchRef --jq '.defaultBranchRef.name')
echo "Default branch: $REF"
```

## Pattern 1: Code Search

Search for code across GitHub with filters:
```bash
gh search code '<terms>' --json path,repository,sha,url,textMatches --limit 30
```

Optional scope filters:
- `--repo owner/repo` — limit to specific repository
- `--owner owner` — limit to specific owner/org
- `--match path` — match in file paths only
- `--match file` — match in file content only

Examples:
```bash
# Search in specific repo
gh search code "NewCmdRoot" --repo cli/cli --json path,repository,sha,url --limit 10

# Search across owner's repos
gh search code "useEffect" --owner facebook --limit 20

# Path-only search
gh search code "README.md" --repo owner/repo --match path --limit 5
```

## Pattern 2: Repo Tree Map

Get recursive tree structure:
```bash
gh api "repos/$REPO/git/trees/$REF?recursive=1" > tree.json
```

Example:
```bash
REPO='cli/cli'
REF='trunk'
gh api "repos/$REPO/git/trees/$REF?recursive=1" > tree.json
```

## Pattern 3: Filter Tree Paths

Filter tree output with jq:
```bash
jq -r '.tree[] | select(.type=="blob" and (.path | startswith("src/"))) | .path' tree.json | head
```

Examples:
```bash
# Find all TypeScript files
jq -r '.tree[] | select(.type=="blob" and (.path | endswith(".ts"))) | .path' tree.json

# Find files in specific directory
jq -r '.tree[] | select(.type=="blob" and (.path | startswith("pkg/cmd/"))) | .path' tree.json

# Count files by extension
jq -r '.tree[] | select(.type=="blob") | .path' tree.json | sed 's/.*\.//' | sort | uniq -c
```

## Pattern 4: Directory Entries via Contents API

List directory contents:
```bash
# Specific directory
gh api "repos/$REPO/contents/$DIR?ref=$REF" --jq '.[] | [.type, .path] | @tsv'

# Repo root
gh api "repos/$REPO/contents?ref=$REF" --jq '.[] | [.type, .path] | @tsv'
```

Examples:
```bash
REPO='cli/cli'
REF='trunk'

# List root directory
gh api "repos/$REPO/contents?ref=$REF" --jq '.[] | [.type, .path] | @tsv'

# List specific directory
gh api "repos/$REPO/contents/pkg?ref=$REF" --jq '.[] | [.type, .path] | @tsv'
```

## Pattern 5: Fetch File to Local Cache

Download and decode file content:
```bash
mkdir -p ".kit/cache/github/$REPO/$(dirname "$FILE")"
gh api "repos/$REPO/contents/$FILE?ref=$REF" --jq .content | tr -d '\n' | base64 --decode > ".kit/cache/github/$REPO/$FILE"
```

Example:
```bash
REPO='cli/cli'
REF='trunk'
FILE='pkg/cmd/root/root.go'

mkdir -p ".kit/cache/github/$REPO/$(dirname "$FILE")"
gh api "repos/$REPO/contents/$FILE?ref=$REF" --jq .content | tr -d '\n' | base64 --decode > ".kit/cache/github/$REPO/$FILE"

echo "Cached to: .kit/cache/github/$REPO/$FILE"
```

## Pattern 6: Refine Locally After Caching

Use local tools on cached files:
```bash
# Search with ripgrep (line numbers)
rg -n '<pattern>' ".kit/cache/github/$REPO"

# Search with grep (line numbers)
grep -rn '<pattern>' ".kit/cache/github/$REPO"

# Find files
find ".kit/cache/github/$REPO" -name "*.ts"

# Count lines
wc -l ".kit/cache/github/$REPO/$FILE"
```

Examples:
```bash
REPO='cli/cli'

# Find all occurrences of "NewCmd"
rg -n "NewCmd" ".kit/cache/github/$REPO"

# Find TypeScript files
find ".kit/cache/github/$REPO" -name "*.ts"

# Search for imports
rg -n "^import.*from" ".kit/cache/github/$REPO"
```

## Pattern 7: Get Line-Numbered Evidence

Extract specific lines with context:
```bash
# Using nl (number lines)
nl -ba ".kit/cache/github/$REPO/$FILE" | sed -n '42,56p'

# Using sed with line numbers
sed -n '42,56p' ".kit/cache/github/$REPO/$FILE" | nl -ba

# Using head/tail
head -n 56 ".kit/cache/github/$REPO/$FILE" | tail -n 15 | nl -ba -v 42
```

Examples:
```bash
REPO='cli/cli'
FILE='pkg/cmd/root/root.go'

# Lines 42-56 with line numbers
nl -ba ".kit/cache/github/$REPO/$FILE" | sed -n '42,56p'

# Context around line 50 (±5 lines)
nl -ba ".kit/cache/github/$REPO/$FILE" | sed -n '45,55p'
```

## Error Handling

### Private Repo Access
If `gh api` returns 404 or 403:
```bash
gh api "repos/owner/private-repo/contents/file.ts?ref=main" 2>&1 | grep -q "404\|403"
if [ $? -eq 0 ]; then
  echo "Access denied or repo not found"
fi
```

### Rate Limiting
Check rate limit status:
```bash
gh api rate_limit --jq '.rate | {limit, remaining, reset: (.reset | strftime("%Y-%m-%d %H:%M:%S"))}'
```

### Validate Repo Exists
```bash
gh repo view "$REPO" --json name,owner,defaultBranchRef > /dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "Repo not found or not accessible: $REPO"
fi
```
