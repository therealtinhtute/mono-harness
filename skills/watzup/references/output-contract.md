# Output Contract

This document is the single source of truth for what `watzup` may print or write to disk. Every rendered output — fast console summary, deep markdown report, deep HTML report — must conform to these rules.

## 1. Forbidden Phrases

Output MUST NOT contain any of the following. They describe how the skill works internally, which is irrelevant to the reader.

- Shell or git commands: `git log`, `git diff`, `git status`, `git branch`, `git show`, `--stat`, `--shortstat`, `--oneline`, `HEAD~`, `..HEAD`, `--graph`, `--decorate`
- Process descriptors: `commit window`, `diff stat`, `last 10 commits`, `last 50 commits`, `fast mode scope`, `deep mode scope`, `analyzed`, `scanned`
- Synthesized scoring: any `Quality: N/10` pattern, any `Score: N/10`, any `/10` rating
- Skill-internal terms: `watzup mode`, `fast mode`, `deep mode` (when describing behavior to the user — naming the invocation is fine)

**Trailing-space check**: the literal substring `git ` (with a trailing space) MUST NOT appear in any rendered output. Project nouns containing `git` without a trailing space (e.g., a directory called `git-tools`) are exempt by this rule.

## 2. Allowed Vocabulary

Use project-language phrases. Numbers are facts about the work, not facts about the skill's process.

- Branch references: "this branch", "branch X", "ahead of main by N commits", "behind main by N commits"
- Working tree: "uncommitted work", "uncommitted files", "staged changes", "clean working tree"
- Recent activity: "this session's work", "recent changes", "what changed", "the changes"
- Quantities (allowed when describing the work): "12 files modified, +450/-120 lines", "5 commits"
- Risks: "tests missing for X", "schema migration without rollback", "breaking change in public API"

Sample sentences:
- "This branch is 5 commits ahead of main with a clean working tree."
- "12 files changed across the auth module: +450 / -120 lines."
- "Tests are missing for the new login flow."

## 3. Session Title Format

Exact pattern, used identically in fast and deep modes:

```
Session — {branch} ({YYYY-MM-DD})
```

- `{branch}` is the branch under review (current branch unless overridden)
- `{YYYY-MM-DD}` is the run date, local timezone
- The em dash (`—`) is required; do not substitute `-` or `--`

There is no user override flag for the title.

## 4. Risk Table Contract

When risks are present, render exactly this table:

```
| Risk | Mức độ | Action |
|------|--------|--------|
| {short noun phrase} | {cao\|vừa\|thấp} | {single concrete action} |
```

Rules:
- Column order is fixed: `Risk`, `Mức độ`, `Action`
- Severity ladder is fixed at 3 levels: `cao`, `vừa`, `thấp`. No additions, no English equivalents in the cell.
- `Risk` column: short noun phrase, no leading verbs ("Schema migration without rollback", not "There is no rollback for the schema migration")
- `Action` column: one concrete action that resolves the risk ("Add rollback migration before merge")
- If the run produces zero risks, the entire Risks section is OMITTED. Do not print "no risks", do not print an empty table.

## 5. Fast Output Layout

Console-only. Total length target: ≤ 25 visible lines.

Section order (omit a section entirely if it has no content, except the title which is mandatory):

1. **Title line** — exact format from Section 3
2. **Trạng thái** — short bullet list. Items: branch position vs. main, uncommitted file count (if any), current branch
3. **Thay đổi chính** — bullet list, max 5 items, grouped by intent (one bullet = one coherent change theme)
4. **Risks** — table per Section 4 (omitted if zero risks)
5. **Next** — single line, one concrete next action

Empty-state branch: when the working tree is clean AND there are no new commits since the last review, output ONLY:

```
Đã sạch — không có thay đổi.
Next: {one-line suggestion, e.g., "Bắt đầu nhánh mới hoặc kéo thay đổi mới nhất."}
```

No file is written in the empty state, even if the user invoked `/watzup deep`.

## 6. Deep Output Layout

Console output: same shape as fast (Section 5), printed for at-a-glance summary.

File output: written to `.kit/reports/watzup/{YYYYMMDD}-{slug}.{ext}` where:
- `{YYYYMMDD}` is run date (no separators)
- `{slug}` is the branch name slugified: lowercase, `/` → `-`, non-alphanumeric → `-`, collapse consecutive `-`
- `{ext}` is `md` (default) or `html` per `--format=` flag
- Existing files at the same path are overwritten (one report per branch per day)
- Directory is created if missing

File section order:
1. **Frontmatter** — YAML, schema in Section 7 (markdown only; HTML embeds equivalent metadata in `<head>`)
2. **Title heading** — `# Session — {branch} ({YYYY-MM-DD})`
3. **Trạng thái** — same as fast
4. **Changes Overview** — counts: commits by type (feat/fix/chore/refactor/test/docs), files modified/added/removed, lines added/removed
5. **Key Changes** — numbered list, each item is `{change} — {impact}`
6. **Quality Assessment** — three lines:
   - `Test Coverage: {increased|decreased|unchanged}`
   - `Documentation: {updated|missing}`
   - `Breaking Changes: {yes|no}`
   These are facts, not scores.
7. **Risks & Blockers** — table per Section 4 (omitted if zero risks)
8. **Next Steps** — numbered list of concrete actions

HTML format additional rules:
- Single self-contained file
- Embed all CSS in a `<style>` block in `<head>`
- MUST NOT contain `<link rel="stylesheet">`, `<script src="...">`, web font URLs, CDN URLs, or any external network reference
- Risk table renders as a real `<table>` with `<thead>` and `<tbody>`
- Severity values are plain text in cells; optional inline `style="color: …"` for visual cue

## 7. Frontmatter Schema (markdown only)

```yaml
---
title: Session — {branch} ({YYYY-MM-DD})
branch: {branch}
commits: {integer}
files: {integer}
created: {YYYY-MM-DD}
tags: [watzup, review, session]
---
```

Required keys: all six. Order: as shown. Tags array is fixed.

## 8. Self-Check (run before printing any output)

Before emitting fast output to the console or writing a deep file, validate:

1. The substring `git ` (with trailing space) is absent from the rendered output.
2. No `Quality: N/10` or `Score: N/10` appears.
3. No phrase from Section 1's Process Descriptors list appears.
4. Title line matches the exact format from Section 3.
5. If a Risks section is rendered, its table columns are exactly `Risk | Mức độ | Action` and severity values are in `{cao, vừa, thấp}`.
6. If empty-state, output is exactly the two lines specified in Section 5.
7. Deep markdown: frontmatter has all 6 keys in order; section order matches Section 6.
8. Deep HTML: contains `<style>`, contains zero `<link `, `<script src=`, or external URLs.

If any check fails, fix the draft and re-run the self-check before printing.
