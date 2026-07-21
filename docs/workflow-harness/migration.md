# Migrating to the zharness-backed workflow chain

This doc is the pilot-migration deliverable: install → import → validate → adopt, proven end-to-end on this repo's own real history (see `docs/workflow-harness/pilot-evidence/2026-07-17-lab-skills-import.md`).

## Install

```bash
bash scripts/install-zharness.sh        # downloads the latest zharness release into ~/.local/bin
zharness --version
```

Releases are cut by pushing a `cli/vX.Y.Z` tag, but goreleaser requires its current-tag to parse as semver, so it publishes the actual GitHub Release under the bare version (e.g. `v0.1.0`), not the `cli/v...` tag used to trigger CI. `install-zharness.sh` resolves the latest release by name (`zharness ...`), not by tag prefix, to account for this. First release (`v0.1.0`) shipped from [#26](https://github.com/therealtinhtute/skills/issues/26).

Building from source instead of installing a release:
```bash
cd cli && go build -o /tmp/zharness ./cmd/zharness
/tmp/zharness --version   # prints a dev-build version; skills' version gate accepts this
```

## New-adopter path (no existing `.kit/`)

Nothing to migrate — just install and initialize:

```bash
bash scripts/install-zharness.sh
zharness init --json
```

`zharness init` creates `.kit/` when needed, creates `harness.db`, and scaffolds missing `.kit/docs/playbooks/` from the CLI's embedded docs. Re-running plain `init` is idempotent and preserves existing generated docs; use `zharness init --refresh-docs --json` when those docs should be replaced with the embedded canonical versions. Every spine skill runs `init` itself when `.kit/docs/` is missing, so a first-time `/brainstorm` or `/work` invocation can self-scaffold. The legacy checklist below applies only when a project already has pre-harness markdown (`workflow-state.yml`, hand-written `.kit/planning/**`) to import.

## Legacy `.kit/` checklist (adopting zharness on a project with existing markdown-only artifacts)

0. **Initialize**: run `zharness init --json` from the project root. It creates `.kit/` when missing, initializes the local database, and scaffolds any missing generated docs. Add `--refresh-docs` only when existing `.kit/docs/` should be refreshed from the embedded canonical copies.
1. **Import**: run `zharness import --json`. This parses `.kit/workflow-state.yml` and `.kit/planning/**` into changesets under `.kit/changesets/` and materializes `harness.db`. It is intentionally minimal: it only creates `story` rows for phases actually referenced by `current_phase`/`entry_phase`/the latest run's phase — not a full historical backfill of every roadmap phase — and it never synthesizes a `checks` row (no check-report body parsing), so `latest_check_id` stays `null` until a real `check record` happens.
2. **Verify**: `zharness query state --json` — confirm `current_phase`/`entry_phase` match your pre-import `workflow-state.yml` values.
3. **Validate**: `zharness validate --json` and `zharness audit --json`. On a project with history that predates the harness, expect real findings here — pre-harness `run`/`check`/`handoff` artifacts won't have the ULID `id`/`run_id`/`check_id` frontmatter the cross-link contract expects. This is a known, filed gap ([#25](https://github.com/therealtinhtute/skills/issues/25)), not a sign the import failed — check `pointer_drift` (should be empty) separately from `contract_violations` (frontmatter completeness) to tell the two apart.
4. **`.gitignore`**: make sure only `harness.db`, any local cache dir, and `.kit/docs/` are ignored — `.kit/docs/` is a generated scaffold (`zharness init` creates missing docs; `zharness init --refresh-docs` rewrites them from the embedded playbooks), so committing it just invites drift. `.kit/changesets/**`, `.kit/planning/**`, `.kit/runs/**`, `.kit/reports/**`, and `.kit/HANDOFF.md` must be git-tracked, or cross-machine resume silently breaks. This repo's own `.gitignore` previously excluded all of `.kit/` and had to be narrowed as part of this pilot.
5. **`trash` the legacy template, not the state file**: once the harness is live, `workflow-state.yml` on a given project becomes redundant with `zharness query state`, but this repo's own copy is left in place as a historical artifact (do not delete another project's live state file without checking whether anything outside the harness still reads it).

## Rollback

Plain and blunt:
- All planning/execution markdown (`SPEC.md`, `ROADMAP.md`, phase `-CONTEXT.md`/`-PLAN.md`, run logs, check reports, `HANDOFF.md`) stays human-readable with no CLI at all — nothing is lost by not running `zharness`.
- `harness.db` is fully rebuildable from `.kit/changesets/**` alone (`zharness init` + `zharness db changeset apply` per file, in ULID filename order) — losing the DB file loses nothing as long as changesets are committed.
- Abandoning the CLI entirely means losing: machine-queryable state (`query state`, `resume`), deterministic gate verdicts (`check record`), and cross-link validation (`validate`/`audit`). The markdown trail alone becomes the only source of truth again, same as before this initiative — no worse off, just back to the original manual-discipline model.

## Contributor playbook: extending zharness without breaking changesets

- Changesets are **append-only, ULID-ordered, and replayed from empty** on every rebuild (`cli/docs/SCHEMA.md`). Never edit or delete a committed changeset file — that breaks every future replay.
- Adding a new mutating command: append a changeset before touching `harness.db`, exactly like every existing command. See any `cli/internal/application/*.go` command for the pattern (build changeset struct → write JSONL → apply to DB in the same transaction).
- Adding a new entity or field that a *skill* needs but the CLI has no dedicated command for yet (e.g. this pilot's own `run` row creation): hand-author a changeset JSON line following the existing op/entity/fields/at shape (see any file under `.kit/changesets/` for a concrete example), then apply it with `zharness db changeset apply <path> --json`. This is the established convention from Phase 5 (skill-adapters) and used again in this pilot.
- If a skill's own reference doc quotes CLI-emitted text verbatim (e.g. `resume`'s recovery strings), keep it in sync with the actual Go source, not just the design doc — this pilot found a live mismatch between `resume.go` and `cli/docs/STATE.md` ([#24](https://github.com/therealtinhtute/skills/issues/24)) precisely because the doc and the code drifted independently.
