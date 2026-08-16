# Security and Privacy Standards

Trust is a product requirement.

## Authentication

Haiven products should use Passage for shared identity.

Production admin surfaces must require authentication.

Local auth bypasses are allowed only when:

- clearly labeled
- non-production only
- impossible to enable accidentally in production
- compatible with the real auth contract

## Authorization

Products must enforce:

- tenant boundaries where applicable
- role checks
- permission checks
- least privilege
- protected routes
- protected APIs

## Sensitive data

Do not log sensitive data.

Sensitive data may include:

- passwords
- tokens
- API keys
- secrets
- private customer data
- PII
- payment data
- provider credentials
- internal session material

## Public identifiers

Public URLs should use opaque tokens or slugs instead of raw UUIDs when appropriate.

Internal IDs should not be exposed unless there is a clear reason.

## Secrets

Secrets must never be committed.

Use environment variables, secret stores, or deployment-specific secret management.

## Webhooks

Webhooks should use:

- signatures
- replay protection where practical
- structured payloads
- versioned contracts
- idempotent handling

## Audit logging

Sensitive actions should write audit entries.

Examples:

- login
- logout
- permission changes
- tenant membership changes
- provider credential changes
- export creation
- destructive actions
- sensitive admin actions

## Definition of done

Security work is complete only when:

- auth is enforced
- permissions are tested
- tenant isolation is tested where applicable
- secrets are not exposed
- logs are privacy-safe
- sensitive actions are audited
- failure modes are safe
