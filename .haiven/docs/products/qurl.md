# qurl

qurl is the shared Haiven QR code generation product.

## Product promise

QR codes for people who do not want weird redirects, ransom analytics, or ugly little checkerboards.

## Core invariant

If a user creates a direct QR code for `https://example.com`, the QR code points to `https://example.com`.

No surprise short links.

No forced redirect domains.

No tracking unless the user explicitly chooses it.

## Responsibilities

qurl owns:

- QR rendering
- QR export formats
- QR styling
- QR asset generation
- QR payload encoding rules
- QR design studio
- QR templates
- print-ready exports
- hosted dynamic QR behavior when explicitly selected
- scan analytics when explicitly selected
- developer API for QR generation

## Not owned by qurl

qurl does not own:

- Passage identity
- Heard restaurant workflows
- product-specific destination authorization
- product-specific customer or guest context
- product-specific analytics beyond QR scan events

## Integration goals

Other Haiven products should send qurl:

- destination URL
- rendering options
- export format
- optional style or template ID

qurl should return:

- QR asset URL
- downloadable asset
- SVG string
- binary payload
- asset reference

The exact return shape should be defined in qurl's OpenAPI contract.

## Privacy rule

Products should not send unnecessary tenant, user, guest, order, restaurant, or operational metadata to qurl.

qurl usually only needs:

- destination URL
- rendering options
- export format

## First slice

The first qurl slice should prove:

1. Anonymous user enters a destination URL.
2. qurl validates the URL.
3. qurl generates a direct QR payload.
4. qurl previews the QR.
5. qurl exports SVG and PNG.
6. The encoded payload is exactly the destination URL for direct QR.
7. The product works without login.

## Future slices

Future qurl slices may include:

- saved QR projects
- account handoff through Passage
- design studio
- templates
- brand kits
- EPS/PDF export
- dynamic hosted QR codes
- explicit tracking and analytics
- bulk generation
- developer API
- webhooks
- CLI
