# Project Context Extraction Guide

Extract repo constraints before reviewing. Goal: compress repo-specific rules into a brief
context block so the review is accurate without reading everything.

## What to Read

Read only files relevant to the changed code:

| File | Extract |
|------|---------|
| `README.md` | Framework, dev commands, test commands |
| `AGENTS.md` / `CLAUDE.md` | Project-specific rules that override this skill |
| `package.json` / `Cargo.toml` / `pyproject.toml` | Scripts, dependencies |
| `.github/workflows/` | Build, test, deploy commands |
| `CHANGELOG.md` | Release conventions and version format |

## What to Compress

After reading, produce a one-paragraph context block:

```
verify_cmd:       [e.g. npm test && tsc --noEmit]
protected_files:  [e.g. dist/, generated/, CHANGELOG.md]
domain_risk:      [e.g. auth middleware, payment flow]
harness_mode:     full / partial / none
release_format:   [e.g. semver tag + CHANGELOG section]
```

## Conflict Rule

When project context and this skill overlap, apply the stricter rule.

If `AGENTS.md` or `CLAUDE.md` defines a verification command → use that, not auto-detection.
If project docs say never auto-commit → skip any autofix that would commit.
If `.kit/planning/` + `.kit/runs/work/` are present → treat artifact alignment as part of the gate, not an optional note.

## Skip Context Extraction When

- Diff is under 30 lines and does not touch config, auth, or CI
- Running `gate` mode only (checks don't need project context)

## Harness Detection

Classify the repo before review:
- `full` — `.kit/planning/SPEC.md` plus roadmap/phase artifacts exist, and `work` run logs are used
- `partial` — some planning artifacts exist, but run logs or phase artifacts are incomplete
- `none` — no harness artifacts present

If harness mode is `full` or `partial`, check `references/artifact-alignment.md` before final sign-off.
