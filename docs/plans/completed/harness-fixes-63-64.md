---
id: 01M0YJCZE48CQXMH2RDTKXJPXX
type: plan
intake_id: 01M0YJHR93EBGG0M8TE5E66M3C
lane: normal
status: completed
created: 2026-08-26
updated: 2026-08-26
---

# Plan: Harness fixes — final-phase full check recording (#63) and installer draft race (#64)

## Outcome
- result: `zharness check record --mode full` can record the final-phase complete review on a `checked` story with the mode persisted on the check row (fixing GitHub #63), and `scripts/install-zharness.sh` refuses loudly instead of silently installing the previous release while a newer draft is publishing (fixing GitHub #64).
- success_signals:
  - Regression test reproduces #63 on unmodified code (`story_not_checkable`), then passes after the fix: a `full` check records on a `checked` story and the story stays `checked`; a `gate` check on a `checked` story still exits 1 with `story_not_checkable`.
  - `query checks` output exposes each check's mode; a recorded `full` check round-trips through `db rebuild --yes` from markdown alone.
  - Live proof: on a real `checked` story, `zharness check record --mode full ...` succeeds and returns a check ID whose row/query output says `mode: full`.
  - Install-script proof: with a simulated draft-newer-than-published release list, the script exits non-zero printing a "release still publishing" message and installs nothing; with no drafts it installs the newest published release unchanged; an explicit tag argument bypasses resolution entirely.

## Authority and Requirements
- authority:
  - GitHub issue #63 (therealtinhtute/mono-harness): `check record` rejects the final-phase full review that `docs/playbooks/handoff.md` step 6 requires.
  - GitHub issue #64: `scripts/install-zharness.sh` races goreleaser's draft window and installs the previous version.
  - `docs/playbooks/handoff.md` step 6 — final-phase closure requires a clean `check full` verdict, verifiable from durable state rather than agent memory.
  - `docs/playbooks/check.md` — modes `gate|full|review|bounded`; only gate/full record checks.
  - `cli/internal/application/check_record.go:59-61` — current guard accepting only `in-progress`.
  - Markdown-as-source-of-truth invariant (ADR 0001): every persisted check field must round-trip through `db rebuild` from plan markdown alone.
- requirements:
  - R1 [accepted]: `zharness check record` accepts a new `--mode {gate|full}` flag (default `gate`) stored on every new check row | source: issue #63 ("persist the mode on the check row")
  - R2 [accepted]: recording a check with `mode=full` on a story whose status is `checked` succeeds and leaves the story `checked`; any other mode on a `checked` story is still refused with exit code 1 and error code `story_not_checkable`; all existing refusals on non-`checked` statuses and the `run_not_latest` guard are unchanged | source: issue #63 suggested fix
  - R3 [accepted]: the `## Validation` entry written by `check record` includes the mode, and `db rebuild` reconstructs the mode column from that markdown alone; schema migration adds the column and bumps the schema version exactly once | source: ADR 0001 round-trip invariant + issue #63
  - R4 [accepted]: `query checks` (and the check output block) reports the recorded mode per check; `docs/playbooks/handoff.md` step 6 is rewritten to verify the final-phase clean `full` verdict from the DB instead of the "does not yet persist" caveat | source: issue #63 related note + handoff.md step 6
  - R5 [accepted]: when the newest zharness release on `$REPO` is a draft newer than the newest published release, `scripts/install-zharness.sh` exits non-zero with an explicit "release still publishing — retry or pass an explicit tag" message before downloading anything; with no such draft its published-release resolution behaves exactly as today; an explicit tag argument skips resolution entirely | source: issue #64 suggested fix

## Non-goals
- NG1: no change to `work`'s in-session gate flow, verdict semantics, judge rules, or proof-link verification.
- NG2: no backfill or re-recording of historical checks — existing rows keep an empty/unknown mode and are never treated as `full`.
- NG3: no wait/poll/retry loop inside `install-zharness.sh` — fail-loud only.
- NG4: no changes to `review`/bounded check modes (they remain zero-write and never call `check record`).

## Approach and Risks
- approach: Two independent phases, #63 first (lifecycle-guard change carries more risk than the shell-script fix). Follow the repo's established rhythm from memory-lifecycle: failing regressions first, then implement, then live proof + full gate per phase. `--mode` defaults to `gate` so every existing invocation behaves identically; the guard widens only along the `mode=full` axis. Markdown-write-before-DB ordering in RecordCheck is preserved. The final-phase closure of this very plan will dogfood the new path (recording a `full` check on a `checked` story).
- constraints:
  - zero new external module dependencies (CLI stays stdlib + oklog/ulid)
  - installer changes stay POSIX-ish bash driven only by `gh release list --json` fields it already uses
- rejected_alternatives:
  - Guard-only widening without persisting mode — rejected by issue #63 itself: handoff step 6 must verify the final-phase `full` verdict from durable state, which requires the persisted mode.
  - Installer waits/polls for the draft window to clear — rejected: unbounded hang risk inside an install script; issue #64 asks for fail-loud.
  - Two separate initiatives — rejected: harness enforces exactly one active plan, so sequential plans would double ceremony for two small fixes.
- risks:
  - Verdict semantics on an already-checked story are ambiguous for `REQUEST_CHANGES`. Decision: a non-clean verdict reopens the story to `in-progress` in both DB and plan phase status, mirroring existing verdict semantics; covered by an explicit regression test.
  - Rebuild compatibility with historical Validation entries that carry no mode — mitigated by an optional regex capture group; absent mode round-trips as empty string and is never treated as `full`; migration column defaults to `''`.
  - handoff.md step 6 rewrite could drift from check.md's Output Format wording — mitigated by reviewing both playbooks together in the doc task.
  - Stub-gh installer test couples to `gh release list` JSON shape — mitigated by pinning only the fields the script consumes (`tagName`, `name`, `isDraft`, `createdAt`).
- recovery: each wave names its stop condition; a wave whose failing-first proof cannot be produced without production changes stops work and escalates to `brainstorm refine`; guard-change fallout in the existing scratch-dir chain test follows the decisions-recorded-deviation rule rather than silent widening.

## Phases and Verification
<!-- Phase and task definitions are immutable after to-plan. Do not add task status fields. Append-only Progress is the sole task execution-status source. Only each phase lifecycle status changes to mirror DB transitions: to-plan=planned; work after run create=in-progress; clean durable check=checked; closing handoff=done. Each planned phase records phase_slug, story_id, status, goal, depends_on, waves, tasks, and checks. -->
- planning_status: planned
- phases:
  - phase_slug: check-mode-persistence
    story_id: 01M0YJVRPKCXJ83TFJ7S0M36Q7
    status: done
    goal: Persist the producing mode on every recorded check and allow a `mode=full` check on an already-`checked` story (GitHub #63)
    depends_on: null
    requirements: [R1, R2, R3, R4]
    allowed_surfaces: [cli/internal/application/check_record.go, cli/internal/application/rebuild.go, cli/internal/application/lifecycle_guard_test.go, cli/internal/application/rebuild_test.go, cli/internal/interfaces/check.go, cli/internal/interfaces/query.go, cli/internal/interfaces/query_checks_test.go, cli/internal/domain/, cli/internal/infrastructure/migrations.go, cli/internal/infrastructure/migrations_test.go, docs/playbooks/handoff.md]
    avoided_surfaces: [embedded playbook sources other than handoff.md step 6, scripts/install-zharness.sh owned by installer-draft-race-guard, scaffold bodies in init.go]
    waves:
      - wave: 1
        tasks:
          - "T1 application-layer guard regressions in lifecycle_guard_test.go: full on a checked story records and leaves status checked; request_changes full on a checked story reopens it to in-progress in DB and plan; gate on a checked story still refuses story_not_checkable; run_not_latest and unknown-run refusals unchanged"
          - "T2 rebuild round-trip regression in rebuild_test.go: a Validation entry carrying mode `full` reconstructs checks.mode='full'; a legacy entry without a mode segment yields ''; schema version bumps exactly once with migration 0014"
          - "T3 interfaces regressions: check record rejects --mode values outside {gate,full}; missing flag defaults to gate; query checks renders a mode field"
        verification: "each fails on unmodified code: cd cli && go test ./internal/application/ -run 'CheckRecord|LifecycleGuard' ; go test ./internal/application/ -run 'Rebuild' ; go test ./internal/interfaces/ -run 'Check' — stop condition: cannot fail-first without touching production → escalate to brainstorm refine"
      - wave: 2
        tasks:
          - "T4 migration 0014_checks_mode: ALTER TABLE checks ADD COLUMN mode TEXT NOT NULL DEFAULT ''; bump CurrentSchemaVersion() once; extend migrations_test.go assertions"
          - "T5 domain + application: domain.Check.Mode with {gate,full} validation; RecordCheck(..., mode) persists the column; guard allows mode=full on status=checked (clean verdict keeps checked, request_changes reopens to in-progress in DB + plan phase status); every other refusal unchanged; Validation entry format gains mode: `<gate|full>`."
          - "T6 rebuild: extend the check-entry regex with an optional mode capture; insert parsed mode (empty for legacy lines)"
          - "T7 interfaces + docs: register --mode flag (default gate, choices enforced); render mode in query checks; rewrite docs/playbooks/handoff.md step 6 to verify the final-phase clean full verdict from query checks and delete the 'does not yet persist' caveat"
        verification: "cd cli && go test ./internal/application/ ./internal/interfaces/ ./internal/domain/ ./internal/infrastructure/"
      - wave: 3
        tasks:
          - "T8 scratch-dir live proof via CLI only: init temp harness, create plan/story/run, check record (gate) → story checked; check record --mode full on same run → succeeds returns ID; another gate attempt → exit 1 story_not_checkable; db rebuild --yes → mode survives; query checks shows both rows with correct modes"
          - "Full gate: cd cli && go test ./... && bash scripts/verify-doc-links.sh"
        verification: "all commands exit 0; live sequence exit codes asserted inline; rebuild preserves both checks with modes"
    checks:
      - "focused suites exit 0 including new regressions"
      - "live gate→full→refuse transcript recorded in Progress with exit codes"
      - "db rebuild round-trips modes from markdown alone"
      - "verify-doc-links.sh 0 findings"
  - phase_slug: installer-draft-race-guard
    story_id: 01M0YJVRPVNCS4SH59H3XTD4YS
    status: done
    goal: Make install-zharness.sh fail loudly before downloading anything when a newer draft zharness release is still publishing (GitHub #64)
    depends_on: null
    requirements: [R5]
    allowed_surfaces: [scripts/install-zharness.sh, scripts/test-install-zharness.sh new]
    avoided_surfaces: [.github/workflows/release.yml, goreleaser config, CLI sources]
    waves:
      - wave: 1
        tasks:
          - "T9 self-contained stub test scripts/test-install-zharness.sh with fake gh on PATH: S1 draft newer than newest published must exit non-zero mentioning 'still publishing' with zero release-download calls; S2 stale older draft installs newest published (no false positive); S3 no drafts unchanged resolution; S4 explicit tag skips resolution entirely"
        verification: "bash scripts/test-install-zharness.sh fails on unmodified script (S1 silently downloads previous release instead of exiting); bash -n both scripts clean"
      - wave: 2
        tasks:
          - "T10 resolution-path fix: one gh release list call fetching isDraft+createdAt+tagName TSV rows; pure-bash max-createdAt comparison independent of gh ordering; newest draft strictly newer than newest published → stderr 'error: release still publishing — retry or pass an explicit tag', exit 1 before any download; ZHARNESS_INSTALL_DIR env override for hermetic tests; explicit-tag branch untouched"
        verification: "bash scripts/test-install-zharness.sh green (all scenarios PASS); bash -n scripts/install-zharness.sh clean"
      - wave: 3
        tasks:
          - "Full gate: cd cli && go test ./... && bash scripts/verify-doc-links.sh"
        verification: "exit 0"
    checks:
      - "stub suite green with failing-first evidence for S1"
      - "S2 proves an older draft does not block install"
      - "full gate exit 0"

## Progress
<!-- Append-only durable entries record timestamp, phase, wave, task, task_status, run_id, trace_id, exact verification/result, and changed surfaces or blocker. -->
- `2026-08-26T08:48:08Z` — wave 1, task check-mode-persistence execution start. task_status: `DONE`. run: `01M0YM0DCSKFVPH62RP8P8X1Q0`. summary: run created for check-mode-persistence; wave 1 failing regressions beginning; parallel sub-agent already implemented installer-draft-race-guard T9/T10 on disjoint surfaces (stub suite 4/4 PASS, failing-first S1 proof captured).
- `2026-08-26T09:14:50Z` — wave 1, task T1 application-layer guard regressions. task_status: `DONE`. run: `01M0YM0DCSKFVPH62RP8P8X1Q0`. summary: check_mode_test.go: full-on-checked succeeds/stays checked, request_changes full reopens to in-progress, gate-on-checked still story_not_checkable, done-story refused, invalid_check_mode rejected, empty mode persists gate; red proof captured (undefined domain.CheckMode*).
- `2026-08-26T09:14:50Z` — wave 1, task T2 rebuild round-trip regression. task_status: `DONE`. run: `01M0YM0DCSKFVPH62RP8P8X1Q0`. summary: round-trip fixture carries mode segment asserting rebuilt mode=full; new TestRebuildFromMarkdownLegacyCheckEntryWithoutMode proves legacy entry yields empty mode; migrations_test wants 0014_checks_mode + schemaVersion 14 twice; red proof captured.
- `2026-08-26T09:14:50Z` — wave 1, task T3 interfaces regressions. task_status: `DONE`. run: `01M0YM0DCSKFVPH62RP8P8X1Q0`. summary: invalid --mode rejected invalid_check_mode at interfaces layer before DB work; default flag value gate persisted; query checks + query check --latest expose mode field; red proof: unknown flag --mode.
- `2026-08-26T09:15:31Z` — wave 2, task T4 migration 0014_checks_mode. task_status: `DONE`. run: `01M0YM0DCSKFVPH62RP8P8X1Q0`. summary: ALTER TABLE checks ADD COLUMN mode TEXT NOT NULL DEFAULT empty-string; CurrentSchemaVersion bumps to 14; migrations_test asserts applied list + schemaVersion 14 on first and second Migrate.
- `2026-08-26T09:15:31Z` — wave 2, task T5 domain + application guard. task_status: `DONE`. run: `01M0YM0DCSKFVPH62RP8P8X1Q0`. summary: domain CheckMode constants + IsValidCheckMode + Check.Mode validation; RecordCheck takes mode (empty normalizes gate), persists column, allows full-on-checked only, request_changes full reopens checked story in DB and plan phase status.
- `2026-08-26T09:15:31Z` — wave 2, task T6 rebuild mode parsing. task_status: `DONE`. run: `01M0YM0DCSKFVPH62RP8P8X1Q0`. summary: checkValidationHeader regex optional mode capture group between run and phase segments; rebuild INSERT carries mode; legacy entries parse as empty string.
- `2026-08-26T09:15:31Z` — wave 2, task T7 interfaces flag + views + docs. task_status: `DONE`. run: `01M0YM0DCSKFVPH62RP8P8X1Q0`. summary: --mode flag registered default gate with interfaces-layer invalid_check_mode rejection; CheckListView and CheckView expose mode; handoff.md step 6 rewritten in BOTH copies (docs/playbooks + cli/docs/embedded) to verify from query checks instead of the does-not-yet-persist caveat; CONTRACT.md updated.
- `2026-08-26T09:15:31Z` — wave 2. run: `01M0YM0DCSKFVPH62RP8P8X1Q0`. summary: wave 2 complete: migration 0014, guard widening, rebuild round-trip, flag+views+docs all implemented; full focused suite green (go test ./internal/... 5 packages ok) and verify-doc-links 0 findings.
- `2026-08-26T09:58:40Z` — wave 3, task T8 scratch-dir live proof. task_status: `DONE`. run: `01M0YM0DCSKFVPH62RP8P8X1Q0`. summary: 6-step transcript all green: gate record moves plan phase to checked; full record on checked story succeeds (the #63 fix) and Validation entry carries mode segment; third gate attempt exits 1 story_not_checkable; query checks shows exactly 2 rows modes [gate full]; rm harness.db + db rebuild --yes reconstructs schema_version=14 checks=2 with modes [gate full] from markdown alone; request_changes full reopens story to in-progress in plan. Proof script /tmp/opencode/t8-live-proof.sh.
- `2026-08-26T09:58:40Z` — wave 3, task Full gate. task_status: `DONE`. run: `01M0YM0DCSKFVPH62RP8P8X1Q0`. summary: cd cli && go test ./... ok (7 packages incl cmd/zharness); bash scripts/verify-doc-links.sh 0 findings; bash -n clean on both installer scripts.
- `2026-08-26T09:58:40Z` — wave 3. run: `01M0YM0DCSKFVPH62RP8P8X1Q0`. summary: check-mode-persistence waves complete: failing-first regressions, implementation, live CLI proof and full gate all green.
- `2026-08-26T09:59:53Z` — wave 3, task gate: name aspect not independently verified. task_status: `DONE`. run: `01M0YM0DCSKFVPH62RP8P8X1Q0`. summary: same-session APPROVED 01M0YR55HA4RA92V512F8XBX6W: the T8 scratch-dir transcript exit codes were asserted by the proof script itself and not re-derived by an independent observer; automated gates (go test ./..., verify-doc-links) were independently re-executed by the CLI proof verifier.
- `2026-08-26T10:00:08Z` — wave 1, task T9 stub regression suite. task_status: `DONE`. run: `01M0YR630Z3CM07ZM070JYM4BW`. summary: scripts/test-install-zharness.sh: hermetic fake-gh harness, S1 race refuses before download (failing-first proof /tmp/opencode/t9-failing-first.txt showed silent exit 0 + 1 download on unmodified script), S2 stale draft installs newest published, S3 no-draft unchanged, S4 explicit tag skips resolution; implemented by parallel sub-agent on disjoint surfaces.
- `2026-08-26T10:00:08Z` — wave 1, task T10 resolution-path fix. task_status: `DONE`. run: `01M0YR630Z3CM07ZM070JYM4BW`. summary: install-zharness.sh: one gh release list call with isDraft+createdAt+tagName TSV rows, pure-bash max-createdAt comparison order-independent, strictly-newer draft exits 1 still-publishing before any download; ZHARNESS_INSTALL_DIR override; explicit-tag path untouched; suite 4/4 PASS post-fix.
- `2026-08-26T10:00:27Z` — wave 3, task Live restore install (incident proof). task_status: `DONE`. run: `01M0YR630Z3CM07ZM070JYM4BW`. summary: fixed installer run against real repo after stub clobber incident: resolved v0.13.0, downloaded, installed to ~/.local/bin, zharness --version = 0.13.0; default-resolution path proven live.
- `2026-08-26T10:00:27Z` — wave 3, task Full gate. task_status: `DONE`. run: `01M0YR630Z3CM07ZM070JYM4BW`. summary: stub suite 4/4 PASS re-run this session; bash -n clean both scripts; cd cli && go test ./... ok; verify-doc-links 0 findings.
- `2026-08-26T10:00:27Z` — wave 3. run: `01M0YR630Z3CM07ZM070JYM4BW`. summary: installer-draft-race-guard waves complete: failing-first S1, fix implemented by parallel sub-agent, hermetic suite green, live default-install proven during incident restore.
- `2026-08-26T10:02:56Z` — handoff recorded. handoff: `01M0YRBN5D2HE4X5GSW5A2X5CS`. run: `01M0YM0DCSKFVPH62RP8P8X1Q0`. check: `01M0YR55HA4RA92V512F8XBX6W`. phase closed.
- `2026-08-26T10:02:56Z` — handoff recorded. handoff: `01M0YRBNHYHK1TDV3486Q62HQN`. run: `01M0YR630Z3CM07ZM070JYM4BW`. check: `01M0YRAYPVYXCB4B97C9VKW20Y`. phase closed. next action: commit the initiative diff, close GitHub #63 and #64, then brainstorm next initiative.

## Decisions
<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->
- `2026-08-26T08:48:08Z` — Phase blocks normalized from markdown-heading form to the canonical YAML-list form used by completed plans (memory-lifecycle.md), because query plan --section phase could not parse the heading form (degraded:true, no phase found); definition content is unchanged, only serialization. (phase: `check-mode-persistence`), task: planning normalization. rationale: to-plan wrote ###-heading phase blocks; the plan parser only discovers list-form blocks (phase_slug/story_id/waves). Without normalization the phase can never be queried, gated, or rebuilt from markdown..
- `2026-08-26T08:48:08Z` — Installer stub-suite incident: the T9 failing-first run executed against the pre-override script, so S2-S4 installed a fake binary into the real ~/.local/bin and clobbered zharness 0.13.0; restored by running the fixed installer itself (live proof of #64 default-resolution path, v0.13.0 reinstalled). Guard added to test-install-zharness.sh: refuses to run when script-under-test lacks ZHARNESS_INSTALL_DIR. (phase: `installer-draft-race-guard`), task: T9. rationale: ZHARNESS_INSTALL_DIR override only exists post-T10; hermeticity was not yet enforced during failing-first. The guard makes an un-overridable script-under-test a hard stop instead of a silent machine-state mutation..
- `2026-08-26T08:48:08Z` — Standalone real-repo install sanity task (former T11) folded away: the incident restore performed the exact default-invocation install against the real repo (v0.13.0 resolved+installed), and S3 covers no-draft resolution hermetically. (phase: `installer-draft-race-guard`), task: T11. rationale: Avoids re-installing over the user live binary as ceremony; both observable behaviors of R5 are proven by restore transcript plus stub suite..
- `2026-08-26T09:15:12Z` — Touched two surfaces outside phase allowed_surfaces: cli/internal/application/layout_backfill.go (checks copy spec now conditionally carries mode via tableHasColumn guard) and cli/docs/CONTRACT.md (check record Args/Errors/Atomic side effects + query check --latest and query checks JSON shapes document the mode field). (phase: `check-mode-persistence`), task: T5/T7. rationale: copyLifecycleRows dropped the mode column so migrate layout parity checks failed with created_at landing in mode (caught by a debug probe, column-order mismatch fixed); leaving it out would corrupt checks on every legacy layout migration. CONTRACT.md is the CLI contract whose claims must stay true or validate/audit doc-truth gates fail — same rule prior initiatives applied to citation repoints..
- `2026-08-26T09:15:12Z` — QueryLatestCheck CheckView gained Mode field; layout migration snapshot comparison now compares modes for parity. (phase: `check-mode-persistence`), task: T7. rationale: handoff step 6 verifies the final-phase full verdict from query check/checks output per R4; both views expose the persisted mode, and layout migration must prove the column survives the copy..
- `2026-08-26T10:03:03Z` — plan completed. rationale: every phase_slug is a done story.

## Validation
<!-- Append-only durable entries record timestamp, phase, exact command/result/output, run_id, check_id, verdict, and proof_gaps. -->
- `2026-08-26T09:59:23Z` — check. verdict: `APPROVED`. check: `01M0YR55HA4RA92V512F8XBX6W`. run: `01M0YM0DCSKFVPH62RP8P8X1Q0`. phase: `check-mode-persistence`. judge: `same-session` (ox-alpha-free).
  - `cd cli && go test ./...` → 7 packages ok
  - `bash scripts/verify-doc-links.sh` → doc links OK 0 findings
- `2026-08-26T10:01:17Z` — check. verdict: `APPROVED`. check: `01M0YR8N1XD8R21NW4MAGWVW1S`. run: `01M0YR630Z3CM07ZM070JYM4BW`. mode: `gate`. phase: `installer-draft-race-guard`. judge: `same-session` (ox-alpha-free).
  - `bash scripts/test-install-zharness.sh` → 4 passed 0 failed
  - `cd cli && go test ./...` → 7 packages ok
  - `bash scripts/verify-doc-links.sh` → doc links OK 0 findings
- `2026-08-26T10:02:33Z` — check. verdict: `APPROVED`. check: `01M0YRAYPVYXCB4B97C9VKW20Y`. run: `01M0YR630Z3CM07ZM070JYM4BW`. mode: `full`. phase: `installer-draft-race-guard`. judge: `same-session` (ox-alpha-free).
  - `bash scripts/test-install-zharness.sh` → 4 passed 0 failed
  - `cd cli && go test ./...` → 7 packages ok
  - `bash scripts/verify-doc-links.sh` → doc links OK 0 findings
  - `bash /tmp/opencode/t8-live-proof.sh` → ALL PROOF STEPS PASSED

## Current State and Next Action
- active_phase: none
- lifecycle_status: completed
- latest_run_id: 01M0YR630Z3CM07ZM070JYM4BW
- latest_trace_ids: [01M0YM2Q5XWXR0TTYK07M6B058]
- latest_check_id: none
- latest_handoff_id: none
- blockers: none
- open_items: [wave 1 failing regressions T1-T3, wave 2 implement T4-T7, wave 3 live proof T8 + gate; installer-draft-race-guard code done by parallel sub-agent pending its own run/gate]
- exact_next_action: commit the initiative diff, close GitHub #63 and #64 with evidence links, then brainstorm the next initiative
