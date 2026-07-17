# Gate Checklist — Per-Stack Commands

## Validation Matrix (harness-aware gate)

When a `zharness` binary passes the version gate and `.kit/planning/` artifacts exist, the automated gate evaluates this lane × proof-class matrix instead of (not in addition to) the generic pass/fail table below. Lane comes from `.kit/planning/SPEC.md`'s frontmatter `lane:` field (set by `intake --lane` at brainstorm time — there is no live CLI query for it, SPEC.md's frontmatter is the source of truth). Every cell is `required` (must have matching evidence or the gate is FAIL), `optional` (nice to have, absence never fails the gate), or `n-a` (not expected for this lane, never requested):

| Lane \ Proof class | unit | integration | e2e | manual-check | command-output |
|---|---|---|---|---|---|
| tiny | optional | n-a | n-a | optional | required |
| normal | required | optional | n-a | optional | required |
| high-risk | required | required | optional | required | required |

Proof-class meaning:
- `unit` — a unit test run covering the changed behavior (`go test ./... -run ...`, `npm test`, etc.)
- `integration` — a test that crosses a real boundary (DB, filesystem, another package's public API) rather than mocking it
- `e2e` — a full user-facing flow exercised end-to-end (browser automation, CLI smoke test against a real binary, etc.)
- `manual-check` — the Phase 2 code-review pass itself (Security/Performance/Architecture/Code Quality from `review-dimensions.md`) counts as this class's evidence; a clean review with no 🔴/🟠 findings satisfies it, a 🔴 finding never satisfies it regardless of lane
- `command-output` — any verification command's actual captured output (build, lint, a one-off script) — the floor every lane always requires, since it's the cheapest proof and the core "no unverified claims" hard stop already demands it

A `required` cell with no matching proof link ⇒ gate FAIL naming that exact missing evidence class (see `check/SKILL.md`'s Harness Gate Flow). This is a hard rule, not a suggestion — the skill does not use judgment to wave through a missing required proof; a human overrides via `zharness intervention` if the missing proof is genuinely acceptable to skip.

## Conditional Reorder Rules

Apply before running checks:
- Auth / data change → secrets scan first, then tests
- Time-critical → tests only, skip lint / polish
- API change → backward-compat check before lint

## Gate Outcome

| Result | Decision |
|--------|----------|
| All pass + on target + artifact aligned (when harness applies) | ✅ APPROVED — ready to commit |
| Minor issues remain | ⚠️ FIX — return to implementation |
| Major gaps | ❌ NEEDS_WORK — re-scope with `to-plan` or re-run `work` |

## Node.js / TypeScript
```bash
npm test                      # or: yarn test / pnpm test
npx tsc --noEmit              # type check
npx eslint . --max-warnings 0
npm run build
```

## Python
```bash
pytest                        # or: python -m pytest
mypy .                        # type check
ruff check .                  # or: flake8
```

## Go
```bash
go test ./...
go vet ./...
staticcheck ./...
go build ./...
```

## Rust
```bash
cargo test
cargo clippy -- -D warnings
cargo build
```

## General: Secrets Scan (run first for auth / data changes)
```bash
git diff HEAD | grep -iE "(password|secret|token|api_key|private_key)" | grep "^\+"
```

## Fallback: No Known Stack
1. Check `package.json` → `scripts.test` / `scripts.build`
2. Check `Makefile` → `make test` / `make build`
3. Check `README.md` → look for "Running tests" section
4. If nothing found: document as `verification: none — no command detected`
   Never claim pass without running a real command.

## Harness Add-on (when `.kit/planning/` is present)
- Read the active phase plan and collect the expected verification commands
- Compare them against the latest matching `.kit/runs/work/*.md`
- If code changed but the proof trail is missing, label `artifact_alignment: ❌ drift` even if local tests pass
- If the diff exceeds allowed surfaces, stop before normal review and route back to `to-plan phase {slug}` or `work`

## Backward-Compat Check (API changes)
```bash
# Grep all callers of the changed interface
grep -r "functionName\|ClassName\|endpoint" . --include="*.ts" -l
```
