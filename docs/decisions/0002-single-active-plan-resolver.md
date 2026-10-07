# 0002 — One resolver owns the active-plan invariant, and it returns a Stop

## Status

Superseded by [0006](0006-v015-authority.md) — historical record, not a current runbook. The `ResolveActivePlan` resolver it ratifies was deleted in v0.15; the at-most-one-active-plan invariant is now dual-encoded in playbook guidance and `zharness_guard_at_most_one_active_plan` in ZGUARD-CORE. Accepted and implemented by commit `7a4195f`, phase `p0-single-active-plan` (`git show 3bab3c0:docs/plans/completed/harness-markdown-truth.md`).

## Context

`docs/audit/consumer-adoption-audit.md` records D1: the harness assumed exactly one plan under `docs/plans/active/` but never enforced it, and every caller that needed the active plan found it independently. Each one failed differently — one raised a validation error, one returned `ok=false` with no explanation, one guessed. An agent hitting the ambiguous case got a different story depending on which command it happened to run, and the usual recovery was to read plan bodies until one looked current, which is exactly the unbounded read the harness exists to avoid.

## Decision

A single function, `ResolveActivePlan`, is the only way to obtain the active plan, and it never returns a bare error for the ambiguous or absent case.

- `cli/internal/application/plan_resolve.go:73` is the sole entry point. Every caller goes through it: `plan_query.go:68`, `plan_write.go:22`, `resume.go:156`, `plan_lifecycle.go:53`, `plan_lifecycle.go:94`, `validate.go:446`.
- Zero plans returns `Stop{Code: "none"}` carrying the recovery `brainstorm lock`. More than one returns `Stop{Code: "ambiguous"}` carrying a bounded candidate list.
- Disambiguation is a three-tier ladder and never reads a plan body. Tier 0 is the index and traces already in the agent's context. Tier 1 is the first ten frontmatter lines per candidate. Tier 2 is a bounded packet that declares what it omitted. When frontmatter is missing or unparseable, the ordering signal is `git log -1 --format=%cI -- <path>` — still never the body.
- Exactly two commands move a plan out of `docs/plans/active/`: `zharness plan complete` and `zharness plan abandon`. There is no "select the active plan" command, because selection would make ambiguity survivable instead of fixable.
- Plan creation fails when a non-empty active plan already exists, and the failure names the existing path.

## Consequences

- The invariant enforced is **at most one** active plan, not exactly one. Zero is a valid idle state — it is the designed result of `plan complete` before the next plan is locked — and `zharness validate` produces no finding for it. Only `Stop{Code: "ambiguous"}` is a finding.
- Playbooks branch on the Stop `code` rather than guessing. `docs/playbooks/work.md` step 1 does this explicitly and instructs the agent to stop rather than pick.
- Ambiguity now has one message, one recovery, and one place to change it.
- The dead `intervention` surface and the `next` command were removed rather than deprecated, since the resolver made them redundant.

## Consequences that are costs

Two active plans block work entirely until the owner runs `plan complete` or `plan abandon`. That is deliberate — a blocked agent is cheaper than an agent silently advancing the wrong initiative — but it does mean the failure surfaces at the least convenient moment, mid-session.

## Authority

- `cli/internal/application/plan_resolve.go:73` — `ResolveActivePlan` and its doc comment naming D1.
- `git show 3bab3c0:docs/audit/consumer-adoption-audit.md` — D1, the original finding.
- `git show 3bab3c0:docs/plans/completed/harness-markdown-truth.md` — R1 through R7 and R13.
