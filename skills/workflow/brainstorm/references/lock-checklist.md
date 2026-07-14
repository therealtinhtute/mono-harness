# Lock Checklist

Run this self-review on `.kit/planning/SPEC.md` (and `.kit/planning/IDEA.md` if present) before showing the user. Fix issues inline; no need to re-review after fixing.

## 1. Placeholder scan
- No `TBD`, `TODO`, `details to be determined`, `similar to above`, or empty sections
- Every requirement is concrete and falsifiable
- Every acceptance criterion is checkable without re-asking the user

## 2. Internal consistency
- Goal and acceptance criteria align — does meeting acceptance prove the goal?
- In Scope and Out of Scope don't contradict each other
- Constraints don't conflict with requirements
- Architecture (if mentioned) matches feature descriptions

## 3. Scope check
- One coherent feature or module — not multiple independent subsystems
- If multiple subsystems detected, surface to user before locking; suggest decomposition into separate specs
- `Deferred Ideas` section captures anything pulled out of scope

## 4. Ambiguity check
- Every requirement has one interpretation, not two
- Every actor is named (no "the user" without specifying which user)
- Every external dependency is named, not implied

## 5. HARD-GATE compliance
- At least 1-2 alternatives are explicitly named
- Rejection reason is captured for each rejected alternative
- `Key Decisions` section documents trade-offs, not just outcomes

## User Review Gate

After self-review passes, show the user:

> "SPEC written to `.kit/planning/SPEC.md`. Please review and let me know if you want changes before I suggest `to-plan`."

Wait for explicit response. If changes requested, edit and re-run this checklist. Only suggest `to-plan` after the user approves.

## Anti-patterns

- Skipping placeholder scan because "the spec looks complete" — placeholders hide in formatted text
- Auto-suggesting `to-plan` without waiting for user response — defeats the gate
- Re-running this checklist after every minor edit — once per lock cycle is enough
