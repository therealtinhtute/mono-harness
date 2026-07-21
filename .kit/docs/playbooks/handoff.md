# Playbook: handoff

## Purpose

Prospective session continuity: capture current session state into `.kit/HANDOFF.md` so the next session (or agent, or developer) can resume without re-deriving context. When harness artifacts exist, anchor the handoff to phase state, work run evidence, and the latest quality-gate verdict. Focus on what the next session needs to pick up — not on re-reviewing the work; that is `check`'s job.

## Preconditions

- **Version gate**: run `zharness --version` before anything else. A `dev` build always satisfies the gate. Otherwise, if the binary is missing or below `0.1.0` (`MIN_ZHARNESS_VERSION`), print `zharness not found or out of date — run: bash scripts/install-zharness.sh` and stop.

## When to Use

- End of a session — capturing state for the next session
- Before a context switch or long break
- Handing off to another developer or agent
- When session context is at risk of being lost

Sacrifice grammar for concision, keep token use lean, and focus on actionable context.

## Arguments

- `[context]` — optional additional context to include in the handoff

## Steps

### Step 1 — Capture Git State

```bash
git status --short --branch
git log --oneline --graph --decorate -5
git diff --stat
git diff --cached --stat
```

Extract: current branch, upstream status, uncommitted changes, untracked files, recent commits, and working-tree summary.

### Step 2 — Identify Active Work

From git status: files being modified, new files being added, files staged for commit.

From recent commits: feature/fix being worked on, scope of changes, progress indicators.

From harness artifacts (preferred when present), in priority order:
1. `.kit/planning/ROADMAP.md` — active phase order
2. phase `-CONTEXT.md` / `-PLAN.md` — locked decisions and remaining tasks; capture decisions/assumptions/rejected-options the next session must not rediscover, plus current wave/task state and remaining work
3. latest `.kit/runs/work/*.md` — task statuses, blockers, and proof trail
4. latest `.kit/reports/check/*.md` — whether the phase gate passed, drifted, or failed

From task tracking (fallback): `find . -name "todo.md" -o -name "tasks.md" -o -name "HANDOFF.md" | head -3`. Read existing task files to understand planned work, completed items, pending items, and blockers.

### Step 3 — Capture Context

**Technical context**: what is being built/fixed; current implementation approach; key decisions made; technologies/frameworks involved.

**Progress context**: what was completed this session; what is in progress; what is blocked; what is next; which phase / run / gate verdict the next session should resume from.

**Environment context**: dependencies installed/updated; configuration changes; environment variables needed; external services involved.

### Step 4 — Identify Blockers

List what is blocked, why it is blocked, and what is needed to unblock. Be specific — vague blockers help no one.

Common blocker types: missing information or requirements; external dependencies (APIs, services); technical challenges or unknowns; failing tests or build errors; merge conflicts; waiting for review/approval.

For each blocker: describe the issue, state what's needed to unblock, suggest next steps, preserve blocker taxonomy when known (`BLOCKED_CONTEXT`, `BLOCKED_SCOPE`, `BLOCKED_VERIFICATION`, `BLOCKED_CONTRACT_DRIFT`).

### Step 5 — Document Next Steps

List 3-5 prioritized actions. Mark the most important one with `→ START HERE`. Each action: verb + file/command + expected outcome. In harness flows, the first action should point to the exact phase, run artifact, or gate result to resume or resolve.

### Step 6 — Record the Handoff Entity, Then Write HANDOFF.md

When the version gate passes, the entity is canonical and `.kit/HANDOFF.md` is its narrative — write the entity first, then the markdown carries its ULID. The `handoffs` entity (written by `zharness handoff record`) is what `resume`/`watzup` read to reconstruct state; never let the markdown diverge from it.

1. Resolve anchors: latest run ULID and latest check ULID (from `zharness resume --json`'s `latest_run_id`/`latest_check_id`), and the open-items list (unresolved blockers / next steps from Steps 4-5).
2. `zharness handoff record --run-id {ulid|omit} --check-id {ulid|omit} --open-items '["...", ...]' --json` → capture the returned `id`.
3. Write `.kit/HANDOFF.md` using the template below, with the entity's `id` in frontmatter (never invent a fresh ULID for it) and `run_id`/`check_id` mirroring what was just recorded. Minimum sections: **Branch**, **Completed**, **In Progress**, **Blockers**, **Next Steps**.
4. When harness artifacts exist, also include `continuity_mode`, `active_phase`, `latest_cook_run`, `latest_check_verdict`, and unresolved concerns or proof gaps.

If the version gate did not pass or `.kit/planning/` does not exist, skip step 2 and write `.kit/HANDOFF.md` alone.

Do not lose: blocker taxonomy from `work` (`BLOCKED_CONTEXT`, `BLOCKED_SCOPE`, `BLOCKED_VERIFICATION`, `BLOCKED_CONTRACT_DRIFT`); artifact drift or proof-gap findings from `check`; any plan boundary the next session must stay inside.

Continuity mode: say `standard` if harness artifacts are missing (fall back to git state + recent commits + working tree only); say `partial-harness` if artifacts partially exist (name which source is missing instead of implying continuity is complete); say `full-harness` when the full chain (roadmap, phase, run, check) is present.

### Step 7 — Verify Handoff Quality

Check: branch state captured? blockers specific? next action has a clear first step? Sensitive data sanitized?

Completeness checklist:
- [ ] Current state clearly documented
- [ ] Progress tracked (completed/in-progress/pending)
- [ ] Blockers identified with unblock criteria
- [ ] Next steps are actionable
- [ ] Technical context sufficient for continuation
- [ ] Continuity anchors captured (phase, latest work run, latest check verdict) when available
- [ ] No sensitive data exposed

## Artifacts

### HANDOFF.md — `.kit/HANDOFF.md`

Emit the skeleton with `zharness scaffold handoff --path .kit/HANDOFF.md --json`, then fill it with the entity `id` in frontmatter (never a fresh ULID) — the CLI carries the full template so it no longer lives in this playbook. Machine frontmatter: `id`, `type: handoff`, `phase`, `lane`, `run_id`, `check_id`, `created`/`updated`. Human frontmatter: `session-date`, `branch`, `status`, `continuity-mode`, `active-phase`, `last-updated`. Body sections: Current State, What We're Building, Continuity Anchors, Progress This Session (Completed / In Progress / Not Started), Key Decisions, Blockers & Issues, Technical Context, Next Steps (mark one `→ START HERE`), Notes.

Rules:
- `id`/`type`/`phase`/`lane`/`run_id`/`check_id`/`created`/`updated` are the machine cross-link contract (used by `validate` and `resume`); `session-date`/`branch`/`status`/`continuity-mode`/`active-phase`/`last-updated` remain the human-narrative fields — keep both in sync (`phase` mirrors `active-phase`, `updated` mirrors `last-updated`) rather than merging them, since narrative and machine consumers read at different cadences
- `run_id`/`check_id` point at the latest RUN/CHECK closed by this handoff; `none` only if the phase had no run or check yet

## Command Reference

- `zharness --version` — version gate
- `zharness resume --json` — resolve `latest_run_id`/`latest_check_id` anchors
- `zharness handoff record --run-id {...} --check-id {...} --open-items '[...]' --json` — record the canonical handoff entity
- `zharness scaffold handoff --path .kit/HANDOFF.md --json` — emit the HANDOFF.md skeleton to fill

## Exit / Handoff Conditions

Complete only when: branch state and upstream status captured; continuity mode stated; active phase or `none` stated; latest work run path or `none` stated; latest check verdict or `none` stated; a single `→ START HERE` next action given; when the version gate passes, the handoff entity was recorded before the markdown was written and its `id`/`run_id`/`check_id` match.

## Error Handling

- `.kit/` missing: create the directory, write the handoff
- No git repo: document working directory only
- No commits: focus on working tree
- Sensitive data found: sanitize before writing, warn the user

## Anti-Patterns

- Writing "continue the work" as next step — too vague to resume cold; name the file, the function, the exact action
- Omitting blockers because they seem minor — next session wastes hours rediscovering what this session already knew
- Capturing what was done but not what was decided — decisions are the expensive part to reconstruct
- Overwriting HANDOFF.md without reading the previous one — loses prior session context that may still be relevant
