# hunt: fix, sweep, and clean up (steps 5–7)

Loaded by `hunt` at the end of step 4 for `fix`, and for `sweep` (step 6 only). `diagnose` never loads it. Steps 1–4 stay in `SKILL.md`.

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
