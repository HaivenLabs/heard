# Passage

Passage is the shared Haiven registration, authentication, identity, account, session, tenant, role, and permission platform.

## Product promise

Registration, identity, and access for products that should feel effortless from the first click.

## Responsibilities

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
- OAuth/OIDC/passkey/passwordless strategy as applicable
- developer integration SDKs for Haiven products

## Not owned by Passage

Passage does not own:

- qurl QR generation
- Heard restaurant workflows
- product-specific business logic
- product-specific authorization decisions after claims are resolved
- product-specific onboarding flows unless delegated by the product

## Integration goals

Haiven products should be able to integrate Passage with minimal friction.

A product should be able to:

- register a new user
- authenticate an existing user
- resolve current identity
- resolve tenant/workspace membership
- resolve roles and permissions
- support anonymous-to-account upgrade
- protect admin routes
- protect APIs
- audit identity-sensitive actions

## First slice

The first Passage slice should prove:

1. A user can create an account.
2. A user can log in.
3. A user can log out.
4. A product can resolve the current user.
5. A product can resolve tenant/workspace membership.
6. A product can enforce role and permission checks.
7. A local fake provider can be used in product repos before Passage is fully deployed.

## Design direction

Passage should feel:

- effortless
- secure
- fast
- boring in the best way
- invisible when it works
- clear when something fails

No product should lose users because Passage feels heavy.
