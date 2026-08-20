import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const campaignBuilder = readFileSync(resolve("app/admin/campaigns/page.tsx"), "utf8");

assert.match(campaignBuilder, /googleReviewURL: "https:\/\/maps\.app\.goo\.gl\/D3cEeXBEtGaKF2Lz8"/);
assert.match(campaignBuilder, /yelpReviewURL: "https:\/\/www\.yelp\.com\/biz\/nom-san-juan-capistrano"/);

console.log("Demo review destinations verified.");
