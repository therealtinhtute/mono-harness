# Context: cli-release

Phase: cli-release
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: low (repo) / medium (external consumers of releases)
Expected Proof: platform (release pipeline), integration (install-path smoke)

## Goal
Ship `cli/v0.2.0` — the first release with embedded docs — through the existing tag → goreleaser pipeline, and prove `scripts/install-zharness.sh` resolves and installs it.

## Scope Boundary
### Allowed Surfaces
- Git tags (`cli/v0.2.0` trigger + bare `v0.2.0` for goreleaser)
- `.github/workflows/**` only if the pipeline breaks (fix-forward, minimal)
- `cli/` version metadata (whatever the build stamps versions from)
- `scripts/install-zharness.sh` only if resolution fails

### Forbidden Surfaces
- `skills/workflow/**` (MIN_ZHARNESS_VERSION bump belongs to thin-triggers)
- New CLI features (feature freeze — this phase releases what exists)

## Spec Hooks
- R11 prerequisite (the version thin-triggers will gate on)
- Constraint: "skill rewrites land only after the CLI that supports them is released" — this phase is the gate between CLI work and skill work

## Locked Decisions
- Version: `0.2.0` (minor bump — additive commands/flags + scaffolding behavior, no breaking command changes)
- Reuse the proven release flow documented in `docs/workflow-harness/migration.md`: `cli/vX.Y.Z` tag triggers CI, local bare-semver tag created for goreleaser, release published under the bare version, installer resolves by release name
- Release smoke = the Phase 3 scratch-dir lifecycle suite run against the *installed* binary, not a rebuilt one

## Assumptions
- The goreleaser config from v0.1.0 works unchanged for v0.2.0 (embed adds no CGO or asset complexity)
- GitHub release permissions/secrets unchanged since #26

## Canonical Refs
- `docs/workflow-harness/migration.md` (release quirks section)
- Issue #26 (v0.1.0 release evidence), commits `6ae91a0`, `15b8acc` (pipeline fixes already landed)

## Rejected Options
- Waiting to release until after thin-triggers (batch skills + CLI) — violates the SPEC constraint that the chain stays operational; skills must gate on an installable version
- v1.0.0 — the agent-agnostic surface is new and unproven until the pilot passes; don't signal API stability yet

## Deferred Ideas
- sha256 checksum verification in install-zharness.sh (deferred initiative, noted in SPEC)

## Escalate If
- Pipeline needs structural rework (not a config fix) → to-plan phase
- Release publishes but installer cannot resolve it after a filter fix attempt → user clarification (naming scheme decision)
