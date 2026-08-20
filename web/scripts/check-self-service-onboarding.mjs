import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const start = readFileSync(resolve("app/start/page.tsx"), "utf8");
const onboarding = readFileSync(resolve("app/onboarding/page.tsx"), "utf8");
const api = readFileSync(resolve("lib/api.ts"), "utf8");
const login = readFileSync(resolve("app/login/page.tsx"), "utf8");
const authGate = readFileSync(resolve("components/auth-gate.tsx"), "utf8");
const logout = readFileSync(resolve("app/api/auth/logout/route.ts"), "utf8");
const demo = readFileSync(resolve("components/flyer-survey.tsx"), "utf8");
const gateway = readFileSync(resolve("app/api/[...path]/route.ts"), "utf8");
const publicAuthSources = [start, login, api];

assert.match(start, /createLocalRegistration/);
assert.match(start, /isProductionAuth/);
assert.match(start, /identityAuthStartUrl/);
assert.match(start, /name="email"/);
assert.match(start, /Create my heard account/);
assert.doesNotMatch(start, /Passage/);
assert.match(start, /No call or approval required/);
assert.match(start, /\/walkthrough/);
assert.doesNotMatch(start, /contact#walkthrough/);
assert.match(onboarding, /\/api\/v1\/onboarding\/activations/);
assert.match(onboarding, /idempotencyKey/);
assert.match(onboarding, /\/api\/v1\/survey-campaigns/);
assert.match(onboarding, /\/api\/v1\/feedback-links/);
assert.match(onboarding, /QR generation is temporarily unavailable/);
assert.match(api, /\/api\/v1\/auth\/local\/registration/);
assert.match(api, /\/api\/v1\/auth\/start/);
assert.match(api, /safeReturnTo/);
assert.match(api, /getStoredSession\(\)[\s\S]*?typeof window === "undefined" \|\| isProductionAuth\(\)/);
assert.match(api, /storeSession\(session: Session\)[\s\S]*?if \(isProductionAuth\(\)\) return;/);
assert.match(login, /identityAuthStartUrl/);
assert.match(login, /isProductionAuth\(\)/);
assert.match(login, /name="email"/);
assert.doesNotMatch(login, /Passage/);
for (const source of publicAuthSources) {
  assert.doesNotMatch(source, /auth\/passage/i);
}
assert.match(api, /resolveSession/);
assert.match(authGate, /resolveSession/);
assert.match(logout, /heard_session/);
assert.match(gateway, /"cookie"/);
assert.doesNotMatch(api, /createPassageAuthRequest|exchangePassageCode|oauth\/token|heard-passage-pkce/);
assert.doesNotMatch(start, /sessionStorage\.setItem\("heard-passage/);
assert.doesNotMatch(login, /sessionStorage\.setItem\("heard-passage/);
assert.match(demo, /\/start\?source=guest_demo/);
assert.match(gateway, /"idempotency-key"/);

console.log("Self-service onboarding contract and conversion paths verified.");
