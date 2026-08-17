# Slice 3: Authenticated customer activation

## Goal

Give an existing restaurant customer a credible path into a protected management console and then into the first campaign workflow, without creating a permanent Heard-owned authentication stack. This slice proves returning-customer access; first-time self-service registration and workspace activation are defined in Slice 6.

## User story

As an existing restaurant customer, I can sign in, see my workspace, and start building a campaign. In local development, the Passage boundary is represented by a development-only adapter so the manager workflow can be evaluated before the shared Passage service is deployed.

## Scope

- Existing-customer sign-in entry from the landing page.
- Local-only Passage adapter that issues signed, expiring development sessions.
- Bearer authentication for manager APIs.
- Tenant-membership and permission enforcement derived from verified identity claims.
- Authenticated management console with real location, campaign, and recovery counts.
- Persistent shared management shell, navigation, and background across every authenticated admin route.
- The management shell uses the same parchment, grid, clay, olive, ink, card, and button language as heard's public surfaces while preserving admin-specific information density.
- Same-origin `/api/*` gateway from the web app to the internal Go API.
- Direct console path to the flyer campaign builder and recovery inbox.
- Production startup guard that rejects the local identity adapter.

## Non-goals

- Production registration, login, password, passkey, OAuth, or account recovery.
- Heard-owned tenant membership or role administration.
- Billing and subscription enforcement.
- A production Passage token verifier before Passage publishes that contract.
- First-time restaurant registration and workspace provisioning, which are delivered as a complete self-service journey in Slice 6.

## API contract

The canonical contract is `docs/openapi.slice1.yaml`, version `0.5.0`.

- `POST /api/v1/auth/local/session` issues a development-only session when `PASSAGE_MODE=local`.
- `GET /api/v1/session` resolves the authenticated identity.
- Manager endpoints require `Authorization: Bearer <token>`.
- Browser clients call the same Heard origin; Next.js forwards `/api/*` to the internal Go API.
- Tenant-scoped manager endpoints also require `X-Heard-Tenant-ID` matching a tenant in the verified identity.
- Actor ID and role are derived from verified claims. Client-provided actor headers are not trusted.

## Identity contract

Heard consumes this provider-neutral identity shape:

- user ID
- email and display name
- tenant memberships
- active role
- permission claims
- identity provider name

The local adapter is a fake provider behind the same interface intended for Passage. It maps every valid local email to the seeded demo tenant and grants the owner role. It cannot be enabled when `APP_ENV=production`.

Slice 6A adds a separate local registration operation that deterministically resolves a dedicated account context for each email. Existing-customer local sign-in continues to use the seeded tenant so the demo and returning-customer workflow remain stable.

## Data model

No persistence migration is required. Passage owns identities, sessions, tenant memberships, roles, and permissions. Heard continues to persist only product-domain tenants and actor IDs on audit records.

## Events emitted

Authentication does not emit a Heard domain event in this slice. Existing campaign and feedback workflows retain their current transactional outbox behavior.

## Permissions

- `tenant:read`, `tenant:create`
- `location:read`, `location:write`
- `campaign:read`, `campaign:write`
- `recovery:read`, `recovery:write`

Each manager route declares the permission it requires. Tenant membership is checked independently from permission claims.

## Failure modes

- Missing Bearer token returns `401`.
- Invalid, tampered, or expired token returns `401`.
- Missing tenant context returns `401`.
- Cross-tenant access returns `403`.
- Missing permission returns `403`.
- Local Passage mode in production prevents API startup.
- Invalid browser sessions are cleared and returned to `/login`.

## Observability

The API logs local session issuance, denied permissions, and denied tenant access using actor and tenant identifiers. It does not log access tokens or email addresses.

## Tests required

- Local session issue and verification.
- Tampered token rejection.
- Authentication-required behavior.
- Tenant isolation.
- Permission denial.
- Production local-adapter guard.
- Production frontend compilation and type checking.
- Browser workflow covering login, console, and campaign entry.

## Passage migration

Replace the local adapter through the `IdentityProvider` interface after Passage publishes token verification and login contracts. Production browser session storage and redirects should then use the Passage SDK or secure server-managed cookie flow. The local session endpoint remains disabled outside local development.

## Definition of done

- Protected manager APIs no longer trust actor headers.
- A local user can sign in, load the console, and enter campaign creation.
- Tenant and permission failures are enforced and tested.
- The local adapter cannot run in production.
- API contract and local-run documentation match implementation.
- Go tests, frontend build, runtime smoke test, and Haiven checks pass.
