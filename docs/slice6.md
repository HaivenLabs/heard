# Slice 6: Self-service restaurant onboarding

## Implementation status

Slice 6 implements the complete local/self-hostable product workflow and the production Passage identity boundary: account context resolution, idempotent restaurant and first-location activation, resumable first-campaign creation, feedback-link handoff, qurl degradation, primary public calls to action, audit coverage, and the `restaurant-workspace-activated` outbox event.

In staging and production, Heard verifies Passage-issued EdDSA product tokens against Passage JWKS and requires the configured issuer, the `heard` audience and product grant, an unexpired token, UUID subject and organization claims, and a recognized organization role. Heard derives tenant context and product permissions only from those verified claims. The local registration endpoint is available only in explicit local runtimes with loopback origins.

## Goal

Let a restaurant operator start using heard immediately without speaking to sales. The primary public journey must create an account through Passage, establish the first restaurant workspace and location, and carry the operator directly into a guided first campaign. The walkthrough remains available as optional help.

## User story

As a restaurant operator new to heard, I can create my account, set up my restaurant and first location, and launch a feedback campaign with as little friction as possible so I can experience the complete product loop before deciding whether I need help.

## Scope

- Primary homepage and guest-demo calls to action start self-service onboarding.
- `/walkthrough` remains a secondary `Request a walkthrough` path for operators who want assistance.
- Passage-owned registration, authentication, session, and account recovery integration.
- Idempotent creation of the first heard tenant, owner membership, restaurant profile, and location after Passage resolves identity.
- A short onboarding sequence that asks only for information required to create the first campaign.
- Sensible defaults and progressive disclosure for optional restaurant, survey, and branding configuration.
- Direct handoff into campaign creation, followed by a clear next action to publish, copy the feedback link, or obtain the qurl-backed QR asset.
- Resumable onboarding when account, tenant, location, or campaign creation is interrupted.
- A useful empty state for authenticated accounts that have not completed workspace activation.
- Source attribution from the public entry point without requiring a marketing-lead form.

## Non-goals

- Requiring a sales qualification call, walkthrough, or approval before activation.
- Building authentication, account recovery, or shared tenant membership inside heard.
- Implementing a permanent heard-owned billing platform.
- POS integration, data import, enterprise hierarchy, or advanced brand configuration during initial onboarding.
- Collecting phone number, challenge narrative, location range, or other sales fields before the operator can use the product.

## API contract

The OpenAPI contract must be updated before implementation. The minimum heard-owned activation surface should:

- Resolve a verified Passage identity and its heard account context.
- Create or resume an idempotent restaurant-workspace activation.
- Create the first location with tenant and owner authorization enforced server-side.
- Return explicit onboarding state and the next incomplete step.
- Reuse the existing campaign APIs rather than creating a separate onboarding-only campaign model.

The production exchange follows Passage's published short-lived product-token and JWKS contract. The local fake continues to exercise onboarding through `POST /api/v1/auth/local/registration` and remains impossible to enable outside an explicit local runtime with loopback origins.

## Data model

- Heard persists restaurant tenant, location, onboarding state, and product attribution only.
- Passage owns identity, sessions, accounts, tenant membership, roles, and permissions.
- Activation requests require idempotency so retries cannot create duplicate tenants or locations.
- Billing and subscription state must use the future shared Haiven billing contract when commercial plan enforcement is introduced.

## Events emitted

Versioned events should describe durable product outcomes without copying identity-provider payloads:

- `restaurant-workspace-activated`
- Existing campaign-created or campaign-published events, when those contracts exist

Activation state and outbox records must be committed atomically where downstream behavior depends on the event.

## Permissions and tenant isolation

- Only a verified Passage identity may start or resume activation.
- Workspace creation grants or verifies the owner relationship through Passage; heard does not trust client-supplied role or tenant claims.
- Every location and campaign operation is scoped to the newly activated tenant.
- Retries and concurrent requests cannot attach a user to another restaurant or reveal another tenant's onboarding state.

## Failure modes

- Passage is unavailable: preserve safe progress, explain that account setup cannot currently continue, and allow retry.
- Workspace or location creation fails: return a structured error and resume from the incomplete step without duplication.
- qurl is unavailable: finish campaign creation, expose the feedback link, and clearly mark QR generation as temporarily unavailable.
- The operator leaves midway: resume at the first incomplete step after the next authenticated visit.
- An already activated operator follows a start CTA: send them to their workspace or first useful campaign action instead of restarting onboarding.
- Billing is not configured: do not block initial product use behind an invented local subscription check.

## Observability

- Measure onboarding started, account resolved, workspace activated, first location created, first campaign created, and first campaign published.
- Measure time to first campaign and drop-off by step.
- Log correlation, actor, tenant, and activation identifiers without logging credentials or unnecessary restaurant/contact data.
- Make activation failures and duplicate-prevention outcomes searchable.

## Tests required

- Contract tests for the Passage boundary and activation APIs.
- BDD browser journey from the primary homepage CTA through first campaign creation.
- Resume-after-interruption tests for every durable step.
- Idempotency and concurrent-activation tests.
- Permission and cross-tenant isolation tests.
- Returning-customer redirect test.
- Passage and qurl outage tests.
- Accessibility, responsive layout, loading, empty, and error-state checks.
- Checks that the walkthrough remains available but is not the primary onboarding CTA.

## Definition of done

- A new restaurant operator can start from the homepage and reach a usable first campaign without contacting sales.
- The onboarding path asks only for information needed for account, workspace, location, and first-campaign activation.
- Registration and identity use Passage; QR generation uses qurl.
- Activation is resumable, idempotent, tenant-safe, permission-safe, observable, and covered by workflow tests.
- Returning users bypass completed onboarding.
- The walkthrough is visibly optional.
- Product and local-run documentation match the implemented journey.
- `haiven check` passes.

Passage hosts registration, browser sessions, verification, recovery, organization membership, and product grants. Heard accepts only Passage's short-lived `aud=heard` product token at its API boundary; signing-key rotation is handled by bounded JWKS caching and forced refresh after signature/key mismatch.

The browser crosses domains through a PKCE-bound one-time authorization code. Passage validates an exact registered callback URI and returns only `code` plus opaque `state`; Heard validates state, exchanges the code server-side, verifies the product token, and stores it in an HttpOnly SameSite=Lax cookie. The code expires after two minutes and is atomically single-use. Neither product tokens nor PKCE verifiers appear in browser URLs or JavaScript storage.

Browser and service origins are configured separately: `PASSAGE_PUBLIC_URL` builds the browser authorize redirect and must be the same public origin that owns the Passage session cookie; `PASSAGE_BASE_URL` is used only for backend token exchange and JWKS. The public URL falls back to the base URL when omitted.
