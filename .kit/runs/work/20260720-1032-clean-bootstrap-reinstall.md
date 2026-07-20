---
id: 01KXYSZ5TJ8E33QNA2WZ91YYS2
type: run
phase: none
plan_id: none
trace_ids: []
lane: normal
mode: simple
plan: .kit/plans/2026-07-20-clean-bootstrap-reinstall/PLAN.md
started: 2026-07-20 10:32 +07:00
created: 2026-07-20
updated: 2026-07-20
status: completed
---

# Work Run — Clean Bootstrap and Harness Reinstallation

## Scope

Execute the approved local cleanup and reinstall plan. Preserve `workflow-core`, update shared skills, install `zharness v0.4.1`, reset old project harness history, initialize a fresh harness, verify all stages, and stop before staging or committing.

## Task status

| Task | Status | Evidence |
|---|---|---|
| Preflight and recovery inventory | DONE | 123 working-tree entries, all under `.kit/`; `master` equals `origin/master`; all eight installed workflow hashes differ from source; protected config fingerprints captured; release `v0.4.1` and Linux AMD64 asset verified |
| Remove stale workflow skills | DONE | Seven shared folders plus standalone Claude `to-plan` moved to trash; unrelated skill inventory preserved |
| Refresh full Claude bootstrap | DONE | Bootstrap stamped at `9e848ac`; settings and workflow-core hashes unchanged; pointer restored; all eight hashes match source; 22 links per shared skill resolve with zero broken |
| Install zharness v0.4.1 | DONE | Release checksum matched; explicit `/usr/bin/install` bypassed a shell alias; binary resolves from `~/.local/bin`, mode `755`, version `0.4.1`; temp directory trashed |
| Initialize fresh project harness | DONE | `zharness init --json` created schema v2 DB; six v0.4.1 playbooks scaffolded and ignored; one new docs-version metadata changeset created; state and resume are empty with readiness `clean`; no legacy state file or historical entity reappeared |
| Final proof | DONE_WITH_CONCERNS | Eight installed skill hashes match source; six spine triggers are 22–24 lines and delegate correctly; all recorded shared links resolve; bootstrap files and stamp verify; `zharness` is v0.4.1; state/resume are empty and clean; audit has zero drift. `validate` exits 1 only because a deliberately empty pre-intake harness has no `.kit/planning/SPEC.md`; this is the sole explained baseline finding |

## Guardrails

- Use `trash`, never `rm`.
- Do not overwrite `~/.claude/settings.json`.
- Preserve `~/.claude/rules/workflow-core.md` and its CLAUDE.md pointer.
- Do not modify project files outside `.kit/`.
- Do not stage, commit, push, or create a PR.

## Final verification

- Installed `SKILL.md` hashes match repository sources for all eight workflow skills.
- The six spine skills are thin triggers at 22–24 lines and point to their matching generated playbooks.
- Generated playbooks match the v0.4.1 embedded playbooks byte-for-byte.
- All 22 recorded shared links resolve for each broadly shared workflow skill; `to-plan` now resolves through its shared folder from Claude Code.
- Bootstrap-managed rules, hooks, statusline, and version stamp match current `master`; `settings.json` remains present and valid; `workflow-core.md` and its single CLAUDE.md pointer remain present.
- `zharness` resolves to `~/.local/bin/zharness`, has mode `0755`, and reports version `0.4.1`.
- `query state` reports no phase, run, or check; `resume` reports no phase, run, check, or handoff, zero drift, and readiness `clean`.
- `audit` exits 0 with zero pointer drift and entropy 5. Its sole contract violation, also returned by `validate`, is the expected missing `.kit/planning/SPEC.md` on a deliberately empty pre-intake harness; no run-link violation remains.
- Repository status is limited to 121 intentional old `.kit/` deletions and three new `.kit/` artifacts (metadata changeset, plan, run); generated docs and DB are ignored; no files outside `.kit/` changed.
- `HEAD` still equals `origin/master`; the index is empty. No staging, commit, push, or PR was performed.
