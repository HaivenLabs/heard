import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const home = readFileSync(resolve("app/page.tsx"), "utf8");
const contact = readFileSync(resolve("app/contact/contact-form.tsx"), "utf8");
const adminShell = readFileSync(resolve("components/admin-shell.tsx"), "utf8");
const publicHeader = readFileSync(resolve("components/public-header.tsx"), "utf8");

for (const source of [home, contact]) {
  assert.match(source, /components\/public-header/);
  assert.match(source, /components\/brand-backdrop/);
  assert.doesNotMatch(source, /<header\b/);
}

assert.match(home, /<h1[^>]*>The guest experience, in your hands<\/h1>/);
assert.match(contact, /<h1[^>]*>Built for restaurant operators, by restaurant operators\.<\/h1>/);
assert.doesNotMatch(contact, /bg-\[#17251d\]/);
assert.doesNotMatch(contact, /function Outcome/);

assert.match(adminShell, /brand-backdrop/);
assert.match(publicHeader, />Try guest experience<\/Link>/);
assert.match(publicHeader, />See heard for your restaurant<\/Link>/);

console.log("Shared brand language verified across public and admin surfaces.");
