# Output Contract

This document is the single source of truth for what `watzup` may print. Every rendered output must conform to these rules.

## 1. Forbidden Phrases

Output MUST NOT contain any of the following:

- Shell or git commands: `git log`, `git diff`, `git status`, `git branch`, `git show`, `--stat`, `--shortstat`, `--oneline`, `HEAD~`, `..HEAD`, `--graph`, `--decorate`
- Process descriptors: `commit window`, `diff stat`, `last 10 commits`, `last 50 commits`, `analyzed`, `scanned`
- Synthesized scoring: any `Quality: N/10` pattern, any `Score: N/10`, any `/10` rating
- Skill-internal terms: `recap mode`, `orient mode` (when describing behavior to the user — naming the invocation is fine)

**Trailing-space check**: the literal substring `git ` (with a trailing space) MUST NOT appear in any rendered output. Project nouns containing `git` without a trailing space (e.g., a directory called `git-tools`) are exempt.

## 2. Allowed Vocabulary

Use project-language phrases. Numbers describe the work, not the skill's process.

- Branch references: "this branch", "branch X", "ahead of main by N commits", "behind main by N commits"
- Working tree: "uncommitted work", "uncommitted files", "staged changes", "clean working tree"
- Committed work: "N commits on this branch", "recent changes", "what changed"
- WIP: "in progress", "incomplete", "partially implemented"
- Quantities: "12 files modified, +450/-120 lines", "5 commits"
- Risks: "tests missing for X", "schema migration without rollback", "breaking change in public API"

## 2.5 Readiness States and Recovery

`readiness` is rendered verbatim from `zharness resume --json` — never a synthesized value. Exactly these four:

| Readiness | Meaning | Recovery / Next action |
|-----------|---------|------------------------|
| `clean` | No drift, nothing pending at the harness level | Continue WIP if present, else `/check review` or `/git cm` |
| `in-progress` | A run recorded, no clean check yet | Continue toward `/check full` for the active phase |
| `drifted` | `resume`'s `drift` array is non-empty | Print the first drift entry's `recovery` field verbatim — do not paraphrase |
| `no-harness` | No `.kit/harness.db` yet (valid snapshot, not an error) | `zharness import` if legacy `.kit/` artifacts exist, else `zharness init` |

Drift type → recovery pattern (from `cli/docs/STATE.md`, reproduced here so watzup's own text stays reviewable without cross-referencing the CLI docs on every recap):

| Drift type | Recovery pattern |
|------------|-------------------|
| `missing_file` | `to-plan phase {slug}` to regenerate, or `git checkout` if a tracked file was deleted accidentally |
| `unknown_phase` | `to-plan phase {slug}` to create the missing story, or correct `current_phase` via `to-plan` if it was a typo |
| `out_of_order` | Re-run `check full` against the latest run — the stale check stays in place, superseded by the new one |

## 3. Title Format

```
Recap — {branch} ({YYYY-MM-DD})
```

- `{branch}` is the branch under review (current branch unless overridden)
- `{YYYY-MM-DD}` is the run date, local timezone
- The em dash (`—`) is required; do not substitute `-` or `--`

## 4. Risk Table Contract

When risks are present, render exactly this table:

```
| Risk | Mức độ | Action |
|------|--------|--------|
| {short noun phrase} | {cao|vừa|thấp} | {single concrete action} |
```

Rules:
- Column order is fixed: `Risk`, `Mức độ`, `Action`
- Severity ladder: `cao`, `vừa`, `thấp`. No additions, no English equivalents in the cell.
- `Risk` column: short noun phrase, no leading verbs
- `Action` column: one concrete action that resolves the risk
- Zero risks → entire Risks section OMITTED. No "no risks" text, no empty table.

## 5. Output Layout

Console-only. Total length target: ≤ 25 visible lines.

Section order (omit a section entirely if it has no content, except the title which is mandatory):

1. **Title line** — exact format from Section 3
2. **Trạng thái** — short bullet list:
   - Branch name and position vs main
   - Uncommitted file count and line delta (if any)
   - Readiness: `clean` | `in-progress` | `drifted` | `no-harness` (verbatim from `resume --json`)
   - In harness repos: artifact chain status
3. **Context** — bullet list, max 2 items:
   - HANDOFF.md summary (where left off, key blocker) or "Không có handoff"
   - Artifact chain state (phase, latest work/check) if present
4. **Thay đổi** — bullet list, max 5 items, grouped by intent:
   - Committed change themes (from branch commits vs main)
   - WIP changes labeled with `[WIP]` prefix
   - Each bullet = one coherent change theme, not a file list
5. **Risks** — table per Section 4 (omitted if zero risks)
6. **Next** — single line, one concrete next action

**Empty-state branch**: when the working tree is clean AND there are no commits ahead of main AND no HANDOFF.md exists, output ONLY:

```
Nhánh sạch — không có thay đổi nào so với main.
Next: {one-line suggestion, e.g., "Bắt đầu task mới hoặc kéo thay đổi mới nhất."}
```

## 6. Self-Check (run before printing any output)

Before emitting output, validate:

1. The substring `git ` (with trailing space) is absent from the rendered output.
2. No `Quality: N/10` or `Score: N/10` appears.
3. No phrase from Section 1's forbidden list appears.
4. Title line matches the exact format from Section 3.
5. If a Risks section is rendered, its table columns are exactly `Risk | Mức độ | Action` and severity values are in `{cao, vừa, thấp}`.
6. If empty-state, output is exactly the two lines specified in Section 5.
7. Total output ≤ 25 visible lines.

If any check fails, fix the draft and re-run before printing.
