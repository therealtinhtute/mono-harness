## Harness

Work runs from this repository: committed markdown is the record, git hooks and repo scripts are the guards. Start at `docs/WORKFLOW.md`; route by work shape — read-only and bounded edits stay reduced and mutate nothing durable; durable planning, execution, checks, and handoffs follow the stage playbooks under `docs/playbooks/` and keep `docs/plans/active/*.md` append-only and true.

If the `zharness` binary exists, `zharness preflight <stage> [--mode <mode>] --json` returns readiness, the playbook path, and drift recovery. A missing binary is not an error: read `docs/playbooks/{stage}.md` directly and continue.

Repository docs, code, tests, and observable behavior are authoritative; any database present is only a lifecycle ledger and recovery index, never a second control plane — no task database, no parallel control-plane state.
