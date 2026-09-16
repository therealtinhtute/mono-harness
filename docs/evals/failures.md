# Failure Ledger

Observed harness failures in this repository, one row per incident, append-only.

Authority: `docs/decisions/0010-local-failure-ledger.md`. This file is maintainer-owned and local.
It is not synced, not scaffolded by `zharness install` or `update`, and not required by any
playbook. Its two consumers are conditional on its existence: `docs/playbooks/check.md:29`
(step 4) reads it and, for every class recorded two or more times, states whether the diff under
review is clean of that class; `docs/playbooks/check.md:37` (step 10) appends one row per finding
when a durable `gate` or `full` returns `REQUEST_CHANGES`. Deleting this file returns the
repository to the behavior it had before the file existed.

## Conventions

- **class** — one of the seven tokens in `docs/playbooks/work-full.md:28`:
  `MISSING_CONTEXT`, `WRONG_TOOL`, `BAD_OUTPUT`, `REPEATED_LOOP`, `UNSAFE_ACTION`,
  `LOST_DECISION`, `UNKNOWN`. No other token is admissible. A row that fits none of them is
  `UNKNOWN` with a stated rationale, never a new token — a private vocabulary would make the
  "recorded two or more times" count meaningless across rows.
- **source** — an immutable commit SHA plus a path and section inside that commit. A row is not
  admissible if its only citation is a path under `docs/plans/active/`: active plans move to
  `docs/plans/completed/` by design, so such a citation rots. Where the incident's own record
  has since moved, the SHA below is one at which the cited path still resolves.
- **coverage** — the guard, contract test, or eval case that would catch a recurrence, or the
  literal token `uncovered`. `uncovered` is the useful state; it marks where a guard is missing
  and is never softened into a promise of future work.
- Rows are appended, never rewritten. A row whose finding is later covered gains a follow-up row
  rather than an edited `coverage` field.

## Incidents

### 2026-09-15 — routing resolved a durable gate with two active plans

- class: `MISSING_CONTEXT`
- rationale: the agent had both plan files available and read them; what it lacked was an
  instruction requiring the *total* active-plan count before matching by initiative name. It
  filtered to the one plan that matched the request and proceeded. The deficit was in the context
  the playbook supplied, not in the tools used or the output produced, so `MISSING_CONTEXT` fits
  and `BAD_OUTPUT` does not — the output was a faithful execution of an under-specified rule.
- surface: `docs/playbooks/check.md`, the `auto` preflight.
- source: `90f5b84`, `docs/audit/deepseek-harness-token-audit.md` §4 S1, routing smoke
  observations table, row "Two active plans, one matching initiative" — *"Initial wording failed:
  agent filtered by initiative and proceeded."*
- coverage: covered. `docs/playbooks/check.md:15` (step 2) now requires exactly one active plan in
  total and to *"list every candidate and stop even when only one matches the request"*; case R07
  in `docs/evals/routing.md` is the standing regression, and the six-trial historical control in
  `docs/evals/runs.md` run 001 measures the fix directly — old side `0dfb5b2^` 0/3, new side
  `0dfb5b2` 3/3, conclusion `reproduced`.

### 2026-09-07 — uninstall destroyed consumer prose because ownership was inferred, not recorded

- class: `LOST_DECISION`
- rationale: ownership of a managed path is decided at install time and was not written down, so
  every later operation re-derived it from the file's current contents. An installer-created
  `AGENTS.md` that the consumer had since appended prose to was re-classified as wholly
  installer-owned and removed. The information needed existed at one moment and was not
  persisted — that is `LOST_DECISION`, not `UNSAFE_ACTION`: the destructive step was authorized
  by the state the code could see.
- surface: `cli/internal/installer/` — uninstall and the `.gitignore` / `docs/WORKFLOW.md` paths
  that shared the root cause (F01, F06, F07).
- source: `b4c1ef0`, `docs/audit/2026-09-07-integrity-review.md` §F01 — *"Consumer prose can be
  deleted with an installer-created AGENTS.md"*; remediation success signals at `2013158`,
  `audit-integrity-remediation.md` (then under `docs/plans/active/`) S1/S2/S3.
- coverage: covered. `docs/decisions/0008-recorded-ownership-and-transactional-recovery.md` is the
  authority; ownership is recorded once in `.zharness/base/ownership.tsv` and never re-derived,
  with `unknown` meaning keep. Contract tests S1, S2 and S3 of
  `docs/plans/completed/audit-integrity-remediation.md` are the standing guard, re-run clean at
  `2013158`.

### 2026-09-07 — incomplete stash recovery reported success and removed the remaining evidence

- class: `BAD_OUTPUT`
- rationale: `update --abort` with one stash payload missing returned success and then deleted
  `.zharness/update-stash/`, so the operation both misreported its result and destroyed what
  would have shown otherwise. The distinguishing feature from the row above is that here the code
  had the information — it knew the inventory was short — and emitted the wrong result anyway.
- surface: `cli/internal/installer/`, stash capture and restore.
- source: `b4c1ef0`, `docs/audit/2026-09-07-integrity-review.md` §F05 — *"Incomplete stash
  recovery returns success and removes remaining evidence"*; remediation success signal at
  `2013158`, `audit-integrity-remediation.md` (then under `docs/plans/active/`) S5.
- coverage: covered. Stash format v2 plus inventory validation; S5 requires `update --abort` with
  one payload removed to return non-zero, name the unrestored path, and leave
  `.zharness/update-stash/` on disk. Re-run clean at `2013158`.

### 2026-09-16 — an installed pre-commit wrapper can go stale against its own template

- class: `UNKNOWN`
- rationale: this is a mechanism gap, not an agent action, so none of the six behavioral tokens
  applies and `UNKNOWN` is recorded with its reason rather than forced into a closer-sounding
  one. The guard *logic* is safe: the installed `.git/hooks/pre-commit` re-extracts the
  `ZGUARD-CORE` block from `scripts/install-git-hooks.sh` on every run (or from the staged blob
  when the installer itself is staged), so a logic edit is live without reinstalling. What is
  baked in at install time is the wrapper's call sequence around that block. Editing the wrapper
  half of the template changes nothing until someone re-runs
  `bash scripts/install-git-hooks.sh --force`, and nothing tells them to.
- surface: `.git/hooks/pre-commit` (untracked, per-clone) against the heredoc template in
  `scripts/install-git-hooks.sh`.
- source: `2013158`, `audit-integrity-remediation.md` (then under `docs/plans/active/`), Decisions entry
  `[2026-09-16 (closure)] (p3-revision-selector)` and the p3 Validation `proof_gaps` line, which
  record the verification and state the residue: *"What can still go stale is the wrapper's call
  sequence, which is baked in at install time. No guard or doc encodes 'reinstall hooks after
  editing the hook template'."*
- coverage: `uncovered`. `scripts/test-guards.sh` exercises the extracted `ZGUARD-CORE` block, not
  the installed wrapper around it, and no guard or document requires reinstalling after a template
  edit. Closing this would touch `scripts/**`, which both the originating initiative and
  this initiative's R10 place under `surfaces_avoided`, so it is recorded
  here rather than fixed.

## Candidates not yet seeded

Observed and recorded elsewhere; not rows until a maintainer decides they belong here. Listed so
the omission is deliberate and visible rather than silent.

- `stashRestore` writes to `filepath.Join(root, e.rel)` with `e.rel` read back from
  `.zharness/update-stash/stash.tsv`, so a `../` component in that file escapes the repository
  root on `zharness update --abort`. Carried in this initiative's own Current State; out of scope under R10.
- Run 001's two baseline failures — H04 upgrading an explicit `check review` into a durable gate
  and writing plan state, and H01 announcing its resolved mode after reading plans against
  `docs/playbooks/work.md:13`. Both are recorded with full evidence in `docs/evals/runs.md` and
  under `docs/evals/evidence/run-001/`. They are not appended here because the gate that found
  them returned `APPROVE_WITH_REQUESTS`, and `docs/playbooks/check.md:37` conditions the append
  on `REQUEST_CHANGES`; seeding them by hand would make this file's provenance ambiguous between
  the playbook's rule and the author's judgment.
