# Migrating workflow procedures into installed skills

**Candidate implemented; consumer verification pending.** This guide records the accepted migration in [ADR 0015](../decisions/0015-skill-owned-procedures-and-safe-retirement.md). It does not identify a released migration binary or claim that consumer smoke has passed. Pin the accepted source and binary identities from the completed migration evidence before using these instructions on a real repository.

## Compatibility and prerequisites

| Source shape | Procedure authority | Binary distribution |
|---|---|---|
| Before migration | Consumer `docs/playbooks/` | WORKFLOW, twelve playbooks, AGENTS block, PROJECT scaffold |
| Reference-copy bridge | Skill references available; existing WORKFLOW still routes to projections | Same old inventory and bytes; safe to pause here |
| Completed transition | Six installed workflow skills and their selected references | WORKFLOW, AGENTS block, PROJECT scaffold; conservative legacy retirement |

Install **brainstorm, to-plan, work, check, handoff and watzup** independently using the [existing skills distribution instructions](../../README.md). Verify the host exposes each skill's metadata and reads references from its installed directory. Binary installation does not install skills. A host without this dependency set cannot execute the workflow; report the missing dependency rather than reading a leftover playbook or installing skills implicitly.

References resolve companions from their own skill directory and consumer data from the consumer root. Cross-stage routing uses host metadata, without assuming all skills share an installation parent. Migration preserves existing stage behavior: normal work remains Goal plus `main`, skips `to-plan`, and has one independent final check.

## Snapshot before upgrade

Record the old release version, binary hash/source identity, candidate source identity including dirty changes, host version/settings and independently installed skill inventory. Capture the complete consumer tree, including `.zharness/base/`, upstream blobs, `ownership.tsv`, originals, unrelated legacy-directory content and host configuration. A source SHA alone does not identify uncommitted candidate bytes.

Keep the three metadata roles separate: the active base records upstream digests; the ownership ledger records origin; `.orig` stores the original consumer payload. Back up all three. Interrupted restoration can also retain path/digest-bound intent under `.zharness/base/retirement-intent/`; capture it with the rest of the base directory. Intent is installer recovery evidence, not workflow state or a replacement for missing base/ownership/original data. A file merely equal to its saved original is still an edited keep case without valid recorded restoration intent. Malformed, stale or linked intent stays unresolved and must not be replaced to manufacture proof. Use disposable consumers and isolated host configuration first; do not validate by changing global skills or a real consumer.

## Expected update and uninstall behavior

- A known legacy projection matching its recorded digest, proven created by the installer and without a saved original, is removed. Unrelated content remains.
- A matching projection with proven preexisting origin and a valid original is restored to those original bytes. Only after verified restoration are that path's obsolete records removed; the restored file is unmanaged.
- Edited files, missing originals, unknown/malformed provenance and contradictory created-origin-plus-original evidence are retained with a reason. They stop being active update targets, but their original/provenance records stay available. Later uninstall keeps both the retained bytes and recovery data.
- Missing base evidence means keep. Corrupt/unreadable base fails before writes. A failure during retirement returns nonzero and preserves evidence for retry; inspect actual partial results rather than assuming nothing changed.
- A repeated update must not rewrite retained files, delete additional content or reintroduce retired base entries. Only an empty legacy directory can be removed. Retirement through symlinked parents is retained with notice to protect external bytes; existing installation directory-symlink behavior remains governed by ADR 0013.

Canonical skills do not read retained old instructions. A leftover file is preserved consumer content, not an alternate workflow authority. Uninstall is not a request to discard unresolved original backups.

## Acceptance before adoption

Installer fixtures must cover M01–M11, including metadata after repeated update/uninstall, malformed ownership, original restoration, partial failure and symlinks. Prepare two consumers from a pinned previous released binary: one fresh install with later edits/unrelated content, and one with a preexisting playbook whose original is demonstrably captured.

Separately install the candidate skills outside the source checkout. Record real host traces proving companion and cross-stage resolution, consumer-root paths and no reads of retained old instructions (M12–M14). Execute bounded behavior against an existing test; interrupt normal durable work after one verified requirement, then resume in a fresh session through watzup/work/check/handoff. Missing authenticated runtime or trace evidence is blocked, not passed.

Run Rust format/lint/tests, the musl binary-size gate, hook fixtures and doc-link checks. Independent final review must examine requirements, consumer artifacts and recovery, not just green command exits. Migration makes no claim about model quality or token savings.

## Recovery and next initiative

Stop on preservation failures. Keep the complete failed fixture, logs and metadata. Restore the captured consumer snapshot offline for consumer rollback. Source rollback reverses the coherent installer/routing transition first to restore the frozen compatibility bridge, then the reference-copy change if needed. Do not use an old updater as a restoration tool: its fresh-overwrite behavior can erase retained edits.

Only after all migration acceptance, independent final approval and closing handoff may the next plan be activated. Archive migration, confirm the active slot is empty, **move** the draft lean-SDLC plan into that slot and record the actual accepted migration completion SHA as baseline B. Obtain commit approval if that identity does not yet exist. No performance evaluation, release or global upgrade is implied by these instructions.
