# Repository Workflow

`AGENTS.md` is the entrypoint. This file defines the shared workflow boundary; stage procedure lives under `docs/playbooks/`.

## Authority

Classify the request before mutation. Read-only requests inspect only what the answer needs and mutate no harness state. Change requests stay within the active stage and user-approved scope; discovery grants no authority to fix adjacent findings.

Repository docs, code, tests, and observable runtime behavior define current truth. A legacy per-machine index, if one exists locally, is only a recovery cache, never product policy.

## Context

The lifecycle needs no binary: route through the table and read only the named playbook and any companion it routes you to. `zharness` (install / update / uninstall) scaffolds and updates these managed docs; it plays no part in running a stage.

| Stage | Playbook |
|---|---|
| brainstorm | `docs/playbooks/brainstorm.md` |
| to-plan | `docs/playbooks/to-plan.md` |
| work | `docs/playbooks/work.md` |
| check | `docs/playbooks/check.md` |
| handoff | `docs/playbooks/handoff.md` |
| watzup | `docs/playbooks/watzup.md` |

`git` and `interview` keep skill-local procedure and are never harness-gated.

## Execution boundary

Reduced mode mutates nothing durable. Durable stages write the active plan's sections exactly as each playbook directs; nothing else writes them. Every proof claim names actual command output or observable evidence. IF repository tooling and a playbook disagree → trust the repository and report the docs mismatch.

escalate_when: ask the owner and stop — locked schema or requirements would change; the same verification command failed twice; a product rule conflicts.
