# Context: cli-stale-drift

Phase: cli-stale-drift
Status: ready
Spec Link: ../../SPEC.md
Roadmap Link: ../../ROADMAP.md
Blast Radius: medium
Expected Proof: unit, integration (scratch-dir lifecycle e2e)

## Goal
`resume` detects and reports `stale_docs` drift (written docs behind the CLI's embedded docs version) with the exact recovery `zharness init --refresh-docs`; the recovery clears the drift; the whole init→check lifecycle is integration-tested on a scratch dir.

## Scope Boundary
### Allowed Surfaces
- `cli/internal/**` (resume/drift logic, application layer)
- `cli/docs/STATE.md` (new drift row in the stale-pointer table)
- Go test files; scratch-dir integration harness

### Forbidden Surfaces
- `skills/workflow/**`
- init scaffolding semantics (fixed in cli-embed-scaffold; bug fixes only)
- Release pipeline

## Spec Hooks
- R3 (complete): stale_docs drift + named recovery + clearing semantics
- Acceptance: "upgrade CLI → resume reports stale_docs naming init --refresh-docs; recovery clears it, rewrites only docs"

## Locked Decisions
- Drift fires iff `meta.docs_version` exists AND differs from the CLI's embedded docs version AND neither is `dev` — dev builds never fire staleness (dogfooding must not drown in drift)
- Missing `meta.docs_version` (project scaffolded pre-embed or imported legacy) does NOT fire `stale_docs`; `resume` may note docs as `unversioned` informationally — blocking old projects is not acceptable
- The recovery string is defined once in Go (single constant) and STATE.md quotes it — the #24 lesson: doc quotes code, not a parallel copy
- `stale_docs` is additive to the drift array; readiness becomes `drifted` per the existing rule (non-empty drift array), no new readiness states

## Assumptions
- `resume --json` consumers (watzup skill, future playbook) render drift entries generically — a new type needs no consumer change beyond the recovery-table row

## Canonical Refs
- `cli/docs/STATE.md` (stale-pointer rules table — add the row here)
- `cli/internal/interfaces/resume.go`
- GitHub issue #24 (recovery-string drift precedent)

## Rejected Options
- Hash-based staleness (compare doc content hashes, not versions) — more precise but noisy: any local doc edit would fire drift; version stamp matches the actual failure mode (CLI upgraded, docs stale)
- Auto-refresh on resume — resume is read-only by contract (watzup writes nothing); mutating there breaks the writer/reader ownership table

## Deferred Ideas
- Per-doc staleness granularity; docs integrity hash as an audit (not resume) finding

## Escalate If
- The existing drift plumbing cannot carry a non-pointer drift type without restructuring → to-plan phase
- watzup's recap contract needs a new vocabulary row it cannot derive from `recovery` verbatim → user clarification (skill contract change belongs to thin-triggers)
