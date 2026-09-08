import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import { resolve } from "node:path";

const tokenJsonPath = resolve("packages/design-tokens/src/tokens.json");
const tokenCssPath = resolve("packages/design-tokens/src/tokens.css");
const centralTokenJsonPath = resolve("../../haiven/packages/design-tokens/src/tokens.json");
const centralTokenCssPath = resolve("../../haiven/packages/design-tokens/src/tokens.css");
const tokens = JSON.parse(readFileSync(tokenJsonPath, "utf8"));
const heardTheme = tokens.product.heard;
const layout = readFileSync(resolve("app/layout.tsx"), "utf8");
const tailwindConfig = readFileSync(resolve("tailwind.config.ts"), "utf8");
const publicHeader = readFileSync(resolve("components/public-header.tsx"), "utf8");
const adminShell = readFileSync(resolve("components/admin-shell.tsx"), "utf8");
const logoComponent = readFileSync(resolve("components/heard-logo.tsx"), "utf8");
const brandDoc = readFileSync(resolve("../docs/brand.md"), "utf8");
const logoPath = resolve("public/brand/heard-logo.svg");

assert.match(layout, /packages\/design-tokens\/src\/tokens\.css/);
assert.match(layout, /data-haiven-product="heard"/);
assert.match(tailwindConfig, /--hv-action-primary-rgb/);
assert.match(tailwindConfig, /--hv-action-primary-hover-rgb/);
assert.match(tailwindConfig, /--hv-color-moss-rgb/);
assert.ok(existsSync(logoPath), "The Heard logo asset must be included with the application.");

const logoSource = readFileSync(logoPath, "utf8");
assert.match(logoSource, /<svg\b/i);
assert.doesNotMatch(logoSource, /<script\b/i);
assert.match(logoSource, new RegExp(heardTheme.accentStrong, "i"));
assert.match(logoSource, new RegExp(heardTheme.accent, "i"));
assert.match(logoSource, new RegExp(heardTheme.tint, "i"));
assert.match(logoComponent, /heard-logo\.svg/);
assert.match(publicHeader, /HeardLogo/);
assert.match(adminShell, /HeardLogo/);
assert.match(brandDoc, /three greens in the logo/i);
assert.match(brandDoc, /business-facing products/i);
assert.equal(heardTheme.accent, tokens.product.haiven.accent);
assert.equal(heardTheme.accentStrong, tokens.product.passage.accent);
assert.equal(heardTheme.tint, tokens.product.qurl.accent);

if (existsSync(centralTokenJsonPath) && existsSync(centralTokenCssPath)) {
  assert.equal(
    readFileSync(tokenJsonPath, "utf8"),
    readFileSync(centralTokenJsonPath, "utf8"),
    "The local token snapshot must match the central Haiven token release."
  );
  assert.equal(
    readFileSync(tokenCssPath, "utf8"),
    readFileSync(centralTokenCssPath, "utf8"),
    "The local CSS token snapshot must match the central Haiven token release."
  );
}

console.log("Heard logo asset and green brand theme token snapshot verified.");
