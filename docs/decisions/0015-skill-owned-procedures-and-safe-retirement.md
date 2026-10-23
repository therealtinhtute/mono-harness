# 0015 — Skills own procedures; retire projections without losing consumer recovery

## Status

Accepted owner decision, 2026-10-13; candidate implementation has passed installer fixtures, independent p3 review, isolated consumer verification and independent final full review. The candidate is locally verified and remains uncommitted and unreleased. This ADR is recorded before implementation because the owner explicitly requires an ADR-first migration. It is not evidence that a released binary implements retirement. Amends the playbook distribution part of [0007](0007-fresh-overwrite-for-playbooks-and-workflow.md); WORKFLOW overwrite and [0011](0011-update-without-three-way-merge.md) remain unchanged.

## Context

The pre-migration source at `156bc96b15715ddd8b43e69bbeaa4bcd0fff8c1a` derives managed playbooks from embedded names (`cli/src/installer/mod.rs:70`), while ownership seeding visits the supplied target inventory (`cli/src/installer/ownership.rs:148`). Dropping embedding first would also drop those paths from recovery. That source's uninstall clears originals after processing active targets (`cli/src/installer/uninstall.rs:63`); leaving that cleanup unchanged would discard consumer originals for retired paths.

The owner wants portable skills to own lifecycle instructions and the binary to remain install/update/uninstall only. Migration preserves procedure semantics, including [0014](0014-outcome-spec-normal-plans.md): normal is Goal plus one `main` phase, without `to-plan` or a duplicate gate. Semantic simplification belongs to the subsequent initiative.

## Decision

1. **Six workflow skills own twelve references.** `brainstorm` owns brainstorm/grill/lock; `to-plan` owns to-plan; `work` owns work/work-full; `check` owns check/check-durable/check-validation/check-review; `handoff` and `watzup` own their respective procedures. Wrappers select relevant references. Companion links resolve inside that skill; cross-stage transitions resolve the destination skill through host metadata. Consumer paths resolve from the consumer root. Neither a source checkout nor a shared skill installation parent is required.
2. **Skill installation is a prerequisite.** Consumers independently install all six skills with the existing skills distribution mechanism before upgrading. The binary neither discovers nor installs skills. Hosts without them are unsupported for workflow execution; retained old files are not a fallback. The binary still manages WORKFLOW, the AGENTS block and the PROJECT scaffold with their existing semantics.
3. **Keep a usable compatibility bridge.** The reference-copy phase keeps both old projection directories and the binary's old managed inventory byte-identical. The next transition changes WORKFLOW/live routing, retires old managed targets and removes old embedding/projections together. There is no intermediate delivery with routes pointing to removed procedures.
4. **Retirement uses evidence, not equality alone.** Only twelve fixed legacy paths are eligible. Load/seed ownership using current and retired inventory before active target removal. The active base manifest/upstream blobs describe managed upstream content; `ownership.tsv` describes provenance; `.orig` contains consumer recovery payload. They are not interchangeable. A digest match alone never licenses destroying an original.

   | Evidence/disposition | File action | Metadata after success |
   |---|---|---|
   | Recorded digest matches, proven installer-created, no original | Unlink managed copy | Remove only that retired base/blob and ownership row |
   | Recorded digest matches, proven preexisting, valid saved original | Atomically restore and verify original bytes | Then remove that base/blob, ownership row and original backup |
   | Edited, unknown/malformed provenance, missing original, or created origin contradicting a saved original | Keep bytes and report reason | Remove active base/blob; retain original and provenance for unresolved recovery |
   | File operation fails | Return nonzero with accurate partial result | Retain that path's prior base and recovery evidence for retry |

   Missing ownership ledger may use ADR 0008 legacy seeding from a proven prior base. An existing malformed ledger is not evidence of absence and does not authorize destructive inference. Missing base means keep; corrupt/unreadable base fails before writes. Unrelated files and ledger entries remain untouched. Remove the old directory only if empty, never recursively. Repeated updates must preserve each disposition. Uninstall handles both distribution shapes and keeps unresolved retired files and recovery records rather than clearing the originals directory wholesale.
5. **Preserve existing symlink policy with a conservative retirement boundary.** [0013](0013-installer-follows-directory-symlinks.md) still governs existing installation/update writes. New retirement must not delete or overwrite external target bytes. A symlinked parent in a retired file or its recovery path is a keep-with-notice case, rather than permission to restore through that parent. Use existing atomic replacement precautions at a file symlink; do not introduce a broad path-policy rewrite. Fixtures must prove the boundary, including backup/metadata paths. The implementation may record narrow restoration intent bound to the allowlisted path, recorded projection digest, original digest and preexisting ownership before restoring. Only this evidence permits cleanup of an already-restored retry; original equality alone does not override edited-file retention, and intent cannot replace missing base evidence.
6. **Proof and recovery are part of the transition.** Installer cases M01–M11 assert bytes, notices and metadata through repeat update and uninstall. Installed consumer cases M12–M14 require actual host traces of bounded work and durable interruption/resume, with isolated configuration and installed skills outside the source tree. Contract tests, doc links and reference resolution cannot substitute for agent behavior. All four repository gates, including binary size, and an independent final review are required.

Rejected: removing embedding before retirement, deleting any file equal to today's upstream, clearing recovery with active targets, and silently falling back to stale procedures. Each either breaks the compatibility bridge or turns missing provenance into authority. Adding a skill installer or lifecycle runtime to the binary is outside its boundary.

## Consequences

Consumers must install the six-skill dependency set independently. Conservative retirement can leave edited or uncertain legacy files and backups; notices explain why they remain, while canonical routing stops using them. The compatibility phase temporarily duplicates procedures and therefore requires byte parity until the atomic transition.

Source rollback does not recover consumer files already removed or restored. Before upgrade, capture the full consumer tree, hidden metadata and host configuration. Restore that snapshot offline when necessary; do not run an old fresh-overwriting updater over retained edits. Source recovery reverses dependent changes in reverse order, restoring the compatibility bridge before the original layout. See the [consumer migration guide](../guides/workflow-harness/skill-procedure-migration.md).

## Authority

- Owner execution request, 2026-10-13: safe migration first, lean SDLC only after accepted migration, no commit/publication/global installation without approval.
- `docs/plans/completed/playbooks-to-skills.md:59` — fixed retirement inventory, M01–M14 and metadata postconditions accepted by the owner; this ADR records those decisions before code changes.
- Pre-migration revision `156bc96b15715ddd8b43e69bbeaa4bcd0fff8c1a`, `cli/src/installer/mod.rs:70` and `cli/src/installer/ownership.rs:148` — original target/legacy-seeding coupling.
- Same pre-migration revision, `cli/src/installer/uninstall.rs:63` — original blanket recovery cleanup that must change for retired targets.
- [0008](0008-recorded-ownership-and-transactional-recovery.md) — recorded ownership, unknown means keep, original preservation.
- [0013](0013-installer-follows-directory-symlinks.md) and [0014](0014-outcome-spec-normal-plans.md) — preserved policy boundaries.
