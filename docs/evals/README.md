# Evals — manual routing evaluation

Maintainer-owned behavioral evaluation surfaces. The routing suite — the case set in
`docs/evals/routing.md`, the run log in `docs/evals/runs.md`, and retained artifacts under
`docs/evals/evidence/` — is opt-in: no playbook, guard, CI job, or installer reads it, and the
protocol below runs only when a maintainer explicitly requests a comparison for a routing
instruction change. `docs/evals/failures.md` is a separate surface with conditional automatic
consumers: `docs/playbooks/check.md` steps 4 and 10 read and append it whenever it exists.

| Path | Role |
|---|---|
| `docs/evals/routing.md` | frozen case set (R01–R10 regression, H01–H04 holdout), plus the fixture template, invocation, and scoring rules every trial uses |
| `docs/evals/runs.md` | run log; one row per case per trial |
| `docs/evals/evidence/` | retained generators, fixtures, traces, snapshots, and scorers |
| `docs/evals/failures.md` | optional failure ledger per `docs/decisions/0010-local-failure-ledger.md`; conditional consumers `docs/playbooks/check.md` steps 4 and 10 |

## Contract

The standing rules of the manual protocol live in this file. They are carried verbatim from the
initiative record that owned the protocol,
`git show 3bab3c0:docs/plans/completed/harness-eval-loop.md`. Disposition of its nine protocol
steps, one by one:

| Step | Disposition |
|---|---|
| 1 Freeze definitions | Executed — all 14 cases are frozen as case-set v1 in `docs/evals/routing.md`. Its case-content rule is carried below; the new-version-on-edit rule is restated there. |
| 2 Build isolated fixtures | Restated in `docs/evals/routing.md` (Fixture template, plan-set tokens); artifacts are captured outside the disposable fixture per step 4. |
| 3 Separate context | Restated in `docs/evals/routing.md` (Invocation: fresh session per trial, request and fixture only, exclusion list); its exposure clause is carried under Trigger and independence. |
| 4 Retain evidence | Carried verbatim below. |
| 5 Score actions | Restated in `docs/evals/routing.md` (Scoring) and encoded per case in `expect_pass`/`expect_fail`. |
| 6 Handle interruptions | Verdict semantics restated in `docs/evals/routing.md` (Scoring); its completeness and no-cherry-picking tail is carried below. |
| 7 Measure the baseline | Executed — run-001 in `docs/evals/runs.md`. Its baseline-definition sentence is carried below for re-baselines after drift; the `historical-report` import was a one-time action. |
| 8 Compare the historical control | Executed — run-001's `reproduced` R07 control (`0dfb5b2^` vs `0dfb5b2`), a one-time comparison with no standing rule left behind. |
| 9 Review future comparisons | Carried verbatim below. |

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

### Cases, trials, and evidence

Every case includes complete input bytes, request, expected reads/checks/writes, allowed
changes, and provenance. Mark reconstructed historical fixtures as reconstructions.

One completed trial per case establishes a smoke baseline on the tested SHA, not automatically
“v0.21.0.”

Use `docs/evals/evidence/<run-id>/` for redacted ordered tool traces, initial/final snapshots
including untracked files, and scoring. Record hashes and repo-relative paths in `runs.md`; a
machine-local transcript pointer alone is insufficient. Remove credentials and unrelated host
data while preserving scoring events. Store historical fixture contents as `.txt` so they are
not treated as live Markdown cross-references.

Missing sessions/artifacts block completion as `BLOCKED_CONTEXT` with a failure class; do not
cherry-pick attempts.

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
