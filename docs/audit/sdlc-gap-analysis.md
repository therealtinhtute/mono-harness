# SDLC System Gap Analysis — Skills, Harness, Tokens, Caching

**Date:** 2026-08-11
**Scope:** the full system — `skills/` (workflow, shipping, craft), `docs/playbooks/`, and the `zharness` CLI — evaluated against five criteria: SDLC stage coverage, CLI harness efficiency, token efficiency, harness in/out rigor, and prompt caching.
**Method:** builds on the measured lifecycle in `sdlc-token-cache-audit.md` (same dogfood repo, same built-from-source binary), plus new empirical probes of the CLI's error envelopes, exit codes, latency, and lock behavior recorded in this pass.
**Companions:** `workflow-harness-ceremony-audit.md` (ceremony counts), `sdlc-token-cache-audit.md` (token + cache economics, findings F1–F5). This document does not repeat their evidence; it cites it.

---

## Verdict summary

| Criterion | Grade | One-line verdict |
|---|---|---|
| 1. Skills & workflow SDLC coverage | **B** | Plan→code→verify→commit is complete and well-gated; **deploy and monitoring do not exist** |
| 2. CLI harness & tool calls | **A−** | ~21 ms/call, zero failed commands across two audits; two redundant round trips remain |
| 3. Token efficiency | **B−** | 41.6% of `work` is ceremony; the one shipped read-optimization is broken (F2) |
| 4. Harness in/out rigor | **B+** | Contract-grade errors, exit codes, locking, replay; **no logging, one accounting drift** |
| 5. Prompt caching | **B** | 77% of raw cost already absorbed; model-scoped switching burns $0.275/phase (F1) |

The system's engineering quality is high — the gaps are concentrated in *coverage* (no post-merge lifecycle) and *economics* (model boundaries, ceremony cadence), not in correctness.

---

## 1. Skills & workflow — SDLC stage coverage

Mapping every skill against the five canonical SDLC stages:

| SDLC stage | Covered by | Status |
|---|---|---|
| **Planning** | `brainstorm` (4 modes) → `to-plan` (phases/waves/tasks/checks), `interview` as optional validator | ✅ Strong — requirements are locked falsifiable (`R1 [accepted]: ... | source:`), verification is written *before* execution (to-plan step 6: "missing verification is a planning blocker") |
| **Coding** | `work` (full/bounded), mode routing via `zharness next` | ✅ Strong — bounded mode's zero-write rule is a genuinely good escape hatch; boundary enforcement (`BLOCKED_CONTRACT_DRIFT`) is unusual and valuable |
| **Testing** | `check` (gate/full/review/bounded) + per-task verification commands in `work` step 6 | ✅ Verification-side complete. ⚠️ No test-*authoring* guidance: nothing tells `to-plan` what a good check is beyond "observable command", so check quality depends entirely on the planning agent's judgment |
| **Deployment** | — | ❌ **Absent.** `git` ends at commit/push/PR/merge. No deploy, no release, no rollback skill. `skills/shipping/` (`create-cli`, `turbo-mono-platform`) is project *scaffolding*, not deployment |
| **Monitoring** | — | ❌ **Absent.** `watzup` is a session recap, not production observation. The only feedback loop is `docs/evals/failures.md` — a regression ledger `check` consults, which is *development* telemetry, not runtime telemetry |

**Gap G1 — the pipeline ends at the merge.** The chain's own diagram (`brainstorm → to-plan → work → check → git → handoff`) closes the loop back to the *repo*, never to *production*. For a personal skills repo this may be a deliberate non-goal — but it is currently an *undeclared* one. Two acceptable resolutions: (a) a `ship` skill owning deploy→verify→rollback with the same preflight/playbook shape, or (b) one paragraph in `skills/workflow/README.md` declaring post-merge lifecycle out of scope. Option (b) costs five minutes and removes the ambiguity; (a) is only worth it if a real deploy target exists.

**Gap G2 — task→agent routing is single-threaded in practice.** `work.md` step 5 permits parallelizing "tasks explicitly marked parallel-safe", but no mechanism exists: no subagent fan-out pattern, no per-task context isolation. Every "parallel-safe" task still executes serially in one growing context — which is exactly what drives the +26%/task cost super-linearity measured in F4. Fan-out of independent wave tasks to fresh-context subagents would both parallelize and flatten that curve. Effort is non-trivial (trace attribution across contexts needs `--run-id` discipline), so this ranks below the cheaper wins.

**What is right and should not change:** the thin-trigger design (6 spine SKILL.md ≤30 lines, logic in CLI-embedded playbooks) is the correct architecture — any agent that can read a file and run a CLI can execute the lifecycle, and playbooks version with the binary (`docs_version` staleness gate in preflight proves it).

---

## 2. CLI harness & tool-call efficiency

**Latency is a non-issue.** Ten consecutive `preflight work --mode full --json` calls: 208 ms total — **~21 ms per invocation**, SQLite included. Both audits' full lifecycles recorded **zero failed commands**. The harness's cost is never the process; it is the conversation round trip wrapped around it (~15–25K cached-token replay per call at mid-phase). That inversion is the single most important fact for prioritization: *removing a call saves 1000× more than speeding one up.*

Redundant calls, in descending order of waste:

| # | Redundancy | Evidence | Fix |
|---|---|---|---|
| C1 | `query plan --section phase` degrades to a whole-file read on every scaffold-shaped plan — worse than not calling it | F2: 5,863 B degraded vs 1,290 B intended vs 5,580 B raw file | One regex (roadmap R1) |
| C2 | Per-task `trace add`: full round trip for a 36-byte return, 5–12×/phase | F3: 41.6% ceremony share; batching −16% | Batch flag (R3) |
| C3 | `check` must call `resume --json` separately because `preflight check` carries no `context` packet — unlike `work`/`watzup`/`handoff`, whose preflights do | Measured: preflight check/to-plan/brainstorm return no `context` field | Extend the stage-shaped context packet (P4 in the convergence plan) to `check` (R4) |
| C4 | `query traces --phase nonexistent` returns `[]` exit 0 — indistinguishable from "phase exists, zero traces", so an agent may burn a diagnostic turn | Probed this pass | Emit `{"phase_known": false}` or exit 1 with `unknown_phase` (R8) |

Already fixed and worth crediting: the ceremony audit's F3 (`--version` as a separate round trip) is closed — `version` now rides the preflight response, and the playbooks read it from there.

---

## 3. Token efficiency

Fully quantified in `sdlc-token-cache-audit.md` §2–6; the ledger:

- **Waste, confirmed:** the degraded plan read (C1), per-task `trace add` cadence (C2), and the `check` model switch (F1 — economically a token issue: 380K cache-read tokens re-billed cold at opus prices every phase).
- **Optimized, confirmed:** preflight context packets eliminate duplicate `query state`/`query phases` reads in `work`/`handoff` (the playbooks explicitly forbid re-fetching — `work.md` step 2: "do not call query state/query phases again"); append-only sections + "CLI owns the pen" prevent whole-plan rewrites; tail-bounded queries (`--tail N`) cap history reads in `handoff`; bounded mode's zero-write rule keeps small changes at 5 ops instead of 62.
- **Truncation/compression:** CLI outputs are already minimal (29 tokens/call average per the ceremony audit) — there is nothing left to compress on the CLI side. The compressible surface is the *conversation*, which is the harness runtime's job (context editing / compaction), not this repo's.
- **Accounting exists but has drifted:** `zharness db status` ships a `context_cost_estimate` block (bytes/4 per stage playbook + active plan). Its own `note` field admits it "reflects today's full-plan-read path; the index read path … is not yet reflected". Honest, but it means the one built-in cost gauge measures a superseded read pattern (R6).

---

## 4. Harness in/out rigor

The strongest area. Probed this pass:

**Contract-grade error surface.** Every failure renders `{"error": {"code": "...", "message": "..."}}` under `--json`, plain text otherwise; codes are stable and enumerated in `cli/docs/CONTRACT.md` (227 lines, ~41 documented I/O shapes). Exit codes are disciplined: 1 = user error, 2 = system error, with `silentExit` for commands whose structured body (e.g. `validate`'s findings) already tells the story. Probes:

```
$ zharness check record --judge banana ... --json
{"error":{"code":"empty_proof_links","message":"check: proof_links required for verdict \"APPROVED\""}}  # exit 1
$ zharness story --json
{"error":{"code":"missing_required_field","message":"story: slug is required"}}                          # exit 1
```

**Concurrency and durability.** A repository-level lock with a 5 s default timeout and *typed* failures (`repository_lock_timeout` / `repository_lock_unsupported` / `repository_lock_failed`, all exit 2) guards writes. State is event-sourced: ULID changesets in `.kit/changesets/` materialize `harness.db`, so the DB is rebuildable (`repository_replay_test.go` pins it) and `db status` exposes `schema_version`, a `fence` ULID, and per-table row counts for drift diagnosis. `check record` re-executes proof commands before recording clean verdicts — the harness refuses to be lied to.

**Retry logic:** none in the CLI, and that is *correct* — a local, idempotent-read/transactional-write CLI should fail fast and let the calling agent decide; the lock timeout is the only wait that belongs inside. The playbooks put retry where it belongs (one targeted fix after a failed verification, then `BLOCKED_VERIFICATION`).

**Gap G3 — no logging, anywhere.** The CLI writes nothing but its stdout response. When a lifecycle goes wrong, the only forensic record is the agent's conversation transcript — which is exactly the artifact that gets compacted, summarized, or lost. A `--log-file` appending one JSONL line per invocation (`ts, argv, exit, ms, error_code`) into gitignored `.kit/log/` would cost ~30 lines of Go and make every future audit's "recorded from a real run" methodology a free byproduct. This is the one rigor gap worth closing (R5).

**Minor:** validation reports first-failure-only (the bad `--judge banana` was masked by `empty_proof_links`) — acceptable for an agent caller who fixes-and-retries, worth a note in CONTRACT.md at most.

---

## 5. Prompt caching

Verified in `sdlc-token-cache-audit.md` §2–3 against the platform's documented semantics (0.10× reads, 1.25× 5-minute writes, model-scoped keys, 20-block lookback, per-model minimum prefixes):

- **Effective hit economics:** caching absorbs **77%** of the chain's raw cost ($4.28 → $0.99/phase). Every stage's prefix clears its model's minimum cacheable size (smallest margin: `watzup` at 5,017 tokens vs haiku's 4,096 floor — worth watching if the watzup playbook ever shrinks).
- **TTL:** the 5-minute default is correct for this workload — stages chain back-to-back, and the harness runtime (Claude Code) owns breakpoint placement and TTL; nothing repo-side to configure.
- **Invalidation logic is where the money leaks:** caches are model-scoped, and the chain's frontmatter declares six model switches per initiative. The `work`(sonnet)→`check`(opus) switch repeats per phase and costs $0.275/phase — 63% of the most expensive stage (F1). This is an *invalidation* problem wearing a *model-assignment* costume.
- **Lookback:** a 5-task phase spans ~3–4 lookback windows (F4). Not a failure — the runtime refreshes breakpoints — but it corroborates the batching case (R3) and the smaller-phases bias already in `to-plan`.
- **Stability hygiene:** playbooks are version-pinned and byte-stable (good — they sit in the cacheable prefix); the plan file mutates per write but is read through tool results in the messages tier, where appends don't invalidate the prefix. No silent invalidators (timestamps/UUIDs in system-position content) found.

---

## 6. Roadmap — prioritized by impact ÷ effort

Items R1–R3 and R7 restate the optimization spec's P1–P4 (kept here so this table is self-contained); R4–R6, R8, G1, G2 are new from this pass.

| # | Item | Impact | Effort | Ratio | Concrete change |
|---|---|---|---|---|---|
| **R1** | Fix `planPhaseHeading` regex | High — restores 4.3× on `work`'s hot read | **Trivial** (1 regex + 1 test) | ★★★★★ | `plan_query.go:34`: also match `^\s*- phase_slug: (\S+)$` list items, slice to next sibling/`## `; test against literal `scaffold plan` output |
| **R2** | Route per-phase gate to `check gate` on sonnet; reserve `check full` (opus) for final phase | Highest $ — $0.275/phase, ~28% of chain | Medium (2 playbook edits + 1 precondition) | ★★★★☆ | `work.md` step 11 → `check gate`; **`handoff.md` step 6 gains precondition: initiative closure requires a `check full` verdict on the final phase** — without this half, the change is a quality cut, not an optimization |
| **R3** | Batch `trace add` per wave | −16% on `work` | Low-med (CLI flag + playbook edit) | ★★★★☆ | `trace add --tasks '[{task,task_status,summary},...]'` mirroring `decision add --decisions`; flush immediately on `BLOCKED`/`NEEDS_CONTEXT` so only clean progress buffers; DB rows stay per-task (resumability unchanged) |
| **R4** | Add `context` packet to `preflight check` | 1 round trip/phase (~15K cached tokens) | Low | ★★★★☆ | Extend the P4 stage-shaped context to the `check` stage; delete the separate `resume --json` step from `check.md` step 1 |
| **R5** | JSONL invocation log | Debuggability + free audit telemetry | Low (~30 lines Go) | ★★★☆☆ | Append `{ts, argv, exit, ms, error_code}` to `.kit/log/zharness.jsonl` (gitignored); no flag needed — always-on, rotation at 1 MB |
| **G1** | Declare or build the post-merge lifecycle | Closes the SDLC coverage gap | Trivial (declare) / High (build) | ★★★☆☆ | Minimum: one README paragraph declaring deploy/monitoring out of scope. If a deploy target exists: a `ship` skill with the standard preflight/playbook shape owning deploy→verify→rollback |
| **R6** | Refresh `context_cost_estimate` | Keeps the built-in gauge honest | Low | ★★★☆☆ | `db_status.go`: model the section-read path (post-R1) instead of full-plan; drop the self-confessed drift note |
| **R7** | Delete orphaned `references/` (work, check, brainstorm, handoff) | Removes drift surface | Trivial | ★★★☆☆ | `git rm -r`; keep `git/` + `interview/`; `verify-doc-links.sh` must stay green |
| **R8** | Distinguish unknown phase from empty result | Saves diagnostic turns | Trivial | ★★☆☆☆ | `query traces/decisions --phase X` where X has no story row: exit 1 `unknown_phase` (or `{"phase_known":false}`) |
| **G2** | Subagent fan-out for parallel-safe wave tasks | Flattens the +26%/task curve; wall-clock | High | ★☆☆☆☆ | Defer until R1–R4 land and remeasure; needs per-task context isolation + trace attribution design |

**Sequencing:** R1+R7+R8 in one trivial-risk commit; R2 next (the only judgment call — the `handoff` precondition is load-bearing); R3+R4 as one "fewer round trips" CLI release; R5+R6 opportunistically; G1 as a README edit now, a skill only if a deploy target materializes; G2 last, re-justified against post-R3 measurements.

**Expected combined result** (R1–R4, from the cost model): chain cost $0.99 → **~$0.63/phase (−36%)**, `work` turns 37 → ~31, `check` per-phase cost −63%. G1/G2 are coverage/latency plays, not token plays, and are deliberately not counted in that number.
