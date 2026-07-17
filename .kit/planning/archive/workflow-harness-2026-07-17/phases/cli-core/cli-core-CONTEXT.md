# Context: cli-core

Phase: cli-core
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: medium
Expected Proof: unit + integration + platform (CI)

## Goal
`zharness` exists: Go module with the four upstream-mirrored layers, SQLite+changeset core (init/migrate/import/db/query), release pipeline, install script.

## Scope Boundary
### Allowed Surfaces
- `cli/**` (new Go module — go.mod, cmd/zharness, internal/{interfaces,application,domain,infrastructure}, testdata)
- `.github/workflows/` (release + CI for cli)
- `scripts/install-zharness.sh` (new)
- Root `.gitignore` (harness.db patterns documented for consumers — repo-local only if this repo runs a .kit)

### Forbidden Surfaces
- Any `skills/workflow/**` file
- Domain commands beyond the core five (Phase 4)
- `CLAUDE.md`, root README (Phase 8)

## Spec Hooks
- R2 (layer layout, `interfaces` naming), R5 (cobra + modernc, CGO=0), R7–R10 (changeset invariants, import), R21 (releases + gh install)
- Constraint: no cgo anywhere in `cli/`; private repo → install via `gh release download`

## Locked Decisions
- Module path `github.com/therealtinhtute/skills/cli`; binary `zharness`
- `modernc.org/sqlite` driver; `github.com/spf13/cobra`; ULID via `github.com/oklog/ulid/v2`
- Changeset append happens before DB apply, both wrapped so a failed apply leaves the changeset written (replay heals) — never DB-then-changeset
- `import` is one-shot idempotent: re-running on an imported project is a no-op
- goreleaser targets: darwin/linux × amd64/arm64, CGO_ENABLED=0

## Assumptions
- Contracts from Phase 2 are frozen; any change requires re-running that phase first
- `gh` auth available in CI and on user machines

## Canonical Refs
- `cli/docs/CONTRACT.md`, `cli/docs/SCHEMA.md`, `cli/docs/STATE.md`
- `~/Lab/harness-experimental/crates/harness-cli/src/infrastructure.rs` (behavior reference — port behavior, not structure)

## Rejected Options
- mattn/go-sqlite3 — cgo breaks clean cross-compilation (SPEC decision)
- Idiomatic flat internal/ layout — SPEC chose upstream-mirror for port traceability
- Homebrew/go-install channels — deferred in SPEC

## Deferred Ideas
- Windows binaries; SQLite read indexes; changeset compaction

## Escalate If
- A SCHEMA.md structure cannot be implemented append-only → to-plan phase harness-contracts
- Upstream behavior contradicts CONTRACT.md → user clarification before coding around it
