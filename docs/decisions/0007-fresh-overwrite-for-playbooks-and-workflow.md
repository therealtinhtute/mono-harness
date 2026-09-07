# 0007 — Playbooks and WORKFLOW.md fresh-overwrite; three-way merge narrows to PROJECT.md and the AGENTS.md block

## Status

Accepted. 2026-09-03. Narrows R9 of `docs/plans/completed/zharness-v015-slim.md` for two of the four managed-file classes; R9 stays accurate for `docs/PROJECT.md` and the `AGENTS.md` marked block.

## Context

`zharness update`'s three-way merge (R9, sourced from prior art: `hoangnb24/repository-harness`) treated every whole-file managed target the same way: `local == oldBase` fast-forwards, `local == newUp` is a no-op, anything else three-way-merges and can conflict. A fourth case existed for drift with no recorded ancestor — `.zharness/base/` missing or predating v0.15 tracking — which the code (R18) refused to touch at all, leaving the file "kept (local edits beyond recorded history)" forever, with no version bump ever reconciling it.

The consumer-facing symptom: a consumer repo could run `zharness update` and see a stage playbook stay on old content indefinitely. Investigation of `cli/internal/installer/{installer,update}.go` found this was not a bug in the merge logic — it was the merge logic doing exactly what R9 specified, applied to files that were never a legitimate customization surface in the first place. Playbooks and `docs/WORKFLOW.md` are pure mirrors of this repo's own `cli/docs/embedded/` content; nothing in the harness model expects a consumer to hand-patch them. `docs/PROJECT.md`, by contrast, is a one-time scaffold the consumer is expected to fill in with real project identity, and the `AGENTS.md` marked block sits inside a file consumers also write their own prose around — both are genuine customization surfaces.

## Decision

Split `Target` (`cli/internal/installer/installer.go`) on a new `Merge bool` field. `AllTargets()` sets it `true` only for the `docs/PROJECT.md` scaffold target; every playbook and `docs/WORKFLOW.md` default to `false`.

- `Merge: false` (playbooks, `docs/WORKFLOW.md`): `install` and `update` always overwrite with upstream bytes unconditionally — no read-and-compare, no diff, no conflict possible, no "kept beyond recorded history" state. A consumer who deletes one of these files locally gets it recreated on the next update; a consumer who hand-edits one loses the edit silently on the next update, with no warning.
- `Merge: true` (`docs/PROJECT.md`, `AGENTS.md` block): unchanged — R9's three-way merge, conflict markers, `--continue`/`--abort`.

Rejected: applying fresh-overwrite to the whole managed set (would silently erase `docs/PROJECT.md`'s consumer-authored identity content and any AGENTS.md prose around the block — those are the two surfaces R9 exists to protect). Rejected: an opt-in `--force`/`--fresh` flag with three-way merge staying the default for playbooks — the owner's call was that playbooks are never a legitimate customization surface, so there is nothing worth defaulting to protect.

## Consequences

- A consumer repo whose `.zharness/base/` was lost, or predates base tracking, now self-heals on the next `update` for playbooks/WORKFLOW.md instead of staying stuck forever.
- Any consumer who *had* hand-patched a playbook or `WORKFLOW.md` loses that patch on their next `update`, unrecoverable except via their own git history — no conflict marker warns them first. Playbooks were never meant to be hand-edited, so this is the accepted cost, not an oversight.
- `--abort` still restores fresh-overwritten files byte-for-byte, because `stashCapture` runs over every target before the per-file branch regardless of `Merge`.
- `install`'s brownfield report can no longer say "drifted ... left untouched; `zharness update` will merge" for a playbook — that message is now reachable only for `docs/PROJECT.md`.

## Authority

- `cli/internal/installer/installer.go:48-51,65-75,312-343` — the `Target.Merge` field and its two call sites.
- `cli/internal/installer/update.go:262-285,530-533` — the fresh-overwrite branch and the `classify()` label fix.
- `docs/plans/completed/zharness-v015-slim.md` — R9 (original three-way-merge decision), R18 (no fabricated ancestor).
- Owner's call, this session, 2026-09-03: scope confirmed as playbooks + WORKFLOW.md only, default behavior (no flag), no equality check before overwrite.
