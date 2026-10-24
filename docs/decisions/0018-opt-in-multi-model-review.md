# 0018 — Multi-model review is an opt-in panel for `full` and high-risk checks

## Status

Accepted 2026-10-24 on merge of the PR that implements the policy for issue #156. The owner asked for the capability during the pstack review (`docs/research/2026-10-22-pstack-retro-5.md:9`). The policy text ships in `skills/stages/check/references/check-review.md`; no executable runner ships. The issue's test criterion is not met (see Consequences).

## Context

`check-review.md` runs the Spec and Standards axes with one model; sub-agents split by axis, not by model. pstack `interrogate` runs one reviewer per model and ranks findings that two or more models report highest. The audit marks the gap "VERIFIED absent, HYPOTHESIS value": no ledger row or eval shows a defect that one judge missed and a second model would have caught (`docs/research/pstack-concepts-audit.md:62`). The cost is N times the review tokens (`docs/research/2026-10-22-pstack-retro-5.md`, section 1).

## Decision

Ship the mode as opt-in policy, not a default. The value is a hypothesis, so the design keeps cost and authority bounded.

| Parameter | Rule |
|---|---|
| Opt-in | Only when the owner asks for it by name and the mode is `full` or `high-risk`. Never `bounded` or `gate`, never automatic. |
| Runner and models | Claude sub-agents with an explicit `model` per `work-fanout.md` "Models per role"; the owner names the models. Other CLIs are out of scope until an owner names one. |
| Reviewer count | 2 to 3. |
| Budget | The owner names it; the ceiling is 3x one single-agent review. No named budget means no panel. A reviewer over its timebox is dropped. |
| Same input | One `base_sha`, `head_sha` and dirty-diff hash, resolved once and written into every brief. A report with other SHAs is rejected (`work-fanout.md:56`). |
| Separate context | One brief per reviewer; no reviewer sees another's output. |
| Axes | Each reviewer reports under `Spec` and `Standards`. A finding carries axis, file:line or requirement ID, and evidence. |
| Dedup | Same axis and file with line ranges overlapping or within 3 lines (Spec without a file: same requirement ID) is one finding. Keep the highest severity, lowest reviewer index on a tie, record `agreed: n/N`. |
| Conflict | The higher severity stands until the lead validates it against the repository and proof output; an unvalidatable finding is listed as unverified and does not block. |
| Priority | Severity, then `agreed`. Agreement orders the list; it never raises severity and is never proof. |
| Fallback | Fewer than 2 valid reports (unavailable runner, timeout, over budget, bad SHA) finishes as a single-agent review and prints `panel: degraded (valid/N)`. |
| Verdict | The `judge:` declaration in `check.md` step 7 is unchanged. The lead's evidence decides the verdict. |

Rejected: a default-on panel (N times cost on every change, no evidenced defect), weighting severity by agreement (turns a vote into proof), and fail-closed on a missing reviewer (an opt-in extra would then weaken the baseline review it sits on).

## Consequences

The owner can buy a second opinion for a risky change at a known ceiling, and the single-agent path is untouched. The cost is prompt-level only: dedup, budget and fallback are followed by the lead, not enforced by code. The issue's criterion "tests cover unavailable reviewers, conflicting findings, and budget limits" is not met, because there is no code to test. The contract test for `check-review.md` still pins only the two-axis text (`cli/src/embedded.rs:434`). Issue #156 stays open for those tests, which need an executable merge step (a script that takes reviewer reports and applies the dedup, conflict and budget rules) before they mean anything. Whether that script is worth building should wait for the first real panel run to show a missed defect.

## Authority

`docs/research/2026-10-22-pstack-retro-5.md:9` (owner request); `docs/research/pstack-concepts-audit.md:62`; `skills/stages/check/references/check-review.md` (Multi-model panel section); `skills/stages/work/references/work-fanout.md:56`; `skills/stages/check/references/check.md:23`; issue #156.
