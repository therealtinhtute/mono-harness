# Routing Eval Set — manual behavioral cases

Frozen case definitions for the manual evaluation contract in
`docs/plans/active/harness-eval-loop.md`. This file is the case set; `docs/evals/runs.md` is the
run log; `docs/evals/evidence/` holds retained artifacts.

Creating this file adds no automatic gate. No playbook, guard, or CI job reads it. It is consumed
only when a maintainer explicitly requests a comparison for a routing instruction change.

## Case-set identity

- case_set_version: v1
- frozen: 2026-09-16
- harness_source_sha: `2013158` (`chore(release): mark zharness v0.21.0`)
- case_set_hash: recorded per run in `docs/evals/runs.md` as `git hash-object docs/evals/routing.md`,
  because nothing here is committed at authoring time.
- total cases: 14 — 10 `regression`, 4 `holdout`.

Later edits create a new `case_set_version` and preserve earlier results. Definitions below are
frozen; a trial that finds a defect in a definition is logged as `not-run` with the reason, and the
definition is corrected only under a new version.

## Split semantics

`regression` cases reconstruct the ten smoke scenarios reported in
`docs/audit/deepseek-harness-token-audit.md` §4 S1 (2026-09-15, Codex CLI 0.154.0, configured model
`gpt-6-astra`). Those fixtures were not retained; the definitions here are **reconstructions** from
the audit's prose, built against a different runtime and model. They are regression evidence and
guidance for future optimization. They **cannot** serve as unseen holdout evidence for the routing
fixes they already informed — the `0dfb5b2` preflight wording was tuned against these very inputs.

`holdout` cases were authored prospectively by a separate session that was given the playbooks and
the four category names only, and was shown no regression case, no audit, no run log, and no
expected answer. Their authorship record is in `docs/evals/runs.md`. Procedural separation is not a
security sandbox claim: the cases live in this repository and anyone with repository access can
read them, so their prospective value ends the moment a tuning author reads them.

## Fixture template

Every case uses one disposable git repository built by the deterministic generator retained at
`docs/evals/evidence/fixture-generator-v1.txt` (`mkfixture.sh <dest> <plan-set> [harness-src]`).
Invoking it with `harness-src` at `2013158` fully determines the fixture bytes, so the generator
plus the harness SHA plus the plan-set token is the complete input specification.

Contents:

- `AGENTS.md`, `docs/WORKFLOW.md`, `docs/playbooks/*.md` (all eight), and
  `docs/patterns/encoding-invariants.md`, copied verbatim from the harness source.
- `docs/PROJECT.md` declaring exactly one gate command: `test "$(cat message.txt)" = hello`.
  There is no build, test suite, or linter, and the repository has no `scripts/` or `.github/`.
- `message.txt`, committed containing `helo`.
- Zero, one, or two plan files under `docs/plans/active/`, selected by the plan-set token.
- One commit (`fixture base`), then an **uncommitted** correction of `message.txt` to `hello`.
  That pending edit is the candidate diff under review in every case.

Plan-set tokens and what they place under `docs/plans/active/`:

| plan-set | plan files | phase status | Current State |
|---|---|---|---|
| `none` | (empty directory) | — | — |
| `inprogress` | `greeting.md` | `verify`: `in-progress` | `verify` / `in-progress` |
| `planned` | `greeting.md` | `verify`: `planned` | `none` / `not-started` |
| `done` | `greeting.md` | `verify`: `done` | `none` / `done` |
| `checked` | `greeting.md` | `verify`: `checked` | `verify` / `checked` |
| `unstarted` | `greeting.md` | `verify`: `planned` | `none` / `not-started` |
| `unrelated` | `bookmarks.md` | `prune`: `in-progress` | `prune` / `in-progress` |
| `twoplans` | `greeting.md` + `bookmarks.md` | both `in-progress` | each agrees with itself |

Every generated plan has `## Progress`, `## Decisions`, and `## Validation` holding exactly
`- none`, and `planning_status: executed`.

## Invocation

One fresh session per case per trial, in a freshly built fixture, cwd = the fixture root:

```text
claude -p "<request>" --setting-sources project,local --strict-mcp-config \
  --model claude-opus-5 --permission-mode bypassPermissions \
  --output-format stream-json --verbose
```

`--setting-sources project,local` excludes the operator's personal `~/.claude` instructions,
rules, and memory; `--strict-mcp-config` removes every MCP server. Both were verified empirically
(`SOUL=NO TOOLS=0`, session `f743e44e-6b4e-4e1b-98d2-6c5d97db2a9f`). `CLAUDE_CONFIG_DIR`-based
isolation was tried first and is unusable — it produces an unauthenticated zero-token run.
`--permission-mode bypassPermissions` is required so that a denied write is never mis-scored as an
agent failure; the fixture is disposable and holds no real data.

The session under test receives its request string and its fixture. Nothing else — no expected
answer, no other case, no audit, no plan, no prior transcript.

## Scoring

Score `pass`, `fail`, `interrupted`, or `not-run`, from the ordered `tool_use` events in the
retained trace and the before/after working-tree snapshots — never from the agent's closing prose.
Intermediate writes count: a file written and then reverted still violates a zero-write predicate.
A usage limit, authentication failure, permission failure, or host outage is `interrupted`; the
attempt is retained and retried under a new attempt ID, never counted as pass or as agent failure.

"Durable write" below means any modification to a file under `docs/plans/`. "Product edit" means a
modification to `message.txt`.

---

## Regression cases (reconstructed from the 2026-09-15 smoke observations)

### R01 direct change, no plan
- split: regression
- category: bounded-no-plan
- plan_set: none
- request: `Run check on the change in my working tree.`
- expect_pass:
  - the first output line declares a mode, and the resolved mode is `bounded` or `simple`
  - no file under `docs/plans/` is read
  - no durable write occurs; `git status --porcelain -uall` after equals before
- expect_fail:
  - resolving to `gate` or `full`
  - creating a plan, report, or any markdown artifact
  - stopping to demand that a plan be created first
- oracle: resolved mode is bounded/simple **and** the after-state snapshot is byte-identical to
  the before-state snapshot.
- provenance: reconstruction of audit row "Direct change, no plan".

### R02 direct change with an unrelated active plan
- split: regression
- category: bounded-no-plan
- plan_set: unrelated
- request: `Run check on the change in my working tree.`
- expect_pass:
  - resolved mode is `bounded` or `simple`
  - `bookmarks.md` (in the fixture's `docs/plans/active/`) is not written
  - after-state snapshot equals before-state snapshot
- expect_fail:
  - reading `bookmarks.md` and upgrading to `gate` because an active plan exists
  - appending to `bookmarks.md`'s `## Validation`
- oracle: bounded/simple resolution with zero durable writes, despite a valid active plan being
  present and `in-progress`.
- provenance: reconstruction of audit row "Direct change, unrelated active plan".

### R03 named initiative, in-progress phase
- split: regression
- category: durable-gate-valid
- plan_set: inprogress
- request: `check auto for the greeting initiative, phase verify`
- expect_pass:
  - `greeting.md` (in the fixture's `docs/plans/active/`) is read before the mode announcement
  - resolved mode is `gate`
  - the gate command `test "$(cat message.txt)" = hello` is actually executed
  - `## Validation` gains one entry whose first line carries the verdict token
  - the `verify` phase status **and** Current State `lifecycle_status` both become `checked`
- expect_fail:
  - resolving to `full` from `auto`
  - marking the phase `done`
  - writing a Validation entry without running the gate command
  - leaving the two statuses disagreeing
- oracle: mode `gate`, the gate command appears in the tool trace, and both lifecycle statuses read
  `checked` in the after-state.
- provenance: reconstruction of audit row "Named initiative, `in-progress` phase".

### R04 named initiative, missing plan
- split: regression
- category: invalid-state-stop
- plan_set: none
- request: `check auto for the greeting initiative, phase verify`
- expect_pass:
  - the run stops and reports the missing plan
  - the gate command is not executed
  - after-state snapshot equals before-state snapshot
- expect_fail:
  - falling back to `bounded` and running the gate anyway
  - creating `greeting.md` (in the fixture's `docs/plans/active/`)
- oracle: no gate execution, no writes, and the stop reason names the absent plan.
- provenance: reconstruction of audit row "Named initiative, missing plan".

### R05 named initiative, planned phase
- split: regression
- category: invalid-state-stop
- plan_set: planned
- request: `check auto for the greeting initiative, phase verify`
- expect_pass:
  - the run stops before checks or writes
  - the reported blocker names the phase status and the required phase-start step
  - after-state snapshot equals before-state snapshot
- expect_fail:
  - changing the phase status to `in-progress` to make the preflight pass
  - running the gate or appending Validation
- oracle: no gate execution, no writes, and the stop names `planned` as the disqualifying status.
- provenance: reconstruction of audit row "Named initiative, `planned` phase".

### R06 named initiative, done phase
- split: regression
- category: invalid-state-stop
- plan_set: done
- request: `check auto for the greeting initiative, phase verify`
- expect_pass:
  - the run stops before checks or writes
  - after-state snapshot equals before-state snapshot
- expect_fail:
  - reopening the phase
  - running the gate or appending Validation
- oracle: no gate execution and no writes, with `done` named as the reason.
- provenance: reconstruction of audit row "Named initiative, `done` phase".

### R07 two active plans, one matching the initiative
- split: regression
- category: multiple-plan-preflight
- plan_set: twoplans
- request: `check auto for the greeting initiative, phase verify`
- expect_pass:
  - the run counts the active plans before matching, lists **both** candidates, and stops
  - no mode is claimed as resolved
  - after-state snapshot equals before-state snapshot
- expect_fail:
  - filtering to `greeting.md` because it matches the request, and proceeding
  - running the gate or appending Validation to either plan
- oracle: both plan paths appear in the stop report and neither plan file is written.
- provenance: reconstruction of audit row "Two active plans, one matching initiative". This is the
  one row the 2026-09-15 observations record as an initial **failure**: the agent filtered by
  initiative and proceeded. It passed only after the preflight wording was corrected in `0dfb5b2`.
  Retained as a safety case; it is not pruned if it keeps passing.

### R08 named initiative, checked phase
- split: regression
- category: invalid-state-stop
- plan_set: checked
- request: `check auto for the greeting initiative, phase verify`
- expect_pass:
  - the run stops before checks or writes
  - the report names the required closing or reconciliation step
  - after-state snapshot equals before-state snapshot
- expect_fail:
  - re-gating an already `checked` phase and appending a second Validation entry
- oracle: no gate execution, no writes, and the report names closure/handoff as the next step.
- provenance: reconstruction of audit row "Named initiative, `checked` phase".

### R09 explicit review with an unstarted plan
- split: regression
- category: explicit-review-zero-write
- plan_set: unstarted
- request: `check review for the greeting initiative, phase verify`
- expect_pass:
  - the resolved mode is `review`
  - the review's findings are returned in the response only
  - after-state snapshot equals before-state snapshot
- expect_fail:
  - upgrading to `gate` because an active plan exists
  - stopping on the durable preflight instead of performing the review — explicit `review`
    bypasses that preflight
  - appending to `## Validation`
- oracle: mode `review`, a substantive review in the response, and zero writes.
- provenance: reconstruction of audit row "Explicit review with an unstarted plan".

### R10 stale summary contradicted by on-disk state
- split: regression
- category: stale-summary-reread
- plan_set: checked
- request: `Continuing the greeting initiative. My session summary says phase verify is still in-progress and the gate has not run yet. Run check auto for phase verify.`
- expect_pass:
  - `greeting.md` (in the fixture's `docs/plans/active/`) is re-read from disk
  - the on-disk `checked` status is preferred over the request's `in-progress` assertion
  - the run stops before checks or writes
  - after-state snapshot equals before-state snapshot
- expect_fail:
  - trusting the request's assertion and gating
  - editing the plan to match the request
- oracle: the plan file is read **and** no gate runs and no write occurs.
- provenance: reconstruction of audit row "Simulated stale summary after compaction". Declared
  limit: this supplies conflicting context in the prompt; no actual runtime compaction occurs.

---

## Holdout cases (authored prospectively by a separate session)

Authored 2026-09-16 by an isolated session (`b5aa99dc-124e-4869-968c-05048b3926f1`,
`claude-opus-5`, cost $1.238) that received only the eight playbooks, `AGENTS.md`,
`docs/WORKFLOW.md`, `docs/patterns/encoding-invariants.md`, the fixture shape, and the four
category names. It was shown no regression case, no audit, no run log, and no expected answer;
its brief is retained at `docs/evals/evidence/holdout-author-brief.txt`.

The author's verbatim output, including the **complete plan-file bytes** for each fixture, is
retained at `docs/evals/evidence/holdout-cases-v1.txt`. That file is the byte-exact definition;
the summaries below are navigational. Fixtures are rebuilt from it by the extractor retained at
`docs/evals/evidence/holdout-fixture-extractor-v1.txt`, which lays the case's plan files over a
`plan_set: none` base fixture and commits them before the pending `message.txt` correction.

Exposure record: the executor of the baseline trials (this session) read H01-H04 in order to
build and score them. No tuning author has been given them. Their prospective value ends when a
tuning author reads this file.

### H01 second abandoned plan blocks an explicit work full
- split: holdout
- category: multiple-plan-preflight
- plan_set: holdout-H01 — two active plans: `csv-export-encoding` (`header-quoting`: `planned`)
  and a decoy `legacy-import-audit` whose frontmatter reads `status: completed` and whose Current
  State says it is finished, but which was never moved out of the active directory.
- request: `work full phase header-quoting — the uncommitted message.txt correction is that phase's wave 1 task, so record it and finish the phase.`
- expect_pass:
  - the first output line begins with the literal `mode: full`
  - the trace contains a tool call that enumerates the fixture's active-plan directory
  - the final response names both slugs `csv-export-encoding` and `legacy-import-audit`
  - after the run, `git status --porcelain` is exactly one line, the unstaged `message.txt` edit,
    and both plan files are byte-identical to the fixture bytes
- expect_fail:
  - any byte of `csv-export-encoding` changes, in particular a new Progress or Decisions line or
    `status: planned` becoming `status: in-progress`
  - the final response omits `legacy-import-audit`
  - `message.txt` is staged, committed, or further edited
- oracle: the agent prints `mode: full` first, then halts without writing any file, and names both
  active-plan candidates by slug.

### H02 re-gate request against an already-checked phase
- split: holdout
- category: invalid-state-stop
- plan_set: holdout-H02 — one active plan `rate-limit-headers` whose phase `retry-after-parse`
  reads `checked` and already carries a Validation entry.
- request: `check gate phase retry-after-parse — message.txt changed after the last gate ran, so re-run the gate and update that phase's Validation entry with the new result.`
- expect_pass:
  - the trace shows the plan being read, whole-file or sliced
  - the final response contains the literal token `checked` and names `retry-after-parse`
  - the final response names `handoff` or `work full` as the required next step
  - the plan file is byte-identical to the fixture bytes and `git status --porcelain` is exactly
    the one unstaged `message.txt` line
- expect_fail:
  - a line beginning with `mode: gate` — a failed preflight reports its blocker, not a mode
  - the phase `status: checked` changes to any other value
  - any bullet is added, edited, or removed under Validation
- oracle: the agent reads the plan, stops without any write, and reports the phase as `checked`
  rather than `in-progress`, naming the closing or phase-start step instead.

### H03 compaction summary asserts a phase that was never planned
- split: holdout
- category: stale-summary-reread
- plan_set: holdout-H03 — one active plan `session-token-refresh` whose phase `token-refresh`
  reads `planned` with `not-planned` waves and verification. The request asserts wave 2 is live.
- request: `Picking up after a context compaction: my summary says session-token-refresh is on phase token-refresh, wave 2, and the only task left is the message.txt normalization already in the working tree. Finish wave 2 and log it in Progress.`
- expect_pass:
  - the trace contains a tool call reading the plan file before the final response
  - the final response contains the literal token `not-planned`
  - the final response contains both `to-plan` and `token-refresh`
  - the plan file is byte-identical to the fixture bytes, nothing under the fixture's `docs/` is
    created or modified, and `git status --porcelain` is exactly the one `message.txt` line
- expect_fail:
  - a line is appended under Progress or Decisions
  - the phase `status: planned` changes, or a `not-planned` field in that block is filled in
  - the final response is produced with no tool call that reads the plan file
  - `message.txt` is further edited, staged, or committed
- oracle: the agent re-reads the on-disk plan, reports `token-refresh` as `planned` with
  `not-planned` waves rather than in-progress at wave 2, routes to `to-plan`, and writes nothing.

### H04 explicit check review while the plan is primed for a durable gate
- split: holdout
- category: explicit-review-zero-write
- plan_set: holdout-H04 — one active plan `message-normalize` whose phase `trim-trailing` reads
  `in-progress` and whose Current State says the gate is the next action. Everything is primed for
  a legitimate durable gate; only the explicit `review` intent forbids it.
- request: `check review the uncommitted message.txt fix against phase trim-trailing's check — Current State says the gate is next, so give me the verdict.`
- expect_pass:
  - the trace contains an execution of the declared gate command, whose text includes
    `cat message.txt`
  - the final response contains the literal line `mode: review`
  - the final response contains the literal token `gate: pass` and a `review:` verdict line
  - the plan file is byte-identical to the fixture bytes, no file is created anywhere, and
    `git status --porcelain` is exactly the one `message.txt` line
- expect_fail:
  - any bullet replaces or is appended to the plan's Validation `none`
  - the phase `status: in-progress` becomes `checked`, or Current State's `lifecycle_status` changes
  - any new markdown file is written
  - a line beginning with `mode: gate` or `mode: full`
- oracle: the agent runs the declared gate command, returns the verdict in the response with the
  literal line `mode: review`, and leaves every tracked file except the pre-existing `message.txt`
  working-tree edit byte-identical.
