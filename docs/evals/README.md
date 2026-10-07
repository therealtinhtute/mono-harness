# Evals — manual routing evaluation

Maintainer-owned, opt-in behavioral evaluation for routing instruction changes. Nothing in this
directory is consumed automatically: no playbook, guard, CI job, or installer reads it. The
protocol below runs only when a maintainer explicitly requests a comparison for a routing
instruction change.

| Path | Role |
|---|---|
| `docs/evals/routing.md` | frozen case set (R01–R10 regression, H01–H04 holdout), plus the fixture template, invocation, and scoring rules every trial uses |
| `docs/evals/runs.md` | run log; one row per case per trial |
| `docs/evals/evidence/` | retained generators, fixtures, traces, snapshots, and scorers |
| `docs/evals/failures.md` | optional failure ledger per `docs/decisions/0010-local-failure-ledger.md`; its consumers are `docs/playbooks/check.md` steps 4 and 10, not this protocol |

## Contract

The rules below are carried verbatim from the initiative record that owned this protocol,
`git show 3bab3c0:docs/plans/completed/harness-eval-loop.md` — requirements R4/R5, the exposure
clause of manual-protocol step 3, step 9, and the run-log header paragraph. Steps 1–8 of that
protocol describe work already recorded in `docs/evals/runs.md` run-001 or restated in
`docs/evals/routing.md`; the full historical text remains in git history at that pin.

### Trigger and independence

The maintainer explicitly requests a comparison for a routing instruction change. A reviewer
other than the change author consumes the run log and artifacts. Equal scores without observed
regressions mean `no-detected-regression`, not improvement. A behavioral improvement claim
requires an observed targeted gain, no observed regression elsewhere, and the repeated
comparison defined below.

Baseline/comparison executors differ from the instruction-change author. The holdout
author/evaluator must not author candidate tuning. Record identities and exposure; a fresh
session alone does not establish that holdout contents were unseen.

Do not expose prospective holdout contents or detailed failures to tuning authors. If exposure
occurs, retain that case as regression evidence and author a replacement under a new ID before
the next prospective comparison. Record the exposure without rewriting its original split. This
is procedural separation, not a security sandbox claim.

### Reviewing a comparison

Pin baseline/candidate texts and use the same corpus, model, settings, and environment;
runtime/model drift requires a new baseline. Compare both revisions on every case. An
improvement claim requires three fresh trials per revision per case, raw counts, no newly
failing predicates, and independent review. Mixed evidence is `inconclusive`. An observed
regression blocks acceptance under this optional protocol. The maintainer records the
decision/run ID in the requesting review or initiative. Existing playbooks do not invoke this
procedure automatically.

### Run records

The run log has one row per case/trial. Headers carry date/purpose, full repo SHA, case-set SHA,
target hashes, runtime/version/model/settings, instruction author, fixture author, executor,
reviewer, exposure status, enforcement evidence, artifact paths/hashes, and decision.
