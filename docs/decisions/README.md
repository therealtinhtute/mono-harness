# Decision Records

Architecture decisions that are already made, already implemented, and expensive to reverse. Each one exists because the reasoning was otherwise reachable only through a commit message or a session transcript.

An ADR here is a record, not a proposal. It is written after the decision has landed in code, and it states what was decided, why, what was rejected, and what it costs. Nothing in this directory is a plan — plans live in `docs/plans/`. ADRs are the absorb path for expensive-to-reverse decisions; a completed plan is a run log, not project knowledge.

Exception: the owner explicitly requested ADRs 0015 and 0016 before implementation. Its status distinguishes the accepted decision from pending implementation; phase execution remains in the plan.

## Index

| # | Decision | Status |
|---|---|---|
| [0001](0001-markdown-as-source-of-truth.md) | Markdown is the source of truth; `harness.db` is a derived, rebuildable index | Superseded by [0006](0006-v015-authority.md) |
| [0002](0002-single-active-plan-resolver.md) | One resolver owns the "at most one active plan" invariant, returning a Stop contract | Superseded by [0006](0006-v015-authority.md) |
| [0003](0003-durable-memory-not-wired-into-playbooks.md) | Durable memory ships as an opt-in CLI surface, unwired from the spine playbooks | Superseded by [0006](0006-v015-authority.md) |
| [0004](0004-docs-directory-deletion-655c6ac.md) | Recovery position after commit `655c6ac` deleted `docs/` | Accepted |
| [0005](0005-authored-documentation-boundary.md) | Audit guards authored-document presence, not authored-document truth | Superseded by [0006](0006-v015-authority.md) |
| [0006](0006-v015-authority.md) | v0.15 deleted the derived index; live authority is ARCHITECTURE + CONTRACT | Accepted |
| [0007](0007-fresh-overwrite-for-playbooks-and-workflow.md) | Playbooks and WORKFLOW.md fresh-overwrite; three-way merge narrows to PROJECT.md and the AGENTS.md block | Accepted |
| [0008](0008-recorded-ownership-and-transactional-recovery.md) | Ownership is recorded, not inferred; stash restore is a transaction | Accepted |
| [0009](0009-guard-revision-selection-is-centralized.md) | One revision selector for every guard; completed plans are validated like active ones | Accepted |
| [0010](0010-local-failure-ledger.md) | A local failure ledger is maintainer-owned and optional for consumers | Accepted |
| [0011](0011-update-without-three-way-merge.md) | `update` drops three-way merge: hash-guarded AGENTS block, write-once PROJECT.md | Accepted |
| [0012](0012-five-section-plan-and-update-migration.md) | Plans have five sections; `update` migrates one nine-section active plan, Validation bytes unchanged | Accepted; decision 2 amended by [0014](0014-outcome-spec-normal-plans.md) |
| [0013](0013-installer-follows-directory-symlinks.md) | The installer follows symlinks at directory components; deferred to its own initiative | Accepted |
| [0014](0014-outcome-spec-normal-plans.md) | A normal plan is its Goal plus one phase that lock writes; `to-plan` is high-risk only | Accepted |
| [0015](0015-skill-owned-procedures-and-safe-retirement.md) | Skills own procedures; legacy retirement preserves consumer recovery | Accepted decision; candidate implemented, consumer verification pending |
| [0016](0016-lean-workflow-method-latitude.md) | Revise prompt-level method constraints; keep product and authority boundaries | Accepted decision; candidate pending B/C evaluation |
| [0017](0017-no-project-verification-skill.md) | No project-verification skill; the evidenced gaps are executable and already gated | Accepted |
| [0018](0018-opt-in-multi-model-review.md) | Multi-model review is an opt-in panel for `full` and high-risk checks | Accepted; policy only, tests pending |
| [0019](0019-arena-and-transcript-recall.md) | Implementation arena and transcript recall: defer both, with limits fixed now | Accepted |

## Writing a new one

Number it sequentially, name it `NNNN-kebab-case-title.md`, and add a row above. Keep the same five headings: Status, Context, Decision, Consequences, Authority. Every structural claim carries a `path:line` citation that resolves at merge time.
