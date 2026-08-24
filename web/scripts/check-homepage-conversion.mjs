import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const source = readFileSync(resolve("app/page.tsx"), "utf8");
const headerSource = readFileSync(resolve("components/public-header.tsx"), "utf8");
const firstSectionEnd = source.indexOf("</section>");

assert.doesNotMatch(source, /Live signal/i);
assert.doesNotMatch(source, /Recovery ready/i);
assert.match(source, />Just received</);
assert.match(source, />Needs follow-up</);
assert.match(source, />Sent to Recovery inbox</);
assert.match(source, /rating === 3 \? "drop-shadow/);
assert.doesNotMatch(source, /rating === 2 \? "drop-shadow/);

assert.match(headerSource, /\/start\?source=homepage/);
assert.match(headerSource, />Start using heard</);
assert.match(headerSource, /\/walkthrough/);
assert.doesNotMatch(headerSource, /contact#walkthrough/);
assert.match(headerSource, />Try guest experience</);
assert.match(headerSource, /href=["']\/login["'][^>]*>Sign in</);
assert.match(source, />Start using heard</);
assert.match(source, />Try guest experience</);
assert.match(source, />Request a walkthrough</);
assert.ok(source.indexOf("From signal to save") < firstSectionEnd, "The signal-to-save story must remain in the opening viewport section.");
assert.ok(source.indexOf("Route the signal") < firstSectionEnd, "The complete signal-to-save sequence must appear in the opening viewport section.");
assert.ok(source.lastIndexOf("lg:grid-cols-[0.7fr_1fr]", source.indexOf("From signal to save")) !== -1, "The signal-to-save story must use the full-width layout.");

console.log("Homepage recovery story and conversion paths verified.");
