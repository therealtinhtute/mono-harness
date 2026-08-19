---
id: 01M0C5EMT7HMP1RGZH9MV5GKET
type: plan
intake_id: 01M0C5EQETZ60HJDWXG3JJHVTA
lane: normal
status: active
created: 2026-08-19
updated: 2026-08-19
---

# Plan: docs architecture — authored project docs, then consumer scaffold

## Outcome
- result: every file under `docs/` has an unambiguous owner (managed, authored, or scaffold-once), and this repository gains real authored documentation reachable from one entrypoint — first by hand for this repo, then by `zharness init` for consumer repos.
- success_signals:
  - A fresh agent session answers "what is this repo's architecture and which decisions are locked" by reading `docs/README.md` plus one linked file, with zero `grep` over source.
  - Every path under `docs/` appears in the ownership table in `docs/README.md`; an unlisted path is a defect.
  - `bash scripts/verify-doc-links.sh` and `cd cli && go test ./...` both pass.
  - Phase A merges and stays useful even if Phase B never ships.

## Authority and Requirements
- authority:
  - `cli/internal/application/managed_docs.go:107-112` — membership in the embedded FS is the sole trigger for hash tracking; `AGENTS.md` is the only exclusion. A file outside that tree can never be tracked or conflict-staged.
  - Verified 2026-08-19: `grep -n "Remove\|Delete\|prune\|trash" cli/internal/application/managed_docs.go` returns empty — no code path deletes a local file, so non-embedded files survive `init --refresh-docs` untouched.
  - `cli/internal/application/init.go:59-104` — `writeAgentsManagedBlock`'s absent-file branch (`os.IsNotExist` → write, else leave alone) is the existing shape for a write-once-if-absent scaffold.
  - `cli/internal/application/init_test.go:12,28,70` — `docsVersion` is a per-call parameter against a `fstest.MapFS` fixture, not a global constant; bumping the real docs version does not break this test.
  - `docs/audit/consumer-adoption-audit.md:135` — the 26,365-token figure is explicitly "Inferred, not observed"; it is a file size, attributed to the ambiguous-active-plan invariant, which commit `7a4195f` already closed.
  - `docs/plans/completed/durable-memory.md:38` (NG3) — wiring `zharness memory` into playbooks is deferred to a separate initiative.
  - Measured 2026-08-19 on a clean tree at `1838a7b`: `bash scripts/verify-doc-links.sh` reports **16 broken doc cross-references** and exits non-zero. `CLAUDE.md` links two docs that do not exist (`docs/prompt-engineering-principles.md`, `docs/workflow-harness/migration.md`); four playbooks link three missing audit files; `skills/workflow/README.md` links six missing files. The repository's own `check` gate is therefore red before any change in this initiative, which is direct observed evidence — unlike the inferred token figure — that authored documentation is missing rather than merely unrouted.
  - Root cause of the 16 broken references, traced 2026-08-19: commit `655c6ac` "Delete docs directory" (2026-08-16, authored via the GitHub web UI, which runs no gate) removed all 26 files under `docs/`, 4,285 lines. The next `zharness init` regenerated exactly the 8 files present in the embedded FS (`WORKFLOW.md` plus 7 playbooks) and none of the other 18, ~3,733 lines. The correspondence is exact with no exception. The managed half silently regenerating is what hid the authored half's loss, which makes this initiative incident recovery rather than tidying. All 18 remain retrievable at `655c6ac^`.
  - Per-reference triage, 2026-08-19: only 2 of the 10 missing targets are actually required. `docs/prompt-engineering-principles.md` (143 lines) is an unexecutable mandatory instruction — `CLAUDE.md:98` orders it read before any skill or rule edit. `docs/workflow-harness/migration.md` (52 lines) is the sole pointer for legacy adoption at `CLAUDE.md:94`. `docs/evals/failures.md` is a checker false positive: `check.md:35` states a repository without it "skips this without it being an error". The remaining 7 files, 2,249 lines of audits and history, are cited only as provenance for rationale already written inline at the citation site.
  - Fixability constraint, verified 2026-08-19: `docs/playbooks/{work,check,watzup,handoff}.md` are byte-identical to their `cli/docs/embedded/playbooks/` sources, and 8 of the 16 broken references originate there. Removing those citations honestly requires editing the embedded source and cutting a `cli/v*` release, placing them in Phase B. Phase A can only mark them via `.claimignore`. The remaining 8 references live in `skills/workflow/README.md` (6) and `CLAUDE.md` (2), both authored and freely editable in Phase A.
  - Owner decision, 2026-08-19: split the work so markdown-only lands first and binary changes are a separate, optional second phase.
  - `hoangnb24/repository-harness` — upstream docs taxonomy (`docs/README.md` router, `ARCHITECTURE.md`, `decisions/NNNN-*.md` with an index, `templates/`), and its stated principle "start with the smallest authoritative surface".
- requirements:
  - R1 [accepted]: `docs/README.md` exists in this repo and contains an ownership table naming, for every existing path under `docs/`, exactly one class — `managed` (present in `cli/docs/embedded/`), `authored` (human-written, never embedded), or `scaffold-once` (written by `init` only when absent). | source: `managed_docs.go:107-112`
  - R2 [accepted]: `docs/README.md` is the single entrypoint; every authored doc is reachable by link from it. No authored doc is discoverable only by directory listing. | source: upstream `repository-harness` docs router
  - R3 [accepted]: no file created by this initiative is added to `cli/docs/embedded/`. Verification: `find cli/docs/embedded -name README.md -o -name ARCHITECTURE.md -o -name decision.md` returns nothing. | source: `managed_docs.go:112`
  - R4 [accepted]: `docs/ARCHITECTURE.md` states the harness's own design in prose, and every structural claim cites a real `path:line` that exists at merge time — at minimum: markdown as source of truth, `harness.db` as a derived rebuildable index, managed-docs hash tracking with conflict staging, and the single `ResolveActivePlan` contract. | source: `docs/plans/completed/harness-markdown-truth.md`, `cli/internal/application/plan_resolve.go:73`
  - R5 [accepted]: `docs/decisions/` holds numbered ADRs with an index README, and the decisions already made and verified — the D1 single-resolver contract, markdown-as-truth, and the deliberate deferral of memory-playbook wiring — are each recorded as an ADR rather than left only in commit messages or session transcripts. | source: `docs/plans/completed/*.md`, commit `7a4195f`
  - R6 [accepted]: Phase A touches only new files plus `docs/README.md`-adjacent authored paths; it changes no Go source and requires no release. Verification: the Phase A diff contains no `.go` file and no path under `cli/docs/embedded/`. | source: owner decision 2026-08-19
  - R7 [accepted]: Phase B adds a scaffold-once class to `zharness init` that writes `docs/README.md`, `docs/decisions/README.md`, and a decision template only when each is absent, records no `managed_docs` row, and leaves consumer edits byte-identical across `init --refresh-docs` with zero new `.kit/conflicts/` entries. | source: `init.go:59-104`, verified absence of any deletion path
  - R8 [accepted]: Phase B adds routing by convention only, by repairing the sentence that is already the defect. `cli/docs/embedded/AGENTS.md` line 5 currently reads "Read `docs/WORKFLOW.md`, then only the returned stage playbook and repository material relevant to the requested outcome" — the final clause names no path, so in a fresh consumer repo it resolves to a blind search. Phase B rewrites that clause to name `docs/README.md` and to state that its absence is not an error. No second routing surface is added: `cli/docs/embedded/WORKFLOW.md` is not modified for routing, no new `preflight` field is introduced, and no per-repo configuration exists. | source: `cli/docs/embedded/AGENTS.md:5`; owner decision 2026-08-19 recorded as decision `01M0C99PNHZX3A1CRK7EGHQ23T`
  - R9 [accepted]: Phase B ships behind a `cli/v*` release tag; the bare `vX.Y.Z` tag is never pushed. | source: existing release convention in `.github/workflows`
  - R10 [accepted]: after Phase A, `bash scripts/verify-doc-links.sh` exits zero. Each of the 16 broken references is resolved by exactly one of: writing the missing doc, retargeting the link to a doc that exists, or adding a `.claimignore` entry carrying a `# reason`. Deleting a link to hide a genuinely missing doc is not an accepted resolution. | source: measured gate failure at `1838a7b`; `CLAUDE.md` gate-commands section
  - R11 [accepted]: `docs/prompt-engineering-principles.md` and `docs/workflow-harness/migration.md` are restored from `655c6ac^` in Phase A, reviewed for staleness against current behaviour, and corrected where they contradict it. The other 8 deleted files stay deleted. | source: per-reference triage above
  - R12 [accepted]: every `.claimignore` entry added for a citation that still lives in an embedded playbook states in its `# reason` that it is a Phase B deferral, not a permanent exemption, and names the embedded source path to edit. A `.claimignore` entry that reads as a permanent exemption for these 8 references fails this requirement. | source: fixability constraint above
  - R13 [accepted]: Phase B removes `git.md` from the projected doc set. The file moves from `cli/docs/embedded/playbooks/git.md` to `skills/workflow/git/references/workflow.md` — the git skill's existing `references/` convention — the `"git"` entry is deleted from `preflightPlaybooks`, and both of its two repo-wide citations are retargeted. `git.md` fails the embedded-membership test that governs this whole initiative: `cli/internal/interfaces/preflight.go:39` records that git owns no harness entity (D7), git is absent from `contextEligibleStages`, and `cli/docs/embedded/WORKFLOW.md` already states git keeps its skill-local procedure — yet git.md is the largest projected playbook at 148 lines / 7,330 bytes (13% of the 55,115-byte projection) while mentioning `zharness` only 4 times. No prune path is added to `managed_docs.go`: `planManagedDocActions` at `managed_docs.go:108` walks the embedded FS, so a stale `managed_docs` row in an already-initialized consumer repo is never visited, never errors, and its orphaned `docs/playbooks/git.md` is inert once nothing routes to it. That inertness is proven, not assumed. | source: measured inventory; `preflight.go:30,39`; `managed_docs.go:108`; decision `01M0C9YJV2Y3QVC22E1E9XGXRG`
  - R14 [accepted]: Phase B makes `zharness init` ensure `CLAUDE.md` carries an `@AGENTS.md` import inside the `ZHARNESS:BEGIN`/`END` markers. Anthropic's documentation states plainly that "Claude Code reads `CLAUDE.md`, not `AGENTS.md`" and prescribes exactly this import; native `AGENTS.md` support is issue #34235, still open. Without it the managed block reaches Codex, Cursor, and Gemini but never reaches Claude Code, which is this repository's primary consumer. The import satisfies "both files carry identical content" without duplicating a byte, because it is the same file expanded at launch. The symlink alternative the same documentation lists is rejected: it clobbers an existing project `CLAUDE.md` and requires Administrator rights on Windows. One consequence must be proven rather than assumed — the same documentation states block-level HTML comments in `CLAUDE.md` are stripped before injection, and the `ZHARNESS` markers are HTML comments; if stripping swallows the import, the fallback is to place `@AGENTS.md` outside the markers. | source: https://code.claude.com/docs/en/memory; https://github.com/anthropics/claude-code/issues/34235; decision `01M0CD04RKSXK91MR5M5S7EA7X`

## Non-goals
- NG1: no JSON, YAML, or frontmatter routing configuration. The reader is always an LLM, and the flat markdown table already in `WORKFLOW.md` proves markdown suffices.
- NG2: no edits to any file inside `cli/docs/embedded/` during Phase A, and no edits to the projected `docs/WORKFLOW.md` or `docs/playbooks/*.md` at any point in this initiative except the single `WORKFLOW.md` sentence in Phase B made at the embedded source. Projected copies are hash-tracked; local edits are conflict-staged to gitignored `.kit/conflicts/` and lost.
- NG3: no wiring of `zharness memory` into any spine playbook. `durable-memory.md:38` deferred that decision to a later initiative; bundling it here reopens closed scope.
- NG4: no placeholder or filler content. A doc ships with real content or is not created. Specifically, no empty `ARCHITECTURE.md` skeleton is scaffolded into consumer repos in Phase B.
- NG5: no auto-generated, AST-derived documentation (the `codesight` model). Reviewed and rejected for this repo — the docs here describe decisions, not code shape.
- NG6: the 26,365-token figure is not used as justification for any part of this work. It is inferred rather than measured and its stated cause was already fixed. This initiative is justified by ownership ambiguity and the absence of authored project docs, not by that number.
- NG7: no move or rename of `cli/docs/{CONTRACT,SCHEMA,STATE}.md`. They stay at their current paths and are indexed from `docs/README.md`.

## Approach and Risks
- approach: ship this repository's authored documentation first as markdown only — no Go, no release — then optionally teach `zharness init` to scaffold the same entrypoint into consumer repos. Ownership is decided by embedded-tree membership, so everything this initiative creates lives outside `cli/docs/embedded/` and is therefore never hash-tracked, never conflict-staged, and never deleted. The 16 broken references are cleared three different ways, chosen per reference by the triage in Authority: restore the 2 files a repository instruction actually requires, retarget the dead history pointers in `skills/workflow/README.md` at a new ADR that records the deletion, and add 3 target-substring `.claimignore` entries for the citations baked into embedded playbooks, each marked a Phase 2 deferral rather than a permanent exemption.
- rejected alternatives:
  - Restore all 18 deleted files from `655c6ac^`. Rejected: 2,249 of those lines are audits and history whose rationale is already written inline at every citation site, so restoring them dilutes current authority and contradicts "start with the smallest authoritative surface".
  - Clear all 16 references by editing the playbooks that contain them. Rejected: `docs/playbooks/{work,check,watzup,handoff}.md` are byte-identical projections of `cli/docs/embedded/`; a local edit is conflict-staged into gitignored `.kit/conflicts/` and lost on the next `init --refresh-docs`.
  - One phase covering both markdown and binary. Rejected by the owner 2026-08-19: it gates every documentation improvement behind a release.
- constraints:
  - Phase `p1-doc-authority` changes no `.go` file and no path under `cli/docs/embedded/` (R6).
  - `.claimignore` matches by target substring, so one entry clears every reference to that target; each entry still needs a `# reason` or the gate exits 2.
  - Both gate commands pass before any commit: `bash scripts/verify-doc-links.sh` and `cd cli && go test ./...`.
- risks:
  - The two restored docs may contradict current behaviour after three days of drift. Mitigation: task `p1.w1.t1` reviews both against current code and `zharness --help` before the phase gate, correcting contradictions in place rather than restoring verbatim.
  - `.claimignore` entries silently becoming permanent exemptions. Mitigation: R12 requires each to name the embedded source path and the Phase 2 deferral; `p2.w2.t2` removes them.
  - A target-substring entry masking a future genuinely broken link to the same target. Accepted: those targets were deliberately deleted in `655c6ac`, so a new link to them would itself be the defect.
  - Phase 2 never shipping. Accepted without mitigation — surviving that is the reason for the split.
- recovery: every deleted file remains at `655c6ac^` and is retrievable with `git checkout 655c6ac^ -- <path>`. `p1-doc-authority` touches only markdown and reverts with a single `git revert`.

## Phases and Verification
<!-- Phase and task definitions are immutable after to-plan. Do not add task status fields. Append-only Progress is the sole task execution-status source. Only each phase lifecycle status changes to mirror DB transitions: to-plan=planned; work after run create=in-progress; clean durable check=checked; closing handoff=done. Each planned phase records phase_slug, story_id, status, goal, depends_on, waves, tasks, and checks. -->
- planning_status: planned
- phases:
  - phase_slug: p1-doc-authority
    story_id: 01M0C8CE4E3ZANQM9PY9N5WQSG
    status: planned
    goal: restore the two required deleted docs, clear all 16 broken references, and give this repo authored documentation reachable from one entrypoint. Markdown only, no Go, no release.
    depends_on: none
    surfaces_allowed: docs/README.md, docs/ARCHITECTURE.md, docs/decisions/**, docs/prompt-engineering-principles.md, docs/workflow-harness/migration.md, .claimignore, skills/workflow/README.md
    surfaces_avoided: cli/** (all Go and all embedded docs), docs/playbooks/**, docs/WORKFLOW.md, CLAUDE.md, docs/plans/completed/**
    waves:
      - wave: 1
        name: recover what is required, mark what is not
        tasks:
          - id: p1.w1.t1
            requirement: R11
            do: run `git checkout 655c6ac^ -- docs/prompt-engineering-principles.md docs/workflow-harness/migration.md`, then read both in full and correct any statement that contradicts current behaviour before staging.
            check: `test -s docs/prompt-engineering-principles.md && test -s docs/workflow-harness/migration.md && bash scripts/verify-doc-links.sh 2>&1 | grep -c '^CLAUDE.md'` prints `0`.
            stop: if either file documents a command, flag, or path that no longer exists, correct it in place; restoring verbatim is a failure of this task, not a shortcut.
          - id: p1.w1.t2
            requirement: R10, R12
            do: add exactly three `.claimignore` entries, for targets `docs/audit/sdlc-token-cache-audit.md`, `docs/audit/workflow-harness-ceremony-audit.md`, and `docs/evals/failures.md`. Each `# reason` names commit `655c6ac`, states that the citation lives in `cli/docs/embedded/playbooks/`, and marks removal as deferred to `p2-consumer-scaffold`.
            check: `bash scripts/verify-doc-links.sh 2>&1 | grep -c '^docs/playbooks/'` prints `0`, and `grep -c 'cli/docs/embedded/playbooks' .claimignore` prints `3`.
      - wave: 2
        name: write the decision record and the architecture doc
        tasks:
          - id: p1.w2.t1
            requirement: R5
            do: create `docs/decisions/README.md` as a numbered index plus four ADRs — `0001-markdown-as-source-of-truth.md`, `0002-single-active-plan-resolver.md`, `0003-durable-memory-not-wired-into-playbooks.md`, `0004-docs-directory-deletion-655c6ac.md`. The fourth records what `655c6ac` removed, which 2 files were restored, which 8 stay deleted, and that all remain retrievable at `655c6ac^`.
            check: `ls docs/decisions/*.md | wc -l` prints `5`, and `bash scripts/verify-doc-links.sh 2>&1 | grep -c 'docs/decisions'` prints `0`.
          - id: p1.w2.t2
            requirement: R4
            do: write `docs/ARCHITECTURE.md` covering markdown as source of truth, `harness.db` as a derived rebuildable index, managed-docs hash tracking with conflict staging, and the single `ResolveActivePlan` contract. Every structural claim carries a `path:line` citation.
            check: `grep -oE '[A-Za-z0-9_./-]+\.go:[0-9]+' docs/ARCHITECTURE.md | sort -u | while IFS=: read -r f l; do { [ -f "$f" ] && [ "$(wc -l < "$f")" -ge "$l" ]; } || echo "BAD $f:$l"; done` produces no output.
      - wave: 3
        name: build the entrypoint and clear the last dead pointers
        tasks:
          - id: p1.w3.t1
            requirement: R1, R2
            do: write `docs/README.md` with an ownership table classifying every existing path under `docs/` as `managed`, `authored`, or `scaffold-once`, plus a link to every authored doc including `cli/docs/{CONTRACT,SCHEMA,STATE}.md` at their current paths.
            check: `find docs -mindepth 1 -maxdepth 1 | sed 's|^docs/||' | while read -r p; do grep -q "$p" docs/README.md || echo "MISSING $p"; done` produces no output.
          - id: p1.w3.t2
            requirement: R10
            do: in `skills/workflow/README.md`, retarget the six dead history citations at `docs/decisions/0004-docs-directory-deletion-655c6ac.md`, keeping each surrounding claim intact.
            check: `bash scripts/verify-doc-links.sh 2>&1 | grep -c 'skills/workflow/README.md'` prints `0`.
      - wave: 4
        name: phase gate
        tasks:
          - id: p1.w4.t1
            requirement: R10
            do: run both repository gate commands.
            check: `bash scripts/verify-doc-links.sh; echo "exit=$?"` prints `exit=0`, and `cd cli && go test ./...` reports `ok` for every package with no `FAIL`.
          - id: p1.w4.t2
            requirement: R3, R6
            do: prove the phase honoured its surface constraints.
            check: `git diff --name-only $(git merge-base master HEAD)..HEAD -- '*.go' 'cli/docs/embedded/**'` produces no output, and `find cli/docs/embedded \( -name README.md -o -name ARCHITECTURE.md -o -name decision.md \)` produces no output.
  - phase_slug: p2-consumer-scaffold
    story_id: 01M0C8CE4RBXFFPBHV793RQ3VT
    status: planned
    goal: teach `zharness init` to scaffold-once the authored-docs entrypoint into consumer repos without hash-tracking it, route to it by convention, and ship behind a `cli/v*` tag.
    depends_on: p1-doc-authority
    surfaces_allowed: cli/internal/application/init.go, cli/internal/application/init_test.go, cli/docs/embedded/AGENTS.md, cli/docs/embedded/playbooks/{work,check,watzup,handoff}.md, cli/docs/embedded/playbooks/git.md (deletion only, R13), cli/internal/interfaces/preflight.go (the `preflightPlaybooks` map only, R13), skills/workflow/git/{SKILL.md,references/workflow.md}, docs/playbooks/git.md (deletion only — orphaned projected copy), CLAUDE.md (the `ZHARNESS` managed block only, R14), .claimignore
    surfaces_avoided: cli/docs/embedded/WORKFLOW.md, cli/internal/application/managed_docs.go, cli/internal/application/preflight.go, docs/** apart from the single `docs/playbooks/git.md` deletion (projected copies are regenerated, never hand-edited)
    waves:
      - wave: 1
        name: scaffold-once class
        tasks:
          - id: p2.w1.t1
            requirement: R7
            do: add a scaffold-once helper to `cli/internal/application/init.go`, shaped like `writeAgentsManagedBlock`'s absent-file branch at lines 59-104 and called from `ScaffoldDocs` after it. It writes `docs/README.md`, `docs/decisions/README.md`, and `docs/decisions/templates/decision.md` from Go string constants only when each is absent, and records no `managed_docs` row. `managed_docs.go` is not modified.
            check: `cd cli && go build ./... && go test ./...` reports `ok` for every package.
          - id: p2.w1.t2
            requirement: R7
            do: add a test to `cli/internal/application/init_test.go` that runs `SyncManagedDocs` with `force=true` against a fixture whose `docs/README.md` holds consumer-authored content, asserting the bytes are unchanged and no `managed_docs` row was inserted.
            check: `cd cli && go test ./internal/application -run ScaffoldOnce -v` prints `PASS`.
          - id: p2.w1.t3
            requirement: R14
            definition_added: 2026-08-19, before any run or trace existed for this phase, per decision `01M0CD04RKSXK91MR5M5S7EA7X`. This task did not exist in the original `to-plan` output.
            do: generalize `writeAgentsManagedBlock` in `cli/internal/application/init.go` to take a target path and a body, then call it twice from `ScaffoldDocs` — once for `AGENTS.md` with the embedded body as today, once for `CLAUDE.md` with the single line `@AGENTS.md`. Keep all three existing branches (absent → create; present without markers → prepend, preserving human text; present with markers → replace in place). Do not write the 675-byte block into `CLAUDE.md`, and do not create a symlink.
            check: `cd cli && go test ./internal/application -run ClaudeMdImport -v` prints `PASS`; in a temp repo whose `CLAUDE.md` already holds human content, `zharness init` leaves that text byte-identical and adds the marker block, and running `zharness init` a second time leaves the file byte-identical to the first run.
          - id: p2.w1.t4
            requirement: R14
            definition_added: 2026-08-19, per decision `01M0CD04RKSXK91MR5M5S7EA7X`.
            do: prove the import actually loads and is not eaten by Claude Code's block-level HTML-comment stripping. If it is stripped, move the `@AGENTS.md` line above the `ZHARNESS:BEGIN` marker and keep only commentary inside the markers, then re-prove.
            check: in a fresh Claude Code session started in a temp repo scaffolded by the new binary, `/context` lists `AGENTS.md` under **Memory files**. Record the observed `/context` output in the Validation section; a task that cannot show that output is not complete.
      - wave: 2
        name: routing by convention, and paying off the p1 deferral
        tasks:
          - id: p2.w2.t1
            requirement: R8
            definition_amended: 2026-08-19, before any run or trace existed for this phase, per decision `01M0C99PNHZX3A1CRK7EGHQ23T`. The original definition targeted `cli/docs/embedded/WORKFLOW.md`.
            do: rewrite the final clause of `cli/docs/embedded/AGENTS.md` line 5 so it names `docs/README.md` as the repository's authored documentation map and states that its absence is not an error. Leave `cli/docs/embedded/WORKFLOW.md` untouched — a second routing surface is exactly what R8 excludes.
            check: `grep -c 'docs/README.md' cli/docs/embedded/AGENTS.md` prints at least `1`; `git diff --name-only $(git merge-base master HEAD)..HEAD -- cli/docs/embedded/WORKFLOW.md cli/internal/application/preflight.go` produces no output (`cli/internal/interfaces/preflight.go` was dropped from this assertion on 2026-08-19 because R13 now edits its `preflightPlaybooks` map in wave 3; R8's "no new preflight field" is asserted instead by `p2.w3.t2`); and `zharness init --refresh-docs` in a temp repo rewrites the `AGENTS.md` managed block to contain the new clause while leaving text outside the `ZHARNESS:BEGIN`/`END` markers byte-identical.
          - id: p2.w2.t2
            requirement: R12
            do: delete the dead audit citations from `cli/docs/embedded/playbooks/{work,check,watzup,handoff}.md`, keeping each surrounding rationale sentence intact, then remove the three `.claimignore` entries added by `p1.w1.t2`.
            check: `grep -c 'cli/docs/embedded/playbooks' .claimignore` prints `0`, and `bash scripts/verify-doc-links.sh; echo "exit=$?"` prints `exit=0`.
      - wave: 3
        name: deproject git.md
        definition_added: 2026-08-19, before any run or trace existed for this phase, per decision `01M0C9YJV2Y3QVC22E1E9XGXRG`. This wave did not exist in the original `to-plan` output; the release wave moved from 3 to 4 so the release still runs last.
        tasks:
          - id: p2.w3.t1
            requirement: R13
            do: `git mv cli/docs/embedded/playbooks/git.md skills/workflow/git/references/workflow.md`, then retarget both citations — delete the `"git": "docs/playbooks/git.md"` line from `preflightPlaybooks` in `cli/internal/interfaces/preflight.go:30`, and rewrite `skills/workflow/git/SKILL.md:13` so it names `{baseDir}/references/workflow.md` instead of the returned `playbook` path. Add `references/workflow.md` to that skill's existing `<references>` list. Trash the orphaned projected copy `docs/playbooks/git.md`. Do not add a prune path to `managed_docs.go`.
            check: `find cli/docs/embedded -name git.md` produces no output; `grep -rn 'docs/playbooks/git.md' --include=*.md --include=*.go --include=*.sh .` produces no output; `cd cli && go build ./... && go test ./...` reports `ok` for every package; and `zharness preflight git --json | grep -c '"playbook":""'` prints `1`.
          - id: p2.w3.t2
            requirement: R13, R8
            do: prove the two claims this wave rests on. First, that `git` still has a usable procedure without a returned playbook — `skills/workflow/git/SKILL.md` already carries the full Core Workflow inline and declares preflight never blocks. Second, that an already-initialized consumer repo survives the removal: its stale `managed_docs` row and orphaned `docs/playbooks/git.md` must be inert, not an error.
            check: in a temp repo initialized with the *previous* binary and then re-run with the new one, `zharness init --refresh-docs; echo "exit=$?"` prints `exit=0`, `docs/playbooks/git.md` is still present and byte-unchanged, `find .kit/conflicts -type f | wc -l` prints `0`, and `sqlite3 harness.db "SELECT count(*) FROM managed_docs WHERE path='docs/playbooks/git.md'"` still prints `1` — the row survives untouched and no command fails because of it. Separately, `git diff $(git merge-base master HEAD)..HEAD -- cli/internal/interfaces/preflight.go` shows only the deleted `"git"` map line and no added struct field.
      - wave: 4
        name: integration proof and release
        tasks:
          - id: p2.w4.t1
            requirement: R7
            do: prove idempotence end to end — build the binary, create a temp repo, `git init`, run `zharness init`, edit every scaffolded file, run `zharness init --refresh-docs`, and compare.
            check: the scaffolded files are byte-identical to the edited versions (`diff` produces no output) and `find .kit/conflicts -type f | wc -l` prints `0`.
          - id: p2.w4.t2
            requirement: R9, R3
            do: cut the release by pushing only the `cli/v*` tag; never push the bare `vX.Y.Z` tag, which is created local-only inside CI.
            check: `git ls-remote --tags origin 'cli/v*'` lists the new tag, `git ls-remote --tags origin | grep -c 'refs/tags/v[0-9]'` prints `0`, and `find cli/docs/embedded \( -name README.md -o -name ARCHITECTURE.md -o -name decision.md \)` produces no output.

## Progress
<!-- Append-only durable entries record timestamp, phase, wave, task, task_status, run_id, trace_id, exact verification/result, and changed surfaces or blocker. -->
- none

## Decisions
<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->
- `2026-08-19T05:53:23Z` — Relocate the R8 routing sentence from cli/docs/embedded/WORKFLOW.md to the AGENTS.md managed block, and amend task p2.w2.t1 to match. (phase: `p2-consumer-scaffold`), task: p2.w2.t1. rationale: AGENTS.md line 5 is the defect itself: it orders the agent to read "repository material relevant to the requested outcome" without naming any path, so in a fresh consumer repo that instruction resolves to a blind grep. Fixing that line in place repairs the exact broken instruction instead of adding a second routing surface, and AGENTS.md is read before any stage is selected, so authored project docs load once per session rather than once per stage. Owner decision 2026-08-19. Amending p2.w2.t1 contradicts the immutability rule in to-plan.md line 22; the exception is recorded here rather than applied silently, and is bounded by p2-consumer-scaffold having status planned with zero runs and zero traces at the time of the change..
- `2026-08-19T06:04:47Z` — Remove git.md from the embedded/projected doc set: move it to skills/workflow/git/references/workflow.md, drop the git entry from preflightPlaybooks, and add the work to p2-consumer-scaffold as a new wave 3 ahead of the release wave. (phase: `p2-consumer-scaffold`), task: p2.w3. rationale: git.md fails the embedded-membership test. cli/internal/interfaces/preflight.go:39 already records that git owns no harness entity (D7) and git is absent from contextEligibleStages; cli/docs/embedded/WORKFLOW.md already states git keeps its skill-local procedure. Yet git.md is the single largest projected playbook (148 lines, 7330 bytes, 13% of the 55115-byte projection) while mentioning zharness only 4 times: its headings are Security check, Split decision, Commit, Push, PR, Merge, Anti-Patterns - repo-specific git policy that no two consumer repos would want byte-identical. Verified safe: only 2 references exist repo-wide (preflight.go:30 and skills/workflow/git/SKILL.md:13); git SKILL.md carries its full Core Workflow inline and declares preflight never blocks; and managed_docs.go:108 walks the embedded FS, so an orphaned managed_docs row in an already-initialized consumer repo is never visited, never errors, and needs no prune path. Amending p2 contradicts the immutability rule at to-plan.md:22; the exception is recorded here rather than applied silently and is bounded by p2-consumer-scaffold having status planned with zero runs and zero traces at the time of the change. Owner decision 2026-08-19..
- `2026-08-19T06:58:04Z` — Teach zharness init to ensure CLAUDE.md carries an @AGENTS.md import inside the ZHARNESS managed markers, rather than leaving the harness contract in a file Claude Code never reads. Added to p2-consumer-scaffold wave 1 as R14. (phase: `p2-consumer-scaffold`), task: p2.w1.t3. rationale: Verified against Anthropic's own documentation (code.claude.com/docs/en/memory): 'Claude Code reads CLAUDE.md, not AGENTS.md.' Native AGENTS.md support is issue #34235, still open. This overturns an intermediate conclusion drawn from the binary string 'Claude Code hardcodes CLAUDE.md / AGENTS.md discovery', which belongs to the Codex /import path (a one-time copy), not ambient loading. Consequence: zharness init currently writes its entire managed block into AGENTS.md, a file Claude Code never loads in any configuration - so the harness contract reaches Codex, Cursor and Gemini but never reaches the primary consumer. The docs prescribe the fix directly: an @AGENTS.md import line in CLAUDE.md, expanded into context at launch. That satisfies the owner requirement that both files carry identical content without duplicating a single byte, since it is literally the same file. The symlink alternative the docs also list is rejected: it would clobber an existing project CLAUDE.md and needs Administrator rights on Windows. One risk must be proven rather than assumed - the docs state block-level HTML comments in CLAUDE.md are stripped before injection, and the ZHARNESS markers are HTML comments, so the task must confirm via /context that AGENTS.md appears under Memory files, and fall back to placing the import outside the markers if stripping swallows it. Amending p2 contradicts the immutability rule at to-plan.md:22; the exception is recorded here rather than applied silently and is bounded by p2-consumer-scaffold having status planned with zero runs and zero traces at the time of the change. Owner decision 2026-08-19..

## Validation
<!-- Append-only durable entries record timestamp, phase, exact command/result/output, run_id, check_id, verdict, and proof_gaps. -->
- none

## Current State and Next Action
- active_phase: p1-doc-authority
- lifecycle_status: planned
- latest_run_id: none
- latest_trace_ids: []
- latest_check_id: none
- latest_handoff_id: none
- blockers: none
- open_items: none
- exact_next_action: work full p1-doc-authority
