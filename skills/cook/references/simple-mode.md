# Simple Mode — Lightweight Execution Workflow

Activate when mode resolves to `simple` (explicit arg or auto-detected). Do NOT touch `.planning/`.

## When Simple Mode Is Appropriate

- Fix or feature scope is known upfront: clear file(s), clear change, clear success criterion
- Input source is a direct prompt or a brainstorm explore file (`.kit/reports/brainstorm/*.md`)
- Work does NOT require multi-phase coordination, multi-team visibility, or schema migrations
- Root cause is already known — if the bug cause is unknown, route to `/hunt` first

## When to Reject Simple Mode

- Task clearly touches > 5 files, > 100 lines, or enters an unfamiliar subsystem (even if user says "just do it")
- Task is a bug with unknown root cause — use `/hunt`
- Task is a scope-defining question — use `/brainstorm`
- Task is a complete feature with multiple phases — use `cook full`

---

## 7-Step Workflow

### Step 1 — Prompt Quality Check

Evaluate the input immediately. A good prompt answers:
- What to change (file, function, behavior)
- Where (file path or subsystem name)
- What "done" looks like (visible output, test passing, UI state)

**If the prompt is thin** (< 1 sentence or missing 2+ of the above):
- Ask via `AskUserQuestion` — 1 to 2 targeted questions, not an open-ended dump
- Example questions: "Which file or component should change?", "What should the output look like after the fix?"

**If still thin after AskUserQuestion**:
- Stop and suggest `/prompt-leverage` to strengthen the prompt before re-invoking `/cook simple`

**Prompt quality rubric:**

| Signal | Verdict |
|--------|---------|
| File path + change description + success criterion | Proceed |
| File path + change description, no success criterion | Proceed — infer criterion from description |
| Subsystem name + vague change | Ask 1 question about the specific change |
| "Fix the bug" / "Make it better" | Ask: which file, what behavior is wrong |

### Step 2 — Quick Research

Minimal read pass — only what is needed to execute safely.

- Read each explicitly referenced file
- Grep for the symbol or pattern being changed (1-2 targeted greps)
- Skim `CLAUDE.md` or nearest convention file for code style rules
- Find 1 similar existing implementation as a pattern reference if adding new behavior

Hard limit: **5 file reads maximum**. If more are needed, the task likely exceeds simple mode scope.

### Step 3 — State Approach

Output 2-3 sentences inline before touching any code:
- What changes (specific function, line, or block)
- Where it lives (file:line if known)
- Why this approach (not an alternative)

No SPEC. No design doc. No options list. One sentence per point.

### Step 4 — Scope Guard (hard stop)

After Step 2, count the impact:

| Threshold | Action |
|-----------|--------|
| ≤ 5 files AND ≤ 100 lines AND known subsystem | Proceed to Step 5 |
| > 5 files OR > 100 lines OR unknown subsystem | **HARD STOP** — emit `scope-guard` message from `routing.md`, do not continue |

The scope guard does not negotiate. Even if the user says "just do it", the stop is unconditional — they must explicitly invoke `cook full` to proceed past this gate.

### Step 5 — Execute

- Prefer inline edits for 1-3 line changes in a single file
- Use a subagent only for one heavy isolated task (e.g., generate a new file from a clear template); never for exploration
- Do not add waves, phases, or task lists — this is single-shot execution
- Do not write to `.planning/` — write nothing outside the files identified in Step 2-3
- If a new assumption surfaces mid-execution that would expand scope past the guard, stop and surface it

### Step 6 — Light Verify

Run the narrowest possible check that proves the change works:
- Run the test file(s) most relevant to the changed code
- Run lint on changed files if the project has a fast lint command
- If no test exists for the changed code, state that explicitly — do not claim "tests pass" vacuously

Capture the verification output verbatim. "It should work" is not a verification result.

### Step 7 — Handoff Suggest

After verification:
1. State what changed (files, line delta)
2. State verification result (command + output)
3. Suggest next move — pick the most natural one:
   - `/check review` — if the change is non-trivial or touches a security boundary
   - `/git cm` — if the change is clean and ready to commit
   - `/handoff` — if stopping for the day
   - Nothing — if it was a one-liner the user will handle themselves

Never auto-run these. Suggest, then stop.

---

## Simple Mode Output Log (optional)

If the project has a `.kit/reports/` directory, write a one-line summary to:
`.kit/reports/cook/{YYYYMMDD}-{slug}.md`

Format:
```yaml
---
date: YYYY-MM-DD
mode: simple
input: {one-line description of prompt or @file ref}
files_changed: [path1, path2]
lines_delta: +N -N
verification: {command} → pass / fail
---
```

This log is optional but helps `/watzup` capture simple-mode work alongside full-pipeline phases.

---

## Anti-Patterns

- Accepting a thin prompt and guessing scope — always ask via `AskUserQuestion` first
- Writing to `.planning/` — simple mode does not own planning artifacts
- Dispatching a subagent for exploration — research happens inline in Step 2
- Treating scope guard as advisory — it is unconditional
- Saying "tests pass" without running them — always capture actual output
