# Watzup Execution Modes

Two modes: `fast` (default, console-only) and `deep` (console + file). Both speak project language only — refer to `output-contract.md` for vocabulary, forbidden phrases, table formats, and section ordering rules.

## Fast Mode

**Purpose:** Quick session summary for routine wrap-ups.

**When to use:**
- Daily session wrap-ups
- Quick at-a-glance status check at the start of a fresh session
- Before short breaks

**Output destination:** Console only. No file is written.

**Layout:** See `output-contract.md` Section 5. Section order: Title → Trạng thái → Thay đổi chính → Risks (omit if none) → Next. In harness repos, include readiness / artifact-chain state inside `Trạng thái` without adding a new section. Total length target ≤ 25 visible lines.

**Empty-state branch:** When the working tree is clean and there is no new activity since the last review, print only the two-line empty-state message defined in `output-contract.md` Section 5. Do not write any file.

## Deep Mode

**Purpose:** Comprehensive review for PR preparation, milestone wrap-up, or shareable session report.

**When to use:**
- Before creating a pull request
- After completing a major milestone or sprint
- When a written report is needed for handoff or archival

**Output destination:**
- Console — same fast-mode shape, printed for at-a-glance summary
- File — written to `.kit/reports/watzup/{YYYYMMDD}-{slug}.{ext}`
  - `{YYYYMMDD}` = run date, no separators
  - `{slug}` = branch name slugified
  - `{ext}` = `md` (default) or `html` per `--format=`

**Format options:**
- `--format=md` (default) — markdown report with YAML frontmatter (schema in `output-contract.md` Section 7)
- `--format=html` — single self-contained HTML file with embedded `<style>`, no external resources

**Layout:** See `output-contract.md` Section 6. Section order: Frontmatter → Title → Trạng thái → Changes Overview → Key Changes → Quality Assessment → Risks & Blockers (omit if none) → Next Steps.

In harness repos, the deep file should also surface artifact-chain readiness and recurring drift inside `Trạng thái`, `Key Changes`, or `Risks & Blockers` without breaking the required section order.

**Empty-state branch:** Identical to fast mode — clean tree + no new activity → two-line empty-state message, no file. The `--format=` flag is ignored when no file would be written.

## Mode Selection

Invocation patterns:

```
/watzup                                    # fast mode on current branch
/watzup feature/my-branch                  # fast mode on a specific branch
/watzup deep                               # deep mode on current branch, markdown
/watzup feature/my-branch deep             # deep mode on a specific branch, markdown
/watzup deep --format=html                 # deep mode, HTML output
/watzup feature/my-branch deep --format=html
```

Argument shape:
- Positional 1 (optional): branch name (default = current branch)
- Positional 2 (optional): `fast` or `deep` (default = `fast`)
- Flag (optional, only meaningful with `deep`): `--format=md|html` (default = `md`)
