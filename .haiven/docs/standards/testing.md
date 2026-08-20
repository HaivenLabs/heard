# Testing Standards

Tests are a first-class deliverable in Haiven products.

## Required test types

Use the test types that match the change:

- unit tests for domain logic
- integration tests for APIs
- contract tests for providers and generated clients
- permission tests for protected workflows
- tenant isolation tests for tenant-scoped data
- event emission tests for event-driven behavior
- failure-mode tests for critical workflows
- regression tests for bug fixes
- end-to-end tests for critical user flows

## Test-first expectation

New behavior should be testable before implementation is considered complete.

AI agents must not treat tests as cleanup work.

## Critical workflow testing

Critical workflows should prove:

- happy path
- validation failures
- permission denial
- tenant isolation
- provider failure
- retry behavior where applicable
- event emission where applicable
- audit entries where applicable
- observability hooks where practical

## BDD-style scenarios

Use BDD-style scenarios for multi-step workflows.

Example:

```gherkin
Given a user belongs to a tenant
When the user creates a protected resource
Then the resource is scoped to the tenant
And an audit entry is written
And unauthorized tenants cannot read it
```

## Definition of done

Testing is complete only when:

- relevant tests exist
- tests pass locally
- CI validates tests
- failure cases are covered
- contract changes are tested
- bug fixes include regression tests
