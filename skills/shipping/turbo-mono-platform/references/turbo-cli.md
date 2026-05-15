# Turbo CLI — Filter, Affected, Debug

## --filter Syntax

```bash
# Single package by name
bun run build -- --filter=web
bun run build -- --filter=@rp/ui

# All in a directory
bun run build -- --filter=./apps/*
bun run build -- --filter=./packages/*

# Package + all its dependencies (...)
bun run build -- --filter=web...

# Package + all its dependents (reverse)
bun run build -- --filter=...@rp/ui

# Exclude a package
bun run build -- --filter=!@rp/deprecated

# Combine filters
bun run build -- --filter=@rp/ui --filter=web
```

## --affected — Changed Packages Only

```bash
# Changed since last commit (CI default)
bun run build -- --affected

# Changed since specific branch
bun run test -- --affected --filter="...[origin/main]"

# Changed in specific commit range
bun run test -- --filter="...[HEAD~3]"
```

Requires full git history — use `fetch-depth: 0` in GitHub Actions.

## turbo-ignore — Skip CI When Nothing Changed

```bash
# Check if web app changed (returns exit 1 if changed, 0 if not)
bunx turbo-ignore web

# Check specific task
bunx turbo-ignore --task=test

# Compare to specific branch
bunx turbo-ignore --fallback=main
```

Use in GitHub Actions:
```yaml
- name: Check changes
  id: check
  run: bunx turbo-ignore web
  continue-on-error: true

- name: Deploy
  if: steps.check.outcome == 'failure'  # changed = deploy
  run: bun run deploy
```

## Concurrency

```bash
# Limit parallel tasks (useful in CI with limited RAM)
bun run build -- --concurrency=4
bun run build -- --concurrency=50%   # 50% of CPU cores
```

## Common Workflows

```bash
# Dev all apps
bun run dev

# Build only changed packages vs main branch
bun turbo build --affected --filter="...[origin/main]"

# Lint only packages that import @rp/ui (dependents)
bun turbo lint --filter=...@rp/ui

# Type-check a specific app + all its package deps
bun turbo type-check --filter=web...

# Run test only for apps/ (not packages/)
bun turbo test --filter=./apps/*

# Clean all build artifacts
bun turbo clean && rm -rf node_modules .turbo
```

## Boundaries (Experimental — Turbo 2.9)

Enforce package import rules:

```json
{
  "boundaries": {
    "tags": {
      "app": { "cannotDependOn": ["app"] },
      "package": { "canDependOn": ["package"] }
    }
  }
}
```

Tag packages in `package.json`:
```json
{ "turbo": { "tags": ["app"] } }
```

Check violations:
```bash
bunx turbo boundaries
```
