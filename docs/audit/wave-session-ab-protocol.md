# Wave-boundary vs same-session — A/B protocol

**Status:** measurement only. Does **not** change `docs/playbooks/work.md` step 11 (all waves + in-session gate remain the default). Lock a STOP-between-waves policy only after this protocol produces numbers.

**Why.** Consumer measurement (llmhub, 2026-09-02) showed one parent session dominating cost while cache-read volume was high. A new session pays cold-entry cost (`docs/audit/sdlc-token-cache-audit.md` F1 / consumer-adoption cold entry). Those two costs have not been compared on the same phase.

## Variants

| ID | Flow | Default? |
|---|---|---|
| **A** | Current: every wave of the phase, then `check.md` gate, **same session**. Do not dispatch `/check` (opus pin). | yes |
| **B** | After each wave: flush `## Progress` + `## Current State`, **end session**. Next session: `work full {phase}` for the next incomplete wave. Gate still in-session on the **last** wave only. | no |

Do not run `/handoff` between waves. Do not disable prompt cache. Do not use an OMP-only adapter as the measurement harness — Claude Code, Codex CLI, Pi/OMP, or any host that `codeburn` can read.

## Paired runs (required)

Do **not** replay “the same phase on separate days” on a live consumer tree. That cannot be the same work.

For each matched pair:

1. Create **two disposable git worktrees or clones** from the **same baseline commit**.
2. Copy in an **identical** multi-wave plan (same phase, waves, tasks, proof commands).
3. Pin **identical** host settings: model, provider, thinking/effort, and tool/permission profile. Record the pin in the pair log. Any host (Claude Code, Codex CLI, Pi/OMP, …) is fine; both arms of a pair must use the same host.
4. Randomize which clone is A vs B (fair coin). Do not always run A first.
5. Run A and B to completion. **Both quality gates must pass** with equivalent outcomes (same checks green, same files in the phase surfaces, no `BLOCKED_*` on one arm only). A cheaper incomplete B does not count.
6. Collect from `codeburn report --project <clone> --from <start> --to <end>`: total cost (include cache-write and cold-entry), cache-read tokens, cache-hit %, calls, session count, longest-parent cost, elapsed wall time. Exploration mix is diagnostic only.

Run **≥3** matched pairs. Compare **medians** across pairs, not a single `B < A`.

## Decision rule

Lock variant B into embedded `docs/playbooks/work.md` (edit `cli/docs/embedded/playbooks/work.md` first) only if **median total cost of B < median total cost of A**, both arms gated equivalent, n≥3 pairs. If A wins, they tie, or any pair fails the outcome check, keep step 11. Do not publish a savings percentage from one run.

## Out of scope

Skill-body diets, watzup rewrites, installer-managed `record-check.sh` on consumers, runtime-specific session killers as core zharness behavior.
