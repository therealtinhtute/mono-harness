# SPEC: workflow-harness — zharness CLI runtime for the workflow skill chain

Status: locked
Input Type: new-initiative
Lane: high-risk
Risk Flags: public-contract, data-model, existing-behavior, cross-platform, multi-domain
Affected Surfaces: api, db, docs
Downstream: to-plan full
Updated At: 2026-07-16

## Source Mode
files

## Source Inputs
- GitHub issue therealtinhtute/skills#23 (10-issue draft plan, Vietnamese)
- Revised 13-issue set from /think session 2026-07-16 (scratchpad: workflow-harness-issues.md)
- Reference implementation: `~/Lab/harness-experimental` (hoangnb24/repository-harness — Rust harness-cli, changeset/SQLite model, docs/HARNESS.md lifecycle)
- Current chain: `skills/workflow/{watzup,brainstorm,to-plan,work,interview,check,git,handoff}`

## Scenario
project bootstrap (new runtime layer inside existing repo)

## Goal
Evolve the workflow skill chain from prompt-only orchestration into a harness-backed runtime: a new Go CLI **`zharness`** (full port of harness-cli's surface plus workflow commands) becomes the mandatory durable layer — SQLite materialized from committed JSONL changesets under `.kit/` — and all workflow skills are rewritten CLI-first so intake, phases, traces, verdicts, and handoffs are machine-recorded, deterministically gated, and exactly resumable across sessions and machines. Workflow chain UX (skill order and intent) is preserved.

## Users / Actors
- **therealtinhtute** — repo owner, solo maintainer of skills + CLI.
- **Coding agents** (Claude Code and other skills.sh-compatible agents) — primary CLI callers; execute SKILL.md instructions.
- **Workflow chain users** — humans running the chain on their projects; install zharness via script, read markdown artifacts and recaps.

## Requirements

### A. Repo & runtime structure
1. Repo gains top-level `cli/` (Go module `github.com/therealtinhtute/skills/cli`, binary `zharness`) and `docs/workflow-harness/` (concept, gap matrix, migration guide). `skills/workflow/` layout is unchanged.
2. `cli/` mirrors upstream's four layers as Go packages: `cmd/zharness/main.go` + `internal/{interfaces,application,domain,infrastructure}` (`interfaces` because `interface` is a Go keyword; epoch-fence logic lives in `infrastructure`).
3. Per-project runtime lives in `.kit/`: `harness.db` (+`-wal`/`-shm`) gitignored; `changesets/*.changeset.jsonl` committed; planning/run/report markdown stays where it is today.
4. All contract docs live in `cli/docs/` (`CONTRACT.md` commands, `SCHEMA.md` database, `STATE.md` workflow-state semantics). No `docs/contracts/` directory.

### B. zharness CLI
5. Tech stack: cobra for the command tree; `modernc.org/sqlite` (pure Go, CGO_ENABLED=0).
6. Full command surface: `init, migrate, import, intake, story, decision, backlog, tool, intervention, trace, score-trace, score-context, audit, propose, db (changeset apply|status), query` ported from harness-cli, plus workflow additions `resume`, `check record`, `validate`. Every command supports `--json`.
7. Every mutating command appends a changeset before touching the DB; rebuilding the DB by replaying all changesets yields identical state; applying the same changeset twice changes nothing (idempotent).
8. One changeset file per command batch, named by ULID (`{ulid}.changeset.jsonl`); replay order = ULID order. No append-to-shared-file mode.
9. All entity IDs are ULIDs (single format across intake/story/trace/run/check/handoff).
10. `import` seeds a database from a legacy `.kit/` project: every `workflow-state.yml` field maps to a DB representation or is explicitly documented as dropped; planning markdown is linked by generated IDs.
11. `validate` walks SPEC→PLAN→RUN→CHECK→HANDOFF by frontmatter IDs, enforces required keys, cross-links, proof links, and pointer freshness; exits non-zero with machine-readable findings on violation.
12. `resume` emits a continuity snapshot: workflow position (phase status enum `planned|in-progress|checked|done`), latest run/check/handoff IDs, drift findings, and a named recovery action per finding.
13. SQLite schema adds workflow entities (phases, runs, checks, handoffs) alongside ported entities (stories, decisions, backlog, traces, intakes, interventions, tools); every table has a changeset entity type.
14. `workflow-state.yml` is retired: no skill or command reads or writes it after migration; its template file is removed from `to-plan/references/`.

### C. Artifact contracts
15. SPEC/PLAN/RUN/CHECK/HANDOFF templates gain required YAML frontmatter: stable ULID, artifact type, phase/story ref, intake lane, timestamps; markdown bodies stay human-first.
16. Cross-link rules: PLAN→SPEC id; RUN→PLAN id + trace ids; CHECK→RUN id + machine-readable proof links; HANDOFF→latest RUN/CHECK ids.

### D. Skills rewritten CLI-first
17. All workflow SKILL.md files are rewritten with `zharness` interaction as the explicit backbone of each step — commands inline in the flow, not tucked into references. The CLI is mandatory: each skill starts with a version gate (`zharness --version` ≥ documented minimum) and stops with install instructions when absent. No markdown fallback path.
18. Adapter behaviors: brainstorm runs `intake` at SPEC lock (lane persisted, intake ID in SPEC frontmatter); to-plan runs `init` + `story` per phase and records phase pointers; work emits `trace add` per wave with RUN linkage; check runs `audit` + `score-trace`, evaluates the proof matrix, records verdicts via `check record`; handoff records a handoff entity; watzup renders `resume` output.
19. check verdicts are deterministic: a validation matrix (intake lane × proof class) defines required evidence; missing required proof always yields FAIL naming the missing evidence.
20. watzup and handoff share unified readiness states (`clean | in-progress | drifted | no-harness`); `no-harness` routes to install/import guidance; drift always surfaces a specific recovery step.

### E. Distribution, docs, migration
21. Release pipeline (goreleaser + GitHub Actions) publishes darwin/linux × amd64/arm64 binaries to GitHub Releases; `scripts/install-zharness.sh` installs via `gh release download` (works against the private repo through existing gh auth).
22. Docs set: root README quickstart (install → init/import → run chain); `skills/workflow/README.md` concept doc (lifecycle Intent→Intake→Story/Plan→Trace→Proof→Handoff/Resume, 4-layer model, skill↔command↔entity mapping table); `docs/workflow-harness/` gap matrix + migration guide with rollback notes (markdown stays readable without CLI; DB always rebuildable from changesets) + contributor playbook (adding a command/contract without breaking existing changesets).
23. Repo `CLAUDE.md` and all workflow references are purged of `workflow-state.yml` semantics.
24. A pilot runs the full chain (intake→plan→work→check→handoff→cross-machine watzup resume) on a real task as soon as adapters + continuity land — before research-command polish — and files an issue per gap with a go/no-go verdict.

## Boundaries

### In Scope
- `cli/` Go module, release pipeline, install script
- Rewrites of all 8 `skills/workflow/*` SKILL.md files + their references
- Artifact template contracts (brainstorm/work/check/handoff references)
- `docs/workflow-harness/`, `skills/workflow/README.md`, root README, CLAUDE.md updates
- Legacy `.kit/` import path and migration guide
- Pilot run + evidence

### Out of Scope
- Repo restructure beyond adding `cli/` and `docs/workflow-harness/` (skills stay in place)
- Any skill outside `skills/workflow/` (craft/, shipping/ untouched)
- Markdown fallback / CLI-optional compatibility mode (explicitly rejected)
- New maintenance skill owning audit/propose (deferred)
- Homebrew tap, `go install` channel, Windows binaries
- Multi-user/concurrent-writer support for the DB (solo/agent use assumed)

## Constraints
- Go (single new language; approved), cobra, modernc.org/sqlite — no cgo anywhere in `cli/`.
- Private repo: every install/docs path must work through `gh` auth; no raw public URLs.
- skills.sh standard preserved: each skill directory remains independently installable.
- Solo maintainer: port scale is ~6k lines of upstream Rust infrastructure — build order must be layered (core → domain → research) so the chain can pilot before the full surface is polished.
- Changesets are append-only and committed; nothing may require editing a past changeset.

## Acceptance Criteria
- Clean machine: `install-zharness.sh` → `zharness --version` passes; legacy project: `init && import && query state --json` returns correct state derived from old `workflow-state.yml`.
- CI proves: changeset replay reproduces identical DB; double-apply is a no-op; `validate` fails a broken-crosslink fixture and passes the fixed one.
- Full chain on a sample project produces DB rows + changesets + artifacts whose IDs cross-reference (verified by `zharness validate`), with zero writes to `workflow-state.yml`.
- Kill a session mid-work, clone on another machine, install, run watzup: recap matches the last handoff exactly; a deliberately staled pointer yields a specific recovery step.
- Same check inputs always produce the same verdict; a fixture missing one required proof fails naming that proof.
- A new user completes install→import→watzup from docs alone, without prior chat context.

## Validation Expectations
- Unit: domain validation, changeset idempotency, ULID ordering, legacy-yml field mapping.
- Integration: init→import→query round-trip; replay-rebuild equality; validate fixtures (pass + fail).
- E2E: pilot chain run with cross-machine resume; evidence (changeset log, query outputs, audit report, recap) committed with the pilot.
- Platform: release artifacts for darwin/linux × amd64/arm64 built with CGO_ENABLED=0.

## Dependencies / Assumptions
- `~/Lab/harness-experimental` stays available as the port reference (source of truth for command semantics and changeset model).
- `gh` CLI authenticated on user machines (already required by the git skill).
- to-plan will decompose this spec into phases roughly along: concept/gap docs → contracts → CLI core+release → CLI domain → adapters → check gate → continuity → CLI research → docs/migration → pilot; phase independence per plan rules.
- Assumption (fragile, named): porting the Rust infrastructure to Go stays tractable for a solo maintainer. Mitigation baked into requirements: layered build order and early pilot (R24) — if core+domain slips badly, the pilot surfaces it before research commands are built.

## Key Decisions
- **CLI mandatory, no markdown fallback** — contracts without an enforcer are just longer prose; rejected fallback mode because dual paths re-create the drift the harness exists to kill. Cost accepted: existing users break until they install; mitigated by install script + version-gate error messages.
- **SQLite + committed changesets (upstream model)** over files-as-database — rejected stateless-CLI-over-markdown because the user chose upstream fidelity and query power; two-sources-of-truth objection resolved by changesets being the committed truth and the DB a rebuildable materialization.
- **New Go port (`zharness`)** over reusing the Rust harness-cli binary — rejected reuse to own the surface (workflow additions: resume/check record/validate), keep the repo single-language for contributors, and control releases.
- **Mirror upstream 4-layer packages** over idiomatic flat `internal/{store,changeset,…}` — chosen for line-by-line port traceability; `interface` renamed `interfaces` (Go keyword).
- **cobra + modernc.org/sqlite** over urfave/cli and mattn/go-sqlite3 — command tree depth needs cobra; pure-Go driver keeps cross-compilation trivial.
- **Per-change ULID changeset files** over session/daily append files — zero merge conflicts between branches; replay order from filename.
- **ULID everywhere** over UUIDv4/sequential — sortable, offline-generated, doubles as filename.
- **SKILL.md rewritten CLI-first with inline commands** over per-skill `references/cli-adapter.md` — user chose explicitness: the workflow skills now revolve around interacting with the harness CLI; flag-change churn across 8 files accepted.
- **`workflow-state.yml` retired outright** over keeping it as a rendered read-only view — one canonical read path (`query`/`resume`); glanceability lost on GitHub accepted.
- **Contracts consolidated in `cli/docs/`** over splitting with `docs/contracts/` — contracts version with the code that enforces them.
- **Install via GitHub Releases + `gh release download` script** over adding go-install/Homebrew — private repo makes gh auth the only smooth path; extra channels deferred.

## Open Questions
- Exact story↔phase field mapping (harness `story` entity ↔ to-plan phase slugs) — owner: gap-matrix phase; blocks SCHEMA.md finalization, not planning.
- Trace quality tier definitions to port for `score-trace` — owner: CLI research phase; upstream tiers are the default unless the pilot shows misfit.

## Deferred Ideas
- Homebrew tap and `go install` distribution channels
- Skill adoption of `propose` and `score-context` (ship reserved in CLI)
- Dedicated maintenance skill (harness health/audit UX, like /health)
- Windows binaries
- SQLite read-index optimizations / complex cross-story graph queries
- Rendered read-only state snapshot for GitHub glanceability (only if retiring yml proves painful)

## Ambiguity Report
- Goal clarity: locked — runtime evolution with preserved UX, mandatory CLI.
- Scope clarity: locked — in/out lists explicit; skills outside workflow untouched.
- Constraints clarity: locked — stack, private-repo distribution, append-only changesets, solo-maintainer build order.
- Acceptance clarity: locked — every criterion checkable by command, fixture, or documented walkthrough.
