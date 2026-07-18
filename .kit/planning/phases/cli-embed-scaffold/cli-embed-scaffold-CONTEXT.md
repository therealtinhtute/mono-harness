# Context: cli-embed-scaffold

Phase: cli-embed-scaffold
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: medium
Expected Proof: unit, integration (scratch-dir init)

## Goal
`zharness` carries the doc set via `go:embed`; `init` becomes the one-command project scaffold: creates `.kit/`, writes docs + shim into `.kit/docs/`, manages `.gitignore`, stamps `docs_version` in meta, and supports `--refresh-docs`.

## Scope Boundary
### Allowed Surfaces
- `cli/internal/**` (new embed package, init command, application layer)
- `cli/cmd/zharness/**`
- `cli/docs/embedded/**` (consumed; content fixes only if a playbook bug blocks embedding)
- `cli/docs/SCHEMA.md`, `cli/docs/STATE.md` (document docs_version)
- Go test files

### Forbidden Surfaces
- `skills/workflow/**` (no skill edits yet)
- `resume` drift logic (that is cli-stale-drift)
- Release pipeline files (that is cli-release)

## Spec Hooks
- R1 (embed mechanism), R2 (init scaffolding + .gitignore), R3 partial (docs_version stamp + --refresh-docs)
- Constraint: changesets append-only; meta changes via the existing changeset mechanism; CGO disabled toolchain unchanged

## Locked Decisions
- Write-out target: `.kit/docs/` (playbooks under `.kit/docs/playbooks/`), `AGENTS.md` shim at repo root only if absent — never overwrite an existing root AGENTS.md; instead print a notice naming the shim content location
- `docs_version` = the CLI's own version string for release builds, `dev` for dev builds; stored in `meta` via changeset like other meta fields; `dev` never triggers staleness (defined in cli-stale-drift)
- `init` idempotency matrix: no `.kit/` → full scaffold; existing `.kit/` without docs → add docs only; existing docs → leave untouched unless `--refresh-docs`
- `--refresh-docs` rewrites `.kit/docs/**` + updates the meta stamp, touches nothing else (R3)
- `.gitignore` management: ensure entries for `.kit/harness.db` and `.kit/cache/` exist (append if missing, never rewrite the file wholesale)

## Assumptions
- schema_version bump only needed if meta storage requires a new column/table; if meta is key-value, no migration needed — decide in Wave 1 after reading the schema
- Embedded docs total size is trivial for the binary (< 1 MB)

## Canonical Refs
- `cli/docs/SCHEMA.md`, `cli/docs/STATE.md`
- `docs/workflow-harness/migration.md` (the `db_not_writable` footgun R2 fixes)
- Phase 1 output: `cli/docs/embedded/**`

## Rejected Options
- Fetching docs from GitHub at init time — breaks offline/private, rejected in SPEC
- Writing docs into `docs/` at repo root (upstream's layout) — collides with consumer projects' own docs/; `.kit/` is already this harness's namespace

## Deferred Ideas
- `--claude` style CLAUDE.md block install (SPEC open question, default: untouched)
- Checksums/dry-run for install script (deferred initiative)

## Escalate If
- Meta storage cannot hold docs_version without a breaking schema change → to-plan phase (migration phase needed)
- Root AGENTS.md policy (never-overwrite) proves insufficient for a real project layout → user clarification
