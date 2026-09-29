---
id: 01M1PROMPTDIET7Q2WZ4
type: plan
intake_id: 01M1PROMPTDIETINTK7Q2WZ4
lane: high-risk
status: active
created: 2026-09-29
updated: 2026-09-29
---

# Plan: prompt-token-diet — fewer instruction tokens, two missing sensors

## Outcome
- result: always-loaded instructions (AGENTS.md, rules/) and the six heaviest skill bodies carry each instruction once, with detail in `references/`; dangling skill references and secret leaks are caught by scripts, not prose.
- success_signals:
  - S1: `bash scripts/validate-skill.sh` fails on a SKILL.md that names a skill absent from `skills/*/*/`, and passes on all current skills.
  - S2: `bash scripts/test-guards.sh` shows the pre-commit hook rejecting a staged `AKIA…` key, with output naming `file:line` but not the key.
  - S3: `wc -w` drops ≥20% on each trimmed file; `bash scripts/verify-doc-links.sh` 0 findings.
  - S4: a Codex diff-review of each trimmed skill lists no dropped behavior the owner rejects.

## Authority and Requirements
- authority:
  - Owner, 2026-09-29: audit (Claude + Codex second review) findings #1–#6; answers — close the old plan, keep 🥷 lines, enforce secrets in pre-commit, verify trims with Codex diff-review.
- requirements:
  - R1 [accepted]: `create-skill`, `create-cli`, `turbo-mono-platform` defer to `check`, not the nonexistent `review`; `validate-skill.sh` resolves every backticked name under "Defer To" against the skill inventory. | source: audit #1
  - R2 [accepted]: ZGUARD-CORE pre-commit blocks staged secret patterns and prints `file:line pattern` only; `git/SKILL.md` Step 2 shrinks to "the hook blocks; on block, unstage or move to env". | source: audit #2
  - R3 [accepted]: AGENTS.md keeps layout, gates, harness; install commands, orkit history, SSH note move to README. | source: audit #3
  - R4 [accepted]: `rules/karpathy-guidelines.md` states each instruction once; installed copy matches repo. | source: audit #4
  - R5 [accepted]: `create-cli`, `prompt-leverage`, `librarian`, `create-skill` drop body text duplicated in their own `references/`. | source: audit #5
  - R6 [accepted]: `think`, `hunt` state each rule once and drop explanatory "why" sentences. | source: audit #6

## Non-goals
- NG1: Keep the 🥷 / verdict-first line in every skill (other agents lack CLAUDE.md).
- NG2: No new eval suites; no CI secret-scan job.
- NG3: No behavior change to any skill beyond removing duplication.

## Phases and Verification
- approach: sensors first (real bugs, deterministic checks), then always-loaded text (paid every session), then skill bodies (paid on invocation, highest regression risk). Rejected: one big trim pass — no way to attribute a regression.
- constraints:
  - Hook edits stay inside ZGUARD-CORE; bash 3.2 safe.
  - One skill per commit in p3 so each trim reverts alone.
- risks:
  - Secret regex false positives on docs mentioning "token"/"password" → match value-shaped patterns (AKIA[0-9A-Z]{16}, `-----BEGIN .*PRIVATE KEY`, `://user:pass@`) not bare words; fixture proves a doc line with "token" passes.
  - Trim drops a load-bearing constraint → Codex diff-review per skill; owner approves before commit.
  - Rule edits change behavior in every session → reinstall only after owner reads the diff.
- recovery: `git revert` the single commit; hook block reverts independently of skill edits.
- seams: `scripts/validate-skill.sh`, `scripts/test-guards.sh`, `scripts/verify-doc-links.sh`, `codex exec -s read-only` diff-review (owner-approved 2026-09-29).
- phase_slug: `p1-sensors`
  status: checked
  - goal: R1, R2 | depends_on: none
  - surfaces: skills/craft/create-skill, skills/shipping/{create-cli,turbo-mono-platform}, skills/workflow/git, scripts/validate-skill.sh, scripts/install-git-hooks.sh (ZGUARD-CORE), scripts/test-guards.sh | avoided: cli/, rules/
  - escalate_when: the secret regex needs an allowlist mechanism, or the hook fixture cannot run without network
  - wave 1:
    - T1 replace `review` with `check` in 3 skills; validator parses "Defer To" backticked names against `ls skills/*/*/` — output: validator rejects a temp SKILL.md deferring to `nosuch` — check: `for f in skills/*/*/SKILL.md; do bash scripts/validate-skill.sh "$f" || echo FAIL "$f"; done` — stop_if: an existing skill fails for a legitimate external name
    - T2 add secret scan to ZGUARD-CORE + fixtures (AKIA key blocked, value absent from output; doc line with "token" passes); shrink git Step 2 — output: guard blocks and redacts — check: `bash scripts/test-guards.sh` — stop_if: any existing fixture regresses
  - phase check: `bash scripts/test-guards.sh && bash scripts/verify-doc-links.sh`
- phase_slug: `p2-always-loaded`
  status: planned
  - goal: R3, R4 | depends_on: p1-sensors
  - surfaces: AGENTS.md, README.md, rules/karpathy-guidelines.md, ~/.claude/rules/ | avoided: skills/
  - escalate_when: a removed AGENTS.md line is referenced by a playbook or script
  - wave 1:
    - T3 move install/orkit/SSH text from AGENTS.md to README — output: AGENTS.md ≤70 lines — check: `wc -l AGENTS.md && bash scripts/verify-doc-links.sh` — stop_if: link check fails
    - T4 dedupe karpathy rule; owner reads diff; reinstall all rules — output: ≤360 words, installed = repo — check: `wc -w rules/karpathy-guidelines.md && for f in rules/*.md; do diff -q "$f" ~/.claude/rules/$(basename "$f"); done` — stop_if: owner rejects diff
  - phase check: `bash scripts/verify-doc-links.sh`
- phase_slug: `p3-skill-trims`
  status: planned
  - goal: R5, R6 | depends_on: p2-always-loaded
  - surfaces: skills/craft/{create-skill,librarian,prompt-leverage}, skills/shipping/create-cli, skills/workflow/{think,hunt} (SKILL.md and references/) | avoided: rules/, scripts/, cli/
  - escalate_when: Codex review flags a dropped constraint the owner wants kept but no reference fits it
  - wave 1 (independent, one commit each):
    - T5 create-cli · T6 prompt-leverage · T7 librarian · T8 create-skill — output: body minus text already in its references/ — check: `wc -w <file>` ≥20% drop, `bash scripts/validate-skill.sh <file>`, `git diff <file> | codex exec -s read-only "list behaviors removed without a pointer to references/"` — stop_if: Codex lists a dropped constraint
    - T9 think · T10 hunt — output: each rule once, no "why" sentences — check: same three commands — stop_if: same
  - phase check: `for f in skills/*/*/SKILL.md; do bash scripts/validate-skill.sh "$f" || echo FAIL "$f"; done; bash scripts/verify-doc-links.sh`

## Current State and Next Action
- active_phase: p1-sensors
- lifecycle_status: checked
- blockers: none
- open_items: failure-ledger row for the 5 gate rounds needs a committed SHA to cite; append after the p1 commit.
- exact_next_action: commit p1 (git), then work full phase p2-always-loaded

## Log
- 2026-09-29 plan written; previous active plan closed to completed/.
- 2026-09-29T08:00:50Z — p1-sensors — decision — heading renamed to `## Current State and Next Action` to match work-full.md; file was uncommitted.
- 2026-09-29T08:00:50Z — p1-sensors — start — T1, T2.
- 2026-09-29T08:03:14Z — p1-sensors/T1 — done — red: validator failed create-skill, create-cli, turbo-mono-platform ("Defers to non-existent skill(s): review"); green: all skills pass `validate-skill.sh`, scratch fixture deferring to `nosuch` exits 1. Surfaces: scripts/validate-skill.sh, 3 SKILL.md.
- 2026-09-29T08:03:14Z — p1-sensors/T2 — done — red: 3 SECRETS fixtures failed (function missing); green: `bash scripts/test-guards.sh` 58 passed, 0 failed; hook reinstalled and run on the real index: passed. Surfaces: scripts/install-git-hooks.sh (ZGUARD-CORE + pre-commit wiring), scripts/test-guards.sh, skills/workflow/git/SKILL.md.
- 2026-09-29T08:03:14Z — p1-sensors/T2 — decision — git Step 2 keeps a redacting scan (`git grep --cached -n … | cut -d: -f1-2`) instead of "the hook blocks": the skill runs in repos without this hook. R2's "shrinks to one line" is not met; the ~70-word saving is dropped (904 → 918 words).
- 2026-09-29T08:08:04Z — p1-sensors — decision — independent gate (Codex) REQUEST_CHANGES. Fixed: "++" content lines skipped, non-ASCII paths (now -z), binary blobs (now --text), rename untested; fixtures added (`bash scripts/test-guards.sh` 63 passed, 0 failed).
- 2026-09-29T08:08:04Z — p1-sensors — decision — owner approved: (a) *.example/*.sample/*.template exempt only from filename + url-credentials checks, AWS/private keys still rejected; (b) R2 deviation, git Step 2 keeps the redacting scan; (c) 3-line pre-commit wiring outside ZGUARD-CORE.
- 2026-09-29T08:09:39Z — p1-sensors — decision — re-gate REQUEST_CHANGES: textconv could hide a key; context lines did not advance the counter. Fixed (--no-textconv --no-ext-diff, interHunkContext=0, count context lines); fixture added, `bash scripts/test-guards.sh` 64 passed, 0 failed. Bash 3.2 runtime unverified (only bash 5.2 on this host).
- 2026-09-29T08:10:43Z — p1-sensors — decision — re-gate round 3 REQUEST_CHANGES: type changes (T) skipped. Fixed (--diff-filter=ACMRT) + symlink→file fixture; `bash scripts/test-guards.sh` 65 passed, 0 failed.
- 2026-09-29T08:12:01Z — p1-sensors — decision — re-gate round 4 REQUEST_CHANGES: pathspec magic in a staged filename bypassed the scan. Fixed (--literal-pathspecs) + fixture; `bash scripts/test-guards.sh` 66 passed, 0 failed.
- 2026-09-29T08:13:37Z — p1-sensors — decision — failures.md ledger rows deferred: its source rule requires an immutable commit SHA, and nothing is committed yet.

## Validation
- 2026-09-29T08:13:37Z — phase `p1-sensors` — verdict: APPROVED — mode: gate
  - `bash scripts/test-guards.sh` — 66 passed, 0 failed
  - `bash scripts/verify-doc-links.sh` — doc links OK (0 findings)
  - scope: on target — R1, R2 surfaces plus owner-approved 3-line pre-commit wiring
  - requirements: R1 met (validator rejects unknown Defer-To names; all skills pass) | R2 met with owner-approved deviation (git Step 2 keeps a redacting scan)
  - rollback_point: 8eae489
  - requests: none; rounds 1–4 REQUEST_CHANGES fixed (++ lines, quoted paths, binary, suffix exemption, textconv, context numbering, type change, pathspec magic)
  - residual: Bash 3.2 runtime unverified (host has bash 5.2 only)
  - judge: independent
  - judge_model: codex-cli 0.158.0 (default model)
