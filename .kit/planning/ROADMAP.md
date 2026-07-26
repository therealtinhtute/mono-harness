# ROADMAP: Harness Convergence Pass v3

## Planning Basis
- source spec: `.kit/planning/SPEC.md`
- planning mode: `full`
- entry phase: **universal-preflight**
- execution mode: `work full`
- lane: high-risk
- sequencing: linear; each phase is independently mergeable and leaves a usable harness

## Phase 1: universal-preflight
**Status:** ready

**Goal:** Every active workflow skill receives deterministic CLI routing from one read-only preflight contract while the current layout remains operational.

**Deliverables:**
- `zharness preflight <stage> [--mode] --json` with ready/reduced/blocked routing.
- Table-driven stage/mode/DB/docs matrix and zero-write proof.
- Eight workflow skill adapters use preflight; no stage invents readiness behavior.

**Dependencies:**
- locked v3 SPEC only

**Risks / Watch-fors:**
- preflight must not mutate files or DB, including when state is missing.
- current skill UX and existing full lifecycle must remain usable before the layout migration.

## Phase 2: root-layout
**Status:** ready

**Goal:** Move to one root database and root managed docs while preserving `.kit/changesets` and the current lifecycle artifact chain.

**Deliverables:**
- root `harness.db` as the only DB; `.kit/changesets` unchanged.
- root docs projection, marked AGENTS block, `managed_docs` hash guard.
- `zharness migrate layout --to v2 [--dry-run]` with replay/parity/rollback behavior.
- dogfood migration of this repository without deleting legacy lifecycle markdown yet.

**Dependencies:**
- universal-preflight

**Risks / Watch-fors:**
- old DB stays active until temporary root DB replay and normalized parity pass.
- root docs may be consumer-owned; conflicts must stop rather than overwrite.

## Phase 3: one-plan-lifecycle
**Status:** ready

**Goal:** Replace parallel SPEC/ROADMAP/phase/RUN/CHECK/HANDOFF files with one evolving durable plan while retaining typed DB lifecycle guards.

**Deliverables:**
- `docs/plans/active/{slug}.md` lifecycle and plan scaffold.
- DB commands no longer require run/check/handoff markdown paths.
- automatic story status transitions through `done`.
- legacy history consolidated, verified, and obsolete artifact trees removed with `trash`.

**Dependencies:**
- root-layout

**Risks / Watch-fors:**
- bounded work must remain zero-write.
- resume/validate/watzup/git must work after report files disappear.

## Phase 4: docs-first-contract
**Status:** ready

**Goal:** Rewrite the managed entrypoint, workflow map, stage playbooks, and public contracts around upstream’s docs-first authority model plus the intentional local CLI guard.

**Deliverables:**
- AGENTS managed block + WORKFLOW combined ≤1,000 words.
- concise stage playbooks with no duplicated global policy or dead commands.
- updated CLI/state/schema/migration docs and workflow README.
- final clean tree, full Go tests, live-binary smoke checks, and migration evidence.

**Dependencies:**
- one-plan-lifecycle

**Risks / Watch-fors:**
- docs must not claim harness rows prove product behavior.
- no CI workflow, compaction, updater, extra DB, or invented consumer runbook may enter scope.
