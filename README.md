# Heard

Open-source guest feedback, recovery, and restaurant intelligence.

Heard helps restaurants capture customer feedback, recover unhappy guests, and turn guest experiences into operational insight. It is designed to be self-hostable, API-first, contract-first, test-first, observable, secure, and extensible.

## Why Heard

Restaurants need a better way to know what guests experienced before those issues become churn, bad reviews, or invisible operational drag.

Heard is built around a simple loop:

1. Capture feedback quickly.
2. Understand the guest experience.
3. Route issues to the right person.
4. Recover the guest.
5. Learn from the signal.
6. Improve operations.

## Core Product Areas

- Frictionless feedback
- QR and link-based feedback collection
- Guest recovery workflows
- Recovery inbox and case management
- SMS and email provider abstractions
- Response templates
- Basic offers and make-it-right tracking
- Experience identity and attribution
- Audit logs
- Webhooks
- Open APIs
- Local development with fake providers

Future product areas include:

- Review management
- Listings management
- AI operational insights
- Advanced automations
- POS and ordering integrations
- Call-to-text
- Multi-location reporting
- Enterprise governance

## Architecture Principles

Heard is designed as a modular Go monolith first.

That means:

- Simple local development
- Clear domain boundaries
- Strong contracts
- No premature microservices
- Clean extraction paths later if scale requires it

Core principles:

- API-first
- Contract-first
- Test-first
- Behavior-driven for critical workflows
- Event-driven internally
- Transactional outbox for reliable event publishing
- Multi-tenant by default
- UUID-based internal identity
- Provider abstractions
- OpenTelemetry-based observability
- Secure and privacy-aware by default
- Self-hostable without paid vendors

## Preferred Stack

Backend:

- Go
- PostgreSQL
- Internal domain events
- Transactional outbox
- Redis only when needed
- NATS adapter later if deployment scale justifies it
- OpenTelemetry
- Docker-first local development

Frontend:

- React
- Next.js
- TypeScript
- Tailwind

APIs:

- REST externally
- OpenAPI contracts
- Generated server/client code where practical

Testing:

- Unit tests
- Integration tests
- Contract tests
- BDD-style tests for critical workflows
- Multi-tenant isolation tests
- Permission tests

## Identity Model

Heard owns its internal IDs.

Internal entity IDs use UUIDs. External provider IDs are stored as external references, not as Heard primary keys.

The minimal Experience identity model is:

```json
{
  "id": "heard-owned-uuid",
  "tenant_id": "uuid",
  "location_id": "uuid",
  "source_system": "toast",
  "source_entity_id": "provider-owned-id"
}
```

The identity tuple is:

```text
tenant_id + location_id + source_system + source_entity_id
```

Provider-specific data such as item details, server, table, subtotal, ticket number, or fulfillment type is optional enrichment. It must not be required for core feedback, recovery, or reporting flows.

## License

Heard is licensed under the GNU Affero General Public License v3.0. See [LICENSE](./LICENSE).

## Trademarks

The Heard name, logo, domain names, and related brand assets are not licensed under the AGPL-3.0 license.

You may fork, modify, and self-host the software under the terms of the AGPL-3.0 license, but you may not use the Heard name, logo, or branding to represent your fork or hosted service as the official Heard project or service without written permission.

See [TRADEMARKS.md](./TRADEMARKS.md) for details.

## Contributing

Contributions are welcome.

Before contributing, read:

- [CONTRIBUTING.md](./CONTRIBUTING.md)
- [SECURITY.md](./SECURITY.md)
- [TRADEMARKS.md](./TRADEMARKS.md)

## Project Status

Heard is early-stage software. APIs, schemas, and architecture may change before the first stable release.

## Local Run

Heard is implemented as a Docker-first vertical slice with:

- `db`: PostgreSQL
- `api`: Go API server
- `worker`: Go outbox worker
- `web`: Next.js guest/admin app

Start it locally:

```bash
docker compose up --build
```

Compose binds PostgreSQL, the API, and the web app to `127.0.0.1` by default so development owner-token routes are not exposed to the LAN. An intentional host publish may set `HEARD_BIND_ADDRESS`, but a non-loopback value is accepted only when `PASSAGE_MODE=jwks` and `HEARD_SEED_DEMO=false`; place TLS and a forwarding-header-sanitizing proxy in front of the deployment. Local session and registration routes are registered only when the explicit local Passage adapter is active.

Then open:

- `http://localhost:3010` for the web app
- `http://localhost:3010/start` to create a local self-service account context
- `http://localhost:3010/onboarding` to resume restaurant, location, and first-campaign setup
- `http://localhost:3010/walkthrough` for the optional walkthrough journey
- `http://localhost:3010/api/v1/healthz` for the same-origin API health endpoint
- `http://localhost:8082/api/v1/healthz` for direct API development and debugging
- `http://localhost:3010/login` for existing-customer sign-in
- `http://localhost:3010/admin` for the authenticated management console
- `http://localhost:3010/admin/campaigns` to create a printed flyer survey campaign
- `http://localhost:3010/f/demo-heard` for the seeded flyer survey flow
- `http://localhost:3010/admin/recovery` for the manager recovery inbox

Default seeded tenant header:

```text
11111111-1111-1111-1111-111111111111
```

`/start` is the primary self-service journey. Restaurant and first-location names are collected once, then the operator deliberately chooses Google or email-and-password identity. Email signup requires a password and email verification; email login requires the existing password. Passage returns the verified product identity and Heard activates only that identity's restaurant workspace. Retries are idempotent, and a qurl outage does not prevent the feedback link from working. Docker uses the real Passage JWKS boundary by default; the passwordless local adapter is reserved for explicit test scenarios and is never a customer-facing login path.

`/walkthrough` remains an optional request for a tailored walkthrough and never gates product access. `APP_ENV` is required and accepts only `local`, `docker`, `test`, `staging`, or `production`; unknown and misspelled values prevent startup. The test-only local Passage adapter and demo seed are allowed only in explicit local runtimes, and additionally require loopback public/web/CORS origins. Docker publishes database, API, and web ports on loopback only and uses Passage JWKS by default. Staging and production require secure browser cookies, HTTPS Passage endpoints, and `HEARD_SEED_DEMO=false`.

Production browser entry starts at the provider-neutral `GET /api/v1/auth/start`. Heard exposes native Sign in with Google and Sign up with Google controls while Passage owns the Google OIDC exchange behind that Heard contract. Registration carries a short-lived Heard-owned onboarding draft so Passage can return the verified identity directly to Heard and Heard can atomically activate the named restaurant and first location without exposing Passage UI. Heard stores PKCE verifier, state, return path, and registration draft in short-lived HttpOnly cookies; Passage returns a single-use authorization code; Heard exchanges it server-side using `PASSAGE_CLIENT_ID` and the exact registered `PASSAGE_CALLBACK_URL`. The product JWT is verified and kept only in the Secure/HttpOnly `heard_session` cookie. It is never placed in a browser URL or local storage.

`PASSAGE_PUBLIC_URL` is the browser-visible identity origin used only for authorization redirects (local combined-stack default `http://localhost:3020`). In production it must be a Heard-owned custom auth domain with Heard-branded Google credentials and consent screen; do not expose a Passage hostname or Passage branding. `PASSAGE_BASE_URL` remains the backend-reachable origin used for token exchange and JWKS. If `PASSAGE_PUBLIC_URL` is unset it falls back to `PASSAGE_BASE_URL` for compatibility. Never configure a Docker-only hostname such as `host.docker.internal` as the public URL because the browser must send the identity origin's session cookie.

Heard owns its Google OAuth client ID and secret. When those values are set,
the Heard API configures the `heard/google` connection once at startup through
Passage's service-authenticated provider-configuration API. Passage encrypts
the secret before persistence and performs Google's code exchange; normal
sign-in requests never contain the client secret. Both backends must receive
the same 32+-character `PASSAGE_PROVIDER_CONFIGURATION_TOKEN`.

Browser traffic uses the Heard origin at `http://localhost:3010/api/*`. Next.js forwards those requests over the Docker network to the Go API at `api:8080`. The container-only port stays `8080`, while Heard publishes host port `8082` by default for direct API development and debugging; override the host port with `HEARD_API_PORT` when needed.

Public marketing-lead, feedback-session, and feedback-response writes accept only bounded `application/json` requests. Text, category lists, and arbitrary metadata have contract limits. Rate limits use bounded in-memory buckets only in explicit local runtimes; staging and production use atomic PostgreSQL buckets shared across API replicas and fail closed if that protection is unavailable. `X-Forwarded-For` is ignored unless the immediate peer belongs to `TRUSTED_PROXY_CIDRS`; an approved edge proxy must overwrite, not append to, untrusted client-supplied forwarding headers.

qurl may return an HTTPS asset URL or inline SVG. Heard converts inline SVG to an isolated `data:image/svg+xml` image source and never exposes provider markup for DOM insertion. Public HTTP assets are rejected; loopback HTTP assets are allowed only in explicit local runtimes. The web app also sends a restrictive CSP that disables object embedding and framing.

The worker drains up to `WORKER_BATCH_SIZE` available events per poll and logs attempted, successful, and failed counts. A tenant owner with `outbox:replay` may safely requeue a failed event through `POST /api/v1/outbox-events/{id}/requeue`; the operation is tenant-scoped, resets retry state, and writes an audit record. Other event states cannot be replayed through this endpoint.

Run checks directly with `go test ./...` from `backend`, and `pnpm --filter @heard/web test`, `pnpm --filter @heard/web lint`, and `pnpm --filter @heard/web build` from the repository root. Set `HEARD_TEST_DATABASE_URL` to run the PostgreSQL onboarding isolation and concurrency coverage; CI must provide it so tenant-boundary tests cannot be silently skipped.

Reference docs:

- [Slice 1 implementation](./docs/slice1.md)
- [Slice 2 implementation](./docs/slice2.md)
- [Slice 3 authenticated activation](./docs/slice3.md)
- [Slice 4 branded guest experience](./docs/slice4.md)
- [Slice 5 optional walkthrough](./docs/slice5.md)
- [Slice 6 self-service onboarding](./docs/slice6.md)
- [Slice 1 OpenAPI contract](./docs/openapi.slice1.yaml)
