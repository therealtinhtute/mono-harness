---
id: 01KYS4ERXT5SANHBG7W0XBNR1J
intake_id: 01KYS4ERXWYXQ7K2M2E2NQYT7M
slug: eval-layer
status: active
lane: normal
created: 2026-07-30
updated: 2026-07-30
---

# Eval Layer

## Outcome

The harness gains the one layer it does not have: a way to detect that agent output is *wrong* rather than merely that a command exited 0.

Today `check` proves that tests ran and that the diff matches the plan. It cannot detect an agent-authored instruction that is internally plausible and externally false, and it cannot detect that a failure class has already been seen before. Both are invisible to every existing gate.

Done when:

1. A broken repo-relative cross-reference in any tracked doc fails a deterministic gate before commit.
2. Every `check` verdict states whether it was produced by an independent judge or by the same session that wrote the code, and which model produced it.
3. A `REQUEST_CHANGES` finding becomes a durable ledger entry, and a class seen twice is required to graduate into a deterministic check.

## Authority and Requirements

**Evidence that motivated this initiative** (audited 2026-07-30, this repo, 113 tracked markdown files):

| Claim class | Result |
|---|---|
| `zharness <subcommand>` references in docs | **clean** — every claimed subcommand exists in the real CLI surface; only false positive is the prose `zharness not found` |
| repo-relative doc cross-references | **11 broken**, concentrated in `skills/craft/create-skill/references/` (a `references/x.md` link written from *inside* `references/`, resolving to `references/references/x.md`) plus 2 stale targets |
| illustrative example paths (`.kit/planning/*`, `.vercel/project.json`, `package.json`) | 32 hits, **not defects** — proof that any gate here needs an explicit ignore mechanism or noise will kill it |

The 11 real broken references:

- `skills/craft/create-skill/references/plugin-marketplace-overview.md` → `references/plugin-marketplace-{hosting,schema,sources,troubleshooting}.md` (4)
- `skills/craft/create-skill/references/skill-anatomy-and-requirements.md` → `references/{metadata-quality-criteria,script-quality-criteria}.md` (2)
- `skills/craft/create-skill/references/skill-creation-workflow.md` → `references/benchmark-optimization-guide.md` (1)
- `skills/craft/write/references/write-vi-notion-report.md` → `references/write-vi-notion-illustrations.md` (1)
- `skills/shipping/turbo-mono-platform/references/companion-skills.md` → `references/shadcn-rules.md` (1)
- `skills/craft/write/examples/eval-queries.md` → `skills/write/SKILL.md` (actual: `skills/craft/write/SKILL.md`) (1)
- `skills/workflow/README.md` → `docs/plans/active/harness-convergence-pass-v3.md` (moved to `docs/plans/completed/`) (1)

**Requirements**

- R1 — A deterministic gate command fails with a non-zero exit and a `file -> path` list when a tracked doc references a repo-relative path that does not resolve, either from the repository root or from the referencing file's own directory.
- R2 — That gate carries an explicit, file-based ignore convention so illustrative example paths do not produce findings.
- R3 — The gate is declared as a repository gate command in `CLAUDE.md`, not embedded into the shipped playbook, because the shipped playbook must remain runnable in projects that do not have this repository's scripts.
- R4 — All 11 known broken references are repaired in the same phase that introduces the gate, so the gate is green at merge.
- R5 — `check`'s output block states `judge: independent | same-session` and `judge_model: {model identifier}`.
- R6 — When `judge: same-session`, an `APPROVED` verdict must name at least one thing that was not independently verified.
- R7 — `docs/evals/failures.md` exists as an append-only ledger with `date | phase-slug | failure class | how it was caught | permanent check`.
- R8 — `check` reads that ledger during plan alignment and, for any failure class recorded two or more times, explicitly states whether the current diff is clean of it.
- R9 — Durable `gate`/`full` appends one ledger row per `REQUEST_CHANGES` finding. `review` and `bounded` write nothing, preserving the existing zero-write rule.
- R10 — A failure class recorded a second time must be graduated into a deterministic check under `scripts/`; a ledger note alone is not acceptable closure.
- R11 — Every playbook edit is authored in `cli/docs/embedded/playbooks/` and re-projected, so `TestProjectionDrift_RootDocsMatchEmbed` stays green.
- R12 — No schema migration, no new `zharness` subcommand, no `MIN_ZHARNESS_VERSION` bump.

## Non-goals

- **Structured trace fields** — changing `trace add` to carry tool-call outcomes needs a schema migration and has no demonstrated demand yet. Revisit only if the ledger shows trace-blindness causing repeated misses.
- **Confidence score / autonomous merge** — requires historical revert rate and eval trajectory, neither of which exists. Premature.
- **Spawning a different-family judge from inside the playbook** — breaks the portability contract (`skills/workflow/README.md`: any agent that can read a file and run a CLI runs the same lifecycle). This plan makes the blind spot *visible*; it does not mandate a mechanism.
- **Adopting `langchain-skills` / `Agent-EvalKit`** — those evaluate LLM applications. This repository ships instruction files. Wrong instrument.
- **Verifying `zharness` command claims** — audited clean; the Go CLI already has its own CI. Building a gate for a defect class with zero instances is speculative work.
- **Fixing the 32 illustrative-path hits** — they are correct as written.

## Approach and Risks

**Chosen approach: deterministic first, declaration second, memory third.**

Split by the same rule the source material gets right — a judge decides semantic questions, plain code decides objective ones. Almost every defect found in the audit is objective (does this path resolve), so it is solved by a shell script that cannot be wrong, not by a model that can. Only the genuinely semantic gap (who judged, and were they independent) is handled by making the agent declare it.

Phase order is by evidence strength, not size: `link-integrity` has 11 proven instances, `judge-hygiene` is an hour of text with a real structural payoff, `regression-ledger` is the compounding layer that only pays off over time.

**Rejected alternative — one combined "eval" phase.** It would touch `scripts/`, `CLAUDE.md`, the embedded playbook, the projection, and 8 doc files in a single reviewable unit, and a failure anywhere would block all of it. The three phases here each merge alone and each leave the repository better on its own.

**Rejected alternative — model-graded doc review.** A judge pass over every changed doc would catch more classes but is non-deterministic, slow, and unverifiable. The audit shows the highest-frequency defect is trivially machine-checkable; spending a model on it would be strictly worse.

**Risks**

| Risk | Mitigation | Recovery |
|---|---|---|
| Ignore list becomes a dumping ground and the gate stops detecting anything | Every `.claimignore` entry requires a trailing `# reason` comment; the script refuses to run if any non-comment line lacks one | Delete unjustified entries; the gate re-fires on the next run |
| Gate produces false positives on new legitimate example docs | Allowlist is prefix-based (`skills/ docs/ rules/ cli/ setup/ references/`), so anything outside is skipped by default rather than flagged | Add a justified `.claimignore` line |
| Playbook edit lands in the projection but not the embedded source, or vice versa | R11 makes the embedded file the only authoring surface; `go test ./...` in `cli/` fails on drift | Re-run `zharness init --refresh-docs` |
| Phase 3 appends to a file during `review` mode and breaks the zero-write rule | R9 is stated as an explicit precondition on the step, and Phase 3 wave 3 tests `review` for zero writes | Revert the step-10 edit; the ledger is read-only until fixed |
| Ledger is written but never read, becoming dead documentation | R8 wires reading into step 4 of the gate, so a gate run that ignores it is itself a finding | If unread after 3 real gates, kill `docs/evals/` and record why |

**Stop conditions**

- Phase 1: if the script's known-bad control does not fail, stop — the gate is measuring nothing.
- Phase 2: if `go test ./...` in `cli/` fails on projection drift, stop and re-project before continuing.
- Phase 3: if `check review` produces any file write, stop and revert the step-10 edit.

**Premise**

This plan assumes stale and broken cross-references are a recurring cost in this repository, not a one-time accumulation. Evidence for recurrence: `skills/workflow/README.md` points at a plan that a *previous completed initiative* moved. If the next two gate runs after Phase 1 find zero new instances, Phase 1 was a cleanup rather than a gate, and Phase 3's ledger should record that verdict.

## Phases and Verification

### Phase 1 — `link-integrity`

- **Story ID**: `01KYS4F41SPGQSP5HZKP1S8XT6`
- **Status**: `checked`
- **Depends on**: none
- **Goal**: A deterministic gate fails on broken repo-relative doc cross-references, with an explicit ignore convention; all 11 known instances repaired.
- **Touched surfaces**: `scripts/verify-doc-links.sh` (new), `.claimignore` (new), `CLAUDE.md` (Development Commands section), the 8 doc files holding the 11 broken references.
- **Avoided surfaces**: `cli/**`, `docs/playbooks/**`, `cli/docs/embedded/**`, `harness.db`, `.kit/**`.

**Wave 1 — build the gate**

| Task | Detail | Verification |
|---|---|---|
| T1.1 | Write `scripts/verify-doc-links.sh`. Build the file list with `find docs skills rules setup -name '*.md' -type f` plus root `CLAUDE.md` and `README.md`, into a temp file consumed via `xargs -a` — never a shell variable, which silently produced a false clean during the audit. For each file, extract backtick-quoted tokens matching `[A-Za-z0-9._/-]+/[A-Za-z0-9._/-]+\.(md\|sh\|go\|json\|yml\|toml\|py)`, drop any containing `{`, `}` or `*`, and keep only paths whose first segment is one of `skills docs rules cli setup references`. Report a finding when the path resolves neither from the repository root nor from the referencing file's directory. Print `file -> path` per finding, exit 1 if any. | `bash scripts/verify-doc-links.sh; echo "exit=$?"` runs and prints a finding list |
| T1.2 | Add `.claimignore` at the repository root: one `substring  # reason` line per justified exception, comments with `#`. Seed with exactly the 5 known illustrative references: `references/api-endpoints-auth.md`, `references/api-endpoints-payments.md`, `references/api-endpoints-users.md`, `references/api-patterns.md`, `references/ffmpeg-encoding.md`. Make the script exit 2 with a named error if any non-comment line lacks a `#` reason. | `printf 'references/foo.md\n' >> .claimignore && bash scripts/verify-doc-links.sh; echo "exit=$?"` prints exit=2, then revert the line |
| T1.3 | Test the test against a known-bad and a known-good control before trusting either result. Known-bad: a scratch file under the scanned tree referencing `docs/definitely-not-here.md`. Known-good: the same file removed. | known-bad run exits 1 and names `docs/definitely-not-here.md`; known-good run does not name it |

**Wave 2 — repair and wire**

| Task | Detail | Verification |
|---|---|---|
| T2.1 | Repair all 11 broken references listed in Authority and Requirements. The 8 `references/`-prefixed ones are wrong-prefix links written from inside `references/` — drop the prefix. `eval-queries.md` → `skills/craft/write/SKILL.md`. `skills/workflow/README.md` → `docs/plans/completed/harness-convergence-pass-v3.md`. | `bash scripts/verify-doc-links.sh; echo "exit=$?"` prints exit=0 with zero findings |
| T2.2 | Declare the gate in `CLAUDE.md` under Development Commands as a required pre-commit gate command, satisfying R3 without touching any shipped playbook. | `grep -n 'verify-doc-links' CLAUDE.md` returns a line |
| T2.3 | Confirm no shipped surface was touched. | `git diff --name-only` contains no path under `cli/` or `docs/playbooks/` |

**Expected output**: a green deterministic link gate, 11 repairs, one justified ignore file, one line in `CLAUDE.md`.

**Exit / escalation**: T1.3 controls must both behave as specified before Wave 2 starts. If the known-bad control passes, stop with `BLOCKED_VERIFICATION` — the gate is measuring nothing and repairing links would hide that.

---

### Phase 2 — `judge-hygiene`

- **Story ID**: `01KYS4F4209JX12J99XGTNXVE7`
- **Status**: `checked`
- **Depends on**: none
- **Goal**: Every check verdict discloses judge independence and judge model version.
- **Touched surfaces**: `cli/docs/embedded/playbooks/check.md`, `docs/playbooks/check.md` (projection only, via refresh).
- **Avoided surfaces**: `scripts/**`, `cli/internal/**`, `cli/cmd/**`, schema, any other playbook, `docs/evals/**`.

**Wave 1 — author the playbook change**

| Task | Detail | Verification |
|---|---|---|
| T1.1 | In `cli/docs/embedded/playbooks/check.md` Output Format block, add two lines after `review:` — `judge: independent \| same-session` and `judge_model: {model identifier}`. `same-session` means the reviewing agent also authored the diff under review. | `grep -n 'judge_model' cli/docs/embedded/playbooks/check.md` returns a line |
| T1.2 | In the same file, extend step 8 (choose the verdict) with one sentence: when `judge: same-session`, an `APPROVED` or `APPROVE_WITH_REQUESTS` verdict must name at least one aspect that was not independently verified. Do not add a mechanism, a tool, or a subagent requirement — declaration only, so the playbook stays runnable by any file-reading, CLI-running agent. | `grep -n 'same-session' cli/docs/embedded/playbooks/check.md` returns at least 2 lines |
| T1.3 | Extend the Exit Conditions section so gate/full completion requires both new fields to be present. | `grep -c 'judge' cli/docs/embedded/playbooks/check.md` returns 4 or more |

**Wave 2 — project and prove**

| Task | Detail | Verification |
|---|---|---|
| T2.1 | Re-project managed docs so the root playbook matches the embedded source. | `zharness init --refresh-docs` exits 0; `diff <(sed -n '/Output Format/,/```$/p' cli/docs/embedded/playbooks/check.md) <(sed -n '/Output Format/,/```$/p' docs/playbooks/check.md)` prints nothing |
| T2.2 | Prove the projection-drift test is green. | `cd cli && go test ./... 2>&1 \| tail -20` shows no FAIL |
| T2.3 | Run `check review` on the working diff and confirm the response block carries both new fields with honest values. | response block contains `judge: same-session` and a concrete `judge_model` value |

**Expected output**: two new declared fields on every check verdict, one added verdict constraint, projection green.

**Exit / escalation**: if `go test ./...` reports projection drift, stop with `BLOCKED_VERIFICATION` and re-run the refresh before any further edit.

---

### Phase 3 — `regression-ledger`

- **Story ID**: `01KYS4F428R9GE4TGWHJYHDWH6`
- **Status**: `checked`
- **Depends on**: `judge-hygiene` (both phases edit `cli/docs/embedded/playbooks/check.md`; sequencing avoids a conflicting edit on the same file)
- **Goal**: A failure recorded once is read on every later gate, and a failure recorded twice is forced to become a deterministic check.
- **Touched surfaces**: `docs/evals/failures.md` (new), `cli/docs/embedded/playbooks/check.md`, `docs/playbooks/check.md` (projection only).
- **Avoided surfaces**: `scripts/**`, `cli/internal/**`, `cli/cmd/**`, schema, all other playbooks.

**Wave 1 — create the ledger**

| Task | Detail | Verification |
|---|---|---|
| T1.1 | Create `docs/evals/failures.md` with the header row `date \| phase-slug \| failure class \| how it was caught \| permanent check` and a short preamble stating the file is append-only and that a class appearing twice must graduate into a deterministic check under `scripts/`. | `test -f docs/evals/failures.md && head -1 docs/evals/failures.md` |
| T1.2 | Seed one real row from this initiative's own audit, not a synthetic example: class `broken-doc-cross-reference`, caught by the 2026-07-30 repository audit, permanent check `scripts/verify-doc-links.sh`. | `grep -c 'broken-doc-cross-reference' docs/evals/failures.md` returns 1 |

**Wave 2 — wire reading and writing**

| Task | Detail | Verification |
|---|---|---|
| T2.1 | In `cli/docs/embedded/playbooks/check.md` step 4, add: when `docs/evals/failures.md` exists, read it, and for every failure class recorded two or more times state explicitly whether the current diff is clean of that class. Phrase the step so an absent file is not an error, keeping the playbook portable. | `grep -n 'failures.md' cli/docs/embedded/playbooks/check.md` returns a line inside step 4 |
| T2.2 | In step 10's `REQUEST_CHANGES` branch, add: append one ledger row per finding. State on the step that this applies to durable gate/full only, so `review` and bounded keep the zero-write rule in the playbook's own Zero-write section. | `grep -n 'gate/full only' cli/docs/embedded/playbooks/check.md` returns a line in step 10 |
| T2.3 | Re-project and prove projection is green. | `zharness init --refresh-docs` exits 0; `cd cli && go test ./... 2>&1 \| tail -20` shows no FAIL |

**Wave 3 — prove both directions**

| Task | Detail | Verification |
|---|---|---|
| T3.1 | Append a second `broken-doc-cross-reference` row to the ledger, run a durable gate on a real diff, and confirm the output names that class as recurring. | gate response explicitly names `broken-doc-cross-reference` and states clean or not-clean |
| T3.2 | Run `check review` on the same diff and confirm zero writes. | `git status --porcelain docs/evals/` prints nothing after the run |
| T3.3 | Confirm the graduation rule is stated in the ledger preamble and that the seeded class already points at its deterministic check. | `grep -n 'scripts/verify-doc-links.sh' docs/evals/failures.md` returns a line |

**Expected output**: an append-only ledger read by every durable gate, written by every `REQUEST_CHANGES`, with the zero-write rule proven intact.

**Exit / escalation**: if T3.2 shows any write during `review`, stop with `BLOCKED_CONTRACT_DRIFT` and revert T2.2 before anything else.

## Progress

- `2026-07-30T09:12Z` — phase `link-integrity` started. wave: —. task: phase-start. task_status: `in-progress`. run: `01KYS4NST8ACHAJAC9S5V12PBF`. changed surfaces: none yet. verification: `zharness query phases --json` → `link-integrity: in-progress`.
- `2026-07-30T09:16Z` — wave 1, task T1.1. task_status: `DONE`. run: `01KYS4NST8ACHAJAC9S5V12PBF`. changed surfaces: `scripts/verify-doc-links.sh` (new). verification: `bash scripts/verify-doc-links.sh; echo "exit=$?"` → 11 findings, `exit=1`.
- `2026-07-30T09:16Z` — wave 1, task T1.2. task_status: `DONE`. run: `01KYS4NST8ACHAJAC9S5V12PBF`. changed surfaces: `.claimignore` (new, 5 seeded exceptions). verification: appended a reason-less line → `ERROR: .claimignore:11 has no \`# reason\` comment`, `exit=2`; line reverted, gate back to `exit=1` with 11 findings.
- `2026-07-30T09:16Z` — wave 1, task T1.3. task_status: `DONE_WITH_CONCERNS`. run: `01KYS4NST8ACHAJAC9S5V12PBF`. changed surfaces: none (controls created and removed). verification: known-bad `rules/_control-bad.md` referencing `docs/definitely-not-here.md` → detected (count 1); control removed → not reported (count 0). concern: first real run returned 25 findings, not the 11 predicted at planning time — resolved by decision D1 below, not by relaxing the gate.
- `2026-07-30T09:17Z` — wave 1 complete. trace: `01KYS4SKZG39631B23ACF3V2D1`. run: `01KYS4NST8ACHAJAC9S5V12PBF`. verification: 4/4 controls pass; gate at documented baseline of 11 findings.
- `2026-07-30T09:21Z` — wave 2, task T2.1. task_status: `DONE`. run: `01KYS4NST8ACHAJAC9S5V12PBF`. changed surfaces: `skills/craft/create-skill/references/plugin-marketplace-overview.md`, `skills/craft/create-skill/references/skill-anatomy-and-requirements.md`, `skills/craft/create-skill/references/skill-creation-workflow.md`, `skills/craft/write/references/write-vi-notion-report.md`, `skills/craft/write/examples/eval-queries.md`, `skills/shipping/turbo-mono-platform/references/companion-skills.md`, `skills/workflow/README.md`. verification: `bash scripts/verify-doc-links.sh` → `doc links OK (0 findings)`, `exit=0`.
- `2026-07-30T09:22Z` — wave 2, task T2.2. task_status: `DONE`. run: `01KYS4NST8ACHAJAC9S5V12PBF`. changed surfaces: `CLAUDE.md` (new `## Gate Commands` section). verification: `grep -n 'verify-doc-links' CLAUDE.md` → line 76.
- `2026-07-30T09:22Z` — wave 2, task T2.3. task_status: `DONE`. run: `01KYS4NST8ACHAJAC9S5V12PBF`. changed surfaces: none. verification: `git status --porcelain | awk '{print $2}' | grep -E '^(cli/|docs/playbooks/)'` → `none`.
- `2026-07-30T09:22Z` — wave 2 complete. trace: `01KYS4W94KXRFJF9T99Y1R5S8N`. run: `01KYS4NST8ACHAJAC9S5V12PBF`. verification: gate green at `exit=0`, surface boundary clean.
- `2026-07-30T09:36Z` — phase `judge-hygiene` started. wave: —. task: phase-start. task_status: `in-progress`. run: `01KYS5FQ8MWHFJK0F6C4XB0J15`. changed surfaces: none yet. verification: `zharness query phases --json` → `judge-hygiene: planned` → set `in-progress`. note: phase 1's approved diff is still uncommitted and touches `scripts/verify-doc-links.sh`, a phase-2 avoided surface; that is checked prior-phase work, not drift — phase 2's own diff is scoped to `cli/docs/embedded/playbooks/check.md` and its projection.
- `2026-07-30T09:38Z` — wave 1, task T1.1. task_status: `DONE`. run: `01KYS5FQ8MWHFJK0F6C4XB0J15`. changed surfaces: `cli/docs/embedded/playbooks/check.md` (Output Format block). verification: `grep -n 'judge_model' cli/docs/embedded/playbooks/check.md` → line 70.
- `2026-07-30T09:38Z` — wave 1, task T1.2. task_status: `DONE`. run: `01KYS5FQ8MWHFJK0F6C4XB0J15`. changed surfaces: `cli/docs/embedded/playbooks/check.md` (step 8). verification: `grep -n 'same-session' cli/docs/embedded/playbooks/check.md` → 3 lines (40, 69, 79), ≥2 required. The `same-session` definition was placed in step 8 prose rather than inside the output block, so the block stays a field list; no mechanism, tool, or subagent was introduced.
- `2026-07-30T09:38Z` — wave 1, task T1.3. task_status: `DONE`. run: `01KYS5FQ8MWHFJK0F6C4XB0J15`. changed surfaces: `cli/docs/embedded/playbooks/check.md` (Exit Conditions, Gate bullet). verification: `grep -c 'judge' cli/docs/embedded/playbooks/check.md` → 4, ≥4 required.
- `2026-07-30T09:39Z` — wave 1 complete. trace: `01KYS5HHGWG6HY3JW3DN6MEE5Z`. run: `01KYS5FQ8MWHFJK0F6C4XB0J15`. verification: 3/3 task greps pass; root `docs/playbooks/check.md` not yet re-projected — drift is expected until wave 2 T2.1.
- `2026-07-30T09:44Z` — wave 2, task T2.1. task_status: `DONE_WITH_CONCERNS`. run: `01KYS5FQ8MWHFJK0F6C4XB0J15`. changed surfaces: `docs/playbooks/check.md` (projection only). verification: after building a dev binary from source, `zharness init --refresh-docs` → `scaffolded managed docs (docs_version=dev)`, `exit=0`; `diff` of the Output Format block → empty; full-file `diff cli/docs/embedded/playbooks/check.md docs/playbooks/check.md` → empty, `exit=0`. concern: the installed `zharness 0.6.0` could not perform this projection — see decision D2.
- `2026-07-30T09:45Z` — wave 2, task T2.2. task_status: `DONE`. run: `01KYS5FQ8MWHFJK0F6C4XB0J15`. changed surfaces: none. verification: `cd cli && go test ./... 2>&1 | tail -20` → every package `ok`, no `FAIL`; `TestProjectionDrift_RootDocsMatchEmbed` green in `internal/embedded`.
- `2026-07-30T09:47Z` — wave 2, task T2.3. task_status: `DONE`. run: `01KYS5FQ8MWHFJK0F6C4XB0J15`. changed surfaces: none (`check review` is response-only). verification: `zharness preflight check --mode review --json` → `mode: reduced`; `bash scripts/verify-doc-links.sh` → `exit=0`; response block carried `judge: same-session` and `judge_model: claude-opus-5`, and the `APPROVED` verdict named the unverified aspect as the new rule requires. Zero writes: no `check record`, no plan Validation append, no changeset.
- `2026-07-30T09:48Z` — wave 2 complete. trace: `01KYS5NPEJC6H0PRD6B9WYMAFX`. run: `01KYS5FQ8MWHFJK0F6C4XB0J15`. verification: `git diff --name-only -- cli docs/playbooks` → exactly `cli/docs/embedded/playbooks/check.md` and `docs/playbooks/check.md`; no path under `cli/internal/`, `cli/cmd/`, `scripts/`, or `docs/evals/`; no managed doc other than `check.md` re-projected.
- `2026-07-30T10:14Z` — phase `regression-ledger` started. wave: —. task: phase-start. task_status: `in-progress`. run: `01KYS7VW1D8GQA4X5CEMHM3TFM`. changed surfaces: none yet. verification: `git status --porcelain` → clean (phases 1-2 pushed as `41d8278..82b1328`); `zharness query phases --json` → `regression-ledger: planned` → set `in-progress`. note: dependency `judge-hygiene` is `checked`, which is as far as a dependency can advance before `handoff` — `check` never marks a phase `done`.
- `2026-07-30T10:17Z` — wave 1, task T1.1. task_status: `DONE`. run: `01KYS7VW1D8GQA4X5CEMHM3TFM`. changed surfaces: `docs/evals/failures.md` (new). verification: `test -f docs/evals/failures.md && head -1` → `# Failure Ledger`, `exit=0`. Preamble states append-only, the graduation rule, the `none yet` placeholder for an ungraduated class, and that the class name must stay stable across rows.
- `2026-07-30T10:17Z` — wave 1, task T1.2. task_status: `DONE`. run: `01KYS7VW1D8GQA4X5CEMHM3TFM`. changed surfaces: `docs/evals/failures.md`. verification: `grep -c 'broken-doc-cross-reference' docs/evals/failures.md` → `1`. The seeded row is real audit history, not a synthetic example: 11 unresolvable references across 113 tracked markdown files, graduated to `scripts/verify-doc-links.sh`.
- `2026-07-30T10:18Z` — wave 1 complete. trace: `01KYS7XFCN7CDGZ4E10WXC5TCK`. run: `01KYS7VW1D8GQA4X5CEMHM3TFM`. verification: both task checks pass; `bash scripts/verify-doc-links.sh` still `exit=0` with the new doc in the scanned tree.
- `2026-07-30T10:22Z` — wave 2, task T2.1. task_status: `DONE`. run: `01KYS7VW1D8GQA4X5CEMHM3TFM`. changed surfaces: `cli/docs/embedded/playbooks/check.md` (step 4). verification: `grep -n 'failures.md' cli/docs/embedded/playbooks/check.md` → line 36, inside step 4. Phrased so an absent ledger is skipped rather than treated as an error, keeping the playbook portable to repositories without `docs/evals/`.
- `2026-07-30T10:22Z` — wave 2, task T2.2. task_status: `DONE`. run: `01KYS7VW1D8GQA4X5CEMHM3TFM`. changed surfaces: `cli/docs/embedded/playbooks/check.md` (step 10, `REQUEST_CHANGES` branch). verification: `grep -n 'gate/full only' cli/docs/embedded/playbooks/check.md` → line 44, inside step 10. The clause points back at the playbook's own Zero-write section rather than restating it, so `review` and bounded keep one authority for that rule.
- `2026-07-30T10:23Z` — wave 2, task T2.3. task_status: `DONE`. run: `01KYS7VW1D8GQA4X5CEMHM3TFM`. changed surfaces: `docs/playbooks/check.md` (projection only). verification: source-built dev binary per decision D2 → `zharness init --refresh-docs` `exit=0`; `diff cli/docs/embedded/playbooks/check.md docs/playbooks/check.md` → empty, `exit=0`; `cd cli && go test ./... 2>&1 | tail -20` → every package `ok`, no `FAIL`.
- `2026-07-30T10:23Z` — wave 2 complete. trace: `01KYS7ZE25Q8N2VJAJRVKZ0NZY`. run: `01KYS7VW1D8GQA4X5CEMHM3TFM`. verification: 3/3 task checks pass; projection byte-identical and drift test green.
- `2026-07-30T10:29Z` — wave 3, task T3.1 (part A: append). task_status: `DONE_WITH_CONCERNS`. run: `01KYS7VW1D8GQA4X5CEMHM3TFM`. changed surfaces: `docs/evals/failures.md`. verification: `grep -c 'broken-doc-cross-reference' docs/evals/failures.md` → `2`; mechanical recurrence scan over the class column → `broken-doc-cross-reference x2`. concern: the first draft of the row broke the link gate — see decision D3. Part B (a durable gate naming the class as recurring) is discharged by this phase's own `check full` and recorded in Validation, not here.
- `2026-07-30T10:31Z` — wave 3, task T3.2. task_status: `DONE`. run: `01KYS7VW1D8GQA4X5CEMHM3TFM`. changed surfaces: none. verification: captured `sha256sum docs/evals/failures.md` plus `git status --porcelain docs/evals/` before and after a full `check review` pass (preflight `mode: reduced`, link gate, `go test ./...`, projection diff, recurring-class scan). Both snapshots identical — `aaf90399…a30b209`, `?? docs/evals/` — so `review` wrote nothing. Deviation from the literal task text: the plan's check was `git status --porcelain docs/evals/` printing nothing, which cannot hold while the ledger is still untracked; a before/after comparison of that same command plus a content hash tests the same property more strictly and works at any tracking state.
- `2026-07-30T10:31Z` — wave 3, task T3.3. task_status: `DONE`. run: `01KYS7VW1D8GQA4X5CEMHM3TFM`. changed surfaces: none. verification: `grep -n 'Graduation rule' docs/evals/failures.md` → line 5; `grep -c 'scripts/verify-doc-links.sh' docs/evals/failures.md` → `2`, so both recorded rows of the graduated class point at their deterministic check.
- `2026-07-30T10:33Z` — wave 3 complete. trace: `01KYS84CTE8X9VE749N7PN7WG5`. run: `01KYS7VW1D8GQA4X5CEMHM3TFM`. verification: read path proven (recurrence scan surfaces `broken-doc-cross-reference x2`), zero-write path proven (identical hash and git status across a full `check review`), graduation rule and its deterministic check both present. Remaining for the phase gate: T3.1 part B, a durable gate that names the recurring class.

## Decisions

- **D1 — `docs/plans/**` is excluded from the scan by category, not by `.claimignore` exception.** Affected phase/task: `link-integrity` / T1.1.
  - Discovered during T1.3: the first real run returned 25 findings instead of the 11 recorded in Authority and Requirements. 9 came from this plan itself (it quotes broken paths as evidence and names `docs/evals/failures.md`, a file phase 3 will create) and 5 from `docs/plans/completed/**` (immutable historical records naming docs as they existed then).
  - Rationale: a plan artifact must be able to name a file it will create, and a completed plan is a record rather than a live cross-reference. Silencing 14 findings through `.claimignore` would have been a per-instance fudge that grows with every future plan; a categorical `-not -path 'docs/plans/*'` in the file list is correct once and stays correct.
  - Result: the gate returns exactly the 11 findings predicted at planning time. Requirements R1 and R2 are unchanged; only the scan scope is narrowed.

- **D2 — projecting an embedded-doc edit requires rebuilding the binary first; the plan's T2.1 assumed the installed binary would suffice.** Affected phase/task: `judge-hygiene` / T2.1.
  - Discovered during T2.1: `zharness init --refresh-docs` on the installed `zharness 0.6.0` printed `exists harness.db` and exited 0 without writing anything. The embedded doc set is compiled into the binary via `go:embed`, so a binary built on 2026-07-28 still carries the pre-edit `check.md` and had nothing new to project. `TestProjectionDrift_RootDocsMatchEmbed`, by contrast, reads the embed from source at compile time, so it saw the edit and would have failed — the two disagreed precisely because one is a binary artifact and the other is source.
  - Rationale: built a dev binary from source into the session scratchpad (`CGO_ENABLED=0 go build ./cmd/zharness`) and ran `--refresh-docs` with that instead of reinstalling over the user's `~/.local/bin/zharness`. The projection is a repository artifact; replacing the installed binary is a release action, belongs to the release path, and was not requested by this phase. Avoided surfaces `cli/internal/**` and `cli/cmd/**` were read but not modified.
  - Result: `docs/playbooks/check.md` is byte-identical to its embedded source and `go test ./...` is green. Consequence recorded as open item O2: until the next release, the installed `0.6.0` binary carries a stale embedded `check.md`, so running `zharness init --refresh-docs` with it would revert the projection. The drift test catches that, so it fails loudly rather than silently.

- **D3 — the ledger records history, and the link gate cannot tell history from a live reference; the ledger's prose gives way, not the gate.** Affected phase/task: `regression-ledger` / T3.1.
  - Discovered during T3.1: the first draft of the two seeded rows quoted the historical and illustrative paths in backticks. `bash scripts/verify-doc-links.sh` immediately returned 3 findings — `references/x.md`, `references/references/x.md`, and `skills/write/SKILL.md` — none of which are live references. The gate was right; the prose was wrong.
  - Rationale: three fixes were available and two were out of authority. Excluding `docs/evals/**` categorically means editing `scripts/verify-doc-links.sh`, an avoided surface for this phase; adding `.claimignore` lines means editing phase 1's artifact, also outside this phase's touched surfaces. Rewording the rows to describe paths rather than quote them stays entirely inside `docs/evals/failures.md`, loses no meaning, and keeps the gate at full strength. A ledger that has to weaken the gate to describe a link failure would be self-defeating.
  - Result: `bash scripts/verify-doc-links.sh` → `exit=0`, and the class is still recorded twice with both rows pointing at their deterministic check. Recorded as open item O3: if the ledger later accumulates many historical paths, the correct fix is a categorical `docs/evals/**` exclusion mirroring D1, which belongs to a phase that owns `scripts/`.

## Validation

- `2026-07-30T09:30Z` — `check full` on phase `link-integrity`. verdict: **APPROVED**. check: `01KYS53W8AQP03RNFK40FRN0XM`. run: `01KYS4NST8ACHAJAC9S5V12PBF`. judge: `same-session` (R5/R6 land in phase 2; declared here voluntarily).
  - Gate 1 — `bash scripts/verify-doc-links.sh` → `doc links OK (0 findings)`, `exit=0`.
  - Gate 2 — `cd cli && go test ./...` → every package `ok`, `exit=0`; includes `TestProjectionDrift_RootDocsMatchEmbed`.
  - `bash -n scripts/verify-doc-links.sh` → `exit=0`. `shellcheck` not installed on this machine; not run.
  - Test-the-test — known-bad control detected (`exit=1`), known-good control clean (`exit=0`), reason-less `.claimignore` line rejected (`exit=2`). Satisfies the phase exit condition.
  - Manual existence check of all 11 repaired targets → 11/11 present on disk.
  - `zharness audit --json` → `pointer_drift: []`, `contract_violations: []`, `unlinked_proofs: []`.
  - Surface boundary — `git diff --name-only` contains no path under `cli/` or `docs/playbooks/`; R11/R12 untouched.
  - Required-proof matrix, lane `normal` (unit + command output): unit satisfied by the T1.3 controls, the `.claimignore` `exit=2` control, and `go test ./...`; command output satisfied by both gate runs. No class missing.
  - **Not independently verified** (same-session judge): the gate script was written and reviewed by the same session, so its allowlist and regex were never read by an outside evaluator; the 11-finding baseline was reproduced by that same script rather than by a second, independently written scanner.

- `2026-07-30T09:52Z` — `check full` on phase `judge-hygiene`. verdict: **APPROVED**. check: `01KYS5QNPGZ1DPFVB75Y9MZR3V`. run: `01KYS5FQ8MWHFJK0F6C4XB0J15`. judge: `same-session`. judge_model: `claude-opus-5`. First gate run under the rule this phase introduced.
  - Gate 1 — `bash scripts/verify-doc-links.sh` → `doc links OK (0 findings)`, `exit=0`.
  - Gate 2 — `cd cli && go test ./...` → every package `ok`, `exit=0`; `TestProjectionDrift_RootDocsMatchEmbed` green, which is the unit test that actually guards this change.
  - Projection — `diff cli/docs/embedded/playbooks/check.md docs/playbooks/check.md` → empty, `exit=0`.
  - Task greps — `judge_model` at line 70; `same-session` on 3 lines (≥2 required); `judge` count 4 (≥4 required).
  - Surface boundary — `git diff --name-only -- cli docs/playbooks` → exactly the two touched-surface files; no path under `cli/internal/`, `cli/cmd/`, `scripts/`, or `docs/evals/`. `check.md` is the only managed doc re-projected.
  - Sibling-instance sweep — `check.md` is the only playbook with an Output Format block, and no Go code or other doc parses these fields (the `proof_gaps` hits in `cli/**_test.go` belong to the plan-template Validation contract, untouched). Coverage is complete; R12 holds.
  - Lifecycle audit — `zharness audit --json` reported one `out_of_order` pointer drift whose stated recovery is "record a new check for the latest run", i.e. this gate itself. `contract_violations: []`, `unlinked_proofs: []`. Cleared by the record above.
  - Required-proof matrix, lane `normal` (unit + command output): unit satisfied by the projection-drift test; command output satisfied by the gate runs, projection diff, and task greps. No class missing.
  - **Not independently verified** (same-session judge): the playbook wording was authored and reviewed in the same session, so no outside evaluator read the new step-8 sentence for ambiguity. More fundamentally, `judge:` is a self-declared field with no mechanism behind it — nothing in this diff can detect an agent that writes `independent` while reviewing its own work. That is the deliberate cost of keeping the playbook portable, recorded here rather than hidden.

- `2026-07-30T10:36Z` — `check full` on phase `regression-ledger`. verdict: **APPROVED**. check: `01KYS860JK6Z53EH14Q27XYJBM`. run: `01KYS7VW1D8GQA4X5CEMHM3TFM`. judge: `same-session`. judge_model: `claude-opus-5`.
  - Gate 1 — `bash scripts/verify-doc-links.sh` → `doc links OK (0 findings)`, `exit=0`.
  - Gate 2 — `cd cli && go test ./...` → every package `ok`, `exit=0`; `TestProjectionDrift_RootDocsMatchEmbed` green.
  - **Recurring-class statement (the new step 4, first real exercise)** — the ledger records `broken-doc-cross-reference` twice. This diff **is clean of that class**: `scripts/verify-doc-links.sh`, the deterministic check the class graduated into, returns zero findings over the whole tracked doc tree including the new ledger. This discharges T3.1 part B.
  - Zero-write proof — `sha256sum` and `git status --porcelain docs/evals/` captured before and after a complete `check review` pass were byte-identical (`aaf90399…a30b209`, `?? docs/evals/`). R9 holds: `review` reads the ledger and writes nothing.
  - Projection — `diff cli/docs/embedded/playbooks/check.md docs/playbooks/check.md` → empty, `exit=0`. R11 holds.
  - Surface boundary — changed paths are `docs/evals/failures.md` (new), `cli/docs/embedded/playbooks/check.md`, `docs/playbooks/check.md`, plus this plan. No path under `scripts/`, `cli/internal/`, `cli/cmd/`, or any other playbook. R12 holds: no schema migration, no new subcommand, no version bump.
  - Lifecycle audit — `zharness audit --json` reported the expected pre-record `out_of_order` pointer drift whose stated recovery is this gate itself; `contract_violations: []`, `unlinked_proofs: []`.
  - Required-proof matrix, lane `normal` (unit + command output): unit satisfied by the projection-drift test; command output satisfied by both gates, the recurrence scan, the before/after hash comparison, and the step-4/step-10 greps. No class missing.
  - **Not independently verified** (same-session judge): the same session wrote the ledger, wrote the playbook clauses that read and write it, and then judged whether both work — no outside evaluator checked that the step-4 wording actually forces the statement rather than merely inviting it. Two further gaps are structural, not oversights: the ledger is append-only by convention only, with nothing preventing a later agent from rewriting or deleting a row; and the recurrence count here was produced by an ad-hoc `awk` scan written for this gate, not by a shipped parser, so a future gate reading the table by eye could miscount. The scan also treats the markdown separator row as a class, which is harmless against the two-or-more rule but shows the format is not machine-clean.

## Current State and Next Action

- **Active phase**: `regression-ledger`
- **Lifecycle status**: `checked`
- **Latest run ID**: `01KYS7VW1D8GQA4X5CEMHM3TFM`
- **Latest check ID**: `01KYS860JK6Z53EH14Q27XYJBM` (phase `regression-ledger`, APPROVED)
- **Latest trace ID**: `01KYS84CTE8X9VE749N7PN7WG5` (phase `regression-ledger`, wave 3)
- **Blockers**: none
- **Open items**:
  - **O1 — the link gate's path regex requires a `/`, so sibling-file references are not covered.** Nine of the eleven repairs dropped the wrong `references/` prefix, which moved those links out of gate coverage; they were verified by hand in phase 1, but a future rename would not be caught. Spec-conformant (T1.1 defines that regex), non-blocking, and a candidate for the phase 3 ledger as a first-recorded failure class.
  - **O2 — the installed `zharness 0.6.0` carries a stale embedded `check.md` until the next release.** Running `zharness init --refresh-docs` with it would revert this phase's projection. `TestProjectionDrift_RootDocsMatchEmbed` catches that loudly, so it is a nuisance rather than a silent regression. See decision D2.
  - **O3 — the ledger will keep colliding with the link gate as it accumulates historical paths.** D3 solved this instance by rewording prose, which does not scale: every future row describing a path that has since moved reopens the same conflict. The categorical fix is to exclude `docs/evals/**` from the link gate the way D1 excludes `docs/plans/**` — both hold historical rather than live references. That edit belongs to `scripts/verify-doc-links.sh`, an avoided surface for this phase, so it is deferred to whatever initiative next owns `scripts/`.
  - **O4 — the two-clean-runs premise is now decided.** The plan predicted that if the two gate runs after phase 1 found zero new link failures, phase 1 was a cleanup rather than a gate. Both runs (`judge-hygiene`, `regression-ledger`) came back clean. The honest reading is the weaker one: the gate has not yet caught a failure it did not itself create the baseline for, so its value is still prospective — it prevents reintroduction rather than having proven detection. That is the correct expectation for a regression gate and not a reason to remove it, but it should not be claimed as evidence the gate works until it fails on a diff written without it in mind.
- **Next action**: `git` to commit phase `regression-ledger`, then `handoff` to close the initiative.
