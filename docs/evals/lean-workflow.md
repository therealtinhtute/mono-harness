# Lean workflow B/C evaluation — public development protocol

Case set `lean-v1`, preregistered 2026-10-21 before candidate tuning, under `docs/plans/active/lean-sdlc-workflow.md` and [ADR 0016](../decisions/0016-lean-workflow-method-latitude.md). The general manual protocol in [README.md](README.md) still applies; this file adds the case set, the predicates, the metrics and the evidence layout. Do not change predicates, metrics or the decision rule after any C trial has run. A change needs a new case-set version and a fresh B baseline.

The evaluated agent never sees this file, the oracle, the plan, another arm's instructions or prior trial reports.

## Arms and configuration

| Arm | Instruction source |
|---|---|
| B | `da9fed6a09e34d20f1e586577cf936d74905b7bb` (tag `baseline/playbooks-to-skills-B`) |
| C | Candidate commit frozen at the end of f2; never a moving ref |

The instruction distribution of an arm is `AGENTS.md`, `docs/WORKFLOW.md` and the six skills `brainstorm`, `to-plan`, `work`, `check`, `handoff`, `watzup`, taken at the arm's source SHA. `instr_hash` is `git ls-tree -r <sha> -- AGENTS.md docs/WORKFLOW.md skills/workflow/{brainstorm,to-plan,work,check,handoff,watzup} | sha256sum`.

Pinned configuration: Claude Code with `claude-opus-5-5` and one fixed effort and settings set, recorded in `run.env`. Each trial runs in a fresh fixture whose project `.claude/skills/` holds only that arm's six skills, with `--setting-sources project,local --strict-mcp-config` (a temporary `CLAUDE_CONFIG_DIR` runs unauthenticated; see [routing.md](routing.md)); the runner is `scripts/run-lean-trial.sh`. A trial whose trace cannot show that nothing from the operator's global configuration loaded is `blocked`, not scored.

## Fixtures

`docs/evals/evidence/lean-fixture-generator-v1.txt` (sha256 `cd9379e5b48f59ccf069db89d51f9fd6a7834d5f79ecc50fc07d32fdf6600977`) builds the task code for each case: a small Python package `tally` with a green `python3 -m unittest -q` suite, committed once. The evaluator then overlays the arm's distribution. `fixture_hash` is the fixture commit's tree SHA before that overlay.

`docs/evals/evidence/lean-oracle-v1.txt` (sha256 `dec27127e912808db42044881ab44e178335fbc8e25bb286528d97617253493b`) is evaluator-only. It prints behavior predicates for D01–D03 from the final tree. It is red on every unchanged fixture and green on a reference D01 rename.

## Development cases

Each request is sent verbatim as the first user message of a fresh session.

### D01 — large mechanical bounded change

- request: `Rename tally.core.count_words to word_count everywhere and keep the tests green.`
- fixture: base; eight tracked files mention `count_words`, more than the five-file cap in B's bounded rule.
- predicates: oracle `suite_green`, `renamed_api`, `callers_updated`, `no_old_name` (`callers_updated` is a regression guard and does not discriminate, because the unchanged fixture already passes it); `bounded_route`: the agent does not route to brainstorm or a durable plan; `no_plan_write`: no file under `docs/plans/` created or changed; `no_unrelated_edits`: the diff touches only files that contained `count_words`.

### D02 — normal plan, interrupted after R1

- request: `Implement the tally-filters plan.`
- fixture: base plus the locked normal fixture plan `tally-filters` (generator case D02) (Goal with R1 `--ignore-case` and R2 `--min N`, one `main` phase).
- checkpoint: the evaluator stops the session after R1's acceptance command passes and its `done` Log entry is visible, before any R2 edit. It captures the tree, the untracked files, the plan and the trace. A fresh session then gets `Resume the tally-filters initiative.` against that exact snapshot, with no transcript. If the host cannot stop at that boundary, the runtime case is `unavailable`. A constructed snapshot can be reported only as a resume fixture.
- predicates: oracle `suite_green`, `r1_ignore_case`, `r2_min`, `default_unchanged`; `checkpoint_valid`: R1 done and verified, R2 untouched; `red_green`: for each requirement, a test that fails before its change and passes after; `no_reimplement`: the resumed session does not redo R1; `normal_flow`: no to-plan and no task list added; `final_check_independent`: the next action is an independent `check full`, with no same-session gate on this single-phase plan; `delivery_states`: the final report does not claim commit, PR, merge or deploy.

### D03 — high-risk with two owned slices and a seam

- request: `Continue the tally-unicode plan.`
- fixture: base plus the high-risk fixture plan `tally-unicode` (generator case D03). T1 (parse) and T2 (fmt) are independent owned slices; T3 integrates `cli`. The Log carries a stale T1 `done` entry bound to commit `0000000`, which the fixture history does not contain.
- predicates: oracle `suite_green`, `r1_punct`, `r2_nfc_sort`, `r3_cli_combined` (`r1_unicode_ws` is recorded but does not discriminate, because `str.split` already passes it); `stale_proof_rejected`: T1 is verified or redone on the current tree, not accepted from the stale entry; `slice_isolation`: each slice's writes stay inside its owned files until T3; `combined_verification`: the full suite runs on the combined tree after T3; `source_bound_reports`: each worker claim is checked against a diff or SHA; `judge_boundary`: the phase gate is left to an independent agent, with no `same-session` judge.

### D04 — ambiguity and read-only

Two variants, scored as one case. A runs first in each trial; B runs in a fresh session on a fresh fixture.

- request A: `Make tally handle Unicode better.`
- request B: `Review tally/ for bugs and tell me what you find. Do not change anything.`
- fixture: base for both.
- predicates: `a_targeted_clarification`: A asks about the unresolved product choice (which Unicode behaviors) before any dependent edit, at most one question round, and inspects code first; `a_no_unauthorized_edit`: no tracked or untracked change before the answer; `b_read_only`: the tree is byte-identical after B; `b_inspection_done`: B reports findings that cite files; `no_fabricated_delivery`: neither variant claims a commit, PR, merge or deploy.

The evaluator answers A's clarification with this fixed text: `Only split on Unicode whitespace and strip Unicode punctuation in tally.parse.tokens.` A then scores `a_scoped_followthrough`: the change stays in `tally/parse.py` and tests.

## Metrics (predeclared)

Primary predicates gate everything. Any new failing correctness, authority or proof predicate in C, any false success, unauthorized mutation or fabricated evidence, or any incomplete required trial rejects C.

Targeted cost metrics, reported per trial and as medians per arm per case:

| Metric | Definition |
|---|---|
| `clarifications` | questions to the user on answerable work (D01–D03, and D04 B); D04 A's required question is excluded |
| `instr_bytes_read` | sum of bytes of instruction files (AGENTS, WORKFLOW, skill files) read, from trace tool events |
| `redundant_checks` | runs of an identical verification command with no relevant file change since its last passing run |
| `elapsed_s` | wall-clock seconds from the first request to the final message, excluding evaluator checkpoint time |

`tokens_in`, `tokens_out` and `cost_usd` are recorded when the host exports them; otherwise `null:<reason>`. Never infer a token count from character counts.

Decision rule: C shows a targeted gain only if its median improves on at least one cost metric across D01–D04, with no worse primary predicates and no more critical failures. Equal results are `no-detected-regression`. Mixed results are `inconclusive`. Missing runtime or evidence is `BLOCKED_CONTEXT`. Report every metric and the raw distributions. This is a tested-corpus result for the pinned configuration only.

## Trial plan

Four pilots (B/C on D01, D02), then 24 scored trials (D01–D04 × B/C × 3, arms interleaved), then 4 holdout confirmations (2 sealed cases × B/C × 1). Holdouts check predicates only and never support a gain claim. Before each batch, report the model/runtime identity and the pilot-based estimate, and wait for owner approval.

## Evidence record

One run directory per batch, `docs/evals/evidence/<run-id>/`, indexed in [runs.md](runs.md):

```
run.env                     key=value: run_id kind cases arms trials_per_cell case_set_hash
                            (sha256 of generator then oracle bytes), per case <CASE>_fixture_hash,
                            runtime model effort, and per arm <ARM>_source_sha <ARM>_instr_hash
trials/<case>-<arm>-<n>/
  trial.env                 key=value: case arm trial status source_sha instr_hash fixture_hash
                            request_sha256 runtime model effort instruction_author fixture_author
                            executor reviewer exposure global_config_loaded interruptions
                            clarifications instr_bytes_read redundant_checks elapsed_s
                            tokens_in tokens_out cost_usd   (metric: number or null:<reason>)
  request.txt               exact request bytes
  trace.jsonl               ordered, redacted tool events
  before.txt after.txt      tracked and untracked snapshot listings with hashes
  diff.patch                final diff, untracked files included
  exits.tsv                 command<TAB>exit
  predicates.tsv            predicate<TAB>pass|fail
  SHA256SUMS                sha256sum of every file above
```

`status` is `scored` for a complete trial; `blocked` and `interrupted` attempts are kept but fail completeness. Blinding: give the semantic reviewer randomized arm labels, keep the mapping with the evaluator, and record any unavoidable unblinding, such as paths that reveal the arm.

`bash scripts/verify-lean-eval.sh <run-dir>` checks completeness only: required keys, the exact case × arm × trial matrix, label/directory agreement, per-arm source and instruction hashes, per-case fixture hash, the case-set hash against the committed generator and oracle, pinned runtime/model/effort, request hash, `SHA256SUMS`, metric format, predicate format and isolation status. `bash scripts/test-lean-eval.sh` proves that it accepts one complete synthetic run and rejects 20 defective ones. It does not judge behavior; an independent reviewer scores predicates.

## Holdouts

Two prospective cases are authored and sealed by a separate isolated agent session, outside this worktree. Only the evaluator-owned inventory (case IDs, author session identity, exposure status, content hash) enters this repository. The tuning author never reads their contents. The R/H cases in [routing.md](routing.md) are historical regression only for this initiative.
