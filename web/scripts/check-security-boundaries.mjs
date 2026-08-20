import assert from "node:assert/strict";
import { readFileSync } from "node:fs";

const campaign = readFileSync(new URL("../app/admin/campaigns/page.tsx", import.meta.url), "utf8");
const proxy = readFileSync(new URL("../app/api/[...path]/route.ts", import.meta.url), "utf8");
const nextConfig = readFileSync(new URL("../next.config.mjs", import.meta.url), "utf8");
const api = readFileSync(new URL("../lib/api.ts", import.meta.url), "utf8");
const compose = readFileSync(new URL("../../docker-compose.yml", import.meta.url), "utf8");

assert.doesNotMatch(campaign, /dangerouslySetInnerHTML/, "qurl provider markup must never be injected into the admin DOM");
assert.match(campaign, /qr_asset_url/, "campaign QR preview must consume the isolated image asset");
assert.doesNotMatch(proxy, /request\.arrayBuffer\(\)/, "the API gateway must not buffer unbounded request bodies");
assert.match(proxy, /MAX_PROXY_REQUEST_BODY_BYTES/, "the API gateway must enforce a request body bound");
assert.match(nextConfig, /object-src 'none'/, "CSP must prohibit active object content");
assert.match(nextConfig, /frame-ancestors 'none'/, "CSP must prevent framing of authenticated surfaces");
assert.match(api, /!\["local", "docker", "test"\]\.includes/, "unknown frontend environments must fail closed to Passage authentication");
assert.match(compose, /HEARD_BIND_ADDRESS:-127\.0\.0\.1/, "local Compose services must bind to loopback by default");
assert.match(compose, /HEARD_API_PORT:-8082\}:8080/, "Heard must publish its API on host port 8082 while retaining container port 8080");
assert.match(proxy, /API_INTERNAL_BASE_URL \?\? "http:\/\/localhost:8082"/, "local web development must proxy to Heard's host API port");

console.log("security boundary checks passed");
