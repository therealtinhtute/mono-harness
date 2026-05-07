# Cook Examples

Three worked scenarios covering the most common entry states.

---

## Example 1 — Missing spec (fail-fast routing)

**Input**: "cook this feature, build the inbox triage logic"

**State detection**: `.planning/SPEC.md` not found.

**Cook's response**:
```
🥷 No `.planning/SPEC.md` found. Lock the problem first.

Run: `/brainstorm` with your idea, notes, or @file: refs.
```

**No execution attempted.** Cook does not write a placeholder spec, does not guess scope, does not start coding.

---

## Example 2 — Ready to execute (auto mode, two waves)

**Input**: `/cook` (no args, default `auto`)

**State detection**:
- `.planning/SPEC.md` ✅
- `.planning/ROADMAP.md` ✅ (lists 3 phases: `auth-foundation`, `inbox-ui`, `triage-rules`)
- Phase 1 `auth-foundation` artifacts ✅, all waves complete (verified by checking `-PLAN.md` status section)
- Phase 2 `inbox-ui` artifacts ✅, no waves complete → **selected**

**Cook's response (condensed)**:
```
🥷 Phase `inbox-ui` ready. 2 waves, 6 tasks.

Wave 1 (parallel-safe, 3 tasks):
- T1: scaffold `apps/web/inbox/page.tsx`
- T2: scaffold `apps/web/inbox/list.tsx`
- T3: add inbox route to nav config

Wave 2 (sequential, 3 tasks):
- T4: wire `useInboxQuery` hook to tRPC
- T5: empty state + loading skeleton
- T6: integration test

Starting Wave 1.
```

**Execution**:
- Wave 1: dispatched 3 inline edits in parallel (small file scaffolds). Each task ran its verification command (`pnpm typecheck` for T1/T2, `pnpm dev:check-routes` for T3). All `DONE`.
- Wave 2: T4 dispatched as a subagent (touched 4 files: hook + tRPC router + types + barrel export). T5 and T6 inline. Verification per task: `pnpm test inbox.spec.ts`. T6 returned `DONE_WITH_CONCERNS` (one flaky timing assertion). Cook surfaced the concern; user said "ack, continue".
- Phase gate: invoked `/check full` against the phase diff. Clean.

**Cook's wrap**:
```
🥷 Phase `inbox-ui` clean. Diff: 11 files (+412 -38). Check passed.

Next:
- `/git cm` to checkpoint this phase
- `/cook` again to continue with phase 3 (`triage-rules`)
- `/handoff` if you're stopping for the day
```

---

## Example 3 — Mid-flight blocker (sub-agent reports BLOCKED)

**Input**: `/cook phase triage-rules`

**State detection**: ready.

**Execution**:
- Wave 1, Task 2: subagent dispatched to "implement priority-scoring rule engine in `packages/triage/scorer.ts`". Subagent returns:
  ```
  Status: BLOCKED
  Reason: SPEC §4.2 says "score by sender + keyword + recency". Plan task 2 says "score by sender + keyword". Recency dimension has no acceptance criterion in either artifact and no data field on the message type to derive it from.
  ```

**Cook's response**:
```
🥷 BLOCKED on `triage-rules` Wave 1 Task 2.

Conflict between SPEC and PLAN:
- SPEC §4.2 expects score by sender + keyword + recency
- PLAN task 2 only covers sender + keyword
- No recency field on `Message` type to derive from

This is a scope/spec gap, not an implementation issue. Two options:

1. `/brainstorm refine` — clarify §4.2 (drop recency or define it concretely)
2. `/plan phase triage-rules` — if SPEC is fine, refresh the plan to cover all three dimensions

Stopping the wave. Phase `triage-rules` is not clean.
```

**No further execution.** Cook does not invent a recency field, does not silently drop the dimension, does not mark the task `DONE_WITH_CONCERNS`. The blocker is surfaced and routed back to upstream.

---

## Notes

- `auto` mode walks the roadmap forward; `phase <slug>` runs one named phase (useful for re-runs after a fix).
- Cook never modifies `.planning/SPEC.md` or `.planning/ROADMAP.md` — those are owned by `brainstorm` and `plan` respectively.
- The three statuses that pause execution (`NEEDS_CONTEXT`, `BLOCKED`, non-clean phase gate) are the only stop conditions cook respects on its own. Everything else continues.
