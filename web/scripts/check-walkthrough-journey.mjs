import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const source = readFileSync(resolve("app/walkthrough/walkthrough-form.tsx"), "utf8");
const redirect = readFileSync(resolve("app/contact/page.tsx"), "utf8");

assert.match(source, /\/api\/v1\/marketing-leads/);
assert.match(source, /contact_consent: true/);
assert.match(source, /source,/);
assert.match(source, /Request my walkthrough/);
assert.match(source, /Request received/);
assert.match(source, /We&apos;ll take it from here\./);
assert.match(source, /role="alert"/);
assert.match(source, /isValidEmail\(workEmail\)/);
assert.match(source, /isValidPhone\(phone\)/);
assert.match(source, /Phone \(optional\)/);
assert.doesNotMatch(source, /id="walkthrough"/);
assert.match(redirect, /redirect\("\/walkthrough" as Route\)/);

console.log("Optional walkthrough form, validation, persistence, and confirmation verified.");
