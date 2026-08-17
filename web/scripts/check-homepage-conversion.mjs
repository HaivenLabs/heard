import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const source = readFileSync(resolve("app/page.tsx"), "utf8");

assert.doesNotMatch(source, /Live signal/i);
assert.doesNotMatch(source, /Recovery ready/i);
assert.match(source, />Just received</);
assert.match(source, />Needs follow-up</);
assert.match(source, />Sent to Recovery inbox</);

assert.ok((source.match(/\/contact/g) ?? []).length >= 3, "Primary navigation must include the restaurant contact path");
assert.ok((source.match(/>See heard for your restaurant</g) ?? []).length >= 2, "Restaurant CTA must appear in the header and hero");
assert.ok((source.match(/>Try guest experience</g) ?? []).length >= 2, "Guest demo CTA must appear in the header and hero");
assert.match(source, /href=["']\/login["'][^>]*>Sign in</);

console.log("Homepage recovery story and conversion paths verified.");
