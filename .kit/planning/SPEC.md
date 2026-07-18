---
id: 01KXSS59QBF2GFF968WGNRZ5ZA
type: spec
phase: none
lane: high-risk
intake_id: 01KXSS7DWDT03WF2N70QRGWWAR
created: 2026-07-18
updated: 2026-07-18
---

# SPEC: Agent-agnostic workflow chain — embedded playbooks, thin-trigger skills

Status: locked
Input Type: new-initiative
Lane: high-risk
Risk Flags: existing-behavior, public-contract, cross-platform
Affected Surfaces: docs, db
Downstream: to-plan full
Updated At: 2026-07-18

## Source Mode
idea

## Source Inputs
- Session analysis: gap comparison between `skills/workflow/` + `zharness` and
  `hoangnb24/repository-harness` (upstream concept source)
- `.kit/planning/IDEA.md` (this initiative's raw idea + brainstorm decisions)
- `skills/workflow/README.md` (current 4-layer model), `cli/docs/STATE.md`
  (state contract), `docs/workflow-harness/migration.md` (init footguns)

## Scenario
feature bootstrap

## Goal
Invert the workflow chain's architecture so its operating logic lives in
canonical playbook docs embedded in the `zharness` binary — written into each
project by `zharness init` — instead of inside Claude Code `SKILL.md` files.
Skills become thin triggers. Any agent (Codex, Cursor, Claude Code) can then
operate the full lifecycle by reading the docs and calling `zharness`.

## Users / Actors
- The repo owner running the chain daily via Claude Code skills
- Non-Claude coding agents (Codex, Cursor, others) entering via `AGENTS.md`
- Anyone installing `skills/workflow/*` from this public repo
- Future contributors extending `zharness`

## Requirements
1. `zharness` embeds one canonical playbook doc per spine stage (brainstorm,
   to-plan, work, check, handoff, watzup) plus `AGENTS.md` shim,
   `CONTEXT_RULES.md`, and a request-class authority doc. Embedded content is
   compiled into the binary (`go:embed`), not fetched at runtime.
2. `zharness init` scaffolds the project surface: creates `.kit/` (fixing the
   documented `db_not_writable` footgun), writes the playbook docs and
   `AGENTS.md` shim, and ensures `.gitignore` covers only `harness.db` and
   `.kit/cache/`.
3. Written docs carry a `docs_version` stamp recorded in the harness meta.
   `zharness resume` reports drift type `stale_docs` (with a named recovery:
   `zharness init --refresh-docs`) when the stamp is behind the running CLI's
   docs version; `init --refresh-docs` rewrites the docs and updates the stamp
   without touching other state.
4. Each playbook is self-sufficient for a competent agent: stage purpose,
   preconditions, exact `zharness` commands with arguments, artifact paths and
   templates, and exit/handoff conditions — no reference back to any
   `SKILL.md` content.
5. `CONTEXT_RULES.md` maps each lifecycle stage to exactly the docs it must
   read (including `git`'s single `query check --latest` step), so an agent
   never needs to over-read.
6. The request-class authority doc codifies read-only vs change requests:
   read-only requests (answer, review, status) must not run `init`, `intake`,
   `trace`, or any mutating command.
7. The 6 spine `SKILL.md` files are rewritten as thin triggers (target ≤ 30
   lines each): trigger conditions + version gate + "read playbook X, follow
   it". Their `references/` content is either absorbed into the embedded
   playbooks or deleted; no workflow logic remains in the skill layer.
8. `git` and `interview` skills stay as-is (minimal-integration); their
   SKILL.md files are not rewritten in this initiative.
9. A pilot proves agent-agnosticism: at least one lifecycle pass
   (intake → story → trace → check record) driven by a second, non-Claude
   agent using only the written docs + CLI, with the evidence recorded under
   `docs/workflow-harness/pilot-evidence/`.
10. Existing behavior is preserved for Claude Code users: the 6 skills keep
    their names, trigger semantics, and version gate; a chain run end-to-end
    on this repo after the rewrite produces the same artifact chain
    (SPEC → ROADMAP → runs → check reports → HANDOFF) as before.
11. `MIN_ZHARNESS_VERSION` is bumped to the release that ships embedded docs;
    thin-trigger skills gate on it.

## Boundaries
### In Scope
- `cli/`: embedded docs, `init` scaffolding + `--refresh-docs`, `stale_docs`
  drift in `resume`, docs_version in meta (schema migration if needed)
- Rewrites of 6 spine `SKILL.md` files + pruning their `references/`
- New embedded doc set: 6 playbooks + `AGENTS.md` shim + `CONTEXT_RULES.md`
  + authority doc
- Second-agent pilot + evidence doc
- Updates to `skills/workflow/README.md`, root `README.md`, `CLAUDE.md`,
  `docs/workflow-harness/migration.md` reflecting the inversion

### Out of Scope
- `git` / `interview` skill rewrites
- Any skill outside `skills/workflow/`
- New harness entities or lifecycle stages (this changes where logic lives,
  not what the lifecycle is)
- Mechanical story verification, backlog outcome loop, tool registry, propose
  activation (separate initiatives — see Deferred Ideas)
- Markdown fallback / CLI-optional mode (still explicitly rejected)
- Multi-language doc localization

## Constraints
- Go binary, cobra, `modernc.org/sqlite`, CGO disabled — unchanged toolchain
- Changesets remain append-only; docs_version lands in meta via the existing
  changeset mechanism, never by editing committed changesets
- Docs must render as plain markdown (no agent-specific syntax) so any runtime
  can consume them
- The chain must stay operational on this repo throughout (it is the
  daily-driver); skill rewrites land only after the CLI that supports them is
  released

## Acceptance Criteria
- Fresh directory: `zharness init` alone yields `.kit/` + docs + shim +
  correct `.gitignore`, and `resume --json` reports `readiness: clean` with a
  recorded docs_version
- Upgrading the CLI then running `resume` on a project with older docs reports
  `stale_docs` drift naming `zharness init --refresh-docs`; running that
  recovery clears the drift and rewrites only docs
- Each of the 6 spine SKILL.md files is ≤ 30 lines and contains no workflow
  step content — verified by review against R7
- Second-agent pilot evidence exists: the non-Claude agent completed
  intake → story → trace → check record on a sample task without reading any
  SKILL.md, and `zharness validate --json` passes on the produced chain
- A full chain pass on this repo (Claude Code) after the rewrite produces a
  valid artifact chain with zero `pointer_drift` in `audit --json`

## Validation Expectations
- Unit: Go tests for embed integrity (every declared doc present, non-empty,
  version-stamped), init scaffolding idempotency, `--refresh-docs` rewrite
  semantics, `stale_docs` drift trigger and clearing
- Integration: end-to-end `init → intake → story → trace → check record →
  resume` on a scratch directory, asserting scaffolded files and drift states
- Platform: goreleaser build across existing target matrix; install script
  still resolves the release
- Pilot: recorded second-agent session as e2e evidence (R9)

## Dependencies / Assumptions
- Depends on `zharness` v0.1.0 release pipeline already shipped (cli/v0.1.0)
- Assumes a second agent runtime (Codex CLI or Cursor) is available to the
  repo owner for the pilot phase
- Assumes current playbook content can be distilled from the existing 6
  SKILL.md files + references without semantic loss (they are the de facto
  spec of each stage today)
- Filed gaps #24/#25 remain separate work items; this initiative must not
  regress them further

## Key Decisions
- **Embed docs in the binary** (vs installer-copy from the skills repo, vs
  shim pointing at the repo): version-locked with the CLI, single source of
  truth, immune to the #24 doc/code drift class. Installer-copy rejected — N
  drifting copies needing merge tooling; repo-pointer rejected — breaks
  offline/private use.
- **6 spine skills only** (vs all 8): `git`/`interview` carry no workflow
  logic worth embedding; forcing them in bloats the binary with generic
  content. Their minimal-integration status is already the documented design.
- **Version-stamp + `stale_docs` drift** (vs write-once at init): reuses the
  existing drift/recovery mechanism; silent doc staleness is the exact
  failure mode this repo already hit (#24).
- **Real second-agent pilot** (vs structural review): agent-agnosticism is
  the initiative's whole point; without a live pass it stays a claim
  (weak-proof flag would otherwise apply).
- **Thin triggers keep the skills** (vs deleting the skill layer): Claude
  Code users keep discovery, trigger semantics, and the version gate; the
  cost (≤ 30 lines each) is negligible and carries no logic to drift.

## Open Questions
- Which second agent runs the pilot (Codex CLI vs Cursor) — decide at the
  pilot phase based on availability
- Whether `AGENTS.md` shim install should offer a `--claude` style CLAUDE.md
  block like upstream, or leave CLAUDE.md untouched — default: leave
  untouched, revisit at planning

## Deferred Ideas
- Mechanical story verification (`verify_command` on stories, `story verify`)
  — borrowed from upstream, separate initiative
- Backlog predicted/outcome loop and `propose` activation
- Tool registry (capability + degrade ladder)
- Auto-scoring traces at record time
- sha256 checksum + `--dry-run` for `install-zharness.sh`

## Ambiguity Report
- Goal clarity: high — inversion target and end state are concrete
- Scope clarity: high — skill list, doc set, and exclusions enumerated
- Constraints clarity: high — toolchain and changeset rules carried over
- Acceptance clarity: high — every criterion maps to a requirement and is
  mechanically or pilot-checkable
