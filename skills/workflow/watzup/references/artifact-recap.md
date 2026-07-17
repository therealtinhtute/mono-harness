# Artifact-Aware Recap

Use this when the repo follows the harness flow (a `zharness` binary is present and `resume --json` returns something other than `readiness: no-harness`).

## Source: `resume --json` only

`resume` is read-only over everything (`run`, `check`, `handoff`, `story` state — see `cli/docs/STATE.md`'s Writer/Reader Ownership table) and returns a single resolved snapshot:

```json
{"position": {"current_phase": "...", "status": "..."}, "latest_run_id": "ulid"|null, "latest_check_id": "ulid"|null, "latest_handoff_id": "ulid"|null, "drift": [{"type": "...", "detail": "...", "recovery": "..."}], "readiness": "clean"|"in-progress"|"drifted"|"no-harness"}
```

Do not separately read `.kit/workflow-state.yml`, `.kit/planning/ROADMAP.md`, phase `-CONTEXT.md`/`-PLAN.md`, `.kit/runs/work/*.md`, or `.kit/reports/check/*.md` to reconstruct phase/run/check state — `resume` already resolved all of them by ULID, including path renames the legacy yml pointer file couldn't survive. Re-deriving from those files independently risks disagreeing with the canonical snapshot `resume` returns.

`.kit/HANDOFF.md` is still read directly, but only for its narrative content (where the previous session left off, in prose) — its machine fields (`id`/`run_id`/`check_id`) should already agree with `resume`'s `latest_handoff_id`/`latest_run_id`/`latest_check_id`; a mismatch there is drift worth flagging even if `resume`'s own `drift` array didn't catch it (e.g., a `HANDOFF.md` written by an older, non-CLI-aware handoff run).

## What to Summarize

### Session Context
- `position.current_phase` / `position.status`
- Whether `latest_run_id` / `latest_check_id` / `latest_handoff_id` are populated (execution reached `work` / `check` / `handoff`)
- `readiness`, rendered verbatim — one of `clean | in-progress | drifted | no-harness`

### Drift → Recovery Mapping

`resume`'s `drift` array already carries a `recovery` string per entry — print it, don't paraphrase or re-derive it. For reference, the type → recovery pattern is (from `cli/docs/STATE.md`):

| Drift type | Recovery pattern |
|------------|-------------------|
| `missing_file` | `to-plan phase {slug}` to regenerate, or `git checkout` if a tracked file was deleted accidentally |
| `unknown_phase` | `to-plan phase {slug}` to create the missing story, or correct `current_phase` via `to-plan` if it was a typo |
| `out_of_order` | Re-run `check full` against the latest run — the stale check stays in place, superseded by the new one |

### Risk Mapping

| Pattern | Default severity | Example action |
|---------|-----------------|----------------|
| `drift: missing_file` / `unknown_phase` / `out_of_order` | vừa | Print that entry's `recovery` field |
| `readiness: no-harness` on a repo with existing `.kit/` artifacts | cao | `zharness import` |
| `HANDOFF.md` anchors disagree with `resume`'s latest IDs | vừa | Re-run `handoff` to resync the entity and the markdown |

## Output Integration

In the recap output:
- **Context** section: include `position.current_phase`, whether a run/check/handoff has been recorded
- **Risks** section: include drift-derived risks alongside git-derived risks
- **Readiness**: the `resume` value directly — no separate "artifact chain health" blend

## Skip Rule

If `zharness` is missing (fails the version gate) or `resume --json` isn't callable at all, that's a hard stop per the version gate, not a skip. If `resume` runs and returns `readiness: no-harness`, render that state — do not silently omit the harness section, since `no-harness` is itself the answer to "what's the harness state."
