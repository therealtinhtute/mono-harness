# Plan: cli-release

Phase: cli-release
Status: ready
Wave Count: 2
Execution Owner: work
Updated At: 2026-07-18

## Goal
Tag, build, publish `cli/v0.2.0` with embedded docs; prove the install path end-to-end.

## Inputs
- Green `cli/` tree from cli-stale-drift (all tests passing on master)
- `docs/workflow-harness/migration.md` release-flow notes

## Wave 1
### T1 — Pre-release verification
- type: test
- inputs: master at the release commit
- touches: nothing (read-only gate)
- avoid: sneaking in feature changes
- steps:
  1. `cd cli && go build ./... && go test ./...` clean at the tagged commit
  2. Confirm the binary self-reports the docs version correctly for a release build (goreleaser ldflags path)
  3. Confirm CHANGELOG-worthy summary of v0.1.0 → v0.2.0 changes for the release notes
- expected outputs: go/no-go record in the run artifact
- verification: command outputs captured; version string check
- stop if: any test red
- escalate to: check

### T2 — Tag and publish
- type: implementation
- inputs: T1 go
- touches: git tags, GitHub release
- avoid: workflow edits unless CI fails
- steps:
  1. Push `cli/v0.2.0` trigger tag per the documented flow (CI creates the bare-semver tag for goreleaser)
  2. Watch the release workflow; on failure, fix-forward minimally and re-tag per the pipeline's documented recovery
  3. Verify release assets exist for the full platform matrix
- expected outputs: published GitHub release `v0.2.0` (zharness name-resolved)
- verification: `gh release view` shows assets; download one asset and run `--version`
- stop if: pipeline needs structural rework
- escalate to: to-plan phase

## Wave 2
### T3 — Install-path smoke
- type: test
- inputs: published release
- touches: nothing in-repo (scratch env)
- avoid: local dev binary contamination (clear PATH shadowing)
- steps:
  1. Fresh scratch: `bash scripts/install-zharness.sh` → `zharness --version` reports 0.2.0
  2. Run the scratch-dir lifecycle suite (Phase 3 T4) against the installed binary: init scaffold → lifecycle → resume/validate/audit clean
  3. Verify stale-docs story end-to-end for real: scaffold with 0.1.0-written docs absent stamp → confirm no false drift; then simulate an old stamp → drift fires → `init --refresh-docs` clears
- expected outputs: install + smoke evidence in the run artifact
- verification: captured command transcripts
- stop if: installer resolves the wrong release after one filter fix attempt
- escalate to: user clarification

## Risks / Watch-fors
- goreleaser current-tag semver quirk — follow migration.md exactly; do not improvise tag names
- Failed tags are never reused (upstream lesson worth honoring): a broken `cli/v0.2.0` becomes `cli/v0.2.1`, not a force-moved tag
