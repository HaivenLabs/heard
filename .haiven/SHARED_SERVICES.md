# Shared Services

This document defines shared Haiven services that product repos should consume instead of reimplementing.

---

## Shared platform capability rule

Haiven products should consume common services instead of rebuilding the same platform primitives in every product repo.

If a capability is needed by more than one product, or is likely to become a reusable ecosystem primitive, evaluate it as a shared Haiven service before implementing it product-locally.

Shared platform candidates include:

- registration
- authentication
- identity
- account management
- tenant or workspace membership
- roles and permissions
- billing and subscriptions
- notifications
- email/SMS provider abstraction
- QR generation
- audit logging
- telemetry
- design tokens
- admin shell primitives
- file/export services
- webhooks
- API client generation

Product repos may create temporary local implementations only when the shared service does not exist yet, the implementation is explicitly marked as temporary, the implementation is wrapped behind a provider/client interface, the contract matches the future shared service, and migration to the shared service is documented.

No product may create a permanent local replacement for an existing Haiven shared service without a documented architecture exception.

---

## Passage

Passage is the shared registration, authentication, identity, account, session, tenant, role, and permission platform.

### Responsibilities

Passage owns:

- registration
- login
- account creation
- sessions
- identity profiles
- anonymous-to-account handoff
- tenant or workspace membership
- roles
- permissions
- auth provider integrations
- OAuth/OIDC/passkey/passwordless strategy as applicable
- developer integration SDKs for Haiven products

### Product integration contract

Every Haiven product that needs user identity should integrate with Passage.

Products should receive:

- authenticated user identity
- tenant/workspace/account context
- role claims
- permission claims or permission lookup
- session state
- account status
- profile information where appropriate

Products should own:

- product-specific authorization checks
- product-specific onboarding after identity is resolved
- product-specific account/workspace screens
- product-specific audit events triggered by product actions

### Local development

Products may use a local fake Passage provider during early development.

The fake provider must be:

- clearly labeled
- restricted to non-production environments
- compatible with the real Passage contract
- covered by tests

---

## qurl

qurl is the shared QR code generation system.

### Responsibilities

qurl owns:

- QR code rendering
- SVG, PNG, JPG, EPS, PDF, and future export formats where supported
- QR styling
- QR design options
- QR templates
- QR asset generation
- direct QR payload rules
- hosted dynamic QR behavior when explicitly selected
- scan analytics when explicitly selected

### Product integration contract

Products should send qurl:

- destination URL
- rendering options
- export format
- optional style or template ID

Products should not send qurl unless explicitly required:

- raw tenant IDs
- raw user IDs
- guest data
- order data
- restaurant operational metadata
- sensitive attribution context

Products should store:

- qurl asset reference if useful
- generated file reference if downloaded or persisted
- product-specific link/entity association

### Direct QR rule

If a user creates a direct QR code for `https://example.com`, the QR code points to `https://example.com`.

No surprise short links.

No forced redirect domains.

No tracking unless the user explicitly chooses it.

### Local development

Products may use a local fake qurl provider during early development.

The fake provider must:

- return predictable test assets or asset references
- avoid pulling in a full QR rendering engine unless needed for tests
- preserve the production integration contract

---

## Future shared services

Potential shared services should be added here only when at least two products need them or when the capability is clearly platform-level.

Candidates:

- shared billing
- shared notification provider
- shared design token package
- shared audit/event viewer
- shared telemetry package
- shared admin shell
- shared export/file service

Do not create shared services prematurely.

Shared services must earn their existence.
