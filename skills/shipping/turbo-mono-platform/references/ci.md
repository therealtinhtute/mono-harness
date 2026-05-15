# CI/CD — GitHub Actions + Bun + Turbo

## Full CI Pipeline

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  ci:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0          # required for --affected

      - uses: oven-sh/setup-bun@v2
        with:
          bun-version: latest

      - name: Install dependencies
        run: bun install --frozen-lockfile

      - name: Lint
        run: bun run lint -- --affected
        env:
          TURBO_TOKEN: ${{ secrets.TURBO_TOKEN }}
          TURBO_TEAM: ${{ vars.TURBO_TEAM }}

      - name: Type check
        run: bun run type-check -- --affected
        env:
          TURBO_TOKEN: ${{ secrets.TURBO_TOKEN }}
          TURBO_TEAM: ${{ vars.TURBO_TEAM }}

      - name: Build
        run: bun run build -- --affected
        env:
          TURBO_TOKEN: ${{ secrets.TURBO_TOKEN }}
          TURBO_TEAM: ${{ vars.TURBO_TEAM }}
```

## Deploy to Vercel (Production + Preview)

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]
  pull_request:

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: oven-sh/setup-bun@v2

      - name: Install Vercel CLI
        run: bun add -g vercel

      - name: Pull Vercel env
        run: vercel pull --yes --environment=preview --token=${{ secrets.VERCEL_TOKEN }}

      - name: Build
        run: vercel build --token=${{ secrets.VERCEL_TOKEN }}

      - name: Deploy Preview
        if: github.ref != 'refs/heads/main'
        run: vercel deploy --prebuilt --token=${{ secrets.VERCEL_TOKEN }}

      - name: Deploy Production
        if: github.ref == 'refs/heads/main'
        run: vercel deploy --prebuilt --prod --token=${{ secrets.VERCEL_TOKEN }}
```

## Required GitHub Secrets

| Secret | Value | Where |
|--------|-------|-------|
| `TURBO_TOKEN` | Vercel remote cache token | Settings → Secrets |
| `TURBO_TEAM` | Vercel team slug | Settings → Variables |
| `VERCEL_TOKEN` | Vercel personal access token | Settings → Secrets |
| `VERCEL_ORG_ID` | From `.vercel/project.json` | Settings → Variables |
| `VERCEL_PROJECT_ID` | From `.vercel/project.json` | Settings → Variables |

## Skip CI with turbo-ignore

```yaml
- name: Check if web changed
  id: changed
  run: bunx turbo-ignore web --fallback=HEAD~1
  continue-on-error: true

- name: Build web
  if: steps.changed.outcome == 'failure'
  run: bun run build -- --filter=web
```

## Local Cache in CI (no Vercel remote cache)

```yaml
- name: Cache turbo
  uses: actions/cache@v4
  with:
    path: .turbo
    key: turbo-${{ runner.os }}-${{ hashFiles('**/bun.lockb') }}-${{ github.sha }}
    restore-keys: |
      turbo-${{ runner.os }}-${{ hashFiles('**/bun.lockb') }}-
      turbo-${{ runner.os }}-
```

## DB Migration in CI

```yaml
- name: Run DB migrations
  run: bun run db:migrate
  env:
    DATABASE_URL: ${{ secrets.DATABASE_URL }}
```

## Root `package.json` CI scripts

```json
{
  "scripts": {
    "ci:lint": "turbo lint --affected",
    "ci:build": "turbo build --affected",
    "ci:typecheck": "turbo type-check --affected"
  }
}
```
