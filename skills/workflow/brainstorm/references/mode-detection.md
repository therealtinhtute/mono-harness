# Mode Detection

The merged `brainstorm` skill operates in 4 modes. Mode is detected from input shape, then **always confirmed via `AskUserQuestion`** before producing output. Detection is a hint, not a commitment — users may pivot mid-session.

## Detection rules

| Input shape | Detected mode | Confirm before |
|-------------|---------------|----------------|
| Vague trade-off question without lock intent (e.g. "REST or GraphQL?", "monolith or microservices?") | `explore` | Generating options |
| Raw idea, notes, or partial draft (e.g. "I want an AI inbox for small teams") | `lock-from-idea` | Writing IDEA.md or SPEC.md |
| `@file:` references to RFC / PRD / README / markdown | `lock-from-files` | Extracting and clarifying |
| Existing `.kit/planning/SPEC.md` is present and user mentions revising, updating, or refining | `refine` | Editing SPEC.md |

When multiple shapes co-occur (e.g. raw idea plus a file reference), default to `lock-from-files` and confirm.

## Ambiguous cases

Resolve via `AskUserQuestion` before producing output:

- User pastes a file but says "what do you think?" → could be `explore` (alternatives) or `lock-from-files` (extract). Ask which.
- User provides an idea but the working tree already contains `.kit/planning/SPEC.md` → could be `refine` or new `lock-from-idea`. Ask whether to revise or replace.
- Trade-off question that touches an existing locked spec → `explore` first, then offer `refine` if exploration changes scope.

## Mode upgrade rules

- `explore` → `lock-from-idea`: when user says "lock this," "make this a spec," or equivalent. Re-confirm scope, then transition.
- `lock-*` → `explore`: when user says "wait, let me think about alternatives." Surface options, then return to lock with the chosen path.
- `refine` → `lock-from-idea`: only if the existing SPEC.md will be discarded. Warn the user before overwriting.

Never silently change modes mid-session. Every transition surfaces via `AskUserQuestion`.

## Anti-patterns

- Producing `.kit/planning/SPEC.md` when the user only asked an exploratory question
- Generating a recommendation report when the user pasted a PRD and said "lock this"
- Refining an existing SPEC.md without showing the user what changed
- Writing both `.kit/planning/SPEC.md` and a `.kit/reports/brainstorm/` artifact in the same session unless the user explicitly asked for both
