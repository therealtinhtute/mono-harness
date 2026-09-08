# Integrity audit: installer ownership, recovery, and verification boundaries

**Repository:** `therealtinhtute/mono-harness`
**Reviewed revision:** `a45642227ea26ec72ea111ddb2fefd626e3026ab`
**Review date:** 2026-09-07
**Scope of this change:** audit documentation and opt-in reproduction fixtures only; no installer, hook, CI, plan, or product-policy changes.

## Executive assessment

The repository has a useful, deliberately small architecture: committed documents carry the workflow, and `zharness` manages installation, update, and removal rather than orchestrating agents. Its distinction between bounded and durable work, independent final review, and absorption of lessons into native guards is sound. The implementation also contains meaningful tests and explicit ownership and rollback contracts. [S01], [S02], [S07]

However, the reviewed implementation does not consistently uphold those contracts. This audit confirms **eight findings: four P1 and four P2**, plus two separately identified enforcement limitations. The most consequential issues are deletion of newly added consumer text during uninstall, incomplete recovery being accepted as success, and validation checks omitted at completion and across multi-commit push ranges.

**Recommendation:** request changes before relying on blanket consumer-data-preservation or fail-closed-verification guarantees. This is not a recommendation to replace the markdown-first architecture. Fix ownership accounting, transaction boundaries, and the integration between guards and Git state first.

This report is **not** a security certification, full-repository test result, or measurement of agent productivity. No exploitability, compromise, credential leakage, or CVE claim is made.

## Method, provenance, and evidence limits

The current branch was resolved through the connected GitHub API and the review was pinned to the full revision above. The source review covered the installer, update/uninstall helpers, three-way merge implementation, hook core and wrappers, CI/release workflows, release installer, selected tests, and lifecycle documentation. Optional skills were sampled through their workflow integration; this was not an exhaustive review of every skill, historical commit, dependency, or bootstrap script.

The execution environment could read source through the GitHub connector but could not clone GitHub: DNS resolution failed. It provided Go **1.23.2**; the original module declares Go **1.25.0** and Cobra dependencies. Consequently, **the complete CLI was not built, and the repository's full test/vet/doc-link suites were not executed**. [S08]

Instead, the audit compiled selected original Go functions in a temporary standard-library-only package and exercised the original shell guard functions with controlled fixtures. Some scenarios compose the same helper calls as the production path; they do not invoke the complete `zharness` binary. The tests use only temporary files/repositories and harmless `true`/`false` proof commands. They do not run proof commands from actual project plans.

`uninstall.go` and `.github/workflows/cli-ci.yml` were reconstructed in full and their Git blob hashes matched the connector metadata. Other local inputs are explicitly labeled source excerpts. Function-level SHA-256 hashes, input blob identifiers, test outputs, and the distinction between excerpts and full files are recorded in the evidence. The original-source MIT notice accompanies the standalone audit bundle; source copies are not added by this proposed PR.

### Executed results

| Category | Checks | Result |
|---|---:|---|
| Installer contract probes | 5 | All 5 exposed the recorded defect |
| Guard integration contract probes | 4 | All 4 exposed defects; two cover the same F02 |
| Enforcement-limit probes | 2 | Both confirmed the stated limitation |
| Positive/negative controls | 9 | All 9 passed |
| Total | 20 | 11 contract failures; 9 controls passed |

The opt-in runner therefore exits **1**, intentionally reflecting failed desired contracts. This is **not 20 passing tests** and is not a failing result from the repository's existing suite. Exit **2** identifies harness/prerequisite errors or a failed control; exit **0** means the included expectations all hold.

Artifacts: [runner and instructions](2026-09-07-integrity-review/README.md), [machine-readable results](2026-09-07-integrity-review/probe-results.json), [console output](2026-09-07-integrity-review/probe-console.txt), and [source provenance](2026-09-07-integrity-review/source-provenance.json).

## Severity and findings index

P1 means high-priority correction because ordinary lifecycle operations can lose consumer content/recovery or omit a promised verification boundary. P2 means a material correctness or local-enforcement defect with narrower impact. These are engineering priorities, not CVSS scores.

| ID | Priority | Finding | Evidence |
|---|---|---|---|
| F01 | P1 | Uninstall deletes consumer prose added to installer-created `AGENTS.md` | Original helper execution |
| F02 | P1 | Completed plan additions/renames bypass proof and judge validation | Git index fixtures + original guards |
| F03 | P1 | Push CI validates only the last commit's active-plan changes | Multi-commit Git fixture + exact CI selector |
| F04 | P2 | Abort deletes an empty file that existed before update | Original stash helpers |
| F05 | P1 | Missing stash blob is skipped; recovery then reports success and drops backups | Fault-injected original stash helpers |
| F06 | P2 | Reinstall records generated bytes as a pre-install original | Original capture/removal helper composition |
| F07 | P2 | Uninstall removes a preexisting consumer-owned ignore rule | Original cleanup helpers + caller inspection |
| F08 | P2 | Local single-plan check examines worktree, not staged state | Real index/worktree divergence fixture |

### F01 — Consumer prose can be deleted with an installer-created AGENTS.md

**Location:** `cli/internal/installer/uninstall.go`, `removeAgentsBlock`; ownership originates in `captureOriginal` and `Install`. [S03], [S04]

**Precondition:** `AGENTS.md` did not exist before installation. The user later adds their own instructions outside the managed block.

`removeAgentsBlock` computes the remainder outside the markers, but first branches on the absence of a pre-install original. In that branch it deletes the entire file, regardless of the remainder. Initial creation by the installer is treated as perpetual ownership of every later byte.

**Observed:** the fixture added an independent deployment-approval instruction after the block. The helper removed `AGENTS.md`; the test reported `consumer prose lost: exists=false`. A control with a preexisting original retained surrounding prose, isolating the problematic ownership branch.

**Impact:** normal uninstall can erase new consumer instructions, contradicting the advertised surgical block ownership and consumer preservation. A lost local base/original directory may create a similar unknown-ownership state, but that additional scenario was not exercised here.

**Suggested correction:** strip only the managed block; delete the whole file only when the remaining bytes are known installer boilerplate with no consumer changes. Treat missing provenance as unknown, not proof of exclusive ownership. Preserve content inside a locally modified block or explicitly define and warn about its removal policy.

**Required regression coverage:** new consumer prefix/suffix; initially absent versus initially present file; whitespace-only original; missing provenance; ordinary generated boilerplate removal.

### F02 — Completing a plan can omit its final proof and judge checks

**Location:** pre-commit path selection in `scripts/install-git-hooks.sh`; active/completed steps in `.github/workflows/cli-ci.yml`. [S05], [S06]

Both wrappers select active-plan changes for `zharness_guard_entries_of_file`. Completed plans go through the separate phase-consistency guard, which does not execute proof commands or require an independent judge.

**Observed:** two fixtures were exercised: direct addition under `completed/`, and moving an active plan into `completed/` while changing its final proof to `false`. The production active-path selector returned no files; the phase guard accepted all-`done` phases. Running the original entry guard directly on the same content rejected the failing proof, so the issue is integration coverage, not failure detection inside that helper.

**Impact:** the final evidence can be absent from the very checks intended to justify closing an initiative. This does not require a malformed Markdown entry. Not every rename is necessarily skipped in every Git configuration; both reproduced scenarios and the omitted destination namespace are sufficient to establish the defect.

**Suggested correction:** process changes in both namespaces, explicitly handle rename source/destination paths and closing transitions, and apply proof/judge requirements to new final Validation entries. Preserve predecessor identity for comparison rather than treating every move as unrelated history.

**Required regression coverage:** add, modify, and rename into completed; passing/failing proof; same-session final review; unchanged historic entries; direct completed-plan creation. A phase-consistency pass must not substitute for evidence validation.

### F03 — A multi-commit push can hide an earlier unverified plan entry

**Location:** `hook-guard`, `Run guards against pushed commit's plan changes`; `zhuards_guard_plans` head-mode old/new selection. [S05], [S06]

The workflow fetches two commits and compares `HEAD~1 HEAD`. Its head-mode wrapper also uses the immediate parent. For a push containing multiple commits, this validates only the tip delta, not the full pushed range.

**Observed:** a fixture committed a failing approved Validation entry in commit A, then made an unrelated README change in commit B. The exact CI selector returned no plans at B. A comparison from the fixture's initial base selected the plan, and the original guard rejected its `false` proof.

**Impact:** an earlier commit in one push can introduce unverified evidence even when the last commit's hook-guard step returns success. The independently applied single-active-plan check does not close this gap. This finding specifically concerns the `push` path; the GitHub-generated pull-request merge ref may cover more than a single contributor commit, and no claim is made that every multi-commit PR is skipped.

**Suggested correction:** use event-appropriate bases and sufficient history. For final-state integrity, compare the push `before` SHA with `after`, and the appropriate PR base/merge base with its head. For an append-only history guarantee, also validate each new evidence-bearing transition: an endpoint diff alone cannot detect an entry introduced and later removed within the range. Define first-push, missing-parent, force-push, merge, and shallow-history behavior explicitly.

**Required regression coverage:** a two-commit push whose final change is unrelated; multiple plan changes; root commits; missing bases; valid and invalid PR merge states.

### F04 — Empty existing files are not restored as existing files

**Location:** `stashCapture` and `stashRestore` in `cli/internal/installer/update.go`. [S09]

The capture code distinguishes `existed` temporarily but writes no existence bit into the TSV metadata. Both an absent file and an existing empty file become a zero-byte stash blob. Restore interprets every zero-length blob as a request to remove the destination.

**Observed:** an existing empty `.gitignore` was captured, changed, and restored. The original helper deleted it. Controls confirmed ordinary nonempty restoration and removal of a file that was originally absent.

**Impact:** the pre-update state is not faithfully restored. The problem is the collision between existence and content length, not diff3 merging.

**Suggested correction:** store explicit existence and relevant file metadata per entry. Restore an empty file by writing zero bytes when `existed=true`; remove only when `existed=false`. Define legacy empty-blob recovery conservatively because the old format cannot recover the missing information.

**Required regression coverage:** absent, empty, nonempty and permission-sensitive files; interruption during restoration; legacy stash metadata. Use these tests against the exported update/abort flow as well as the helpers.

### F05 — Incomplete stash recovery returns success and removes remaining evidence

**Location:** `stashRestore`; the `RunUpdate` abort branch. [S09]

A read error on any stash payload executes `continue` rather than returning an error. After iterating, `stashRestore` removes the entire stash directory. Its caller sees `nil` and can announce restoration success.

**Observed:** the fixture captured two nonempty files, changed both, and removed one backing blob to simulate damaged/incomplete recovery data. Restore returned `nil`, left that file's changed bytes in place, and removed the stash directory: `error=<nil> stashExists=false a="changed a\n"`.

**Impact:** the user receives no actionable recovery failure, and the remaining recovery artifacts are destroyed. This scenario needs an incomplete stash, not merely an empty file, and is distinct from F04.

**Suggested correction:** validate the complete inventory before restoring; treat missing, malformed, unreadable or mismatched payloads as errors; keep the stash until all restores succeed. Propagate removal/write failures and report exactly what did and did not restore. Checksums help detect corrupted payloads but do not replace transactional bookkeeping.

**Required regression coverage:** missing payload, malformed metadata, unreadable payload, failed target write/removal, partial interruption, and retry. Fault injection should work without privileged accounts or modifying real consumer files.

### F06 — Repeated install loses the distinction between created and preexisting files

**Location:** `Install` unconditionally invokes `captureOriginal` for targets; `captureOriginal` and `removeManagedFile`. [S03], [S04]

On first install, an absent managed file produces no `.orig` record. After the installer writes the generated content, a second install sees an existing file and no original record, and captures those generated bytes as though they preceded installation.

**Observed:** the helper composition mirrors that sequence. A clean generated `docs/WORKFLOW.md` remained after removal logic ran because it had become its own supposed pre-install original. The single-install control removed the same kind of file normally. This was not an end-to-end CLI invocation.

**Impact:** install/reinstall/uninstall is no longer a reversible ownership lifecycle. It can leave managed documents behind while removing their bookkeeping. This is a narrower cleanup/provenance issue than F01's data loss.

**Suggested correction:** persist origin explicitly, for example `created_by_installer` or `preexisting`, on the first ownership decision. Reinstallation should use that record and never invent a new pre-install origin. Include a migration policy for installations lacking the bit.

**Required regression coverage:** one, two, and three install cycles; upgrade between cycles; preexisting consumer documents; local modifications; uninstall and reinstall after provenance loss.

### F07 — Ignore cleanup removes rules the installer did not add

**Location:** `appendGitignoreEntries`, `Uninstall`, `dropLines`/`dropLine`. [S03], [S04]

Installation avoids appending an ignore line that already exists. Uninstall nevertheless passes the full managed-line list to `dropLines`, without tracking which lines were actually introduced by the installer. A captured `.gitignore` original is not used by this cleanup path.

**Observed:** starting with a consumer-owned `/.zharness/` rule and adding only the installer marker, cleanup removed both the marker and the consumer rule. The fixture retained an unrelated `keep/` rule to demonstrate targeted ownership loss rather than wholesale file deletion.

**Impact:** consumer ignore policy changes after uninstall even when the installer did not add that rule. No claim is made that a secret is necessarily exposed. Separately, `dropLine` normalizes blank lines and terminal newlines; that byte-preservation concern is source-reviewed, not counted as another reproduced finding.

**Suggested correction:** record inserted lines/blocks and remove only owned additions; preserve any preexisting identical rule. Prefer a reversible marked insertion over globally stripping matching strings. Preserve an originally empty file's existence and unrelated formatting.

**Required regression coverage:** preexisting exact rule, whitespace variants, CRLF, original missing trailing newline, empty original, unrelated edits after installation, and repeated installs.

### F08 — The local one-plan guard validates a different state than the commit

**Location:** `zharness_guard_at_most_one_active_plan` and its pre-commit invocation. [S05]

The helper counts nonempty files in the working directory. Unlike the Validation-entry checker, the pre-commit wrapper does not provide staged blobs for this invariant.

**Observed:** two nonempty active plans were staged; the working-tree copy of one was then truncated without staging that change. The guard returned success, although `git show :docs/plans/active/b.md` still contained a nonempty plan. A control with both worktree files nonempty was rejected.

**Impact:** the local commit check can accept two staged active plans. Conversely, an untracked extra plan may reject an otherwise valid staged commit. The existing CI count on its clean checkout should catch the first case when CI runs, so this is specifically a local-enforcement gap, not proof of a complete CI bypass.

**Suggested correction:** make the data source explicit: index for pre-commit, checked-out commit for CI. Enumerate staged paths safely and read their blob sizes/content from Git, including deletions and renames. Keep worktree advisory checks separate from commit gating.

**Required regression coverage:** staged/worktree disagreement in both directions, untracked plans, deletions, empty files, renamed plans, unusual filenames, and linked worktrees.

## Confirmed limitations, not additional undisclosed bugs

### L01 — Judge declarations are negative string checks, not schema or provenance verification

The guard rejects explicit `judge: same-session`, but a `mode: full` entry with a passing proof and no judge declaration is accepted. The corresponding fixture reproduced this, while explicit same-session and independent-judge controls behaved as expected. [S05], [S07]

A schema validator can require the field and an allowed value. That still cannot establish independence: verification of a real review session requires a separate trusted mechanism. Keep these two promises distinct. This report does not mislabel a self-reported `independent` string as cryptographic attestation.

### L02 — Unknown entry formats can be ignored rather than rejected

A verdict nested below a narrative first line was ignored, and its `false` proof was never run. Both `check.md` and the existing guard-fixture comments explicitly document the first-line grammar and ignoring a verdictless first line. This is an existing design limitation, not a newly discovered parser regression. [S05], [S07], [S10]

Preserve the useful rule that quoted nested verdicts must not shadow a real entry header. If stronger enforcement is desired, add a separate schema check for required Validation records and final lifecycle transitions instead of broadening a regex so that arbitrary quoted prose becomes executable evidence.

## Architecture, maintainability, and delivery observations

**Keep the narrow installer boundary.** The three-verb design and repository-as-authority model reduce unnecessary runtime state. Read-only/bounded modes and stage-owned plan sections remain useful. These findings do not justify reintroducing an orchestration database. [S01], [S02], [S07]

**Separate policy, evidence syntax, and execution.** Today Markdown regular expressions, filename selectors, Git object selection, and shell execution jointly define the guarantee. Most guard findings are failures between those layers. Centralize event-to-revision selection, test transition-level scenarios, and make malformed required records visibly fail without interpreting unrelated prose as commands.

**Strengthen persistence as a lifecycle, not as isolated helper calls.** The source uses atomic replacement for individual files, but ownership, manifest/base publication, and multi-file rollback need coherent invariants. `saveBase` removes the upstream blob directory before completing the next snapshot; crash consistency and concurrent invocation therefore deserve dedicated fault-injection/locking review. This is a source-level risk, not an additional reproduced defect in this audit. [S03], [S09]

**Do not treat deliberate overwrites as an undisclosed merge bug.** The current documented policy intentionally overwrites playbooks and `WORKFLOW.md`; only selected consumer-owned surfaces merge. Improve warnings, previews, extension points, and upgrade guidance, but evaluate behavior against the actual documented policy. [S01]

**Consumer installation is not automatic enforcement.** The default install does not distribute a hook/CI guarantee, and `check.md` describes hook/CI use as opt-in. Keep status output precise about `local-only`, hook, and CI coverage. Do not describe a command that merely exits zero as proof that a product requirement was adequately tested. [S01], [S07]

**Release integrity has an improvement opportunity.** GoReleaser emits `checksums.txt`, while the examined download installer fetches the archive and installs the binary without checking that file. Consider release checksums/attestations and explicit version pinning. This is a source-reviewed hardening recommendation, not a demonstrated supply-chain attack. CI testing in the reviewed workflow is Ubuntu-based while binaries target Linux and macOS on two architectures; cross-compilation alone is not execution coverage. [S06], [S11], [S12], [S13]

## Suggested remediation sequence and acceptance gates

1. **Protect user content and recovery first:** F01, F04, F05, F07. Add exported-operation integration tests with temporary consumer repos, not only helper tests. Verify rollback returns actionable errors and retains evidence.
2. **Repair validation coverage:** F02 and F03. Define event bases and plan transitions, then run a matrix across additions, renames, completion, batch pushes, and missing history.
3. **Restore ownership and staging consistency:** F06 and F08. Make provenance durable and choose the Git object under validation explicitly.
4. **Clarify optional guarantees:** decide whether L01/L02 warrant schema enforcement; keep independent-review provenance and semantic proof quality outside the current guarantee unless implemented.

Before a corrective PR is approved, run the original repository gates on a full checkout with the declared toolchain: build, vet, complete Go tests, guard fixtures, and doc-link verification. Add macOS execution coverage for the changed filesystem/shell paths where supported. Record exact command output and any remaining unavailable environments rather than inferring a pass from this audit's controls.

The supplied probes are a starting point, not a permanent replacement for native tests. They extract selected functions and pin the reviewed selectors; function signature changes or refactored wrappers can require updating the harness. Do not wire its intentionally failing baseline into required CI until fixes and expectations are reconciled.

## Proposed PR scope and remaining gaps

This proposed PR adds only this report, an opt-in runner, test fixtures, and evidence. It does not fix F01–F08, change branch protections, install hooks, alter existing plans, or certify production readiness. The two enforcement-limit probes document a stronger desired contract and remain separately classified.

Not executed: complete `go test ./...`, `go vet ./...`, production CLI builds, existing complete `test-guards.sh`, repository-wide doc-link verification, macOS/Windows runs, dependency vulnerability scans, release-asset verification, hosted PR CI, or an independent human/model review. The local artifact checks and focused executions must not be substituted for those gates.

## Pinned source references

[S01]: https://github.com/therealtinhtute/mono-harness/blob/a45642227ea26ec72ea111ddb2fefd626e3026ab/README.md
[S02]: https://github.com/therealtinhtute/mono-harness/blob/a45642227ea26ec72ea111ddb2fefd626e3026ab/docs/ARCHITECTURE.md
[S03]: https://github.com/therealtinhtute/mono-harness/blob/a45642227ea26ec72ea111ddb2fefd626e3026ab/cli/internal/installer/installer.go
[S04]: https://github.com/therealtinhtute/mono-harness/blob/a45642227ea26ec72ea111ddb2fefd626e3026ab/cli/internal/installer/uninstall.go
[S05]: https://github.com/therealtinhtute/mono-harness/blob/a45642227ea26ec72ea111ddb2fefd626e3026ab/scripts/install-git-hooks.sh
[S06]: https://github.com/therealtinhtute/mono-harness/blob/a45642227ea26ec72ea111ddb2fefd626e3026ab/.github/workflows/cli-ci.yml
[S07]: https://github.com/therealtinhtute/mono-harness/blob/a45642227ea26ec72ea111ddb2fefd626e3026ab/cli/docs/embedded/playbooks/check.md
[S08]: https://github.com/therealtinhtute/mono-harness/blob/a45642227ea26ec72ea111ddb2fefd626e3026ab/cli/go.mod
[S09]: https://github.com/therealtinhtute/mono-harness/blob/a45642227ea26ec72ea111ddb2fefd626e3026ab/cli/internal/installer/update.go
[S10]: https://github.com/therealtinhtute/mono-harness/blob/a45642227ea26ec72ea111ddb2fefd626e3026ab/scripts/test-guards.sh
[S11]: https://github.com/therealtinhtute/mono-harness/blob/a45642227ea26ec72ea111ddb2fefd626e3026ab/scripts/install-zharness.sh
[S12]: https://github.com/therealtinhtute/mono-harness/blob/a45642227ea26ec72ea111ddb2fefd626e3026ab/cli/.goreleaser.yaml
[S13]: https://github.com/therealtinhtute/mono-harness/blob/a45642227ea26ec72ea111ddb2fefd626e3026ab/.github/workflows/cli-release.yml
