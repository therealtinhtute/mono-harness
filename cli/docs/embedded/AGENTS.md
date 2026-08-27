## Harness

Work runs from this repository: committed markdown is the record, git hooks and repo scripts are the guards. Start at `docs/WORKFLOW.md`; route by work shape — read-only and bounded edits stay reduced and mutate nothing durable; durable planning, execution, checks, and handoffs follow the stage playbooks under `docs/playbooks/` and keep `docs/plans/active/*.md` append-only and true.

The `zharness` binary manages the doc set (install / update / uninstall) once those verbs ship; it plays no part in running the lifecycle.

Repository docs, code, tests, and observable behavior are authoritative; any legacy per-machine index is only a recovery cache, never a second control plane — no task database, no parallel control-plane state.
