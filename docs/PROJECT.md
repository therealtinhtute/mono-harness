# PROJECT — identity (answer inline; this is the single forced write step at
# brainstorm lock; keep the whole file at or under 50 lines)

## What is this project?
- zharness: an install/update/uninstall binary that scaffolds and maintains a
  markdown-first workflow harness (playbooks, identity docs, guards) in a
  consumer git repository. v0.15 deleted SQLite and every lifecycle command.
  v0.16 is the protocol on that surface: absorb at close, at most one active
  plan, independent judge for `full` checks. The binary stays off the PATH.

## Who is it for?
- The owner and contributor agents working in this repository, plus consumers
  who `zharness install` the managed doc set into their own repos.

## Non-goals
- No hidden or deprecated lifecycle commands; no parallel control-plane state
  (no task database, no derived index).
- No edits to consumer repositories beyond the managed set; no fabricated
  backfill of consumer history.
- No application runtime, credentials, schema validation, or product policy.
- No scanning or merging of `~/.claude`, `~/.codex`, `~/.agents`,
  `~/.config/opencode` (except the single codex config line in R7).

## How do we run the tests?
- `bash scripts/verify-doc-links.sh`
- `cd cli && go build ./... && go vet ./... && go test ./...`
- Phase gates per plan: doc links, go tests, S4 `rg -i "sqlite|harness\.db" cli/`
  = 0, kill-list bounded scan = 0 actionable, kill-switch smoke.

## Architecture in one breath
- runtime shape: one Go binary (`cli/cmd/zharness`) exposing exactly install /
  update / uninstall; everything else is git-committed markdown under `docs/`
  plus fail-closed pre-commit guards (proof re-execution, high-risk and full
  independent-judge, at most one active plan).
- where state lives: `docs/plans/active/*.md` (append-only Progress /
  Decisions / Validation) and `.zharness/base/` (manifest + content-addressed
  upstream blobs) — no SQLite anywhere.
- entrypoints: `cli/internal/interfaces/root.go`; embedded doc set under
  `cli/docs/embedded/` projected to `docs/`; hooks via
  `scripts/install-git-hooks.sh`.

## What are we working on right now?
- none (absorb-encode-protocol closed)
