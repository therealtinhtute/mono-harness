# Pilot Evidence — dogfooding `Lab/skills` itself

Date: 2026-07-17
Pilot target: this repo (`Lab/skills`), confirmed by user 2026-07-17 (not a synthetic target)
Phase: pilot-migration, Wave 1 T1

## Changeset log (ULID order, `.kit/changesets/`)

| File | Op | Entity | Summary |
|---|---|---|---|
| `01KXR8RR9GAT35QFS05NCRJN8F` | create | story | `continuity` (status: in-progress) |
| `01KXR8RR9HFV81RGAQ42V898M3` | create | story | `harness-concept` (status: planned) |
| `01KXR8RR9HFV81RGAQ48PD21BE` | create | story | `pilot-migration` (status: planned) |
| `01KXR8RR9J1P5SP9KJD1R57MVZ` | create | run | continuity run, artifact `.kit/runs/work/20260717-2140-continuity.md` |
| `01KXR8RR9J1P5SP9KJD32K4APV` | update | meta | `current_phase=pilot-migration`, `entry_phase=harness-concept`, `latest_run_id` |
| `01KXR94MPXB2JF7Y9MH6ACBEEB` | create | run | pilot-migration run, artifact `.kit/runs/work/20260717-2200-pilot-migration.md` |
| `01KXR95D04X6ZE6WSN0YSDHF2V` | update | meta | `latest_run_id` → pilot-migration run |

First 5 written by `zharness import` against this repo's pre-harness `.kit/workflow-state.yml`; last 2 hand-authored (no `run create` CLI command exists — established convention from Phase 5/skill-adapters).

## `init` + `import` (SPEC acceptance criterion: "legacy project: init && import && query state --json returns correct state derived from old workflow-state.yml")

```
$ zharness init --json
{"db_path":".kit/harness.db","schema_version":1,"status":"created"}

$ zharness import --json
{"imported":5,"skipped":0,"changesets_written":[<5 ULID paths above>]}

$ zharness query state --json
{"current_phase":"pilot-migration","entry_phase":"harness-concept","schema_version":1,"latest_run_id":"01KXR8RR9J1P5SP9KJCY27BHMK","latest_check_id":null}
```
Matches pre-import `workflow-state.yml` (`current_phase: pilot-migration`, `entry_phase: harness-concept`) exactly. `latest_check_id: null` is correct-by-design: `import` never synthesizes a `checks` row (verdict is NOT NULL, yml has no check-report body to map) — confirmed by reading `cli/internal/application/import.go`'s doc comment, not assumed.

## Rebuild-from-changesets (cross-machine resume proof, real history)

Copied `.kit/` (excluding `harness.db`, `cache/`) to a scratch directory simulating a second machine, ran `zharness init` + `zharness db changeset apply` on all 7 changesets in ULID order:

```
$ zharness resume --json   # scratch (rebuilt)
{"position":{"current_phase":"pilot-migration","status":"planned"},"latest_run_id":"01KXR93JGZTBHA5BVYPNAJ0N0X","latest_check_id":null,"latest_handoff_id":null,"drift":[],"readiness":"in-progress"}

$ zharness resume --json   # original
{"position":{"current_phase":"pilot-migration","status":"planned"},"latest_run_id":"01KXR93JGZTBHA5BVYPNAJ0N0X","latest_check_id":null,"latest_handoff_id":null,"drift":[],"readiness":"in-progress"}
```
Byte-identical. Zero divergence — rebuild-from-changesets stop condition did not trigger.

Note: this used a directory copy, not `git clone`, because this phase's own commit (which first brings `.kit/` under git tracking — see the `.gitignore` fix below) had not yet landed at evidence-capture time. The changeset-rebuild mechanism itself does not depend on git; a `git clone`-based repeat is trivial to re-run after `/git cm` lands, same procedure as continuity's T4.

## `validate` — real gap found (not hypothetical)

```
$ zharness validate --json
{"valid": false, "findings": [39 entries — see below]}
```

Every `.kit/runs/work/*.md` and `.kit/reports/check/*.md` file from phases 1–6 (harness-concept through validation-gate, all already gated APPROVE and committed) fails cross-link validation: `RUN` files are missing `id`/`phase`/`plan_id` (or hold non-ULID placeholder values like `none`, `continuity-PLAN.md`, `01J0000000000000000HARNCT`), `CHECK` reports are missing `id`/`run_id`, and `.kit/HANDOFF.md` is missing `id`/`run_id`/`check_id`. `zharness audit --json` reports `entropy_score: 100` (max) and zero `pointer_drift` (the live DB pointers are internally consistent — the gap is entirely in the markdown artifacts' own frontmatter, not the DB).

**Root cause**: phases 1–6 were executed and gated *before* the harness existed to backfill these fields — the run/check artifact templates only started requiring real ULID `id`/`run_id`/`check_id` values once `continuity` (phase 7) rewired the skills to be CLI-first. Phases 1–6's markdown artifacts were never retrofitted.

**This is the real "gap" pilot-migration's own Wave 2 (T2) is designed to surface** — see `skills/workflow/README.md`'s go/no-go section and the filed GitHub issue for the backfill/retrofit decision. Not hotfixed here per this phase's own `avoid` rule ("hotfixing cli/ or skills mid-run — record gaps instead").

## Verdict inputs
- `init && import && query state --json` — **pass**, satisfies SPEC's literal acceptance wording on real (not synthetic) history
- rebuild-from-changesets — **pass**, byte-identical resume output
- `validate`/`audit` — **fail on historical artifact completeness**, entropy 100, zero DB-level pointer drift; scoped as a filed gap, not a phase-blocking defect (T1's own stop condition — "`import` cannot correctly derive current state" — did not trigger; state derivation is correct, only pre-harness artifact backfill is missing)
