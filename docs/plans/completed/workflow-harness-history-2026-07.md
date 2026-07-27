# Workflow Harness Legacy Lifecycle History — July 2026

## Record status and authority

This document is the completed historical record for the legacy parallel lifecycle-artifact system. It preserves initiative chronology, durable decisions, gate outcomes, and exact source coverage before those working-tree artifacts are removed.

It does **not** declare Harness Convergence Pass v3 complete. At summary time:

- v3 Phase 1 `universal-preflight` is done.
- v3 Phase 2 `root-layout` is done.
- v3 Phase 3 `one-plan-lifecycle` Waves 1–3 are complete and verified. Wave 3 is backed by trace `01KYGWA9M8GCHBAD0EB8XT0ZPQ` and an APPROVED semantic review.
- v3 Phase 3 Wave 4 is active; this history record is only its summary-writing step.
- v3 Phase 4 `docs-first-contract` is pending.

The canonical source for current execution state is `docs/plans/active/harness-convergence-pass-v3.md`. That file was untracked at summary time and is not a legacy deletion target.

## Chronological initiative record

### 1. Original workflow-harness import and pilots — 2026-07-17 to 2026-07-19

The initial dogfood import established the durable core on this repository's real history: `zharness init` plus `import` reproduced the prior current/entry phase, and replaying the ULID-ordered changesets into a fresh database produced byte-identical `resume --json` output. Validation also exposed honest historical incompleteness: pre-harness RUN/CHECK/HANDOFF markdown lacked the later ULID cross-link contract even though DB pointers were internally consistent.

The first independent second-agent pilot was a literal **NO-GO**. A cold Codex run read no `SKILL.md`, followed the scaffolded docs, implemented and tested the requested code, but simple-mode run registration failed because `runs.story_slug` required a story that simple mode did not create. The resulting `check record` also failed because the run had not been registered. The same pilot independently reconfirmed broken root `AGENTS.md` relative links. These were recorded as harness defects rather than hidden or misattributed to the agent.

After the #38/#39/#40 fix cycles, the final isolated second-agent pilot was a literal **GO**: an auth-only temporary HOME, no skill exposure, no harness coaching, a full `intake -> story -> run -> trace -> check` chain, 6/6 product tests passing, `zharness validate --json` returning `valid: true`, empty resume drift, and empty audit pointer drift. One nonblocking multi-file `artifact_path` audit finding remained and was filed separately.

Raw retained evidence:

- `docs/workflow-harness/pilot-evidence/2026-07-17-lab-skills-import.md`
- `docs/workflow-harness/pilot-evidence/2026-07-19-second-agent-pilot.md`
- `docs/workflow-harness/pilot-evidence/2026-07-19-agent-pilot-final.md`

### 2. Clean bootstrap and reinstall — 2026-07-20

The clean-bootstrap initiative intentionally removed the prior project `.kit` history, refreshed all eight workflow skills from the synchronized repository, preserved the local `workflow-core` customization, installed `zharness v0.4.1`, and initialized an empty schema-v2 harness without importing old state. Skill hashes, shared links, generated playbooks, CLI version, clean state/resume, and audit output were verified. The expected missing-SPEC result on a deliberately empty pre-intake harness was documented rather than treated as unexplained drift.

Gate: **APPROVED**, check `01KXYV4Z69YAJMPQQ034GSXZVD`.

### 3. README workflow refresh — 2026-07-20

The public workflow documentation was refreshed around the then-current four-layer model and skill-facing lifecycle. The README, canonical HTML flowchart, matching SVG, and migration guidance covered full, simple, resume, and legacy-adoption paths without teaching retired `plan`/`cook` names as the primary UX. Markup parity, links, stale terms, visual budgets, Go tests, build/vet, validation, audit, and scope were verified.

Gate: **APPROVED**, check `01KXYXFABNHPX05J8G8F2C8NNJ`.

### 4. Harness architecture audit — 2026-07-21

The architecture audit concluded that the system's append-only replay core, thin triggers, and execution discipline were worth retaining, but its platform-shaped mass was not justified for a single-user harness. Its central recommendation was to **evolve by subtraction, not rebuild**:

- retain changeset replay, traceability, verify-per-task discipline, and the lane×proof gate;
- move durable mutations from LLM-authored JSONL into CLI commands;
- remove dead entities and commands;
- remove score/entropy proxies that measured string length or finding counts rather than evidence;
- single-source playbooks;
- reduce token and maintenance cost without weakening proof.

Source: `.kit/reports/audit/20260721-harness-architecture-audit.md`.

### 5. Harness Subtraction Pass — 2026-07-21 to 2026-07-22

The subtraction pass executed four linear phases:

1. **write-boundary** — added `zharness run create`, made `check record` own `latest_check_id`, and removed playbook instructions to hand-author durable changesets. Gate: **APPROVED**, check `01KY1CD8MSWR9FYR9EYVXG1V3W`.
2. **dead-surface-removal** — removed unconsumed `decision`, `backlog`, `tool`, `propose`, and `score-context` command/application/domain/schema surface while preserving replay/import behavior. Gate: **APPROVED**, check `01KY42FNJ8GR9EZN6RXGC858NV`.
3. **scoring-removal** — removed `score-trace` and `entropy_score` while leaving the lane×proof validation matrix unchanged as the meaningful gate. Gate: **APPROVED**, check `01KY44PFE6KB8KMDNFVMNH3B33`.
4. **single-source-playbooks** — made the Go embed canonical and added a projection-drift test for `.kit/docs/playbooks`. Gate: **APPROVED**, check `01KY4BR73A4ZFJXQ4MQP6AE5M1`.

A separate post-write-boundary state-readiness audit returned **APPROVE WITH REQUESTS**, check `01KY2DVZDEFCNHCFEQSGBYPRCB`. Source code was re-verified green, but the local database had not replayed 23 pulled changesets, the installed binary predated `run create`, and local scaffolded playbooks were stale. Those were operational synchronization requests, not reversals of the write-boundary code gate.

The pass closed with all four implementation phases gated and the raw details retained in Git history. Its durable result was subtraction without losing replay or execution discipline.

### 6. Slim Playbooks S1–S3 — 2026-07-21 to 2026-07-22

The slim-playbooks track moved deterministic prose into tested CLI behavior to reduce per-run token load:

- **S1 — CLI scaffold:** extracted SPEC/RUN/CHECK/HANDOFF skeletons into embedded templates and added `zharness scaffold`. Gate: **APPROVE WITH REQUESTS**, report ID `01KY3PA3RM6FJQ0B97W7TYXY15`; requests covered a stale installed binary and an intentionally unregistered ad-hoc audit report, not scaffold correctness.
- **S2 — `zharness next`:** moved mode/phase routing tables into a tested command while leaving git-dependent contract-drift and plan-lifecycle judgment agent-side. Gate: **APPROVE WITH REQUESTS**, report ID `01KY3VHZ3K4BSM893HJ8P86FWE`.
- **S3 — `resume --facts`:** moved recap rendering and its formatting/validation rules into the CLI and slimmed `watzup`. Gate: **APPROVED**, report frontmatter ID `01KY3WXHS72E066EEJDFP75Y3R-check-s3` (the report's displayed run label is `check-20260722-slim-playbooks-s3`).

These stages were an informal track without registered RUN rows; the reports state that limitation explicitly.

### 7. Harness Convergence Pass v3 — 2026-07-26 to 2026-07-27

The v3 initiative converges on repository authority, one root DB, mandatory preflight, and one evolving plan without removing typed lifecycle rows.

#### Phase 1 — universal-preflight

The first gate returned **REQUEST CHANGES**, check `01KYF6VS9TCHEZH41X91ZKH9DK`, because `work --mode auto` could reduce instead of block when a durable plan existed without a DB, and missing managed docs still produced a nonexistent playbook path. Both routing defects were fixed. The repeated gate was **APPROVED**, check `01KYF71N4RMWYGTSJTD4BWC740`, with all eight stages using common read-only preflight and zero-write proof.

Status: **done**.

#### Phase 2 — root-layout

The phase activated one ignored root `harness.db`, preserved `.kit/changesets`, projected root managed docs, added installed-hash conflict protection, and proved dry-run migration, replay parity, rollback safety, dogfood migration, and fresh-clone reconstruction. Legacy lifecycle markdown was deliberately retained pending Phase 3 coverage.

Gate: **APPROVED**, check `01KYF9NC8RXVCNYM8M56YQX0RF`.

Status: **done**.

#### Phase 3 — one-plan-lifecycle

Waves 1–2 added the one-plan schema/scaffold contract, removed report-path requirements from lifecycle readers, completed public `planned -> in-progress -> checked -> done` transitions, and proved bounded work makes no lifecycle writes.

Their gates followed an honest request/fix/review cycle:

1. `20260727-0228-one-plan-lifecycle.md`, ID `01KYGPE7H8HJ6DFM6491GWS7PX`: **REQUEST CHANGES** after the full suite found one stale lifecycle assertion; no check record was written.
2. `20260727-0237-one-plan-lifecycle.md`, ID `01KYGPXS1T37ARFBVK04AQG3KM`: **REQUEST CHANGES** despite green build/test/vet because the real-binary fixture bypassed public lifecycle commands, `STATE.md` still described file-based semantics, and ULID validation accepted overflow encodings.
3. `20260727-0248-one-plan-lifecycle.md`, ID `01KYGQK2NVCY4FT34AF7AWTVHA`: **APPROVED** after the fixture, docs, and strict ULID fixes; audit was clean and lifecycle status was `checked`.

Wave 3 rewrote the six stage playbooks around one evolving plan, established Progress as the task-status source, preserved immutable task definitions, kept review/bounded routes zero-write, and added two-dependent-phase real-binary lifecycle proof. Targeted, expanded, and full Go suites passed; semantic review was **APPROVED** after two review-fix cycles. Trace: `01KYGWA9M8GCHBAD0EB8XT0ZPQ`. No separate durable Wave 3 check ID was recorded; the full Phase 3 gate remains after Wave 4.

Wave 4 was active at summary time. This file completes only the legacy-history summary step. Legacy removal, file-independent replay/resume/validate proof, and the full Phase 3 gate were still pending.

#### Phase 4 — docs-first-contract

Pending. Its planned scope is the final concise authority/work-shape contract and public documentation rewrite. It must not be described as complete by this history record.

## Durable decisions retained

1. **The CLI owns durable mutations.** Agents invoke intent-named commands; they do not hand-author lifecycle changeset JSONL.
2. **`.kit/changesets` is append-only replay provenance.** Existing changesets are never rewritten or compacted as part of these initiatives.
3. **There is exactly one writable database.** The active layout uses one ignored root `harness.db`; no `.harness`, second DB, or baseline snapshot DB is introduced.
4. **Repository truth outranks the ledger.** Repository docs, code, tests, and observable runtime behavior are authority for intent and proof. SQLite is a durable execution ledger and recovery index.
5. **One evolving plan replaces parallel lifecycle markdown.** Durable initiatives use `docs/plans/active/{slug}.md`, moved to `docs/plans/completed/{slug}.md` after final validation, while typed intake/story/run/trace/check/handoff/intervention rows remain queryable.
6. **Bounded work is zero-write.** It may invoke preflight, but creates no lifecycle row, plan, report, changeset, or workflow markdown artifact.
7. **Managed docs protect local edits.** Installed versions and SHA-256 hashes are stored in `managed_docs`; untouched files may refresh, while local/upstream conflicts stage recovery content instead of overwriting consumer edits.
8. **Score and entropy proxies are removed; lane×proof remains.** Required evidence classes, not string-length tiers or weighted finding counts, determine gate completeness.
9. **Playbooks are single-sourced and projected.** The embedded canonical text is guarded against drift in managed copies.
10. **Raw provenance remains available.** `docs/workflow-harness/` retains selected import, migration, gap, and pilot evidence; Git history retains the complete bodies and evolution of removed legacy artifacts.

## Gate chronology

| Date | Initiative / phase | Verdict | Durable ID or evidence |
|---|---|---|---|
| 2026-07-17 | Real-history import/replay | PASS for state import and byte-identical replay; validation exposed historical markdown incompleteness | retained pilot evidence; no CHECK ID |
| 2026-07-19 | Initial independent second-agent pilot | NO-GO on literal R9 because simple-mode RUN FK registration failed | retained pilot evidence; issue #38 |
| 2026-07-19 | Final isolated second-agent pilot | GO; `validate` returned `valid:true` | retained pilot evidence; lifecycle CHECK `01KXX77N09FEXQ8C0KYKB5YAAC` was `APPROVE_WITH_REQUESTS` for the pilot product chain |
| 2026-07-20 | Clean bootstrap/reinstall | APPROVED | `01KXYV4Z69YAJMPQQ034GSXZVD` |
| 2026-07-20 | README workflow refresh | APPROVED | `01KXYXFABNHPX05J8G8F2C8NNJ` |
| 2026-07-21 | Harness architecture audit | Evolve by subtraction; preserve replay and execution discipline | audit report; no CHECK ID |
| 2026-07-21 | Harness Subtraction: write-boundary | APPROVED | `01KY1CD8MSWR9FYR9EYVXG1V3W` |
| 2026-07-21 | State-readiness audit | APPROVE WITH REQUESTS | `01KY2DVZDEFCNHCFEQSGBYPRCB` |
| 2026-07-22 | Harness Subtraction: dead-surface-removal | APPROVED | `01KY42FNJ8GR9EZN6RXGC858NV` |
| 2026-07-22 | Harness Subtraction: scoring-removal | APPROVED | `01KY44PFE6KB8KMDNFVMNH3B33` |
| 2026-07-22 | Harness Subtraction: single-source-playbooks | APPROVED | `01KY4BR73A4ZFJXQ4MQP6AE5M1` |
| 2026-07-22 | Slim Playbooks S1 | APPROVE WITH REQUESTS | `01KY3PA3RM6FJQ0B97W7TYXY15` |
| 2026-07-22 | Slim Playbooks S2 | APPROVE WITH REQUESTS | `01KY3VHZ3K4BSM893HJ8P86FWE` |
| 2026-07-22 | Slim Playbooks S3 | APPROVED | `01KY3WXHS72E066EEJDFP75Y3R-check-s3` |
| 2026-07-26 | v3 universal-preflight initial gate | REQUEST CHANGES | `01KYF6VS9TCHEZH41X91ZKH9DK` |
| 2026-07-26 | v3 universal-preflight resolved gate | APPROVED | `01KYF71N4RMWYGTSJTD4BWC740` |
| 2026-07-26 | v3 root-layout | APPROVED | `01KYF9NC8RXVCNYM8M56YQX0RF` |
| 2026-07-27 02:28 | v3 one-plan Waves 1–2 first gate | REQUEST CHANGES; stale full-suite assertion | `01KYGPE7H8HJ6DFM6491GWS7PX` |
| 2026-07-27 02:37 | v3 one-plan Waves 1–2 second gate | REQUEST CHANGES; public lifecycle fixture/docs/ULID issues | `01KYGPXS1T37ARFBVK04AQG3KM` |
| 2026-07-27 02:48 | v3 one-plan Waves 1–2 resolved gate | APPROVED | `01KYGQK2NVCY4FT34AF7AWTVHA` |
| 2026-07-27 04:10 | v3 one-plan Wave 3 semantic review | APPROVED; full Phase 3 gate still pending | trace `01KYGWA9M8GCHBAD0EB8XT0ZPQ`; no CHECK ID |
| 2026-07-27 summary time | v3 one-plan Wave 4 | ACTIVE; no completion verdict | canonical active plan |
| pending | v3 docs-first-contract | PLANNED | no RUN or CHECK |

## Source coverage appendix

### Inventory snapshot

The exact working-tree inventory at summary time contains **62 tracked legacy deletion targets** plus **4 current untracked legacy evidence files**. One tracked run was modified. The canonical active v3 plan was also untracked, but it is current authority rather than a legacy deletion target.

Status-sensitive current evidence:

- `docs/plans/active/harness-convergence-pass-v3.md` — untracked; canonical current plan; retain.
- `.kit/runs/work/20260726-1328-one-plan-lifecycle.md` — tracked and modified.
- `.kit/runs/work/20260727-0308-one-plan-lifecycle.md` — untracked.
- `.kit/reports/check/20260727-0228-one-plan-lifecycle.md` — untracked.
- `.kit/reports/check/20260727-0237-one-plan-lifecycle.md` — untracked.
- `.kit/reports/check/20260727-0248-one-plan-lifecycle.md` — untracked.

### `.kit/planning/` — 24 tracked

1. `.kit/planning/ROADMAP.md`
2. `.kit/planning/SPEC.md`
3. `.kit/planning/archive/harness-subtraction-pass-2026-07-21/ROADMAP.md`
4. `.kit/planning/archive/harness-subtraction-pass-2026-07-21/SPEC.md`
5. `.kit/planning/archive/harness-subtraction-pass-2026-07-21/phases/dead-surface-removal/dead-surface-removal-CONTEXT.md`
6. `.kit/planning/archive/harness-subtraction-pass-2026-07-21/phases/dead-surface-removal/dead-surface-removal-PLAN.md`
7. `.kit/planning/archive/harness-subtraction-pass-2026-07-21/phases/scoring-removal/scoring-removal-CONTEXT.md`
8. `.kit/planning/archive/harness-subtraction-pass-2026-07-21/phases/scoring-removal/scoring-removal-PLAN.md`
9. `.kit/planning/archive/harness-subtraction-pass-2026-07-21/phases/single-source-playbooks/single-source-playbooks-CONTEXT.md`
10. `.kit/planning/archive/harness-subtraction-pass-2026-07-21/phases/single-source-playbooks/single-source-playbooks-PLAN.md`
11. `.kit/planning/archive/harness-subtraction-pass-2026-07-21/phases/write-boundary/write-boundary-CONTEXT.md`
12. `.kit/planning/archive/harness-subtraction-pass-2026-07-21/phases/write-boundary/write-boundary-PLAN.md`
13. `.kit/planning/archive/readme-workflow-refresh-2026-07-20/ROADMAP.md`
14. `.kit/planning/archive/readme-workflow-refresh-2026-07-20/SPEC.md`
15. `.kit/planning/archive/readme-workflow-refresh-2026-07-20/phases/readme-workflow-refresh/readme-workflow-refresh-CONTEXT.md`
16. `.kit/planning/archive/readme-workflow-refresh-2026-07-20/phases/readme-workflow-refresh/readme-workflow-refresh-PLAN.md`
17. `.kit/planning/phases/docs-first-contract/docs-first-contract-CONTEXT.md`
18. `.kit/planning/phases/docs-first-contract/docs-first-contract-PLAN.md`
19. `.kit/planning/phases/one-plan-lifecycle/one-plan-lifecycle-CONTEXT.md`
20. `.kit/planning/phases/one-plan-lifecycle/one-plan-lifecycle-PLAN.md`
21. `.kit/planning/phases/root-layout/root-layout-CONTEXT.md`
22. `.kit/planning/phases/root-layout/root-layout-PLAN.md`
23. `.kit/planning/phases/universal-preflight/universal-preflight-CONTEXT.md`
24. `.kit/planning/phases/universal-preflight/universal-preflight-PLAN.md`

### `.kit/plans/` — 5 tracked

1. `.kit/plans/2026-07-20-clean-bootstrap-reinstall/PLAN.md`
2. `.kit/plans/2026-07-20-readme-workflow-flow/PLAN.md`
3. `.kit/plans/2026-07-21-slim-playbooks/PLAN.md`
4. `.kit/plans/2026-07-22-slim-playbooks-s2/PLAN.md`
5. `.kit/plans/2026-07-22-slim-playbooks-s3/PLAN.md`

### `.kit/runs/` — 9 tracked plus 1 current untracked

Tracked:

1. `.kit/runs/work/20260720-1032-clean-bootstrap-reinstall.md`
2. `.kit/runs/work/20260720-1126-readme-workflow-refresh.md`
3. `.kit/runs/work/20260721-1027-write-boundary.md`
4. `.kit/runs/work/20260722-1135-dead-surface-removal.md`
5. `.kit/runs/work/20260722-1300-scoring-removal.md`
6. `.kit/runs/work/20260722-1424-single-source-playbooks.md`
7. `.kit/runs/work/20260726-1204-universal-preflight.md`
8. `.kit/runs/work/20260726-1242-root-layout.md`
9. `.kit/runs/work/20260726-1328-one-plan-lifecycle.md` — modified at summary time.

Current untracked:

10. `.kit/runs/work/20260727-0308-one-plan-lifecycle.md`

### `.kit/reports/` — 14 tracked plus 3 current untracked checks

Tracked:

1. `.kit/reports/audit/20260721-harness-architecture-audit.md`
2. `.kit/reports/check/2026-07-22-slim-playbooks-s3-gate.md`
3. `.kit/reports/check/20260720-1104-clean-bootstrap-reinstall.md`
4. `.kit/reports/check/20260720-1145-readme-workflow-refresh.md`
5. `.kit/reports/check/20260721-1044-write-boundary.md`
6. `.kit/reports/check/20260721-2029-state-readiness-audit.md`
7. `.kit/reports/check/20260722-1230-dead-surface-removal.md`
8. `.kit/reports/check/20260722-1330-scoring-removal.md`
9. `.kit/reports/check/20260722-1445-single-source-playbooks.md`
10. `.kit/reports/check/20260722-slim-playbooks-s1-w3-gate.md`
11. `.kit/reports/check/20260722-slim-playbooks-s2-gate.md`
12. `.kit/reports/check/20260726-1236-universal-preflight.md`
13. `.kit/reports/check/20260726-1240-universal-preflight.md`
14. `.kit/reports/check/20260726-1325-root-layout.md`

Current untracked:

15. `.kit/reports/check/20260727-0228-one-plan-lifecycle.md` — ID `01KYGPE7H8HJ6DFM6491GWS7PX`, REQUEST CHANGES.
16. `.kit/reports/check/20260727-0237-one-plan-lifecycle.md` — ID `01KYGPXS1T37ARFBVK04AQG3KM`, REQUEST CHANGES.
17. `.kit/reports/check/20260727-0248-one-plan-lifecycle.md` — ID `01KYGQK2NVCY4FT34AF7AWTVHA`, APPROVED.

### `.kit/HANDOFF.md` — 1 tracked

1. `.kit/HANDOFF.md`

The handoff is a partial-harness checkpoint from 2026-07-26. It accurately distinguished approved Phases 1–2 from then-unverified Phase 3 work; later active-plan Progress and Validation entries supersede its execution snapshot.

### `.kit/docs/` — 9 tracked

1. `.kit/docs/AGENTS.md`
2. `.kit/docs/AUTHORITY.md`
3. `.kit/docs/CONTEXT_RULES.md`
4. `.kit/docs/playbooks/brainstorm.md`
5. `.kit/docs/playbooks/check.md`
6. `.kit/docs/playbooks/handoff.md`
7. `.kit/docs/playbooks/to-plan.md`
8. `.kit/docs/playbooks/watzup.md`
9. `.kit/docs/playbooks/work.md`

These are legacy managed projections. Their historical decisions and operating changes are represented above; canonical/current managed docs live under root `docs/` and embedded CLI sources.

### Protected `.kit/changesets` — not deletion targets

All changesets are protected, whether tracked or currently untracked. At summary time there were **68 tracked changesets** and **5 untracked changesets**, for **73 total**. They remain append-only replay provenance and are excluded from legacy deletion.

Current untracked protected changesets:

1. `.kit/changesets/01KYGP9S4X3QSBWHXN16KNJJW5.changeset.jsonl`
2. `.kit/changesets/01KYGPXS1T37ARFBVK05CJW2Q1.changeset.jsonl`
3. `.kit/changesets/01KYGQK2NVCY4FT34AF8GA70X3.changeset.jsonl`
4. `.kit/changesets/01KYGRQGV0ZS281GCB82N3ETDT.changeset.jsonl`
5. `.kit/changesets/01KYGWA9M8GCHBAD0EB9EFXWCS.changeset.jsonl`

### Retained `docs/workflow-harness/` provenance — not deletion targets

1. `docs/workflow-harness/gap-matrix.md`
2. `docs/workflow-harness/migration.md`
3. `docs/workflow-harness/pilot-evidence/2026-07-17-lab-skills-import.md`
4. `docs/workflow-harness/pilot-evidence/2026-07-19-agent-pilot-final.md`
5. `docs/workflow-harness/pilot-evidence/2026-07-19-second-agent-pilot.md`

These files remain the selected raw provenance set. Git history remains the raw archive for every legacy artifact body summarized and indexed here.

## Coverage reconciliation

- `.kit/planning/`: 24 tracked
- `.kit/plans/`: 5 tracked
- `.kit/runs/`: 9 tracked + 1 untracked
- `.kit/reports/`: 14 tracked + 3 untracked
- `.kit/HANDOFF.md`: 1 tracked
- `.kit/docs/`: 9 tracked
- **Legacy deletion-target total:** 62 tracked + 4 untracked
- **Protected state:** 68 tracked + 5 untracked `.kit/changesets` files; none are deletion targets
- **Retained raw provenance:** 5 tracked `docs/workflow-harness/` files; none are deletion targets

This index demonstrates representation of every legacy lifecycle source present in `git ls-files` and every modified/untracked current evidence path reported by status at summary time, without duplicating entire artifact bodies.
