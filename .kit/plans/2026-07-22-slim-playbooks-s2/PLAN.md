# PLAN: Slim Playbooks S2 — `zharness next` + work.md routing slim

**Date:** 2026-07-22
**Origin:** `.kit/plans/2026-07-21-slim-playbooks/PLAN.md` Phase S2 (locked). Continues S1 (scaffold command, done + gated 2026-07-22, see `.kit/reports/check/20260722-slim-playbooks-s1-w3-gate.md`).
**Lane:** high-risk (same as S1 — public CLI contract change, embedded-playbook rewrite).

## Scope decision (stated up front, not silently assumed)

`next` folds every row of work.md's Mode-Resolution + Full-Mode-Detection tables **except `contract-drift` and `stale-plan`**. `contract-drift` requires diffing the working tree against a phase's Allowed/Forbidden Surfaces + task `touches`/`avoid` — a git operation the CLI has no access to, identical to the constraint `resume` already carries (its git/WIP block stays agent-gathered, per S1's `PLAN.md` item 2). `stale-plan` (a discovery made mid-implementation, not in the original scoping) requires checking whether a phase plan references files/symbols that no longer exist — but a plan legitimately references files it hasn't created yet, so a naive existence check false-positives on almost every real plan (including this very plan, which references `cli/internal/application/next.go` before it existed). `next`'s Go implementation will not attempt either; work.md keeps a one-line agent-side note for each instead of dropping the checks.

## Steps

1. **Domain**: add `NextView` / `StopInfo` types + pure resolution function(s) in `internal/application/next.go` — no new domain package needed, mirrors `resume.go`'s shape (`{mode, active_phase, stop}` per PLAN.md's exact contract).
   - Inputs: parsed argument (`""`/`auto`, `simple`, `simple @file`, `full`, `full phase <slug>`, `phase <slug>`), DB state (`query state`, `query phases`/stories), filesystem reads (`.kit/planning/SPEC.md` exists+locked, `ROADMAP.md` exists+parseable ordered phase list, per-phase `-PLAN.md`/`-CONTEXT.md` exist, PLAN.md scanned for `TBD`/`TODO`/"similar to"/"implement later" and wave-completion markers, `.kit/reports/brainstorm/*.md` presence).
   - Output states: `no-spec`, `no-plan`, `no-phase`/`no-context`, `placeholder-plan`, `multiple-incomplete`, `all-phases-done`, `ready`, plus simple-mode `ready` and auto-mode's `ambiguous` stop. (`stale-plan` excluded — see Scope decision above.)
   - Verify: unit — one test per table row (10ish cases), ported directly from the Mode-Resolution + Full-Mode-Detection tables so removing the prose from work.md doesn't silently drop a case (table→test parity, the plan's stated invariant).

2. **Interface**: `internal/interfaces/next.go` — new `zharness next [argument] --json` cobra command, wired into `root.go`, following the `scaffold`/`resume` command shape (JSON via `--json`, human-readable one-liner otherwise).
   - Verify: command-output — `go build && zharness next` smoke against a fresh-build binary, a few real states in this repo (e.g. current write-boundary phase) and a scratch fixture tree for the stop cases.

3. **Integration test**: exercise `next` against a real temp `.kit/` fixture (tmp dir, real files, real sqlite db via existing test helpers) for at least: no-spec, no-plan, no-phase, ready, simple-mode ready. Crosses a real filesystem + DB boundary — satisfies the `integration` proof-class cell.

4. **Playbook slim**: strip work.md's `Mode Resolution` (L23-29), `Full Mode Detection Table` (L41-58), and `Stop message shapes` (L60-66) sections (~54 lines) → replace with a `zharness next` call + a short "follow its `stop`/`mode` output; recovery commands come from `next`'s `stop.recovery` field" note. Keep one-line agent-side notes for `contract-drift` and `stale-plan` (git-blind / plan-lifecycle-blind, CLI can't check either) so those rows aren't silently dropped. Update Command Reference section.

5. **Gate**: run `check full` on the S2 diff (same pattern as S1-W3), persist report, skip `check record` again (no RUN artifact for this informal track — same as S1).

## Non-scope (unchanged from master plan)

- `contract-drift` and `stale-plan` detection themselves — stay agent-side.
- Simple-mode carve-out prose compression is bundled into this same diff (master plan lists it under S2), but is a small text edit, not new logic — `next` returning `mode: simple` already covers the FK carve-out; the prose collapses to one line + `CONTRACT.md` ref.
- S3 (`resume` text rendering) is separate, not touched here.

## Rollback

Additive command + playbook-text-only change, same as S1 — revertable independently via git, no schema change.
