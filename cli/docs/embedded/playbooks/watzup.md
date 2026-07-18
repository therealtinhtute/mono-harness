# Playbook: watzup

## Purpose

Session recap. Answer one question: where is this branch, what state is the code in, what should happen next — regardless of whether code is committed or not. Read everything available (git state, diffs, handoff, artifact chain via `resume`), summarize concisely, and recommend one concrete next action. Read-only: this playbook never implements code, runs gates, writes files, or modifies artifacts.

## Preconditions

- **Version gate**: run `zharness --version` before anything else. A `dev` build always satisfies the gate. Otherwise, if the binary is missing or below `0.1.0` (`MIN_ZHARNESS_VERSION`), print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and stop.
- If the gate passes, run `zharness resume --json`. A `readiness: "no-harness"` response is a valid successful snapshot, not an error — it means no `.kit/harness.db` exists yet. Route on it, do not fall back to independent prose re-derivation: if `.kit/` already has legacy planning artifacts, recommend `zharness import`; otherwise recommend `zharness init` (fresh project) followed by `brainstorm`/`to-plan`. A `db_unreadable` exit (code 2) is a real error (DB present but unreadable/corrupt) — surface it directly, do not silently treat it as `no-harness`.

## Arguments

- `[branch]` — branch under review (default: current branch)

## When to Use

- Start of a new session — orient before coding
- Resuming after a break or context switch
- Quick status check on any branch

## Steps

### Step 1 — Branch State

```bash
git status -sb
git log --oneline main..HEAD
git rev-list --left-right --count main...HEAD
```

Extract: branch name, commits ahead/behind main, working tree cleanliness (staged, unstaged, untracked counts).

### Step 2 — Load Harness State

Run `zharness resume --json` (already called once by the version gate for routing — reuse that output, don't call twice). Extract, verbatim, no re-derivation:
- `position.current_phase`, `position.status`
- `latest_run_id`, `latest_check_id`, `latest_handoff_id`
- `drift` — array of `{type, detail, recovery}`
- `readiness` — one of `clean | in-progress | drifted | no-harness`

This snapshot is the single source of truth for phase/run/check/handoff state — do not additionally read `.kit/planning/ROADMAP.md`, phase `-CONTEXT.md`/`-PLAN.md`, run logs, or check reports to reconstruct state; `resume` already resolved them.

Read `.kit/HANDOFF.md` if present — this is narrative only (where left off, human blocker description, `→ START HERE` action), not a state source; its `id`/`run_id`/`check_id` should already match `resume`'s `latest_handoff_id`/`latest_run_id`/`latest_check_id`. If they don't, that's drift — add it to the Risks section even if `resume`'s own `drift` array missed it.

### Step 3 — Committed Work Summary

From `git log --oneline main..HEAD`: group commits by type (feat/fix/refactor/etc.), identify change themes. Max 3 themes.

From the diff between the branch and main: total files and line delta.

### Step 4 — WIP Analysis

From the working-tree and staged diffs: identify uncommitted files and line delta.

Read the actual diff content for uncommitted changes. Look for:
- Incomplete implementations (TODO, FIXME, HACK, partial functions)
- Quality signals (missing error handling at boundaries, hardcoded values, dead code from this change)
- What the WIP is trying to accomplish (change intent)

Cap analysis at the top 5 most significant changed files if the diff is large.

### Step 5 — Risk Assessment

Flag issues from git-derived signals AND from `resume`'s `drift` array — map each drift entry's `type` to its recovery per the Drift → Recovery table below (do not invent recovery text; use the `recovery` field `resume` already returned):

| Signal | Default severity |
|--------|-----------------|
| Missing tests for new behavior | vừa |
| Breaking changes in public API | cao |
| Large uncommitted diff (> 200 lines) | vừa |
| `drift: missing_file` | vừa |
| `drift: unknown_phase` | vừa |
| `drift: out_of_order` | vừa |
| `readiness: no-harness` on a repo with existing `.kit/` artifacts | cao |
| Explicit blockers from HANDOFF.md | cao |
| Hardcoded credentials or secrets | cao |
| Schema/migration without rollback | cao |

If zero risks, omit the Risks section entirely.

### Step 6 — Next Action

Based on all evidence, recommend ONE concrete next action. `readiness` from `resume` drives the primary branch; git WIP state breaks the tie within `clean`/`in-progress`:

| State | Recommended action |
|-------|-------------------|
| `readiness: no-harness`, legacy `.kit/` present | `zharness import` |
| `readiness: no-harness`, no `.kit/` artifacts | `zharness init`, then `brainstorm` |
| `readiness: drifted` | Follow the first drift entry's `recovery` field verbatim |
| `readiness: clean` or `in-progress`, WIP present | Continue the in-progress work (name the specific file/function) |
| `readiness: clean` or `in-progress`, no WIP, HANDOFF.md has `→ START HERE` | Follow that action |
| `readiness: clean` or `in-progress`, no WIP, no HANDOFF.md action | `check review` or a commit/push request |

## Readiness State

Render `resume --json`'s `readiness` field verbatim — one of:

- `clean` — no drift, no pending work at the harness level
- `in-progress` — a run recorded, no clean check yet, or check pending review
- `drifted` — `resume`'s `drift` array is non-empty
- `no-harness` — no `.kit/harness.db` yet (valid snapshot, not an error)

Never derive this value independently from git state or file reads — it comes from `resume --json` only.

## Output Contract

This section is the single source of truth for what this playbook may print. Every rendered output must conform to these rules.

### 1. Forbidden Phrases

Output MUST NOT contain any of the following:
- Shell or git commands: `git log`, `git diff`, `git status`, `git branch`, `git show`, `--stat`, `--shortstat`, `--oneline`, `HEAD~`, `..HEAD`, `--graph`, `--decorate`
- Process descriptors: `commit window`, `diff stat`, `last 10 commits`, `last 50 commits`, `analyzed`, `scanned`
- Synthesized scoring: any `Quality: N/10` pattern, any `Score: N/10`, any `/10` rating
- Playbook-internal terms: `recap mode`, `orient mode` (when describing behavior to the user — naming the invocation is fine)

**Trailing-space check**: the literal substring `git ` (with a trailing space) MUST NOT appear in any rendered output. Project nouns containing `git` without a trailing space (e.g., a directory called `git-tools`) are exempt.

### 2. Allowed Vocabulary

Use project-language phrases. Numbers describe the work, not the process.

- Branch references: "this branch", "branch X", "ahead of main by N commits", "behind main by N commits"
- Working tree: "uncommitted work", "uncommitted files", "staged changes", "clean working tree"
- Committed work: "N commits on this branch", "recent changes", "what changed"
- WIP: "in progress", "incomplete", "partially implemented"
- Quantities: "12 files modified, +450/-120 lines", "5 commits"
- Risks: "tests missing for X", "schema migration without rollback", "breaking change in public API"

### 2.5 Readiness States and Recovery

`readiness` is rendered verbatim from `zharness resume --json` — never a synthesized value. Exactly these four:

| Readiness | Meaning | Recovery / Next action |
|-----------|---------|------------------------|
| `clean` | No drift, nothing pending at the harness level | Continue WIP if present, else `check review` or a commit/push request |
| `in-progress` | A run recorded, no clean check yet | Continue toward `check full` for the active phase |
| `drifted` | `resume`'s `drift` array is non-empty | Print the first drift entry's `recovery` field verbatim — do not paraphrase |
| `no-harness` | No `.kit/harness.db` yet (valid snapshot, not an error) | `zharness import` if legacy `.kit/` artifacts exist, else `zharness init` |

Drift type → recovery pattern:

| Drift type | Recovery pattern |
|------------|-------------------|
| `missing_file` | `to-plan phase {slug}` to regenerate, or revert if a tracked file was deleted accidentally |
| `unknown_phase` | `to-plan phase {slug}` to create the missing story, or correct `current_phase` via `to-plan` if it was a typo |
| `out_of_order` | Re-run `check full` against the latest run — the stale check stays in place, superseded by the new one |

### 3. Title Format

```
Recap — {branch} ({YYYY-MM-DD})
```

- `{branch}` is the branch under review (current branch unless overridden)
- `{YYYY-MM-DD}` is the run date, local timezone
- The em dash (`—`) is required; do not substitute `-` or `--`

### 4. Risk Table Contract

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

### 5. Output Layout

Console-only. Total length target: ≤ 25 visible lines.

Section order (omit a section entirely if it has no content, except the title which is mandatory):

1. **Title line** — exact format from Section 3
2. **Trạng thái** — short bullet list: branch name and position vs main; uncommitted file count and line delta (if any); readiness (`clean`|`in-progress`|`drifted`|`no-harness`, verbatim from `resume --json`); in harness repos, artifact chain status
3. **Context** — bullet list, max 2 items: HANDOFF.md summary (where left off, key blocker) or "Không có handoff"; artifact chain state (phase, latest work/check) if present
4. **Thay đổi** — bullet list, max 5 items, grouped by intent: committed change themes (from branch commits vs main); WIP changes labeled with `[WIP]` prefix; each bullet = one coherent change theme, not a file list
5. **Risks** — table per Section 4 (omitted if zero risks)
6. **Next** — single line, one concrete next action

**Empty-state branch**: when the working tree is clean AND there are no commits ahead of main AND no HANDOFF.md exists, output ONLY:

```
Nhánh sạch — không có thay đổi nào so với main.
Next: {one-line suggestion, e.g., "Bắt đầu task mới hoặc kéo thay đổi mới nhất."}
```

### 6. Self-Check (run before printing any output)

Before emitting output, validate:
1. The substring `git ` (with trailing space) is absent from the rendered output.
2. No `Quality: N/10` or `Score: N/10` appears.
3. No phrase from Section 1's forbidden list appears.
4. Title line matches the exact format from Section 3.
5. If a Risks section is rendered, its table columns are exactly `Risk | Mức độ | Action` and severity values are in `{cao, vừa, thấp}`.
6. If empty-state, output is exactly the two lines specified in Section 5.
7. Total output ≤ 25 visible lines.

If any check fails, fix the draft and re-run before printing.

## Examples

### Example 1 — Harness repo, in-progress with WIP

**Scenario**: mid-phase, harness present, uncommitted WIP, HANDOFF.md exists.

```
Recap — feature/inbox-ui (2026-07-18)

Trạng thái
- Nhánh: feature/inbox-ui, 4 commits ahead of main
- Uncommitted: 3 files, +45/-12 lines
- Readiness: in-progress

Context
- Handoff: phase inbox-ui hoàn tất wave 1, đang chờ verify wave 2
- Phase: inbox-ui | latest run: recorded | latest check: chưa có

Thay đổi
- Hệ thống list/detail view cho inbox
- API integration cho inbox endpoints
- [WIP] Filter dropdown logic — chưa hoàn thành

Next: Hoàn thành filter dropdown rồi chạy check full cho phase inbox-ui.
```

### Example 2 — Harness repo, drifted

**Scenario**: `resume --json` returns a non-empty `drift` array with an `unknown_phase` entry.

```
Recap — feature/triage-rules (2026-07-18)

Trạng thái
- Nhánh: feature/triage-rules, 2 commits ahead of main
- Uncommitted: 1 file
- Readiness: drifted

Context
- Handoff: dừng ở phase triage-rules, chờ xác nhận scope
- Phase: triage-rules | drift: unknown_phase

Thay đổi
- Triage rule engine cơ bản
- [WIP] Rule configuration

Risks
| Risk | Mức độ | Action |
|------|--------|--------|
| Phase triage-rules chưa có story tương ứng | vừa | to-plan phase triage-rules để tạo story còn thiếu |

Next: to-plan phase triage-rules để tạo story còn thiếu (recovery verbatim từ resume).
```

### Example 3 — Empty state

```
Nhánh sạch — không có thay đổi nào so với main.
Next: Bắt đầu task mới hoặc kéo thay đổi mới nhất.
```

### Example 4 — No harness yet, legacy artifacts present

**Scenario**: `resume --json` returns `readiness: "no-harness"`, and `.kit/planning/` already has legacy files.

```
Recap — feature/legacy-migrate (2026-07-18)

Trạng thái
- Nhánh: feature/legacy-migrate, 1 commit ahead of main
- Working tree: sạch
- Readiness: no-harness

Context
- Không có handoff
- .kit/planning/ tồn tại nhưng chưa có harness.db

Thay đổi
- Cập nhật tài liệu planning ban đầu

Next: zharness import để nạp các artifact hiện có vào harness.
```

## Command Reference

- `zharness --version` — version gate
- `zharness resume --json` — the single source of truth for phase/run/check/handoff/drift/readiness state

## Exit / Handoff Conditions

Complete only when: the version gate passed; `resume --json` was called exactly once and its `readiness` value rendered verbatim; the self-check in Section 6 passed; total output stays within the 25-line target (or the exact two-line empty-state form was used).

## Anti-Patterns

- Saying `readiness: clean` without it coming from `resume --json` — optimistic self-certification
- Copying commit messages as the summary — that's raw log output, not a recap; summarize themes
- Skipping WIP analysis because "it's just uncommitted" — uncommitted code is the most actionable part of a recap
- Ignoring HANDOFF.md when it exists — the whole point of recap is to bridge sessions
- Re-deriving phase/drift state from planning files instead of `resume --json` — `resume` already resolved them; reading them independently risks disagreeing with the canonical snapshot
