# ROADMAP: Agent-agnostic workflow chain — embedded playbooks, thin-trigger skills

Entry phase: `playbook-authoring` · Execution mode: phase-by-phase via `work`, gated by `check`
Spec: `.kit/planning/SPEC.md` (intake `01KXSS7DWDT03WF2N70QRGWWAR`, lane high-risk)

## Planning Basis
- source spec: `.kit/planning/SPEC.md`
- planning mode: `full`

## Phase 1: playbook-authoring
**Goal:** Author the canonical embedded doc set — 6 stage playbooks + `AGENTS.md` shim + `CONTEXT_RULES.md` + `AUTHORITY.md` — under `cli/docs/embedded/`, distilled from the 6 spine SKILL.md files + references with no semantic loss.

**Deliverables:**
- `cli/docs/embedded/playbooks/{brainstorm,to-plan,work,check,handoff,watzup}.md` (R1, R4)
- `cli/docs/embedded/AGENTS.md` (shim), `CONTEXT_RULES.md` (R5), `AUTHORITY.md` (R6)
- Command-accuracy pass: every `zharness` invocation in the playbooks verified against the real CLI surface

**Dependencies:** none — pure content phase; unblocks everything else

**Risks / Watch-fors:**
- Distillation loss: a playbook silently dropping a contract the SKILL.md enforced (e.g. watzup's output-contract forbidden phrases)
- Writing Claude-specific phrasing into docs meant for any agent

## Phase 2: cli-embed-scaffold
**Goal:** `zharness` embeds the doc set (`go:embed`) and `init` scaffolds the full project surface: `.kit/` creation, docs + shim write-out, `.gitignore` management, `docs_version` stamped in meta, `init --refresh-docs`.

**Deliverables:**
- Embed package with integrity unit tests (R1)
- `init` scaffolding + idempotency + `--refresh-docs` (R2, R3 partial)
- `docs_version` in meta via changeset (schema migration if needed)

**Dependencies:** Phase 1 (the doc set is the embed source)

**Risks / Watch-fors:**
- Schema/meta change must go through a versioned migration + changeset, never a hand-edit
- `init` must stay idempotent for existing projects (no clobbering a live `.kit/`)

## Phase 3: cli-stale-drift
**Goal:** `resume` reports `stale_docs` drift with the named recovery `zharness init --refresh-docs` when written docs lag the CLI; recovery clears the drift; a scratch-dir integration test covers `init → intake → story → trace add → check record → resume`.

**Deliverables:**
- `stale_docs` drift detection + recovery string, kept in sync with `cli/docs/STATE.md` (R3)
- Integration test suite on a scratch directory

**Dependencies:** Phase 2 (`docs_version` exists)

**Risks / Watch-fors:**
- Recovery string drift between `resume.go` and STATE.md — the exact #24 failure class; single-source the string
- Drift must not fire on projects that predate embedded docs in a way that blocks them — `no docs_version` needs a defined behavior

## Phase 4: cli-release
**Goal:** Ship `cli/v0.2.0` with embedded docs through the existing tag → goreleaser pipeline; `install-zharness.sh` resolves the new release.

**Deliverables:**
- Tagged, published release with all platform assets (R11 prerequisite)
- Verified install path from a clean machine/dir

**Dependencies:** Phase 3 (feature-complete CLI)

**Risks / Watch-fors:**
- The `cli/vX.Y.Z` trigger-tag vs bare-semver goreleaser quirk (already documented in migration.md) — reuse the proven flow, don't reinvent

## Phase 5: thin-triggers
**Goal:** Rewrite the 6 spine SKILL.md files as ≤30-line thin triggers gating on the new MIN_ZHARNESS_VERSION; prune absorbed `references/`; update repo docs; prove Claude-chain parity on this repo.

**Deliverables:**
- 6 thin SKILL.md files, references pruned (R7); `git`/`interview` untouched (R8)
- MIN_ZHARNESS_VERSION bumped in `skills/workflow/README.md` (R11)
- Updated root `README.md`, `CLAUDE.md`, `docs/workflow-harness/migration.md`
- Full-chain pass on this repo with zero `pointer_drift` in `audit` (R10)

**Dependencies:** Phase 4 (skills gate on a released version)

**Risks / Watch-fors:**
- The chain is the daily driver — land skill rewrites atomically per skill, verify each against its playbook before moving on
- ~/.claude installed copies vs repo copies must be resynced after rewrite

## Phase 6: agent-pilot
**Goal:** A second, non-Claude agent (Codex CLI or Cursor — decided at phase start) completes one lifecycle pass `intake → story → trace add → check record` on a sample task using only the written docs + CLI; evidence published.

**Deliverables:**
- `docs/workflow-harness/pilot-evidence/{date}-second-agent-pilot.md` (R9)
- `zharness validate --json` pass on the produced chain

**Dependencies:** Phase 4 (released CLI + docs). Can run in parallel with Phase 5 — the pilot must not read any SKILL.md by design.

**Risks / Watch-fors:**
- Pilot failures are findings, not hotfix targets — file gaps and route them to the next cycle unless they block the acceptance criterion itself

## Phase 7: harness-mode-parity
**Goal:** Close the gap Phase 6's pilot exposed (GitHub #38 and its check-side twin): `work`/`check` simple mode currently attempts DB registration against a schema that has no concept of a story-less run or run-less check (`runs.story_slug` and `checks.run_id` are both `NOT NULL` FKs), and `validate` has no mode-awareness so it flags the resulting gaps as hard failures. Make `zharness validate --json` capable of returning `valid:true` on a simple-mode-produced chain for the first time in this repo's history, without weakening full-mode validation.

**Deliverables:**
- `work.md`: RUN artifact frontmatter gains `mode: full|simple`; Step 2 branches by mode — full mode registers via the existing two-line changeset (unchanged), simple mode skips DB registration entirely (no story to satisfy the FK) and records why inline
- `check.md`: persisted report frontmatter gains `mode: full|simple` (inherited from the RUN it gates); Step 4 skips `zharness check record` when the gated run is simple-mode (no run row to FK against)
- `validate.go`: reads `mode` from RUN/CHECK frontmatter; `mode: simple` exempts phase-existence, `plan_id` ULID, and DB stale-pointer checks (artifact hygiene — `id` must still be a well-formed ULID — still enforced); `mode: full` or absent keeps today's strict behavior unchanged
- `CONTRACT.md`: `validate`'s documented issue enum gains `not_yet_implemented` (already emitted by code, undocumented — pre-existing drift); new mode-aware carve-out documented
- New CLI release past v0.2.0 through the proven tag → goreleaser pipeline; `MIN_ZHARNESS_VERSION` bumped if the fix changes required playbook behavior

**Dependencies:** Phase 6 (`agent-pilot` surfaced #38 as the trigger; this phase's `-CONTEXT.md` supersedes agent-pilot's Forbidden Surface on `cli/**` for this scope only)

**Risks / Watch-fors:**
- Must not loosen full-mode validation — every change is additive/mode-gated, guarded by regression tests on existing full-mode artifacts
- The check-side twin (no CLI path for ad-hoc/simple-mode checks, already backlogged as `01KXWH4YNC9RRFR1VPE6DK8P14`) must be resolved by the same mode-aware design, not patched separately
- Old artifacts predating the `mode` field must default to full-mode strictness, not silently pass

## Phase 8: agent-pilot-rerun
**Goal:** Re-run the second-agent pilot (same protocol as Phase 6, Codex CLI) against the Phase 7 release to confirm R9's literal acceptance bar is now met: `zharness validate --json` returns `valid:true` on a genuinely cold, simple-mode-produced chain.

**Deliverables:**
- `docs/workflow-harness/pilot-evidence/{date}-agent-pilot-rerun.md` — new dated evidence doc, citing `validate --json` output showing `valid: true`
- R9 formally closed (GO verdict) or a new, distinct blocking finding routed for a further cycle

**Dependencies:** Phase 7 (fixed CLI + docs must be released first)

**Risks / Watch-fors:**
- Same Forbidden Surfaces as Phase 6: scratch target + `pilot-evidence/` only, read-only elsewhere, no `cli/**`, no `skills/workflow/**`, no this repo's live `.kit/` — pilot failures route to a new finding, never a live hotfix mid-pilot
