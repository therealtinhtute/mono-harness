# Migrating to the zharness-backed workflow chain

This doc is the pilot-migration deliverable: install → import → validate → adopt, proven end-to-end on this repo's own real history. The original pilot-evidence log was removed by commit `655c6ac` and is retrievable at `655c6ac^`; see `docs/decisions/0004-docs-directory-deletion-655c6ac.md`.

## Install

```bash
bash scripts/install-zharness.sh        # downloads the latest zharness release into ~/.local/bin
zharness --version
```

Releases are cut by pushing a `cli/vX.Y.Z` tag, but goreleaser requires its current-tag to parse as semver, so it publishes the actual GitHub Release under the bare version (e.g. `v0.1.0`), not the `cli/v...` tag used to trigger CI (`cli/.goreleaser.yaml`). `install-zharness.sh` resolves the latest release by name (`zharness ...`), not by tag prefix, to account for this. First release (`v0.1.0`) shipped from [#26](https://github.com/therealtinhtute/mono-harness/issues/26).

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

`zharness init` creates `.kit/` when needed, creates `harness.db` at the repo root, and projects the CLI's embedded doc set into `docs/` — `docs/WORKFLOW.md`, `docs/playbooks/*.md`, and an `AGENTS.md` managed block (`cli/internal/application/init.go:33`). Re-running plain `init` is idempotent and preserves locally changed managed docs by staging the conflict under `.kit/conflicts/`; use `zharness init --refresh-docs --json` when those docs should be refreshed from the embedded canonical versions, and `--force-docs` to overwrite locally changed ones. Spine skills do not self-scaffold: when `docs/` is missing, `zharness preflight <stage> --json` returns a `stop` whose recovery is `zharness init` (`cli/internal/application/preflight.go:76-83`), and the skill runs that before continuing. Playbooks are edited in `cli/docs/embedded/playbooks/` only; `docs/playbooks/` is generated output, hash-tracked in the `managed_docs` table — never hand-edit the projection. The legacy checklist below applies only when a project already has pre-harness markdown (`workflow-state.yml`, hand-written `.kit/planning/**`) to import.

## Legacy `.kit/` checklist (adopting zharness on a project with existing markdown-only artifacts)

0. **Initialize**: run `zharness init --json` from the project root. It creates `.kit/` when missing, initializes `harness.db`, and projects any missing managed docs into `docs/`. Add `--refresh-docs` only when the existing `docs/playbooks/` and `docs/WORKFLOW.md` should be refreshed from the embedded canonical copies.
1. **Import**: run `zharness import --json`. This parses `.kit/workflow-state.yml` and the planning markdown it points to, writing rows straight into `harness.db` (`cli/internal/application/import.go`) — the changeset log this step once produced was retired in `p3-retire-changesets`. It is intentionally minimal: it only creates `story` rows for phases actually referenced by `current_phase`/`entry_phase`/the latest run's phase — not a full historical backfill of every roadmap phase — and it never synthesizes a `checks` row (no check-report body parsing), so `latest_check_id` stays `null` until a real `check record` happens. It is idempotent by pre-check: a second run with unchanged legacy input writes nothing new.
2. **Verify**: `zharness query state --json` — confirm `current_phase`/`entry_phase` match your pre-import `workflow-state.yml` values.
3. **Validate**: `zharness validate --json` and `zharness audit --json`. On a project with history that predates the harness, expect real findings here — pre-harness `run`/`check`/`handoff` artifacts won't have the ULID `id`/`run_id`/`check_id` frontmatter the cross-link contract expects. This is a known, filed gap ([#25](https://github.com/therealtinhtute/mono-harness/issues/25)), not a sign the import failed — check `pointer_drift` (should be empty) separately from `contract_violations` (frontmatter completeness) to tell the two apart.
4. **`.gitignore`**: ignore `harness.db` (plus its `-wal`/`-shm` sidecars) and every per-machine `.kit/` subdirectory — `.kit/cache/`, `.kit/conflicts/`, `.kit/log/`. Nothing inside `.kit/` needs to be committed: it holds only local state and is rebuilt by `zharness init`. The durable truth is the plan markdown, which lives outside `.kit/` at `docs/plans/active/{slug}.md` (moved to `docs/plans/completed/{slug}.md` on final closure) and must be tracked. `docs/playbooks/` and `docs/WORKFLOW.md` are the projected managed docs; commit them so a fresh clone can read the lifecycle without the binary. See this repo's own `.gitignore` for the pattern.
5. **`trash` the legacy template, not the state file**: once the harness is live, `workflow-state.yml` on a given project becomes redundant with `zharness query state`, but this repo's own copy is left in place as a historical artifact (do not delete another project's live state file without checking whether anything outside the harness still reads it).

## Rollback

Plain and blunt:
- The plan markdown under `docs/plans/` stays human-readable with no CLI at all — nothing is lost by not running `zharness`.
- `harness.db` is fully rebuildable from the committed plan markdown alone (`zharness db rebuild --yes`) — losing the DB file loses nothing as long as the plans are committed. The DB is a derived index, never the source of truth.
- Abandoning the CLI entirely means losing: machine-queryable state (`query state`, `resume`), deterministic gate verdicts (`check record`), and cross-link validation (`validate`/`audit`). The markdown trail alone becomes the only source of truth again, same as before this initiative — no worse off, just back to the original manual-discipline model.

## Contributor playbook: extending zharness without breaking the rebuild

- Markdown is the source of truth and `harness.db` is derived from it. The gate on any new mutating command is that `zharness db rebuild --yes` still reconstructs the same state from committed plan markdown alone (`cli/docs/SCHEMA.md`). The append-only changeset log this rule was originally phrased against was retired in `p3-retire-changesets`; lifecycle commands now write direct SQL inside an explicit transaction.
- Adding a new mutating command: write the markdown section and the DB row in the same transaction, following any existing `cli/internal/application/*.go` lifecycle command. A field that lands only in the DB and never in markdown will silently vanish on the next rebuild.
- Skills never hand-author rows. If a skill needs an entity the CLI has no command for, add the command — the CLI owns the pen, and the playbooks depend on it appending its own `## Progress` and `## Decisions` lines.
- If a skill's own reference doc quotes CLI-emitted text verbatim (e.g. `resume`'s recovery strings), keep it in sync with the actual Go source, not just the design doc — this pilot found a live mismatch between `resume.go` and `cli/docs/STATE.md` ([#24](https://github.com/therealtinhtute/mono-harness/issues/24)) precisely because the doc and the code drifted independently.
