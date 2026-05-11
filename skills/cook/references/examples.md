# Cook Examples

Three worked scenarios covering the most common entry states.

---

## Example 1 — Missing spec (fail-fast routing)

**Input**: "cook this feature, build the inbox triage logic"

**State detection**: `.kit/planning/SPEC.md` not found.

**Cook's response**:
```
🥷 No `.kit/planning/SPEC.md` found. Lock the problem first.

Run: `/brainstorm` with your idea, notes, or @file: refs.
```

**No execution attempted.** Cook does not write a placeholder spec, does not guess scope, does not start coding.

---

## Example 2 — Ready to execute (auto mode, two waves)

**Input**: `/cook` (no args, default `auto`)

**State detection**:
- `.kit/planning/SPEC.md` ✅
- `.kit/planning/ROADMAP.md` ✅ (lists 3 phases: `auth-foundation`, `inbox-ui`, `triage-rules`)
- Phase 1 `auth-foundation` artifacts ✅, all waves complete (verified by checking `-PLAN.md` status section)
- Phase 2 `inbox-ui` artifacts ✅, no waves complete → **selected**

**Cook's response (condensed)**:
```
🥷 Phase `inbox-ui` ready. Run artifact: `.kit/runs/cook/20240901-1010-inbox-ui.md`.

Preflight:
- required artifacts present
- working tree clean
- scope drift: no

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
🥷 BLOCKED_CONTRACT_DRIFT on `triage-rules` Wave 1 Task 2.

Conflict between SPEC and PLAN:
- SPEC §4.2 expects score by sender + keyword + recency
- PLAN task 2 only covers sender + keyword
- No recency field on `Message` type to derive from

Run artifact updated: `.kit/runs/cook/20240901-1445-triage-rules.md`

This is a scope/spec gap, not an implementation issue. Two options:

1. `/brainstorm refine` — clarify §4.2 (drop recency or define it concretely)
2. `/plan phase triage-rules` — if SPEC is fine, refresh the plan to cover all three dimensions

Stopping the wave. Phase `triage-rules` is not clean.
```

**No further execution.** Cook does not invent a recency field, does not silently drop the dimension, does not mark the task `DONE_WITH_CONCERNS`. The blocker is surfaced and routed back to upstream.

---

---

## Example 4 — Simple mode from a direct prompt

**Input**: `/cook simple` with prompt "Add a `--dry-run` flag to the CLI's `deploy` command that prints the deployment plan without executing it. See `src/cli/deploy.ts`."

**Mode resolution**: explicit `simple` argument.

**Step 1 — Prompt quality check**: file provided (`src/cli/deploy.ts`), change described (add `--dry-run` flag), success criterion implied (prints plan without executing). Proceed.

**Step 2 — Quick research**:
- Read `src/cli/deploy.ts` (1 file)
- Grep for `program.command('deploy')` to confirm entry point (1 grep)
- Read `src/cli/index.ts` to see how other flags are registered (1 file)
Total: 3 reads. Within limit.

**Step 3 — State approach**:
```
🥷 Adding `--dry-run` to `deploy` in `src/cli/deploy.ts:42`. Flag gates the execution block behind an `if (!dryRun)` check and prints the resolved config via `console.log`. No new dependencies needed.
```

**Step 4 — Scope guard**: 2 files, ~15 lines. Within threshold. Proceed.

**Step 5 — Execute**: Edited `src/cli/deploy.ts` inline. One targeted change to the command registration + one guard around the execution call.

**Step 6 — Light verify**:
```
pnpm test src/cli/deploy.test.ts → 8 passed, 0 failed
```

**Cook's wrap**:
```
🥷 Done. `--dry-run` added to deploy command.

Changed: src/cli/deploy.ts (+18 -2)
Verified: pnpm test src/cli/deploy.test.ts → 8 passed

Next:
- `/git cm` to commit
- `/check review` if you want a security pass before merging
```

---

## Example 5 — Simple mode from a brainstorm explore file

**Input**: `/cook simple @.kit/reports/brainstorm/20240901-api-error-format.md`

**Mode resolution**: explicit `simple` argument with @file ref.

**Step 1 — Prompt quality check**: brainstorm file is the source. Read it to extract: what changes, where, success criterion.

Brainstorm file summary (from read):
- Change: standardize API error responses to `{ error: { code, message, details } }` shape
- Where: `packages/api/src/middleware/error-handler.ts`
- Success: all error responses match the schema; existing tests pass

Prompt quality sufficient. Proceed.

**Step 2 — Quick research**:
- Read `packages/api/src/middleware/error-handler.ts`
- Read 1 existing test file `packages/api/src/middleware/error-handler.test.ts`
- Grep for `res.status(` to confirm no other error-response sites in scope
Total: 3 reads. Within limit.

**Step 3 — State approach**:
```
🥷 Updating error-handler middleware at `error-handler.ts:28` to emit `{ error: { code, message, details } }`. All three catch branches get the new shape. Existing tests need shape update only — logic unchanged.
```

**Step 4 — Scope guard**: 2 files, ~40 lines. Within threshold. Proceed.

**Step 5 — Execute**: Edited `error-handler.ts` (middleware shape) and `error-handler.test.ts` (assertion shape). No new files.

**Step 6 — Light verify**:
```
pnpm test packages/api/src/middleware → 12 passed, 0 failed
```

**Cook's wrap**:
```
🥷 Done. Error response shape standardized per brainstorm output.

Changed: packages/api/src/middleware/error-handler.ts (+22 -18), error-handler.test.ts (+9 -9)
Verified: 12 passed

Next:
- `/git cm` to commit
- `/check review` if integration tests are downstream of this middleware
```

---

## Notes

- `auto` mode resolves to `full` or `simple` based on available artifacts. `full` walks the roadmap forward. `simple` executes from prompt or brainstorm explore file.
- `phase <slug>` is an alias for `full phase <slug>` — backward-compatible.
- Cook never modifies `.kit/planning/SPEC.md` or `.kit/planning/ROADMAP.md` — those are owned by `brainstorm` and `plan` respectively.
- Simple mode never writes to `.kit/planning/`; it still writes a run artifact under `.kit/runs/cook/` when execution starts.
- The three statuses that pause full-mode execution (`NEEDS_CONTEXT`, `BLOCKED`, non-clean phase gate) are the only stop conditions cook respects on its own in full mode. In simple mode, the scope guard is the only hard stop.
- Preferred blocker taxonomy: `BLOCKED_CONTEXT`, `BLOCKED_SCOPE`, `BLOCKED_VERIFICATION`, `BLOCKED_CONTRACT_DRIFT`.
