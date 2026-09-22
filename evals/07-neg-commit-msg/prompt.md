---
max_turns: 6
timeout_seconds: 120
allowed_tools: [Skill, Read, Glob]
model: sonnet
runs: 3
plugins: [../../skills/craft/write]
---
Write a commit message for this diff:

diff --git a/src/api/paginate.ts b/src/api/paginate.ts
@@ -12,7 +12,7 @@ export function paginate<T>(items: T[], page: number, size: number): T[] {
-  const start = page * size + 1;
+  const start = (page - 1) * size;
   return items.slice(start, start + size);
 }
