#!/usr/bin/env bash
set -e
ACTION="${1:-push}"

case "$ACTION" in
  push)
    echo "⬆️  Pushing schema to DB..."
    bun run db:push
    ;;
  generate)
    echo "📝 Generating migrations..."
    bun run db:generate
    ;;
  studio)
    echo "🎨 Opening Drizzle Studio..."
    bun run db:studio
    ;;
  *)
    echo "Usage: migrate.sh [push|generate|studio]"
    echo "  push     → push schema directly (dev only)"
    echo "  generate → generate migration files"
    echo "  studio   → open Drizzle Studio UI"
    exit 1
    ;;
esac
