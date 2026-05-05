# Gate Checklist — Per-Stack Commands

## Conditional Reorder Rules

Apply before running checks:
- Auth / data change → secrets scan first, then tests
- Time-critical → tests only, skip lint / polish
- API change → backward-compat check before lint

## Gate Outcome

| Result | Decision |
|--------|----------|
| All pass + on target | ✅ APPROVED — ready to commit |
| Minor issues remain | ⚠️ FIX — return to implementation |
| Major gaps | ❌ NEEDS_WORK — re-scope with `plan` |

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

## Backward-Compat Check (API changes)
```bash
# Grep all callers of the changed interface
grep -r "functionName\|ClassName\|endpoint" . --include="*.ts" -l
```
