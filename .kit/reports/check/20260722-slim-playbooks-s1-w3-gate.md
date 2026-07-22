---
id: 01KY3PA3RM6FJQ0B97W7TYXY15
type: check
phase: slim-playbooks-S1
lane: high-risk
mode: n/a (built outside `work` — no RUN artifact registered for slim-playbooks S1; see Next Action)
run_id: none
proof_links: [{command: "go build ./...", output_ref: "inline below", artifact_path: "cli/"}, {command: "go vet ./...", output_ref: "inline below", artifact_path: "cli/"}, {command: "go test ./...", output_ref: "inline below", artifact_path: "cli/"}, {command: "gofmt -l .", output_ref: "inline below", artifact_path: "cli/"}, {command: "go test ./internal/application/... -run TestScaffold -v", output_ref: "inline below", artifact_path: "cli/internal/application/scaffold.go, cli/internal/application/scaffold_test.go"}, {command: "go test ./internal/embedded/... -run TestBuildManifest -v", output_ref: "inline below", artifact_path: "cli/internal/embedded/manifest_disk_test.go"}, {command: "go build -o /tmp/zharness-fresh ./cmd/zharness && /tmp/zharness-fresh scaffold {run,check,handoff,spec} --json", output_ref: "inline below", artifact_path: "cli/internal/interfaces/scaffold.go"}, {command: "diff templates vs pre-refactor inline playbook blocks (git show 627c0aa~1)", output_ref: "inline below", artifact_path: "cli/docs/embedded/templates/{spec,run,check,handoff}.md"}, {command: "zharness audit --json", output_ref: "inline below", artifact_path: ".kit/"}]
created: 2026-07-22
updated: 2026-07-22
---

# CHECK REPORT

Run ID: check-20260722-slim-playbooks-s1-w3-gate
Scope: full
Artifact Alignment: aligned (source) / drift (one pre-existing ad-hoc report's cross-links, see Findings)
Review Verdict: APPROVE with requests
Phase: slim-playbooks-S1 (informal — no harness story row; `resume` still shows `current_phase: write-boundary`)
Spec: .kit/planning/SPEC.md (Harness Subtraction Pass — slim-playbooks nests under it informally)
Plan: .kit/plans/2026-07-21-slim-playbooks/PLAN.md
Cook Run: none — S1-W1/W2 were built directly, not through `work` (HANDOFF.md, confirmed: no `.kit/runs/work/*slim-playbooks*` file exists)
Created At: 2026-07-22 (session date)

## Gate Evidence
- tests: `go test ./...` (cli/) → pass (all 6 packages ok, incl. 5 new scaffold unit tests + extended manifest-disk guard)
- types: `go vet ./...` (cli/) → pass
- lint: no linter configured (no `.golangci.yml`, no `staticcheck`/`golangci-lint` on PATH); `gofmt -l .` → pass (no unformatted files) used as substitute
- build: `go build ./...` (cli/) → pass

## Artifact Alignment
- status: aligned
- notes:
  - Diff (94fdf5d + 627c0aa, 19 files, +706/-358) matches the locked plan's S1 scope exactly: new `scaffold` command + 4 embedded templates (W1), then playbooks rewired to call it (W2). No surface creep.
  - Byte-for-byte diffed each extracted template (`spec.md`, `run.md`, `check.md`, `handoff.md`) against the inline blocks it replaced in the pre-refactor playbooks (`git show 627c0aa~1`) — all four match exactly; no content lost or altered in extraction.
  - `Templates` embed.FS deliberately kept separate from `FS` (not walked into `.kit/docs` manifest/init projection) — verified by `TestBuildManifest_MatchesDiskTree`, which now unions both embeds against the disk tree.
  - Proof trail present in HANDOFF.md and this diff's own commit messages (build/vet/test/smoke cited); consistent with what re-running those commands shows now.

## Findings
### Critical
- none

### Major
- Installed CLI binary at `~/.local/bin/zharness` (mtime 2026-07-20 10:38) predates commit 94fdf5d (2026-07-21 22:34) that introduced `scaffold`. Running `zharness scaffold ...` against the installed binary fails with `unknown command "scaffold"`, despite HANDOFF.md's claim that the binary was "reinstalled with scaffold." Every playbook now wired to call `zharness scaffold` (check/work/handoff/brainstorm, per 627c0aa) will fail at that step until the local binary is rebuilt. Fix: `cd cli && go build -o ~/.local/bin/zharness ./cmd/zharness` — the public `scripts/install-zharness.sh` won't help; it pulls a tagged GitHub release, which doesn't carry this command yet either.
- `zharness audit --json` reports 2 `contract_violations` against `.kit/reports/check/20260721-2029-state-readiness-audit.md` (added in this diff's scope, commit 94fdf5d): `missing_key` (`run_id: none` is not a valid ULID) and `stale_pointer` (its `id` has no matching row in the `checks` table). Root cause: that report is a deliberate ad-hoc state-audit (`mode: n/a`, not a phase gate), never meant to be registered via `check record` — but the `audit`/`validate` cross-link checker doesn't yet distinguish "intentionally unregistered ad-hoc report" from "unregistered gate check that should have one." Same shape as several pre-existing `unlinked_proofs` entries already in the audit output; not a defect in the scaffold feature itself, but it does touch a file under review, so flagged per the gate's contract-violation rule.

### Minor / Suggestions
- No RUN artifact exists for slim-playbooks S1 (built outside `work`, confirmed by HANDOFF.md and absence in `.kit/runs/work/`). This blocks `zharness check record --run-id ...` from having anything to link against — same root cause as the documented `mode: simple` skip (no row to link), applied here by extension. Recorded as a skip below, not a gate failure.

## Next Action
- Rebuild + reinstall the local `zharness` binary (`cd cli && go build -o ~/.local/bin/zharness ./cmd/zharness`) before relying on `scaffold` in any subsequent skill invocation in this environment.
- `check record` skipped (see harness_verdict below) — no RUN row exists for this phase to link against. If a human wants the S1 diff formally registered (RUN + CHECK rows), retro-author a RUN via `to-plan`/`work` first, then re-run this gate against that RUN's id.
- Ready to continue toward S2 (`zharness next`) per HANDOFF.md's stated sequencing, once the binary is rebuilt.
