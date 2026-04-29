// ❌ This file has bad relative cross-package imports
// Used by eval #2 to test check-imports.sh detection

import { db } from "../../packages/db/src";
import { auth } from "../../packages/auth/src/index";
import { env } from "../../../packages/env/src";
import { redis } from "../../packages/kv/src/index";

export async function getUser(id: string) {
  const session = await auth.api.getSession({ headers: new Headers() });
  return db.query.users.findFirst({ where: (u) => u.id === id });
}
