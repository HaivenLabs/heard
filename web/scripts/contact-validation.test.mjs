import assert from "node:assert/strict";
import test from "node:test";
import { isValidEmail, isValidPhone } from "../lib/contact-validation.ts";

test("accepts practical email formats", () => {
  for (const email of [
    "guest@example.com",
    "first.last+takeout@restaurant.co.uk",
    "OWNER@NOM-KITCHEN.COM"
  ]) {
    assert.equal(isValidEmail(email), true, email);
  }
});

test("rejects malformed email formats", () => {
  for (const email of [
    "guest@example",
    "guest@example.c",
    ".guest@example.com",
    "guest..name@example.com",
    "guest@example..com",
    "guest@-example.com",
    "guest@example_.com",
    "guest @example.com",
    "guest@@example.com"
  ]) {
    assert.equal(isValidEmail(email), false, email);
  }
});

test("accepts formatted NANP and international phone numbers", () => {
  for (const phone of [
    "(206) 555-0142",
    "1-206-555-0142",
    "+1 (206) 555-0142",
    "+44 20 7946 0958"
  ]) {
    assert.equal(isValidPhone(phone), true, phone);
  }
});

test("rejects malformed phone numbers", () => {
  for (const phone of [
    "000-000-0000",
    "123-456-7890",
    "206-155-0142",
    "+01234567890",
    "+1 206 555 014",
    "206-555-014",
    "206-555-CALL",
    "206+555+0142",
    "(206 555-0142",
    "-206-555-0142",
    "206-555-0142-"
  ]) {
    assert.equal(isValidPhone(phone), false, phone);
  }
});
