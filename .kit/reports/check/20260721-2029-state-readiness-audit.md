---
id: 01KY2DVZDEFCNHCFEQSGBYPRCB
type: check
phase: write-boundary
lane: high-risk
mode: n/a (state-audit, not a phase gate — no new RUN in scope)
run_id: none
proof_links:
  - {command: "go build ./... && go vet ./... && go test ./...", output_ref: "inline below", artifact_path: "cli/"}
  - {command: "go build -o /tmp/zharness-fresh ./cmd/zharness && /tmp/zharness-fresh run create --help", output_ref: "inline below", artifact_path: "cli/cmd/zharness"}
  - {command: "zharness --version", output_ref: "inline below", artifact_path: "~/.local/bin/zharness"}
  - {command: "zharness audit --json", output_ref: "inline below", artifact_path: ".kit/"}
  - {command: "zharness db changeset status --json", output_ref: "inline below", artifact_path: ".kit/changesets/"}
  - {command: "diff cli/docs/embedded/playbooks/{check,work}.md .kit/docs/playbooks/{check,work}.md", output_ref: "inline below", artifact_path: ".kit/docs/playbooks/"}
  - {command: "gh release list", output_ref: "inline below", artifact_path: "n/a"}
created: 2026-07-21
updated: 2026-07-21
---

# CHECK REPORT — state-readiness-audit

Run ID: check-20260721-2029-state-readiness-audit
Scope: full (audit-oriented — argument "current code with goal, audit --> update state plan audit")
Artifact Alignment: drift (operational state) / aligned (source)
Review Verdict: APPROVE with requests
Phase: write-boundary (last completed phase; this audits its post-merge readiness, not a new diff)
Spec: .kit/planning/SPEC.md
Plan: .kit/planning/phases/write-boundary/write-boundary-PLAN.md
Cook Run: .kit/runs/work/20260721-1027-write-boundary.md
Created At: 2026-07-21 20:29

## Scope note

Working tree is clean (`git diff --stat HEAD` empty) — there is no fresh code diff to gate. Per the argument's intent ("current code with goal, audit"), this check instead audits whether the repo's **claimed state** (HANDOFF.md / ROADMAP.md: "write-boundary DONE, Phase 1/4") matches **provable reality** right now, before Phase 2 (dead-surface-removal) starts. It does.

## Gate Evidence (source level — re-run fresh, not cited from the prior report)

- tests: `go test ./...` (in `cli/`) → **pass** — all 6 packages ok
- types/vet: `go vet ./...` → **pass**
- build: `go build ./...` → **pass**
- `run create` proof: `go build -o /tmp/zharness-fresh ./cmd/zharness && /tmp/zharness-fresh run create --help` → **pass**, command exists, flags match `CONTRACT.md`

Source-level: write-boundary Phase 1 is real, complete, and still green today. This is not a re-litigation of the prior APPROVED verdict — it's independent re-verification.

## Findings

### Critical
- none

### Major
1. **Local `harness.db` has not replayed the 23 changesets this session's `git pull` brought in.** `zharness db changeset status --json` → `"applied_count":0, "pending":[23 files]`, last applied `01KXX6NBMRTC06H8EVPNBFH6J2`. Direct consequence: `zharness audit --json` reports `entropy_score:100` with `stale_pointer` findings naming **exactly** the write-boundary artifacts under review — `RUN .../20260721-1027-write-boundary.md id ... has no matching row in the runs table` and `CHECK .../20260721-1044-write-boundary.md id ... has no matching row in the checks table`. Per this playbook's own Step 4 rule, any non-empty `pointer_drift`/`contract_violations` touching artifacts under review is a finding rated 🟠 Major minimum — this qualifies. Not a harness bug: it's the expected "derived cache lags committed log" gap the changeset design anticipates; it just hasn't been resolved on this machine yet.
   - **Fix**: `zharness db changeset apply <path>` for each of the 23 pending files in ULID order (list is in the command output above), or wait — no bulk command exists today (see Suggestion below).
2. **The installed `zharness` binary predates write-boundary — Phase 1 is source-complete, not deploy-complete.** `~/.local/bin/zharness` is v0.4.1, file-dated 2026-07-19 19:42 (`gh release list` confirms v0.4.1 was cut 2026-07-19T12:41Z). write-boundary merged 2026-07-21 11:03 (`git log 32cb60c`) — **after** the last release. `zharness run --help` on the installed binary → `unknown command "run"`. No release has been cut since. Practical effect: if `work` or `check` run right now using the installed binary, they will read the stale scaffolded playbooks (finding 3) and fall straight back into hand-authoring changeset JSONL — the exact anti-pattern write-boundary exists to close. **Phase 2 must not start against this binary.**
3. **`.kit/docs/playbooks/{check,work}.md` are drifted from the canonical embed** (confirmed by direct diff — check.md: 2 hunks, work.md: 17 diff lines). Root cause is #2, not a regression of the prior session's "byte-verified" re-scaffold claim: `.kit/docs/` is gitignored (`.gitignore:4`), so it never traveled with `git pull`, and it reflects whatever the *installed* (stale) binary last scaffolded via `zharness init`. `zharness init` alone won't fix this while the binary itself lacks the feature — the binary must be rebuilt/reinstalled first, then `zharness init` re-run.

### Minor / Suggestions
- 💡 No CLI command bulk-replays pending changesets (`db changeset apply` is one file at a time). The pilot evidence names cross-machine replay as a core guarantee; today recovering from a 23-changeset pull lag is 23 manual commands. Worth a small `zharness db changeset apply-all` (or `sync`) addition — low effort, directly serves the "6-months-later DX" goal the architecture audit scored weakest (65/100).
- 💡 The version gate (`zharness --version` / `MIN_ZHARNESS_VERSION`) cannot catch this class of drift because write-boundary shipped without a version bump — a same-numbered binary can silently lack a merged feature. Not blocking now, but worth a note for release discipline: bump the CLI version on any user-visible command addition, even mid-initiative.
- Pre-existing/unrelated drift already present before write-boundary (not this audit's concern, matches prior check report's own carve-out): `autonomous-entry-parity` missing run artifact, `readme-workflow-refresh` broken PLAN link, `SPEC->PLAN not_yet_implemented` (documented known gap, CONTRACT.md:144).

## Pattern-Fix Completeness

The same "installed-binary-lags-source" shape will recur for Phases 2-4 (dead-surface-removal, scoring-removal, single-source-playbooks) unless resolved once now — each phase's `work` invocation reads whatever binary is on PATH. Fixing the sync now (not per-phase) closes it for the rest of the initiative.

## Hard Stops
- [x] No unverified claims — every command above ran this session; output summarized/cited from real terminal output.
- [x] No forbidden-surface edits — this check made zero file edits (read-only throughout, per contract).
- [ ] Missing proof trail — n/a, no new RUN in scope for this audit pass.

## Verdict: APPROVE WITH REQUESTS

Write-boundary's *code* is solid and re-verified green. The *state* (ROADMAP/HANDOFF marking Phase 1 fully closed) overstates operational readiness: local DB is 23 changesets behind, the installed CLI lacks `run create`, and the scaffolded playbooks are stale as a direct result. None of this is a code defect — it is entirely fixable by 3 mechanical sync actions, no design or scope questions involved.

## Next Action (proposed — human or `work` applies; this check does not self-fix)

1. Replay the 23 pending changesets into local `harness.db`: `zharness db changeset apply <path>` per file in `zharness db changeset status --json`'s `pending` list, in the listed (ULID/chronological) order.
2. Rebuild + reinstall `zharness` from current `cli/` source (or cut+install a new release tag once ready) so `~/.local/bin/zharness` includes `run create` / the write-boundary `check record` behavior.
3. Re-run `zharness init` to regenerate `.kit/docs/playbooks/{check,work}.md` from the now-current embed; re-diff against `cli/docs/embedded/playbooks/` to confirm zero drift.
4. Re-run `zharness audit --json` — expect `pointer_drift`/`contract_violations` on the write-boundary artifacts to clear; `entropy_score` should drop back toward baseline (pre-existing unrelated findings will remain, that's expected).
5. Only then start Phase 2 (`dead-surface-removal`) via `to-plan`/`work` — starting it against the current stale binary would silently reintroduce the exact hand-authored-changeset flow Phase 1 just closed.
6. Update `HANDOFF.md` / this initiative's tracking to distinguish "merged + gated" from "deployed + verified live" for Phase 1 — the current wording ("write-boundary Phase 1 of 4 implemented and check-gated APPROVED") reads as fully closed; it should note the deploy step as outstanding until step 2 above runs.
