import assert from "node:assert/strict";
import test from "node:test";
import {
  LOCAL_CANONICAL_APP_ORIGIN,
  browserFacingRequestURL,
  canonicalOrigin,
  canonicalRedirectURL,
  configuredCanonicalAppOrigin
} from "../lib/canonical-app-origin.mjs";

test("normalizes a configured public application origin", () => {
  assert.equal(canonicalOrigin("https://heard.staging.example/"), "https://heard.staging.example");
  assert.equal(canonicalOrigin("http://localhost:3010"), "http://localhost:3010");
});

test("rejects a non-origin application URL", () => {
  for (const value of ["heard.example", "https://heard.example/admin", "https://heard.example?preview=true"]) {
    assert.throws(() => canonicalOrigin(value), /PUBLIC_APP_URL/);
  }
});

test("requires explicit public configuration outside local runtimes", () => {
  assert.equal(configuredCanonicalAppOrigin("docker"), LOCAL_CANONICAL_APP_ORIGIN);
  assert.throws(() => configuredCanonicalAppOrigin("staging"), /required outside local development/);
});

test("redirects an alias before OAuth cookies are issued while preserving the request", () => {
  const redirect = canonicalRedirectURL(
    "http://127.0.0.1:3010/api/v1/auth/start?provider=google&intent=login",
    LOCAL_CANONICAL_APP_ORIGIN
  );
  assert.equal(redirect?.toString(), "http://localhost:3010/api/v1/auth/start?provider=google&intent=login");
  assert.equal(canonicalRedirectURL("http://localhost:3010/login", LOCAL_CANONICAL_APP_ORIGIN), null);
});

test("uses the browser-facing host instead of the container bind address", () => {
  const browserRequest = browserFacingRequestURL(
    "http://0.0.0.0:3010/api/v1/auth/start?provider=google",
    "localhost:3010",
    "http"
  );
  assert.equal(browserRequest.toString(), "http://localhost:3010/api/v1/auth/start?provider=google");
  assert.equal(canonicalRedirectURL(browserRequest, LOCAL_CANONICAL_APP_ORIGIN), null);
});
