#!/usr/bin/env bash
set -e
TABLE="${1:?Usage: gen-schema.sh <table-name>}"
PASCAL="$(echo "$TABLE" | sed -r 's/(^|_)([a-z])/\U\2/g')"
FILE="packages/db/src/schema/${TABLE}.ts"
QUERY="packages/db/src/queries/${TABLE}.ts"

mkdir -p packages/db/src/{schema,queries}

cat > "$FILE" << EOF
import { pgTable, text, timestamp, uuid } from "drizzle-orm/pg-core";

export const ${TABLE} = pgTable("${TABLE}", {
  id: uuid("id").primaryKey().defaultRandom(),
  createdAt: timestamp("created_at").defaultNow().notNull(),
  updatedAt: timestamp("updated_at").defaultNow().notNull(),
});

export type ${PASCAL} = typeof ${TABLE}.\$inferSelect;
export type New${PASCAL} = typeof ${TABLE}.\$inferInsert;
EOF

cat > "$QUERY" << EOF
import { eq } from "drizzle-orm";
import { db } from "../index";
import { ${TABLE} } from "../schema/${TABLE}";
import type { New${PASCAL} } from "../schema/${TABLE}";

export async function get${PASCAL}ById(id: string) {
  return db.query.${TABLE}.findFirst({ where: eq(${TABLE}.id, id) });
}

export async function create${PASCAL}(data: New${PASCAL}) {
  const [row] = await db.insert(${TABLE}).values(data).returning();
  return row;
}

export async function delete${PASCAL}(id: string) {
  await db.delete(${TABLE}).where(eq(${TABLE}.id, id));
}
EOF

echo "✅ Created:"
echo "   $FILE"
echo "   $QUERY"
echo ""
echo "Add to packages/db/src/schema/index.ts:"
echo "  export * from './${TABLE}';"
