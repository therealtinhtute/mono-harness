# Routing — State Detection and Stop Messages

Run this table on every invocation, before any execution.

## Step 0: Mode Resolution

Resolve mode before checking artifact state.

| Argument | Resolved mode |
|----------|---------------|
| `simple` or `simple @file` | `simple` |
| `full`, `full phase <slug>`, `phase <slug>` | `full` |
| No argument | auto-detect (see table below) |

**Auto-detect (no argument)**:

| Observed state | Resolved mode |
|----------------|---------------|
| `.kit/planning/SPEC.md` + `ROADMAP.md` + target phase artifacts all present | `full` |
| `.kit/reports/brainstorm/*.md` present or @ref'd, no `.kit/planning/SPEC.md` | `simple` |
| Only a direct prompt, no planning or brainstorm artifacts | `simple` (after prompt-quality check) |
| No argument, no artifacts, no meaningful prompt | Stop → `/brainstorm` |
| Ambiguous: stale SPEC + new prompt, or SPEC exists but brainstorm file also ref'd | `AskUserQuestion` to clarify |

Once mode is resolved, proceed to the matching section below.

---

## Full Mode Detection Table

| State | Files checked | Signal | Action |
|-------|---------------|--------|--------|
| `no-spec` | `.kit/planning/SPEC.md` | missing or empty | Stop, route to `brainstorm` |
| `no-plan` | `.kit/planning/ROADMAP.md` | missing | Stop, route to `plan` |
| `no-phase` | `.kit/planning/phases/{slug}/{slug}-PLAN.md` | missing for selected phase | Stop, route to `plan phase {slug}` |
| `no-context` | `.kit/planning/phases/{slug}/{slug}-CONTEXT.md` | missing | Stop, route to `plan phase {slug}` |
| `stale-plan` | Phase plan references files/symbols that no longer exist | Detected via grep during context load | Stop, route to `plan phase {slug}` to refresh |
| `placeholder-plan` | Phase plan contains `TBD`, `TODO`, `similar to`, "implement later" | Detected during context load | Stop, route to `plan phase {slug}` |
| `contract-drift` | Working tree or requested scope already touches files outside `Allowed Surfaces` / task `touches`, or conflicts with `Forbidden Surfaces` / task `avoid` | Detected during preflight | Stop, route to `plan phase {slug}` or `brainstorm refine` |
| `multiple-incomplete` | More than one phase has incomplete waves | Default `auto` mode is ambiguous | Ask user via `AskUserQuestion` which phase to run |
| `ready` | All required files present and concrete | — | Proceed to execution loop |

## Stop Message Templates

Use these verbatim so the user always sees the same shape:

### no-spec
```
🥷 No `.kit/planning/SPEC.md` found. Lock the problem first.

Run: `/brainstorm` with your idea, notes, or @file: refs.
```

### no-plan
```
🥷 SPEC exists, no plan. Generate the roadmap first.

Run: `/plan full`.
```

### no-phase / no-context
```
🥷 Phase artifacts missing for `{slug}`. Refresh the phase plan.

Run: `/plan phase {slug}`.
```

### stale-plan / placeholder-plan
```
🥷 Phase plan for `{slug}` is {stale|incomplete}. Specifics:
- {file or symbol that no longer exists}
- {placeholder text and line}

Run: `/plan phase {slug}` to refresh, then re-invoke `/work`.
```

### contract-drift
```
🥷 Contract drift detected before execution for `{slug}`.

Conflict:
- {working-tree file or requested scope item}
- outside allowed surfaces or inside forbidden scope

Run one:
- `/plan phase {slug}` if the phase contract should be refreshed
- `/brainstorm refine` if the spec boundary itself changed
```

### multiple-incomplete (interactive)
Use `AskUserQuestion` with the incomplete phase slugs as options. Mark the first incomplete phase by roadmap order as `(Recommended)`.

## Selecting the Active Phase (auto mode)

1. Parse `.kit/planning/ROADMAP.md` for the ordered phase list.
2. For each phase, check whether its `-PLAN.md` shows all waves complete (look for a status section, completion markers, or — if absent — assume incomplete).
3. The first incomplete phase is the candidate.
4. If two phases are partially done (rare, indicates a previous handoff), trigger `multiple-incomplete` and ask.

## Simple Mode Stop Messages

### scope-guard
```
🥷 Scope guard triggered in `simple` mode.

Research found: {files_count} files / {lines_count} lines / unknown subsystem: {subsystem}.

This exceeds the simple mode limit (≤5 files, ≤100 lines, no unknown subsystem).
Upgrade to full pipeline:
1. `/brainstorm` — lock the spec
2. `/plan full` — generate phase artifacts
3. `/work full` — execute with verification gates
```

### prompt-too-thin (before AskUserQuestion)
```
🥷 Prompt is too vague to execute safely.
```
(Follow with AskUserQuestion targeting: scope, files affected, success criterion.)

### prompt-still-thin (after AskUserQuestion, still insufficient)
```
🥷 Still missing enough context to proceed.

Run: `/prompt-leverage` to strengthen the prompt, then re-invoke `/work simple`.
```

---

## What `work` Never Does Here

- Never invent missing artifacts to "unblock" itself
- Never edit SPEC.md or ROADMAP.md to skip a stop condition
- Never use a run artifact as a substitute for the planning contract
- Never proceed past a `BLOCKED` status without user input
- Never bypass the scope guard in simple mode — not even when the user says "just do it"
