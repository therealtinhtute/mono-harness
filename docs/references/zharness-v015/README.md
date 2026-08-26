# zharness v0.15 "slim" — reference material

Supporting material for `docs/plans/active/zharness-v015-slim.md`. Everything here is an
**archive**: a record of how that plan reached its current shape. None of it is a live plan,
and none of its line-number citations track the repository as it changes.

The live plan is the only executable document. If this material and the plan disagree, the
plan wins.

## What is here

| File | What it is |
|---|---|
| `docs/references/zharness-v015/v013-plan.md` | The v0.13 "slim" plan, the superseded predecessor. Authored and superseded on 2026-08-26. It lived under `.kit/`, which `.gitignore` excludes, so it reached no clone and no CI run — committing it here is what lets the merged plan cite a real path instead of absorbing its content inline. |
| `docs/references/zharness-v015/v015-original-plan.md` | The v0.15 plan as it stood before the review: 7 success signals, 10 requirements, 7 non-goals, `phases: none`. Kept because the review cites it by line number. |
| `docs/references/zharness-v015/review-findings.md` | The read-only review of both plans: 15 findings — 4 blockers, 6 majors, 5 minors — with evidence at `file:line`, a v0.13→v0.15 kept/changed/dropped table, and the recommendation. |
| `docs/references/zharness-v015/interview-spec.md` | The `/interview` pass that closed all 4 blockers and the 6 majors: 7 questions with the options offered and the one chosen, plus the spec accepted at the end. |
| `docs/references/zharness-v015/research-links.md` | The external sources the design was argued from, with per-row verification status, plus the in-repository evidence paths. |

## How the plan got here

```
v0.13 plan  ──superseded──┐
                          ├──► review (15 findings, 4 blockers)
v0.15 draft ──reviewed────┘         │
                                    ▼
                        /interview — 7 decisions
                                    │
                                    ▼
                 merged plan ──► to-plan ──► 5 phases, P0–P4
```

The four blockers were, in short: the original S7 targeted a cost reduction that shipped work
had already banked; its measurement method depended on commands the plan itself deletes; NG1
declared the six spine `SKILL.md` files out of the product path when they are exactly what
`npx skills add` installs; and R1's kill list was derived from a stale contract rather than
from `root.go`, so it missed two live commands and listed one that was never registered.

## Reading these files safely

Every path and line number in the archived documents is **as of 2026-08-26**. Several name
commands that v0.15 deletes, and at least two name files that do not exist yet
(`docs/PROJECT.md`, `docs/memory/`) because the plan's final phase creates them. They are
historical measurements, not live cross-references — `.claimignore` records the ones the
doc-link gate would otherwise flag.
