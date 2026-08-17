# Slice 5: Prospect conversion journey

## Goal

Turn interest from the marketing site and guest demo into a clear, durable request for a heard product conversation without confusing that journey with customer sign-in or promising a free trial that does not exist yet.

## User story

As a restaurant operator evaluating heard, I can understand the outcomes heard provides, request a tailored walkthrough with a short restaurant-specific form, and know what happens next.

As an existing heard customer, I can reach a dedicated sign-in page without seeing development, provider, seed-data, or sandbox terminology.

## Scope

- Separate prospect conversion and existing-customer sign-in journeys.
- Prospect-first homepage navigation with primary restaurant and guest-experience calls to action; customer sign-in remains visually tertiary.
- Plain-language homepage feedback states that connect a newly received low rating directly to the Recovery inbox without implying a separate live-feed product.
- One shared public header across the homepage and contact journey so conversion actions never disappear between steps.
- A consistent light editorial canvas using parchment, a subtle grid, clay and olive atmosphere, ink accent panels, and the same card and button treatment used by the management console.
- Headline-first hierarchy: the distinctive product line is the primary heading and the explanatory promise is its subtitle, not a competing second headline.
- A deliberately concise contact page with one operator story and one next-step explanation beside the lead form; no stacked feature lists or redundant sales sections.
- Public `/contact` page with product outcomes, expectations, and a short lead form.
- Public `POST /api/v1/marketing-leads` endpoint.
- Durable `marketing_leads` storage for sales follow-up.
- Marketing and guest-demo CTAs route to `/contact` with source attribution.
- Public-facing copy contains no local-development or Passage implementation details.

## Non-goals

- Self-serve free-tier registration or a free-trial promise.
- CRM delivery, lead assignment, automated email, or calendar scheduling.
- Passage account creation or production authentication.
- Invented customer logos, performance statistics, or testimonials.

## API contract

The canonical contract is `docs/openapi.slice1.yaml`, version `0.4.0`.

`POST /api/v1/marketing-leads` is public because prospects are not authenticated. It accepts name, work email, optional phone, restaurant name, location count, optional challenge context, source attribution, and explicit contact consent.

## Data model

`marketing_leads` stores a UUID, contact fields, restaurant context, source, consent, status, and creation timestamp. It is intentionally not tenant-scoped because a prospect does not have a heard tenant before qualification and account creation.

## Events emitted

No event is emitted in this slice. The durable lead row is the initial sales queue. CRM delivery will add an explicit provider contract and transactional delivery mechanism when a CRM is selected.

## Permissions

- Anyone may create a contact request.
- No public or restaurant-tenant API lists marketing leads.
- Sales access and lead administration are deferred until a Haiven-level internal-operator authorization model exists.

## Failure modes

- Missing or malformed required fields return a structured `400` error.
- Invalid email or optional phone formats return a structured `400` error.
- Failed persistence keeps the prospect's entered values and displays a visible error.
- Unknown source values and location ranges are rejected.

## Observability

Successful creation logs only lead ID, source, and location range. Contact details are not logged.

## Tests required

- Validation accepts a complete restaurant lead with optional phone omitted.
- Validation rejects missing consent, malformed contact details, unsupported location ranges, and unknown sources.
- API persistence is verified against PostgreSQL.
- Browser journey covers guest-demo CTA, form validation, successful confirmation, and separate customer sign-in.
- Homepage checks prevent operational jargon from replacing the new-feedback and Recovery inbox language, and verify both prospect calls to action remain in the primary navigation.
- Brand-language checks require the shared header and backdrop, enforce headline hierarchy, and prevent the contact page from returning to a disconnected full-dark motif or numbered feature list.
- Production frontend build and type validation pass.

## Definition of done

- Prospects never land on sign-in from a marketing CTA.
- Existing customers retain a direct sign-in path.
- Home, contact, and admin read as one product through shared colors, atmosphere, typography, card shapes, and action hierarchy.
- The contact journey retains the primary public navigation and presents only the context needed to complete the form.
- The form explains heard's value and what happens after submission.
- Contact requests persist with source attribution and consent.
- No public UI exposes Passage, sandbox, seeded workspace, or local-build language.
