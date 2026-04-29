# Dev Commands

```bash
# Analyze project → define package tiers
bash scripts/analyze-project.sh

# Scaffold base tier (always first)
bash scripts/scaffold-base.sh <project-name>

# Scaffold optional packages (after analysis)
bash scripts/scaffold-packages.sh <project-name> <pkg1,pkg2,...>

# Generate Drizzle schema + query helpers
bash scripts/gen-schema.sh <table-name>

# Generate tRPC router boilerplate
bash scripts/gen-trpc-router.sh <router-name>

# Setup @rp/env with t3-env
bash scripts/gen-env.sh <project-name>

# Setup Better Auth full configuration
bash scripts/gen-auth.sh <project-name>

# Setup Hono API with middleware
bash scripts/gen-hono-api.sh <project-name>

# Run Drizzle migrations
bash scripts/migrate.sh

# Add new @rp/<n> package
bash scripts/gen-package.sh @rp/<n>

# Validate no relative cross-package imports
bash scripts/check-imports.sh

# Run Biome check (root task)
bun run check
bun run check:fix
```
