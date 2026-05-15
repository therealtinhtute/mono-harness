# Autoplan Patterns

Use this file when the wrapper skill needs a little more structure without bloating `SKILL.md`.

## 1) Input modes

### Idea-first

Use when the user brings:
- a vague idea
- a desired outcome without a path
- scattered bullets or chat fragments
- a problem statement but no real spec

Primary job:
- shape the problem
- clarify success
- surface constraints
- turn ambiguity into a spec that can actually be planned

### Spec-first

Use when the user brings:
- a markdown spec
- PRD-like notes
- a structured request with headings
- an existing plan that still feels incomplete

Primary job:
- audit completeness
- detect weak assumptions
- fill critical gaps
- convert the spec into executable phases and tasks

## 2) Minimum-question strategy

Ask only questions that change planning quality materially.

Good question buckets:
- **objective**: what outcome matters most?
- **scope**: what is in, what is out?
- **constraints**: time, budget, stack, platform, team, policy
- **success**: how will we know this worked?
- **dependencies**: APIs, repos, data, approvals, environments

Avoid questions that are:
- cosmetic
- reversible later
- already implied by the user's context
- better answered by quick inspection or lightweight research

If many questions exist, batch them into one concise list.

## 3) Research threshold

Do live research when the answer could have changed or when getting it wrong would distort the plan.

Typical triggers:
- current library/framework capabilities
- latest pricing or provider limits
- active competitors or recent launches
- changing operational rules, APIs, or platform constraints

Do **not** over-research when:
- the issue is internal framing, not external facts
- the facts will not materially change planning
- the user asked for a very lightweight pass

## 4) Spec normalization pattern

Normalize messy input into this shape:

```text
Goal
Context
Scope
Non-goals
Constraints
Assumptions
Open questions
Success criteria
```

Notes:
- merge duplicates
- remove decorative wording
- separate user facts from your inference
- call out anything that is still provisional

## 5) Phase and task decomposition

Turn the normalized spec into:
- **phases** = meaningful chunks of work
- **goals** = what each phase must achieve
- **tasks** = concrete actions with visible outputs

Good tasks are:
- specific
- verifiable
- small enough to execute cleanly
- grouped under the right phase

Weak tasks sound like:
- "improve system"
- "work on UX"
- "handle backend"

Stronger tasks sound like:
- "audit current API endpoints and list missing auth checks"
- "draft migration steps for schema changes and rollback"
- "build smoke test list for the deployment path"

## 6) Lite vs full mode

### Autoplan-lite
- fewer questions
- lighter research
- coarser phases
- fewer but still concrete tasks
- strong emphasis on the next recommended action

### Full autoplan
- deeper gap analysis
- more complete spec normalization
- stronger risk/dependency callouts
- more detailed phase and task breakdown

## 7) Recommended closing move

End with one recommended next move:
- **execute now** if the plan is small and ready
- **switch to `goal`** if execution is substantial and should be supervised
- **brief `problem-solving` detour** if the plan is blocked by framing or architecture confusion

The wrapper is successful when the user can immediately decide what to do next without rereading the whole conversation.
