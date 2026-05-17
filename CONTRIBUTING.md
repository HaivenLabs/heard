# Contributing to Heard

Thanks for contributing to Heard.

Heard is an open-source guest feedback, recovery, and restaurant intelligence platform. Contributions should improve the product while preserving the architecture principles that make the system reliable, maintainable, and self-hostable.

## Engineering Standards

Every meaningful contribution should follow these principles:

- API-first
- Contract-first
- Test-first
- Tenant-aware
- Observable
- Secure by default
- Provider-abstracted when third-party services are involved
- Simple enough to run locally
- Clear enough for humans and AI agents to understand

## Architecture

Heard starts as a modular Go monolith.

Do not introduce a new service, queue, database, cache, cloud dependency, or framework unless the need is clear and documented.

Preferred initial process model:

- API server
- Background worker
- Web/admin app
- PostgreSQL
- Optional fake local providers

## Contracts First

For API changes:

1. Update the OpenAPI contract first.
2. Regenerate code where applicable.
3. Implement the change.
4. Add or update contract tests.
5. Update docs and examples.

Do not hand-write API behavior that drifts from the contract.

## Database Changes

All schema changes must use versioned migrations.

Migration requirements:

- Deterministic
- Tested from a clean database
- Safe for self-hosted installs
- Rollback behavior documented when practical
- Destructive changes explicitly called out

## Tenant Isolation

Tenant isolation is mandatory.

Any tenant-scoped data access must include tenant context and authorization checks.

Contributions that add tenant-scoped tables, queries, APIs, exports, jobs, or webhooks must include tenant isolation tests.

## Testing Expectations

Add tests with your change.

Expected test types:

- Unit tests for deterministic domain logic
- Integration tests for APIs and persistence
- Contract tests for providers and APIs
- Permission tests for protected actions
- Tenant isolation tests for tenant-scoped behavior
- BDD-style tests for critical workflows

Critical workflows include:

- Feedback submission
- Recovery case creation
- Messaging
- Offer issuance
- Webhook delivery
- Permission enforcement
- Event publication

## Observability

Critical flows must include:

- Structured logs
- Correlation IDs where applicable
- Metrics where useful
- Tracing for cross-module workflows
- Audit entries for sensitive actions

Do not log sensitive guest data.

## Provider Abstractions

Third-party dependencies must sit behind provider interfaces.

Examples:

- SMS
- Email
- AI
- Storage
- Translation
- POS integrations
- Review integrations

Local fake providers should exist for development and testing.

## Pull Request Checklist

Before opening a pull request, confirm:

- The change has a clear purpose.
- Contracts are updated if APIs changed.
- Migrations are included if the database changed.
- Tests are included.
- Tenant isolation is preserved.
- Permissions are enforced.
- Observability is included where relevant.
- Docs are updated.
- Local development still works.
- No secrets or sensitive data are committed.

## Commit Style

Use clear, direct commit messages.

Examples:

```text
Add feedback session creation API
Add recovery case state machine
Fix tenant isolation in feedback queries
Document QR tokenization model
```

## Code Style

Write code that is boring, readable, and maintainable.

Prefer:

- Small focused functions
- Clear names
- Explicit dependencies
- Simple control flow
- Comments that explain why something exists

Avoid:

- Clever abstractions
- Hidden global state
- Framework magic
- Premature microservices
- Unbounded queries
- Vendor lock-in in core modules

## License

By contributing to Heard, you agree that your contributions are provided under the repository license.
