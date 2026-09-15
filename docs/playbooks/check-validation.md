# Playbook: check (Validation evidence)

Loaded by `docs/playbooks/check.md` step 8 in durable `gate`/`full` mode, including `work`'s in-session gate. Bounded/simple and `review` never load it.

## Validation Entry Format

Every Validation entry must include timestamp, stable phase slug, exact command/result and concise output, verdict, judge declaration, reviewing model identifier, proof gaps, and a grep-able `receipt:` block (`context_sources`, `policy`, `judge`, `judge_model`, `retries`, `rollback_point`, `failure_ledger: absent|{path}`, `enforcement: hook | ci | local-only`, `not_independently_verified`). Two positions are load-bearing, so the commit-time guard can find them:

- The verdict token sits on the entry's own first line. A verdict on a sub-bullet is skipped in silence: no rejection, no re-execution.
- Each proof command is a nested sub-bullet, indented at least two spaces, whose text begins with the bare command in single backticks: no label before the opening backtick, no double backticks, no backtick inside the command (it truncates there and re-executes a broken fragment; use a character class or a shell-free equivalent). Trailing annotation after the closing backtick is safe.

Validation is append-only from an entry's first commit; never replace committed failed evidence or verdicts. An uncommitted entry may still be reshaped into the guard-visible form (the guard hashes only committed entries).

**Proof re-execution contract** — each proof re-runs from the repository root under `sh -c` with no prior `cd`, activated virtualenv, or session function/alias. Exported environment variables are inherited, so a proof must depend on neither their presence nor their absence. Bounded to 300 s IF `timeout` or `gtimeout` is on PATH, else unbounded. Carry any working directory or interpreter inline (`cd cli && go test ./...`). Never test for a binary's presence or an OS version: re-execution happens on a bare checkout.
