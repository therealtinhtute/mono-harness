#!/usr/bin/env bash
# Interactive analysis — meant to be READ by Claude, not run standalone
# Claude uses this to structure the Package Analyzer interview

cat << 'EOF'
🔍 Package Analyzer — turbo-mono-platform

BASE TIER (always included):
  ✅ toolings/typescript-config
  ✅ toolings/biome-config
  ✅ packages/ui  (@rp/ui, shadcn preset b1tMcUv91, pre-init)
  ✅ apps/web     (Next.js 16 + demo page)

OPTIONAL TIER — answer questions to determine:

  Q1: Project type?
      [saas] [internal] [marketplace] [ecommerce] [api-only] [content]

  Q2: Features needed?
      [auth] [db+queries] [realtime] [rate-limiting]
      [feature-flags] [rpc-layer] [hono-api] [payments]
      [email] [file-upload] [background-jobs]

  Q3: Domain entities? (e.g., users, posts, orders)
      Each entity → gen-schema.sh + gen-trpc-router.sh

  Q4: Standalone API needed?
      [yes → apps/api (Hono)] [no → Next.js API routes]

  Q5: Deployment?
      [vercel] [vps] [docker]

PACKAGE DECISION RULES:
  @rp/env          → ALWAYS (every project)
  @rp/supabase     → if any DB, queries, or realtime needed
  @rp/auth         → if user authentication needed
  @rp/kv           → if rate limiting, caching, or feature flags
  @rp/trpc         → if client-side data fetching with type safety
  @rp/validators   → if shared Zod schemas across packages
  @rp/remote-config→ if feature flags needed (requires @rp/kv)
  apps/api         → if mobile clients or standalone REST API

SCAFFOLD COMMANDS (after confirmation):
  bash scripts/scaffold-base.sh <project-name>
  bash scripts/scaffold-packages.sh <project-name> <pkg1,pkg2,...>

Per entity:
  bash scripts/gen-schema.sh <entity>
  bash scripts/gen-trpc-router.sh <entity>

EOF
