import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const layoutSource = readFileSync(resolve("app/admin/layout.tsx"), "utf8");
assert.match(layoutSource, /import \{ AdminShell \} from ["'][^"']*components\/admin-shell["'];/);
assert.match(layoutSource, /import \{ AuthGate \} from ["'][^"']*components\/auth-gate["'];/);
assert.match(layoutSource, /<AdminShell session=\{session\}>/);
assert.match(layoutSource, /<AdminSessionProvider session=\{session\}>/);

const adminPages = [
  "app/admin/page.tsx",
  "app/admin/campaigns/page.tsx",
  "app/admin/recovery/page.tsx"
];

for (const relativePath of adminPages) {
  const source = readFileSync(resolve(relativePath), "utf8");
  assert.doesNotMatch(source, /components\/admin-shell/);
  assert.doesNotMatch(source, /components\/auth-gate/);
  assert.doesNotMatch(source, /<AdminShell\b/);
  assert.doesNotMatch(source, /<AuthGate\b/);
  assert.doesNotMatch(source, /<main className=["'][^"']*min-h-screen[^"']*bg-/);
}

console.log("Admin layout ownership verified across authenticated admin surfaces.");
