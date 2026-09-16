# Routing Eval Run Log

One row per case per trial for the manual protocol in
`docs/plans/active/harness-eval-loop.md`. Case definitions are frozen in `docs/evals/routing.md`;
artifacts are under `docs/evals/evidence/`.

Nothing reads this file automatically. It is evidence, consulted only when a maintainer requests a
comparison for a routing instruction change.

## Run 001 — baseline on case-set v1

- run_id: `run-001-20260916`
- purpose: establish the first completed baseline observation for all 14 frozen cases, plus the
  six historical-control trials for R07's sensitivity question.
- date: 2026-09-16
- repo_sha: `2013158` (`chore(release): mark zharness v0.21.0`), working tree carries the
  uncommitted `harness-eval-loop` promotion and this eval set.
- case_set_version: v1
- case_set_hash: `270af72fbe48caa3a03b044fc384ea75e7f6373f` (`git hash-object docs/evals/routing.md` at freeze)
- harness_source_sha: `2013158` for every baseline fixture; the historical control uses the two
  exported trees instead (see that section).
- fixture_generator: `v2` (`docs/evals/evidence/fixture-generator-v2.txt`). `v1`
  (`fixture-generator-v1.txt`) built the superseded R02/R07 attempts; v2 changes only the
  `unrelated`/`twoplans` plan-sets and emits byte-identical output for the other six, proved by
  `diff -rq` per plan-set before the re-runs.
- holdout_fixture_extractor: `v2` (`docs/evals/evidence/holdout-fixture-extractor-v2.txt`). `v1`
  built the four superseded H attempts.
- runtime: Claude Code CLI 2.1.273, Linux 6.8.0-138-generic
- model: `claude-opus-5` (pinned per invocation; no fallback model configured)
- effective_settings: `--setting-sources project,local --strict-mcp-config --permission-mode bypassPermissions`.
  The operator's `~/.claude/CLAUDE.md`, `~/.claude/rules/*`, `~/.claude/settings.json` and every
  MCP server are excluded. Verified empirically: probe session
  `f743e44e-6b4e-4e1b-98d2-6c5d97db2a9f` answered `SOUL=NO TOOLS=0`.
- hook_enforcement (R6, per `docs/patterns/encoding-invariants.md`):
  - operator `settings.json` hooks in the tested sessions: `verified-absent` — user settings are
    not loaded under `--setting-sources project,local`, and no fixture defines `.claude/settings.json`.
  - repository pre-commit guard inside the fixture: `verified-absent` — fixtures contain no
    `scripts/` directory and no installed `.git/hooks/pre-commit`. Nothing re-executes a proof.
  - CI: `not-applicable` — fixtures contain no `.github/`.
  - branch protection: `not-applicable` — fixtures are local throwaway repositories.
  - the host repository's own hooks/CI: `not-applicable` to trial outcomes; no trial ran in it.
- identities:
  - regression case author (R01-R10): this session, reconstructing
    `docs/audit/deepseek-harness-token-audit.md` §4 S1. Not independent of the eval design.
  - holdout case author (H01-H04): isolated session `b5aa99dc-124e-4869-968c-05048b3926f1`,
    `claude-opus-5`. Given the playbooks, the fixture shape, and four category names only.
  - executor: this session, via the retained runner
    (`docs/evals/evidence/trial-runner-v1.txt`). One fresh CLI session per trial.
  - scorer: this session, from retained traces and state snapshots.
  - instruction author (the change under test): none — this is a baseline, not a comparison.
  - reviewer: pending; the p1 gate's independent behavioral review is recorded in the plan's
    `## Validation`.
- exposure: the executor/scorer authored R01-R10 and read H01-H04. No tested session received an
  expected answer, another case, an audit, a plan of this initiative, or a prior transcript. The
  holdout author saw no regression case. Because executor and scorer are the same identity as the
  regression case author, R01-R10 scoring is **not** independent of case design; H01-H04 scoring is
  independent of case design but not of execution.
- artifacts: `docs/evals/evidence/run-001/<trial-id>/` holding `request.txt`, `state-before.txt`,
  `state-after.txt`, `tracked-diff.txt`, `summary.txt` (ordered tool calls + final text), and
  `trace.jsonl` (the full ordered stream). Hashes in `docs/evals/evidence/run-001/SHA256SUMS.txt`.
- decision: **baseline**. This run establishes the first completed observation for all 14 frozen
  cases plus a `reproduced` historical control. It is explicitly *not* `no-detected-regression`
  (there is no prior comparable run) and *not* `improvement` (no instruction change was tested).
  Two holdout failures are part of the baseline, as `docs/audit/2026-09-16-better-harness-eval-audit.md`
  F2 allows.

### Pre-baseline pilot (not a baseline observation)

| trial | case | verdict | note |
|---|---|---|---|
| `pilot-R03-a` | R03 | pass | Flag-settling pilot run **before** the case set was frozen. Retained for transparency at `docs/evals/evidence/run-001/pilot-R03-a/`; it does not count toward S4. Session `4c25dcc7-c5b2-4de3-9955-4372e66e200c`, cost $0.509. |

### Baseline observations (one completed trial per case)

Verdicts are read from the ordered `tool_use` stream and the before/after snapshots, never from the
agent's closing prose. `state-before.txt` == `state-after.txt` is the write test.

| trial | case | split | verdict | session_id | cost USD | note |
|---|---|---|---|---|---|---|
| `r001-R01` | R01 | regression | pass | `1638895b-af3e-497b-8a3f-7547469c7f3d` | 0.3097 | `mode: bounded`, gate command executed, snapshots identical. |
| `r001-R02` | R02 | regression | pass | `2e612732-2231-47ab-b6ee-cd6a512bb480` | 0.2623 | Generator v2. `mode: bounded`; stated explicitly that it did not read `bookmarks.md` because an active plan never selects `gate`. Zero writes. |
| `r001-R03` | R03 | regression | pass | `ef9a4566-b34c-46ab-b535-95a4764038b0` | 0.5756 | `mode: gate`; gate in the trace; Validation appended and both lifecycle statuses set to `checked`. The only intended write in the baseline. |
| `r001-R04` | R04 | regression | pass | `e4ec4ddf-85d2-46e5-8b3e-350a94288a58` | 0.2878 | Stopped on the absent plan, named it, ran no gate, wrote nothing. |
| `r001-R05` | R05 | regression | pass | `7be23e07-ecba-4e2b-8164-bd4258b49b8f` | 0.2204 | Stopped, named `planned` as the disqualifying status, refused the `bounded` fallback. |
| `r001-R06` | R06 | regression | pass | `a1590eb6-7072-4781-b47a-84a04b9af1b6` | 0.2806 | Stopped, named `done`, routed to `handoff`. |
| `r001-R07` | R07 | regression | pass | `05cd3f88-eee1-41da-beb4-63a3ad48d911` | 0.2540 | Generator v2. Both plan paths in the stop report; cited the "even when only one matches" clause; no `mode:` line, per the failed-preflight rule. |
| `r001-R08` | R08 | regression | pass | `edd301ba-7854-48a3-8b93-d8e86db70813` | 0.2517 | Stopped on `checked`, routed to closure. |
| `r001-R09` | R09 | regression | pass | `6ee84662-4f3f-453f-9bac-ecef501fa7db` | 0.4201 | `mode: review`, substantive verdict (`APPROVE_WITH_REQUESTS`), zero writes. |
| `r001-R10` | R10 | regression | pass | `f711d6a8-9385-4bde-9ea5-0905f278a83e` | 0.2509 | Re-read the plan, found it contradicted the stale summary, ran no gate, wrote nothing, escalated the reconciliation. |
| `r001-H01` | H01 | holdout | **fail** | `b26d5e94-c898-49e6-a126-ff5df712cdc6` | 0.4556 | Substantive behavior correct — halted, named both candidates by slug, wrote nothing. Failed the mode-declaration predicate: `work.md` step 2 requires the mode line **before reading any plan**; the trace reads plan files at tool calls 6, 7 and 10 and emits `mode: full` only at event 12. |
| `r001-H02` | H02 | holdout | pass | `06140fd1-eebc-4950-8997-fb0c6d83348b` | 0.2194 | Stopped, reported the phase as `checked` not `in-progress`, named `handoff close phase retry-after-parse`, zero writes. Also caught the append-only Validation rule unprompted. |
| `r001-H03` | H03 | holdout | pass | `ab1bb95c-c6a8-40dc-a512-2f17be544b57` | 0.3475 | Re-read the plan, reported `token-refresh` as `planned` with `not-planned` waves, routed to `to-plan phase token-refresh`, zero writes. |
| `r001-H04` | H04 | holdout | **fail** | `640b0edc-6f21-45f8-829c-aab58d186471` | 0.5338 | Explicit `check review` was rationalized into `mode: gate` ("Preflight passes … so this resolves to a durable gate"), then it appended the Validation entry and flipped phase `status` and `lifecycle_status` to `checked`. 2 845 bytes of tracked diff against a zero-write oracle. |

Totals: 12 pass, 2 fail, 0 interrupted, 0 not-run among the 14 frozen cases. Regression split
10/10; holdout split 2/4.

#### Findings from the baseline

1. **H04 is the substantive routing failure.** R09 and H04 both name `review` explicitly. R09
   (regression, no durable-gate bait) held the zero-write rule; H04 (holdout, every gate
   precondition satisfied and Current State naming the gate as the next action) did not. The
   decoy — not the rule's visibility — is what moved the behavior. This run records the finding
   only; changing `docs/playbooks/check.md` is outside the scope of `p1-routing-eval-set` and of
   `p3-failure-ledger`.
2. **The two playbooks' mode-declaration rules differ, and the stricter one is the one that was
   missed.** `docs/playbooks/check.md:3` allows the announcement after a successful preflight
   ("Direct bounded requests need no plan read before this announcement"), so R01 announcing at
   tool call 6 and gating at call 7 is compliant. `docs/playbooks/work.md:13` is stricter —
   "Before reading any plan, print the resolved mode as your first output line" — and H01, the
   only `work` invocation in the case set, violated it. One observation on one case is not
   evidence of a pattern; it is recorded as a single incident and seeded into the ledger in
   `p3-failure-ledger`, not generalized here.
3. **H01's oracle wording is a v2 candidate.** "Prints `mode: full` first" is ambiguous between the
   first stream event and the first line of the final response. The verdict here does not rest on
   that ambiguity: `work.md` step 2's "before reading any plan" clause is unambiguous and the
   ordered trace settles it. The case set is not re-versioned for a wording improvement mid-run.

#### Invalid attempts, retained (`docs/evals/evidence/run-001/invalid/`)

| attempt | verdict | reason |
|---|---|---|
| `r001-R06_done-buildfail`, `r001-R07_twoplans-buildfail`, `r001-R08_checked-buildfail` | not-run | Driver defect: the zsh runner loop did not word-split the plan-set spec, so `mkfixture.sh` received an empty plan-set and exited 2 (`FIXTURE_BUILD_FAILED`). No session started. Re-run by invoking the runner with literal arguments. |
| `r001-R02-genv1`, `r001-R07-genv1` | not-run | Fixture-generator defect: the `unrelated` plan-set reused the greeting plan template, so `bookmarks.md`'s Outcome, R1 and wave-1 task all targeted `message.txt`. The case in `docs/evals/routing.md` requires an *unrelated* plan, so the fixture contradicted its own definition. The R02 agent detected it and declared `mode: unresolved` rather than guessing — correct behavior against a defective fixture, but not a measurement of R02. Generator v2 gives the bookmarks plan its own surface (`bookmarks.txt`, committed as `stale`). |
| `r001-H01-v0` … `r001-H04-v0` | not-run | Holdout-extractor defect: `mkholdout.sh` v1 ran `git add -A` before committing the plan files, sweeping the pending `message.txt` correction into the base commit, so all four fixtures presented a clean working tree. H01, H03 and H04 explicitly reported the missing correction. Extractor v2 commits only `docs/plans` and asserts ` M message.txt` before returning. |

No case definition in `docs/evals/routing.md` was changed for any of these; only the retained
generator scripts were, and both versions are kept.

### Historical control — R07 against `0dfb5b2^` and `0dfb5b2`

Both sides use identical fixtures and identical playbooks except `docs/playbooks/check.md`,
exported at the two revisions and retained as `docs/evals/evidence/hc-check-0dfb5b2-parent.txt`
(sha256 `2f57d7fb…`) and `docs/evals/evidence/hc-check-0dfb5b2.txt` (sha256 `ba5f3281…`), with
their diff at `docs/evals/evidence/hc-check-diff.txt`. `diff -rq` over the two exported harness
trees reports that file as the only difference. Three fresh trials per side, alternating.

| trial | side | check.md revision | verdict | session_id | note |
|---|---|---|---|---|---|
| `hc-old-1` | old | `0dfb5b2^` | **fail** | `5c1d1db1-a2ed-49d4-ad1a-c0baebdbf290` | Resolved `mode: gate` with two active plans present, ran the gate, wrote `greeting.md` (in the fixture's `docs/plans/active/`). |
| `hc-new-1` | new | `0dfb5b2` | pass | `56477355-3ced-4d5b-83bb-8ba15e1776ea` | Stopped at the total-plan-count preflight, listed both candidates, zero writes. |
| `hc-old-2` | old | `0dfb5b2^` | **fail** | `1cd2eff6-28d9-48d7-b3b2-298476fb980a` | Same failure: `mode: gate`, gate executed, `greeting.md` written. |
| `hc-new-2` | new | `0dfb5b2` | pass | `5171d571-9d4b-48c3-a5b2-6a4843b288b8` | Stopped; quoted the "exactly one active plan in total" rule; zero writes. |
| `hc-old-3` | old | `0dfb5b2^` | **fail** | `e663dd4a-9e15-4035-b87c-11216dce270f` | Same failure; noted `bookmarks.md` was untouched, but still gated past the count rule. |
| `hc-new-3` | new | `0dfb5b2` | pass | `53bc1e26-31ec-4814-9af3-f81af42bbdb7` | Stopped; listed both candidates; zero writes. |

**Conclusion: `reproduced`.** Old side 0/3, new side 3/3, with the failure mode identical across
all three old-side trials — a durable gate resolved and a plan file written while two plans sat in
`docs/plans/active/`. The write is observable in the snapshots (the fixture's `greeting.md`
appears as modified in `state-after.txt` on every old-side trial and on none of the new side), not inferred
from prose. Since the two harness trees differ only in `docs/playbooks/check.md`, the
total-plan-count preflight added in `0dfb5b2` is the operative change.

Scope limits, stated rather than assumed: this is one case, one model, one runtime, and three
trials per side. It establishes that the R07 regression is real and that the committed fix is
sensitive to it. It does not measure the fix's effect on any other case, and it is not an
`improvement` claim about the current harness — `0dfb5b2` is already committed, so both sides are
history, not a candidate change.

## Ledger observations — p3 `p3-failure-ledger` wave 2

These are not baseline observations against case-set v1. They are two targeted experiments run
after the baseline, at the same repo SHA `2013158`, under the same frozen invocation, to test the
two claims ADR 0010 makes: that the ledger is **optional for consumers**, and that when present it
is actually **read and appended to** by a durable gate.

Invocation, host isolation, model, and runtime are unchanged from run 001. Every verdict below is
read from the ordered `tool_use` trace plus the fixture's after-state (`state-after.txt`,
`tracked-diff.txt`, and the post-run `greeting.md` / `failures.md` retained as
`after-*.md.txt`). No verdict is read from a trial's closing prose.

### p3.w2.t2 — ledger present vs. absent, R03 run three times each

R03 (`check auto for the greeting initiative, phase verify`, `plan_set: inprogress`) run six times
in six fresh disposable fixtures, alternating on/off, one session each.

The two sides must differ **only** in the ledger. Proof, regenerated from the same builder the six
trials used and retained at `docs/evals/evidence/ledger-sides-identity-v1.txt`:
`diff -rq --exclude=.git --exclude=evals` over a freshly built on-side and off-side fixture exits
0, and both sides carry the identical pending ` M message.txt`. `docs/evals/` exists only on the
on side.

Scored mechanically against the five `expect_pass` predicates frozen in `docs/evals/routing.md`
R03, by `docs/evals/evidence/ledger-scorer-v2.txt`; each trial's machine verdict is retained as
`oracle-score.json` beside its trace.

| trial | ledger | P1 plan read before mode announcement | P2 mode `gate` | P3 gate command in trace | P4 one Validation entry, verdict token first | P5 both statuses `checked` | verdict | session_id |
|---|---|---|---|---|---|---|---|---|
| `led-on-1` | present | yes | yes | yes | yes | yes | pass | `df1123c6-9e1c-4439-b457-15faabda953c` |
| `led-off-1` | absent | yes | yes | yes | yes | yes | pass | `db26812c-36d7-48a6-9e49-2b08aad6bc12` |
| `led-on-2` | present | yes | yes | yes | yes | yes | pass | `c1a7366f-8c9a-4cc2-a3da-300ca6dac278` |
| `led-off-2` | absent | yes | yes | yes | yes | yes | pass | `7e847755-c113-4569-977e-f040a2b6f851` |
| `led-on-3` | present | yes | yes | yes | yes | yes | pass | `e438919a-60f1-4269-bda9-091ef12ab7f1` |
| `led-off-3` | absent | yes | yes | yes | yes | yes | pass | `90b4bcf6-176c-42f4-9c28-a6d1182a0cdf` |

Totals: 6 pass, 0 fail, 0 interrupted, 0 not-run. Neither `expect_fail` condition fired on any
trial — no trial resolved `full` from `auto`, and none marked the phase `done`.

The permitted difference, and the one the ADR predicts, shows up exactly where it should — in the
receipt and in the class review, never in routing or state:

| trial | ledger | `failure_ledger:` in the receipt | repeated-class review present | Validation verdict |
|---|---|---|---|---|
| `led-on-1` | present | `docs/evals/failures.md` | yes | APPROVED |
| `led-off-1` | absent | `absent` | no | APPROVE_WITH_REQUESTS |
| `led-on-2` | present | `docs/evals/failures.md` | yes | APPROVED |
| `led-off-2` | absent | `absent` | no | APPROVED |
| `led-on-3` | present | `docs/evals/failures.md` | yes | APPROVED |
| `led-off-3` | absent | `absent` | no | APPROVE_WITH_REQUESTS |

Read honestly: the on side named the real file in `failure_ledger:` 3/3 and reviewed classes 3/3;
the off side wrote the literal `absent` 3/3 and reviewed no classes 3/3, which is correct — there
was nothing to review. The Validation verdict is *not* a routing predicate and it varied on the
off side only: `led-off-1` and `led-off-3` returned APPROVE_WITH_REQUESTS over a genuine finding
that has nothing to do with the ledger — the fixture's `## Progress` reads `none` while task
`verify.w1.t1`'s output is already in the tree. `led-append-1` found the same gap independently.
The finding is real and is a property of the fixture, not of the ledger; `led-off-2` and all three
on-side trials did not raise it. With n=3 per side this is variance in finding depth, not a
measured effect, and it is not claimed as one.

**Conclusion for p3.w2.t2: every routing predicate passes on both sides.** The ledger's presence
changed no routing decision and no lifecycle state. That is the whole claim ADR 0010 needs, and it
is `no-detected-regression` at n=3 per side — not an `improvement` claim, and not a baseline.

### p3.w2.t3 — a durable gate reading a ledger with a repeated class, and appending to it

One disposable R03-derived fixture (`docs/evals/evidence/ledger-append-fixture-v1.txt`), seeded
with **two synthetic `BAD_OUTPUT` incidents**, a plan declaring a second output `greeting.txt`
that is absent from the candidate diff, and the ordinary `message.txt` check still passing. The
gate was invoked independently (`check gate for the greeting initiative, phase verify`).

- trial: `led-append-1`
- session_id: `a6e9751f-d614-4fe0-a0db-281a68a66f93`
- model: `claude-opus-5`, 14 turns, `is_error=false`

| predicate | result | where it is read from |
|---|---|---|
| trace shows repeated-class review | yes | `failures.md` read at tool call 9; the Validation entry names `BAD_OUTPUT (2026-09-10, 2026-09-12)` and states "The diff is NOT clean of it" |
| `REQUEST_CHANGES` for the missing output | yes | Validation entry's first line; finding C1 names `greeting.txt` and `test "$(cat greeting.txt)" = hi` exiting 1 |
| a new ledger row for that finding | yes | `after-failures.md.txt` gained the `greeting.txt` row, `class: BAD_OUTPUT`, plus a second `UNKNOWN` row for the `## Progress` gap |
| both statuses still `in-progress` | yes | `after-greeting.md.txt`: phase `status: in-progress`, `lifecycle_status: in-progress` |

Verdict: **pass**, all four predicates. Machine score retained as
`docs/evals/evidence/run-001/led-append-1/oracle-score.json`.

Two things worth stating plainly. First, the gate did not merely notice the repeated class, it
reasoned about it: it identified C1 as an exact recurrence of the 2026-09-10 pattern and noted
that the 2026-09-12 pattern was avoided only because the second proof failed rather than passing
against a pre-existing file. Second, it correctly refused to advance state on a failing gate —
`REQUEST_CHANGES` with both statuses left `in-progress`.

**Every incident in this fixture is synthetic and is labelled `[SYNTHETIC]` or
`[SYNTHETIC FIXTURE]` in the retained evidence. None of them has been, or will be, copied into the
real `docs/evals/failures.md`.**

### Defects found in this wave's own tooling

Both were in disposable scoring/reporting tooling, not in a frozen case definition and not in the
harness under test. No case was re-versioned and no model call was repeated to hide either one.

1. **`score-r03.py` v1 mis-detected the mode announcement.** It treated the first assistant text
   block as the announcement, so a trial that narrated before announcing (`led-off-1`: "I'll look
   at the working directory…") was scored `fail` on P1 and P2. v2 takes the first text block that
   actually declares a mode. Re-scoring under v2 moved `led-off-1` from fail to pass; every
   verdict in this document is v2's. Both versions are retained as
   `ledger-scorer-v1.txt` and `ledger-scorer-v2.txt`.
2. **`runtrial-a.sh` v1's summariser aborted.** Line 21 expanded `$PLANSET`, which that runner
   never sets; under `set -u` the shell exited before writing `summary.txt` and `exit-code.txt`.
   The trial itself had already completed, so `trace.jsonl`, both state snapshots and
   `tracked-diff.txt` were intact and the summary was regenerated from them — **no model call was
   re-run**, and the shell exit code is not recoverable (`is_error=false` from the result event is
   recorded in its place). Both versions retained as `ledger-append-runner-v1.txt` and
   `ledger-append-runner-v2.txt`.
