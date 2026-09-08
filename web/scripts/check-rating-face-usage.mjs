import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const componentSource = readFileSync(resolve("components/rating-face.tsx"), "utf8");
for (const faceSet of ["heard", "clay", "glass", "minimal", "retro"]) {
  assert.match(componentSource, new RegExp(`(?:id|value): ["']${faceSet}["']`));
}
assert.match(componentSource, /export type RatingFaceSet/);
assert.equal(
  componentSource.match(/<SoftSadExpression\b/g)?.length,
  5,
  "Every face set should use the shared, softened one-star expression"
);

const expectedConsumers = [
  "app/page.tsx",
  "app/admin/campaigns/page.tsx",
  "components/flyer-survey.tsx"
];

for (const relativePath of expectedConsumers) {
  const source = readFileSync(resolve(relativePath), "utf8");
  assert.match(source, /import \{[^}]*RatingFace[^}]*\} from ["'][^"']*rating-face["'];/);
  assert.match(source, /<RatingFace\b/);
  assert.doesNotMatch(source, /<svg[^>]*viewBox=["']0 0 100 106["']/);
}

const campaignSource = readFileSync(resolve("app/admin/campaigns/page.tsx"), "utf8");
assert.match(campaignSource, /name=["']rating_face_set["']/);
assert.match(campaignSource, /rating_face_set: formData\.rating_face_set/);
assert.match(campaignSource, /faceSet=\{formData\.rating_face_set\}/);
assert.match(campaignSource, /size-5 shrink-0 items-center justify-center rounded-full border-2/);
assert.match(campaignSource, /size-2\.5 rounded-full bg-teal/);
assert.doesNotMatch(campaignSource, /\{selected \? "✓" : ""\}/);

const surveySource = readFileSync(resolve("components/flyer-survey.tsx"), "utf8");
assert.match(surveySource, /faceSet=\{survey\?\.campaign\?\.rating_face_set \?\? ["']heard["']\}/);

const tokenCss = readFileSync(resolve("packages/design-tokens/src/tokens.css"), "utf8");
const artworkTokens = [...new Set([...componentSource.matchAll(/art\("([a-z0-9]+)"\)/g)].map((match) => match[1]))];
assert.ok(artworkTokens.length > 0, "Rating artwork must use shared artwork tokens.");
for (const token of artworkTokens) {
  assert.match(tokenCss, new RegExp(`--hv-art-${token}:`), `Missing shared artwork token ${token}`);
}

console.log("Shared RatingFace usage verified across guest, campaign, and marketing surfaces.");
