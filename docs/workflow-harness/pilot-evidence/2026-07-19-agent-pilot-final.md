# Final Second-Agent Pilot — Codex CLI, 2026-07-19

Evidence for the extended `agent-pilot-final` phase after the Phase 6 NO-GO and the #38/#39/#40 fix cycles.

**R9 acceptance criterion (SPEC.md):** a non-Claude agent completes `intake → story → trace → check record` using only written docs + CLI, without reading any `SKILL.md`, and `zharness validate --json` passes on the produced chain.

## Verdict

**GO — R9 is satisfied literally.**

The final, fully isolated Codex run:

- read **zero `SKILL.md` files** (transcript path grep returned no matches),
- received **zero harness-mechanics coaching** and asked no procedural question,
- completed the full `intake → story → registered RUN → linked trace → registered CHECK` lifecycle,
- implemented all six requested product/test files and passed **6/6 tests**,
- produced `zharness validate --json` → **`{"valid":true,...}`**,
- produced `resume --json` with drift `[]` and `audit --json` with pointer drift `[]`.

One nonblocking audit finding remains: Codex encoded two test filenames in one singular `artifact_path` string, so audit reports one `unlinked_proof`. The CHECK row is valid and linked, the latest-check pointer is correct, and validation passes. Filed separately as [#41](https://github.com/therealtinhtute/skills/issues/41); it does not negate any R9 clause.

## Isolation Controls

- **Runtime:** Codex CLI `0.144.5`, non-Claude model/runtime.
- **CLI under test:** released and freshly installed `zharness 0.4.1`.
- **Target:** brand-new scratch directory; before launch it contained only `git init` + `zharness init` output (`AGENTS.md`, `.gitignore`, `.kit/` scaffold, git metadata).
- **Codex HOME:** brand-new temporary HOME/CODEX_HOME containing `auth.json` only.
- **Invocation flags:** `--ignore-user-config --ignore-rules --ephemeral -s workspace-write --skip-git-repo-check`.
- **Prompt and transcript:** stored outside the target directory, so the agent could not inspect its own protocol/transcript files.
- **After launch:** no target edits, no user reply, no session continuation, no coaching.

Invocation shape:

```sh
HOME="$ISOLATED_HOME" CODEX_HOME="$ISOLATED_HOME/.codex" \
  codex exec --ignore-user-config --ignore-rules --ephemeral \
  -C "$TARGET" -s workspace-write --skip-git-repo-check \
  < "$PROMPT_OUTSIDE_TARGET" > "$TRANSCRIPT_OUTSIDE_TARGET" 2>&1
```

## Protocol Prompt

The prompt contained no `zharness` command, stage name, mode, threshold, phase, or acceptance-test hint. The task itself required six product/test files so `work.md`'s documented scope guard naturally selected full mode.

> This repository has an AGENTS.md file at its root. Read it first, and follow whatever process it directs for handling work in this repo. Then complete the following task.
>
> Create a standard-library-only Python inventory package and tests using exactly these six files:
>
> 1. `inventory/__init__.py` — export the package's public types and functions.
> 2. `inventory/models.py` — define a frozen `Item` dataclass with `sku`, `name`, `quantity`, and `unit_price: Decimal`.
> 3. `inventory/store.py` — define `UnknownSKUError` and `Inventory` with replace-by-SKU `add`, quantity `adjust`, unknown-SKU error, and deterministic `items()` ordering.
> 4. `inventory/report.py` — exact Decimal `inventory_value` and inclusive, SKU-sorted `low_stock`.
> 5. `test_store.py` — insertion/replacement, adjustment, ordering, and error tests.
> 6. `test_report.py` — exact total and low-stock filtering/sorting tests.
>
> Use `unittest`. All tests must pass before you're done.
>
> Work autonomously end-to-end. If you hit a product decision only a human can make about the task itself, ask me directly.

## Excluded Attempts (Not Counted as Evidence)

| Attempt | Result | Why excluded |
|---|---|---|
| Phase 8 attempt 1 | Task/full lifecycle/validate passed | Codex read global `/Users/tinhtute/.agents/skills/{work,brainstorm,to-plan,check}/SKILL.md`; violates R9 |
| Phase 8 attempt 2 (`--ignore-user-config --ignore-rules`) | Began cold discovery | Still read a global `SKILL.md`; flags alone do not isolate the skill filesystem |
| Phase 8 attempt 3 (auth-only HOME) | Zero SKILL reads | Stopped for `brainstorm.md`'s mandatory procedural confirmation; no human coaching allowed → filed #39 |
| Phase 9 dry pass 1 | Autonomous, tests passed | Fabricated invalid RUN ULID because no exact generator existed → filed #40 |
| Phase 9 dry pass 2 | Zero SKILL reads/questions, tests passed, `valid:true` | Release-gate evidence only; task correctly selected simple mode, so it did not exercise R9's full story/trace/check chain |

The final run is the first attempt satisfying every clause simultaneously.

## Command Trail (Abridged, Order Preserved)

1. `zharness --version` → `zharness version 0.4.1`.
2. Read root `AGENTS.md`; attempted root `AUTHORITY.md`/`CONTEXT_RULES.md`; hit known broken relative links (#37).
3. Self-recovered unaided by locating `.kit/docs/{AUTHORITY,CONTEXT_RULES}.md` and the stage playbooks.
4. Read `brainstorm.md`; recognized the prompt's explicit execution intent and proceeded without a procedural question (#39 fix exercised).
5. `zharness id --json` → minted SPEC id `01KXX6ZR9S004MT3FFAF58PVS9`.
6. `zharness intake --type new-spec ... --lane tiny --json` → intake `01KXX70TYN0W21P11WCRPNNCZW`.
7. Wrote locked SPEC with `Downstream: to-plan full` (six explicit files exceed simple mode's file guard).
8. Read `to-plan.md`; wrote ROADMAP + one phase CONTEXT/PLAN.
9. `zharness story --slug inventory-package ... --json` → story `01KXX72MYAESC7HHB9ZQCYCJRF`.
10. `zharness id --json` → minted and applied the to-plan meta changeset filename; `current_phase` became `inventory-package`.
11. Read `work.md`; `zharness id --json` minted RUN `01KXX73MJHEVM7HWAS34BCSEBG`; a separate `zharness id --json` minted its changeset filename.
12. Applied the two-line run/meta changeset; RUN was durably registered as `mode: full`.
13. Implemented exactly the six requested files.
14. `python3 -m unittest -v` → 6/6 pass.
15. `zharness trace add --wave 1 ... --run-id 01KXX73MJHEVM7HWAS34BCSEBG --json` → trace `01KXX75ZT2YMN6F2GX8PBK57EK`; RUN frontmatter links it.
16. Read `check.md`; reran the full tests and audit.
17. `zharness check record --verdict APPROVE_WITH_REQUESTS --run-id 01KXX73MJHEVM7HWAS34BCSEBG ... --json` → CHECK `01KXX77N09FEXQ8C0KYKB5YAAC`.
18. `zharness id --json` minted the latest-check meta changeset filename; changeset applied successfully.
19. Process exited naturally with no question/continuation.

## Independent Verification

### Product tests

```text
test_add_inserts_and_replaces_by_sku ... ok
test_adjust_changes_only_quantity ... ok
test_adjust_unknown_sku_raises ... ok
test_items_are_sorted_by_sku ... ok
test_inventory_value_is_exact_decimal_total ... ok
test_low_stock_filters_inclusively_and_sorts_by_sku ... ok

Ran 6 tests in 0.000s
OK
```

### Validation

```json
{"valid":true,"findings":[{"link":"SPEC->PLAN","issue":"not_yet_implemented","detail":"PLAN artifacts don't carry a spec_id field yet; SPEC->PLAN cannot be cross-checked"}]}
```

`not_yet_implemented` is explicitly nonblocking by CONTRACT/implementation; `valid` is true and the process exits 0.

### Resume

```json
{"position":{"current_phase":"inventory-package","status":"planned"},"latest_run_id":"01KXX73MJHEVM7HWAS34BCSEBG","latest_check_id":"01KXX77N09FEXQ8C0KYKB5YAAC","latest_handoff_id":null,"drift":[],"readiness":"in-progress"}
```

### Audit

```json
{
  "pointer_drift": [],
  "contract_violations": [{"link":"SPEC->PLAN","issue":"not_yet_implemented",...}],
  "unlinked_proofs": [{"link":"check","issue":"unlinked_proof","detail":"... artifact_path \"test_store.py and test_report.py\" not found on disk"}],
  "entropy_score": 13
}
```

### Durable lifecycle rows

- intake: `01KXX70TYN0W21P11WCRPNNCZW`
- story/plan id: `01KXX72MYAESC7HHB9ZQCYCJRF`
- run: `01KXX73MJHEVM7HWAS34BCSEBG` (`mode: full`)
- trace: `01KXX75ZT2YMN6F2GX8PBK57EK`
- check: `01KXX77N09FEXQ8C0KYKB5YAAC` (`APPROVE_WITH_REQUESTS`)
- `query phases`, `query artifacts`, and `query check --latest` all resolve these records correctly.

### No SKILL.md exposure / no coaching

The transcript check:

```sh
grep -nE '/SKILL\.md|\.agents/skills|\.claude/skills|\.codex/skills' agent-pilot-final-transcript.log
```

returned **zero matches**. A second grep for `Should I`, `Please review/approve`, `let me know`, and confirmation requests returned no agent question; matches were only normal `codex` event markers.

## Findings

### F1 — Multi-file proof command cannot be represented cleanly by singular `artifact_path` (nonblocking)

Codex recorded one command (`python3 -m unittest -v`) that verified two files, following check.md's "one proof entry per command" wording, but stored `artifact_path: "test_store.py and test_report.py"`. Audit interprets it as one literal path and reports `unlinked_proof`.

- **Severity:** Minor for R9; proof output is real, CHECK is registered, pointers/validation pass.
- **Routing:** [GitHub #41](https://github.com/therealtinhtute/skills/issues/41).

### Known #37 — root AGENTS.md relative links

Hit again on first navigation, then self-recovered via scaffold discovery without help. Already tracked; not a new blocker.

## Closure

R9 is closed **GO**. The architecture claim is now proven by a genuinely isolated non-Claude runtime against released CLI/docs:

- written docs + CLI were sufficient,
- no Claude/Codex skill file supplied workflow mechanics,
- no human supplied workflow mechanics,
- the full durable lifecycle completed,
- the product passed its tests,
- the produced chain validated successfully.
