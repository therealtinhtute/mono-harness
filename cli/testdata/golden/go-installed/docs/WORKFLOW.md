# Repository Workflow

`AGENTS.md` is the entrypoint. This file defines the shared workflow boundary; stage procedure lives under `docs/playbooks/`.

## Authority

Classify the request before mutation. Read-only requests inspect only what the answer needs and mutate no harness state. Change requests stay within the active stage and user-approved scope; discovery grants no authority to fix adjacent findings.

Repository docs, code, tests, and observable runtime behavior define current truth. A legacy per-machine index, if one exists locally, is only a recovery cache, never product policy.

## Context

The lifecycle needs no binary: route through the table and read only the named playbook and any companion it routes you to. `zharness` (install / update / uninstall) scaffolds and updates these managed docs; it plays no part in running a stage.

| Stage | Playbook | Model tier |
|---|---|---|
| brainstorm | `docs/playbooks/brainstorm.md` | deep |
| to-plan | `docs/playbooks/to-plan.md` | deep |
| work | `docs/playbooks/work.md` | standard |
| check | `docs/playbooks/check.md` | deep |
| handoff | `docs/playbooks/handoff.md` | standard |
| watzup | `docs/playbooks/watzup.md` | fast |

`git` keeps skill-local procedure and is never harness-gated.

A tier is a capability class, not a model name. Each host maps tiers once:

| Host | deep | standard | fast |
|---|---|---|---|
| Claude Code | opus | sonnet | haiku |

Do not switch model or agent mid-phase: prompt cache is per model, so a switch re-reads the phase cold.

## Execution boundary

Reduced mode mutates nothing durable. Durable stages write the active plan's sections exactly as each playbook directs; nothing else writes them. Every proof claim names actual command output or observable evidence. Cite `docs/WORKFLOW.md` and `docs/playbooks/*.md` by section and step number, never by line number: `zharness update` overwrites them, so line numbers drift. IF repository tooling and a playbook disagree → trust the repository and report the docs mismatch.

escalate_when: ask the owner and stop — locked schema or requirements would change; the same verification command failed twice; a product rule conflicts.
