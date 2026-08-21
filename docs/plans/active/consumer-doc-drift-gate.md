---
id: 01M0FAVR17J6RGK6S40HRNA7F3
type: plan
intake_id: 01M0FAY5EZSM3CPBCAXF6RTCQB
lane: normal
status: active
created: 2026-08-20
updated: 2026-08-21
---

# Plan: consumer-repo documentation — guard the authored half, generate nothing

## Outcome
- result: authored documentation in a consumer repository gains an integrity signal of the kind managed documentation already has. First the repository's own citation graph is made true and the link gate is made able to see it; then `zharness audit` can tell an agent whether the repository still has authored documentation at all, and whether the top-level architecture document still describes the code it cites. The harness continues to generate no documentation content.
- success_signals:
  - `bash scripts/verify-doc-links.sh` **fails** on the current tree once its two blind spots are closed, and **passes** after the citation repair. A gate that cannot fail on today's tree is not evidence.
  - `docs/audit/workflow-harness-ceremony-audit.md`, `docs/audit/sdlc-token-cache-audit.md`, and `docs/audit/sdlc-gap-analysis.md` exist again, restored verbatim from `655c6ac^`, so the requirements and `.claimignore` reasons that cite them as authority have backing.
  - No file under `cli/docs/` or `docs/` cites a path that does not exist, except where `.claimignore` names a justified reason.
  - The embedded playbooks cite no repository-specific dated record, because a consumer repository does not contain one.
  - `zharness audit` reports a finding when a repository initialized by the harness has lost its authored documentation — the signature of `655c6ac`, which today produces a fully green repository.
  - `zharness audit` reports a finding when a pinned authored document's cited source files have moved past the commit it was pinned to, naming the document, the citations that moved, and the size of the change.
  - A file under `docs/plans/`, `docs/decisions/`, or `docs/audit/` never produces a drift finding, no matter how far the code has moved, because those are point-in-time records.
  - A repository whose authored document carries no pin produces no drift finding and no error — pinning is opt-in, and its absence is not a defect.
  - `zharness init` creates no generated prose. The one new scaffold-once file is an unanswered question form, and `zharness audit` reports it as unanswered rather than counting it as documentation.
  - `zharness db rebuild --yes` followed by `zharness query phases --json` lists every phase slug this plan declares, and loses no story that existed before the rebuild. A row the committed markdown cannot regenerate is the one failure ADR 0001 does not tolerate.
  - `cd cli && go test ./...` passes.
  - Each phase merges and stays useful even if no later phase ships.

## Authority and Requirements
- authority:
  - `docs/research/agent-documentation-evidence.md` — external evidence gathered 2026-08-20. F1: LLM-generated context files cost 0.5–2% task success and over 20% inference cost on repositories that already carry documentation, and only help (+2.7%) once every `.md` and `docs/` is deleted. F2: on-demand wiki retrieval is null on correctness and saves cache tokens only. F3: documentation beats code-only by +0.071 on rationale questions and by +0.007 — effectively zero — on anything derivable from code. F5: four independent products converge on the same four-piece drift mechanism. F6: DeepWiki's free tier is public-repository only; codesight's AST support is TypeScript/JavaScript only; CodeWiki has no staleness answer.
  - Fuchsia RFC guidance, cited in session 2026-08-20: "Don't link to an RFC in code or documentation to explain how a feature works." A durable document that explains current behavior by citing a frozen dated record inverts the dependency and rots when the record moves. This is the direct authority for R11–R14.
  - Measured on the working tree, 2026-08-20, verified by `grep -rno "docs/audit/[a-z0-9-]*\.md"`: **74 citations across the repository name three files that do not exist** — `docs/audit/workflow-harness-ceremony-audit.md` (45), `docs/audit/sdlc-token-cache-audit.md` (21), `docs/audit/sdlc-gap-analysis.md` (8). `git log --diff-filter=D` confirms all three were deleted by `655c6ac` and never restored; `git show 655c6ac^:` recovers them intact at 399, 212, and 126 lines. `cli/docs/CONTRACT.md` holds 17 of the 74 and `cli/docs/SCHEMA.md` holds 2; roughly 35 more are Go source comments. `.claimignore` cites two of the three deleted files as the justification for six of its own exemption entries.
  - Measured on the working tree, 2026-08-20: `cli/docs/CONTRACT.md` lines 23, 61, 75 cite `docs/plans/active/harness-markdown-truth.md` and line 162 cites `docs/plans/active/retrieval-router.md`; `cli/docs/SCHEMA.md` line 132 cites `docs/plans/active/harness-markdown-truth.md` and line 145 cites `docs/plans/active/durable-memory.md`. All four plans now live under `docs/plans/completed/`. `bash scripts/verify-doc-links.sh` nevertheless prints `doc links OK (0 findings)`.
  - Simulated 2026-08-20 by adding `cli` to the `find` on `scripts/verify-doc-links.sh:53` and running the script unchanged: **exit 1, 9 findings** — `cli/docs/CONTRACT.md` 5, `cli/docs/SCHEMA.md` 3, and `cli/testdata/legacy-kit/planning/ROADMAP.md` 1. Four of the nine are the deleted audit files and disappear when R12 restores them; four are `docs/plans/active/` paths for plans now under `completed/`. The ninth is a Go test fixture: `cli/internal/application/import_test.go:30,35` reads `cli/testdata/legacy-kit/` as a frozen copy of this repository's own legacy `.kit/` layout, so its stale path is the fixture's content and repairing it would falsify the test. `cli/testdata/` must therefore be excluded by category, on the same reasoning as `docs/plans/`, not by a `.claimignore` entry.
  - `scripts/verify-doc-links.sh:53` — `find docs skills rules setup -name '*.md'`. `cli` appears only in `ALLOWED_PREFIXES` (line 18) as a valid link *target*, never in the scanned source set, so no file under `cli/docs/` has ever been checked. Confirmed adversarially 2026-08-20: a deliberately broken path injected into `cli/docs/CONTRACT.md` left the gate reporting `0 findings` (edit reverted; `git status --short` clean). Widening the path regex alone would therefore be cosmetic.
  - `cli/docs/embedded/playbooks/` — `handoff.md:36`, `work.md:33`, and `brainstorm.md:42` each cite `docs/audit/consumer-adoption-audit.md`. That path is authored and appears nowhere in `cli/internal/application/init.go`, so no consumer repository ever contains it. The other 14 matches on `docs/plans/` in the same directory are the generic `{slug}` template placeholder and describe a live mechanism, not a dated record.
  - `docs/plans/completed/docs-architecture.md` — NG4 ("no empty `ARCHITECTURE.md` skeleton is scaffolded into consumer repos") and NG5 ("no auto-generated, AST-derived documentation — the `codesight` model — reviewed and rejected"). NG5 stands unchanged. NG4 is narrowed by R15 below, deliberately and with its falsification condition stated.
  - `docs/decisions/0004-docs-directory-deletion-655c6ac.md` — the incident. Commit `655c6ac`, authored through the GitHub web UI, deleted 26 files and 4,285 lines under `docs/`. Its closing sentence states the unclosed gap: "This ADR records the incident; it does not prevent the next one." The 74 dead citations above are that incident's damage, still live and still unmeasured by every gate.
  - `docs/audit/consumer-adoption-audit.md` D4 — the consumer `CLAUDE.md` measured at 349 lines / 12,676 bytes / ~3,169 tokens resident every turn, roughly half of it restating what `Read` and `Glob` already disclose: "spend the budget on gotchas, not on what the filesystem shows."
  - `docs/ARCHITECTURE.md` — carries 21 distinct `path/to/file.go:NN` citations (verified 2026-08-20). These are already the page-to-source link that F5 names as mechanism piece one; only the pinned baseline is missing.
  - `cli/internal/application/managed_docs.go:43,56,107` — `SyncManagedDocs` already implements three-hash comparison (recorded, on-disk, embedded) with conflict staging under `.kit/conflicts/`. It covers managed docs only; no equivalent exists for authored docs. Verified 2026-08-20: `cli/internal/application/rebuild.go` contains no reference to `managed_docs` at all, so after `db rebuild` the table is empty.
  - `cli/internal/application/audit.go:17-49` — verified 2026-08-20: `Audit(db, cliVersion)` composes `Resume` and `Validate` only. It receives no repository root and touches no filesystem or git, and `AuditReport`'s three-array JSON shape is a documented public contract.
  - `docs/README.md` — the ownership table; every path under `docs/` must appear in it, and an unlisted path is declared a defect in that table.
  - Owner instruction, this session: research the consumer-repository documentation question against outside practice before choosing a direction, then record the result as durable spec rather than as conversation.
- requirements:
  - R1 [accepted]: The CLI generates no documentation *content* — no wiki, no overview, no AST-derived or LLM-derived prose — in this initiative or as a consequence of it. Creating an empty form that only a human can fill is not generation and is permitted by R15. | source: `docs/research/agent-documentation-evidence.md` F1; `docs/plans/completed/docs-architecture.md` NG5
  - R2 [accepted]: `zharness audit` gains a finding that fires when the repository carries the harness's managed documentation on disk and no markdown file under `docs/` lies outside the managed path set. The precondition is computed from the binary's embedded doc set — a compile-time constant — and the filesystem, never from the `managed_docs` table, because `rebuild.go` never populates that table and `harness.db` is gitignored, so a fresh clone plus `db rebuild` would leave the old precondition false on exactly the incident it exists to catch. | source: `docs/decisions/0004-docs-directory-deletion-655c6ac.md`; `cli/internal/application/rebuild.go` (verified 2026-08-20); `docs/research/agent-documentation-evidence.md` F5
  - R3 [accepted]: `zharness audit` gains a finding that fires when an authored document declares a pinned commit and at least one source path it cites has commits after that pin. The finding names the document, each moved citation, and lines added/removed since the pin. | source: `docs/research/agent-documentation-evidence.md` F5 (doccupine's four pieces: link, pinned SHA, change signal, size measurement)
  - R4 [accepted]: Drift checking is scoped by directory, not by judgment. Only top-level `docs/*.md` is eligible; `docs/plans/`, `docs/decisions/`, and `docs/audit/` are exempt by construction because rewriting a point-in-time record to match today's code falsifies the record. | source: `docs/research/agent-documentation-evidence.md` F5 (freemansoft's directory rule); `docs/README.md` ownership table
  - R5 [accepted]: Pinning is opt-in and additive. A document with no pin is not a defect, produces no finding, and produces no error; a repository that never pins anything sees byte-identical `audit` output to today's apart from R2. | source: `docs/plans/completed/docs-architecture.md` NG4
  - R6 [accepted]: The findings are severity-graded and distinguishable in `audit --json`, so a consumer can act on one and ignore the other. | source: `docs/research/agent-documentation-evidence.md` F5 ("grade the severity, or the gate gets tuned out")
  - R7 [accepted]: `audit` stays read-only. It creates no WAL/SHM sidecar, writes no file, and repairs nothing it checks — a drift finding is reported, never auto-fixed and never auto-repinned. | source: `cli/internal/application/audit.go` (`OpenReadOnly`); `docs/research/agent-documentation-evidence.md` F5 (Doc Bridge: "CI should not repair the evidence it is checking")
  - R8 [accepted]: A decision record fixes the boundary between what the harness guards and what it delegates to external tooling, and `docs/README.md` names the ownership class for documentation the harness neither writes nor guards. | source: owner instruction, this session; `docs/README.md`
  - R9 [accepted]: Freshness is reported as freshness, never as correctness. Finding text must not imply a pinned, unmoved document is accurate. | source: `docs/research/agent-documentation-evidence.md` F5 (Doc Bridge: "a fresh index can faithfully encode a bad ownership decision")
  - R10 [accepted]: `scripts/verify-doc-links.sh` scans `cli/` in addition to `docs`, `skills`, `rules`, and `setup`. Falsifiable: after the change and before any citation repair, the gate exits non-zero and names at least one finding inside `cli/docs/`. | source: `scripts/verify-doc-links.sh:53` vs `:18` (verified adversarially 2026-08-20)
  - R11 [accepted]: `scripts/verify-doc-links.sh` reports a repository-relative path written as a markdown link target, not only one written inside backticks. Today it matches the backtick form only. This yields **zero findings on the current tree** — every markdown link in the repository resolves — so it is prophylactic hardening, not a repair, and its falsification is synthetic: inject one broken markdown link, confirm it is reported, revert. Relative targets must resolve against the referencing file's directory, including a leading `../`, which the current first-segment allowlist check would reject outright. | source: `scripts/verify-doc-links.sh:64` (regex is backtick-delimited); simulated over all scanned files 2026-08-20
  - R12 [accepted]: The three audit documents deleted by `655c6ac` are restored verbatim from `655c6ac^` to their original paths under `docs/audit/`, and `docs/README.md`'s ownership table continues to cover them. Restoration is chosen over rewriting because 74 live citations — including six `.claimignore` justification reasons and requirements already accepted in completed plans — name these files as authority; the citations are correct and the files are what went missing. | source: `git log --diff-filter=D` and `git show 655c6ac^:` (verified 2026-08-20); `docs/decisions/0004-docs-directory-deletion-655c6ac.md`
  - R13 [accepted]: No file under `cli/docs/` cites a repository path that does not exist. The six citations naming `docs/plans/active/` paths for plans now under `docs/plans/completed/` are repointed to the governing decision record under `docs/decisions/`, never to `docs/plans/completed/`. A completed plan is a retired record whose lasting result belongs in a decision record; a citation aimed at the completed path is a link scheduled to break a second time. Falsifiable: `bash scripts/verify-doc-links.sh` passes with R10 and R11 in force, and `grep -rn 'docs/plans/completed/' cli/docs/` returns nothing. | source: measured working tree 2026-08-20; Fuchsia RFC guidance (authority above); owner decision this session
  - R14 [accepted]: No file under `cli/docs/embedded/` cites a repository-specific dated record. The three citations of `docs/audit/consumer-adoption-audit.md` in `handoff.md`, `work.md`, and `brainstorm.md` are removed or replaced by the rule they encode, because a consumer repository never contains that file. Generic `docs/plans/active/{slug}.md` template placeholders are not affected — they describe a live mechanism. | source: `cli/docs/embedded/playbooks/` measured 2026-08-20; `cli/internal/application/init.go` (path absent from every scaffold set)
  - R15 [accepted]: `zharness init` scaffolds `docs/ARCHITECTURE.md` once, when absent, containing at most five unanswered questions and no prose describing the repository. The questions cover: the problem and its audience; the main use cases in domain language; domain nouns whose meaning is non-standard; invariants that fail silently; and boundaries that must not be crossed. None is answerable by reading the repository — that is the admission test for each question. `zharness audit` reports the file as unanswered while every question is blank, and never counts it as documentation for R2. Falsification and rollback: if two or three real consumer repositories leave it blank, the entry is deleted from `scaffoldOnceDocs` and no migration is required. | source: `docs/research/agent-documentation-evidence.md` F3 (rationale +0.071 vs code-derivable +0.007) and F1 (generation is net-negative, elicitation is not measured as such); narrows `docs/plans/completed/docs-architecture.md` NG4
  - R16 [accepted]: Before this plan moves to `docs/plans/completed/`, every citation naming its active path by slug is repointed. Confirmed clean at lock time — `docs/README.md:15` and `docs/ARCHITECTURE.md:23` link generically to `plans/active/`, not to this slug — so this is a closing step, not a repair. | source: the six dead citations measured above, which are this exact failure mode one initiative later
  - R17 [accepted]: `zharness db rebuild --yes` reconstructs a phase's story row from committed plan markdown regardless of which phase-block form the plan uses. `planFieldValue` (`cli/internal/application/rebuild.go:151`) matches `^[ \t]*<key>:` only, so a field written as a markdown list item -- `- story_id: 01M0...`, the form this plan and `docs/plans/completed/pr60-review-fixes.md` both use -- never matches, and `rebuildStoriesFromPlan` takes its `continue // malformed/hand-edited phase block` branch and drops the story silently. Measured 2026-08-20: `grep -c '^[[:space:]]*story_id:'` returns 2/2 for `docs/plans/completed/docs-architecture.md` (indented form, rebuilds) and 0/5 for this plan (list form, does not). After `zharness db rebuild --yes` all five of this plan's story IDs are absent, and two stories present beforehand -- `pr60-doc-truth` and `pr60-go-correctness` -- are lost. This is a live violation of `docs/decisions/0001-markdown-is-the-source-of-truth.md`: rows exist that markdown cannot regenerate. Falsifiable: after the fix, `zharness db rebuild --yes` followed by `zharness query phases --json` lists all five phase slugs of this plan, and no story present before the rebuild is missing after it. | source: `cli/internal/application/rebuild.go:150-165`; `cli/internal/application/plan_query.go:37,45`; reproduced against the working tree 2026-08-20
  - R18 [accepted]: `zharness db rebuild --yes` fully reconstructs a phase's story row when the plan uses `docs/plans/completed/pr60-review-fixes.md`'s phase-block form for `pr60-go-correctness` and `pr60-doc-truth` (`:81,118`) -- both the phase heading and the scalar field names it writes. Two independent gaps compose here, and closing only one still drops both stories:
    - Heading discovery: the heading reads "### Phase N — {slug}" (em dash). Neither `planPhaseHeading` nor `planPhaseListItem` (`cli/internal/application/plan_query.go:37,45`) matches it, so `planPhaseSlugs` never yields the slug and `extractPlanPhaseBlock` never returns a block at all -- nothing downstream runs. `docs/plans/completed/pr60-review-fixes.md:171` already documents this as a known, then-out-of-scope gap in `extractPlanPhaseHeadingBlock`.
    - Field-name mismatch: even with the block returned, its fields are `- story: \`{ulid}\`` (backtick-wrapped value) and `- depends on: ...` (space, with trailing prose after the value) at `:82-83,119-120` -- not the `story_id:`/`depends_on:` keys `planFieldValue` (R17, `cli/internal/application/rebuild.go:159`) matches. `rebuildStoriesFromPlan` (`cli/internal/application/rebuild.go:192`) reads `planFieldValue(block, "story_id")`, finds nothing, and takes the same `continue // malformed/hand-edited phase block` branch R17 fixed for the list-item case -- so even a heading-only fix still drops both stories, which is why fixing solely `planPhaseHeading` was rejected as an incomplete scope for this requirement (owner decision this session).
    - Reproduced 2026-08-21, run `01M0HAHCF491420NFW88Y6DMGK`: R17's fix alone regenerated all six of this plan's own phase slugs but not these two, confirming the gap is compound, not a single missed pattern.
    - Falsifiable: after the fix, `zharness db rebuild --yes` followed by `zharness query phases --json` additionally lists `pr60-go-correctness` and `pr60-doc-truth` with correct `story_id` values (backticks stripped), and no story present before the rebuild is missing after it. | source: `docs/plans/completed/pr60-review-fixes.md:81-83,118-120,171`; `cli/internal/application/plan_query.go:37,45,163`; `cli/internal/application/rebuild.go:159,192`; reproduced against the working tree 2026-08-21; owner decision this session to route the gap through `brainstorm refine` rather than widen phase `rebuild-phase-parser`'s scope
  - R19 [accepted]: `docs/playbooks/` (the projected copy `zharness init` scaffolds once per repository) is brought back into byte-identical sync with `cli/docs/embedded/playbooks/` (the source R14 edits), and that sync is achieved without relying on `zharness init` to perform it. `writeScaffoldOnceDocs` (`cli/internal/application/init.go:135-151`) writes a scaffold-once path only when it is absent on disk -- unconditionally, with no `--force` or refresh override -- so `zharness init --json` against a repository that already has `docs/playbooks/` (every existing checkout, including this one) leaves the three files R14 edited stale; `diff -r cli/docs/embedded/playbooks docs/playbooks` still reports the pre-R14 text after `init` runs. This was discovered while executing phase `audit-restore`'s own Wave 2, whose second task specifies exactly `zharness init --json` as the sync command and expects that same `diff -r` to report no differences -- a check the command cannot satisfy by the contract `writeScaffoldOnceDocs` itself documents ("the harness never refreshes, overwrites, or deletes it," `cli/internal/application/init.go:158-159`). Restoring sync therefore requires a hand copy of the three files -- a surface (`docs/playbooks/`) that phase `audit-restore`'s `surfaces_allowed` does not list, since its wave was designed assuming `init` would perform the sync. Falsifiable: after the fix, `diff -r cli/docs/embedded/playbooks docs/playbooks` reports no differences, and `git status --short docs/playbooks/` shows exactly the three files R14 touched as modified. | source: `cli/internal/application/init.go:135-151,158-159`; measured against the working tree 2026-08-21 (`zharness init --json` returned `status: exists`, `docs/playbooks/` unchanged); phase `audit-restore` Wave 2 task 2, this plan

## Non-goals
- NG1: No documentation generator of any kind — not AST-derived, not LLM-derived, not a repository wiki, and no prose written into any file by the CLI. F1 measures generation as net-negative on repositories that already have `docs/`, which is every repository that has run `zharness init`. `docs/plans/completed/docs-architecture.md` NG5 stays in force. R15's question form contains no statement about the repository and is therefore not generation.
- NG2: No scaffolded file beyond the one R15 names. The scaffold-once set goes from three files to four and no further. R15 narrows `docs/plans/completed/docs-architecture.md` NG4 — which forbade an empty *skeleton*, meaning headings awaiting prose — to permit a form of questions with an explicit falsification and rollback clause. Every other file remains uncreated.
- NG3: No dependency on DeepWiki, codesight, CodeWiki, or any external documentation service — no MCP server call, no network access, no vendored parser. F6 rules each out on its own terms.
- NG4: No auto-repair, no auto-repin, and no generated pull request. Reporting only (R7). The "detect, then let a human decide" split is deliberate.
- NG5: No semantic or content comparison between a document and the code it cites. Drift here is a commit-range fact computed from git, not an LLM judgment about whether prose is still true.
- NG6: No change to any spine playbook's *procedure* and no new mandatory step. R14 removes three dead citations from three playbooks without altering a single instruction, which is repair, not redesign. This holds the boundary `docs/decisions/0003-durable-memory-not-wired-into-playbooks.md` set: "every mandatory playbook step is paid on every invocation, forever, by every consumer repo."
- NG7: No edits to any consumer or reference repository. `onedrive-cloud`, `harness-experimental`, and `codesight` are cited as evidence only, per `docs/plans/completed/harness-markdown-truth.md` NG1.
- NG8: No structural-forensics analysis (hub files, co-change coupling, generated-file detection). F4 identifies it as the one place automation beats a human brief; it is recorded in the evidence page for a future initiative.
- NG9: No CI workflow, no git hook, and no blocking gate in this initiative. `audit` reports; wiring a report into a blocking check is a separate decision with its own cost.
- NG10: No seeding of `docs/memory/` and no playbook step that writes a memory. The store is complete and deliberately unwired per `docs/decisions/0003-durable-memory-not-wired-into-playbooks.md`; wiring it is a separate initiative that must argue its own per-invocation cost.
- NG11: No repair of the roughly 35 dead `docs/audit/` citations that live in Go source comments rather than markdown. R12's restoration makes them true again as a side effect; if any remain wrong afterwards they are out of scope, because `verify-doc-links.sh` checks markdown and widening it to Go source is a separate decision.

## Approach and Risks
- approach: repair the citation graph first, then teach the gate to see it, then extend `zharness audit` with two filesystem-grounded findings, and only last add the one scaffold-once question form. Repair precedes detection because a gate widened over a broken tree fails immediately and would have to be merged red; restoring the three deleted audit documents removes four of the nine findings before the gate ever runs. The two new audit findings are additive to `AuditReport`'s existing three-array JSON shape and are reached by giving `Audit` a repository root it does not have today. Every phase leaves the repository green.
- constraints:
  - `AuditReport`'s three-array JSON shape is a documented public contract (`cli/docs/CONTRACT.md`); findings are added within it, never by reshaping it.
  - `Audit(db, cliVersion)` takes no repository root and touches no filesystem or git. Phases 3 and 4 change that signature, so every caller and test must move with it.
  - `audit` opens the database read-only and must stay side-effect free (R7): no WAL/SHM sidecar, no file write, no repin.
  - Scaffold-once semantics mean `writeScaffoldOnceDocs` skips any existing path. `docs/ARCHITECTURE.md` already exists in this repository, so R15 cannot be observed here by running `init` — it is verifiable only in a temporary directory or a second consumer repository.
  - `cli/testdata/legacy-kit/` is a Go test fixture asserted against by `cli/internal/application/import_test.go`. Its contents are frozen input, not documentation, and must never be repaired to satisfy a gate.
  - Restored audit documents are historical records. They are restored verbatim from `655c6ac^` and not edited to match today's code, on the same reasoning as R4.
- dependencies:
  - Phase `link-gate-widening` depends on `audit-restore`: restoration removes four of the nine findings the widened gate would otherwise report, so repointing work shrinks from nine to four.
  - Phases `authored-docs-guard`, `pin-drift-finding`, and `architecture-elicitation` depend on the repository being green under the widened gate, because each adds documentation prose or a scaffolded file that the gate will then check.
  - `pin-drift-finding` and `architecture-elicitation` both depend on `authored-docs-guard` because it is the phase that gives `Audit` a repository root; neither depends on the other.
- rejected_alternatives:
  - Rewrite the 74 dead citations instead of restoring the three files. Rejected: the citations are correct and the files are what went missing. Six `.claimignore` reasons and requirements already accepted in completed plans name these documents as authority, so rewriting would erase the authority rather than restore it, at roughly twelve times the edit cost.
  - Silence the nine findings with `.claimignore` entries. Rejected: `.claimignore` line 4 states "Do not add an entry to silence a real broken link - fix the link." Eight of the nine are real; the ninth is a category exclusion, not an exception.
  - Widen the gate first and repair second. Rejected: it requires merging a phase that leaves `check` red, violating phase independence.
  - Give `audit` an LLM or semantic comparison between prose and code. Rejected by NG5; drift here is a commit-range fact computed from git.
- risks:
  - Changing `Audit`'s signature ripples into `preflight`, `db_status`, and their tests. Mitigation: phase 3 is scoped to that single signature change plus one finding, so the ripple is reviewed on its own. Stop condition: if the caller set exceeds what one phase can verify, split the signature change into its own wave before adding the finding.
  - The R2 precondition could fire on a repository that legitimately has no `docs/` — a consumer that ran `init` and deleted docs on purpose. Mitigation: R6 grades severity so it can be read as a warning, and R9 forbids wording it as a correctness claim. Recovery: if a real consumer reports a false positive, the precondition narrows to "managed docs present on disk AND `docs/` contains zero markdown files", which is the `655c6ac` signature exactly.
  - R15 may produce a file every consumer leaves blank. Mitigation: R15 carries its own falsification and rollback — after two or three real consumer repositories, delete the entry from `scaffoldOnceDocs`; no migration is needed because scaffold-once files are consumer-owned from creation.
  - Restoring 737 lines of historical audit text re-exposes those documents to the widened gate, which may find dead citations inside them. Mitigation: they are dated records under `docs/audit/`, already excluded from drift by R4; if they carry broken links, `.claimignore` entries naming the record-immutability reason are the correct response, not edits.
- recovery: every phase is a normal revert. No migration, no schema change, and no on-disk format change is introduced by any phase. `harness.db` is gitignored and rebuildable via `zharness db rebuild --yes` at any point.

## Phases and Verification
<!-- Phase and task definitions are immutable after to-plan. Do not add task status fields. Append-only Progress is the sole task execution-status source. Only each phase lifecycle status changes to mirror DB transitions: to-plan=planned; work after run create=in-progress; clean durable check=checked; closing handoff=done. Each planned phase records phase_slug, story_id, status, goal, depends_on, waves, tasks, and checks. -->
- planning_status: planned
- phases:

### phase_slug: `rebuild-phase-parser`
- story_id: 01M0H8NGCWZ28H6KDSGSAT0V8D
- status: checked
- goal: make `db rebuild` reconstruct a story row from every phase-block form the repository's plans actually use, so the database stays regenerable from committed markdown as ADR 0001 requires.
- depends_on: none
- requirements: R17
- surfaces_allowed: `cli/internal/application/rebuild.go`, `cli/internal/application/rebuild_test.go`
- surfaces_avoided: `docs/`, `scripts/`, `cli/internal/application/plan_query.go`, `cli/internal/application/plan_lifecycle.go`
- waves:
  - **Wave 1 — reproduce the loss**
    - task: pin the defect with a failing test before touching the parser, using the list-item form this plan uses.
      - surfaces: `cli/internal/application/rebuild_test.go`
      - expected_output: a test that rebuilds a plan whose phase block writes `- story_id: <ulid>` and asserts the story row exists; it fails on the current parser.
      - check: `cd cli && go test ./internal/application/ -run Rebuild` fails, and the failure names the missing story row rather than a panic or a compile error.
      - stop_condition: the new test passes before the fix. That would mean the defect is elsewhere; stop and re-diagnose rather than editing `planFieldValue`.
  - **Wave 2 — widen the field matcher**
    - task: allow an optional markdown list marker before the key in `planFieldValue`, changing nothing else about the match.
      - surfaces: `cli/internal/application/rebuild.go:150-156`
      - expected_output: the regex accepts `story_id:`, `  story_id:`, and `- story_id:` and still anchors to line start; no other function is edited.
      - check: `cd cli && go test ./...` passes, including the Wave 1 test.
      - check: `git diff --stat cli/internal/application/rebuild.go` shows a single-line change to the regex.
  - **Wave 3 — prove the database is regenerable**
    - task: rebuild against the real repository and confirm no story is lost and every declared phase returns.
      - command: `zharness db rebuild --yes --json`
      - check: `zharness query phases --json` lists `rebuild-phase-parser`, `audit-restore`, `link-gate-widening`, `authored-docs-guard`, `pin-drift-finding`, and `architecture-elicitation`.
      - check: `zharness query phases --json` still lists `pr60-doc-truth` and `pr60-go-correctness`, which the current parser drops.
      - stop_condition: any story present before the rebuild is absent after it. Do not proceed to `audit-restore` with a lossy rebuild.
- checks:
  - `cd cli && go test ./...` passes.
  - `bash scripts/verify-doc-links.sh` passes (this phase touches no markdown).
  - `zharness db rebuild --yes --json` followed by `zharness query phases --json` shows a story count no lower than before the rebuild.
  - `zharness check record` verdict is `APPROVED` or `APPROVE_WITH_REQUESTS`.
- escalation: if a plan form still fails to rebuild after the regex widens, the gap is in `extractPlanPhaseBlock` (`cli/internal/application/plan_query.go:163`) rather than in field matching. That file is on the avoided list for this phase — stop and route to `brainstorm refine` instead of widening scope.

### phase_slug: `legacy-phase-form`
- story_id: 01M0HHQPDMEWEF9ZET6649GV28
- status: checked
- goal: make `db rebuild` fully regenerate `pr60-doc-truth` and `pr60-go-correctness` from `docs/plans/completed/pr60-review-fixes.md`'s phase-block form — a compound gap (heading discovery, then field-name matching) that `rebuild-phase-parser`'s list-item fix left closed only for the shapes this plan itself uses.
- depends_on: rebuild-phase-parser
- requirements: R18
- surfaces_allowed: `cli/internal/application/plan_query.go`, `cli/internal/application/plan_query_test.go`, `cli/internal/application/rebuild.go`, `cli/internal/application/rebuild_test.go`
- surfaces_avoided: `cli/internal/application/plan_phase_status.go`, `cli/internal/application/plan_phase_status_test.go` (also read `planPhaseHeading`; must keep passing unedited — a widened shared regex, not a rewritten one, is what proves this), `docs/plans/completed/**` (immutable historical record), `scripts/`
- waves:
  - **Wave 1 — reproduce both gaps as failing tests**
    - task: pin the heading-discovery gap.
      - surfaces: `cli/internal/application/plan_query_test.go`
      - expected_output: a test fixture using `### Phase 1 — \`{slug}\`` (em dash) and asserting `planPhaseSlugs`/`extractPlanPhaseHeadingBlock` finds the slug; it fails on the current parser.
      - check: `cd cli && go test ./internal/application/ -run PlanPhase -v` fails, naming the missing slug rather than a panic or compile error.
      - stop_condition: the new test passes unfixed. That means the gap is already closed elsewhere; stop and re-diagnose before touching `planPhaseHeading`.
    - task: pin the field-name gap.
      - surfaces: `cli/internal/application/rebuild_test.go`
      - expected_output: a test fixture reproducing `pr60-review-fixes.md`'s exact block shape (`- story: \`{ulid}\`` and `- depends on: \`{slug}\` — {prose}`, reachable via a heading `planPhaseHeading` already matches, so this test isolates the field gap from the heading gap) and asserting the rebuilt story row exists with `story_id` equal to the bare ulid (backticks stripped); it fails on the current parser.
      - check: `cd cli && go test ./internal/application/ -run RebuildFromMarkdown -v` fails, naming the missing/malformed story rather than a panic or compile error.
      - stop_condition: the new test passes unfixed; stop and re-diagnose before touching `planFieldValue`.
  - **Wave 2 — fix, gated on Wave 1 being red**
    - task: widen `planPhaseHeading` to also match `### Phase {N} — \`{slug}\`` (em dash), changing nothing else about heading matching or block-boundary logic.
      - surfaces: `cli/internal/application/plan_query.go:37`
      - expected_output: the shared regex accepts both the existing `### phase_slug: \`{slug}\`` form and the new em-dash form; `extractPlanPhaseHeadingBlock`'s boundary logic is untouched (it already treats "next match of `planPhaseHeading`" as the sibling boundary regardless of which form matched).
      - check: `cd cli && go test ./internal/application/ -run PlanPhase -v` passes (Wave 1's test); `go test ./internal/application/ -run TestSetPlanPhaseStatus` still passes unedited (proves the active-plan-only consumer is unaffected).
    - task: add legacy field-name aliasing and backtick-value stripping to `planFieldValue`.
      - surfaces: `cli/internal/application/rebuild.go:159-166`
      - expected_output: `story_id` also matches a `story:` key, `depends_on` also matches a `depends on:` key, and a captured value that starts with a backtick is trimmed to the text between its first backtick pair (so `` `01M0EG9Z...` `` → `01M0EG9Z...`, and `` `pr60-go-correctness` — R10 verifies... `` → `pr60-go-correctness`); every existing caller and key (`status`, `goal`) is unaffected since neither the alias map nor the backtick rule fires for them.
      - check: `cd cli && go test ./internal/application/ -run RebuildFromMarkdown -v` passes (Wave 1's test), and `TestRebuildFromMarkdownRoundTrip`, `TestRebuildFromMarkdownDropsUncheckedRun`, and `TestRebuildFromMarkdownListItemStoryFields` all still pass unedited.
  - **Wave 3 — prove the real repository regenerates losslessly**
    - task: rebuild against the real repository and confirm both stories return with correct IDs and nothing else is lost.
      - command: `zharness db rebuild --yes --json`
      - check: `zharness query phases --json` lists `pr60-go-correctness` with story id `01M0EG9ZS5XTQJV3J5J2CZ689B` and `pr60-doc-truth` with story id `01M0EG9ZSERMRZVPE860XRCJPR` (the ulids in `docs/plans/completed/pr60-review-fixes.md:82,119`, backticks stripped).
      - check: `zharness query phases --json` still lists every phase slug of this plan (`rebuild-phase-parser`, `legacy-phase-form`, `audit-restore`, `link-gate-widening`, `authored-docs-guard`, `pin-drift-finding`, `architecture-elicitation`).
      - stop_condition: any story present before the rebuild is absent after it. Do not proceed to `audit-restore` with a lossy rebuild.
- checks:
  - `cd cli && go test ./...` passes.
  - `bash scripts/verify-doc-links.sh` passes (this phase touches no markdown other than this plan).
  - `zharness db rebuild --yes --json` followed by `zharness query phases --json` shows a story count no lower than before the rebuild, including both pr60 stories.
  - `zharness check record` verdict is `APPROVED` or `APPROVE_WITH_REQUESTS`.
- escalation: if either wave's fix causes an existing, unrelated test to require modification, stop and report rather than editing that test — the same rule `pr60-review-fixes.md`'s own Wave 2 used for this exact function family.

### phase_slug: `audit-restore`
- story_id: 01M0FTV4RHEWQG12ND15PACVGW
- status: checked
- goal: make the repository's 74 citations of three deleted audit documents true again, and stop the embedded playbooks citing a path no consumer repository contains.
- depends_on: none
- requirements: R12, R14
- surfaces_allowed: `docs/audit/`, `cli/docs/embedded/playbooks/`, `docs/README.md`
- surfaces_avoided: `scripts/`, `cli/internal/`, `docs/plans/`
- waves:
  - **Wave 1 — restore the records**
    - task: restore the three documents verbatim from the commit before the deletion.
      - command: `for f in workflow-harness-ceremony-audit sdlc-token-cache-audit sdlc-gap-analysis; do git show 655c6ac^:docs/audit/$f.md > docs/audit/$f.md; done`
      - expected_output: three files exist at 399, 212, and 126 lines.
      - check: `wc -l docs/audit/workflow-harness-ceremony-audit.md docs/audit/sdlc-token-cache-audit.md docs/audit/sdlc-gap-analysis.md` prints 399, 212, 126.
      - check: `for f in workflow-harness-ceremony-audit sdlc-token-cache-audit sdlc-gap-analysis; do diff <(git show 655c6ac^:docs/audit/$f.md) docs/audit/$f.md; done` prints nothing — byte-identical, no edits.
      - stop_condition: any of the three fails to extract from `655c6ac^`. Escalate to the owner rather than reconstructing prose.
    - task: confirm `docs/README.md`'s ownership table covers `docs/audit/`, since an unlisted path under `docs/` is declared a defect in that table.
      - check: `grep -n 'docs/audit' docs/README.md` returns at least one row; if absent, add the row naming the authored ownership class.
  - **Wave 2 — deproject the playbook citations**
    - task: remove or replace the `docs/audit/consumer-adoption-audit.md` citation in each of the three embedded playbooks, keeping the rule it encodes and changing no instruction.
      - surfaces: `cli/docs/embedded/playbooks/handoff.md:36`, `cli/docs/embedded/playbooks/work.md:33`, `cli/docs/embedded/playbooks/brainstorm.md:42`
      - expected_output: zero occurrences in `cli/docs/embedded/`; each sentence still states the same constraint without naming a repository-specific record.
      - check: `grep -rn 'consumer-adoption-audit' cli/docs/embedded/` returns nothing.
      - check: `git diff --stat cli/docs/embedded/playbooks/` shows only the three files, and `git diff cli/docs/embedded/playbooks/` shows no added or removed numbered step.
    - task: regenerate the projected copies under `docs/playbooks/` so the scaffolded tree matches the embedded source.
      - command: `zharness init --json`
      - check: `diff -r cli/docs/embedded/playbooks docs/playbooks` reports no differences for the three edited files.
      - check: `git status --short docs/playbooks/` shows exactly the three expected files as modified.
- checks:
  - `cd cli && go test ./...` passes.
  - `bash scripts/verify-doc-links.sh` still passes (this phase does not widen the gate).
  - `grep -rno 'docs/audit/[a-z0-9-]*\.md' . 2>/dev/null | grep -v '^./.git' | awk -F'docs/audit/' '{print $2}' | sort | uniq -c` shows every named file now exists on disk.
  - `zharness check record` verdict is `APPROVED` or `APPROVE_WITH_REQUESTS`.
- escalation: if the embedded playbook edit cannot be made without changing a step, stop and route to `brainstorm refine` — NG6 forbids procedural change.

### phase_slug: `playbook-scaffold-sync`
- story_id: 01M0HM0KJFXZEXN67KTA90CY9Y
- status: checked
- goal: sync `docs/playbooks/` (the projected scaffold-once copy) to match `cli/docs/embedded/playbooks/` (the source `audit-restore` Wave 2 edits for R14), since `zharness init` cannot perform this sync for an existing checkout.
- depends_on: audit-restore
- requirements: R19
- surfaces_allowed: `docs/playbooks/`
- surfaces_avoided: `cli/docs/embedded/`, `docs/audit/`, `cli/internal/`, `scripts/`, `docs/plans/`
- waves:
  - **Wave 1 — hand-copy the projected files**
    - task: copy the three files `audit-restore` edited from the embedded source to the projected copy, byte-for-byte.
      - command: `cp cli/docs/embedded/playbooks/handoff.md cli/docs/embedded/playbooks/work.md cli/docs/embedded/playbooks/brainstorm.md docs/playbooks/`
      - expected_output: the three files under `docs/playbooks/` are byte-identical to their `cli/docs/embedded/playbooks/` counterparts.
      - check: `diff -r cli/docs/embedded/playbooks docs/playbooks` reports no differences.
      - check: `git status --short docs/playbooks/` shows exactly `handoff.md`, `work.md`, and `brainstorm.md` as modified, nothing else.
      - stop_condition: `cli/docs/embedded/playbooks/` and `docs/playbooks/` diverge on any file outside these three (a sign the embedded source moved since `audit-restore` ran). Escalate rather than overwrite an unrelated file.
- checks:
  - `diff -r cli/docs/embedded/playbooks docs/playbooks` reports no differences.
  - `bash scripts/verify-doc-links.sh` still passes.
  - `zharness check record` verdict is `APPROVED` or `APPROVE_WITH_REQUESTS`.
- escalation: if `cli/docs/embedded/playbooks/` and `docs/playbooks/` differ on a file this phase did not expect to touch, stop and route to `brainstorm refine` before overwriting it.

### phase_slug: `link-gate-widening`
- story_id: 01M0FTV4RVKEC0FZWWNPN6JHNV
- status: checked
- goal: make `verify-doc-links.sh` able to see `cli/` and markdown-link targets, exclude the Go test fixture by category, and repair the citations the widened gate then reports.
- depends_on: audit-restore
- requirements: R10, R11, R13
- surfaces_allowed: `scripts/verify-doc-links.sh`, `cli/docs/CONTRACT.md`, `cli/docs/SCHEMA.md`, `.claimignore`
- surfaces_avoided: `cli/internal/`, `cli/testdata/`, `docs/audit/`
- waves:
  - **Wave 1 — widen the source set and prove it fails**
    - task: add `cli` to the `find` on `scripts/verify-doc-links.sh:53` and exclude `cli/testdata/*` by category, with a comment stating the reason in the same style as the existing `docs/plans/**` comment above it.
      - expected_output: the scanned set gains `cli/docs/` and excludes the frozen fixture.
      - check: `bash scripts/verify-doc-links.sh; echo $?` exits **1** and names at least one path under `cli/docs/`. A pass here means the change did not take effect.
      - check: the output contains no path under `cli/testdata/`.
      - stop_condition: findings appear in `cli/` outside `cli/docs/`. Widen the exclusion by category rather than adding `.claimignore` entries.
  - **Wave 2 — repoint the remaining dead citations**
    - task: repoint the four `docs/plans/active/` citations naming plans now under `docs/plans/completed/`, to the completed path or to the governing decision record.
      - surfaces: `cli/docs/CONTRACT.md` lines 23, 61, 75 (`harness-markdown-truth.md`) and 162 (`retrieval-router.md`); `cli/docs/SCHEMA.md` lines 132 (`harness-markdown-truth.md`) and 145 (`durable-memory.md`)
      - expected_output: every cited path resolves; no requirement identifier (R9, R1/R4, P3, P6) is dropped in the rewrite.
      - check: `bash scripts/verify-doc-links.sh` exits 0 and prints `doc links OK (0 findings)`.
      - check: `grep -c 'docs/plans/active/[a-z-]*\.md' cli/docs/CONTRACT.md cli/docs/SCHEMA.md` counts only paths containing `{slug}` or `*`, which are live templates, not records.
  - **Wave 3 — markdown-link coverage**
    - task: extend the claim regex to markdown-link targets in addition to backtick-quoted paths, resolving a relative target against the referencing file's directory including a leading `../`, which the current first-segment allowlist rejects outright.
      - expected_output: zero new findings on the current tree — every markdown link in the repository already resolves — so this wave must not change the exit code.
      - check: `bash scripts/verify-doc-links.sh` still exits 0.
      - check (synthetic, the only falsification available): append `[x](docs/nonexistent-probe.md)` to `docs/README.md`, confirm the gate exits 1 naming that path, then `git checkout docs/README.md` and confirm `git status --short docs/README.md` is empty.
      - check (synthetic, relative form): repeat the probe as `[x](../cli/docs/nonexistent-probe.md)` from a file inside `docs/`, confirm it is reported, then revert and confirm a clean tree.
- checks:
  - `bash scripts/verify-doc-links.sh` exits 0 on the final tree.
  - `cd cli && go test ./...` passes, including `import_test.go`, proving the fixture was not edited.
  - `git diff --stat cli/testdata/` is empty.
  - `zharness check record` verdict is `APPROVED` or `APPROVE_WITH_REQUESTS`.
- escalation: if repointing a citation requires deciding what the sentence should now claim rather than where it points, that is a content decision — stop and route to `brainstorm refine`.

### phase_slug: `authored-docs-guard`
- story_id: 01M0FTV4S4RBAK9D1K3QQQ9161
- status: in-progress
- goal: `zharness audit` reports a finding when a repository carrying the harness's managed documentation has no authored documentation left — the `655c6ac` signature, which today produces a fully green repository.
- depends_on: link-gate-widening
- requirements: R2, R6, R7, R8, R9
- surfaces_allowed: `cli/internal/application/audit.go`, its callers and tests, `cli/docs/CONTRACT.md`, `docs/decisions/`, `docs/README.md`
- surfaces_avoided: `cli/internal/application/init.go`, `scripts/`, `docs/audit/`
- waves:
  - **Wave 1 — give `Audit` a repository root**
    - task: change `Audit(db, cliVersion)` to accept a repository root and update every caller and test.
      - expected_output: signature change only; no behavior change and no new finding yet.
      - check: `cd cli && go build ./... && go test ./...` passes.
      - check: `zharness audit --json` output is byte-identical to a copy captured before the change.
      - check: `zharness audit --json` creates no sidecar — `ls harness.db-wal harness.db-shm` reports both absent (R7).
      - stop_condition: the caller set is larger than this wave can verify. Split rather than proceeding into wave 2 with a red build.
  - **Wave 2 — the finding**
    - task: add the finding, with its precondition computed from the binary's embedded doc set and the filesystem, never from the `managed_docs` table.
      - expected_output: fires when managed docs are present on disk and `docs/` holds no markdown outside the managed path set; silent otherwise.
      - check: in a temporary directory, `zharness init` then delete `docs/` — `zharness audit --json` reports the finding. Then restore and confirm it clears.
      - check: `grep -n 'managed_docs' cli/internal/application/audit.go` returns nothing — the precondition must survive a fresh clone plus `db rebuild`, which leaves that table empty.
      - check: on a fresh clone of this repository plus `zharness db rebuild --yes`, `zharness audit --json` still reports correctly.
      - check: the finding is severity-graded and distinguishable in `--json` (R6), and its text states freshness or presence, never correctness (R9).
      - check: `AuditReport`'s three top-level arrays are unchanged in name and shape (`cli/docs/CONTRACT.md` public contract).
  - **Wave 3 — record the boundary**
    - task: write the decision record fixing what the harness guards versus what it delegates, and name the ownership class in `docs/README.md` for documentation the harness neither writes nor guards (R8).
      - check: the new record exists under `docs/decisions/` following `docs/decisions/templates/decision.md`.
      - check: `bash scripts/verify-doc-links.sh` exits 0.
      - check: `docs/README.md`'s ownership table has no path under `docs/` missing from it.
- checks:
  - `cd cli && go test ./...` passes.
  - `bash scripts/verify-doc-links.sh` exits 0.
  - `zharness audit --json` on this repository reports no new finding, because this repository has authored docs.
  - `zharness check record` verdict is `APPROVED` or `APPROVE_WITH_REQUESTS`.
- escalation: if the precondition cannot be computed without reading `managed_docs`, stop — R2 exists precisely to forbid that, and a workaround needs a refine, not a judgment call.

### phase_slug: `pin-drift-finding`
- story_id: 01M0FTV8M4ERF0H3SKDBED05GW
- status: planned
- goal: an authored document may declare a pinned commit, and `zharness audit` reports when source paths it cites have moved past that pin, naming the document, the moved citations, and the size of the change.
- depends_on: authored-docs-guard
- requirements: R3, R4, R5, R6, R7, R9
- surfaces_allowed: `cli/internal/application/audit.go`, its tests, `cli/docs/CONTRACT.md`, `docs/ARCHITECTURE.md`
- surfaces_avoided: `cli/internal/application/init.go`, `docs/plans/`, `docs/decisions/`, `docs/audit/`
- waves:
  - **Wave 1 — pin format and citation extraction**
    - task: define the opt-in pin declaration and extract `path/to/file.go:NN` citations from a pinned document.
      - expected_output: `docs/ARCHITECTURE.md`'s 21 existing citations are extracted; a document with no pin is skipped silently.
      - check: a document with no pin produces no finding and no error, and `zharness audit --json` on an unpinned tree is byte-identical to before (R5).
      - check: directory scoping holds — a file under `docs/plans/`, `docs/decisions/`, or `docs/audit/` never produces a drift finding even when pinned and stale (R4). Assert this with a test that pins a file in each of the three directories.
  - **Wave 2 — the drift measurement**
    - task: compare each cited path's commits against the pin and report lines added and removed since it.
      - expected_output: the finding names the document, each moved citation, and the size of the change — the four pieces F5 identifies.
      - check: pin `docs/ARCHITECTURE.md` to an old commit in a scratch clone; `zharness audit --json` names the moved citations with non-zero line counts.
      - check: pin it to `HEAD`; the finding clears.
      - check: a citation naming a path that no longer exists is reported distinctly from one that merely moved.
      - check: `audit` performs no repin and no write — `git status --short` is clean after the run, and neither sidecar appears (R7).
      - check: finding text states freshness only; no wording implies a pinned, unmoved document is accurate (R9).
- checks:
  - `cd cli && go test ./...` passes.
  - `zharness audit --json` distinguishes this finding from the `authored-docs-guard` finding by severity and identifier (R6).
  - `bash scripts/verify-doc-links.sh` exits 0.
  - `zharness check record` verdict is `APPROVED` or `APPROVE_WITH_REQUESTS`.
- escalation: if measuring drift requires judging whether prose is still true rather than whether commits exist, stop — that is NG5.

### phase_slug: `architecture-elicitation`
- story_id: 01M0FTV8MEZW4TM96C0JAJSS0W
- status: planned
- goal: `zharness init` scaffolds `docs/ARCHITECTURE.md` once, as at most five unanswered questions that no repository read can answer, and `audit` reports it as unanswered rather than counting it as documentation.
- depends_on: authored-docs-guard
- requirements: R1, R15, R16
- surfaces_allowed: `cli/internal/application/init.go`, its tests, `cli/internal/application/audit.go`, `cli/docs/CONTRACT.md`, `docs/README.md`
- surfaces_avoided: `cli/docs/embedded/`, `scripts/`, `docs/audit/`
- waves:
  - **Wave 1 — the question form**
    - task: add a fourth entry to `scaffoldOnceDocs` in `cli/internal/application/init.go:126` containing at most five questions: the problem and its audience; the main use cases in domain language; domain nouns whose meaning is non-standard; invariants that fail silently; and boundaries that must not be crossed.
      - expected_output: the file contains no statement about the repository — only questions and blank answer space.
      - check: the admission test, applied per question — none is answerable by reading the repository. A question that a `Glob` plus `Read` could answer is deleted, not reworded.
      - check: `scaffoldOnceDocs` has exactly four entries (NG2 caps it there).
      - check: in a temporary directory, `zharness init` creates the file; writing content into it and re-running `init` leaves it untouched, proving scaffold-once semantics.
      - check: this cannot be observed in this repository — `docs/ARCHITECTURE.md` already exists and `writeScaffoldOnceDocs` skips existing paths — so the temporary-directory run is the only valid evidence. Recording a pass from this repository is a false claim.
  - **Wave 2 — audit reports it unanswered**
    - task: `audit` reports the file as unanswered while every question is blank, and never counts it as authored documentation for R2's precondition.
      - check: in a temporary directory, `init` then `zharness audit --json` reports unanswered; the R2 finding must not be suppressed by the scaffolded file's mere existence.
      - check: answering the questions clears the unanswered report.
      - check: the finding text does not assert the answers are correct (R9).
  - **Wave 3 — close the initiative**
    - task: repoint every citation naming this plan's active path by slug before it moves to `docs/plans/completed/` (R16).
      - check: `grep -rn 'consumer-doc-drift-gate' . 2>/dev/null | grep -v '^./.git' | grep -v '^./docs/plans/active/'` returns nothing, or every hit is a generic `plans/active/` link carrying no slug.
      - check: after `zharness plan complete`, `bash scripts/verify-doc-links.sh` exits 0.
- checks:
  - `cd cli && go test ./...` passes.
  - `zharness init` on this repository leaves `docs/ARCHITECTURE.md` byte-identical — `git diff --stat docs/ARCHITECTURE.md` is empty.
  - `bash scripts/verify-doc-links.sh` exits 0.
  - `zharness check record` verdict is `APPROVED` or `APPROVE_WITH_REQUESTS`.
- escalation: if the five questions cannot be written without describing this repository, R15's premise is wrong — stop and route to `brainstorm refine` rather than shipping generated prose under an elicitation label.

## Progress
<!-- Append-only durable entries record timestamp, phase, wave, task, task_status, run_id, trace_id, exact verification/result, and changed surfaces or blocker. -->
- `2026-08-21T04:36:44Z` — handoff recorded. handoff: `01M0H9PS6DFANMV2PHBHYX7YRM`. open items: Phase rebuild-phase-parser is planned and unstarted. It owns R17: planFieldValue (cli/internal/application/rebuild.go:151) anchors on ^[ \t]*<key>: and cannot see a field written as a markdown list item, so rebuildStoriesFromPlan drops the story through its malformed-block continue branch with no error. Unblock: run work full phase rebuild-phase-parser.; harness.db cannot currently be regenerated from committed markdown for this plan. zharness query phases --json lists rebuild-phase-parser (minted directly by the CLI) but none of the five H3-form phases this plan declares, and a db rebuild --yes additionally loses pr60-doc-truth and pr60-go-correctness. Do not trust any phase-count check until R17 lands.; R15 cannot be verified in this repository because docs/ARCHITECTURE.md already exists and scaffold-once skips existing paths; its only valid evidence is a temporary-directory init run.; Follow-on initiative agreed but not locked: delete completed plans outright when done (upstream harness-experimental policy). Four approved phases -- ADR absorption, three docs/plans README policy files, delete the five existing completed plans and repair their non-ADR citations, and change zharness plan complete to delete rather than move. Cannot be locked until this plan completes and frees the single-active-plan slot..
- `2026-08-21T04:47:08Z` — wave 0. run: `01M0HA8TV0J948PJRNPBZC5BN4`. summary: Phase rebuild-phase-parser started (run 01M0HA8TV0J948PJRNPBZC5BN4): fix planFieldValue (rebuild.go:151) to accept markdown list-item fields so db rebuild stops dropping story rows (R17)..
- `2026-08-21T04:47:52Z` — wave 1, task reproduce the loss. task_status: `DONE`. run: `01M0HA8TV0J948PJRNPBZC5BN4`. summary: Added TestRebuildFromMarkdownListItemStoryFields (rebuild_test.go): H3 phase block with list-item story_id fails RebuildResult.Stories=0,want 1 on unmodified parser (cd cli && go test ./internal/application/ -run Rebuild fails naming the missing row, no panic/compile error)..
- `2026-08-21T04:47:54Z` — wave 1. run: `01M0HA8TV0J948PJRNPBZC5BN4`. summary: Wave 1 complete: R17 defect pinned by a failing test before touching the parser..
- `2026-08-21T04:48:12Z` — wave 2, task widen the field matcher. task_status: `DONE`. run: `01M0HA8TV0J948PJRNPBZC5BN4`. summary: planFieldValue (rebuild.go:160) now accepts an optional '- ' list marker before the key via (?:-[ \t]*)? in the anchor; cd cli && go test ./... passes all 6 packages; git diff --stat shows a single-line regex change..
- `2026-08-21T04:48:14Z` — wave 2. run: `01M0HA8TV0J948PJRNPBZC5BN4`. summary: Wave 2 complete: field matcher widened, full suite green, single-line diff..
- `2026-08-21T04:51:31Z` — wave 3, task prove the database is regenerable. task_status: `DONE_WITH_CONCERNS`. run: `01M0HAHCF491420NFW88Y6DMGK`. summary: Rebuilt with a from-source dev binary (installed zharness was stale): all 15 stories recovered, none lost, all 6 of this plan's declared phase slugs present (rebuild-phase-parser, audit-restore, link-gate-widening, authored-docs-guard, pin-drift-finding, architecture-elicitation). Concern: pr60-doc-truth/pr60-go-correctness still absent -- third heading form, out of R17's/this phase's scope (plan_query.go avoided), see decision..
- `2026-08-21T07:09:58Z` — wave 3. run: `01M0HAHCF491420NFW88Y6DMGK`. summary: Wave 3 complete: rebuild against real repo (from-source binary) recovered all 15 stories, none lost; the pr60-doc-truth/pr60-go-correctness gap is out of this phase's authority and routed to phase legacy-phase-form via decision 01M0HJF6MTW7BS957XHKF33H32..
- `2026-08-21T07:13:40Z` — wave 0. run: `01M0HJNKD1EST0FK8BE67T6CJ2`. summary: Phase legacy-phase-form started (run 01M0HJNKD1EST0FK8BE67T6CJ2): close R18 -- widen planPhaseHeading (plan_query.go) for the em-dash heading form and add story:/depends-on: legacy field aliasing + backtick-value stripping to planFieldValue (rebuild.go)..
- `2026-08-21T07:14:40Z` — wave 1, task pin the heading-discovery gap. task_status: `DONE`. run: `01M0HJNKD1EST0FK8BE67T6CJ2`. summary: Added TestExtractPlanPhaseBlockEmDashHeadingFormFindsPhase + LastPhaseStopsAtNextTopLevelHeading (plan_query_test.go): em-dash heading fixture from pr60-review-fixes.md fails not found on unmodified parser (go test -run TestExtractPlanPhaseBlockEmDash -v fails naming not found, no panic/compile error)..
- `2026-08-21T07:14:40Z` — wave 1, task pin the field-name gap. task_status: `DONE`. run: `01M0HJNKD1EST0FK8BE67T6CJ2`. summary: Added TestRebuildFromMarkdownEmDashHeadingLegacyFields (rebuild_test.go): story:/depends on: legacy fields with backtick-wrapped value fail RebuildResult.Stories=0,want 1 on unmodified parser (go test -run TestRebuildFromMarkdownEmDashHeadingLegacyFields -v fails naming the missing row, no panic/compile error)..
- `2026-08-21T07:14:40Z` — wave 1. run: `01M0HJNKD1EST0FK8BE67T6CJ2`. summary: Wave 1 complete: both R18 gaps (heading discovery, field-name matching) pinned by failing tests before touching either parser..
- `2026-08-21T07:15:24Z` — wave 2, task widen planPhaseHeading for the em-dash heading form. task_status: `DONE`. run: `01M0HJNKD1EST0FK8BE67T6CJ2`. summary: plan_query.go:41 planPhaseHeading now matches both phase_slug: and Phase N — (em dash) forms; extractPlanPhaseHeadingBlock boundary logic untouched. go test -run TestExtractPlanPhaseBlockEmDash -v passes, TestSetPlanPhaseStatus* and all TestExtractPlanPhaseBlock* pass unedited..
- `2026-08-21T07:15:24Z` — wave 2, task add legacy field-name aliasing and backtick-value stripping to planFieldValue. task_status: `DONE`. run: `01M0HJNKD1EST0FK8BE67T6CJ2`. summary: rebuild.go: planFieldLegacyAlias maps story_id->story, depends_on->depends on; planFieldValue tries the alias when the exact key misses, then trims a backtick-wrapped value to its first backtick pair. go test -run TestRebuildFromMarkdownEmDashHeadingLegacyFields -v passes; TestRebuildFromMarkdownRoundTrip, TestRebuildFromMarkdownDropsUncheckedRun, TestRebuildFromMarkdownListItemStoryFields all pass unedited..
- `2026-08-21T07:15:24Z` — wave 2. run: `01M0HJNKD1EST0FK8BE67T6CJ2`. summary: Wave 2 complete: both fixes landed, full suite green (7 packages), no unrelated test needed edits..
- `2026-08-21T07:19:42Z` — wave 3, task rebuild against the real repository and confirm both stories return with correct IDs and nothing else is lost. task_status: `DONE`. run: `01M0HK0Z2ABNXGGCQ59CW8QM8R`. summary: Rebuilt ~/.local/bin/zharness from source with both R18 fixes, then zharness db rebuild --yes --json: 18 stories (was 16). All 16 prior slugs present (missing set() empty); new: pr60-go-correctness, pr60-doc-truth. sqlite3 harness.db select on stories confirms IDs 01M0EG9ZS5XTQJV3J5J2CZ689B (pr60-go-correctness) and 01M0EG9ZSERMRZVPE860XRCJPR (pr60-doc-truth) exactly match the ULIDs cited in docs/plans/completed/pr60-review-fixes.md:82,119 -- backtick-stripping confirmed correct..
- `2026-08-21T07:19:42Z` — wave 3. run: `01M0HK0Z2ABNXGGCQ59CW8QM8R`. summary: Wave 3 complete: real-repo rebuild recovers pr60-doc-truth and pr60-go-correctness losslessly, correct IDs, nothing lost. R18 falsification condition satisfied..
- `2026-08-21T07:23:26Z` — wave 1, task restore the three documents verbatim from the commit before the deletion. task_status: `DONE`. run: `01M0HK7JGA9FNKP389B2G7G25D`. summary: git show 655c6ac^:docs/audit/{f}.md > docs/audit/{f}.md for workflow-harness-ceremony-audit, sdlc-token-cache-audit, sdlc-gap-analysis. wc -l: 399, 212, 126 -- exact match. diff against 655c6ac^ blob: empty for all three, byte-identical..
- `2026-08-21T07:23:26Z` — wave 1, task confirm docs/README.md ownership table covers docs/audit/. task_status: `DONE`. run: `01M0HK7JGA9FNKP389B2G7G25D`. summary: grep -n docs/audit docs/README.md:37 already lists docs/audit/ as authored -- no edit needed..
- `2026-08-21T07:23:26Z` — wave 1. run: `01M0HK7JGA9FNKP389B2G7G25D`. summary: Wave 1 complete: three deleted audit docs restored byte-identical from 655c6ac^, ownership table already covers docs/audit/..
- `2026-08-21T07:37:31Z` — wave 1, task phase-level check: bash scripts/verify-doc-links.sh still passes. task_status: `DONE_WITH_CONCERNS`. run: `01M0HK7JGA9FNKP389B2G7G25D`. summary: Restoring the three audit docs verbatim reintroduced 6 broken cross-references to docs/plans/completed/{eval-layer,harness-convergence-pass-v3,harness-memory-ceremony-convergence,workflow-harness-history-2026-07}.md and docs/evals/failures.md -- all 5 targets deleted in the same 655c6ac commit, out of R12's restoration scope. Added 5 matching .claimignore entries (same format/precedent as the 6 existing historical-audit-measurement entries). bash scripts/verify-doc-links.sh now reports doc links OK (0 findings)..
- `2026-08-21T07:37:31Z` — wave 2, task remove or replace the docs/audit/consumer-adoption-audit.md citation in each of the three embedded playbooks. task_status: `DONE`. run: `01M0HK7JGA9FNKP389B2G7G25D`. summary: sed replaced (R1, `docs/audit/consumer-adoption-audit.md` D1) -> (R1, D1) in brainstorm.md:42, (R2, ...) -> (R2, D1) in work.md:33, (R5, ...) -> (R5, D1) in handoff.md:36. grep -rn consumer-adoption-audit cli/docs/embedded/ returns nothing. git diff --stat shows only the three files; git diff shows exactly one changed line per file, no numbered step added or removed..
- `2026-08-21T07:37:31Z` — wave 2, task regenerate the projected copies under docs/playbooks/ so the scaffolded tree matches the embedded source. task_status: `DONE_WITH_CONCERNS`. run: `01M0HK7JGA9FNKP389B2G7G25D`. summary: zharness init --json returned status:exists and left docs/playbooks/ untouched -- writeScaffoldOnceDocs (init.go:135-151) skips any scaffold-once path already on disk, unconditionally. diff -r and git status checks cannot pass via this command. Accepted per decision 01M0HM1EEB798XJ061KH0JB9CH: gap locked as R19, closed by new phase playbook-scaffold-sync (depends_on: audit-restore)..
- `2026-08-21T07:37:31Z` — wave 2. run: `01M0HK7JGA9FNKP389B2G7G25D`. summary: Wave 2 complete with one accepted concern: playbook citations deprojected cleanly (task 1 DONE); docs/playbooks/ sync deferred to new phase playbook-scaffold-sync (R19) since zharness init cannot perform it for an existing checkout (task 2 DONE_WITH_CONCERNS, decision 01M0HM1EEB798XJ061KH0JB9CH)..
- `2026-08-21T07:38:22Z` — wave 1, task copy the three files audit-restore edited from the embedded source to the projected copy, byte-for-byte. task_status: `DONE`. run: `01M0HM32RWXPCFWT7TTNH42YAE`. summary: cp cli/docs/embedded/playbooks/{handoff,work,brainstorm}.md docs/playbooks/. diff -r cli/docs/embedded/playbooks docs/playbooks reports no differences. git status --short docs/playbooks/ shows exactly handoff.md, work.md, brainstorm.md modified, nothing else..
- `2026-08-21T07:38:22Z` — wave 1. run: `01M0HM32RWXPCFWT7TTNH42YAE`. summary: Wave 1 complete: docs/playbooks/ resynced to cli/docs/embedded/playbooks/. R19 falsification condition satisfied..
- `2026-08-21T08:29:09Z` — wave 0. run: `01M0HPWQAGDM4H7DSQAKQZ2H5C`. summary: Phase link-gate-widening started (run 01M0HPWQAGDM4H7DSQAKQZ2H5C). The playbook requests a phase-start task_status=in-progress, but the live trace CLI accepts only DONE, DONE_WITH_CONCERNS, NEEDS_CONTEXT, or BLOCKED; recording the phase start as a wave summary and proceeding with the defined tasks..
- `2026-08-21T08:32:07Z` — wave 1, task add cli to the find and exclude cli/testdata by category. task_status: `DONE`. run: `01M0HPWQAGDM4H7DSQAKQZ2H5C`. summary: Updated scripts/verify-doc-links.sh to scan cli/ and exclude the frozen cli/testdata/ fixture by category; bash scripts/verify-doc-links.sh exited 1 with four cli/docs findings and no cli/testdata findings..
- `2026-08-21T08:32:39Z` — wave 1. run: `01M0HPWQAGDM4H7DSQAKQZ2H5C`. summary: Wave 1 complete: verify-doc-links.sh now scans cli/ while excluding cli/testdata/ as a frozen fixture category; the required adversarial run exits 1 with four cli/docs findings and no cli/testdata findings..
- `2026-08-21T08:37:22Z` — wave 2, task repoint the remaining dead citations. task_status: `DONE`. run: `01M0HPWQAGDM4H7DSQAKQZ2H5C`. summary: Repointed six stale docs/plans/active/ occurrences covering four plan records in cli/docs/CONTRACT.md and cli/docs/SCHEMA.md to governing ADRs 0001 and 0003; verify-doc-links.sh exits 0, and active-plan citation counts are 0 with only generic templates remaining..
- `2026-08-21T08:37:53Z` — wave 2. run: `01M0HPWQAGDM4H7DSQAKQZ2H5C`. summary: Wave 2 complete: all stale active-plan citations in cli/docs/CONTRACT.md and cli/docs/SCHEMA.md now point to governing ADRs; the link gate passes with 0 findings and no stale plan-record citation remains..
- `2026-08-21T08:47:33Z` — wave 3, task extend the claim regex to markdown-link targets and resolve relative paths. task_status: `DONE`. run: `01M0HPWQAGDM4H7DSQAKQZ2H5C`. summary: verify-doc-links.sh now scans markdown-link targets, resolves them from the referencing file directory (including ../), and preserves backtick allowlist behavior; final gate passes, both synthetic broken-link probes exited 1 with exact paths, and docs/README.md was clean after each revert..
- `2026-08-21T08:47:49Z` — wave 3. run: `01M0HPWQAGDM4H7DSQAKQZ2H5C`. summary: Wave 3 complete: markdown-link coverage and file-relative resolution are active; the final gate passes, synthetic root-style and ../ probes both fail as required, and both probes were reverted cleanly..
- `2026-08-21T09:45:18Z` — wave 0. run: `01M0HVAR32DTHHPACKGCF2MV0K`. summary: Phase authored-docs-guard started (run 01M0HVAR32DTHHPACKGCF2MV0K). The live trace CLI accepts only DONE, DONE_WITH_CONCERNS, NEEDS_CONTEXT, or BLOCKED for task entries, so phase start is recorded as a wave summary per the prior workflow decision..
- `2026-08-21T09:59:36Z` — wave 1, task change Audit to accept a repository root and update every caller and test. task_status: `DONE`. run: `01M0HVAR32DTHHPACKGCF2MV0K`. summary: Added repoRoot to application.Audit, passed the repository root from the CLI caller, and updated five application tests; cd cli && go build ./... && go test ./... passed, source versus installed audit --json output was byte-identical, and harness.db-wal/harness.db-shm remained absent..
- `2026-08-21T10:00:14Z` — wave 1. run: `01M0HVAR32DTHHPACKGCF2MV0K`. summary: Wave 1 complete: Audit now accepts a repository-root argument with no behavior change; source and installed audit output match byte-for-byte, the full Go build/test suite passes, and the read-only command creates no WAL/SHM sidecars..
- `2026-08-21T10:20:41Z` — wave 2, task add the embedded-set filesystem authored-docs finding. task_status: `DONE`. run: `01M0HVAR32DTHHPACKGCF2MV0K`. summary: Added a read-only audit scan over embedded managed paths and docs/ Markdown, emitting authored_docs_missing at warning severity without reading the managed-docs table; full go test ./..., verify-doc-links.sh, init-delete-docs-audit-restore, rebuild survival, three-array JSON, presence-only wording, no-sidecar, and no managed_docs-reference checks passed..
- `2026-08-21T10:21:09Z` — wave 2. run: `01M0HVAR32DTHHPACKGCF2MV0K`. summary: Wave 2 complete: audit reports the embedded-set authored_docs_missing warning when managed documentation remains but docs/ has no authored Markdown, clears when authored Markdown returns, survives db rebuild, preserves the three top-level arrays, and remains read-only..
- `2026-08-21T10:26:09Z` — wave 3, task record the authored-documentation boundary and ownership class. task_status: `DONE`. run: `01M0HVAR32DTHHPACKGCF2MV0K`. summary: Added ADR 0005, indexed it in docs/decisions/README.md, and clarified the authored ownership class in docs/README.md; the ADR template-heading check, docs ownership coverage check, bash scripts/verify-doc-links.sh, cd cli && go test ./..., and git diff --check all passed..
- `2026-08-21T10:26:55Z` — wave 3. run: `01M0HVAR32DTHHPACKGCF2MV0K`. summary: Wave 3 complete: ADR 0005 records the presence-versus-truth boundary, docs/README.md names authored ownership and delegates semantic review, the ADR index is updated, link and ownership checks pass, and the full Go suite remains green..

## Decisions
<!-- Append-only durable entries record timestamp, phase/task, decision, and rationale. -->
- `2026-08-20T10:21:54Z` — Guard authored consumer-repo documentation instead of generating it; lock the initiative as consumer-doc-drift-gate.. rationale: External evidence (docs/research/agent-documentation-evidence.md F1) measures LLM-generated context files at -0.5% to -2% task success and +20% cost on repositories that already carry docs, and only positive (+2.7%) once every .md and docs/ is deleted -- which is the opposite of a zharness-initialized repo. F3 puts documentation value at +0.007 (zero) on code-derivable questions and +0.071 on rationale. F6 disqualifies every external generator: DeepWiki is public-repo-only on the free tier, codesight is TS/JS-only, CodeWiki has no staleness answer. F5 shows four independent products converging on a four-piece drift mechanism of which this repo already holds three, missing only the pinned baseline SHA. Generating is crowded and net-negative here; guarding matches what the CLI already is..
- `2026-08-21T04:19:13Z` — R13 chốt: dead docs/plans/active/ citations repoint to the governing decision record, never to docs/plans/completed/.. rationale: The owner has adopted upstream harness-experimental policy — a completed plan is deleted once a decision record absorbs its lasting result. Repointing a citation at docs/plans/completed/ would therefore schedule the same link to break a second time, one initiative later, which is exactly the failure mode R16 exists to prevent..
- `2026-08-21T04:19:13Z` — Add R17 and a new first phase rebuild-phase-parser to fix planFieldValue in rebuild.go.. rationale: Reproduced 2026-08-20: planFieldValue (rebuild.go:151) anchors on ^[ \t]*<key>: and so cannot see a field written as a markdown list item. rebuildStoriesFromPlan then takes its malformed-block continue branch and drops the story with no error. Measured: 0 of 5 story_id lines in this plan match the regex, 2 of 2 match in docs-architecture.md. After db rebuild --yes all five of this plan story IDs are absent and pr60-doc-truth and pr60-go-correctness are lost. Rows that markdown cannot regenerate violate ADR 0001 directly, and PR #67 body claims a query phases verification that does not reproduce locally, so the defect blocks trusting any later phase gate..
- `2026-08-21T04:51:26Z` — Mid-phase, `db rebuild --yes` was accidentally run against the globally-installed zharness binary (built 2026-08-20, predating this phase's fix), which temporarily dropped the rebuild-phase-parser story row and invalidated run 01M0HA8TV0J948PJRNPBZC5BN4 (never check-backed, so the rebuild could not resolve its story_slug and dropped it per the documented uncheck-run tradeoff). Rebuilt a local dev binary from source (go build ./cmd/zharness) and re-ran the rebuild with it; recovered all 15 stories with none lost. Minted a fresh run 01M0HAHCF491420NFW88Y6DMGK to continue recording. (phase: `rebuild-phase-parser`), task: prove the database is regenerable. rationale: The installed CLI is a separately released binary, not auto-synced with source edits; verifying a source-level fix against the real repo requires building from source first. Markdown Progress content survived the incident untouched since it is the write target; only the DB-side run row and its trace/decision linkage needed reconstruction..
- `2026-08-21T04:51:26Z` — Wave 3 check 'query phases lists pr60-doc-truth and pr60-go-correctness' cannot pass within this phase's authority; left both stories unrecovered rather than widening scope. (phase: `rebuild-phase-parser`), task: prove the database is regenerable. rationale: docs/plans/completed/pr60-review-fixes.md uses a third phase-heading form, ### Phase N -- `{slug}`, which its own text (line 171) already documents as a gap in extractPlanPhaseHeadingBlock/plan_query.go, distinct from R17 (planFieldValue). Fixing it requires editing plan_query.go, which is on this phase's surfaces_avoided list. The phase's own escalation clause names this exact fork: stop and route to brainstorm refine rather than widen scope. The blocker's unblock condition (all six of this plan's declared phase slugs regenerate, nothing lost) is otherwise fully met: rebuild with the fixed binary recovered 15 stories (10 pre-existing + 5 new: audit-restore, link-gate-widening, authored-docs-guard, pin-drift-finding, architecture-elicitation), losing none..
- `2026-08-21T07:09:53Z` — Accept Wave 3 concern on rebuild-phase-parser as resolved-by-routing, not by this phase itself. (phase: `rebuild-phase-parser`), task: prove the database is regenerable. rationale: Investigation while scoping the follow-on fix found the pr60-doc-truth/pr60-go-correctness gap is compound: (1) plan_query.go planPhaseHeading does not match the pr60-review-fixes.md em-dash heading form, so the phase block is never found, and (2) even once found, its story:/depends on: field names (backtick-wrapped value) are not matched by planFieldValue, so the story would still be dropped. Both were locked as R18 and phased as legacy-phase-form (story 01M0HHQPDMEWEF9ZET6649GV28, depends_on: rebuild-phase-parser) via brainstorm refine + to-plan phase, per explicit owner choice. rebuild-phase-parser Wave 3s second check ("query phases still lists pr60-doc-truth/pr60-go-correctness") is therefore known-unsatisfiable by this phase alone and is accepted as such; the phases actual stop_condition ("no story present before the rebuild is absent after it") held clean on the from-source rebuild and remains the binding gate for this phase..
- `2026-08-21T07:37:19Z` — audit-restore Wave 2 task 2 (regenerate docs/playbooks/ via zharness init --json) cannot pass its own check: writeScaffoldOnceDocs (cli/internal/application/init.go:135-151) only writes a scaffold-once path when absent, unconditionally, no --force override -- so init leaves docs/playbooks/ stale in every existing checkout. This is R19, locked via brainstorm refine and closed by new phase playbook-scaffold-sync (depends_on: audit-restore). Wave 2 task 2 is accepted as DONE_WITH_CONCERNS: its command ran as specified and its non-scaffold-sync intent (keeping cli/docs/embedded/playbooks/ the corrected source) is satisfied; the diff -r/git status checks are deferred to playbook-scaffold-sync rather than re-litigated here, since task definitions are immutable after to-plan.. rationale: Consistent with the rebuild-phase-parser precedent (decision reconciling an unsatisfiable Wave 3 check after R17/R18 split): fixing a to-plan-time design gap requires a new phase, not a hand-edit of an immutable task definition, and the original task still ran correctly for the part within its actual surfaces_allowed..
- `2026-08-21T08:58:43Z` — Record the phase start as a wave-level summary rather than a task trace with status in-progress. (phase: `link-gate-widening`), task: start phase link-gate-widening. rationale: The work playbook requests task_status=in-progress for phase start, but the live trace CLI rejects that value and accepts only DONE, DONE_WITH_CONCERNS, NEEDS_CONTEXT, or BLOCKED. The attempted command returned invalid_task_status; a wave summary preserves the start evidence without mislabeling a task as complete..

## Validation
<!-- Append-only durable entries record timestamp, phase, exact command/result/output, run_id, check_id, verdict, and proof_gaps. -->
- `2026-08-21T07:11:18Z` — check. verdict: `APPROVED`. check: `01M0HJHSTF2DK59T8CFVC01BBY`. run: `01M0HAHCF491420NFW88Y6DMGK`. phase: `rebuild-phase-parser`. judge: `same-session` (claude-sonnet-5).
  - `cd cli && go test ./internal/application/ -run Rebuild -v` → Validation entry 2026-08-21T04:47:52Z: TestRebuildFromMarkdownListItemStoryFields failed pre-fix (Stories=0), passed post-fix
  - `cd cli && go test ./...` → Validation entry 2026-08-21T04:48:12Z: all 6 packages ok
  - `bash scripts/verify-doc-links.sh` → doc links OK (0 findings)
- `2026-08-21T07:20:02Z` — check. verdict: `APPROVED`. check: `01M0HK1SKVH266TZ3Z4RRRR1NS`. run: `01M0HK0Z2ABNXGGCQ59CW8QM8R`. phase: `legacy-phase-form`. judge: `same-session` (claude-sonnet-5).
  - `cd cli && go test -run TestExtractPlanPhaseBlockEmDash -v ./internal/application/`
  - `cd cli && go test -run TestRebuildFromMarkdownEmDashHeadingLegacyFields -v ./internal/application/`
  - `cd cli && go test ./...`
  - `bash scripts/verify-doc-links.sh`
- `2026-08-21T07:37:47Z` — check. verdict: `APPROVE_WITH_REQUESTS`. check: `01M0HM29N12HZ9FM6JPSW3YPYQ`. run: `01M0HK7JGA9FNKP389B2G7G25D`. phase: `audit-restore`. judge: `same-session` (claude-sonnet-5).
  - `cd cli && go test ./...`
  - `bash scripts/verify-doc-links.sh`
  - `grep -rno 'docs/audit/[a-z0-9-]*\.md' . 2>/dev/null | grep -v '^./.git' | awk -F'docs/audit/' '{print $2}' | sort | uniq -c`
- `2026-08-21T07:38:28Z` — check. verdict: `APPROVED`. check: `01M0HM3HJMK7TWDGE4AJY9FWJ6`. run: `01M0HM32RWXPCFWT7TTNH42YAE`. phase: `playbook-scaffold-sync`. judge: `same-session` (claude-sonnet-5).
  - `diff -r cli/docs/embedded/playbooks docs/playbooks`
  - `bash scripts/verify-doc-links.sh`
- `2026-08-21T08:53:39Z` — check. verdict: `APPROVED`. check: `01M0HRD7G9YXWV41J77E3SWE82`. run: `01M0HPWQAGDM4H7DSQAKQZ2H5C`. phase: `link-gate-widening`. judge: `same-session` (gpt-5.6-luna).
  - `bash scripts/verify-doc-links.sh` → Validation entry 2026-08-21T08:53:03Z: doc links OK (0 findings)
  - `cd cli && go test ./...` → Validation entry 2026-08-21T08:53:03Z: all Go packages passed
  - `git diff --stat -- cli/testdata/` → Validation entry 2026-08-21T08:53:03Z: empty output

## Current State and Next Action
- active_phase: authored-docs-guard
- lifecycle_status: in-progress
- latest_run_id: 01M0HVAR32DTHHPACKGCF2MV0K
- latest_trace_ids: [01M0HXPJHM1YV0TNNWJ616HHG1, 01M0HXR08SY3A3H654RR22KZAV]
- latest_check_id: 01M0HRD7G9YXWV41J77E3SWE82
- latest_handoff_id: 01M0H9PS6DFANMV2PHBHYX7YRM
- blockers: none
- open_items: [R15 cannot be verified in this repository because docs/ARCHITECTURE.md already exists and scaffold-once skips existing paths; its only valid evidence is a temporary-directory init run | Follow-on initiative agreed but not locked — delete completed plans outright when done, per upstream harness-experimental policy; four approved phases covering ADR absorption, three `docs/plans/` README policy files, deletion of the five existing completed plans with their non-ADR citations repaired, and changing `zharness plan complete` to delete rather than move. It cannot be locked until this initiative completes and frees the single-active-plan slot.]
- exact_next_action: execute the in-session durable phase gate for `authored-docs-guard`: run the check playbook gate, record APPROVED or APPROVE_WITH_REQUESTS, and synchronize DB and plan status to checked
