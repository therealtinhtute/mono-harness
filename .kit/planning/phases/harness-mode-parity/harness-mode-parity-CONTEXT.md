# Context: harness-mode-parity

Phase: harness-mode-parity
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: medium
Expected Proof: unit, integration, command-output

## Goal
Make `zharness validate --json` return `valid:true` on a simple-mode-produced chain, without weakening full-mode validation. Root cause (GitHub #38 + its check-side twin, backlog `01KXWH4YNC9RRFR1VPE6DK8P14`): the DB schema has no concept of a story-less run or run-less check (`runs.story_slug` and `checks.run_id` are both `NOT NULL` FKs, no `mode` column anywhere), yet `work.md`/`check.md` instruct simple mode to attempt the same DB registration full mode does, and `validate.go` has no mode-awareness so it flags the resulting gaps (`phase: none` → broken_link, `plan_id: none` → missing_key, missing DB row → stale_pointer) as hard failures instead of expected simple-mode shape.

## Scope Boundary
### Allowed Surfaces
- `cli/internal/application/validate.go` (+ its `_test.go`)
- `cli/docs/embedded/playbooks/work.md`, `cli/docs/embedded/playbooks/check.md`
- `cli/docs/CONTRACT.md`
- `cli/` release surface: version bump, tag, goreleaser trigger (reuse `cli-release`'s proven flow)
- `skills/workflow/README.md` (`MIN_ZHARNESS_VERSION`, only if the fix requires a floor bump)

### Forbidden Surfaces
- The 6 thin-trigger `SKILL.md` files themselves (only the `MIN_ZHARNESS_VERSION` reference in README.md changes, not the skill files)
- `git` / `interview` skills
- Any DB schema migration that adds nullable FKs or a `mode` column — rejected here, see Rejected Options
- This repo's own `.kit/planning/SPEC.md` / `ROADMAP.md` structure beyond this amendment already made by `to-plan`

## Spec Hooks
- R9 (agent-agnosticism pilot) — acceptance criterion's second clause: `zharness validate --json` passes on the produced chain. This phase is the only thing standing between the current NO-GO and a literal pass.
- R4 (each playbook self-sufficient, no reference back to SKILL.md) — the `work.md`/`check.md` edits must stay playbook-native, no new SKILL.md content
- Constraint: "Changesets remain append-only... docs_version lands via changeset, never hand-edit" — any DB-affecting change here goes through the same changeset mechanism, not a schema migration (see Rejected Options)

## Locked Decisions
- **Fix at the data-path + validator layer, not the DB schema.** Simple mode stops attempting DB registration entirely (no changeset authored for simple-mode runs/checks) rather than loosening `runs.story_slug`/`checks.run_id` to nullable. `validate.go` gains mode-awareness instead of forcing simple-mode artifacts to satisfy full-mode FK/link shape.
- **`mode` field added to RUN and CHECK artifact frontmatter** (not just the existing body `Mode: full | simple` line in RUN, which `validate.go`'s `parseFrontmatter` cannot see — it only reads the `---` fenced block). This is the only new field introduced.
- **Old artifacts with no `mode` field default to full-mode strictness** — never silently loosen historical/ambiguous data. Only an explicit `mode: simple` triggers the carve-out.
- **Artifact hygiene stays universal**: even in `mode: simple`, `id` must still be a well-formed ULID. Only phase-existence, `plan_id` ULID, and DB stale-pointer checks are mode-gated.
- **`not_yet_implemented` added to CONTRACT.md's documented issue enum** — it's already emitted by `validate.go` (line 67) for the SPEC→PLAN gap; CONTRACT.md's enum (`missing_key|broken_link|stale_pointer`) was already stale before this phase. Fixed as part of documenting the new mode carve-out, not a separate finding.
- **Check-side twin resolved by the same mechanism**: `check.md`'s Step 4 skips `zharness check record` when the gated RUN's `mode` is `simple` — same shape as the RUN-side skip, not a separate design.
- **Ship a new release** through the existing `cli/vX.Y.Z` tag → goreleaser pipeline (Phase 4's proven flow, not reinvented).

## Assumptions
- Fixing #38 and its check-side twin is sufficient to reach `valid:true` on the pilot's exact chain shape (RUN with `phase: none`/`plan_id: none`, an ad-hoc CHECK report referencing it, no HANDOFF.md in the scratch target). If Phase 8's re-pilot surfaces a chain shape this phase didn't anticipate, that is a new finding routed back here or to a further cycle — not silently patched mid-pilot (Phase 8 stays read-only on `cli/**`).
- `MIN_ZHARNESS_VERSION` bump is likely needed (old CLIs would still emit the crash/false-negative) but the exact version-bump decision belongs to the implementing task, not pre-locked here.
- **Correction (recorded during T7 execution)**: the Forbidden Surfaces line below originally assumed the 6 spine `SKILL.md` files "reference README.md's constant" symbolically. On inspection they hardcode the literal version string per-file (`MIN_ZHARNESS_VERSION (\`0.2.0\` — see skills/workflow/README.md)`), same pattern the `thin-triggers` phase itself used when bumping `0.1.0`→`0.2.0`. Bumping only README.md's prose would leave every skill's actual gate check still passing a `0.2.0` (buggy) binary — the opposite of the bump's purpose. Corrected: the 6 spine `SKILL.md` files' hardcoded version strings are in scope for this one mechanical string replacement (`0.2.0`→`0.3.0`), `interview/SKILL.md`'s separate stale `0.1.0` stays untouched (pre-existing, out of scope per SPEC R8, already backlogged).

## Canonical Refs
- `.kit/planning/SPEC.md` (R9), `.kit/planning/ROADMAP.md` (Phase 7/8)
- `cli/internal/application/validate.go`, `cli/internal/infrastructure/migrations.go` (schema)
- `cli/docs/embedded/playbooks/work.md` (Execution Loop Step 2, Artifacts), `check.md` (Step 4, Artifacts)
- `cli/docs/CONTRACT.md` (`validate` entry)
- GitHub issue #38; backlog item `01KXWH4YNC9RRFR1VPE6DK8P14` (check-side twin); related but distinct: #36 (`plan_id` missing_key, `query phases --json` gap), #30 (check.md Step 4 gating-logic gap for simple mode — this phase resolves #30's root cause too)
- `docs/workflow-harness/pilot-evidence/2026-07-19-second-agent-pilot.md` (the pilot run that surfaced this)

## Rejected Options
- **Nullable `story_slug`/`run_id` + schema migration**: would let simple-mode rows exist in `runs`/`checks` without a story/run to link to, but breaks the FK's actual purpose (every DB row should trace to a real story) and requires a migration + changeset-replay story for existing DBs. Rejected — the data-path fix (don't register what has nothing to register against) is smaller and doesn't touch the schema's invariants.
- **Loosen R9's acceptance bar instead of fixing the harness** (the "Loosen R9" option presented to and rejected by the user this session): rejected — user explicitly chose the full-fix route.
- **Have simple mode call `zharness story` to synthesize a placeholder story**: considered in GitHub #38's suggested-fix list; rejected here because it pollutes the `stories` table with fake phase-less entries that `query phases --json` would then surface as if they were real roadmap phases — worse than not registering at all.

## Deferred Ideas
- A dedicated `zharness run register --mode simple` CLI command (vs. the current changeset-authoring convention) — would make the mode-branch explicit at the CLI layer instead of the playbook layer; deferred, current fix doesn't require a new command surface
- Backfilling `mode` onto historical RUN/CHECK artifacts that predate this phase — not required for R9, those chains already have their own recorded verdicts

## Escalate If
- The check-side twin turns out to need a genuinely different fix shape than the RUN-side one (not just "skip registration") → pause and re-run `to-plan phase harness-mode-parity` before continuing
- Fixing this exposes a validate finding class Phase 8's re-pilot chain doesn't map to any of the ones enumerated here → route the new finding to a further mini-phase rather than patching mid-pilot
