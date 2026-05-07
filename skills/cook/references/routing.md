# Routing — State Detection and Stop Messages

Run this table on every invocation, before any execution.

## Detection Table

| State | Files checked | Signal | Action |
|-------|---------------|--------|--------|
| `no-spec` | `.planning/SPEC.md` | missing or empty | Stop, route to `brainstorm` |
| `no-plan` | `.planning/ROADMAP.md` | missing | Stop, route to `plan` |
| `no-phase` | `.planning/phases/{slug}/{slug}-PLAN.md` | missing for selected phase | Stop, route to `plan phase {slug}` |
| `no-context` | `.planning/phases/{slug}/{slug}-CONTEXT.md` | missing | Stop, route to `plan phase {slug}` |
| `stale-plan` | Phase plan references files/symbols that no longer exist | Detected via grep during context load | Stop, route to `plan phase {slug}` to refresh |
| `placeholder-plan` | Phase plan contains `TBD`, `TODO`, `similar to`, "implement later" | Detected during context load | Stop, route to `plan phase {slug}` |
| `multiple-incomplete` | More than one phase has incomplete waves | Default `auto` mode is ambiguous | Ask user via `AskUserQuestion` which phase to run |
| `ready` | All required files present and concrete | — | Proceed to execution loop |

## Stop Message Templates

Use these verbatim so the user always sees the same shape:

### no-spec
```
🥷 No `.planning/SPEC.md` found. Lock the problem first.

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

Run: `/plan phase {slug}` to refresh, then re-invoke `/cook`.
```

### multiple-incomplete (interactive)
Use `AskUserQuestion` with the incomplete phase slugs as options. Mark the first incomplete phase by roadmap order as `(Recommended)`.

## Selecting the Active Phase (auto mode)

1. Parse `.planning/ROADMAP.md` for the ordered phase list.
2. For each phase, check whether its `-PLAN.md` shows all waves complete (look for a status section, completion markers, or — if absent — assume incomplete).
3. The first incomplete phase is the candidate.
4. If two phases are partially done (rare, indicates a previous handoff), trigger `multiple-incomplete` and ask.

## What `cook` Never Does Here

- Never invent missing artifacts to "unblock" itself
- Never edit SPEC.md, ROADMAP.md, or CONTEXT.md to skip a stop condition
- Never proceed past a `BLOCKED` status without user input
