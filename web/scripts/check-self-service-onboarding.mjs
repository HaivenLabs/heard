import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const start = readFileSync(resolve("app/start/page.tsx"), "utf8");
const onboarding = readFileSync(resolve("app/onboarding/page.tsx"), "utf8");
const api = readFileSync(resolve("lib/api.ts"), "utf8");
const login = readFileSync(resolve("app/login/page.tsx"), "utf8");
const demo = readFileSync(resolve("app/f/[token]/page.tsx"), "utf8");
const gateway = readFileSync(resolve("app/api/[...path]/route.ts"), "utf8");

assert.match(start, /createLocalRegistration/);
assert.match(start, /isProductionAuth/);
assert.match(start, /passageHostedUrl\("signup"/);
assert.match(start, /No call or approval required/);
assert.match(start, /\/contact#walkthrough/);
assert.match(onboarding, /\/api\/v1\/onboarding\/activations/);
assert.match(onboarding, /idempotencyKey/);
assert.match(onboarding, /\/api\/v1\/survey-campaigns/);
assert.match(onboarding, /\/api\/v1\/feedback-links/);
assert.match(onboarding, /QR generation is temporarily unavailable/);
assert.match(api, /\/api\/v1\/auth\/local\/registration/);
assert.match(api, /NEXT_PUBLIC_PASSAGE_HOSTED_URL/);
assert.match(api, /safeReturnTo/);
assert.match(login, /passageHostedUrl\("login"/);
assert.match(login, /isProductionAuth\(\)/);
assert.match(demo, /\/start\?source=guest_demo/);
assert.match(gateway, /"idempotency-key"/);

console.log("Self-service onboarding contract and conversion paths verified.");
