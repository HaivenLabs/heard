import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const adminPages = [
  "app/admin/page.tsx",
  "app/admin/campaigns/page.tsx",
  "app/admin/recovery/page.tsx"
];

for (const relativePath of adminPages) {
  const source = readFileSync(resolve(relativePath), "utf8");
  assert.match(source, /import \{ AdminShell \} from ["'][^"']*components\/admin-shell["'];/);
  assert.match(source, /<AdminShell session=\{session\}>/);
  assert.doesNotMatch(source, /<main className=["'][^"']*min-h-screen[^"']*bg-/);
}

console.log("Shared AdminShell usage verified across authenticated admin surfaces.");
