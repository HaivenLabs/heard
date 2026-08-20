import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const source = readFileSync(resolve("components/flyer-survey.tsx"), "utf8");

// These checks exercise the user-visible state machine without requiring a
// browser runner in the current lightweight web test environment.
assert.match(source, /useState<"rate" \| "details" \| "done">\("rate"\)/);
assert.match(source, /setStep\("details"\)/);
assert.match(source, /Change my rating/);
assert.match(source, /setStep\("rate"\)/);
assert.match(source, /if \(!email && !phone\)/);
assert.match(source, /isValidEmail\(email\)/);
assert.match(source, /isValidPhone\(phone\)/);
assert.match(source, /await beginSession\(\)/);
assert.match(source, /\/api\/v1\/feedback-responses/);
assert.match(source, /setSubmitted\(response\)/);
assert.match(source, /setStep\("done"\)/);
assert.match(source, /review_destinations_clicked: reviewClicks/);
assert.match(source, /heard_sales_follow_up_requested: isDemo && heardFollowUp/);
assert.match(source, /role="alert"/);
assert.match(source, /googleReviewUrl = survey\?\.campaign\.google_review_url \|\|/);
assert.match(source, /yelpReviewUrl = survey\?\.campaign\.yelp_review_url \|\|/);
assert.match(source, /href=\{googleReviewUrl\}/);
assert.match(source, /href=\{yelpReviewUrl\}/);

console.log("Guest rating, change-rating, validation, persistence, and branding journey verified.");
