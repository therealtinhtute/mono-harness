# Playbook: watzup

## Purpose

Session recap. Answer one question: where is this branch, what state is the code in, what should happen next — regardless of whether code is committed or not. Read everything available (git state, diffs, handoff, artifact chain via `resume`), summarize concisely, and recommend one concrete next action. Read-only: this playbook never implements code, runs gates, writes files, or modifies artifacts.

## Preconditions

- **Version gate**: run `zharness --version` before anything else. A `dev` build always satisfies the gate. Otherwise, if the binary is missing or below `0.1.0` (`MIN_ZHARNESS_VERSION`), print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and stop.

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

Run `zharness resume --json` once to fetch `readiness`, `drift`, `position`, and the run/check/handoff IDs. Read `.kit/HANDOFF.md` if present for a one-line summary of where work left off.

Then assemble the facts JSON from Steps 1/3/4/5/6 below and call `zharness resume --facts '<json>'` exactly once more — its stdout is the final answer, printed verbatim. Do not reformat, re-derive, or add text around it: forbidden-phrase safety, the risk-table shape, the title/section format, the 25-line empty-state form, and the drifted-state recovery override are all enforced by the CLI itself. If it returns an error (`invalid_severity`, `forbidden_phrase`, `facts_malformed`), fix the offending field in the facts JSON and retry — do not work around it by hand-formatting output instead.

Facts JSON shape:
```json
{
  "branch": "feature/inbox-ui",
  "ahead": 4,
  "behind": 0,
  "uncommitted_files": 3,
  "uncommitted_adds": 45,
  "uncommitted_dels": 12,
  "handoff_summary": "phase inbox-ui hoàn tất wave 1, đang chờ verify wave 2",
  "changes": ["Hệ thống list/detail view cho inbox", "API integration cho inbox endpoints"],
  "wip": ["Filter dropdown logic — chưa hoàn thành"],
  "risks": [{"risk": "thiếu test cho filter", "severity": "vừa", "action": "viết test trước khi merge"}],
  "next_action": "Hoàn thành filter dropdown rồi chạy check full cho phase inbox-ui."
}
```
`handoff_summary` empty string means no handoff. `next_action` is ignored by the CLI when `readiness` came back `drifted` — it prints the first drift entry's recovery instead, so it's safe to fill in your best normal-state recommendation regardless.

If `resume --json`'s `readiness` was `no-harness`: still call `--facts` the same way; it renders a generic `no-harness` recap. Route the actual recommendation (`zharness import` if legacy `.kit/` artifacts exist, else `zharness init` then `brainstorm`) via `next_action`.

### Step 3 — Committed Work Summary

From `git log --oneline main..HEAD`: group commits by type (feat/fix/refactor/etc.), identify change themes. Max 3 themes — these become `facts.changes`.

From the diff between the branch and main: total files and line delta.

### Step 4 — WIP Analysis

From the working-tree and staged diffs: identify uncommitted files and line delta (`facts.uncommitted_*`).

Read the actual diff content for uncommitted changes. Look for:
- Incomplete implementations (TODO, FIXME, HACK, partial functions)
- Quality signals (missing error handling at boundaries, hardcoded values, dead code from this change)
- What the WIP is trying to accomplish (change intent)

Cap analysis at the top 5 most significant changed files if the diff is large. Each coherent change theme becomes one `facts.wip` entry (prefix `[WIP]` is added by the CLI, don't add it yourself).

### Step 5 — Risk Assessment

Decide which signals count as a risk — from git-derived signals AND from `resume --json`'s `drift` array (each drift entry is itself a risk row: use its `type` as the risk noun and its own `recovery` field as the action, don't invent one):

| Signal | Default severity |
|--------|-----------------|
| Missing tests for new behavior | vừa |
| Breaking changes in public API | cao |
| Large uncommitted diff (> 200 lines) | vừa |
| `drift: missing_file` / `unknown_phase` / `out_of_order` | vừa |
| `readiness: no-harness` on a repo with existing `.kit/` artifacts | cao |
| Explicit blockers from HANDOFF.md | cao |
| Hardcoded credentials or secrets | cao |
| Schema/migration without rollback | cao |

Zero risks → pass an empty `facts.risks` array (the CLI omits the section itself).

### Step 6 — Next Action

Based on all evidence, decide ONE concrete next action for `facts.next_action`. `readiness` drives the primary branch; git WIP state breaks the tie within `clean`/`in-progress`:

| State | Recommended action |
|-------|-------------------|
| `readiness: no-harness`, legacy `.kit/` present | `zharness import` |
| `readiness: no-harness`, no `.kit/` artifacts | `zharness init`, then `brainstorm` |
| `readiness: drifted` | (ignored — CLI prints the drift recovery instead) |
| `readiness: clean` or `in-progress`, WIP present | Continue the in-progress work (name the specific file/function) |
| `readiness: clean` or `in-progress`, no WIP, HANDOFF.md has `→ START HERE` | Follow that action |
| `readiness: clean` or `in-progress`, no WIP, no HANDOFF.md action | `check review` or a commit/push request |

## Command Reference

- `zharness --version` — version gate
- `zharness resume --json` — harness state snapshot (phase/run/check/handoff/drift/readiness)
- `zharness resume --facts '<json>'` — renders the full Recap text; see Step 2 for the shape

## Exit / Handoff Conditions

Complete only when: the version gate passed; `resume --json` was called once for state and `resume --facts` was called exactly once to render; its stdout was printed verbatim with no manual reformatting.

## Anti-Patterns

- Hand-formatting the recap text instead of calling `resume --facts` — that's exactly what the CLI now guarantees correctness on
- Saying `readiness: clean` without it coming from `resume --json` — optimistic self-certification
- Copying commit messages as `facts.changes` — that's raw log output, not a summary; summarize themes
- Skipping WIP analysis because "it's just uncommitted" — uncommitted code is the most actionable part of a recap
- Ignoring HANDOFF.md when it exists — the whole point of recap is to bridge sessions
- Re-deriving phase/drift state from planning files instead of `resume --json` — `resume` already resolved them
