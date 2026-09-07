# Repository Workflow

`AGENTS.md` is the entrypoint. This file defines the shared workflow boundary; stage-specific procedure lives under `docs/playbooks/`.

## Authority

Classify the request before mutation. Read-only requests inspect only what the answer needs and do not mutate harness state. Change requests remain limited to the active stage and the user-approved scope. Discovery does not grant authority to fix adjacent findings.

Repository docs, code, tests, and observable runtime behavior define current truth. A legacy per-machine index, if one still exists locally, is only a recovery cache; it does not define product policy.

## Context

The lifecycle needs no binary: route through the table below and read only the named playbook. The `zharness` binary (install / update / uninstall) scaffolds and updates these managed docs; it plays no part in running a stage.

| Stage | Playbook |
|---|---|
| brainstorm | `docs/playbooks/brainstorm.md` |
| to-plan | `docs/playbooks/to-plan.md` |
| work | `docs/playbooks/work.md` |
| check | `docs/playbooks/check.md` |
| handoff | `docs/playbooks/handoff.md` |
| watzup | `docs/playbooks/watzup.md` |

`git` and `interview` keep their skill-local procedure and are never harness-gated.

## Execution boundary

Reduced mode mutates nothing durable. Durable stages append to the active plan's markdown sections exactly as each playbook directs; nothing else writes them. Every proof claim must name actual command output or observable evidence. If repository tooling and a playbook disagree, trust the repository and report the docs mismatch.

escalate_when: ask the owner and stop — locked schema or requirements would change; the same verification command failed twice; a product rule conflicts.
