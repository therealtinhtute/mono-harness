// ❌ Old Next.js 15 caching pattern — must be migrated
// Used by eval #7: Claude must rewrite using "use cache" directive

import { unstable_cache } from "next/cache";
import { revalidateTag } from "next/cache";
import { db } from "@rp/db";

// Wrong: unstable_cache, single-arg revalidateTag
export const getPosts = unstable_cache(
  async () => {
    return db.query.posts.findMany({ limit: 10 });
  },
  ["posts"],
  { tags: ["posts"] }
);

export async function invalidatePosts() {
  revalidateTag("posts"); // ❌ missing second arg "max"
}
