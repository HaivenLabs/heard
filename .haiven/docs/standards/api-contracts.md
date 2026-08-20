# API and Contract Standards

Haiven products are API-first and contract-first.

## OpenAPI

Public APIs should use OpenAPI unless there is a documented exception.

Requirements:

- OpenAPI is the source of truth.
- Versioned APIs are required.
- Generated types/clients/stubs should be used where practical.
- API examples should be documented.
- Structured machine-readable errors are required.
- Pagination is required for list endpoints.
- Idempotency is required where retries or duplicate creation are plausible.
- Auth and permission behavior must be described.
- Failure modes must be documented.

## JSON Schema

Major entities and event payloads should use JSON Schema.

Requirements:

- Schemas are versioned.
- Breaking changes require new schema versions.
- Shared schemas live in a dedicated contracts package or clearly named contract location.
- Validation tests are required.
- Arbitrary metadata must be intentionally scoped.

## Error shape

Errors should be structured and machine-readable.

Recommended fields:

- `code`
- `message`
- `details`
- `correlation_id`
- `field_errors` where applicable

Do not return random string errors from production APIs.

## Versioning

Default API path:

```text
/api/v1
```

Breaking changes require:

- new version
- migration notes
- compatibility assessment
- updated client generation

## Definition of done

API work is complete only when:

- contract is updated
- generated code is refreshed where practical
- implementation matches contract
- contract tests pass
- permission behavior is covered
- failure behavior is documented
- examples are updated
