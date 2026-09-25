# Retro — operating logic

Output: at most 5 findings, severity-ranked, each in the schema below, exactly
one marked `Next`. No preamble. No file edits — a retro diagnoses, the hand-off
skill changes things and proves the change.

A retro improves the agent's **environment** (navigation, guards, steering,
tools, information), never the code the session produced. Every finding must
cite evidence a reader can open: a digest line, a turn number, or `file:line`.
Drop any candidate you cannot cite.

## 1. Load the session

- Default: `python3 scripts/session-digest.py` — newest Claude Code session for
  the current directory (`~/.claude/projects/<cwd with non-alphanumerics as ->/*.jsonl`).
- A named session: pass its id or `.jsonl` path.
- Another agent (Codex: `~/.codex/sessions/**/*.jsonl`) or a pasted transcript:
  read it directly and collect the same five signals the digest prints —
  tool-call counts, errored results, repeated identical calls, largest results,
  user prompts.
- Open the raw log only to confirm a specific turn the digest points at.

## 2. Read the repository's guardrails before judging

- Steering: `AGENTS.md` / `CLAUDE.md` in the repo, plus global ones if readable. Note line counts.
- Gates: the repo's check command (its AGENTS.md gate list, `package.json` /
  `Cargo.toml` / `Makefile` scripts), git hooks (`.git/hooks/`, a hook
  installer), and CI workflows.
- Review surface: whatever reviews a diff here — a review playbook, a judge,
  a `CODING_STANDARDS.md`, a review bot. Record which exist; assume none.

A check that exists but is unwired, skipped, or silently broken is the finding —
not a new check. A repo with no hook and no CI job running its lint/test is
itself a finding.

## 3. Scan these categories

- **Navigation** — the agent searched long or read wrong files first. Candidate fix: a pointer in the file the agent already reads.
- **Automated checks** — the agent made an error a linter, type check, test, or hook would catch.
- **Review standards** — the review surface missed a defect. Classify first:
  a **mechanical** violation (fixed pattern, banned API, import shape, file
  location) gets a deterministic check; only a **judgement** call (cross-file
  consistency, surrounding style) gets a written review rule.
- **Steering weight** — AGENTS.md/CLAUDE.md carries rules that belong in a check or review rule; it should hold navigation pointers.
- **Tool economy** — large results, repeated identical calls, or a token-heavy custom tool in the digest.
- **No-ops** — a steering line the session shows had no effect on behaviour; removing it would change nothing.
- **Information access** — the agent lacked a log, service, or doc it needed and guessed or asked.

## 4. Rank by severity

| Severity | Meaning |
|---|---|
| `S1` | Wrong result shipped or claimed: a false completion claim, an unproven "done", a defect that passed every gate. |
| `S2` | Cost human steering or rework: a user correction, a failed commit/CI, a reverted change. |
| `S3` | Wasted budget only: tokens, time, repeated calls, no wrong outcome. |

Break ties by recurrence: a friction that will hit every future session outranks a one-off.

## 5. Report

```
### {n}. [{S1|S2|S3}] {category} — {one-line friction}
- evidence: {digest line / turn N / file:line}
- mechanical | judgement: {which, and why}
- smallest fix: {change} at {file or tool}
- hand-off: improve-harness | encode-invariant | human
```

Mark one finding `Next` — highest severity, then highest recurrence. End with
one line naming the hand-off skill and the finding it would take. Stop there.
Success of any fix is proven only by that skill's fresh rerun, never by the retro.

<example id="1">
### 1. [S1] Automated checks — work claimed `cargo test` green; it was never run (Next)
- evidence: digest shows 0 `cargo test` calls; turn 14 says "all tests pass"
- mechanical: a pre-commit guard can require a fresh proof line
- smallest fix: extend the existing proof guard in `scripts/install-git-hooks.sh`
- hand-off: encode-invariant
</example>

<example id="2">
### 3. [S3] Tool economy — full `docs/ARCHITECTURE.md` read four times
- evidence: digest "repeated identical calls: Read x4"
- judgement: the agent needed one section, not a check
- smallest fix: section pointer in `AGENTS.md` Project Structure
- hand-off: improve-harness
</example>
