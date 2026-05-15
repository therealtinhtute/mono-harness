#!/usr/bin/env bash
set -e
ROUTER="${1:?Usage: gen-trpc-router.sh <router-name>}"
PASCAL="$(echo "$ROUTER" | sed -r 's/(^|_)([a-z])/\U\2/g')"
FILE="packages/trpc/src/routers/${ROUTER}.ts"

mkdir -p packages/trpc/src/routers

cat > "$FILE" << EOF
import { z } from "zod";
import { router, publicProcedure, protectedProcedure } from "../init";

export const ${ROUTER}Router = router({
  list: publicProcedure.query(async ({ ctx }) => {
    return [];
  }),

  getById: publicProcedure
    .input(z.object({ id: z.string().uuid() }))
    .query(async ({ input }) => {
      return null;
    }),

  create: protectedProcedure
    .input(z.object({ name: z.string().min(1) }))
    .mutation(async ({ ctx, input }) => {
      return { id: crypto.randomUUID(), ...input };
    }),

  delete: protectedProcedure
    .input(z.object({ id: z.string().uuid() }))
    .mutation(async ({ input }) => {
      return { id: input.id };
    }),
});
EOF

echo "✅ Created $FILE"
echo ""
echo "Add to packages/trpc/src/routers/index.ts:"
echo "  import { ${ROUTER}Router } from './${ROUTER}';"
echo "  export const appRouter = router({ ..., ${ROUTER}: ${ROUTER}Router });"
