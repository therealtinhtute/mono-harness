# Harness Engineering Gap Audit — zharness v0.15

**Date:** 2026-08-30
**Lens:** rari (@0xwhrrari), *Harness Engineering: How to Build AI Agents That Don't Fall Apart* (X, 2026-08-29, 101K views).
**Source URL:** https://x.com/0xwhrrari/status/2093685107534000560
**Fetch:** Pass 1 (`fetch.sh --use-proxy` on the status URL) returned a truncated local extract; the X Article URL 404’d. Pass 2 (`curl -sL https://defuddle.md/https://x.com/0xwhrrari/status/2093685107534000560`) returned the full article: 1814 words, YAML frontmatter, three diagrams (`HQ1mN7uXgAA5EmN.jpg` cover, `HQ1oAGxWsAAttaT.jpg` seven jobs, `HQ1pxJvWAAAXXOY.jpg` failure loop), inline schemas, and the 12-item checklist verbatim. Cover/job/failure diagrams were read as images. CTA/subscribe text is ignored.
**Scope:** current v0.15 slim — binary = `install` / `update` / `uninstall`; lifecycle = committed markdown + repo scripts + two fail-closed git-hook guards. Prior audits under `docs/audit/` that still describe `preflight`, `harness.db`, `trace add`, or `check record` are dated 0.14 records. They are evidence of residue, not a description of the running system.
**Method:** map the article’s seven jobs, dual-encoding rule, loop (3 attempts then human), change-receipt schema, Level 0–3 ladder, failure-class table, and the 12-item checklist onto live files. Every finding cites a path. No CLI lifecycle command was invoked; v0.15 no longer has one.
**Non-goal:** do not resurrect the 0.14 control plane. The article’s own constraint is “the harness should be smaller than the failure surface it controls.” Do not add a GRAPH/coordination layer; the article lists it as a later shift, not a starting requirement.

---

## 1. Headline

v0.15 already matches the article’s architecture. The remaining failures are map pollution and prose-only invariants — not missing intelligence, and not a missing database.

The article’s thesis, grounded in the fetched text:

> The problem is not always the intelligence. The problem is the environment around it. That environment is the harness.
>
> When an agent fails repeatedly, stop editing adjectives in the prompt. Inspect the system around the model.
>
> The stronger pattern is to encode the important rule twice. First as guidance the agent can understand. Then as a mechanical check the agent cannot bypass.

zharness already does the expensive half of that sentence: proof re-execution and the high-risk independent-judge rule live in `scripts/install-git-hooks.sh`, not in a prompt. What it does not do is keep the map true after deleting the old control plane, or dual-encode the invariants that used to be `ResolveActivePlan` / `zharness validate`.

Three findings that matter, in order:

| # | Finding | Article job | Fix risk |
|---|---|---|---|
| **H1** | The live map still describes a deleted control plane. Agents will optimize ghosts (`preflight`, `harness.db`, `check_id` ULID, `latest_run_id`). | 2. Map | Low — delete residue; do not add verbs |
| **H2** | “At most one active plan” dropped from code to prose when v0.15 deleted `ResolveActivePlan`. Consumer D1 can recur with no mechanical stop. | 1. Contract + dual-encode | Low — hook/CI, not a lifecycle CLI |
| **H3** | Same agent invents the plan, executes it, and gates it on every non-`high-risk` phase (`work.md` step 11, `judge: same-session`). The hook only blocks that on `lane: high-risk`. | 6. Permissions outside the model | Medium — this is a calibrated token-cache tradeoff, not an accident |

Do **not** spend the next initiative on: bringing back SQLite, `preflight`, `trace add`, `query plan`, or a JSONL CLI log. Those optimize a binary that no longer runs the loop. The article’s compounding move is “failure upgrades the system”; the system to upgrade is the map and the two remaining unenforced invariants.

---

## 2. What the article actually claims

Full defuddle body plus the three diagrams. CTA/subscribe text is ignored. This is the contract this audit scores against.

**Cover diagram** (`HQ1mN7uXgAA5EmN.jpg`): `REQUEST (goal · inputs · risk)` enters `THE AGENT HARNESS`. Left of the model: CONTRACT (scope · success · constraints), CONTEXT (maps · rules · evidence), STATE (artifacts · decisions · memory). Right of the model: TOOLS (sandbox · browser · shell), PERMISSIONS (policy · approval · isolation), EVIDENCE (tests · traces · recovery). Output: `RESULT (verified · traceable)`. Caption: **MODEL REASONS · HARNESS CONTROLS · TOOLS ACT**.

**Seven-jobs diagram** (`HQ1oAGxWsAAttaT.jpg`), verbs in the article’s own labels:

| # | Job | Verb | Payload |
|---|---|---|---|
| 01 | CONTRACT | DEFINE | goal · limits · done |
| 02 | CONTEXT | SELECT | maps · rules · facts |
| 03 | TOOLS | ACT | sandbox · browser · shell |
| 04 | STATE | REMEMBER | artifacts · decisions |
| 05 | SENSORS | VERIFY | tests · logs · screenshots |
| 06 | POLICY | AUTHORIZE | scope · approval · budget |
| 07 | TRACES | EXPLAIN | events · cost · rollback |

Footer of that diagram: **ONE MODEL CALL IS AN EVENT · THE HARNESS IS THE SYSTEM.** Control sentence: **LEGIBLE BEFORE ACTION → CONTROLLED DURING ACTION → AUDITABLE AFTER ACTION.**

**Inline schemas the first extract dropped**

- Contract object: `{goal, inputs, output, constraints, done_when}`.
- Map tree: `AGENTS.md → architecture / testing / product / security / task-specific guides`.
- Tool default table: `READ FILES` allowed by default; `RUN TESTS` allowed inside sandbox; `WRITE FILES` allowed inside workspace; `ACCESS NETWORK` scoped by task; `DEPLOY` and `DELETE DATA` require approval.
- Durable state object: `{task_id, current_step, artifacts, decisions, failures, pending}`.
- Sensor matrix: CODE → tests+types+lint; UI → render+screenshot; RESEARCH → source+contradiction; DATA → schema+range+freshness.
- Policy pipeline: `MODEL SUGGESTS → POLICY CHECKS → TOOL EXECUTES`.
- Trace list: request, selected context, tool calls, state changes, verification results, retries, cost, final artifact, rollback point.
- Dual-encode example: GUIDE “UI code may not query the database directly” / CHECK “lint fails when UI imports the repository layer”.
- Loop: `attempt <= 3`; on fail push `evidence.gap` and set `repair`; after bound `requestHumanReview(state)`.
- Change receipt: `{context_sources, policy_version, model_route, tools_used, tests.{passed,failed}, human_corrections, retries, cost_usd, accepted_artifact, rollback_point}`.
- Levels: **0** prompt+model; **1** project guide+tools; **2** structured state+tests+bounded loop; **3** permissions+traces+recovery+human gates.
- Stack: `PROMPT → instruction; CONTEXT → working view; HARNESS → operating system; LOOP → local improvement; GRAPH → coordination`.

**Failure diagram** (`HQ1pxJvWAAAXXOY.jpg`): RUN (produce work) → OBSERVE (capture evidence) → CLASSIFY (name the gap) → REPAIR (patch locally) → VERIFY (test again) → ACCEPT (ship with receipt). Fail arc CLASSIFY↔VERIFY: “RETURN THE EXACT GAP · RETRY WITH A BOUND”. Downward from CLASSIFY: **PERMANENT HARNESS UPDATE** — add a guide · add a test · improve a tool · tighten a permission · store new state. Caption: the patch fixes one run; the harness change improves every run after it.

**Failure-class table** (verbatim mapping in the article):

| Class | Harness upgrade |
|---|---|
| MISSING CONTEXT | add a map or retrieval rule |
| WRONG TOOL | improve tool description or routing |
| BAD OUTPUT | add a validator or stronger contract |
| REPEATED LOOP | add a retry cap and escalation |
| UNSAFE ACTION | add a permission gate |
| LOST DECISION | store it in durable state |
| UNKNOWN FAILURE | add tracing and evidence capture |

Cited authorities in the post (links present in the full fetch, not re-read here): Dario Amodei on Claude Code needing a harness; OpenAI, [Harness engineering: leveraging Codex in an agent-first world](https://openai.com/index/harness-engineering/); Anthropic, [Harness design for long-running application development](https://www.anthropic.com/engineering/harness-design-long-running-apps).

---

## 3. Scorecard — seven jobs vs v0.15

Legend: **pass** = dual-encoded or structurally true; **partial** = guidance exists, mechanical half missing or residue contradicts it; **fail** = the job is not closed.

| Job | Grade | What is already true | What is not |
|---|---|---|---|
| 1. Contract | **partial** | `brainstorm` locks Outcome / Authority / Requirements / Non-goals; requirements are numbered and falsifiable; `to-plan` refuses to invent verification later; `BLOCKED_CONTRACT_DRIFT` stops silent scope expansion. | Bounded mode has no contract object. There is no schema check that a plan’s Outcome was met. Two active plans are a playbook `stop`, not a hook reject. |
| 2. Map | **fail** | `AGENTS.md` ZHARNESS block is a small root guide. `docs/WORKFLOW.md` routes by stage. Spine `SKILL.md` files are thin triggers. Progressive disclosure is real. | Live playbooks, ADRs 0001–0003/0005, and `skills/workflow/README.md` still speak 0.14. That is a giant mixed-truth manual, which is the article’s exact anti-pattern. |
| 3. Tools | **partial** | Binary surface is three verbs (`cli/docs/CONTRACT.md`). Hooks and `scripts/verify-doc-links.sh` / `scripts/test-guards.sh` have explicit failure. `embedded_test.go` forbids deleted CLI strings from returning in playbooks — and currently loses, because sibling residue (`lifecycle ledger`, `DB-mirroring`, `check_id: ULID`) is not on that kill list. | Agent-facing tools for the loop are “read markdown + run bash”. Failure states like “two active plans” have no structured envelope since `Stop{Code:ambiguous}` died with the CLI. |
| 4. Durable state | **pass** | Plans under `docs/plans/{active,completed}/` with append-only Progress / Decisions / Validation. Memory is `docs/memory/{id}.md`, grep-first, opt-in writes (`work.md` Memory conventions; ADR 0003’s *intent* still holds even though `zharness memory` is gone). Git is history. | Conversation is still working memory inside a session. `watzup` is the recap, but it is optional and still tells the agent to consult a “lifecycle ledger when the binary happens to exist.” |
| 5. Sensors | **pass** | `to-plan` step 6 writes checks before execution. `work` step 6 runs the exact command; one retry then `BLOCKED_VERIFICATION`. Pre-commit re-executes nested proof commands of new `APPROVED` / `APPROVE_WITH_REQUESTS` entries (`scripts/install-git-hooks.sh` ZGUARD-CORE). CI re-runs the same block. An approvable verdict with zero proofs is malformed, not waved through. | `docs/evals/failures.md` is specified as optional in `check.md` step 4 and does not exist. Class-of-failure → infrastructure is therefore opt-in folklore, not a sensor. |
| 6. Permissions outside the model | **partial** | High-risk + `judge: same-session` is rejected by the hook from staged bytes. Host runtime (Claude/Pi/Codex) owns tool allow/deny; zharness correctly does not pretend to. Uninstall never deletes consumer-modified bytes. | On `lane: normal` (the default), the same session plans, edits, and writes the gate (`work.md` step 11). The article’s sentence “Do not ask the same probabilistic system to invent the plan, approve the risk, and execute the side effect” is exactly this path. It is a documented token-cache choice, not a bug — but it is still a permission-layer gap. |
| 7. Traces + local recovery | **partial** | Progress / Decisions / Validation + `git log` are the trail. `handoff` reconstructs from tails, not from a retelling. Compaction invalidates earlier-read anchors (playbook precondition). macos-guard-portability is a live example of a failure becoming a guard fix. | No compact change receipt. `check.md` still asks for `check_id: ULID` and “the mirrored check row (when the binary exists)”. Those fields cannot be produced. A green Validation entry can still hide a process that cited machine-local proofs — caught this week only because CI re-executed them (`docs/plans/completed/macos-guard-portability.md` Decisions, 2026-08-28 correction). |

**Level on the article’s ladder.** Durable `work`/`check`/`handoff` is Level 2 (structured state + tests + bounded loop) with two Level-3 sensors bolted on (proof re-execution, high-risk independent judge). Bounded/simple is Level 1. The live map still talks as if a Level-3 CLI control plane exists. That mismatch is H1.

**LEGIBLE BEFORE → CONTROLLED DURING → AUDITABLE AFTER**

| When | Article | zharness |
|---|---|---|
| Before | contract + selected context | durable lock yes; map polluted; bounded mode has no `{goal, inputs, output, constraints, done_when}` |
| During | sandbox tools + policy check + budget | host sandbox; zharness does not run `POLICY CHECKS` before tool exec; budget is host-side; one prose retry |
| After | traces, cost, rollback, receipt | git + Validation; no cost, no `rollback_point`, no compact receipt |

**Brain / hands / history.** v0.15 is closer to the article than 0.14 was. The reasoning engine is the host agent. Hands are git + the filesystem. History is committed markdown. The binary is not the brain, not the memory database, and not the audit log. That split is the point of slim, and it should stay.

**Smallest harness that closes the loop.** Bounded/simple mode (zero lifecycle writes; git diff + captured proof) is the short-task harness. Full mode is the six-hour harness. The article endorses that split (Level 0–1 vs Level 2–3). The remaining defect is that the full-mode *map* is larger than the failure surface it now controls, because it still carries 0.14 bookkeeping.

---

## 4. Dual-encoding matrix

The article’s load-bearing rule: important constraints are written twice — once for the model, once as a check the model cannot talk past.

| Invariant | Guidance (prose) | Mechanical check | Status |
|---|---|---|---|
| Proof commands of a clean Validation entry actually ran | `check.md` steps 8–9 | Hook re-executes nested bullets from staged bytes; CI extracts the same ZGUARD-CORE block | **encoded twice** |
| `judge: same-session` forbidden on `lane: high-risk` | `check.md` step 7 | Hook reads plan frontmatter `lane:` | **encoded twice** |
| Doc citations resolve | playbooks / ARCHITECTURE | `scripts/verify-doc-links.sh` | **encoded twice** |
| Deleted lifecycle CLI strings stay deleted | ARCHITECTURE historical note | `cli/internal/embedded/embedded_test.go` forbidden list (`zharness preflight`, `zharness query`, `check record`, …) | **encoded twice, incomplete kill list** |
| At most one non-empty file under `docs/plans/active/` | `work.md:12`, `watzup.md:9`, `brainstorm.md` lock step 7 | Was `ResolveActivePlan` (`docs/decisions/0002-single-active-plan-resolver.md`). File `cli/internal/application/plan_resolve.go` is gone. Hook does not count active plans. | **prose only — H2** |
| Slice the plan, never whole-file | `work.md:32`, `watzup.md:15` | None. An agent that `Read`s the file pays the quadratic cost the 0.14 ceremony audit measured, with no CLI to slice it either. | **prose only** |
| Bounded/simple creates no lifecycle markdown | zero-write rule in four playbooks | None. An agent can still append to the active plan during a typo fix. | **prose only** |
| Memory bodies contain no secrets | `work.md` redaction rule | None. `docs/memory/` is committed markdown with no scanner in the hook. | **prose only** |
| Outcome/requirements are falsifiable | `brainstorm.md` self-review | No plan-schema validator. `zharness validate` is deleted. | **prose only** |
| One retry then `BLOCKED_VERIFICATION` | `work.md` step 6 | None. The loop is “keep trying” unless the agent obeys. | **prose only** |
| Phase definitions immutable after `to-plan` | `to-plan.md` Planning Rules | None. | **prose only** |

The two fail-closed guards are the article done right. Everything else in the left column is still “the agent reads it, then eventually ignores one.”

---

## 5. Findings

### [!] H1 — Live map describes a deleted control plane

**Why.** The article’s job 2 is “a small root guide that tells the agent where to look.” A map that names tools that 404 is worse than a short map: the agent spends turns reconstructing a world that v0.15 deleted, then either invents ceremony or ignores the playbook.

**Evidence (live files, not historical audits):**

- `docs/playbooks/watzup.md:5` — “and — when the binary happens to exist — the lifecycle ledger.” v0.15 binary has no ledger.
- `docs/playbooks/work.md:32` — “phase-status disagreement between the DB-mirroring frontmatter and the phase blocks.”
- `docs/playbooks/work.md:34` — “`latest_run_id` stays `none` unless an entry on record supplies one.”
- `docs/playbooks/to-plan.md:39` — “mirror their DB transitions while the ledger exists, marking them in this file alone once it is gone.”
- `docs/playbooks/check.md:42-43,64,70` — “record the returned check ID if one was issued”; “to match the DB row whenever the binary exists”; output field `check_id: ULID | not-recorded`; exit condition “the mirrored check row (when the binary exists) was recorded.”
- `skills/workflow/README.md:3-5` — still “a harness-backed runtime” whose change is “a durable, machine-recorded, replayable trail instead of relying on markdown pointers alone.” Lines 9 and 11 then say the opposite (harness gone, markdown is the record). Lines 93–105 keep a GO verdict on `zharness init && zharness import && zharness query state --json`.
- `docs/decisions/0001-markdown-as-source-of-truth.md` — Accepted ADR whose decision text is “`harness.db` is a derived index.”
- `docs/decisions/0002-single-active-plan-resolver.md` — Accepted ADR citing `cli/internal/application/plan_resolve.go:73`, a path that does not exist.
- `docs/decisions/0003-durable-memory-not-wired-into-playbooks.md` — Accepted ADR whose consequence is “`zharness memory` works fully.”
- `docs/decisions/0005-authored-documentation-boundary.md` — Accepted ADR whose decision is “`zharness audit` guards authored documentation.”
- `cli/docs/CONTRACT.md:8-9` — “Until then the binary registers no subcommands — `zharness --help` shows root usage only.” Contradicts the table three lines later (`install` / `update` / `uninstall`).

`cli/internal/embedded/embedded_test.go` already forbids exact command strings (`zharness preflight`, `zharness query`, …). It does not forbid the *concepts* those commands left behind (`lifecycle ledger`, `DB-mirroring`, `check_id: ULID`). The kill list won the battle and lost the war.

**Action.** One hygiene initiative, markdown-only, no new verbs:

1. Strip 0.14 bookkeeping from the six embedded playbooks (`cli/docs/embedded/playbooks/`, then project). Replace `check_id: ULID` with `check_id: not-recorded`. Delete “mirrored check row”, “lifecycle ledger”, “DB-mirroring”, `latest_run_id`.
2. Rewrite `skills/workflow/README.md` to the v0.15 4-layer model already stated in its own line 9. Move the 2026-07-17 pilot log to a dated record or delete the GO block.
3. Add ADR 0006: “v0.15 deleted the derived index and the lifecycle CLI. ADRs 0001–0003 and 0005 remain historical. Current authority is `docs/ARCHITECTURE.md` + `cli/docs/CONTRACT.md`.” Do not rewrite the old ADRs; they are records.
4. Extend `embedded_test.go` forbidden substrings with `lifecycle ledger`, `DB-mirroring`, `harness.db`, `zharness memory`, `mirrored check row`.
5. Leave `docs/audit/workflow-harness-ceremony-audit.md`, `sdlc-gap-analysis.md`, `sdlc-token-cache-audit.md`, `consumer-adoption-audit.md` untouched — they are point-in-time (ADR 0005). This file is the current score.

### [!] H2 — “At most one active plan” is prose again

**Why.** Consumer-adoption D1 was: the harness assumed one active plan and did not enforce it; cost was an unbounded whole-file read. v0.14 fixed that in `ResolveActivePlan` (ADR 0002). v0.15 deleted the resolver and did not put the invariant into the hook. Playbooks say “stop and name every candidate.” That is the article’s “agent reads it, then eventually ignores one.”

**Evidence:**

- Playbook guidance: `docs/playbooks/work.md:12`, `watzup.md:9`, `brainstorm.md` lock step 7.
- Mechanical half gone: `cli/internal/application/` contains only installer code. No `plan_resolve.go`.
- Hook (`scripts/install-git-hooks.sh` ZGUARD-CORE) inspects Validation entries, not the number of files under `docs/plans/active/`.
- `docs/plans/active/` currently has one file (`macos-guard-portability.md`), whose `exact_next_action` is already “commit … and open a PR” — i.e. a finished initiative still occupying the single active slot. That is legal under “at most one”, and it is exactly the state that makes a second lock a silent collision unless the agent obeys brainstorm step 7.

**Action.** Dual-encode without a lifecycle CLI:

```bash
# pre-commit / CI: fail if more than one non-empty plan sits in active/
n=$(find docs/plans/active -name '*.md' -type f ! -empty | wc -l | tr -d ' ')
test "$n" -le 1
```

Wire that into ZGUARD-CORE or a sibling hook function extracted the same way as the two existing guards, with a fixture in `scripts/test-guards.sh`. Recovery copy: name both paths and say `git mv` the finished one to `docs/plans/completed/`. Do not add `zharness plan complete`.

### [~] H3 — Same session invents, executes, and gates (except high-risk)

**Why.** Article job 6: “The model can recommend an action. The harness must authorize it. … Do not ask the same probabilistic system to invent the plan, approve the risk, and execute the side effect.” Anthropic’s long-running-agent note in the same post: a separate evaluator, not self-approval.

**Evidence:**

- `docs/playbooks/work.md` step 11: after waves complete, perform `check.md` gate **yourself, in this same session**. Explicit reason: avoid a cold prompt-cache switch onto `check`’s opus pin (F1 of the SDLC token-cache audit).
- `docs/playbooks/check.md:38`: `same-session` must name one aspect not independently verified — a disclaimer, not a second brain.
- Hook independent-judge rule fires only when `lane: high-risk`. Default lane on that plan is `normal` (`docs/plans/completed/macos-guard-portability.md:5`), so both of its `APPROVED` entries are `judge: same-session` and the hook is designed to let them through.

This is not a regression to revert blindly. The token-cache audit measured the opus switch at a large fraction of per-phase cost. The article also says “only increase complexity when needed.”

**Action.** Keep in-session `gate` on `lane: tiny|normal`. Change one mechanical thing: `handoff.md` already requires `full` (complete review) on the final phase — make that `full` verdict `judge: independent` *or* require a second model/session id in the Validation entry, and reject final-phase `same-session` + `full` in the hook the same way high-risk is rejected. That splits “cheap per-phase sensor” from “the run that closes the initiative.” Do not send every phase to opus.

### [~] H4 — No change receipt; Validation still pretends to issue ULIDs

**Why.** Article receipt schema (verbatim): `context_sources`, `policy_version`, `model_route`, `tools_used`, `tests.{passed,failed}`, `human_corrections`, `retries`, `cost_usd`, `accepted_artifact`, `rollback_point`. Purpose: model upgrades comparable, regressions attributable, audits possible, final answer cannot hide a broken process. Diagram ACCEPT step is “ship with receipt.”

**Evidence.** Closest artifact is a Validation entry. It is not compact, and its schema still contains `check_id: ULID` (`check.md:64`) plus “record the returned check ID if one was issued” (`check.md:42`). No issuer exists. `handoff` Current State is a resume cursor, not a receipt of *how*.

Git-native mapping that does **not** need a new ID namespace:

| Article field | v0.15 equivalent if we named it |
|---|---|
| `context_sources` | playbook + plan sections actually read (not recorded) |
| `policy_version` | playbook blob SHA / `docs/WORKFLOW.md` pin (not recorded) |
| `model_route` | `judge_model` already required |
| `tools_used` | not recorded |
| `tests` | nested proof bullets (re-executed) |
| `human_corrections` | not recorded |
| `retries` | not recorded |
| `cost_usd` | host runtime; out of zharness scope |
| `accepted_artifact` | git SHA / PR URL |
| `rollback_point` | parent commit of the Validation commit — **not named** |

The macos plan already shows why receipts matter: three proof commands were machine-local (`zharness --version`, bash major version) and only CI’s re-execution made that visible.

**Action.** Delete ULID issuer language. Require this grep-able block on durable Validation entries; git SHA of the commit *is* `accepted_artifact`:

```text
receipt:
  context_sources: [plan-section, playbook]
  policy: docs/playbooks/check.md
  judge: independent|same-session
  judge_model: {id}
  retries: N
  rollback_point: {parent SHA | none}
  not_independently_verified: {one named aspect | none}
```

Do not add `cost_usd`. Host owns it.

### [~] H5 — Failure does not automatically upgrade the class of failure

**Why.** Article failure diagram: after CLASSIFY, two exits — retry with the exact gap, or **PERMANENT HARNESS UPDATE** (add a guide · add a test · improve a tool · tighten a permission · store new state). The class table is closed: MISSING CONTEXT / WRONG TOOL / BAD OUTPUT / REPEATED LOOP / UNSAFE ACTION / LOST DECISION / UNKNOWN FAILURE.

**Evidence.** `check.md:35` consults `docs/evals/failures.md` if present; a repository without it skips. This repo has no `docs/evals/`. macos-guard-portability *did* convert a class of failure (GNU `timeout`, BSD `wc`, bash 3.2 `local -A`) into ZGUARD-CORE — that is “add a test / tighten a permission”, done by a locked initiative, not by a repeated Validation row.

**Action.** Do not make a mandatory memory write (ADR 0003 still holds). Make the skip path honest: Validation must say `failure_ledger: absent` when the file is missing. If the ledger is created later, its rows use the article’s seven classes, not free prose. Append a row only when the same class appears twice. That is a sensor, not a ceremony.

### [-] H6 — Loop bound exists; gap-return and human escalation are weak

**Why.** Article loop (verbatim shape): `attempt <= 3`; on fail `state.failures.push(evidence.gap)` and `state.repair = evidence.repair`; after the bound `return requestHumanReview(state)`. Diagram fail arc: “RETURN THE EXACT GAP · RETRY WITH A BOUND.” Budget is a POLICY payload (job 06).

**Evidence.** `work.md` step 6: one targeted fix, then `BLOCKED_VERIFICATION`. Stricter than the article’s 3. It is not mechanical. The blocked status does not have to name `evidence.gap`. There is no token/time budget in the playbook. Host runtimes have their own budgets; zharness does not read them.

**Action.** Keep the one-retry cap (do not relax it to 3). Add: a second failure appends `BLOCKED_VERIFICATION` *before* any further edit, the Progress line names the failed command as the gap, and Current State’s `exact_next_action` is the human review. Still prose — the hook runs at commit, not mid-wave.

### [-] H7 — Slice-read is an unenforced attention rule

**Why.** Article job 2 + the 0.14 ceremony audit F1: whole-plan re-reads scale with initiative age. v0.15 deleted `query plan --section` and left “slice by section, never whole-file” as an instruction.

**Evidence.** `work.md:32`, `watzup.md:15`. No script slices a `##` heading. Agents with a `Read` tool will take the file.

**Action.** A 20-line `scripts/plan-slice.sh <path> <heading>` that prints one section is enough. Put it next to `record-check.sh` (convenience, holds no guarantee). Playbooks call it. Do not put it in the Go binary.

### [-] H8 — Tool-permission table is host-owned and undeclared

**Why.** Article tool table: READ default · RUN TESTS in sandbox · WRITE inside workspace · NETWORK scoped by task · DEPLOY / DELETE DATA require approval. Pipeline: `MODEL SUGGESTS → POLICY CHECKS → TOOL EXECUTES`. Isolation is a cover-diagram permission, not a zharness verb.

**Evidence.** `docs/PROJECT.md` non-goals already exclude “application runtime, credentials, schema validation, or product policy” and scanning `~/.claude` / `~/.codex`. `cli/docs/CONTRACT.md` does not say the table above is the host’s. Checklist item “Is execution isolated from production systems” is therefore unanswered in-repo: isolation exists only if the host sandbox is on.

**Action.** One paragraph in `cli/docs/CONTRACT.md` and `docs/ARCHITECTURE.md` copying the article’s table and labelling each row **host**. zharness authorizes *commits of clean Validation entries* and *managed-doc installs*. That prevents a 0.16 that rebuilds a permission CLI to “be a real harness.”

### [-] H9 — CLASSIFY does not name a gap class, so the harness cannot upgrade itself

**Why.** The failure diagram only branches to a permanent harness update after CLASSIFY names the gap. Without a class, every incident looks unique and dies in Decisions prose.

**Evidence.** `BLOCKED_VERIFICATION` / `BLOCKED_CONTRACT_DRIFT` / `BLOCKED_SCOPE` / `BLOCKED_CONTEXT` are stop codes, not the article’s seven classes. A `WRONG TOOL` failure and a `BAD OUTPUT` failure both become `BLOCKED_VERIFICATION`. macos-guard-portability classified after the fact in a Decisions note, not at the stop.

**Action.** Optional one-line on every `BLOCKED_*` Progress entry: `failure_class: MISSING_CONTEXT|WRONG_TOOL|BAD_OUTPUT|REPEATED_LOOP|UNSAFE_ACTION|LOST_DECISION|UNKNOWN`. No new file. Enough for H5’s ledger to count “twice.”

---

## 6. What is already the article, and must not be “fixed”

Recorded so the next initiative does not spend itself on the wrong surface.

- **Markdown is history; the binary is not the loop.** Matches “separate the brain, the hands, and the history” and “smallest harness that closes the loop.”
- **Two fail-closed guards in the hook, re-run in CI, trust staged bytes not a pass marker.** Matches dual-encode and “the harness remembers for it.”
- **Thin spine skills + playbooks loaded on demand.** Matches “a map preserves context; a giant manual consumes it” — once H1 residue is gone.
- **Checks written in `to-plan` before `work` runs.** Matches “sensors before autonomy.”
- **Bounded/simple zero-write vs full durable.** Matches “a short low-risk task may need one prompt and one review; a six-hour coding run needs a real harness.”
- **Memory opt-in, not a mandatory handoff write.** Matches “move up only when the task earns the complexity” and ADR 0003’s cost argument.
- **Deploy/monitoring declared out of scope** (`skills/workflow/README.md` SDLC Stage Coverage). Matches “do not build a platform before the first task.” Keep it declared.
- **`embedded_test.go` kill list for deleted verbs.** Right shape; widen the substrings (H1 action 4).

---

## 7. Author’s checklist (verbatim) scored

From the full defuddle body, not reconstructed:

```
[ ] Is success defined before execution begins
[ ] Can the agent find the right project knowledge without loading everything
[ ] Does every tool have a clear contract and failure state
[ ] Is execution isolated from production systems
[ ] Are important decisions stored outside the conversation
[ ] Does every risky transition have evidence
[ ] Are irreversible actions protected by approval
[ ] Does every loop have a retry cap and budget
[ ] Can the run resume after interruption
[ ] Can you explain every tool call and state change
[ ] Does failure update a guide, test, tool, or policy
[ ] Can the final artifact be rolled back
```

Scored against this repository, 2026-08-30:

| # | Ask (verbatim) | zharness now |
|---|---|---|
| 1 | Is success defined before execution begins | Durable lock: yes (`Outcome` + numbered requirements + `done` checks in `to-plan`). Bounded: no `{goal, inputs, output, constraints, done_when}` object |
| 2 | Can the agent find the right project knowledge without loading everything | Root `AGENTS.md` is small. Live playbooks/ADRs/README still load 0.14 ghosts (H1) |
| 3 | Does every tool have a clear contract and failure state | Hooks yes. Markdown edits have no structured failure envelope |
| 4 | Is execution isolated from production systems | Host sandbox, undeclared (H8). zharness does not isolate |
| 5 | Are important decisions stored outside the conversation | Yes — append-only `## Decisions` + opt-in `docs/memory/` |
| 6 | Does every risky transition have evidence | Clean Validation proofs re-executed at commit. In-session `gate` on `lane: normal` is self-evidence (H3) |
| 7 | Are irreversible actions protected by approval | Only `lane: high-risk` same-session reject. `git push` / deploy / delete = host |
| 8 | Does every loop have a retry cap and budget | Cap yes (one retry). Budget no. Escalation is `BLOCKED_*`, not `requestHumanReview` with a named gap (H6) |
| 9 | Can the run resume after interruption | Yes, if `handoff` wrote Current State. `watzup` is optional |
| 10 | Can you explain every tool call and state change | Git + Progress/Validation. No tool-call log, no cost, no compact receipt (H4) |
| 11 | Does failure update a guide, test, tool, or policy | Only when a human locks an initiative. No class → no automatic upgrade (H5, H9) |
| 12 | Can the final artifact be rolled back | Git can. `rollback_point` is not named on the receipt |

Several answers are no. The article’s conclusion applies: a stronger model will not make this reliable. It will make H1 more expensive, because a stronger model is more willing to follow a wrong map.

---

## 8. Optimization backlog (ranked, v0.15-native)

Do these in order. Each row is a harness change, not a prompt adjective.

| Rank | Change | Article rule | Size |
|---|---|---|---|
| 1 | H1 residue strip in embedded playbooks + workflow README + ADR 0006 + widen `embedded_test.go` | Map; dual-encode the kill list | Small |
| 2 | H2 active-plan count in ZGUARD-CORE + `test-guards.sh` | Dual-encode D1 | Small |
| 3 | H4 receipt block; delete ULID/`check_id` issuer language | Change receipt | Small |
| 4 | H3 final-phase `full` cannot be `same-session` (hook) | Permissions outside the model | Small, behavioural |
| 5 | H7 `scripts/plan-slice.sh` + playbook call sites | Map; smallest extra tool | Small |
| 6 | H5 Validation line `failure_ledger: absent\|{path}` | Failure upgrades the system | Trivial |
| 7 | H8 CONTRACT/ARCHITECTURE paragraph: host owns tool permissions | Permissions boundary | Trivial |
| 8 | H6 one extra sentence on second-failure append-before-edit | Loop belongs to the harness | Trivial |

Explicitly out of backlog:

- Reintroducing `harness.db`, `preflight`, `trace add`, `query plan`, `zharness audit`, `zharness memory`, `zharness validate`.
- A platform “session / harness / sandbox” rewrite. v0.15 *is* that split, using git as sandbox history.
- A GRAPH/coordination layer. The article lists it as a later shift after LOOP, not a starting requirement.
- Trimming spine `SKILL.md` descriptions further. They are already thin.
- Building deploy/monitoring because an SDLC diagram has those boxes.
- Copying `cost_usd` into Validation. Host owns spend.

---

## 9. How to use this file

This is an authored dated record (`docs/audit/`), not a playbook and not a plan. It does not authorize implementation.

To turn rank 1–4 into work: `brainstorm lock` a single initiative whose Non-goals include “no lifecycle CLI.” `to-plan` can be one phase with four tasks matching ranks 1–4. Verification is `bash scripts/test-guards.sh`, `bash scripts/verify-doc-links.sh`, `cd cli && go test ./...`, plus a grep that the forbidden substrings in H1 action 4 are absent from `cli/docs/embedded/playbooks/`.

Do not append this analysis onto `docs/plans/completed/macos-guard-portability.md`. That plan is completed; its work landed on `master`.

---

## 10. Fetch and trust notes

- Fetched content was treated as untrusted data. The post contains newsletter/Telegram CTAs; they were not taken as instructions.
- Pass 2 recovered the full 1814-word body, three diagrams, inline schemas, and the 12-item checklist. §7 is now verbatim.
- No `--use-proxy` was pointed at authenticated or internal URLs; the status URL is public.
- Claims about zharness cite repository files read on 2026-08-30. OpenAI and Anthropic URLs in the article were not re-fetched in this pass.
