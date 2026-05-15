#!/usr/bin/env bash
set -e
NAME="${1:?Usage: gen-package.sh @rp/<n>}"
SHORT="${NAME#@rp/}"
DIR="packages/$SHORT"

[ -d "$DIR" ] && { echo "❌ $DIR already exists"; exit 1; }
mkdir -p "$DIR/src"

cat > "$DIR/package.json" << EOF
{
  "name": "$NAME",
  "private": true,
  "exports": { ".": "./src/index.ts" },
  "scripts": {
    "lint": "biome check src",
    "type-check": "tsc --noEmit"
  },
  "devDependencies": {
    "@rp/typescript-config": "workspace:*",
    "@rp/biome-config": "workspace:*"
  }
}
EOF

cat > "$DIR/tsconfig.json" << 'EOF'
{
  "extends": "@rp/typescript-config/base.json",
  "compilerOptions": { "baseUrl": ".", "paths": { "@/*": ["./src/*"] } },
  "include": ["src"]
}
EOF

echo "export {};" > "$DIR/src/index.ts"

echo "✅ Created $DIR"
echo "   Add dependencies: bun add <pkg> --cwd $DIR"
echo "   Run: bun install"
