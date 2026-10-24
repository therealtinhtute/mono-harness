# 0017 — No project-verification skill; the evidenced gaps are executable and already gated

## Status

Accepted 2026-10-24 on merge of the PR that resolves issue #155. Decision: reject a new skill.

## Context

Issue #155 asks whether a project-specific verification skill fills a gap that the check matrix, repository scripts and hooks leave open, and says to add one only if a concrete gap remains. The audit rates verification "mono-harness ahead" and says to keep it (`docs/research/pstack-concepts-audit.md:32`).

Failure history searched (fix commits in the log, the audit's findings F1 and F2):

1. A skill validator that never ran. `f5af8e8` revived `scripts/ci/validate-skill.sh` after the pre-commit hook pointed at a stale path. Now the hook calls it (`setup/install-git-hooks.sh:559`) and `scripts/ci/test-guards.sh:944` proves it fails on a bad fixture.
2. Embedded playbook contracts drifting from reworded skills. `4d0beb3` fixed a contract that no longer matched its playbook. `AGENTS.md` now says to run `cd cli && cargo test` after editing `skills/stages/*/references/`, and the `cli-ci` workflow runs on `skills/stages/**` (`.github/workflows/cli-ci.yml:9`, `:17`).
3. A stale installed hook wrapper (audit F1). The CI job `hook-guard` re-runs the guard core, so a stale wrapper delays rejection but does not bypass it.
4. Auto-mode routing wrote plan state (`1fbf4ac`, `fcdae1f`, `0dfb5b2`). The fix lives in the `check.md` preconditions, a rule in the existing playbook.

One residual gap is reproducible: `validate-skill.sh` runs only from the optional pre-commit hook. `grep -rc validate-skill .github/workflows` returns 0 for every workflow file. A malformed `SKILL.md` outside `skills/stages/**` is therefore not caught on a pull request from a clone that has no hook. The script reserves errors for load-time breakage (`scripts/ci/validate-skill.sh:4`).

## Decision

Do not add a project-verification skill. Every gap in the history was closed by an executable guard, a CI job, or a line in an existing playbook. A skill is prompt text: it cannot enforce what a script and a CI job do, and it would repeat the proof classes in `check.md` step 6 and the gate list in `AGENTS.md`. That is the duplicate verification layer the issue warns against.

Rejected: a skill that lists project constraints (duplicates `AGENTS.md` Gate Commands and `check.md`), and a skill wrapping the existing scripts (adds a second entry point to the same commands).

The residual gap is a CI wiring question, not a skill. It is recorded here and not changed in this PR, because adding a CI job is an enforcement change that needs its own negative fixture and the owner's call. IF the owner wants it, the smallest fix is one `docs-ci` or `cli-ci` step that runs `validate-skill.sh` on changed `SKILL.md` files, reusing the existing bad-skill fixture.

## Consequences

No new surface to maintain and no second description of the gates. The cost: the residual gap stays open until the owner chooses to wire the validator into CI, and a pull request that changes a skill outside `skills/stages/**` without the hook is validated only by review.

## Authority

`docs/research/pstack-concepts-audit.md:32`; commits `f5af8e8`, `4d0beb3`; `setup/install-git-hooks.sh:559`; `scripts/ci/test-guards.sh:944`; `.github/workflows/cli-ci.yml:9`; `AGENTS.md` Layout and Gate Commands; issue #155 (acceptance criteria resolved by this record).
