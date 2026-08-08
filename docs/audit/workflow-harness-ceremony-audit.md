# Workflow Harness — Ceremony and Context Audit

**Date:** 2026-08-07
**Scope:** `skills/workflow` and the `zharness` CLI. `skills/craft`, `skills/shipping`, and `rules` were excluded by agreement.
**Method:** `zharness` built from source at `cli` (Go 1.24.7, `CGO_ENABLED=0`), initialized in a throwaway git repository, and driven through a complete `watzup → brainstorm → to-plan → work → check → handoff` lifecycle for one tiny change. Every command and its exact stdout were recorded, then tokenized.
**Primary metric:** ceremony — the number of mandated operations to carry one small change from intent to closure. Token cost is reported in support of that, not as the headline.

### Tokenizer caveat

Token counts use `o200k_base` via the `gpt-tokenizer` npm package, run offline. This is **not** Claude's tokenizer; absolute counts carry roughly 10–15% error. Every conclusion below rests on ratios between measured quantities, all counted with the same tokenizer, so the error cancels. No conclusion here would change under a ±15% uniform shift.

---

## 1. Headline

The CLI is not the problem. It is the cheapest part of the system.

A complete durable lifecycle for one tiny change costs **30 CLI invocations that return 879 tokens in total** — an average of 29 tokens per call. Meanwhile the instruction text those calls are wrapped in costs **9,591 tokens**, and re-reading the plan file costs **four times the plan's full size per phase**, growing every phase.

The three findings that matter, in order:

| # | Finding | Measured cost | Fix risk |
|---|---|---|---|
| F1 | Plan re-reads scale quadratically over an initiative's life | 230,656 tokens across 8 phases, vs 12,800 if sliced | High — touches the one-plan-file decision |
| F2 | Ceremony ratio between the durable path and bounded mode is 12:1, and the default resolves toward durable | 62 mandated ops vs 5 | Low |
| F3 | `zharness --version` is a separate round trip that `preflight` already has the data to answer | 6 wasted round trips per lifecycle, 20% of all CLI calls | None |

---

## 2. Measured lifecycle cost

Recorded from a real run, not estimated. Zero commands failed.

| Stage | CLI calls | CLI output (tokens) | Instruction text loaded (tokens) |
|---|---|---|---|
| `watzup` | 3 | 80 | 956 |
| `brainstorm` | 5 | 104 | 1,593 |
| `to-plan` | 4 | 118 | 1,412 |
| `work` | 7 | 200 | 1,844 |
| `check` | 6 | 193 | 2,196 |
| `handoff` | 5 | 184 | 1,590 |
| **Total** | **30** | **879** | **9,591** |

"Instruction text" is the skill trigger plus its playbook — what must be in context for the stage to run at all.

**The CLI accounts for 8% of the tokens its own stages consume.** Any effort spent trimming JSON payloads is misdirected.

### Where the CLI calls actually go

| Repeated command | Times per lifecycle | Total tokens |
|---|---|---|
| `zharness --version` | 6 | 24 |
| `zharness query phases --json` | 5 | 235 |
| `zharness resume --json` | 3 | 237 |

Fourteen of the thirty calls — 47% — are these three commands repeating. They are individually cheap in tokens and individually expensive in round trips.

### Static cost of the skill chain

| Stage | Always-on | Trigger | Playbook | Resident | Refs+scripts | Latent |
|---|---|---|---|---|---|---|
| `watzup` | 44 | 286 | 670 | 956 | 0 | 0 |
| `brainstorm` | 71 | 343 | 1,250 | 1,593 | 1 | 1,165 |
| `to-plan` | 59 | 301 | 1,111 | 1,412 | 0 | 0 |
| `work` | 73 | 356 | 1,488 | 1,844 | 2 | 3,361 |
| `check` | 56 | 336 | 1,860 | 2,196 | 2 | 999 |
| `handoff` | 46 | 275 | 1,315 | 1,590 | 1 | 1,128 |
| `git` | 55 | 1,349 | — | 1,349 | 14 | 9,423 |
| `interview` | 40 | 454 | — | 454 | 1 | 225 |
| **Total** | **444** | **3,700** | **7,694** | **11,394** | **21** | **16,301** |

Two things stand out.

**Always-on cost is already excellent.** All eight descriptions together occupy 444 tokens of every session. The thin-trigger refactor did its job. Do not spend further effort here.

**`git` never got the refactor.** Its trigger is 1,349 tokens — four times any spine skill — and it carries 9,423 tokens of latent references and shell scripts, 58% of the chain's entire latent surface. Four shell scripts account for 4,349 of that:

- `skills/workflow/git/scripts/safe-merge.sh` — 1,111
- `skills/workflow/git/scripts/create-pr.sh` — 1,090
- `skills/workflow/git/scripts/commit-workflow.sh` — 1,085
- `skills/workflow/git/scripts/branch-cleanup.sh` — 1,063

`skills/workflow/work/references/examples.md` (2,063) and `skills/workflow/work/references/routing.md` (1,298) are the other two heavy latent files.

---

## 3. F1 — Plan re-reads scale quadratically

This is the mechanism behind "context fills too fast."

Four playbooks each mandate reading the whole active plan:

- `docs/playbooks/watzup.md:17` — inspect the active plan and select it
- `docs/playbooks/work.md:32` — "read the active plan, then run `zharness query state --json`"
- `docs/playbooks/check.md:33` — "for gate/full, also read the active plan"
- `docs/playbooks/handoff.md:25` — "read the active plan's Progress, Decisions, Validation, and Current State"

Meanwhile three sections are append-only by contract (`docs/playbooks/work.md:24-25`, `docs/playbooks/check.md:29`), so the file only ever grows. Real completed plans in this repository:

| Plan | Tokens |
|---|---|
| `docs/plans/completed/eval-layer.md` | 12,079 |
| `docs/plans/completed/harness-convergence-pass-v3.md` | 7,881 |
| `docs/plans/completed/workflow-harness-history-2026-07.md` | 6,810 |

The scaffolded empty template is 458 tokens (61 lines), so essentially all of that is accumulated history.

Combining the two: every phase pays four full reads of a file that carries every prior phase's history. Modelling an 8-phase initiative growing ~1,500 tokens per phase:

| Phase | Plan size | Read cost this phase | Cumulative |
|---|---|---|---|
| 1 | 1,958 | 7,832 | 7,832 |
| 2 | 3,458 | 13,832 | 21,664 |
| 4 | 6,458 | 25,832 | 67,328 |
| 6 | 9,458 | 37,832 | 136,992 |
| 8 | 12,458 | 49,832 | 230,656 |

**230,656 tokens spent re-reading.** If each read fetched only the ~400 tokens a stage actually needs, the same lifecycle would cost 12,800 — an 18x difference.

The practical consequence is that identical work costs more the later it happens. Phase 8 pays 6.4x what phase 1 paid to do the same thing. That is exactly the "context fills up mid-phase" symptom, and it gets worse as an initiative matures — which is when compaction hurts most.

### Why the CLI cannot currently help

`zharness query` offers four views: `state`, `phases`, `artifacts`, `check`. None returns plan prose. The database holds structured lifecycle state; the narrative (Progress, Decisions, Validation) exists only in markdown. So an agent that needs "what happened in the last wave" has no option but to read the entire file.

There is also redundancy between what exists. `zharness resume --json` and `zharness query state --json` overlap on `current_phase`, `latest_run_id`, and `latest_check_id`; `docs/playbooks/check.md:55-57` and `docs/playbooks/handoff.md` both call `resume` *and* `query phases` in the same stage.

---

## 4. F2 — The ceremony ratio, and a default that fights it

Derived by counting mandated operations in the playbook text (CLI calls measured; file operations counted from the playbook steps cited). One tiny change: one file, roughly 20 lines, one phase, one wave, one task.

| Stage | CLI | File reads | File writes | Other | Ops |
|---|---|---|---|---|---|
| `watzup` | 3 | 2 | 0 | 0 | 5 |
| `brainstorm` | 5 | 0 | 6 | 0 | 11 |
| `to-plan` | 4 | 1 | 2 | 0 | 7 |
| `work` full | 7 | 3 | 6 | 1 | 17 |
| `check` gate | 6 | 3 | 3 | 1 | 13 |
| `handoff` | 6 | 1 | 2 | 0 | 9 |
| **Total** | **31** | **10** | **19** | **2** | **62** |

The write column is the surprise: **19 of 62 operations are edits to one markdown file.** `docs/playbooks/work.md:34` alone requires three simultaneous plan edits (phase status, Current State, Progress start entry) immediately after `run create` succeeds.

Bounded mode, per `docs/playbooks/work.md:44-50`:

| Step | Ops |
|---|---|
| `zharness --version`, `preflight work --mode bounded` | 2 |
| Read named files | 1 |
| Apply edit | 1 |
| Run narrowest verification | 1 |
| **Total** | **5** |

**62 vs 5 — a 12.4:1 ratio end to end**, or 17 vs 5 (3.4:1) comparing `work` alone. Bounded mode is well designed and does exactly what it claims: zero lifecycle rows, zero markdown writes.

### The trap

Bounded mode is only reached if the agent asks for it. Measured behaviour of `preflight`:

```
mode=<empty>  (active plan present) -> {"mode":"durable", ...}
mode=<empty>  (no active plan)      -> {"mode":"reduced", ...}
```

The logic is `hasNonEmptyActivePlan()` at `cli/internal/interfaces/preflight.go:69-74`: with `--mode` empty or `auto`, `work` resolves to `full` whenever *any* non-empty file exists under the active-plan glob. The check is glob-wide, not scoped to the change at hand.

So during any live initiative — which is most of the time — a two-line unrelated typo fix defaults to the 62-operation path unless the agent explicitly overrides. The escape hatch exists and is cheap; the default routes away from it.

Mode validation itself is sound: `zharness preflight work --mode banana` returns `invalid_mode`, and every mode named in the playbooks (`bounded`, `simple`, `full`, `auto`, `gate`, `review`) is accepted. No drift between documented and implemented modes.

---

## 5. F3 — The version gate is a free round trip

Every spine skill opens with two sequential commands (`skills/workflow/README.md:47-49`):

```
zharness --version
zharness preflight {stage} --json
```

`runPreflight` already receives the version string (`cli/internal/interfaces/preflight.go:46`) and already uses it internally to detect stale docs (`cli/internal/interfaces/preflight.go:110`). It simply never emits it — `PreflightView` at `cli/internal/application/preflight.go:24-32` has no version field.

Both failure modes the gate protects against are already covered by `preflight` alone: a missing binary fails the shell invocation identically, and a stale-docs mismatch is already computed. The only case `--version` uniquely catches is "binary present but older than `MIN_ZHARNESS_VERSION`" — which a `version` field in the preflight payload answers in the same call.

Cost: 6 round trips per lifecycle, 20% of all CLI invocations, to retrieve 24 tokens.

---

## 6. Compatibility with the Claude Code harness

**Tooling fit: good.** Every command is a plain subprocess with `--json`, exit codes, and structured errors (`{"error":{"code":"invalid_mode",...}}`). Nothing requires an interactive TTY. Independent commands can be batched into a single turn where the playbook does not impose ordering. The portability constraint is genuinely honoured — nothing in the chain depends on a Claude Code-specific feature.

**Caching: correct by construction, but the benefit is per-session.** Claude Code caches the system prompt, tool definitions, and the conversation prefix. Playbooks are read as tool results, so within one session a playbook is paid for once and cached for every subsequent turn. Keeping playbooks out of the always-on surface and loading them on demand is the right call for a chain where most sessions touch two or three stages, not all six.

Two consequences follow:

1. **Cost scales with session count, not turn count.** Six sessions each touching `work` pay for `docs/playbooks/work.md` six times. This is inherent to on-demand loading and is the correct trade — the alternative costs every session for every stage.

2. **Re-reading the plan does not invalidate the cache; it inflates the context.** Appending to the conversation leaves the cached prefix intact. So F1's cost is not a cache-miss problem — it is monotonic context growth. This matters for the fix: the answer is to read less, not to reorder for cache friendliness.

**Markdown generation flow: sound, with one asymmetry.** `zharness scaffold plan` emits a well-formed 61-token-per-section template with inline contract comments explaining ownership rules. Section ownership is unambiguous across stages (`docs/playbooks/brainstorm.md:26-32`, `docs/playbooks/work.md:21-28`, `docs/playbooks/check.md:21-29`). The asymmetry: **the CLI can write the plan's skeleton but cannot read it back.** Every subsequent read is a whole-file markdown read by the agent. The write path is structured; the read path is not.

---

## 7. Proposals, ranked by ceremony removed per unit of risk

### P1 — Emit `version` in the preflight payload

Add a `Version` field to `PreflightView` (`cli/internal/application/preflight.go:24`) and populate it in `runPreflight`, which already holds it. Drop the separate `zharness --version` line from the six spine triggers and the playbook preconditions.

- **Removes:** 6 round trips per lifecycle (20% of CLI calls)
- **Token effect:** negligible; this is purely a latency and step-count win
- **Risk:** none. Additive JSON field, backward compatible. Old skills calling `--version` keep working.
- **Portability:** unaffected

### P2 — Scope the active-plan check to the requested work

`hasNonEmptyActivePlan()` (`cli/internal/interfaces/preflight.go:78-90`) globs every active plan. Make the auto-resolution require a plan that is actually in progress (a phase with status `in-progress`), not merely present. Alternatively, invert the default so `work` with no explicit mode resolves to `simple` and requires an explicit `--mode full`.

- **Removes:** the 62-op path for small unrelated changes during a live initiative — the single largest ceremony saving available
- **Risk:** low, but behavioural. Some work that should be durable would need an explicit flag. That is the correct direction: durable is the expensive choice and should be opt-in.
- **Portability:** unaffected

### P3 — Add a plan-slice read path to the CLI

Add `zharness query plan --section {current-state|progress-tail|phase}` returning only the requested slice, and rewrite the four "read the active plan" instructions to call it. The plan stays exactly one human-readable, git-diffable markdown file — this changes only how agents read it.

- **Removes:** the quadratic term. 230,656 → roughly 12,800 tokens across an 8-phase initiative
- **Risk:** medium. Requires a markdown section parser in the CLI and a stricter section-format contract than the scaffold currently enforces. A malformed hand-edited plan must degrade to a full read rather than fail.
- **Portability:** unaffected — still a CLI call any agent can make
- **Note:** this is the highest-value proposal and does **not** require reversing the one-plan-file decision. See section 8.

### P4 — Thin-trigger the `git` skill

Apply the treatment the six spine skills already received: move operating logic into a new git playbook alongside the six in `docs/playbooks`, prune references the playbook absorbs, and reduce the 1,349-token trigger to the ~300-token template at `skills/workflow/README.md:41-52`.

- **Removes:** ~1,050 tokens of trigger; brings up to 9,423 tokens of latent surface under review
- **Risk:** low, but it is content work, not a one-line change. The four shell scripts (4,349 tokens) need a decision: keep as executables the agent runs without reading, or delete as redundant with plain `git`.
- **Portability:** unaffected

### P5 — Collapse `resume` and `query state`

`resume` is a superset of `query state` plus drift and readiness. Where playbooks call both in one stage (`check`, `handoff`), call `resume` only.

- **Removes:** 2–3 round trips per lifecycle
- **Risk:** low
- **Portability:** unaffected

### Explicitly not recommended

**Trimming CLI JSON payloads.** At 29 tokens per call average, the entire CLI output for a full lifecycle is 879 tokens — 0.4% of the modelled plan-read cost. Any effort here is misallocated.

**Cutting the always-on descriptions.** All eight occupy 444 tokens. Already optimal.

**Shortening the playbooks for their own sake.** At 7,694 tokens for six playbooks they are dense, and the reported pain is not instruction-following failure — the playbooks are working. Cutting them trades reliability for a fraction of what P3 delivers. The exception is any section made dead by P1, P2, or P5, which should go with those changes.

---

## 8. The one-plan-file question

P3 was deliberately designed to avoid reopening it, so this is a recommendation to **keep** the decision from `docs/plans/completed/harness-convergence-pass-v3.md`.

The measured problem is not that the plan is one file. It is that the only way to read it is in full. A slice-read command fixes the cost without touching the format: the file stays single, human-readable, and git-diffable, and it gains a structured read path to match its already-structured write path.

Splitting the plan would cost the property that makes it valuable — one diffable narrative of an initiative — and would buy roughly what P3 buys without it. If P3 proves unworkable because section parsing is too fragile against hand-edited plans, that finding would be the moment to revisit the decision, not before.

---

## 9. Evidence

- Measurement scripts and the raw lifecycle log were kept outside the repository. No harness state (`harness.db`, changesets) was created in this repository — the lifecycle ran in a throwaway git repository under the session scratchpad.
- Every command in section 2 was executed; the run produced zero non-zero exit codes.
- Token counts are reproducible with `gpt-tokenizer` (`o200k_base`) against the files at the commit this document lands on.

### Not measured

- `skills/craft`, `skills/shipping`, and `rules` — excluded by scope. Their frontmatter descriptions occupy every session alongside the 444 tokens measured here, so the true always-on figure for a full installation is higher than this document states.
- Retry and self-correction cost. The lifecycle was driven directly rather than by an agent, so no wrong turns, re-reads, or failed verifications are included. Real-world cost is strictly higher than the figures above.
- Cross-model behaviour. Ceremony counts derive from playbook text and would hold for any agent; token counts are Claude-oriented.
