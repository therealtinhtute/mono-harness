# Citation Rules and Evidence Discipline

Evidence-first discipline for GitHub code research. These rules prevent hallucination
and ensure all claims are backed by observable tool output.

## Core Principle

**Never speculate beyond observed tool output.** If you didn't see it in a command
result or cached file, don't present it as fact.

## Citation Types

### 1. Code Content Claims

**Rule:** Must cite cached file with line range.

**Format:** `.kit/cache/github/owner/repo/path:lineStart-lineEnd`

**Examples:**

✅ **Correct:**
```
The `NewCmdRoot` function is defined in `.kit/cache/github/cli/cli/pkg/cmd/root/root.go:42-56`
```

❌ **Wrong:**
```
The function is in root.go (no line range)
```

❌ **Wrong:**
```
Based on the search results, it's probably in pkg/cmd/root.go (not cached, speculative)
```

❌ **Wrong:**
```
The gh search textMatches show it's defined here (textMatches are not proof)
```

**Why:** Code claims require you to have actually read the file. Search results and
textMatches are hints, not evidence.

### 2. Path/Metadata Claims

**Rule:** Cite command output or `owner/repo:path` format.

**Format:** `owner/repo:path` or `.kit/cache/github/owner/repo/path`

**Examples:**

✅ **Correct:**
```
The repo contains these TypeScript files:
- cli/cli:pkg/cmd/root/root.ts
- cli/cli:pkg/cmd/version/version.ts
(from gh api tree output)
```

✅ **Correct:**
```
The file exists at `.kit/cache/github/cli/cli/README.md`
```

❌ **Wrong:**
```
The repo probably has a README.md (not observed)
```

**Why:** Path claims are about structure, not content. Tree/search output is sufficient
proof that a path exists.

### 3. Partial Evidence

**Rule:** State what is confirmed and what remains uncertain.

**Examples:**

✅ **Correct:**
```
Confirmed: `NewCmdRoot` is defined in pkg/cmd/root/root.go:42
Uncertain: Whether this is the only definition (only searched cli/cli repo)
```

✅ **Correct:**
```
Found 3 usage examples in facebook/react (cached and cited below).
Note: Limited to --limit 20 search results; more examples may exist.
```

❌ **Wrong:**
```
This is the only place where NewCmdRoot is defined (overconfident)
```

**Why:** Honesty about search scope prevents false confidence.

## What Counts as Evidence

### ✅ Valid Evidence

1. **Cached file content** — you ran Read tool on `.kit/cache/github/owner/repo/path`
2. **Command output** — you ran `gh api` or `gh search` and saw the result
3. **Local tool output** — you ran `rg`, `grep`, `find` on cached files

### ❌ Not Evidence

1. **Search textMatches** — these are snippets, not full context
2. **Assumptions** — "it's probably in src/" without checking
3. **Prior knowledge** — "I know React uses hooks" without citing code
4. **Speculation** — "this might be related to X" without proof

## Citation Format Examples

### Code with Line Range
```
Function definition: `.kit/cache/github/cli/cli/pkg/cmd/root/root.go:42-56`

    42  func NewCmdRoot() *cobra.Command {
    43      cmd := &cobra.Command{
    44          Use:   "gh",
    45          Short: "GitHub CLI",
    ...
    56  }
```

### Multiple Locations
```
Found in 3 files:
1. `.kit/cache/github/cli/cli/pkg/cmd/root/root.go:42-56` — main definition
2. `.kit/cache/github/cli/cli/pkg/cmd/root/root_test.go:15-20` — test usage
3. `.kit/cache/github/cli/cli/cmd/gh/main.go:8` — import statement
```

### Path-Only Citation
```
Repository structure (from tree API):
- cli/cli:pkg/cmd/root/
- cli/cli:pkg/cmd/version/
- cli/cli:pkg/cmd/auth/
```

## Snippet Guidelines

### Keep Snippets Short

**Rule:** 5-15 lines max. Longer snippets add noise without value.

✅ **Good:**
```go
// .kit/cache/github/cli/cli/pkg/cmd/root/root.go:42-46
func NewCmdRoot() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "gh",
        Short: "GitHub CLI",
    }
```

❌ **Bad:**
```go
// 50 lines of code pasted here...
```

### Include Line Numbers

**Rule:** Always show line numbers in snippets for traceability.

✅ **Good:**
```
42  func NewCmdRoot() *cobra.Command {
43      cmd := &cobra.Command{
```

❌ **Bad:**
```
func NewCmdRoot() *cobra.Command {
    cmd := &cobra.Command{
```

### Add Context When Needed

**Rule:** Brief explanation of what the snippet proves.

✅ **Good:**
```
The function signature shows it returns *cobra.Command:

    42  func NewCmdRoot() *cobra.Command {
```

❌ **Bad:**
```
42  func NewCmdRoot() *cobra.Command {
(no explanation of why this matters)
```

## Private Repo Constraints

**Rule:** If access fails, report the constraint explicitly.

✅ **Correct:**
```
Cannot access owner/private-repo: gh api returned 404.
This repo is either private (requires authentication) or does not exist.
```

❌ **Wrong:**
```
The repo doesn't exist (might be private, not confirmed)
```

## Search Scope Transparency

**Rule:** State search parameters and limits.

✅ **Correct:**
```
Searched cli/cli repo with --limit 30.
Found 12 matches for "NewCmd" pattern.
```

✅ **Correct:**
```
Searched across facebook org repos (--owner facebook --limit 20).
Results may not be exhaustive.
```

❌ **Wrong:**
```
Found all occurrences of X (overstates completeness)
```

## When Evidence is Insufficient

**Rule:** Say so explicitly and suggest next steps.

✅ **Correct:**
```
Current evidence: Found 2 definitions in cli/cli repo.
Insufficient to confirm: Whether other repos also define this.
Next step: Search across --owner cli with broader scope.
```

❌ **Wrong:**
```
Only 2 definitions exist (unverified claim)
```

## Summary Checklist

Before citing anything, verify:
- [ ] Did I cache and read the file? (for code claims)
- [ ] Did I observe this in command output? (for path claims)
- [ ] Are line numbers included? (for code snippets)
- [ ] Is the snippet 5-15 lines? (not too long)
- [ ] Did I state search scope and limits? (transparency)
- [ ] Did I acknowledge uncertainty? (partial evidence)
