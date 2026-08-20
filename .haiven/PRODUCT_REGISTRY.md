# Product Registry

This document defines the current Haiven product boundaries.

The goal is to prevent product repos from duplicating shared capabilities or creating conflicting platform primitives.

---

## Shared platform rule

A product repo owns its domain workflow. It should not permanently own common platform capabilities that belong at the Haiven ecosystem layer.

Before implementing reusable capabilities such as registration, authentication, identity, account management, tenant/workspace membership, roles, permissions, billing, notifications, QR generation, audit logging, telemetry, webhooks, file/export services, or API client generation, the product must evaluate whether the capability belongs in an existing shared service or should become a new shared service.

Temporary product-local implementations are allowed only when the shared service does not exist yet, the implementation is explicitly marked temporary, it is wrapped behind a provider/client interface, and migration to the shared service is documented.

No product may create a permanent local replacement for an existing Haiven shared service without a documented architecture exception.

---

## Product: Passage

Repo: `passage`

Status: planned

Purpose:

Passage is the shared Haiven registration, authentication, identity, account, session, tenant, role, and permission platform.

Passage owns:

- registration
- login
- account creation
- identity profiles
- sessions
- anonymous-to-account handoff
- tenant or workspace membership
- roles
- permissions
- auth provider integrations
- developer integration SDKs for Haiven products

Passage does not own:

- product-specific business workflows
- product-specific admin screens beyond identity/account administration
- product-specific authorization decisions after identity and claims are resolved

Products that should use Passage:

- qurl
- Heard
- future Haiven products

Rule:

No Haiven product should build a standalone auth stack unless Passage is unavailable and the repo documents a temporary adapter.

---

## Product: qurl

Repo: `qurl`

Status: active

Purpose:

qurl is the shared Haiven QR code generation product.

qurl owns:

- QR rendering
- QR export formats
- QR styling
- QR asset generation
- QR payload encoding rules
- QR design studio
- QR templates
- hosted dynamic QR behavior when explicitly selected
- scan analytics when explicitly selected
- developer API for QR generation

qurl does not own:

- restaurant feedback context
- Heard feedback workflows
- Passage identity
- product-specific destination authorization
- tenant-specific business rules from other products

Products that should use qurl:

- Heard
- any future Haiven product that needs QR generation

Rule:

If another Haiven product needs QR generation, it should call qurl instead of adding a QR rendering engine.

---

## Product: Heard

Repo: `heard`

Status: active

Purpose:

Heard is the open restaurant guest experience platform.

Heard owns:

- restaurant tenants
- locations
- guest feedback workflows
- survey campaigns
- recovery cases
- review routing
- offer/recovery workflows
- restaurant communications
- restaurant analytics
- POS/provider integrations
- feedback links and opaque tokens
- attribution context for restaurant interactions

Heard does not own:

- QR rendering
- shared identity/auth
- generic communication provider infrastructure beyond product-specific adapters
- shared Haiven design system primitives unless extracted into shared packages

Dependencies:

- Passage for auth and identity.
- qurl for QR generation.

Rule:

Heard should stay focused on restaurant guest experience workflows and avoid becoming a generic identity, QR, or messaging platform.

---

## Adding a product

A new Haiven product must document:

- product name
- repo name
- purpose
- owned capabilities
- explicitly not-owned capabilities
- shared services it depends on
- first vertical slice
- product-specific rules
- integration contracts with other Haiven products

Add the product to this registry before implementation begins.
