import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const campaignBuilder = readFileSync(resolve("app/admin/campaigns/page.tsx"), "utf8");
const onboarding = readFileSync(resolve("app/onboarding/page.tsx"), "utf8");
const adminShell = readFileSync(resolve("components/admin-shell.tsx"), "utf8");
const overview = readFileSync(resolve("app/admin/page.tsx"), "utf8");

for (const source of [campaignBuilder, onboarding]) {
  assert.doesNotMatch(source, /maps\.app\.goo\.gl\/D3cEeXBEtGaKF2Lz8/);
  assert.doesNotMatch(source, /yelp\.com\/biz\/nom-san-juan-capistrano/);
}
assert.doesNotMatch(campaignBuilder, /restaurantName: "nom"|logoURL: "\/brands\/nom/);
assert.doesNotMatch(adminShell, /North Star Noodles/);
assert.doesNotMatch(overview, /North Star Noodles/);

console.log("Customer onboarding is free of seeded demo ownership and destinations.");
