import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const expectedConsumers = [
  "app/page.tsx",
  "app/admin/campaigns/page.tsx",
  "app/f/[token]/page.tsx"
];

for (const relativePath of expectedConsumers) {
  const source = readFileSync(resolve(relativePath), "utf8");
  assert.match(source, /import \{ RatingFace, RatingValue \} from ["'][^"']*components\/rating-face["'];/);
  assert.match(source, /<RatingFace\b/);
  assert.doesNotMatch(source, /<svg[^>]*viewBox=["']0 0 100 106["']/);
}

console.log("Shared RatingFace usage verified across guest, campaign, and marketing surfaces.");
