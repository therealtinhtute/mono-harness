# Playbook: work (bounded mode)

Loaded by [work.md](work.md) for `bounded`.

**Zero-write rule:** bounded/simple mode creates no plans, reports, changesets, or markdown artifacts and never edits an active plan. Evidence: the diff plus captured proof in the response, with the delivery state stated as implemented or locally verified; commit and later states run only through `git` on the owner's explicit request.

1. **Scope** — name the files to change and the one command that proves the change. No known check → stop and route to `brainstorm`.
2. **Baseline** — run that command once before editing; a failure that predates your work is reported, not fixed silently.
3. **Change** — read before editing. A behavior change runs **red → green**: one failing test that asserts what users observe, then the least code that passes. Never delete, skip, or weaken a test to get green; a step you choose not to run is reported as `skip: <reason>`.
4. **Verify** — re-run the command; on failure make one targeted fix, and the same failure again without new evidence → stop and report it.
5. **Report** — the diff summary, the command and its output, and the delivery state.
