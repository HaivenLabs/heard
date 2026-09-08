import assert from "node:assert/strict";
import { createHash } from "node:crypto";
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
assert.equal(
  createHash("sha256").update(readFileSync(logoPath)).digest("hex"),
  "a2db25103790a80ca05d685a9f461f3c1758e5441582ace1da731a8c52b01ad0",
  "The Heard logo asset must remain an exact copy of the supplied source file."
);
function logoFill(className) {
  const match = logoSource.match(new RegExp(`\\.${className}\\s*\\{fill:([#][0-9A-F]{6})`, "i"));
  assert.ok(match, `The supplied logo must define a ${className} fill.`);
  return match[1].toLowerCase();
}
const logoTint = logoFill("fil1");
const logoPrimary = logoFill("fil3");
const logoStrong = logoFill("fil2");
assert.equal(heardTheme.accent, logoPrimary);
assert.equal(heardTheme.accentStrong, logoStrong);
assert.equal(heardTheme.focus, logoPrimary);
assert.equal(heardTheme.tint, logoTint);
assert.match(logoComponent, /heard-logo\.svg/);
assert.match(publicHeader, /HeardLogo/);
assert.match(adminShell, /HeardLogo/);
assert.match(brandDoc, /exact greens in its supplied vector logo/i);
assert.match(brandDoc, /business-facing operational energy/i);

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

console.log("Exact Heard logo asset and green brand theme token snapshot verified.");
