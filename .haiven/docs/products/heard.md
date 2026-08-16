# Heard

Heard is the open restaurant guest experience platform.

## Product promise

A restaurant guest experience platform that helps restaurants collect private feedback, recover unhappy guests, route happy guests to public reviews, and understand what is happening across locations.

## Responsibilities

Heard owns:

- restaurant tenants
- brands
- locations
- feedback links
- survey campaigns
- guest feedback
- recovery cases
- review routing
- restaurant communications
- restaurant analytics
- POS/provider integrations
- restaurant-specific attribution context
- feedback workflows
- recovery workflows

## Not owned by Heard

Heard does not own:

- QR rendering
- shared identity/auth
- generic account/session infrastructure
- generic QR design studio
- shared provider infrastructure beyond product-specific adapters

## Shared service dependencies

Heard should use:

- Passage for auth, identity, accounts, roles, sessions, and tenant membership.
- qurl for QR generation, QR rendering, QR styling, and QR exports.

## QR rule

Heard generates the feedback destination URL.

Heard sends that URL to qurl.

qurl generates the QR asset.

Heard stores the qurl asset reference and associates it with the feedback link or campaign.

Heard should not pass unnecessary restaurant, guest, tenant, order, or operational metadata to qurl.

## First slice

The first Heard slice should prove:

1. A restaurant tenant can be created.
2. A location can be created.
3. A feedback link can be generated.
4. QR generation goes through qurl or a fake qurl provider.
5. A guest can open a feedback form.
6. A guest can submit feedback.
7. Negative feedback creates a recovery case.
8. A manager can see the recovery case.
9. The system is observable, tested, tenant-safe, and polished.

## Design direction

Heard should feel:

- calm
- modern
- restaurant-friendly
- operationally useful
- trustworthy
- fast
- polished enough for public screenshots

The guest feedback flow must feel frictionless.

The manager experience must feel like a serious SaaS product, not an internal admin panel.
