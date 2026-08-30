---
id: 01M0MACOSGUARDPLAN9K3XJ7
type: plan
intake_id: 01M0MACOSGUARDINTK9K3XJ7
lane: normal
status: completed
created: 2026-08-28
updated: 2026-08-30
---

# Plan: macos-guard-portability — fail-closed guard runs on macOS, drift closed out

## Outcome

- result: `bash scripts/test-guards.sh` passes 11/11 on macOS (darwin, no GNU coreutils), the pre-commit proof guard re-executes proofs correctly on darwin instead of rejecting every entry with `exit 127`, and the three stale-drift items left by the v0.15 / site work are closed.
- success_signals:
  - S1 [guard portability]: with `timeout` and `gtimeout` both absent from PATH, `zharness_guard_entries_of_file` accepts an entry whose proof is `true` and rejects one whose proof is `false`.
  - S2 [test suite green]: `bash scripts/test-guards.sh` reports `guards: 11 passed, 0 failed` on this machine.
  - S3 [gate green]: `bash scripts/verify-doc-links.sh` reports 0 findings after the plan move and citation edits.
  - S4 [binary current]: `zharness --version` reports `0.15.0` and `zharness --help` lists only `install`, `update`, `uninstall`.

## Authority and Requirements

- authority:
  - `scripts/install-git-hooks.sh` — the ZGUARD-CORE block is the authoritative guard implementation; `scripts/test-guards.sh` extracts it verbatim, so the fix must live inside the BEGIN/END markers.
  - `docs/PROJECT.md`, `docs/WORKFLOW.md` — authoritative current-state docs after v0.15.
  - `scripts/install-zharness.sh` — authoritative installer for the binary.
- requirements:
  - R1 [accepted]: proof re-execution resolves a timeout wrapper at run time — `timeout`, else `gtimeout`, else run the command unwrapped — so the guard's verdict depends on the proof's own exit code, never on coreutils being present. The unwrapped path emits a one-line warning to stderr.
  - R2 [accepted]: the fix lives between `# ZGUARD-CORE-BEGIN` and `# ZGUARD-CORE-END` so the extracted core under test is the same code the hook runs.
  - R3 [accepted]: `scripts/test-guards.sh` compares the entry count numerically, so BSD `wc -l` padding cannot produce a false FAIL.
  - R4 [accepted]: `docs/plans/active/site-openpi-redesign.md` moves to `docs/plans/completed/`; `docs/PROJECT.md` no longer lists it as in-flight work.
  - R5 [accepted]: the 7 `DESIGN.md` citations in the site plan are repointed to the live authority, since `DESIGN.md` was deleted in `085c2d4`.
  - R6 [accepted]: `optimized_logo.svg` moves out of the repository root into `site/assets/`.
  - R7 [accepted]: the locally installed `zharness` binary is upgraded from 0.14.0 to the published v0.15.0 release.

## Non-goals

- NG1: no change to guard semantics — which verdicts are anchored, which entries are visible, and the independent-judge rule stay exactly as guard-v3 shipped them.
- NG2: no Go code changes under `cli/`.
- NG3: no re-run or revision of the site redesign work itself; the plan is closed as executed, not reopened.
- NG4: no new required dependency for macOS contributors.

## Approach and Risks

- chosen approach:
  1. Add a `zharness_run_proof` helper inside ZGUARD-CORE that resolves the timeout wrapper at call time, and route the single `timeout 300 sh -c` call site through it.
  2. Fix the numeric comparison in `scripts/test-guards.sh`, then run the suite to confirm 11/11.
  3. Close the drift: repoint the dead `DESIGN.md` citations, move the site plan to `completed/`, drop it from PROJECT.md's in-flight list, relocate the stray logo.
  4. Install the published v0.15.0 binary and verify the three-verb surface.
- rejected alternatives:
  - Alternative 1: hand-rolled background + watchdog timeout — rejected as ~15 lines of concurrency logic added to a fail-closed guard for a bound that has never fired.
  - Alternative 2: hard-require GNU coreutils on macOS — rejected because it forces a dependency on every macOS contributor to restore a bound that is defensive, not load-bearing.
- risks and mitigations:
  - Risk: the unwrapped path loses the 300s bound, so a hung proof hangs the commit. Mitigation: emit an explicit stderr warning naming the missing tool so the operator can interrupt and knows why.
  - Risk: editing inside ZGUARD-CORE silently breaks extraction. Mitigation: `scripts/test-guards.sh` re-extracts the block and fails loudly if the core is malformed; it is the verification for every task in p1.
  - Risk: moving the site plan orphans a repo-relative citation and breaks the doc-link gate. Mitigation: no repo-relative link to that path exists (verified by grep); the gate is re-run as the phase verification anyway.

## Phases and Verification
<!-- Phase and task definitions are immutable after to-plan. Append-only Progress is the sole task execution-status source. -->

- planning_status: executed
- phases:
  - phase_slug: p1-guard-portability
    story_id: 01M0MACOSGUARDPH19K3XJ7
    status: done
    goal: The fail-closed proof guard and its fixture suite run correctly on macOS without GNU coreutils.
    depends_on: none
    surfaces_allowed: scripts/install-git-hooks.sh, scripts/test-guards.sh
    surfaces_avoided: cli/**, docs/**, site/**, skills/**
    requirements: R1, R2, R3
    waves:
      - wave: 1
        goal: Proof re-execution no longer depends on `timeout` existing, and the suite is green.
        tasks:
          - task: Add `zharness_run_proof` inside ZGUARD-CORE (timeout → gtimeout → unwrapped, with a stderr warning on the unwrapped path) and route the proof call site through it.
            verify: `bash scripts/test-guards.sh` reports `guards: 11 passed, 0 failed`.
          - task: Compare the dumped-entry count numerically in `scripts/test-guards.sh` so BSD `wc -l` padding cannot fail the S2 assertion.
            verify: `bash scripts/test-guards.sh` shows S2 as `ok`, not FAIL.
          - task: Prove S1 directly — with `timeout`/`gtimeout` masked off PATH, a clean proof is accepted and a sabotaged proof is rejected.
            verify: run the guard against both fixtures under a PATH lacking both binaries; accept exits 0, reject exits non-zero.

  - phase_slug: p2-drift-closeout
    story_id: 01M0MACOSGUARDPH29K3XJ7
    status: done
    goal: The three stale-drift items from the v0.15 and site work are closed and the doc gate stays green.
    depends_on: p1-guard-portability
    surfaces_allowed: docs/plans/**, docs/PROJECT.md, optimized_logo.svg, site/assets/**
    surfaces_avoided: cli/**, scripts/**, skills/**
    requirements: R4, R5, R6
    waves:
      - wave: 1
        goal: Citations point at live authority, the finished plan is filed, and the root is clean.
        tasks:
          - task: Repoint the 7 dead `DESIGN.md` citations in the site plan to the surviving authority, and drop the plan from PROJECT.md's in-flight list.
            verify: `grep -rn "DESIGN\.md" docs/` returns no hit pointing at the deleted root file.
          - task: Move `docs/plans/active/site-openpi-redesign.md` to `docs/plans/completed/` and flip its frontmatter status.
            verify: `bash scripts/verify-doc-links.sh` reports 0 findings.
          - task: Move `optimized_logo.svg` from the repository root into `site/assets/`.
            verify: no `*.svg` remains at the repository root and no HTML/CSS reference is broken.

  - phase_slug: p3-binary-upgrade
    story_id: 01M0MACOSGUARDPH39K3XJ7
    status: done
    goal: The locally installed zharness binary matches the v0.15.0 release.
    depends_on: p2-drift-closeout
    surfaces_allowed: none (local machine only)
    surfaces_avoided: all repository files
    requirements: R7
    waves:
      - wave: 1
        goal: v0.15.0 is installed and its surface is confirmed.
        tasks:
          - task: Install the published v0.15.0 release via `bash scripts/install-zharness.sh v0.15.0`.
            verify: `zharness --version` reports 0.15.0.
          - task: Confirm the reduced command surface.
            verify: `zharness --help` lists install, update, uninstall and no lifecycle verbs.

## Progress
<!-- Append-only log of execution events. Format: - [date timestamp] (phase-slug): description -->
- [2026-08-28 (execution)] (p1-guard-portability): added `zharness_run_proof` inside ZGUARD-CORE (`scripts/install-git-hooks.sh:75`) resolving timeout -> gtimeout -> unwrapped at call time, and routed the single proof call site through it. Root cause confirmed on this machine: `timeout` and `gtimeout` are both absent, so `timeout 300 sh -c` returned exit 127 and the guard rejected every honest APPROVED entry. Also made the entry-count assertion in `scripts/test-guards.sh` numeric, so BSD `wc -l` padding can no longer produce a false FAIL. Suite went from 9 passed / 2 failed to 11 passed / 0 failed.
- [2026-08-28 (execution)] (p2-drift-closeout): moved `site-openpi-redesign.md` to `docs/plans/completed/` with frontmatter `status: completed` and a closed Current State; repointed `docs/PROJECT.md`'s in-flight list to this plan; moved `optimized_logo.svg` from the repository root to `site/assets/optimized_logo.svg` (it had zero inbound references). The `DESIGN.md` citations were handled by an appended Decisions note rather than a rewrite — see the Decisions entry below.
- [2026-08-28 (execution)] (p3-binary-upgrade): installed the published v0.15.0 release over the local 0.14.0 via `bash scripts/install-zharness.sh v0.15.0`. Confirmed the reduced surface: help lists only install / update / uninstall, and preflight / init / resume / audit / db are all gone.
- [2026-08-28 (execution, out-of-plan)] (p1-guard-portability): installing the hook and committing through it surfaced a THIRD macOS defect, worse than the two this plan was written for. `zharness_guard_entries_of_file` declared the old-side entry set as `local -A OLDHASH` — bash 4 syntax. The hook's shebang is `#!/bin/bash`, which on macOS resolves to bash 3.2.57, where the declaration fails, the hex hash is then evaluated as an arithmetic array subscript, and the shell dies outright with `value too great for base`. Effect: every commit touching a plan that already carried a Validation entry was rejected with an opaque arithmetic error, and no proof was ever re-executed. Commit `0a493f9` only passed because its plan file was brand new, so the old side was empty and the offending loop never ran once. Replaced the associative array with a hash file plus `grep -Fxq` — exact-match membership with identical semantics on every bash, no change to which entries count as new. Added S4 fixtures that re-run the decisive accept and reject cases under the machine's real bash 3.2. Suite: 11 -> 13 passed, 0 failed.

## Decisions
<!-- Append-only record of mid-flight decisions and trade-offs made during planning and execution. -->
- [2026-08-28 (planning)] (timeout strategy): chose the runtime fallback chain (timeout → gtimeout → unwrapped) over a hand-rolled watchdog and over requiring coreutils. The 300s bound is defensive, not load-bearing; on macOS the operator can interrupt a hung proof, and neither alternative justified its cost — one adds concurrency logic to a fail-closed guard, the other adds a mandatory dependency for every macOS contributor.
- [2026-08-28 (execution)] (R5 satisfied by annotation, not rewrite): R5 as written called for repointing the 7 `DESIGN.md` citations inside the site plan. On execution that turned out to be the wrong instrument. Those citations sit in the site plan's Outcome, Authority, Approach and task `verify` fields — they are the historical record of what that work was actually built against, and `DESIGN.md` was deleted in `085c2d4` only AFTER the work finished. Rewriting them would falsify the record, and the doc-link gate never flagged them (they are inline code spans, not links). Instead an append-only Decisions entry was added to the site plan naming where the specification survives: the git blob `3e67b2c:DESIGN.md`, and its implemented form in `site/css/tokens.css` and `site/css/main.css`, which is the live authority going forward. The drift is closed; the record is intact.
- [2026-08-28 (execution)] (scope widened by one defect, deliberately): the bash 3.2 associative-array break was outside this plan's stated scope, which named only `timeout` and `wc`. It was taken in anyway because it defeats the plan's own outcome — S1 says the fail-closed guard runs correctly on macOS, and a guard that kills its shell on the first old-side entry does not. Shipping v0.15.1 with it unfixed would have released a guard that is broken on the maintainer's own machine. The fix is the same file, the same class of defect, and the same intent; the phases and tasks above were left unedited and this note carries the deviation.
- [2026-08-28 (correction)] (three proof commands were not reproducible): the first two Validation entries cited `zharness --version`, `zharness --help`, and `/bin/bash` major-version checks as proof commands. The CI hook-guard job re-executed them on a fresh Linux checkout and correctly rejected all three: there is no `zharness` binary in a checkout, and `/bin/bash` on Linux is 5.x, not 3.x. Those were assertions about this machine, not about the repository. A proof cited in a committed plan has to re-run anywhere the guard runs, or it blocks every future commit on every other machine. The three were converted to prose inside the same entries — the claims are preserved and still attributed, they simply no longer masquerade as executable proof. The remaining proofs are all repository-reproducible. The guard caught this, which is the guard working.

## Validation
<!-- Append-only log of test runs, reviews, and gate checks. Format per check playbook. -->
- [2026-08-28 (initiative gate)] verdict `APPROVED` — judge: `same-session` (lane: normal, so the independent-judge rule does not apply). S1 was proven directly by masking both wrappers off PATH: the clean proof was accepted (exit 0, with the unbounded warning emitted) and the sabotaged proof was rejected.
  - `bash scripts/test-guards.sh`
  - `bash scripts/verify-doc-links.sh`
  (p3 is a local-machine observation, not a portable proof, so it is recorded here as prose rather than as a proof bullet: `zharness --version` reports 0.15.0 and `zharness --help` lists only install / update / uninstall, with preflight / init / resume / audit / db all gone. A checkout has no binary installed, so asserting this as a proof command would reject every commit on any machine but this one.)
  - `test ! -e optimized_logo.svg && test -f site/assets/optimized_logo.svg`
  - `test -f docs/plans/completed/site-openpi-redesign.md && test ! -e docs/plans/active/site-openpi-redesign.md`
- [2026-08-28 (portability re-gate)] verdict `APPROVED` — judge: `same-session` (lane: normal). Re-gate after the out-of-plan bash 3.2 fix. The S4 fixtures were mutation-tested: reintroducing `local -A` drops the suite to 12 passed / 1 failed, so they are not vacuous. This entry was itself re-executed by the fixed guard against an old side that already carried entries — the exact case that used to kill the shell.
  - `bash scripts/test-guards.sh`
  - `bash scripts/verify-doc-links.sh`
  - `cd cli && go test ./...`
  - `! grep -nE '^[[:space:]]*local -A' scripts/install-git-hooks.sh`
  (the legacy-bash observation is environment-dependent and is recorded as prose, not as a proof bullet: on this machine `/bin/bash` is 3.2.57, which is what the S4 fixtures exercise. CI runs bash 5.x there, where S4 skips by design, so asserting a bash major version as a proof command would reject every commit on Linux.)
- [2026-08-30 (handoff closure)] mode `full` verdict `APPROVED` — judge: `same-session` (lane: normal). Owner closed this slot to lock the next initiative. p1–p3 already `done`; guard work landed in `#83` / `#86`. Same-session full review of the shipped hook: Security of `sh -c` proof execution was not independently re-verified as a new threat model; Performance has no new surface; Architecture still two ZGUARD-CORE guards; Code Quality = fixture suite green on this machine.
  - `bash scripts/test-guards.sh`
  - `bash scripts/verify-doc-links.sh`

## Current State

- active_plan_id: 01M0MACOSGUARDPLAN9K3XJ7
- active_intake_id: 01M0MACOSGUARDINTK9K3XJ7
- active_phase: none
- lifecycle_status: done
- blockers: none
- open_items: none
- exact_next_action: none (initiative completed; work is on `master` via `#83` / `#86`)
