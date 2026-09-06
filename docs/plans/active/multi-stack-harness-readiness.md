---
id: 01M0MULTISTACKHRN9K4XJ2
type: plan
intake_id: 01M0MULTISTACKHRNINTK9K4XJ2
lane: normal
status: active
created: 2026-09-06
updated: 2026-09-06
---

# Plan: multi-stack-harness-readiness — spine docs and the git skill stop assuming a Node repo, and the projection invariant gets a guard

## Outcome
- result: the spine playbooks, the identity template, and the `git` skill state
  what is actually true for a Go, Rust, or Python repository — the base branch is
  resolved rather than assumed, the proof re-execution contract (cwd, shell
  state, time bound) is written down, the enforcement claim matches what
  `zharness install` really ships, the identity file has one slot per gate class,
  and dependency/test staging is not `package.json`-shaped. The
  embedded-to-projected byte invariant is enforced by a check that runs in CI
  instead of being asserted once in a completed plan.
- success_signals:
  - A one-byte edit to any file under `docs/playbooks/` that is not mirrored into
    `cli/docs/embedded/playbooks/` fails an automated check, and that check runs
    on a projection-only change.
  - `docs/playbooks/watzup.md` step 1 produces a working base-branch comparison in
    a repository whose default branch is `master` and in one with no remote.
  - A Validation proof written as `cd cli && go test ./...` survives pre-commit
    guard re-execution while a bare `go test ./...` fails it, and `check.md`
    states which of the two is correct and why.
  - `docs/playbooks/check.md` never claims an enforcement level higher than what
    the repository actually installed.
  - A filled `docs/PROJECT.md` answers `tests`, `types`, `lint`, `build`, and
    `format` separately, or marks each `n/a`.
  - `skills/workflow/git/references/workflow.md` classifies a `go.mod`/`go.sum`,
    `Cargo.toml`/`Cargo.lock`, or `pyproject.toml`/`uv.lock` pair as one
    dependency change, and keeps a `*_test.go` with the source it covers.

## Authority and Requirements
- authority:
  - `docs/audit/multi-stack-agent-readiness-audit.md` — 13 findings across 6 rubric questions (cited below as `A1-<id>`).
  - `docs/audit/multi-stack-harness-audit.md` — 11 findings in 5 cause families (cited below as `A2-<id>`).
  - `cli/docs/CONTRACT.md:71` — `cli/docs/embedded/playbooks/` is canonical, projected to `docs/playbooks/`.
  - `README.md:72` — `zharness install` does not install git hooks.
  - `docs/patterns/encoding-invariants.md:59-71` — the Local validation / Optional hook / CI / Branch protection enforcement ladder.
  - `docs/plans/completed/playbook-truth-and-guards.md:17,83` — invariant S2 (`docs/playbooks/*.md` byte-identical to the embedded source) was accepted and proven once by `diff -q`, then never encoded as a guard.
  - `docs/decisions/0007-fresh-overwrite-for-playbooks-and-workflow.md` — fresh-overwrite for playbooks and WORKFLOW.md on update; three-way merge only for `docs/PROJECT.md` and the marked `AGENTS.md` block.
  - Owner decisions, 2026-09-06 — the bounded-mode line threshold moves to its own ADR; both audit documents are committed and `multi-stack-harness-audit-prompt.md` was trashed.
  - Measured this session — `cli/internal/installer/threeway.go` merge behaviour, `scripts/install-git-hooks.sh` receipt handling, `.github/workflows/cli-ci.yml` path filters, and current projection parity.
- requirements:
  - R1 [accepted]: A repository-resident automated check fails when any file under `docs/playbooks/` or `docs/WORKFLOW.md` differs by one byte from its `cli/docs/embedded/` source, names the drifted file, and passes when every pair matches. | source: `cli/docs/CONTRACT.md:71`; invariant S2 in `docs/plans/completed/playbook-truth-and-guards.md:17`
  - R2 [accepted]: That check lives under `cli/` and both `paths:` lists in `.github/workflows/cli-ci.yml` include `docs/playbooks/**` and `docs/WORKFLOW.md`, so a projection-only edit triggers CI. | source: measured — the current lists are `cli/**`, `docs/plans/**`, `scripts/install-git-hooks.sh`, `.github/workflows/cli-ci.yml`, none of which match a projected playbook
  - R3 [accepted]: `watzup` step 1 resolves the base branch through `origin/HEAD`, then a verified local `main`, then `master`, then the current branch, and produces a usable comparison in a repository with no remote. | source: `A1-X1`; `docs/playbooks/watzup.md:14` hardcodes `main` in all three commands
  - R4 [accepted]: `check.md` claims the pre-commit hook as the proof guarantee only for repositories that installed it, names the honest level otherwise using the `encoding-invariants.md:59-71` ladder, and points at `bash scripts/install-git-hooks.sh --force` as the opt-in. Both the step 9 claim and the step 7 `mode: full` rejection claim carry the qualifier. | source: `A2-C1`; `README.md:72`; `docs/playbooks/check.md:38,40`
  - R5 [accepted]: `check.md` states the proof re-execution contract in the Validation entry contract: proofs are re-executed from the repository root under `sh -c` with no inherited shell state (no activated virtualenv, no prior `cd`, no exported environment), bounded to 300 s where `timeout` or `gtimeout` is on PATH and unbounded otherwise, so a proof needing a directory or interpreter carries it inline. | source: `A1-G2`, `A1-P1`, `A1-R1`, `A2-G2`; `scripts/install-git-hooks.sh` `zharness_run_proof()` and the hook body's `ROOT="$(git rev-parse --show-toplevel)"` with no `cd`
  - R6 [accepted]: The `receipt:` field list in `check.md`'s Validation entry contract and in its Output Format block both carry `enforcement: hook | ci | local-only`, and `bash scripts/test-guards.sh` still passes unchanged. | source: measured — `grep 'receipt\|context_sources\|enforcement' scripts/install-git-hooks.sh` returns nothing, so the guard never parses the receipt block and the field is additive
  - R7 [accepted]: `cli/docs/embedded/templates/project.identity.md` carries one gate section with `run from:`, `tests:`, `types:`, `lint:`, `build:`, and `format:` slots, each accepting `n/a`, and the file stays at or under 50 lines. The section replaces `## How do we run the tests?` in place rather than being appended after it, and this repository's own `docs/PROJECT.md` is re-answered against the new section in the same phase, so mono-harness does not ship a template its own identity file no longer follows. | source: `A1-X2`, `A1-G3`, `A1-R2`, `A1-P3`, `A2-R2`, `A2-P2`; `check.md:34` runs a four-class gate against a template asking one question
  - R8 [accepted]: `zharness update` against a repository whose `docs/PROJECT.md` was filled from the old template conflicts on that file, writes conflict markers and a `.zharness/conflicts.json` entry, and `zharness update --abort` restores the pre-update file byte-identically. The conflict is deliberate: it is the only mechanism that puts the new gate questions in front of the owner. This is proven by a test under `cli/internal/installer/` driving a filled-template fixture — extending the shape of `TestUpdate_FastForward_Kept_AutoMerge_ConflictAbort` — never by a manual scratch-repo run, because the pre-commit guard re-executes every Validation proof from a bare checkout. | source: measured against `cli/internal/installer/threeway.go` — `diffHunks` absorbs change regions separated by 3 or fewer common lines, so a filled 22-line template is one hunk spanning lines 5-22 and any in-place edit overlaps it (CONFLICT), while an append at end of file does not (clean merge)
  - R9 [accepted]: `work.md` step 1 carries an absence branch for `scripts/plan-slice.sh` in the same shape `check.md:34` already uses for `scripts/record-check.sh`, naming a heading-range extraction that needs no distributed script. | source: `A2-C2`; `AllTargets()` in `cli/internal/installer/installer.go` distributes only `WORKFLOW.md`, the project template, and the playbooks — no scripts; `watzup.md:15` already hedges with "Prefer" but `work.md:32` states it as the method
  - R10 [accepted]: `skills/workflow/git/references/workflow.md` step 1 reviews `git status --porcelain` and stages explicit paths instead of running `git add -A`. | source: `A1-P2`; that same file's Anti-Patterns section already forbids `git add -A` because it "catches `.env`, secrets, `node_modules`"
  - R11 [accepted]: The `deps:` class in that file matches `go.mod`/`go.sum`, `Cargo.toml`/`Cargo.lock`, and `pyproject.toml`/`requirements*.txt`/`poetry.lock`/`uv.lock` alongside `package.json` and its lockfiles, and states that a manifest and its lockfile commit together. | source: `A2-G1`, `A2-R1`, `A2-P1`
  - R12 [accepted]: That file's `test:`/`code:` split states that a changed `*_test.go` commits with the source file it covers when both changed. | source: `A1-G1`; Go places `foo_test.go` beside `foo.go`, so the current split fragments one logical change into two commits, neither of which builds and passes alone
  - R13 [accepted]: Every playbook edit in this initiative is authored in `cli/docs/embedded/playbooks/` and projected to `docs/playbooks/` in the same commit, with R1's check passing on that commit. | source: `cli/docs/CONTRACT.md:71`; `docs/decisions/0007-fresh-overwrite-for-playbooks-and-workflow.md` — a consumer-side playbook edit is destroyed by the next `zharness update`
  - R14 [accepted]: `docs/audit/multi-stack-agent-readiness-audit.md` and `docs/audit/multi-stack-harness-audit.md` are tracked in git, so the requirements above cite authority that exists on a bare checkout. | source: owner decision, 2026-09-06

## Non-goals
- NG1: The bounded-mode ceiling at `work.md:17` ("five files or roughly 100 changed lines") is not touched. `A2-G3`, `A2-R3`, and `A2-P3` argue raw line count misfires when lockfile and generated churn inflates a diff, but changing what counts as a changed line is externally observable policy, not a defect. Owner decision 2026-09-06: it gets its own ADR under `docs/decisions/`.
- NG2: No `# timeout:` proof annotation is added. `A1-R1` proposes one, but `zharness_run_proof()` documents its own 300 s bound as "defensive, not load-bearing" and falls through to running the proof unbounded when neither `timeout` nor `gtimeout` is on PATH, which is the stock macOS case. R5 documents the real behaviour instead.
- NG3: `/.kit/` is not added to `gitignoreWants` in `cli/internal/installer/installer.go`. `A1-X3` is false for this repository — `.gitignore:5` already carries `**/.kit/` — and for consumers it is out of scope, because `zharness install` ships no skills and therefore never creates `.kit/`.
- NG4: No installer, hook, or CLI behaviour change. No new command, flag, config field, or schema. `AGENTS.md` — the binary is install / update / uninstall only.
- NG5: `zharness install` still does not install git hooks. R4 makes the documentation match that fact rather than changing the fact.
- NG6: Neither `scripts/plan-slice.sh` nor `scripts/record-check.sh` is added to the distributed set. R9 makes the playbook work without them; what `install` owns is a separate scope question.
- NG7: No rewrite of either audit document. They are committed as-authored evidence under R14, including the three findings this plan rejects.

## Approach and Risks
- approach: land the projection guard first so every later playbook edit is
  mechanically checked, then correct the spine playbooks, then the identity
  template, then the `git` skill. R13 is not a phase of its own — it is the
  working rule that Phase 1 makes enforceable and Phases 2-4 inherit. Phase 1
  adds a Go test under `cli/docs/embedded/` rather than a shell script plus a
  new CI step, because R2 puts the check under `cli/` and a test in that
  package rides the existing `go test ./...` step of the `build-test` job with
  no new workflow surface. Phases 3 and 4 open surfaces Phase 2 never touches
  (`cli/docs/embedded/templates/`, `cli/internal/installer/`,
  `skills/workflow/git/`), so they carry no dependency, can land in any order,
  and each leaves the repository usable alone.
- rejected_alternatives:
  - A `scripts/verify-projection.sh` invoked from a new CI step — rejected: R2
    requires the check under `cli/`, and a separate workflow step is a second
    thing to keep in sync with the `paths:` filters.
  - Making the parity check a pre-commit guard — rejected:
    `docs/patterns/encoding-invariants.md:59-71` forbids installing hooks as
    part of encoding a rule without separate authorization, and `zharness
    install` ships no hooks (NG5), so a hook-only guard is invisible to every
    consumer and to CI.
  - Appending the gate section to `project.identity.md` instead of replacing
    `## How do we run the tests?` — rejected: measured against
    `cli/internal/installer/threeway.go`, an end-of-file append auto-merges into
    unfilled placeholders nobody is prompted to answer, closing zero findings.
    The conflict R8 pins is the notification mechanism, not a defect.
- constraints:
  - Every playbook edit is authored in `cli/docs/embedded/playbooks/` and
    projected to `docs/playbooks/` in the same commit (R13); a
    projection-only edit is destroyed by the next `zharness update` per
    `docs/decisions/0007-fresh-overwrite-for-playbooks-and-workflow.md`.
  - No installer, hook, CLI, flag, config, or schema behaviour change (NG4).
    Phase 3's only Go change is a test.
  - Every Validation proof must survive pre-commit re-execution from the
    repository root under `sh -c` with no inherited shell state, so each proof
    carries its own `cd` and uses no binary-presence or OS-version check.
  - A proof that asserts a negative must still exit 0; write it as
    `sh -c '! <command that must fail> >/dev/null 2>&1'`.
  - `docs/PROJECT.md` already carries an uncommitted edit from `brainstorm`.
    Phase 3 owns the remaining change to that file; no other phase touches it.
- risks:
  - The parity test resolves `../../../docs/` from `cli/docs/embedded/`, which
    does not exist for someone running `go test` after `go install`.
    Mitigation: none added deliberately — the `cli` module is only tested inside
    this repository, and a `t.Skip` on a missing tree would let the gate pass
    silently, which is the exact failure mode invariant S2 already had.
  - A green parity test proves nothing about detection;
    `docs/patterns/encoding-invariants.md:55-56` requires an exercised
    violation. Mitigation: Phase 1 ships a `DetectsDrift` subtest that feeds a
    synthetic one-byte difference through the same comparator in the same run.
  - Adding `docs/playbooks/**` and `docs/WORKFLOW.md` to the CI `paths:` filters
    makes every playbook edit run the full Go build/vet/test job. Accepted —
    that is precisely what R2 asks for, and the job builds one small module.
  - R7's in-place template rewrite conflicts on `zharness update` for every
    consumer holding a filled `docs/PROJECT.md`. Deliberate per R8; recovery is
    `zharness update --abort`, which Phase 3's test proves restores bytes
    exactly.
  - R6 adds a key to the `receipt:` block. Measured today no guard parses
    receipts, but a future strict parser could reject an unknown key.
    Mitigation: Phase 2 re-runs `bash scripts/test-guards.sh` as a proof.
- stop_conditions:
  - Phase 1: if the parity test fails on an otherwise unmodified tree, the
    projection has already drifted. Stop, record `BLOCKED_CONTRACT_DRIFT`,
    reconcile `docs/playbooks/` against `cli/docs/embedded/playbooks/` as its
    own commit, then add the gate.
  - Phase 3: if the installer test shows a filled template auto-merging instead
    of conflicting, R8's premise is wrong. Stop and re-open the template-shape
    decision rather than weakening the assertion.
  - Any phase: if a requirement can only be satisfied by an installer, hook, or
    CLI behaviour change, stop — that is NG4 and needs new authority.
- escalation: record the blocker in `## Progress`, set `blockers:` in
  `## Current State and Next Action`, and route to `brainstorm refine` for a
  scope or authority change. Do not work around a stop condition inside `work`.
- recovery: each phase is one commit on `master` with its proofs recorded in
  `## Validation`. `git revert` of that commit is the rollback; no phase leaves
  external state changed.

## Phases and Verification
<!-- Phase and task definitions are immutable after to-plan. Do not add task status fields. Append-only Progress is the sole task execution-status source. Only each phase lifecycle status changes: to-plan=planned; work=in-progress; clean durable check=checked; closing handoff=done. Each planned phase records phase_slug, story_id, status, goal, depends_on, waves, tasks, and checks. -->
- planning_status: planned
- phases:
  - phase_slug: projection-parity-gate
    story_id: 01M0MULTISTACKHRNS1P9K4XJ2
    status: checked
    goal: one-byte drift between `cli/docs/embedded/` and its projection under
      `docs/` fails an automated check that CI actually runs on a
      projection-only change, and the audit documents this plan cites exist on
      a bare checkout. Satisfies R1, R2, R14 and makes R13 enforceable.
    depends_on: none
    allowed_surfaces: `cli/docs/embedded/parity_test.go` (new),
      `.github/workflows/cli-ci.yml`, git tracking of `docs/audit/*.md`.
    avoided_surfaces: `docs/playbooks/`, `cli/docs/embedded/playbooks/`,
      `cli/docs/embedded/templates/`, `cli/internal/`, `scripts/`,
      `skills/workflow/`.
    waves:
      - wave: W1 — the check exists
        tasks:
          - task: T1.1 — add `cli/docs/embedded/parity_test.go` in package
              `embedded` with `TestProjectionParity`. It walks `FS` for
              `WORKFLOW.md` and every `playbooks/*.md`, compares each against
              `../../../docs/WORKFLOW.md` and `../../../docs/playbooks/<name>`
              byte for byte, and fails naming the drifted or missing file. It
              then lists `../../../docs/playbooks/` and fails naming any file
              with no embedded counterpart, so the assertion runs in both
              directions. Reading a missing tree is a hard failure — no
              `t.Skip`, because a skip lets the gate pass silently, which is
              how invariant S2 was lost. A `DetectsDrift` subtest feeds an
              in-memory pair differing by one byte through the same comparator
              and asserts it reports a mismatch naming that file, satisfying
              `docs/patterns/encoding-invariants.md:55-56`.
            proof: `cd cli && go test ./docs/embedded/ -run TestProjectionParity -count=1`
      - wave: W2 — CI sees it, and the cited authority is tracked
        tasks:
          - task: T1.2 — add `docs/playbooks/**` and `docs/WORKFLOW.md` to both
              the `push.paths` and the `pull_request.paths` lists in
              `.github/workflows/cli-ci.yml`, so a projection-only edit
              triggers `build-test`.
            proof: `sh -c 'test "$(grep -c "docs/playbooks/\*\*" .github/workflows/cli-ci.yml)" = 2 && test "$(grep -c "docs/WORKFLOW.md" .github/workflows/cli-ci.yml)" = 2'`
          - task: T1.3 — stage
              `docs/audit/multi-stack-agent-readiness-audit.md` and
              `docs/audit/multi-stack-harness-audit.md` unmodified (NG7) so
              both are tracked by this commit.
            proof: `git ls-files --error-unmatch docs/audit/multi-stack-agent-readiness-audit.md docs/audit/multi-stack-harness-audit.md`
            note: this reads the index, so it passes during pre-commit
              re-execution exactly when both files are staged in this commit,
              which is the condition R14 asks for.
    expected_outputs: new `cli/docs/embedded/parity_test.go`; modified
      `.github/workflows/cli-ci.yml`; two audit documents tracked.
    checks:
      - `cd cli && go build ./... && go vet ./... && go test ./...`
      - `bash scripts/verify-doc-links.sh`
      - `check full`

  - phase_slug: playbook-truth
    story_id: 01M0MULTISTACKHRNS2P9K4XJ2
    status: checked
    goal: the three spine playbooks stop asserting things that are false
      outside a Node repository with a `main` branch and an installed hook —
      resolved base branch, written-down proof re-execution contract, honest
      enforcement claim, declared enforcement level, and a plan-slice absence
      branch. Satisfies R3, R4, R5, R6, R9 under R13.
    depends_on: projection-parity-gate
    allowed_surfaces: `cli/docs/embedded/playbooks/{watzup,work,check}.md` and
      their projections at `docs/playbooks/{watzup,work,check}.md`.
    avoided_surfaces: `work.md:17`'s bounded-mode ceiling (NG1), `scripts/`,
      `cli/internal/`, `cli/docs/embedded/templates/`, `skills/workflow/git/`.
    waves:
      - wave: W1 — authoring, independent files
        tasks:
          - task: T2.1 — rewrite `cli/docs/embedded/playbooks/watzup.md` step 1
              to resolve the base branch before comparing: `origin/HEAD`
              stripped of its `origin/` prefix, else the first of `main` then
              `master` that `git show-ref --verify --quiet refs/heads/<c>`
              confirms, else `git branch --show-current` — which is used rather
              than `git rev-parse --abbrev-ref HEAD` because it succeeds on an
              unborn branch in a fresh `git init`. `git log` and
              `git rev-list --left-right --count` then use that value, and the
              step states that when the resolved base equals the current
              branch the comparison is empty and the recap says so.
            proof: `sh -c 'f=cli/docs/embedded/playbooks/watzup.md; grep -q "origin/HEAD" "$f" && grep -q "show-ref --verify" "$f" && grep -q "^1\\. \\*\\*Read branch state" "$f" && ! awk "/^1\\. \\*\\*Read branch state/,/^2\\./" "$f" | grep -q "main\\.\\.HEAD"'` and `sh -c 'd=$(mktemp -d); git -C "$d" init -q; git -C "$d" -c user.name=t -c user.email=t@t commit -q --allow-empty -m x; git -C "$d" update-ref refs/remotes/origin/master HEAD; git -C "$d" symbolic-ref refs/remotes/origin/HEAD refs/remotes/origin/master; test "$(git -C "$d" symbolic-ref --quiet --short refs/remotes/origin/HEAD | sed "s|^origin/||")" = master'`
            note: the first half asserts the step-1 heading still exists before
              the `awk` range runs, so a renamed label cannot make the
              `main..HEAD` absence vacuously true. The second half builds a
              throwaway repository with an `origin/HEAD` symref rather than
              reading this checkout, because a CI clone
              (`actions/checkout` with `fetch-depth: 2`) has no
              `refs/remotes/origin/HEAD` and no local `main`.
          - task: T2.2 — add the no-remote branch of the same resolution as an
              executable proof, run against a throwaway `git init` repository
              with no remote, no commit, and no configured identity.
            proof: `sh -c 'grep -q "branch --show-current" cli/docs/embedded/playbooks/watzup.md'` and `sh -c 'd=$(mktemp -d); git -C "$d" init -q; base=$(git -C "$d" symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null | sed "s|^origin/||"); [ -n "$base" ] || for c in main master; do git -C "$d" show-ref --verify --quiet "refs/heads/$c" && base=$c && break; done; [ -n "$base" ] || base=$(git -C "$d" branch --show-current); test -n "$base"'`
          - task: T2.3 — add an absence branch for `scripts/plan-slice.sh` to
              `cli/docs/embedded/playbooks/work.md` step 1, in the shape
              `check.md` already uses for `scripts/record-check.sh`: if the
              script exists, slice with it; if it is absent, extract the
              heading range with `awk` and still never read the whole file.
              `AllTargets()` distributes no scripts, so the absent case is the
              consumer default.
            proof: `sh -c 'f=cli/docs/embedded/playbooks/work.md; grep -q "plan-slice.sh" "$f" && grep -q "absent" "$f" && grep -q "awk" "$f"'`
      - wave: W2 — `check.md`, three edits to one file
        tasks:
          - task: T2.4 — qualify both enforcement claims in
              `cli/docs/embedded/playbooks/check.md`: the step 9 "sole proof
              guarantee" line and the step 7 `mode: full` rejection line each
              state that this holds in a repository that installed the hook,
              name the honest ladder level from
              `docs/patterns/encoding-invariants.md` otherwise, and point at
              `bash scripts/install-git-hooks.sh --force` as the opt-in.
            proof: `sh -c 'test "$(grep -c "install-git-hooks.sh --force" cli/docs/embedded/playbooks/check.md)" -ge 2'`
          - task: T2.5 — state the proof re-execution contract in the Validation
              entry contract: proofs are re-executed from the repository root
              under `sh -c` with no inherited shell state — no activated
              virtualenv, no prior `cd`, no exported environment — bounded to
              300 s where `timeout` or `gtimeout` is on PATH and unbounded
              otherwise, so a proof needing a directory or interpreter carries
              it inline.
            proof: `sh -c 'f=cli/docs/embedded/playbooks/check.md; grep -q "repository root" "$f" && grep -q "no inherited shell state" "$f" && grep -q "300" "$f"'`
          - task: T2.6 — add `enforcement: hook | ci | local-only` to the
              `receipt:` field list in the Validation entry contract and to the
              `receipt:` line of the Output Format block.
            proof: `sh -c 'test "$(grep -c enforcement cli/docs/embedded/playbooks/check.md)" -ge 2'`
      - wave: W3 — projection
        tasks:
          - task: T2.7 — copy the three edited files to
              `docs/playbooks/{watzup,work,check}.md` byte-identically in this
              same commit (R13), and confirm the Phase 1 gate passes on it.
            proof: `cd cli && go test ./docs/embedded/ -run TestProjectionParity -count=1`
    expected_outputs: modified `cli/docs/embedded/playbooks/{watzup,work,check}.md`
      and their byte-identical projections under `docs/playbooks/`.
    checks:
      - `cd cli && go test ./docs/embedded/ -run TestProjectionParity -count=1`
      - `bash scripts/test-guards.sh`
      - `bash scripts/verify-doc-links.sh`
      - `check full`

  - phase_slug: identity-gate-slots
    story_id: 01M0MULTISTACKHRNS3P9K4XJ2
    status: checked
    goal: the identity template asks one question per gate class instead of one
      question about tests, this repository answers the new shape, and the
      resulting `zharness update` conflict is proven to be recoverable.
      Satisfies R7 and R8.
    depends_on: none
    allowed_surfaces: `cli/docs/embedded/templates/project.identity.md`,
      `docs/PROJECT.md`, `cli/internal/installer/installer_test.go`.
    avoided_surfaces: `cli/internal/installer/threeway.go`,
      `cli/internal/installer/installer.go`, and every other non-test file
      under `cli/internal/` (NG4).
    waves:
      - wave: W1 — the template and this repository's answer
        tasks:
          - task: T3.1 — replace `## How do we run the tests?` in
              `cli/docs/embedded/templates/project.identity.md` in place with a
              single gate section carrying `run from:`, `tests:`, `types:`,
              `lint:`, `build:`, and `format:` slots, each explicitly accepting
              `n/a`. Replacement, not an append: an append auto-merges and
              notifies nobody. The file stays at or under 50 lines. Preserve
              the pre-edit bytes for T3.3 via
              `git show HEAD:cli/docs/embedded/templates/project.identity.md`.
            proof: `sh -c 'f=cli/docs/embedded/templates/project.identity.md; for k in "run from:" "tests:" "types:" "lint:" "build:" "format:" "n/a"; do grep -q "$k" "$f" || exit 1; done; ! grep -q "^## How do we run the tests?" "$f"; test "$(wc -l < "$f")" -le 50'`
          - task: T3.2 — re-answer `docs/PROJECT.md` against the new section so
              mono-harness does not ship a template its own identity file no
              longer follows: run from the repository root; `tests:`
              `cd cli && go test ./...`; `types:` `n/a` (the Go compiler is the
              type gate); `lint:` `cd cli && go vet ./...`; `build:`
              `cd cli && go build ./...`; `format:` `gofmt -l cli`. The
              existing `bash scripts/verify-doc-links.sh` gate and the phase
              gates stay recorded.
            proof: `sh -c 'f=docs/PROJECT.md; for k in "run from:" "tests:" "types:" "lint:" "build:" "format:"; do grep -q "$k" "$f" || exit 1; done'`
      - wave: W2 — the conflict is proven, not assumed
        tasks:
          - task: T3.3 — add
              `TestUpdate_IdentityTemplateGateSlots_ConflictsAndAborts` to
              `cli/internal/installer/installer_test.go`, extending the shape of
              `TestUpdate_FastForward_Kept_AutoMerge_ConflictAbort`. Capture the
              real shipped template bytes by calling
              `srcBytesImpl(Target{Src: projectTemplate})` before any
              `withSource` call, because two `withSource` calls layer and the
              first override becomes `prev`. Install against the pre-edit
              template literal as the recorded base, write a filled fixture over
              the target, `withSource` the captured post-edit bytes as upstream,
              then assert: `RunUpdate` returns non-nil, the target contains
              `conflictOpenTag`, `.zharness/conflicts.json` names the file, and
              `RunUpdate` with `Abort: true` restores the filled fixture byte
              for byte. Never a manual scratch-repo run — the pre-commit guard
              re-executes proofs from a bare checkout.
            proof: `cd cli && go test ./internal/installer/ -run TestUpdate_IdentityTemplateGateSlots_ConflictsAndAborts -count=1`
    expected_outputs: modified
      `cli/docs/embedded/templates/project.identity.md`, modified
      `docs/PROJECT.md`, one new test in
      `cli/internal/installer/installer_test.go`.
    checks:
      - `cd cli && go build ./... && go vet ./... && go test ./...`
      - `bash scripts/verify-doc-links.sh`
      - `check full`

  - phase_slug: git-skill-multistack
    story_id: 01M0MULTISTACKHRNS4P9K4XJ2
    status: checked
    goal: the `git` skill stages explicit paths instead of contradicting its own
      Anti-Patterns section, and classifies Go, Rust, and Python dependency and
      test files correctly. Satisfies R10, R11, R12.
    depends_on: none
    allowed_surfaces: `skills/workflow/git/references/workflow.md` only. This
      file is not projected — its own header states
      `cli/docs/embedded/playbooks/` holds no `git` entry — so R13 does not
      apply to it.
    avoided_surfaces: the `pr` and `merge` sections' `main` defaults at lines
      22-23 and 78, which no requirement covers; `skills/workflow/git/SKILL.md`;
      every playbook.
    waves:
      - wave: W1 — one file, three independent edits
        tasks:
          - task: T4.1 — rewrite Step 1 to review `git status --porcelain` and
              stage explicit paths, removing `git add -A`, which that same
              file's Anti-Patterns section already forbids for catching `.env`,
              secrets, and `node_modules`.
            proof: `sh -c '! awk "/^### Step 1/,/^### Step 2/" skills/workflow/git/references/workflow.md | grep -q "git add -A"'`
          - task: T4.2 — extend the `deps:` class in Step 3 to match
              `go.mod`/`go.sum`, `Cargo.toml`/`Cargo.lock`, and
              `pyproject.toml`/`requirements*.txt`/`poetry.lock`/`uv.lock`
              beside `package.json` and its lockfiles, and state that a
              manifest commits together with its lockfile.
            proof: `sh -c 'f=skills/workflow/git/references/workflow.md; for k in go.mod go.sum Cargo.toml Cargo.lock pyproject.toml poetry.lock uv.lock; do grep -q "$k" "$f" || exit 1; done'`
          - task: T4.3 — state in the `test:`/`code:` split that a changed
              `*_test.go` commits with the source file it covers when both
              changed, because Go places `foo_test.go` beside `foo.go` and the
              current split yields two commits, neither of which builds and
              passes alone.
            proof: `sh -c 'awk "/^### Step 3/,/^### Step 4/" skills/workflow/git/references/workflow.md | grep -q "_test.go"'`
    expected_outputs: modified `skills/workflow/git/references/workflow.md`.
    checks:
      - `bash scripts/validate-skill.sh skills/workflow/git/SKILL.md`
      - `bash scripts/verify-doc-links.sh`
      - `check full`

## Progress
<!-- Append-only durable entries record timestamp, phase, wave, task, task_status, exact verification/result, and changed surfaces or blocker. -->
- none
- timestamp: 2026-09-06T14:12:16Z | phase: projection-parity-gate | wave: W1 | task: T1.1 | task_status: in-progress | verification: run started | surfaces: cli/docs/embedded/parity_test.go
- timestamp: 2026-09-06T14:12:16Z | phase: projection-parity-gate | wave: W1 | task: T1.1 | task_status: DONE | verification: `cd cli && go test ./docs/embedded/ -run TestProjectionParity -count=1` rc=0 (TestProjectionParity, TestProjectionParity_DetectsDrift both PASS) | surfaces: cli/docs/embedded/parity_test.go (new)
- timestamp: 2026-09-06T14:12:16Z | phase: projection-parity-gate | wave: W1 | summary: wave complete — parity test added, drift-detection subtest passes
- timestamp: 2026-09-06T14:12:16Z | phase: projection-parity-gate | wave: W2 | task: T1.2 | task_status: DONE | verification: `sh -c 'test "$(grep -c "docs/playbooks/\*\*" .github/workflows/cli-ci.yml)" = 2 && test "$(grep -c "docs/WORKFLOW.md" .github/workflows/cli-ci.yml)" = 2'` rc=0 | surfaces: .github/workflows/cli-ci.yml
- timestamp: 2026-09-06T14:12:16Z | phase: projection-parity-gate | wave: W2 | task: T1.3 | task_status: DONE | verification: `git ls-files --error-unmatch docs/audit/multi-stack-agent-readiness-audit.md docs/audit/multi-stack-harness-audit.md` rc=0 (both staged) | surfaces: docs/audit/multi-stack-agent-readiness-audit.md, docs/audit/multi-stack-harness-audit.md (git add, no content change)
- timestamp: 2026-09-06T14:12:16Z | phase: projection-parity-gate | wave: W2 | summary: wave complete — CI path filters extended, both audit docs tracked
- timestamp: 2026-09-06T14:26:34Z | phase: playbook-truth | wave: W1 | task: T2.1 | task_status: in-progress | verification: run started | surfaces: cli/docs/embedded/playbooks/watzup.md
- timestamp: 2026-09-06T14:26:34Z | phase: playbook-truth | wave: W1 | task: T2.1 | task_status: DONE | verification: content half rc=0 (`origin/HEAD`, `show-ref --verify`, step-1 heading present, no `main..HEAD` in step 1) and synthetic-repo half rc=0 | surfaces: cli/docs/embedded/playbooks/watzup.md
- timestamp: 2026-09-06T14:26:34Z | phase: playbook-truth | wave: W1 | task: T2.2 | task_status: DONE | verification: `sh -c 'grep -q "branch --show-current" cli/docs/embedded/playbooks/watzup.md'` rc=0 and throwaway no-remote/no-commit repo resolution rc=0 | surfaces: cli/docs/embedded/playbooks/watzup.md
- timestamp: 2026-09-06T14:26:34Z | phase: playbook-truth | wave: W1 | task: T2.3 | task_status: DONE | verification: `sh -c 'f=cli/docs/embedded/playbooks/work.md; grep -q "plan-slice.sh" "$f" && grep -q "absent" "$f" && grep -q "awk" "$f"'` rc=0 | surfaces: cli/docs/embedded/playbooks/work.md
- timestamp: 2026-09-06T14:26:34Z | phase: playbook-truth | wave: W1 | summary: wave complete — base-branch resolution and plan-slice absence branch authored
- timestamp: 2026-09-06T14:26:34Z | phase: playbook-truth | wave: W2 | task: T2.4 | task_status: DONE_WITH_CONCERNS | verification: `sh -c 'test "$(grep -c "install-git-hooks.sh --force" cli/docs/embedded/playbooks/check.md)" -ge 2'` rc=0, after first wording broke `TestOnePlan_PlaybookContract` (pinned phrase) and was reworded — see Decisions | surfaces: cli/docs/embedded/playbooks/check.md
- timestamp: 2026-09-06T14:26:34Z | phase: playbook-truth | wave: W2 | task: T2.5 | task_status: DONE | verification: `sh -c 'f=cli/docs/embedded/playbooks/check.md; grep -q "repository root" "$f" && grep -q "no inherited shell state" "$f" && grep -q "300" "$f"'` rc=0 | surfaces: cli/docs/embedded/playbooks/check.md
- timestamp: 2026-09-06T14:26:34Z | phase: playbook-truth | wave: W2 | task: T2.6 | task_status: DONE | verification: `sh -c 'test "$(grep -c enforcement cli/docs/embedded/playbooks/check.md)" -ge 2'` rc=0 (count=3); `bash scripts/test-guards.sh` 25 passed, 0 failed — the new receipt key did not break the guard parser | surfaces: cli/docs/embedded/playbooks/check.md
- timestamp: 2026-09-06T14:26:34Z | phase: playbook-truth | wave: W2 | summary: wave complete — enforcement claims qualified, re-execution contract stated, `enforcement:` receipt key added
- timestamp: 2026-09-06T14:26:34Z | phase: playbook-truth | wave: W3 | task: T2.7 | task_status: DONE | verification: `cd cli && go test ./docs/embedded/ -run TestProjectionParity -count=1` rc=0 after copying all three files to `docs/playbooks/` | surfaces: docs/playbooks/{watzup,work,check}.md
- timestamp: 2026-09-06T14:26:34Z | phase: playbook-truth | wave: W3 | summary: wave complete — projection byte-identical, Phase 1 gate passes on it
- timestamp: 2026-09-06T14:32:43Z | phase: identity-gate-slots | wave: W1 | task: T3.1 | task_status: in-progress | verification: run started; pre-edit template bytes captured via `git show HEAD:cli/docs/embedded/templates/project.identity.md` | surfaces: cli/docs/embedded/templates/project.identity.md
- timestamp: 2026-09-06T14:32:43Z | phase: identity-gate-slots | wave: W1 | task: T3.1 | task_status: DONE | verification: gate-slot proof rc=0 (all six slots + `n/a` present, `## How do we run the tests?` gone, 27 lines <= 50) | surfaces: cli/docs/embedded/templates/project.identity.md
- timestamp: 2026-09-06T14:32:43Z | phase: identity-gate-slots | wave: W1 | task: T3.2 | task_status: DONE | verification: `sh -c 'f=docs/PROJECT.md; for k in "run from:" "tests:" "types:" "lint:" "build:" "format:"; do grep -q "$k" "$f" || exit 1; done'` rc=0; file is 48 lines, under the 50-line ceiling | surfaces: docs/PROJECT.md
- timestamp: 2026-09-06T14:32:43Z | phase: identity-gate-slots | wave: W1 | summary: wave complete — template asks one question per gate class, this repository answers the new shape
- timestamp: 2026-09-06T14:32:43Z | phase: identity-gate-slots | wave: W2 | task: T3.3 | task_status: DONE | verification: `cd cli && go test ./internal/installer/ -run TestUpdate_IdentityTemplateGateSlots_ConflictsAndAborts -count=1` rc=0; additionally proved the test has teeth with a throwaway scratch test showing the appended-section variant auto-merges (`err = <nil>`, no conflict markers), then removed it | surfaces: cli/internal/installer/installer_test.go
- timestamp: 2026-09-06T14:32:43Z | phase: identity-gate-slots | wave: W2 | summary: wave complete — the update conflict is proven executable, not assumed
- timestamp: 2026-09-06T14:38:25Z | phase: git-skill-multistack | wave: W1 | task: T4.1 | task_status: in-progress | verification: run started | surfaces: skills/workflow/git/references/workflow.md
- timestamp: 2026-09-06T14:38:25Z | phase: git-skill-multistack | wave: W1 | task: T4.1 | task_status: DONE | verification: `sh -c '! awk "/^### Step 1/,/^### Step 2/" skills/workflow/git/references/workflow.md | grep -q "git add -A"'` rc=0; `git add -A` now appears only at line 145, where Anti-Patterns forbids it | surfaces: skills/workflow/git/references/workflow.md
- timestamp: 2026-09-06T14:38:25Z | phase: git-skill-multistack | wave: W1 | task: T4.2 | task_status: DONE | verification: `sh -c 'f=skills/workflow/git/references/workflow.md; for k in go.mod go.sum Cargo.toml Cargo.lock pyproject.toml poetry.lock uv.lock; do grep -q "$k" "$f" || exit 1; done'` rc=0 | surfaces: skills/workflow/git/references/workflow.md
- timestamp: 2026-09-06T14:38:25Z | phase: git-skill-multistack | wave: W1 | task: T4.3 | task_status: DONE | verification: `sh -c 'awk "/^### Step 3/,/^### Step 4/" skills/workflow/git/references/workflow.md | grep -q "_test.go"'` rc=0 | surfaces: skills/workflow/git/references/workflow.md
- timestamp: 2026-09-06T14:38:25Z | phase: git-skill-multistack | wave: W1 | summary: wave complete — the skill stages explicit paths and classifies Go/Rust/Python deps and co-located tests correctly
- timestamp: 2026-09-06T16:02:00Z | phase: playbook-truth | wave: R1 (remediation) | task: M1/J1 — verify the `origin/HEAD`-derived base before use | task_status: DONE | verification: `bash regress-j1.sh` in a throwaway clone with `origin/HEAD -> origin/main` and no local `main` — old snippet `base=main -> git log EXIT 128`, patched snippet `base=feat -> git log OK`; T2.1 and T2.2 proofs still pass | surfaces: cli/docs/embedded/playbooks/watzup.md
- timestamp: 2026-09-06T16:06:00Z | phase: playbook-truth | wave: R2 (remediation) | task: M2/J2 — qualify the `CI re-runs it` clause | task_status: DONE | verification: `sh -c '! grep -qF "install-git-hooks.sh\`; CI re-runs it" cli/docs/embedded/playbooks/check.md'` -> pass; T2.4/T2.5/T2.6 still pass; the `TestOnePlan_PlaybookContract` pinned phrase "The repository's pre-commit hook is the sole proof guarantee" verified intact | surfaces: cli/docs/embedded/playbooks/check.md
- timestamp: 2026-09-06T16:08:00Z | phase: playbook-truth | wave: R2 (remediation) | task: J3 — align step 7's ladder level to step 9's | task_status: DONE | verification: `grep -c "Local validation" cli/docs/embedded/playbooks/check.md` -> 0; both steps now name `Optional hook` | surfaces: cli/docs/embedded/playbooks/check.md
- timestamp: 2026-09-06T16:10:00Z | phase: playbook-truth | wave: R2 (remediation) | task: M3b — state the guard-visible entry shape in the step-8 field contract | task_status: DONE | verification: contract now specifies the verdict on the entry's first line and the bare single-backtick proof sub-bullet, and records that append-only binds from an entry's first commit | surfaces: cli/docs/embedded/playbooks/check.md
- timestamp: 2026-09-06T16:12:00Z | phase: playbook-truth | wave: R3 (remediation) | task: J5 + J6 — parameterize the absence-branch heading, drop the internal Go symbol | task_status: DONE | verification: `{heading}` now appears in both branches of step 1 (count 2); `grep -c "AllTargets()" cli/docs/embedded/playbooks/work.md` -> 0; T2.3 still passes | surfaces: cli/docs/embedded/playbooks/work.md
- timestamp: 2026-09-06T16:15:00Z | phase: playbook-truth | wave: R4 (remediation) | task: T2.7 — project the three edited playbooks (R13) | task_status: DONE | verification: `cmp` byte-identical for watzup/work/check; `cd cli && go test ./docs/embedded/ -run TestProjectionParity -count=1` -> ok | surfaces: docs/playbooks/{watzup,work,check}.md
- timestamp: 2026-09-06T16:18:00Z | phase: playbook-truth | wave: R4 (remediation) | task: M3 — reshape `## Validation` entries into the guard-visible shape | task_status: DONE | verification: 5 first-line verdicts added, 32 proof bullets converted from the labelled to the bare form; the M3 regression proof flipped from `entries=5 first-line-verdicts=0` (exit 1) to `entries=7 first-line-verdicts=7` (exit 0); a real `zharness_guard_entries_of_file <path> <old> <new>` run re-executed **24** proofs with **0** rejections, against **0** re-executed before the reshape | surfaces: docs/plans/active/multi-stack-harness-readiness.md
- timestamp: 2026-09-06T16:19:00Z | phase: playbook-truth | wave: R1-R4 (remediation) summary | task: — | task_status: DONE | verification: every phase check passes — `cd cli && go test ./... -count=1` ok (5 pkgs), `bash scripts/test-guards.sh` `guards: 25 passed, 0 failed`, `bash scripts/verify-doc-links.sh` `0 findings`, `gofmt -l cli` empty, `go vet ./...` clean | surfaces: cli/docs/embedded/playbooks/{watzup,work,check}.md + projections + the plan's Validation section
- timestamp: 2026-09-06T16:40:00Z | phase: playbook-truth | wave: R5 | task: T2.5 | task_status: done | verification: `cd cli && go test ./... -count=1` pass; `cmp cli/docs/embedded/playbooks/check.md docs/playbooks/check.md` byte-identical; guard simulation 46 re-executed / 0 rejections / exit 0 (first run rejected this entry's own negative grep, which exits 1 when the string is correctly absent; rewritten as `! grep -q ...` so absence proves as exit 0) | surfaces: cli/docs/embedded/playbooks/check.md, docs/playbooks/check.md

## Decisions
<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->
- timestamp: 2026-09-06T14:26:34Z | phase/task: playbook-truth / T2.4 | decision: keep the exact sentence "The repository's pre-commit hook is the sole proof guarantee" intact and append the qualifier "in a repository that installed it (`bash scripts/install-git-hooks.sh --force`)" to it, rather than rewriting the sentence as first drafted. | rationale: `cli/internal/embedded/embedded_test.go:252` pins that phrase as a required one-plan contract string via a case-sensitive `strings.Contains`, and the first wording deleted it, failing `TestOnePlan_PlaybookContract`. `cli/internal/` is in this phase's avoided_surfaces, so editing the test was out of authority. The rewording satisfies R5's honesty requirement and the pinned contract simultaneously with no avoided-surface edit and no escalation.
- timestamp: 2026-09-06T14:32:43Z | phase/task: identity-gate-slots / T3.2 | decision: also refresh `docs/PROJECT.md`'s "What are we working on right now?" line from `(active, not-planned)` to `(active, in-progress)` while re-answering the gate section. | rationale: `docs/PROJECT.md` is an allowed surface for this phase and the status word had been factually stale since to-plan minted the phases; leaving a known-false line in the file this phase exists to make truthful would contradict the phase goal. One word, same file, no other surface touched.
- timestamp: 2026-09-06T15:45:00Z | phase/task: playbook-truth / M3 | decision: owner ruled that the M3 fix reshapes all existing `## Validation` entries in place — moving each verdict token onto the entry's first line and rewriting proof bullets to the guard-matchable shape — rather than only shaping entries authored from now on. | rationale: check.md:28 declares Validation append-only, but nothing in this plan is committed, the file is still untracked, and `zharness_guard_entries_of_file` computes old-side entry hashes from the committed version, so there is no prior commit for append-only to protect and no hash the reshape can invalidate. Leaving the five existing entries unreshaped would permanently hide the four APPROVED gate entries from the proof guard, which is the defect M3 names. Owner chose reshape over the literal reading; the append-only rule is honored from first commit onward.
- timestamp: 2026-09-06T15:45:00Z | phase/task: git-skill-multistack / A1-X1 | decision: owner ruled the A1-X1 base-branch residual IN SCOPE as a new requirement R15 rather than an explicit non-goal or a separate follow-up plan. R15: mirror the corrected `watzup` base resolution (`origin/HEAD` → verified local candidate → current branch) into `skills/workflow/git/references/workflow.md:88`, fix `branch-management.md:33`, and restate the `pr`/`merge` defaults at `:22-23` as "the repository's default branch" rather than `main`. The `git-skill-multistack` phase reopens from `checked` to `in-progress` to carry it. | rationale: the independent judge confirmed the class is INCOMPLETE and reproduced the failure in this repository (`origin/main` absent; `git log origin/main...origin/master` exits 128), so `/git pr` is broken here today. It also identified the mechanical cause the residual passed its own gate: phase 4's proof at line 482 is a negative grep bounded to Step 1 of `workflow.md`, blind to `:88` and `branch-management.md:33`. R15's verification must therefore scan the whole file, not Step 1. Closing the plan with a known-broken shipped skill was rejected.
- timestamp: 2026-09-06T16:07:00Z | phase/task: playbook-truth / M2 | decision: qualify only the trailing clause ``(`scripts/install-git-hooks.sh`; CI re-runs it)`` and leave the `bash scripts/install-git-hooks.sh --force` referral in place, rather than adding a script-absence branch as first framed. | rationale: R4 (line 54) explicitly requires check.md to point at that command as the opt-in, so an absence branch for the referral would contradict the requirement it was meant to satisfy. The independent judge self-corrected to the same narrower reading. The surviving falsehood was only "CI re-runs it", which is true for mono-harness (`.github/workflows/cli-ci.yml:56,61` extracts the `ZGUARD-CORE` block) and false for every consumer, since `AllTargets()` ships no workflow — so the clause is now scoped to a repository that checks in such a workflow.
- timestamp: 2026-09-06T16:11:00Z | phase/task: playbook-truth / M3b | decision: encode the owner's append-only ruling directly in the step-8 field contract — append-only binds from an entry's first commit, and reshaping an uncommitted entry into the guard-visible shape is permitted. | rationale: the ruling resolves a real contradiction between check.md's append-only wording and the M3 fix, and leaving it only in this plan's Decisions would let the next agent re-litigate it. The mechanism is stated alongside it so the rule is checkable rather than asserted: the guard derives old-side entry hashes from the committed version, so an uncommitted entry has no hash the reshape can invalidate. Same file and same commit as the M2 edit, per R13.
- timestamp: 2026-09-06T16:13:00Z | phase/task: playbook-truth / J4 | decision: DEFER J4 — do not widen the `enforcement: hook | ci | local-only` vocabulary to cover "script present, hook not installed". | rationale: that vocabulary is R6's accepted requirement text, and `work.md`'s escalate_when stops execution when locked requirements would change. Widening it is an owner decision for `brainstorm`/`to-plan`, not a work-phase edit. Recorded as an open item instead; the ladder in `docs/patterns/encoding-invariants.md` has four levels against the key's three, so the gap is real but out of this phase's authority.
- timestamp: 2026-09-06T16:14:00Z | phase/task: playbook-truth / J7 | decision: DEFER J7 — leave `docs/PROJECT.md`'s `also required:` line as it is. | rationale: `docs/PROJECT.md` and `cli/docs/embedded/templates/` are this phase's avoided_surfaces; they belong to `identity-gate-slots`, which is already `checked`. Touching either here would be contract drift. Recorded as an open item for whoever reopens that phase.
- timestamp: 2026-09-06T16:17:00Z | phase/task: playbook-truth / M3 | decision: reshape the passing proof bullets to the bare single-backtick form but leave the deliberately-failing regression proofs in their labelled form. | rationale: the guard re-executes every matched proof bullet under an APPROVED verdict. A regression proof exists to fail while a defect is present; making it guard-matchable would arm a command designed to exit non-zero if that entry's verdict were ever read as clean. The labelled form keeps them as readable evidence while structurally unable to be re-executed. Verified that the bare form tolerates trailing annotation, so no evidence text was lost in the conversion.
- timestamp: 2026-09-06T16:26:00Z | phase/task: playbook-truth / M4 | decision: rewrite the M2 gate proof to use a character class instead of a literal backtick, and record M4 as an open item rather than widening the guard's extraction pattern in this phase. | rationale: the first full guard run over the reshaped plan REJECTED that proof — the extraction pattern takes a proof bullet as a bare command between single backticks, so a command containing its own backtick is truncated at the inner one and re-executes as `sh: unexpected EOF while looking for matching`. The defect was found by the guard the same phase made effective, which is the intended failure mode working. The proof was rewritten as `install-git-hooks[.]sh.{0,3}; CI re-runs it` and verified to still discriminate: it passes on the corrected file and still matches the original false clause in a fixture. Changing the pattern itself lives in `scripts/`, this phase's avoided surface, so M4 is recorded for an owner decision rather than fixed here.
- timestamp: 2026-09-06T16:38:00Z | phase/task: playbook-truth / M4 | decision: name the no-inner-backtick constraint in the step-8 proof-shape contract, and correct the same sentence's false claim that a proof bullet is the command "and nothing else". | rationale: the contract as first written was itself a truth defect of the kind this phase exists to remove — `drive5.sh` and the gate entry's own 12 bullets prove trailing annotation after the closing backtick is ignored and safe, so an agent obeying the text literally would strip its own evidence. Naming M4's constraint here is the half of M4's fix that does not touch `scripts/`, this phase's avoided surface; the guard-side fix stays deferred. | alternatives: leave the wording (rejected: propagates a false shape rule through every consumer repo).
- timestamp: 2026-09-06T16:42:00Z | phase/task: playbook-truth / J8 | decision: set `playbook-truth` to `checked`, matching the three sibling phases. | rationale: `work.md` step 11 contradicts itself — it orders check.md's gate steps 1-4 and 6-11 performed in-session, and step 10 mandates `checked` on `APPROVED`, then the closing sentence forbids exactly that. Phases 1, 3 and 4 all reached `checked` through this same in-session gate, and this plan's own check.md follow-up item already reasons from "`work.md` step 11 has already set every phase to `checked`". Leaving phase 2 alone would make the plan internally inconsistent and would show `watzup` a clean gate with an open status. Which clause of step 11 wins is a spine-playbook truth defect for the owner, recorded as J8. | alternatives: leave `in-progress` and add a `check gate phase playbook-truth` hop (rejected: a whole stage invocation whose only effect is a status flip already earned).

## Validation
<!-- Append-only durable entries record timestamp, phase, exact command/result/output, verdict, judge, receipt, and proof_gaps. -->
- timestamp: 2026-09-06T14:20:00Z | phase: projection-parity-gate | mode: gate — verdict `APPROVED` — judge: same-session
  - scope: on target
  - depth: standard
  - `cd cli && go build ./... && go vet ./... && go test ./...` -> pass (all packages ok, TestProjectionParity + TestProjectionParity_DetectsDrift PASS)
  - `bash scripts/verify-doc-links.sh` -> pass (`doc links OK (0 findings; 10 claim(s) under known-removed v0.15 surfaces)`)
  - `cd cli && go test ./docs/embedded/ -run TestProjectionParity -count=1` -> pass
  - `sh -c 'test "$(grep -c "docs/playbooks/\*\*" .github/workflows/cli-ci.yml)" = 2 && test "$(grep -c "docs/WORKFLOW.md" .github/workflows/cli-ci.yml)" = 2'` -> pass
  - `git ls-files --error-unmatch docs/audit/multi-stack-agent-readiness-audit.md docs/audit/multi-stack-harness-audit.md` -> pass
  - plan alignment: R1, R2, R14 satisfied; expected_outputs match (new `cli/docs/embedded/parity_test.go`, modified `.github/workflows/cli-ci.yml`, both audit docs tracked); no avoided_surfaces touched; no Decisions required
  - verdict: APPROVED
  - judge: same-session
  - judge_model: claude-sonnet-5
  - proof_gaps: none
  - receipt: context_sources: docs/plans/active/multi-stack-harness-readiness.md, docs/audit/multi-stack-agent-readiness-audit.md, docs/audit/multi-stack-harness-audit.md, cli/docs/embedded/embed.go, .github/workflows/cli-ci.yml; policy: docs/playbooks/work.md#11, docs/playbooks/check.md#gate; judge: same-session; judge_model: claude-sonnet-5; retries: 1 (first parity_test.go draft swept the entire embedded FS and docs/ tree, over-matching AGENTS.md and non-playbook docs files; corrected to WORKFLOW.md + playbooks/*.md only per T1.1's spec, then reran clean); rollback_point: pre-phase commit aba7057; failure_ledger: absent; not_independently_verified: the CI path-filter trigger was verified by grep against `.github/workflows/cli-ci.yml`, not by observing an actual GitHub Actions push/PR run

- timestamp: 2026-09-06T14:30:00Z | phase: playbook-truth | mode: gate — verdict `APPROVED` — judge: same-session
  - scope: on target
  - depth: standard
  - `cd cli && go test ./docs/embedded/ -run TestProjectionParity -count=1` -> pass
  - `bash scripts/test-guards.sh` -> pass (`guards: 25 passed, 0 failed`)
  - `bash scripts/verify-doc-links.sh` -> pass (`doc links OK (0 findings; 10 claim(s) under known-removed v0.15 surfaces)`)
  - `cd cli && go build ./... && go vet ./... && go test ./...` -> pass (all packages ok, including `TestOnePlan_PlaybookContract`)
  - `sh -c 'f=cli/docs/embedded/playbooks/watzup.md; grep -q "origin/HEAD" "$f" && grep -q "show-ref --verify" "$f" && grep -q "^1\. \*\*Read branch state" "$f" && ! awk "/^1\. \*\*Read branch state/,/^2\./" "$f" | grep -q "main\.\.HEAD"'` -> pass
  - `sh -c 'f=cli/docs/embedded/playbooks/work.md; grep -q "plan-slice.sh" "$f" && grep -q "absent" "$f" && grep -q "awk" "$f"'` -> pass
  - `sh -c 'test "$(grep -c "install-git-hooks.sh --force" cli/docs/embedded/playbooks/check.md)" -ge 2'` -> pass
  - `sh -c 'f=cli/docs/embedded/playbooks/check.md; grep -q "repository root" "$f" && grep -q "no inherited shell state" "$f" && grep -q "300" "$f"'` -> pass
  - `sh -c 'test "$(grep -c enforcement cli/docs/embedded/playbooks/check.md)" -ge 2'` -> pass
  - plan alignment: R3, R4, R5, R6, R9 satisfied under R13 (all three edits projected byte-identically in this same change); expected_outputs match; no avoided_surfaces touched — `work.md:17`'s bounded ceiling, `scripts/`, `cli/internal/`, `cli/docs/embedded/templates/`, and `skills/workflow/git/` are unmodified; one Decision recorded for the T2.4 contract-phrase conflict
  - verdict: APPROVED
  - judge: same-session
  - judge_model: claude-sonnet-5
  - proof_gaps: none
  - receipt: context_sources: docs/plans/active/multi-stack-harness-readiness.md, cli/docs/embedded/playbooks/{watzup,work,check}.md, cli/internal/embedded/embedded_test.go, docs/patterns/encoding-invariants.md; policy: docs/playbooks/work.md#11, docs/playbooks/check.md#gate; judge: same-session; judge_model: claude-sonnet-5; retries: 1 (T2.4's first wording deleted the phrase pinned by `TestOnePlan_PlaybookContract`; reworded to keep it verbatim and reran the full Go suite clean); rollback_point: pre-phase commit aba7057; failure_ledger: absent; enforcement: local-only (the pre-commit hook is not installed in this checkout — `bash scripts/install-git-hooks.sh --force` is the opt-in, and CI re-runs the guards); not_independently_verified: the rewritten `watzup.md` base-branch resolution was proven by grep and by synthetic throwaway repositories, not by running the recap end-to-end against a repository whose `origin/HEAD` points at a non-`master` default
- timestamp: 2026-09-06T14:40:00Z | phase: identity-gate-slots | mode: gate — verdict `APPROVED` — judge: same-session
  - scope: on target
  - depth: standard
  - `cd cli && go test ./internal/installer/ -run TestUpdate_IdentityTemplateGateSlots_ConflictsAndAborts -count=1` -> pass
  - `cd cli && go build ./... && go vet ./... && go test ./...` -> pass (all packages ok)
  - `bash scripts/verify-doc-links.sh` -> pass (`doc links OK (0 findings; 10 claim(s) under known-removed v0.15 surfaces)`)
  - `sh -c 'f=cli/docs/embedded/templates/project.identity.md; for k in "run from:" "tests:" "types:" "lint:" "build:" "format:" "n/a"; do grep -q "$k" "$f" || exit 1; done; ! grep -q "^## How do we run the tests?" "$f"; test "$(wc -l < "$f")" -le 50'` -> pass
  - `sh -c 'f=docs/PROJECT.md; for k in "run from:" "tests:" "types:" "lint:" "build:" "format:"; do grep -q "$k" "$f" || exit 1; done'` -> pass
  - plan alignment: R7 and R8 satisfied; expected_outputs match (template rewritten in place, `docs/PROJECT.md` re-answered, one new installer test); NG4 held — `threeway.go`, `installer.go`, and every other non-test file under `cli/internal/` are unmodified; one Decision recorded for the PROJECT.md status-word refresh
  - verdict: APPROVED
  - judge: same-session
  - judge_model: claude-sonnet-5
  - proof_gaps: none
  - receipt: context_sources: docs/plans/active/multi-stack-harness-readiness.md, cli/docs/embedded/templates/project.identity.md, docs/PROJECT.md, cli/internal/installer/installer_test.go, cli/internal/installer/installer.go; policy: docs/playbooks/work.md#11, docs/playbooks/check.md#gate; judge: same-session; judge_model: claude-sonnet-5; retries: 0; rollback_point: pre-phase commit aba7057; failure_ledger: absent; enforcement: local-only (the pre-commit hook is not installed in this checkout — `bash scripts/install-git-hooks.sh --force` is the opt-in, and CI re-runs the guards); not_independently_verified: the conflict was proven through the installer's own update path in a temp repository, not by running the released `zharness update` binary against a real consumer repository that had installed the pre-edit template
- timestamp: 2026-09-06T14:45:00Z | phase: git-skill-multistack | mode: gate — verdict `APPROVED` — judge: same-session
  - scope: on target
  - depth: quick
  - `bash scripts/validate-skill.sh skills/workflow/git/SKILL.md` -> pass (`PASS WITH WARNINGS - 6 warning(s)`; all six are pre-existing `SKILL.md` structure notes — missing `version` field, no `<instructions>` XML section, no examples, no output-format spec — and none concern `references/workflow.md`, the only file this phase changed)
  - `bash scripts/verify-doc-links.sh` -> pass (`doc links OK (0 findings; 10 claim(s) under known-removed v0.15 surfaces)`)
  - `sh -c '! awk "/^### Step 1/,/^### Step 2/" skills/workflow/git/references/workflow.md | grep -q "git add -A"'` -> pass
  - `sh -c 'f=skills/workflow/git/references/workflow.md; for k in go.mod go.sum Cargo.toml Cargo.lock pyproject.toml poetry.lock uv.lock; do grep -q "$k" "$f" || exit 1; done'` -> pass
  - `sh -c 'awk "/^### Step 3/,/^### Step 4/" skills/workflow/git/references/workflow.md | grep -q "_test.go"'` -> pass
  - plan alignment: R10, R11, R12 satisfied; expected_outputs match (only `skills/workflow/git/references/workflow.md` modified); avoided_surfaces held — the `pr`/`merge` `main` defaults at lines 22-23 and 78 are untouched, as are `skills/workflow/git/SKILL.md` and every playbook; no Decisions required
  - verdict: APPROVED
  - judge: same-session
  - judge_model: claude-sonnet-5
  - proof_gaps: the initiative-level `check full` (complete Security, Performance, Architecture, Code Quality review with `judge: independent`) has not run; this is the final phase, so `handoff.md` step 6 requires it exactly once before closure
  - receipt: context_sources: docs/plans/active/multi-stack-harness-readiness.md, skills/workflow/git/references/workflow.md, docs/audit/multi-stack-agent-readiness-audit.md; policy: docs/playbooks/work.md#11, docs/playbooks/check.md#gate; judge: same-session; judge_model: claude-sonnet-5; retries: 0; rollback_point: pre-phase commit aba7057; failure_ledger: absent; enforcement: local-only (the pre-commit hook is not installed in this checkout — `bash scripts/install-git-hooks.sh --force` is the opt-in, and CI re-runs the guards); not_independently_verified: the rewritten Step 1 and Step 3 guidance was proven by grep against the file, not by running the `git` skill end-to-end on a Go, Rust, or Python repository to observe the resulting commit split

- timestamp: 2026-09-06T15:00:00Z | phase: git-skill-multistack | mode: full — verdict `REQUEST_CHANGES` — judge: independent
  - scope: on target
  - depth: deep
  - note: initiative-level `full` run as the `handoff.md` step 6 closure precondition. It reviews all four phases, not only the selected one; the two blocking findings below land in phase `playbook-truth`, which is therefore reopened to `in-progress` by step 10.
  - note: check.md precondition 2 and step 1 require the selected phase read `in-progress`, but every phase read `checked` because `work.md` step 11 already gated them. That is structurally unavoidable for an initiative-level `full` at handoff closure. Proceeded past the literal precondition rather than reopening a phase backwards to satisfy it; recorded as a check.md follow-up in open_items.
  - `cd cli && go test ./... -count=1` -> pass (uncached; docs/embedded, internal/embedded, internal/installer, internal/interfaces all ok; cmd/zharness has no test files)
  - `cd cli && go test ./docs/embedded/ -run TestProjectionParity -count=1` -> pass (TestProjectionParity and TestProjectionParity_DetectsDrift both PASS)
  - `cd cli && go test ./internal/installer/ -run TestUpdate_IdentityTemplateGateSlots_ConflictsAndAborts -count=1` -> pass
  - `bash scripts/verify-doc-links.sh` -> pass (`doc links OK (0 findings; 10 claim(s) under known-removed v0.15 surfaces)`)
  - `bash scripts/test-guards.sh` -> pass (`guards: 25 passed, 0 failed`)
  - `gofmt -l cli` -> pass (empty output)
  - `sh -c 'for f in WORKFLOW.md playbooks/check.md playbooks/watzup.md playbooks/work.md; do diff -q "cli/docs/embedded/$f" "docs/${f}" >/dev/null || exit 1; done'` -> pass (R13 projection parity holds for every file this initiative edited)
  - `sh -c 'cd cli && go list ./... | grep -q "cli/docs/embedded"'` -> pass (the parity test is inside the package set CI's `go test ./...` executes, so a docs-only drift is actually caught)
  - proof (FAILS ON PURPOSE — M1 regression): `sh -c 'd=$(mktemp -d); git init -q "$d/up" >/dev/null 2>&1; cd "$d/up" || exit 9; git config user.email t@t; git config user.name t; git symbolic-ref HEAD refs/heads/main; echo a>a; git add a; git commit -qm i >/dev/null; cd "$d" || exit 9; git clone -q up cl >/dev/null 2>&1; cd cl || exit 9; git config user.email t@t; git config user.name t; git checkout -q -b feat; git branch -D main -q >/dev/null 2>&1; base=$(git symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null | sed "s|^origin/||"); [ -n "$base" ] || for c in main master; do git show-ref --verify --quiet "refs/heads/$c" && base=$c && break; done; [ -n "$base" ] || base=$(git branch --show-current); git log --oneline "$base..HEAD" >/dev/null 2>&1'` -> FAIL (exit 128) — reproduces M1; passes once the resolved base is verified
  - proof (FAILS ON PURPOSE — M2 regression): ``sh -c '! grep -qF "install-git-hooks.sh`; CI re-runs it" cli/docs/embedded/playbooks/check.md'`` -> FAIL (exit 1) — the false consumer-repo enforcement claim is still present; passes once reworded
  - finding M1 (MAJOR, architecture) — `cli/docs/embedded/playbooks/watzup.md:17` and its projection `docs/playbooks/watzup.md:17`: the `origin/HEAD`-derived base is the one candidate never verified. `sed 's|^origin/||'` turns `origin/main` into the bare name `main` and hands it straight to `git log`/`git rev-list`; `git show-ref --verify` guards only the `main`/`master` fallback candidates. R3 requires a *verified* base. Failure scenario reproduced in a real throwaway clone: `origin/HEAD -> origin/main` with no local `main` (the standard shape of a CI checkout that fetched only the PR ref, or any clone where the default branch was never checked out) resolves `base=main`, then `git log --oneline "main..HEAD"` and `git rev-list --left-right --count "main...HEAD"` both exit 128 with `fatal: ambiguous argument`. This breaks step 1 of the session-start playbook — the first command an agent runs. Fix: drop the strip (`origin/main..HEAD` resolves fine), or gate it with `git show-ref --verify --quiet "refs/heads/$base" || base=""` before the fallback loop.
  - finding M2 (MAJOR, correctness/honesty) — `cli/docs/embedded/playbooks/check.md:42` (step 9) and `:40` (step 7), plus both projections: the phase named `playbook-truth` left a false enforcement claim inside the sentence it edited. Step 9 asserts the hook is the guarantee, tells the reader to run `bash scripts/install-git-hooks.sh --force`, and states ``(`scripts/install-git-hooks.sh`; CI re-runs it)``. Verified against `cli/internal/installer/installer.go:66-76` (`AllTargets()`): the installer distributes only `WORKFLOW.md`, `templates/project.identity.md` -> `docs/PROJECT.md`, and the playbooks — no `scripts/`, no `.github/`; `grep -n 'scripts/\|\.github\|install-git-hooks' cli/internal/installer/installer.go` returns nothing and `ls cli/docs/embedded/` shows no scripts directory. Failure scenario: a consumer agent that ran `zharness install` reaches step 9, follows the instruction, gets `No such file or directory`, and has no fallback branch — and the cited CI does not exist in that repo either. This is the exact class of false enforcement claim R4 exists to remove, reintroduced by R4's own fix. Fix must land in `cli/docs/embedded/playbooks/check.md` and be projected to `docs/playbooks/check.md` in the same commit (R13).
  - finding F1 (minor, code quality) — `cli/docs/embedded/parity_test.go:48` builds an `fs.FS` path with `filepath.Join`, which emits `\` on Windows and is invalid for `fs.FS` (forward-slash only). CI is `ubuntu-latest` with no Windows matrix, so this is latent, not live. Correct fix is `path.Join`. Non-blocking.
  - finding D1 (minor, record accuracy) — the phase-4 `git-skill-multistack` entry above cites the `BASE=${TO_BRANCH:-main}` line as **78**; the real line is **88** (78 is a code fence). Validation is append-only, so the citation is corrected here rather than rewritten there.
  - finding D2 (minor, record conformance) — the phase-1 `projection-parity-gate` receipt omits the `enforcement:` field that R6 requires. Chronologically defensible, since R6's field was introduced by the phase-2 edit itself, but the entry is non-conforming as written.
  - sibling-instance coverage (check.md step 5): **Node-only assumption class — COMPLETE.** `grep -rn 'package\.json\|npm \|pnpm \|yarn \|node_modules'` over `docs/playbooks/`, `cli/docs/embedded/`, and `docs/WORKFLOW.md` returns zero hits. **Hardcoded base-branch (A1-X1) class — INCOMPLETE.** Residual instances: `skills/workflow/git/references/workflow.md:22` and `:23` (documented `pr`/`merge` defaults of `main`), `:88` (`BASE=${TO_BRANCH:-main}`, consumed unquoted at `:90`), and `skills/workflow/git/references/branch-management.md:33` (`git rebase origin/main`). Confirmed to fail in this repository today: `origin/HEAD` is `origin/master`, `origin/main` is absent, so `git log origin/main...origin/master` exits with `fatal: ambiguous argument`. Not covered by R3 (scoped to `watzup` step 1), not by R10-R12 (staging, `deps:` breadth, `*_test.go` split), and not excluded by any of NG1-NG7 — so it is an unrecorded open item, not a declared non-goal. It does not by itself force `REQUEST_CHANGES` (step 7 reserves that for critical issues and material plan contradictions), but the watzup fix must not be recorded as closing the A1-X1 class.
  - security review: no findings. Every new shell expansion is quoted (`"$base..HEAD"`, `"$base...HEAD"`, `"refs/heads/$c"`); the loop candidate comes from a literal `main`/`master` list; `sed 's|^origin/||'` takes the branch name as input, not as script. No secrets introduced in the diff. `git add -A` now survives in `skills/` only at `workflow.md:145`, where it is the prohibition.
  - performance review: no findings. The parity test is a fixed byte comparison over a handful of small files (0.00s). Widening the CI path filters means a docs-only edit runs the Go build/vet/test on a small module — a cost the plan accepts explicitly under Approach and Risks.
  - plan alignment: R1, R2, R7, R8, R10, R11, R12, R13, R14 satisfied. R3 **not** satisfied — M1 leaves the `origin/HEAD` path unverified. R4 **not** satisfied — M2 leaves a false enforcement claim in the sentence R4 governs. R5, R6, R9 satisfied. No avoided_surfaces breached; NG1-NG7 all held.
  - verdict: REQUEST_CHANGES
  - judge: independent
  - judge_model: conflicting self-report — the reviewing agent stated its system prompt names both `claude-sonnet-5` (environment preamble) and `claude-opus-5` (env block) and declined to pick silently; it was spawned with `model: opus`. Recorded as the agent reported it rather than resolved.
  - proof_gaps: none for the automated gate. M1 and M2 are cited as deliberately failing regression proofs per step 9; both flip to pass once fixed.
  - receipt: context_sources: docs/plans/active/multi-stack-harness-readiness.md, cli/docs/embedded/playbooks/{check,watzup,work}.md, docs/playbooks/{check,watzup,work}.md, cli/docs/embedded/parity_test.go, cli/docs/embedded/templates/project.identity.md, cli/internal/installer/installer.go, cli/internal/installer/installer_test.go, .github/workflows/cli-ci.yml, docs/PROJECT.md, skills/workflow/git/references/{workflow,branch-management}.md, docs/patterns/encoding-invariants.md; policy: docs/playbooks/check.md#full, docs/playbooks/handoff.md#6; judge: independent; judge_model: see the conflicting self-report above; retries: 0; rollback_point: pre-initiative commit aba7057 (working tree still uncommitted; verified byte-identical to a pre-review snapshot after the judge's probes); failure_ledger: absent; enforcement: local-only (the pre-commit hook is not installed in this checkout — `bash scripts/install-git-hooks.sh --force` is the opt-in, and CI re-runs the guards); not_independently_verified: the CI path-filter trigger was proven by reading `.github/workflows/cli-ci.yml` and by `go list ./...`, not by observing a real GitHub Actions run on a docs-only commit; and the `zharness update` conflict path was proven through the installer's own code in a temp repository, not by running a released binary against a real consumer repo.

- `2026-09-06T15:20:00Z` | phase: playbook-truth | mode: full (step-11 verification follow-up) — verdict `REQUEST_CHANGES` — judge: independent
  - scope: on target | depth: deep — this entry is appended, not a rewrite; Validation is append-only and the entry above stands as written.
  - trigger: check.md step 11 requires verifying durable synchronization. Verifying that the entry above would actually survive the commit-time guard surfaced a third major finding that neither the automated gate nor the independent judge had exercised.
  - proof (FAILS ON PURPOSE — M3 regression): `b=$(sed -n "/^## Validation/,/^## Current State/p" docs/plans/active/multi-stack-harness-readiness.md); n=$(printf "%s\n" "$b" | grep -c "^- "); v=$(printf "%s\n" "$b" | grep -cE "^- .*verdict[^A-Za-z]*(APPROVED|APPROVE_WITH_REQUESTS|REQUEST_CHANGES)"); echo "entries=$n first-line-verdicts=$v"; [ "$n" -eq "$v" ]` -> FAIL (exit 1; `entries=5 first-line-verdicts=0`) — passes once every entry carries its verdict token on its own first line.
  - finding M3 (MAJOR, correctness/honesty) — every `## Validation` entry in this plan is invisible to the commit-time proof guard. The guard reads the verdict token from the entry's **first line only** (`scripts/install-git-hooks.sh:70-77`, `zharness_anchored_verdict`, documented there as the guard-v3 R1 first-line rule and asserted as intended behavior by `scripts/test-guards.sh:51-57`, "R1 verdictless first line yields no verdict despite body prose"). Every entry here puts `verdict:` on an indented sub-bullet, so the guard resolves an empty verdict, hits `*) continue` at `scripts/install-git-hooks.sh:162-165`, and skips the entry entirely. Reproduced by sourcing the extracted guard core and running `zharness_dump_entries` plus `zharness_anchored_verdict` over this file: 5 entries split correctly, all 5 report `verdict=[]` and `proofs=0`. Failure scenario: the four APPROVED gate entries for phases 1-4 commit with **zero** proof re-execution in a repository that installed the hook — the R2 proof guarantee that check.md step 9 calls "the sole proof guarantee" and that each of those receipts leans on never runs. This is M2's class (asserting enforcement that does not apply) reached through entry shape rather than through prose.
  - finding M3b (MAJOR, playbook truth) — the same defect is reachable by an agent that follows `check.md` correctly. Step 8 and the entry-field contract at `docs/playbooks/check.md:28` require "timestamp, stable phase slug, exact command/result and concise output, verdict, judge declaration ... with every proof command written as a nested sub-bullet under the entry so the commit-time guard can find them". That sentence states the sub-bullet rule for **proofs** and states nothing about where the **verdict** must sit, so an entry authored to the letter of check.md is guard-invisible. The proof-bullet shape is also under-specified: the guard matches `^[[:space:]]{2,}- ` immediately followed by a single-backtick-delimited command (`scripts/install-git-hooks.sh:79-84`), so the labeled form used throughout this plan (`- proof (…): \`cmd\``) and any double-backtick wrapping are silently dropped. Fix belongs in `cli/docs/embedded/playbooks/check.md` and its projection (R13), same phase as M2.
  - security review: no new findings beyond the entry above. M3 is an enforcement-visibility defect, not an injection or secrets defect.
  - performance review: no findings.
  - plan alignment: unchanged from the entry above — R3 and R4 remain unsatisfied. M3/M3b add no new requirement breach; R6's `enforcement: local-only` declaration in every receipt is literally accurate for this checkout, which is why the defect stayed invisible until the guard was executed rather than read.
  - verdict: REQUEST_CHANGES
  - judge: independent
  - judge_model: conflicting self-report as recorded in the entry above; this follow-up finding was produced by the reviewing agent executing the guard, not by the spawned judge, and is labeled as such rather than credited to it.
  - proof_gaps: none. M1, M2 and M3 all carry deliberately failing regression proofs that flip to pass once fixed.
  - receipt: context_sources: scripts/install-git-hooks.sh, scripts/test-guards.sh, docs/plans/active/multi-stack-harness-readiness.md, docs/playbooks/check.md; policy: docs/playbooks/check.md#full, docs/playbooks/check.md#11; judge: independent; judge_model: see above; retries: 0; rollback_point: pre-initiative commit aba7057; failure_ledger: absent; enforcement: local-only (the hook is not installed in this checkout; the guard core was extracted with the same awk range CI uses and executed directly against this file); not_independently_verified: the guard was exercised through its extracted core functions in-process, not by installing the hook and running a real `git commit`.

- `2026-09-06T15:30:00Z` | phase: playbook-truth | mode: full (independent judge report) — verdict `REQUEST_CHANGES` — judge: independent
  - scope: on target | depth: deep | lane: normal (step-6 required proof = unit + command output; both present)
  - trigger: fresh-context reviewing agent that did not author this diff, reporting the complete Security / Performance / Architecture / Code Quality review. First `full`-shaped pass over the change — the four phase entries above all declare `mode: gate`, so per check.md step 5 none of them performed the complete manual review.
  - gate commands re-run independently, real results:
    - `cd cli && go build ./...`
    - `cd cli && go vet ./...`
    - `cd cli && go test ./... -count=1`
    - `gofmt -l cli`
    - `bash scripts/verify-doc-links.sh`
    - `bash scripts/test-guards.sh`
    - all six pass: doc links `0 findings; 10 claim(s) under known-removed v0.15 surfaces`; guards `25 passed, 0 failed`; gofmt empty.
  - negative proof (parity test is not vacuous): appending one byte to `docs/playbooks/handoff.md` produced `parity_test.go:84: playbooks/handoff.md: embedded and projected copies differ`; file restored and re-verified byte-identical to HEAD. Disclosed: this probe used `git checkout --` on that one tracked path, which carried no uncommitted work; post-hoc `git diff --quiet -- docs/playbooks/handoff.md` confirms no loss.
  - negative proof (installer test is not vacuous): in a scratchpad copy of `cli/`, replacing the shipped template with HEAD-plus-appended gate section — the plan's explicitly rejected alternative — produced `installer_test.go:722: expected the gate-slot rewrite to conflict with a filled answer FAIL`. The test discriminates exactly the in-place-vs-append property R8 rests on.
  - CI path filters parsed independently: both `push.paths` and `pull_request.paths` contain `docs/playbooks/**` and `docs/WORKFLOW.md`; `build-test` carries no job-level `if`. R2 satisfied.
  - finding J1 (major, sharpens M1) — `cli/docs/embedded/playbooks/watzup.md:17`: the `origin/HEAD`-derived base is the one candidate never verified. `sed 's|^origin/||'` strips `origin/main` to a bare `main` and hands it to `git log` / `git rev-list`; `git show-ref --verify` is applied only to the `main`/`master` fallback candidates. Reproduced: in a checkout where `origin/HEAD -> origin/main` with no local `refs/heads/main`, `git log --oneline "main..HEAD"` exits 128 — the standard shape of a CI checkout that fetches only the PR ref, breaking the first command of the session-start playbook. R3 requires a *verified* base. Fix: drop the strip, or gate it with `git show-ref --verify --quiet "refs/heads/$base" || base=""` before the fallback loop.
  - finding J2 (major, narrows M2) — `cli/docs/embedded/playbooks/check.md:42`: R4 (plan:54) explicitly sanctions pointing at `bash scripts/install-git-hooks.sh --force` as the opt-in, so naming the script is *not* the defect. What survives is the trailing unqualified clause ``(`scripts/install-git-hooks.sh`; CI re-runs it)``: `AllTargets()` (`cli/internal/installer/installer.go:66-76`) distributes no workflow file, so "CI re-runs it" is false for every consumer repository and true only for mono-harness itself. R4's success signal is unmet by that clause, unqualified, inside the sentence the truth phase edited.
  - finding J3 (minor) — `check.md:40` vs `:42` assign the same hook-not-installed state two different ladder levels (`Local validation` at step 7, `Optional hook` at step 9). An agent reading both cannot emit a consistent `enforcement:` receipt. Align step 7 to `Optional hook`, matching `docs/patterns/encoding-invariants.md:64-70`.
  - finding J4 (minor) — `check.md` receipt vocabulary `enforcement: hook | ci | local-only` has no correct value for "script present, hook not installed": three values do not map onto the four ladder levels R4 cites, so mono-harness's own state is unrepresentable in the key R6 added.
  - finding J5 (minor) — `cli/docs/embedded/playbooks/work.md:35`: the absence-branch awk hardcodes `## Progress` with no `{heading}` slot while the `plan-slice.sh` line at `:32` directly above is parameterized, and the prose at `:38` tells the agent to read Phases, Progress AND Decisions. An agent without `plan-slice.sh` copies it literally and gets Progress only.
  - finding J6 (minor) — `cli/docs/embedded/playbooks/work.md:32` leaks the internal Go symbol `AllTargets()` into a consumer-shipped playbook, the same category as the problem that sentence fixes.
  - finding J7 (minor) — `docs/PROJECT.md` adds an `also required:` slot the identity template does not define, mildly against R7's aim that mono-harness not diverge from the template it ships.
  - coverage, base-branch class (A1-X1): INCOMPLETE — do not record as closed. Fixed in `watzup.md` step 1 but unverified there (J1), and three call sites still hardcode `main`: `skills/workflow/git/references/workflow.md:22-23`, `:88`, and `branch-management.md:33`. Reproduced in this repository: `origin/HEAD -> origin/master`, `refs/remotes/origin/main` absent, `git log origin/main...origin/master` exits 128. Mechanically, phase 4's proof at plan:482 is a negative grep bounded to Step 1 of `workflow.md`, which is why the residual passed its own gate without ever seeing `:88` or `branch-management.md:33`.
  - coverage, Node-assumption class (R11, A2-G1/R1/P1): COMPLETE across shipped surfaces. The audit's own grep leaves three hits, all correct: `workflow.md:28` and `:145` (anti-pattern mentions of `node_modules`) and `workflow.md:52` (the corrected multi-stack `deps:` line). Zero Node-shaped assumptions remain in `cli/docs/embedded/`, `docs/playbooks/`, or `docs/WORKFLOW.md`; the identity template's gate slots are stack-neutral.
  - security review: no findings. Every expansion in the new snippets is quoted; `git` ref-name rules exclude spaces and metacharacters; the only `--force` occurrences are `install-git-hooks.sh --force`. Improvement recorded: `git add -A` now survives in `skills/` only at `workflow.md:145`, where it is the prohibition.
  - performance review: no findings. `TestProjectionParity` is a fixed 7-file byte compare; the widened CI path filters run the Go job on doc-only edits, a cost the plan accepts explicitly.
  - code quality: no blocking findings. The `reporter` interface lets `TestProjectionParity_DetectsDrift` capture failures rather than fatal the outer test; drift is caught in both directions plus an extra-projected-file loop; `AGENTS.md` is correctly excluded (embedded and repo-root copies genuinely differ). Vacuous pass ruled out: `t.Fatalf` on an unreadable projection root, no `t.Skip` in the package, and an emptied embedded tree fails at `go:embed` build time. `identityPreEdit` is byte-identical to `git show HEAD:cli/docs/embedded/templates/project.identity.md` (543 bytes) — a frozen fixture guarded by the `## What are the gate commands?` assertion. No `t.Parallel()`, so the `srcBytesImpl` swap is safe. R7's 50-line ceiling holds at 27 lines.
  - plan alignment: R1, R2, R7, R8, R10-R14 satisfied. R3 and R4 remain unsatisfied (J1, J2). Findings M3/M3b from the entry above are outside this judge's scope — it reviewed the diff, not the Validation entries' guard visibility — and remain open.
  - verdict: `REQUEST_CHANGES`
  - judge: independent
  - judge_model: conflicting self-report inside the reviewing session — its preamble states `claude-sonnet-5`, its environment block states `claude-opus-5`; spawned model was opus. Recorded unresolved rather than guessed.
  - proof_gaps: none for `lane: normal` (unit + command output both present)
  - receipt: context_sources: diff + plan + both audits + `installer.go` + `parity_test.go` + `installer_test.go` + `cli-ci.yml` + `encoding-invariants.md` / policy: check.md full / judge: independent / judge_model: see above / retries: 0 / rollback_point: `aba7057` / failure_ledger: absent / enforcement: local-only / not_independently_verified: the reviewing agent re-ran the gate commands and both negative probes itself, but no commit was made and no pre-commit hook was installed, so nothing here was proven under real commit-time enforcement.

- `2026-09-06T16:22:00Z` | phase: playbook-truth | mode: gate (remediation waves R1-R4) — verdict `APPROVED` — judge: same-session
  - scope: on target | depth: standard | lane: normal (step-6 required proof = unit + command output; both present)
  - remediates M1/J1, M2/J2, J3, M3b, J5, J6 and M3 from the two `REQUEST_CHANGES` entries above. All edits authored in `cli/docs/embedded/playbooks/` and projected byte-identically in the same commit (R13).
  - `cd cli && go test ./docs/embedded/ -run TestProjectionParity -count=1` -> pass
  - `bash scripts/test-guards.sh` -> pass (`guards: 25 passed, 0 failed`)
  - `bash scripts/verify-doc-links.sh` -> pass (`doc links OK (0 findings; 10 claim(s) under known-removed v0.15 surfaces)`)
  - `cd cli && go build ./... && go vet ./... && go test ./... -count=1` -> pass (all packages ok, including `TestOnePlan_PlaybookContract`)
  - `gofmt -l cli` -> pass (empty output)
  - `sh -c 'd=$(mktemp -d); git init -q "$d/up"; cd "$d/up" || exit 9; git config user.email t@t; git config user.name t; git symbolic-ref HEAD refs/heads/main; echo a>a; git add a; git commit -qm i; cd "$d" || exit 9; git clone -q up cl; cd cl || exit 9; git config user.email t@t; git config user.name t; git checkout -q -b feat; git branch -D main -q; base=$(git symbolic-ref --quiet --short refs/remotes/origin/HEAD 2>/dev/null | sed "s|^origin/||"); [ -z "$base" ] || git show-ref --verify --quiet "refs/heads/$base" || base=""; [ -n "$base" ] || for c in main master; do git show-ref --verify --quiet "refs/heads/$c" && base=$c && break; done; [ -n "$base" ] || base=$(git branch --show-current); git log --oneline "$base..HEAD" >/dev/null 2>&1'` -> pass (M1/J1: the same scenario that exits 128 under the unverified snippet now resolves to `feat` and succeeds)
  - `sh -c 'f=cli/docs/embedded/playbooks/watzup.md; awk "/^1\. \*\*Read branch state/,/^2\./" "$f" | grep -q "refs/heads/\$base"'` -> pass (the playbook itself carries the verification gate, not just the prose)
  - `sh -c '! grep -qE "install-git-hooks[.]sh.{0,3}; CI re-runs it" cli/docs/embedded/playbooks/check.md'` -> pass (M2/J2: the unqualified clause is gone; the R4-sanctioned `--force` referral is retained. Written with a character class rather than a literal backtick — see M4: the guard's proof-extraction pattern stops at the first inner backtick, so a command containing one is truncated and re-executes as a shell syntax error.)
  - `sh -c '! grep -q "Local validation" cli/docs/embedded/playbooks/check.md'` -> pass (J3: steps 7 and 9 now name the same ladder level)
  - `sh -c 'f=cli/docs/embedded/playbooks/check.md; grep -q "verdict token on the entry.s own first line" "$f" && grep -q "single backticks" "$f"'` -> pass (M3b: the step-8 contract now states both positions)
  - `sh -c '! grep -q "AllTargets()" cli/docs/embedded/playbooks/work.md'` -> pass (J6: no internal Go symbol in a consumer-shipped playbook)
  - `sh -c 'f=cli/docs/embedded/playbooks/work.md; grep -q "plan-slice.sh" "$f" && grep -q "absent" "$f" && grep -q "awk" "$f"'` -> pass (T2.3 still holds after the J5 heading-slot change)
  - `sh -c 'for f in WORKFLOW.md playbooks/check.md playbooks/watzup.md playbooks/work.md; do diff -q "cli/docs/embedded/$f" "docs/${f}" >/dev/null || exit 1; done'` -> pass (R13)
  - `sh -c 'b=$(sed -n "/^## Validation/,/^## Current State/p" docs/plans/active/multi-stack-harness-readiness.md); n=$(printf "%s\n" "$b" | grep -c "^- "); v=$(printf "%s\n" "$b" | grep -cE "^- .*verdict[^A-Za-z]*(APPROVED|APPROVE_WITH_REQUESTS|REQUEST_CHANGES)"); [ "$n" -eq "$v" ]'` -> pass (M3: every entry now carries its verdict on its own first line; this same command exited 1 with `entries=5 first-line-verdicts=0` before the reshape)
  - M3 end-to-end evidence: a real `zharness_guard_entries_of_file <path> <old-file> <new-file>` run over this plan re-executed **24** proof commands with **0** rejections and exit 0. The identical run before the reshape re-executed **0**. The four APPROVED gate entries are now actually proof-guarded rather than silently skipped.
  - plan alignment: R3 and R4 are now satisfied — the two requirements the previous entries recorded as unmet. R1, R2, R5, R6, R7, R8, R9, R10-R14 unchanged and still satisfied. No avoided surface was touched: `work.md:17`'s bounded-mode ceiling (NG1), `scripts/`, `cli/internal/`, `cli/docs/embedded/templates/` and `skills/workflow/git/` are all unmodified by this wave.
  - deferred with rationale in Decisions, not silently dropped: J4 (would change R6's accepted `enforcement:` vocabulary — escalate_when), J7 (`docs/PROJECT.md` is this phase's avoided surface), R15/A1-X1 (owner-ruled a new requirement for `to-plan phase git-skill-multistack`), F1 and D2 (minors).
  - verdict: `APPROVED`
  - judge: same-session
  - judge_model: claude-opus-5 (this session's environment block; the same session authored these edits, hence the same-session declaration)
  - not independently verified (required for a same-session clean verdict): the M2/J2 rewording is asserted to be *honest* for a consumer repository on the strength of reading `AllTargets()` at `cli/internal/installer/installer.go:66-76`, not by running `zharness install` into a scratch repository and observing which files land. The claim that no workflow is distributed is therefore source-read, not behaviourally proven. Likewise, the guard evidence comes from sourcing the extracted `ZGUARD-CORE` block in-process; no hook was installed and no real `git commit` was attempted.
  - proof_gaps: none for `lane: normal`. This is a `gate`, not a `full`: per check.md step 5 it performs no complete Security/Performance/Architecture/Code Quality review, and per `handoff.md` step 6 the initiative still owes exactly one `full` review with an independent judge before closure.
  - receipt: context_sources: the two `REQUEST_CHANGES` entries above + the independent judge report + `install-git-hooks.sh` + `installer.go` + `cli-ci.yml` + `encoding-invariants.md` / policy: work.md step 11 in-session gate / judge: same-session / judge_model: claude-opus-5 / retries: 0 / rollback_point: `aba7057` / failure_ledger: absent / enforcement: local-only / not_independently_verified: see the dedicated line above
- `2026-09-06T16:45:00Z` | phase: playbook-truth | mode: gate (wave R5, work.md step 11 in-session) — verdict `APPROVED` — judge: same-session
  - scope: on target — `cli/docs/embedded/playbooks/check.md` + its projection only.
  - depth: standard
  - `cd cli && go build ./... && go vet ./... && go test ./... -count=1` -> pass
  - `gofmt -l cli` -> pass (no output)
  - `cmp cli/docs/embedded/playbooks/check.md docs/playbooks/check.md` -> pass (byte-identical, R13 holds)
  - `bash scripts/verify-doc-links.sh` -> pass
  - `bash scripts/test-guards.sh` -> pass (25 passed, 0 failed)
  - `grep -c "sole proof guarantee" cli/docs/embedded/playbooks/check.md` -> pass (1; TestOnePlan_PlaybookContract pin intact)
  - `! grep -q "and nothing else, so a label prefix" cli/docs/embedded/playbooks/check.md` -> pass (absent; the false shape rule and its garbled inline example are gone)
  - `grep -c "no backtick anywhere inside the command" cli/docs/embedded/playbooks/check.md` -> pass (1; M4's constraint is now named in the contract)
  - not_independently_verified: this is the same-session author gating their own diff. The claim not independently checked is that the reworded contract is *complete* — it is verified against the extraction pattern in `scripts/install-git-hooks.sh` and against the shapes exercised by `drive5.sh`, but no consumer repository has been run against the new wording.
  - proof_gaps: none for `lane: normal`. Still a `gate`, not a `full`; the initiative's one complete review remains owed at `handoff.md` step 6.
  - receipt: context_sources: the 2026-09-06T16:22:00Z gate entry + advisor review + `scripts/install-git-hooks.sh` extraction pattern / policy: work.md step 11 in-session gate / judge: same-session / judge_model: claude-opus-5 / retries: 0 / rollback_point: `aba7057` / failure_ledger: absent / enforcement: local-only / not_independently_verified: see the dedicated line above

## Current State and Next Action
- active_phase: playbook-truth
- lifecycle_status: checked
- latest_anchor: 2026-09-06T16:45:00Z — remediation waves R1-R5 gated `APPROVED` in-session per `work.md` step 11; phase set to `checked`, matching the three sibling phases (see decision J8).
- blockers:
  - none for this phase. M1/J1, M2/J2, J3, M3b, J5, J6 and M3 are all fixed, projected under R13, and carry passing proofs in the Validation entry dated 2026-09-06T16:22:00Z. R3 and R4 are now satisfied.
- open_items:
  - R15 (owner-ruled in scope, 2026-09-06) A1-X1 base-branch coverage is INCOMPLETE: `skills/workflow/git/references/workflow.md:22`, `:23`, `:88` (consumed unquoted at `:90`) and `branch-management.md:33` still hardcode `main`; `/git pr` fails in this repo today (`origin/main` absent; `git log origin/main...origin/master` exits 128). Authoring R15 and its tasks belongs to `to-plan phase git-skill-multistack`, which must also reopen that phase from `checked` to `in-progress` — neither `work` nor `check` reopened it, since it was not the phase under gate. R15's verification must scan the whole of `workflow.md`, not Step 1: the judge showed phase 4's proof is a negative grep bounded to Step 1, structurally blind to `:88` and `branch-management.md:33`. The corrected resolution to mirror is the one now in `watzup.md` step 1: `origin/HEAD` stripped, kept only if `git show-ref --verify --quiet refs/heads/<name>` confirms it, then `main`/`master`, then `git branch --show-current`.
  - J4 (minor, DEFERRED — see Decisions) `check.md`'s `enforcement: hook | ci | local-only` has no value for "script present, hook not installed"; three values against the ladder's four levels. Widening it changes R6's accepted text, so it is an owner decision for `brainstorm`/`to-plan`, not a work-phase edit.
  - J7 (minor, DEFERRED — see Decisions) `docs/PROJECT.md` carries an `also required:` slot the shipped identity template does not define. That file is this phase's avoided surface; it belongs to `identity-gate-slots`.
  - M4 (minor, found by the guard itself; HALF FIXED in wave R5) a proof command containing a literal backtick is unrepresentable: the extraction pattern reads a bare command between single backticks, truncates at the first inner one, and re-executes the fragment as a shell syntax error. `check.md`'s step-8 contract now names the constraint, so no agent writes such a bullet unknowingly. The guard-side fix lives in `scripts/install-git-hooks.sh` — this phase's avoided surface — and stays an owner decision.
  - J8 (minor, NEW — spine-playbook truth defect) `work.md` step 11 contradicts itself: it orders `check.md`'s gate steps 1-4 and 6-11 performed in-session, step 10 of which mandates setting `checked` on `APPROVED`, and then its closing sentence says "Do not mark the phase checked or done". All four phases of this initiative resolved the contradiction toward `checked`. Which clause wins is an owner call: either drop the closing prohibition, or exclude step 10 from the borrowed range. Fixing it edits `work.md`, already this phase's surface, but the choice is a requirements change, so it routes to `brainstorm`/`to-plan` per `work.md`'s own `escalate_when`.
  - F1 (minor) `cli/docs/embedded/parity_test.go:48` uses `filepath.Join` for an `fs.FS` path; latent on Windows only (CI is ubuntu-latest). Fix is `path.Join`.
  - D2 (minor) the phase-1 `projection-parity-gate` receipt omits R6's `enforcement:` field.
  - check.md follow-up: precondition 2 and step 1 demand the selected phase read `in-progress`, which is structurally unsatisfiable for an initiative-level `full` at handoff closure, since `work.md` step 11 has already set every phase to `checked`.
  - all four phases remain uncommitted; rollback point is `aba7057`.
- exact_next_action: `to-plan phase git-skill-multistack` — author R15 (mirror the corrected `watzup` base resolution into `skills/workflow/git/references/workflow.md:88`, fix `branch-management.md:33`, restate the `pr`/`merge` defaults at `:22-23` as "the repository's default branch"), give it a whole-file verification rather than a Step-1-bounded grep, and reopen that phase from `checked` to `in-progress`. Then `work full phase git-skill-multistack`. After that phase gates clean, run the initiative's single `check full` with an independent judge (`handoff.md` step 6 requires it exactly once, on the final phase) — and before `git full`, re-run `zharness_guard_entries_of_file <path> <old-file> <new-file>` (three arguments; a one-argument call blocks awk on stdin and returns a vacuous exit 0) to confirm every Validation entry is still guard-visible.
