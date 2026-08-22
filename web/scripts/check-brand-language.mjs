import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const home = readFileSync(resolve("app/page.tsx"), "utf8");
const walkthrough = readFileSync(resolve("app/walkthrough/walkthrough-form.tsx"), "utf8");
const adminShell = readFileSync(resolve("components/admin-shell.tsx"), "utf8");
const publicHeader = readFileSync(resolve("components/public-header.tsx"), "utf8");
const login = readFileSync(resolve("app/login/page.tsx"), "utf8");
const start = readFileSync(resolve("app/start/page.tsx"), "utf8");

for (const source of [home, walkthrough]) {
  assert.match(source, /components\/public-header/);
  assert.match(source, /components\/brand-backdrop/);
  assert.doesNotMatch(source, /<header\b/);
}

for (const source of [login, start]) {
  assert.match(source, /components\/public-header/);
  assert.match(source, /<PublicHeader \/>/);
}

assert.match(home, /<h1[^>]*>The guest experience, in your hands<\/h1>/);
assert.match(walkthrough, /<h1[^>]*>Built for restaurant operators, by restaurant operators\.<\/h1>/);
assert.doesNotMatch(walkthrough, /bg-\[#17251d\]/);
assert.doesNotMatch(walkthrough, /function Outcome/);
assert.doesNotMatch(walkthrough, /guest demo again/i);
assert.match(walkthrough, />Try guest experience<\/Link>/);

assert.match(adminShell, /brand-backdrop/);
assert.match(publicHeader, />Try guest experience<\/Link>/);
assert.match(publicHeader, />Start using heard<\/Link>/);
assert.match(publicHeader, />Walkthrough<\/Link>/);

console.log("Shared brand language verified across public and admin surfaces.");
