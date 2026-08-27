# Repository Workflow

`AGENTS.md` is the entrypoint. This file defines the shared workflow boundary; stage-specific procedure lives under `docs/playbooks/`.

## Authority

Classify the request before mutation. Read-only requests inspect only what the answer needs and do not mutate harness state. Change requests remain limited to the active stage and the user-approved scope. Discovery does not grant authority to fix adjacent findings.

Repository docs, code, tests, and observable runtime behavior define current truth. Any present `harness.db` is a derived lifecycle ledger and recovery index; it does not define product policy.

## Context

If the `zharness` binary exists, `zharness preflight <stage> [--mode <mode>] --json` returns readiness, the playbook path, and drift recovery; if it is absent, proceed straight to the table below. When a returned `playbook` is present, read that file and no other stage playbook.

| Stage | Playbook |
|---|---|
| brainstorm | `docs/playbooks/brainstorm.md` |
| to-plan | `docs/playbooks/to-plan.md` |
| work | `docs/playbooks/work.md` |
| check | `docs/playbooks/check.md` |
| handoff | `docs/playbooks/handoff.md` |
| watzup | `docs/playbooks/watzup.md` |

`git` and `interview` may consult preflight but keep their skill-local procedure.

## Execution boundary

Reduced mode mutates nothing durable. Durable stages append to the active plan's markdown sections exactly as each playbook directs; while the binary exists its commands reconcile the ledger after that markdown write. Every proof claim must name actual command output or observable evidence. If repository tooling and a playbook disagree, trust the repository and report the docs mismatch.
