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
