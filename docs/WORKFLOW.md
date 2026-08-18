# Repository Workflow

`AGENTS.md` is the entrypoint. This file defines the shared workflow boundary; stage-specific procedure lives under `docs/playbooks/`.

## Authority

Classify the request before mutation. Read-only requests inspect only what the answer needs and do not mutate harness state. Change requests remain limited to the active stage and the user-approved scope. Discovery does not grant authority to fix adjacent findings.

Repository docs, code, tests, and observable runtime behavior define current truth. `harness.db` records lifecycle position, evidence links, and recovery state; it does not define product policy.

## Context

Run `zharness preflight <stage> [--mode <mode>] --json`. Follow a returned stop and recovery. When `playbook` is present, read that file and no other stage playbook.

| Stage | Playbook |
|---|---|
| brainstorm | `docs/playbooks/brainstorm.md` |
| to-plan | `docs/playbooks/to-plan.md` |
| work | `docs/playbooks/work.md` |
| check | `docs/playbooks/check.md` |
| handoff | `docs/playbooks/handoff.md` |
| watzup | `docs/playbooks/watzup.md` |

`git` and `interview` use preflight but keep their skill-local procedure.

## Execution boundary

Reduced mode is read-only with respect to harness state. Durable stages require an initialized database and use changeset-first CLI commands. Every proof claim must name actual command output or observable evidence. If the live CLI and a playbook disagree, trust `--help` and report the docs mismatch.
