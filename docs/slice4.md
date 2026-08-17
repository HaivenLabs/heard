# Slice 4: Branded guest experience refinement

## Goal

Make the guest survey feel like a polished, branded restaurant experience while improving completion, contact quality, and demo lead capture.

## User story

As a guest, I can recognize the restaurant brand, choose a modern visual rating, revise that choice, provide feedback and contact details in one final step, and understand exactly what happens after submission.

As a Heard prospect using the seeded demo, the information I enter is persisted with a demo lead source and an explicit Heard follow-up preference.

## Scope

- Supplied nom logo and nom-derived teal, cream, coral, and ink colors.
- Lowercase `heard` and `nom` brand presentation.
- Product-owned SVG rating faces with consistent rendering across browsers and devices.
- A visible `Change my rating` action before submission.
- One final screen for feedback, optional public review links, and contact capture.
- Completion language based on the visible selected feeling rather than an unseen number.
- Client and server email/phone format validation.
- Shared format rules accept practical tagged email addresses, formatted US/Canada numbers, and `+`-prefixed international numbers while rejecting malformed domains, impossible NANP prefixes, misplaced country prefixes, letters, and unbalanced punctuation.
- Demo response metadata for lead source and requested Heard sales follow-up.

## Non-goals

- Verifying email or phone ownership with an external service.
- A CRM, lead assignment queue, or automated sales outreach.
- General restaurant theme configuration beyond the current nom-branded slice.
- Replacing qurl or Passage responsibilities.

## API contract

The existing feedback session and response endpoints remain canonical in `docs/openapi.slice1.yaml`.

Flyer giveaway feedback requires at least one valid contact method. Submitted name, phone, email, rating, comment, categories, review clicks, consent flags, and metadata are persisted in the feedback response transaction.

Demo submissions add:

- `entry_surface: heard_guest_demo`
- `lead_source: heard_guest_demo`
- `heard_sales_follow_up_requested: boolean`

## Data model

No migration is required. Contact fields already belong to `feedback_sessions` and `feedback_responses`; lead attribution and follow-up preference are intentionally scoped to response metadata until a real CRM integration exists.

## Events emitted

The existing `feedback-submitted` event includes the persisted response identity and continues through the transactional outbox. Contact details are not copied into the event payload.

## Permissions

The guest submission remains public through an opaque campaign token. Reading persisted responses remains tenant-protected.

## Failure modes

- Missing phone and email blocks giveaway submission.
- Malformed email or phone blocks submission in the browser and API.
- Failed persistence keeps the guest on the final step with entered data intact.
- Changing the rating preserves comment, category, and contact state.
- Missing brand data falls back to the restaurant name and a branded initial.

## Observability

Existing structured API errors and outbox processing logs cover submission failures. Demo lead intent is queryable through response metadata without logging contact details.

## Tests required

- Server accepts valid email or formatted phone.
- Server rejects missing giveaway contact.
- Server rejects malformed email and phone.
- Production frontend build and type validation.
- Browser checks for all rating faces at desktop and mobile widths.
- Browser checks for change-rating state preservation, inline validation, successful persistence, and lowercase branding.

## Definition of done

- nom logo and colors appear in the seeded guest experience.
- Rating faces are crisp, custom, and unclipped on every guest step.
- Guests can change ratings before submission.
- Contact capture and optional review links share the final step.
- Demo contact information and lead metadata persist.
- Contact format validation is enforced by both client and server.
- Guest-facing Heard branding is lowercase.
