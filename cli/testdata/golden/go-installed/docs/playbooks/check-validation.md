# Playbook: check (Validation evidence)

Loaded by `docs/playbooks/check.md` step 8 in durable `gate`/`full` mode, including `work`'s phase gate. `bounded` never loads it.

## Validation Entry Format

```markdown
- <UTC> — phase `<phase-slug>` — verdict: APPROVED — mode: gate
  - `<command>` — <result in at most 3 lines>
  - scope: on target | drift | incomplete — <one line>
  - requirements: R1 met (<proof>) | R2 not met (<gap>) | R3 partial (→ <phase>)
  - rollback_point: <commit the phase can be reverted to>
  - requests: none | <numbered findings with path:line>
  - not_independently_verified: <aspect> (same-session only)
  - judge: independent | same-session
  - judge_model: <model identifier>
```

Every Validation entry must include timestamp, stable phase slug, exact command/result and concise output, verdict, requirement coverage, rollback point, judge declaration, reviewing model identifier, and proof gaps. Two positions are load-bearing, so the commit-time guard can find them:

- The first line carries the anchored token `verdict: <VERDICT>`. A verdict on a sub-bullet, or without the `verdict:` anchor, is skipped in silence: no rejection, no re-execution.
- Each proof command is a nested sub-bullet, indented at least two spaces, whose text begins with the bare command in single backticks: no label before the opening backtick, no double backticks, no backtick inside the command (it truncates there and re-executes a broken fragment; use a character class or a shell-free equivalent). Trailing annotation after the closing backtick is safe; keep it to at most 3 lines of output.

Validation is append-only from an entry's first commit; never replace committed failed evidence or verdicts. The one exception is closing compaction in `handoff.md`, which keeps the last entry per phase byte-identical. An uncommitted entry may still be reshaped into the guard-visible form (the guard hashes only committed entries).

**Proof re-execution contract** — each proof re-runs from the repository root under `sh -c` with no prior `cd`, activated virtualenv, or session function/alias; wrap bash-only syntax such as process substitution in `bash -c '…'`. Exported environment variables are inherited, so a proof must depend on neither their presence nor their absence. Bounded to 300 s IF `timeout` or `gtimeout` is on PATH, else unbounded. Carry any working directory or interpreter inline (`cd cli && go test ./...`). Never test for a binary's presence or an OS version: re-execution happens on a bare checkout.
