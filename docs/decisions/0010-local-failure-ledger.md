# 0010 — A local failure ledger is maintainer-owned and optional for consumers

## Status

Accepted. 2026-09-16. Authority for creating `docs/evals/failures.md`, executed in phase
`p3-failure-ledger` of `harness-eval-loop.md` (then under `docs/plans/active/`) (R8).

## Context

Two of the six spine playbooks already read a ledger that has never existed.

`docs/playbooks/check.md:29` (step 4) says: *"IF `docs/evals/failures.md` exists → for every
failure class recorded two or more times, state whether the diff is clean of it; an absent file
is not an error."* `docs/playbooks/check.md:37` (step 10) says a `REQUEST_CHANGES` verdict
appends one row per finding *"IF `docs/evals/failures.md` exists"*. Both clauses have been
dormant since they were written. `docs/playbooks/work-full.md:28` supplies the vocabulary they
would use — `MISSING_CONTEXT|WRONG_TOOL|BAD_OUTPUT|REPEATED_LOOP|UNSAFE_ACTION|LOST_DECISION|UNKNOWN`
— and that taxonomy has no store either, so a class is named once in a `BLOCKED_*` Progress line
and never counted again.

The path is not new. `.claimignore:22` carries an exception for `docs/evals/failures.md`,
recorded as a *"historical audit measurement cited by `docs/audit/sdlc-gap-analysis.md`"* that
was deleted in `655c6ac` along with its audit and deliberately not restored — R12 of that
recovery restored the three audit documents only. `docs/decisions/0004-docs-directory-deletion-655c6ac.md`
is the record of that scoping call: *"Restore only what is required."* In 2026-08 nothing read
the file, so not restoring it was correct. The playbook clauses above are what changed; the file
now has live consumers.

What makes the question hard is `docs/decisions/0003-durable-memory-not-wired-into-playbooks.md`.
That ADR added a repo-scoped memory store and then deliberately declined to wire it into the
spine: *"No spine playbook gains a mandatory memory step."* The stated cost of wiring was
recurring — every session paying a read, every closure paying a write, on a store whose value at
the moment of reading is unproven. A failure ledger is the same shape of artifact and invites the
same mistake.

The measurement that makes it worth revisiting is in `docs/evals/runs.md` run 001. Fourteen
frozen routing cases produced two failures, and both are recurrences of classes already named
elsewhere in the repository rather than novel one-offs: H04 upgraded an explicit response-only
`check review` into a durable gate and wrote plan state, and H01 announced its resolved mode
after reading plans, against `docs/playbooks/work.md:13`. Neither has any guard. A class that
recurs and that no guard covers is precisely what `check.md` step 4's "two or more times" clause
was written to surface, and it cannot surface without somewhere to count.

## Decision

Create `docs/evals/failures.md` as a maintainer-owned, append-only markdown ledger in this
repository, and change no playbook to require it.

- **Ownership is local.** The ledger records failures observed in *this* repository's harness
  work. It is not a shared registry, it is not synced, and `zharness install` / `update` does not
  scaffold it into a consumer repository. A consumer that wants one creates the path itself.
- **Adoption stays optional, because the consumers are already conditional.** `check.md` steps 4
  and 10 are both written `IF … exists`, with "an absent file is not an error" stated inline. No
  clause is added, edited, or promoted to mandatory by this ADR, so a consumer that never creates
  the file sees behavior identical to today's. This is the narrow difference from ADR 0003: that
  decision declined to *add* mandatory playbook steps for a store with no consumers, while this
  one creates the store for optional steps that already exist and are dormant.
- **Classes come from the existing taxonomy only.** `docs/playbooks/work-full.md:28`'s seven
  tokens are the complete vocabulary (R7). A row that fits none of them is `UNKNOWN` with a
  rationale, never a new token — a private vocabulary would make the "recorded two or more times"
  count meaningless across rows.
- **Every row carries an immutable source.** A row cites a commit SHA plus a path and section, so
  it survives the cited file moving or being rewritten. A row whose only citation is an active
  plan path is not admissible: active plans move to `docs/plans/completed/` by design.
- **Coverage is stated, including its absence.** Each row records the guard or eval that would
  catch a recurrence, or the literal token `uncovered`. `uncovered` is the useful state — it is
  what tells a maintainer where a guard is missing — and it is never smoothed into a vague
  promise of future work.

**Rejected: a derived index or database.** `docs/decisions/0001-markdown-as-source-of-truth.md`
already settled that markdown is authoritative and any index is rebuildable, and v0.15 deleted
the SQLite store outright. A ledger that needs a binary to read would be unreadable by exactly
the agents `check.md` step 4 addresses, since the playbooks are written so that any agent that
can read a file and run git can execute the lifecycle.

**Rejected: making the ledger mandatory, or gating commits on it.** That is ADR 0003's recurring
cost with the names changed. A mandatory read costs every session; the value at the moment of
reading is still unmeasured. Phase `p3-failure-ledger` measures exactly that — R03 run three
times with the ledger present and three times without, in separate disposable fixtures, with both
sides required to pass the routing oracle (R9). Making it mandatory before that measurement would
be the assertion this repository's audits keep finding.

**Rejected: reusing `docs/memory/`.** The memory store from ADR 0003 is opt-in and unwired, and
its retrieval is keyword-ranked. Ledger rows are counted by class, not retrieved by relevance,
and `check.md` names a literal path. Overloading the memory store would couple a live playbook
clause to a surface that decision deliberately left unwired.

## Consequences

- The two dormant clauses in `docs/playbooks/check.md` become live in this repository the moment
  the file exists. A durable `gate` or `full` that returns `REQUEST_CHANGES` now has an append
  obligation it did not have before, and step 4 now has a file to read. Nothing enforces either:
  both are authoring discipline at the `Optional hook` level of
  `docs/patterns/encoding-invariants.md` unless a guard is later written for them.
- The ledger will drift toward stale if nobody prunes it, and pruning is a judgment call this ADR
  does not automate. An append-only file whose rows are never revisited is a cost with no
  benefit, and the "two or more times" clause makes a stale row actively misleading rather than
  merely inert.
- The `.claimignore:22` exception for `docs/evals/failures.md` becomes obsolete the moment the
  real file exists, and is removed in the same phase. Leaving it would mean the link verifier
  permanently excuses a path that now resolves — the exact class of silent-exception drift that
  `docs/decisions/0005-authored-documentation-boundary.md` warns about.
- Rollback is one deletion: `git rm docs/evals/failures.md`. Both consumers are conditional on
  the file's existence, so removing it returns the repository to today's behavior with no playbook
  edit and no migration. This is the property that makes accepting the ADR cheap, and it is the
  reason the decision is reversible in a way ADR 0003's wiring would not have been.
- Consumers of `zharness` see no change at all. No template, no embedded doc, and no installer
  path references the ledger.

## Authority

- `docs/playbooks/check.md:29` and `docs/playbooks/check.md:37` — the two conditional consumers
  that already name `docs/evals/failures.md`.
- `docs/playbooks/work-full.md:28` — the seven-token failure taxonomy this ledger stores.
- `docs/decisions/0003-durable-memory-not-wired-into-playbooks.md` — the recurring-cost argument
  against mandatory playbook wiring, which this decision honors rather than overturns.
- `docs/decisions/0004-docs-directory-deletion-655c6ac.md` — the "restore only what is required"
  scoping that left this path unrestored in `655c6ac`, and the reason it is created now rather
  than restored.
- `docs/decisions/0001-markdown-as-source-of-truth.md` — why the ledger is a markdown file and
  not an index.
- `.claimignore:22` — the obsolete exception removed alongside this decision.
- `docs/evals/runs.md` — run 001, the measured baseline whose two failures are recurrences of
  uncovered classes.
- `harness-eval-loop.md` (then under `docs/plans/active/`) — R7, R8, R9, and R10; the owner's 2026-09-16 approval
  of the corrected draft is the authority for the phase, and this ADR is the authority for the
  file.
