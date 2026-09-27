---
name: hunt
version: "1.1.0"
model: opus
description: "Find the root cause before any fix: build a red-capable repro loop, rank falsifiable hypotheses, prove, fix, sweep siblings. Also triages bug issues/PRs. Use on errors, crashes, regressions."
argument-hint: "[diagnose|fix|bisect|regression|sweep|triage] [symptom, error, issue/PR ref]"
compatibility: Designed for Claude Code
---

# Hunt: Diagnose Before You Fix

Prefix your first line with `🥷` inline. Verdict first: the root-cause sentence, then evidence. A patch applied to a symptom creates a new bug somewhere else.

## Outcome Contract

- Outcome: the root cause is identified and proven before any fix is applied.
- Done when: one sentence explains the cause, every observed symptom fits it, and the fix (or handoff) is verified by the same loop that reproduced the bug.
- Evidence: the loop command and its output, source trace, logs or state, targeted test/build output, runtime evidence for UI or native defects.
- Authorization: "diagnose", "investigate", "why", "look into", "debug", "xem thử", "sao lỗi" is report-only. Edit code only when the request asks to fix, change, or implement, or that authorization is still in force for the same unfinished task. A fix authorization never carries commit, push, publish, or anything destructive.

**Do not touch code until you can write:**
> "I believe the root cause is [X at file:line / condition] because [evidence]."

"A state management issue" is not testable. "Stale cache in `useUser` at `src/hooks/user.ts:42` because the dependency array omits `userId`" is. If you cannot be that specific, you do not have a hypothesis yet.

## Modes

Resolve from a leading subcommand, else from the request shape.

| Mode | Activate when | Difference from the default loop |
|---|---|---|
| `diagnose` (default) | Error, crash, test failure, "not working", "why" | Steps 1–4, report only |
| `fix` | "fix it", "sửa đi" | Full loop, steps 1–7 |
| `bisect` | "used to work", "broke after update", a known-good commit or version | Step 1 becomes a bisect harness, see Bisect below |
| `regression` | Same issue after a fix, or a "good" screenshot/version/file to compare against | The reference is the oracle, see Regression below |
| `sweep` | After a root-cause fix, or "anywhere else like this?", "còn chỗ nào giống vậy không" | Step 6 only, on a named pattern |
| `triage` | Issue/PR queue, "look at #42", "what needs my attention", label/close/brief an issue | Load `references/triage.md` |

## The Loop

### 1. Build a feedback loop (this is the skill)

With a tight pass/fail signal that goes red on *this* bug, bisection, hypotheses, and instrumentation all become mechanical. Without one, staring at code will not save you. Spend disproportionate effort here; load `references/feedback-loops.md` for the ten ways to build one, how to tighten it, and non-deterministic bugs.

Exit criterion: **one command you have already run**, shown with its (redacted) output, that is:

- Red-capable: it drives the real code path and asserts the user's exact symptom, not "didn't crash".
- Deterministic: same verdict every run (flaky bug: a pinned, high reproduction rate).
- Fast: seconds, not minutes.
- Agent-runnable: unattended; a human only through `scripts/hitl-loop.template.sh`.

Reading code to build a theory before this command exists is the exact failure this skill prevents. If you genuinely cannot build a loop, stop: list what you tried and ask for environment access, a redacted captured artifact (HAR, log, core dump, recording), or permission for temporary instrumentation. When the reporter's environment is the missing piece, hand them a read-only probe script (prints env, the disputed value, the state the hypothesis turns on; nothing secret) with one command to run and one block to paste back.

### 2. Reproduce and minimise

- Confirm the loop shows the failure the **user** described, not a nearby one. Wrong bug, wrong fix.
- List every symptom in the user's words, including the ones they wave off as unrelated.
- Cut inputs, callers, config, data, and steps one at a time, re-running after each cut. Done when every remaining element is load-bearing.

### 3. Hypothesise: 3–5 ranked, falsifiable

Single-hypothesis generation anchors on the first plausible idea. Each hypothesis states its prediction:
> "If X is the cause, then changing Y makes the bug disappear / changing Z makes it worse."

A hypothesis with no prediction is a vibe; sharpen or discard it. A hypothesis that explains only some symptoms is a symptom-level guess. Show the ranked list to the user before testing when they are around; they often re-rank instantly. Do not block on it.

### 4. Instrument, then confirm or discard

- Every probe maps to one prediction; change one variable at a time.
- Tool order: debugger/REPL, then targeted logs at the boundaries that separate hypotheses. Never "log everything and grep".
- Tag every temporary log with a unique prefix (`[DEBUG-a4f2]`) so cleanup is one grep. Full playbook: `references/logging-techniques.md`.
- Pick the instrument by bug class: pure logic (formula, off-by-one) → static reading; lifecycle/async/event-order → add the log while forming the hypothesis; rendering/compositor → DevTools layers before logs; performance → baseline numbers first, then bisect, then re-measure.
- Run the one probe that would fail if the hypothesis were wrong. If evidence contradicts it, discard it completely and re-orient on what the probe showed. If adding a log changes the behavior, that is timing/lifecycle evidence, not noise.

For a symptom that has recurred or smells like runtime state (caches, queues, generated output, PATH, locale, timeouts, entry points), read only `## Gotchas` in `references/failure-patterns.md`, then list its headings (`grep -n '^## '`) and read only the section matching the symptom, before adding a second fix. Domain traps: `references/ime-unicode.md` (IME, cursor drift, emoji), `references/rendering-debug.md` (PDF, print, fonts).

### 5. Fix with a regression guard

Write the regression test **before** the fix, at a **correct seam**: one where the test reproduces the real bug pattern as it occurs at the call site. If the only seam is too shallow to reproduce it, that is a finding: report that the architecture prevents locking this bug down, instead of adding false confidence.

1. Turn the minimised repro into a failing test at that seam; run it and watch it go red.
2. Apply the fix within the authorized scope; watch it go green.
3. Re-run the step 1 loop against the original, un-minimised scenario.

Test rules: expected values come from an independent source (a known-good literal, the spec, the reference), never recomputed the way the code does. A negative assertion ("output must not contain X") needs a paired positive case proving the assertion can fail. Red-green is **run**, not assumed; for red on unfixed code use a temporary worktree, never a stash or revert of the shared tree.

### 6. Sweep the blast radius

The same shape often hides in N other places. Extract the pattern signature (function, regex, API call, selector, missing lock, skipped validation, input boundary) and search it across the repo, excluding generated, build, and vendored paths; for class-of-bug patterns search the surrounding shape, not only the literal. For every match write one of: same bug / safe (why) / unsure (ask). Unrelated bugs the sweep surfaces are listed, not fixed, unless the user agrees.

### 7. Clean up

- The step 1 loop no longer reproduces; the regression test passes (or the missing seam is documented).
- `grep` for your `[DEBUG-` prefix returns nothing; throwaway harnesses are deleted or clearly marked.
- The hypothesis that proved correct goes into the commit/PR message, including why the bug recurred if it did.

## Bisect

- Protect the worktree first: `git status --short --branch -uall`. Any modified, staged, or untracked file means bisect runs in a temporary detached worktree, removed afterwards.
- If last-good is only a few releases back, read `git diff <good>..HEAD -- <suspect path>` first; bisect only when the diff is too large or the culprit is not obvious.
- Bisect with a non-interactive pass/fail command defined up front (`git bisect run <loop>`). Read the culprit diff down to the line, then `git bisect reset`.

## Regression

Treat the reference (last-good commit, old build, fixture, screenshot, described expected state) as the oracle, not decoration. Define the pass/fail check against it before editing, then name the exact current-vs-reference delta. Do not generalize a broken render, race, or state path into "style polish".

## Rationalization Smells

| Thought | Means |
|---|---|
| "Let me just try this" | No hypothesis. Write one first. |
| "I'm confident" | Run the probe that proves it. |
| "Probably the same issue" | Re-read the execution path from scratch. |
| "Works on my machine" | Enumerate environment differences before dismissing. |
| "One more restart" | Read the last error verbatim; never restart twice without new evidence. |
| "The log says it's fine" | Trust the user's observation; the gap is an un-instrumented path. |

## Hard Rules

- **Same symptom after a fix is a hard stop.** The hypothesis is unfinished; re-read the execution path before touching code again.
- **Three failed hypotheses → stop** and use the Handoff format. Ask how to proceed.
- **Measure the lower layer first** for system or tooling symptoms: OS capture vs post-processing, service vs UI, toolchain vs assertion, network vs client handling.
- **External tool failure: diagnose before switching** (server up? key valid? config right?).
- **Magic number tuned three times: stop.** The bug is structural; unify the values into one named token and find the missing constraint.
- **Performance claims need before/after numbers.** "Feels faster" is not evidence.
- **Redact every secret** in commands, outputs, and artifacts you show (`<REDACTED>`); build loops against env vars.
- **Fix the cause inside the authorized scope.** A prerequisite refactor is fine once you state why; an unrequested feature or unrelated refactor is not.

## Output

Open with one plain line: the outcome and whether changes are committed.

```
Root cause:        [what was wrong, file:line]
Loop:              [the red-capable command]
Fix:               [what changed, file:line]  (diagnose mode: "not applied")
Sibling sweep:     [N sites checked: N fixed / safe / unsure]  or  [not run, why]
Confirmed:         [loop red → green, original scenario re-run]
Regression guard:  [test file:line, red run shown]  or  [none: no correct seam, why]
```

Status: **resolved**, **resolved with caveats** (state them), or **blocked**.

**Handoff (blocked, or after three failed hypotheses):** the symptom in one sentence; each hypothesis with its probe and why it was ruled out; evidence collected (log excerpts, repro steps, versions, config); what is still unknown; next steps naming any tool, access, or context the user must supply.

Defer to: `check` for review of a diff with no concrete symptom; `think` for "should this exist / is it worth it"; `brainstorm` when the fix needs a product decision; `git` to commit after a fix.
