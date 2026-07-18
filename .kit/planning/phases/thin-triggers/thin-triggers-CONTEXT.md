# Context: thin-triggers

Phase: thin-triggers
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: high
Expected Proof: inspection (≤30-line + no-logic review), e2e (full Claude-chain pass on this repo)

## Goal
Rewrite the 6 spine SKILL.md files as thin triggers gating on the new MIN_ZHARNESS_VERSION; prune absorbed references; update repo docs; prove the Claude Code chain still produces a valid artifact chain (R10).

## Scope Boundary
### Allowed Surfaces
- `skills/workflow/{brainstorm,to-plan,work,check,handoff,watzup}/**` (SKILL.md rewrite, references pruning)
- `skills/workflow/README.md` (MIN_ZHARNESS_VERSION, model description)
- Root `README.md`, `CLAUDE.md`, `docs/workflow-harness/migration.md`
- `~/.claude/skills/` resync (installed copies)

### Forbidden Surfaces
- `skills/workflow/git/**`, `skills/workflow/interview/**` (R8 — verbatim untouched)
- `cli/**` (CLI is frozen at v0.2.0 for this phase; playbook bugs found here get filed, not hotfixed)
- Any skill outside `skills/workflow/`

## Spec Hooks
- R7 (≤30-line thin triggers, no workflow logic), R8, R10 (chain parity), R11 (MIN bump)
- Constraint: chain stays operational — rewrite and verify one skill at a time

## Locked Decisions
- Thin-trigger canonical shape (uniform across the 6): frontmatter (name/description unchanged for skills.sh compatibility) → version gate line (MIN_ZHARNESS_VERSION = 0.2.0) → "ensure docs present: run `zharness init` if `.kit/docs/` missing" → "read `.kit/docs/playbooks/{stage}.md` and follow it" → defer-to list (1 line)
- ≤30 lines counts rendered lines of SKILL.md including frontmatter
- references/ pruning rule: delete a reference file only if its content is verifiably inside the corresponding playbook (diff-checked during Phase 1); anything not absorbed stays and is a finding against playbook-authoring
- Trigger semantics (description frontmatter, when-to-use) stay Claude-facing — that is the legitimate skill-layer content
- Order of rewrite: watzup first (read-only, safest), then handoff, check, work, to-plan, brainstorm — reverse-lifecycle so the entry skills flip last
- Resync `~/.claude/skills/` after each skill lands (repo is source of truth)

## Assumptions
- skills.sh install flow only needs frontmatter intact; body length is unconstrained by the format
- R10's "same artifact chain" means structural parity (all artifacts present, cross-linked, audit-clean), not byte-identical prose

## Canonical Refs
- `.kit/docs/playbooks/**` (as scaffolded by the released v0.2.0 — the docs the triggers point at)
- `skills/workflow/README.md` (version-gate contract)
- SPEC R7 wording

## Rejected Options
- Deleting the skill layer entirely — loses Claude discovery/trigger UX for zero gain (SPEC key decision)
- Keeping references/ as duplicated backup — two sources of truth, the exact drift class this initiative kills

## Deferred Ideas
- Thin-trigger equivalents for git/interview; automated ≤30-line lint in CI

## Escalate If
- A playbook proves insufficient mid-rewrite (agent needs SKILL.md content the playbook lacks) → file against playbook-authoring, pause that skill's rewrite; if systemic → to-plan phase
- Chain parity run surfaces a CLI defect → file issue; only block if it breaks the R10 acceptance criterion
