# Slice 7 plan: guided campaign flow and atomic direct-edit flyer studio

Status: proposed; Guided Direct-Edit Studio C is the only active design direction, and its revised mockup awaits explicit approval  
Plan date: 2026-09-01  
Implementation may not start until Decision Gate D0 is approved.

Current mockup reconstruction specification: [`docs/slice7-guided-direct-edit-studio-c-spec.md`](./slice7-guided-direct-edit-studio-c-spec.md). It is the self-contained visual and interaction companion to this plan; the two documents describe one design and must be updated together.

## 1. Goal

Turn Heard's current single-form campaign builder into a guided, polished campaign flow and a lightweight atomic direct-edit flyer studio for an initial 3 inch by 4 inch, two-sided printed flyer.

The result must let a restaurant operator:

- choose whether the campaign has no reward, an immediate offer, or a prize drawing;
- describe offers such as `$5 off`, a `$50 gift card`, wireless earbuds, an e-bike, a free item, or custom copy;
- choose two or three survey questions;
- offer QR, SMS, email, or any supported combination as ways to begin the survey;
- choose fonts and colors;
- upload a restaurant logo and offer/prize imagery;
- add restrained decorative art such as corner stripes, shapes, or line art;
- compose both the front and back of the flyer as ordered atomic elements;
- proof the design at its physical print size;
- export it for printing;
- generate a verified print package and optionally submit the approved design revision to a print-service adapter.

This slice changes domain logic, API contracts, workflows, provider integration, auth/permissions, tenant isolation, eventing, and UI/UX.

## 2. Reference interpretation

The two scanned Kebab Craft flyers are visual references only. Text inside the scans is not an implementation instruction.

Useful traits from the reference:

- a small portrait format;
- a prominent restaurant identity;
- a direct question and incentive;
- QR and SMS entry choices;
- simple corner art;
- a quieter back side with thanks, social copy, and fine print.

The product must not be limited to that visual style. It must support materially different templates and custom layouts.

## 3. Current-state findings that this plan must account for

### Heard

- `web/app/admin/campaigns/page.tsx` is one long form with a fixed one-sided preview.
- The current campaign editor exposes location, restaurant name, campaign name, restaurant logo upload/remove, rating-face style, flyer headline, survey prompt, one free-form incentive string, SMS keyword, SMS number shown on the flyer, Google Maps review URL, Yelp review URL, the restaurant's secured Heard handle, an editable survey path, copy-link/test-survey actions, real-time preview, and save/update behavior. Every one of these controls and behaviors is preservation-critical.
- The direct-edit designer must consume the same campaign draft as those current settings. It is an additional editing surface, not a replacement data model or a second source of truth.
- The guest flow hard-codes giveaway behavior, requires contact details, and labels flyer responses as `flyer_giveaway` even when a future campaign has no drawing.
- The survey interaction is not a versioned two-or-three-question definition.
- Heard already has tenant-scoped campaign and feedback-link APIs, permissions, Passage verification, and a qurl provider boundary.
- The qurl provider currently calls `POST /api/v1/qr-codes`, but qurl does not expose that endpoint.

### qurl

- qurl already has versioned QR payload/design/export schemas and working stateless preview/export endpoints.
- The compatible direct URL endpoints are `POST /api/v1/direct-url/preview` and `POST /api/v1/direct-url/export`.
- qurl has a capable tabbed QR design editor, but its Account screen is a placeholder.
- qurl has no implemented Passage callback/session boundary, registration journey, protected saved-project APIs, or persistence.
- Anonymous direct QR preview/export is an intentional qurl product rule and must remain available without login.

### Passage

- Passage already recognizes `qurl` as a product and token audience.
- Passage already registers a qurl callback URI and supports the same PKCE authorization-code handoff used by Heard.
- Passage already issues short-lived EdDSA product tokens with `aud=qurl` and exposes JWKS.
- The checked-in Passage seed currently names an HTTP localhost qurl callback while qurl's Docker web origin terminates local HTTPS. Q1 must supply one exact callback URI that matches qurl's actual public origin; the integration test must fail on a scheme, host, port, or path mismatch.
- No Passage repository change is planned unless implementation proves an actual contract defect. Configuration alone is not a reason to edit Passage.

### Worktree safety

Both Heard and qurl currently contain user-owned uncommitted changes. Every implementation agent must:

1. run `git status --short --branch` before editing;
2. inspect diffs in every file it intends to touch;
3. preserve existing changes;
4. never reset, checkout, or overwrite the worktree to obtain a clean state;
5. stop and report a real overlap that cannot be safely merged.

## 4. Decision Gate D0 and selected design direction

Three concepts were reviewed. Options A and B are retained only as closed decision history. They are not implementation choices. Guided Direct-Edit Studio C is the sole active direction, and production implementation remains blocked until its revised mockup is explicitly approved.

The current Guided Direct-Edit Studio C mockup is documented element by element, state by state, and breakpoint by breakpoint in [`docs/slice7-guided-direct-edit-studio-c-spec.md`](./slice7-guided-direct-edit-studio-c-spec.md). That file supplies the reconstruction detail for this plan, and both files are required.

### Closed alternative A: Guided wizard

Structure:

1. Offer
2. Questions
3. Ways to respond
4. Brand
5. Front and back
6. Proof and export

Strengths:

- lowest cognitive load;
- closest to qurl's progressive design flow;
- best for a first campaign and small screens;
- easy to validate before the operator advances.

Tradeoff:

- freeform layout can feel hidden or slow for experienced users.

### Closed alternative B: Outcome recipe

The operator starts with the business outcome:

- bring guests back now;
- collect more signal with a drawing;
- request feedback without a reward.

Heard then proposes a complete recipe containing the hook, question count, reward, response methods, and visual direction. Every line remains editable.

Strengths:

- starts in the operator's language rather than design terminology;
- creates better first drafts quickly;
- makes campaign logic and flyer copy feel like one coherent story.

Tradeoff:

- operators who already know the exact layout may want direct access to the canvas.

### Selected direction C: Guided direct-edit studio

A campaign studio that keeps the current settings recognizable and makes the flyer itself the advanced editor:

- the current controls are grouped into Campaign, Survey, Offer, Responses, Brand, and Link and reviews sections;
- the generated 3 by 4 inch front/back flyer remains visible beside those settings;
- selecting text, an image, the QR/contact block, or a shape on the flyer opens a small contextual editor for that item;
- operators drag items on the flyer, pull a corner handle to resize, double-click text to edit it inline, or use exact position/size fields for fine control;
- new Design controls add templates, fonts, colors, uploads, atomic text/image/QR/shape elements, safe area, bleed, front/back pages, proofing, print-file export, and print-service ordering;
- templates create a strong first draft but do not lock copy, placement, size, color, typography, imagery, or art;
- guided settings and direct canvas edits update one shared campaign/design draft.

Strengths:

- preserves the existing Heard campaign-builder experience while supporting a lightweight direct-edit design experience;
- gives non-designers a guided starting point while experienced operators edit items directly on the flyer;
- is the best fit for experienced operators and reusable templates.

Tradeoff:

- requires careful progressive disclosure so the added canvas does not make the familiar settings harder to use.

### Current direction

Use **Guided Direct-Edit Studio C**, with progressive disclosure inside the selected canvas direction:

1. Preserve every current campaign control and behavior; reorganize them into concise section tabs without changing their stored values or meaning.
2. Keep the generated front/back flyer visible while settings are edited so the experience feels like one continuous flow.
3. Let the operator click any flyer item to edit the actual copy or asset, drag it to move, resize it with handles, and fine-tune font, color, alignment, position, and dimensions in a contextual editor.
4. Persist every item with explicit z-order for deterministic rendering and export; customers select and arrange those items directly on the flyer.
5. Add structured mechanics, two/three questions, QR/SMS/email combinations, templates, typography, colors, uploads, atomic art primitives, and front/back composition as expansions of the existing form.
6. Autosave every draft change with offline recovery and conflict handling.
7. Finish in the same studio with onscreen proofing, downloadable print files, and a separately confirmed Send to print service flow.

The user selected C, required preservation of the existing Heard experience, and chose direct manipulation of flyer items. D0 remains open until the revised mock is explicitly approved. No current control, preview behavior, link action, or persisted field may be deleted or silently replaced during implementation.

Decision Gate D0 checklist:

- [x] Three concepts mocked up.
- [x] Front and back states shown.
- [x] Small-screen behavior mocked.
- [x] User selected Guided Direct-Edit Studio C.
- [x] Selected interaction rules and visual direction are recorded here.
- [x] Scope changes from the review are incorporated into the task cards below.
- [x] Direct-manipulation revision is mocked.
- [ ] Revised Guided Direct-Edit Studio C mock is explicitly approved.

## 5. Product rules

### 5.1 Campaign mechanics

`campaign_mechanic` is authoritative server-side and must be one of:

- `feedback_only`
- `instant_offer`
- `prize_drawing`

Do not infer the mechanic from flyer copy or client metadata.

Rules:

- `feedback_only` does not require guest contact information unless the guest requests follow-up.
- `instant_offer` displays configured redemption instructions after completion. POS-backed redemption is a later integration.
- `prize_drawing` requires at least one valid contact method and explicit drawing terms or an official-rules URL before publication.
- Marketing consent is always separate, optional, and unchecked by default.
- Transactional contact consent is requested only when the chosen mechanic or follow-up requires contact.
- The UI must never call every flyer a giveaway.

### 5.2 Survey questions

The first flyer format supports exactly two or three visible survey questions. Contact collection and consent do not count as survey questions.

Question rules:

- Question 1 is an overall 1-to-5 rating because recovery and sentiment logic depend on it.
- Question 2 or 3 may be `single_select`, `multi_select`, or `short_text`.
- Each question has an ID, type, prompt, required flag, position, and type-specific options.
- Options are bounded, ordered, unique, and plain text.
- The published campaign stores an immutable question revision so historical responses remain interpretable after a draft changes.
- The guest API accepts answers keyed by question ID. It does not accept arbitrary answer metadata.

### 5.3 Ways to respond

Supported response methods are separate from visual elements:

- `web_qr`: qurl encodes the Heard feedback URL directly.
- `sms_keyword`: an inbound provider maps a verified number and unique keyword to the campaign and replies with the Heard feedback URL.
- `email_alias`: an inbound provider maps a verified address/alias to the campaign and replies with the Heard feedback URL.

Rules:

- The operator may choose any enabled combination.
- QR can be used alone.
- SMS or email may be published only when the provider route passes a verification check.
- A disabled or unconfigured provider must not produce a dead phone number or email address on a flyer.
- SMS STOP/HELP behavior, consent copy, and transactional/marketing separation are required.
- Local development uses fake inbound SMS and email adapters with inspectable messages.

### 5.4 Physical flyer specification

Initial format ID: `flyer-3x4-duplex-v1`.

- Trim size: 3.00 by 4.00 inches, portrait.
- Bleed: 0.125 inch on every edge.
- Document size with bleed: 3.25 by 4.25 inches.
- Canonical coordinate system: 72 points per inch.
- Trim canvas: 216 by 288 points.
- Bleed box: 234 by 306 points.
- Default safe inset: 0.125 inch inside the trim edge.
- Raster export target: 975 by 1275 pixels at 300 DPI, including bleed.
- Two named sides: `front` and `back`.
- Preview may zoom, but stored geometry is always physical-unit based.

### 5.5 Atomic flyer elements

Version 1 exposes a fixed palette of explicit, bounded element types. Templates pre-populate the right collection of elements; operators may add more instances up to the per-side limit:

- `text`
- `image`
- `shape`
- `decorative_art`
- `qr`
- `rating_scale`

Every element has:

- opaque UUID;
- side;
- name;
- position and size in points;
- rotation;
- z-order;
- conditional visibility state where a campaign setting controls whether the item is rendered;
- type-specific content/style;
- accessibility/meaning label for editor use.

Do not store executable markup, arbitrary CSS, arbitrary HTML, or unbounded metadata in an element.

Atomicity rules:

- one text element contains one plain-text value and one typography/style record; rich-text runs and multiple font sizes inside one text element are out of scope;
- differently sized or styled copy is represented by separate text elements;
- an offer card is a shape element plus separate title and detail text elements;
- a response area is a QR element plus separate instruction, phone, and email text elements;
- a restaurant logo is an image element and opens image controls, never text controls;
- a QR element opens QR controls and cannot contain instruction copy;
- a shape element has a `rectangle`, `square`, `circle`, or `squircle` variant and opens only geometry/fill/border controls;
- templates may place related elements near each other but do not merge them into a mixed-content editing box;
- each side permits at most 40 elements in this slice.

### 5.6 Editor behavior

Required interactions:

- keep the current settings available in recognizable Campaign, Survey, Offer, Responses, Brand, and Link and reviews sections;
- update the flyer canvas immediately from guided setting edits and update the same shared draft from direct canvas edits;
- select a flyer item by clicking it directly on the canvas;
- edit the exact displayed wording from the contextual editor or inline on the canvas, including response instructions independent of the configured SMS keyword;
- show type-correct controls: typography for text, replace/crop/fit for images, render settings for QR, and geometry/fill/border for shapes;
- drag, resize, and rotate within bounded rules;
- bring forward/send backward, duplicate, and delete from contextual actions;
- keyboard nudge by 1 point and accelerated nudge by 10 points;
- align to page, safe area, and nearby element guides;
- undo and redo within the current browser editing session;
- edit plain text, font, size, weight, alignment, color, line spacing, and letter spacing;
- crop and fit images without modifying the original asset;
- switch between front and back;
- reset one side to its selected template after confirmation;
- autosave with visible `saving`, `saved`, `offline changes`, and `save failed` states.

Use bounded autosave after the current form is migrated. Persist campaign and design changes as one draft revision, retain dirty state locally on network failure, and require explicit conflict resolution instead of silently overwriting another editor.

The same responsive interface is used at all breakpoints. On narrow screens the guided/contextual editor stacks above the canvas and the canvas zooms; functionality is not split into a separate mobile product. The persisted document retains explicit element z-order for rendering and export.

Preservation rule: opening an existing campaign and saving it without deliberate edits must produce no semantic change to any current field, feedback link, or review destination. New structured fields use explicit defaults and compatibility adapters; the UI must never infer that a legacy incentive is a drawing and overwrite it.

### 5.7 Fonts, colors, imagery, and art

Fonts:

- ship a reviewed, license-compatible curated set;
- store the font family by stable ID, not a remote CSS URL;
- embed or outline fonts in export only when the license permits it;
- custom font upload is out of scope for this slice.

Colors:

- provide a tenant brand palette plus per-element color controls;
- store normalized 6- or 8-digit hex values;
- warn on low contrast for essential text and QR codes.

Images:

- accept bounded PNG, JPEG, and WebP uploads;
- decode and re-encode uploads, remove metadata, and reject malformed files;
- reject SVG uploads until an approved sanitizer exists;
- store tenant-scoped asset records and never trust a client-supplied cross-tenant asset reference;
- calculate effective DPI at the placed physical size and warn below the print threshold.

Decorative art:

- start with code-owned SVG primitives such as stripes, arcs, bursts, stars, blobs, checks, and simple food/celebration marks;
- allow recolor, scale, rotate, and corner placement;
- keep art decorative and separate from actionable text.

### 5.8 QR ownership and data minimization

- Heard creates the public feedback URL and owns its opaque path/token.
- qurl renders the QR.
- Heard sends qurl only the destination URL and QR render configuration.
- Heard does not send tenant IDs, user IDs, restaurant operations, guest data, question answers, or offer details to qurl.
- Direct QR output must encode the Heard URL directly. Tracking or redirect behavior is opt-in and out of scope here.
- qurl markup is consumed only as an isolated image. It is never injected into Heard's DOM as trusted HTML.

### 5.9 Preview, export, and print service

These are three distinct outcomes:

- `Preview & proof` is an onscreen, non-downloadable inspection mode for front/back appearance, safe area, bleed, effective image DPI, QR readability, overflow, and contrast.
- `Download print files` is the file-export action and produces the verified design revision as an exact-size two-page PDF plus optional 300 DPI PNG and SVG files.
- `Send to print service` begins a provider-neutral print-order flow from the exact proofed export revision.

Print-service rules:

- the source of truth remains `heard.flyer-design.v1`; a print provider receives an immutable export, not the editable design document;
- Heard uses a provider adapter selected during implementation discovery and does not couple campaign logic to one vendor;
- the operator chooses quantity, stock/finish options supported for the 3-by-4 duplex product, shipping address, and shipping speed;
- Heard shows the provider quote, taxes, shipping, estimated delivery window, and the exact proof before the operator confirms payment/order submission;
- order submission requires explicit confirmation and an idempotency key; autosave never places or charges an order;
- webhook/order status is authenticated, replay-safe, tenant-scoped, and visible as submitted, accepted, in production, shipped, delivered, canceled, or failed;
- provider failure never damages the Heard design or print export, and retry never duplicates an order.

## 6. Heard contract and data plan

OpenAPI must change before handlers or UI.

### 6.1 New schemas

Add versioned schemas for:

- `CampaignMechanicV1`
- `CampaignOfferV1`
- `SurveyQuestionV1`
- `SurveyAnswerV1`
- `ResponseMethodV1`
- `FlyerDesignV1`
- `FlyerSideV1`
- `FlyerElementV1` and each explicit atomic element subtype
- `FlyerProofResultV1`
- `FlyerExportRequestV1`
- `FlyerExportV1`
- `DesignAssetV1`
- `PrintQuoteRequestV1`
- `PrintQuoteV1`
- `PrintOrderRequestV1`
- `PrintOrderV1`

Canonical design schema version: `heard.flyer-design.v1`.

### 6.2 Endpoints

Campaign:

- `POST /api/v1/survey-campaigns`
- `PATCH /api/v1/survey-campaigns/{id}`
- `POST /api/v1/survey-campaigns/{id}/publish`
- `GET /api/v1/public/surveys/{token}` and path equivalent return the published question revision and mechanic-safe public fields.

Flyer design:

- `GET /api/v1/survey-campaigns/{campaign_id}/flyer-design`
- `PUT /api/v1/survey-campaigns/{campaign_id}/flyer-design`
- `POST /api/v1/flyer-designs/{id}/proof`
- `POST /api/v1/flyer-designs/{id}/exports`
- `GET /api/v1/flyer-exports/{id}`

Print service:

- `POST /api/v1/flyer-exports/{id}/print-quotes`
- `POST /api/v1/print-orders`
- `GET /api/v1/print-orders/{id}`
- provider-specific webhook routes remain behind the print-provider adapter.

Assets:

- `POST /api/v1/design-assets`
- `GET /api/v1/design-assets?cursor=...&page_size=...`
- `DELETE /api/v1/design-assets/{id}` only when no published design references the asset.

Provider verification:

- `POST /api/v1/response-methods/sms/verify`
- `POST /api/v1/response-methods/email/verify`

Inbound provider webhooks are provider-adapter routes and must have signature verification, replay protection, bounded bodies, structured errors, and idempotency.

### 6.3 Concurrency and idempotency

- `PUT flyer-design` requires a revision/ETag and returns `409 revision_conflict` on a stale write.
- Create, publish, export, print-quote, print-order, and provider-verification mutations require idempotency keys.
- Autosave never publishes.
- Publish snapshots the campaign questions, offer, response methods, and flyer revision atomically.
- Duplicate inbound messages, export retries, and print-order retries return the original durable result.

### 6.4 Persistence

Add migrations for:

- campaign mechanic and structured offer;
- ordered survey questions and published question revisions;
- configured response methods;
- flyer designs and immutable published revisions;
- tenant-scoped design assets;
- flyer exports and status;
- print quotes, orders, immutable submitted artwork revision, and provider status history;
- inbound message deduplication where SMS/email are enabled.

Prefer normalized columns for authorization, lifecycle, and lookup fields. Store the validated versioned design document as JSONB because element shapes are explicit and schema-validated. Do not put tenant ID only inside JSON.

### 6.5 Migration of current campaigns

- Existing non-empty `incentive_text` campaigns migrate to a draft `prize_drawing` offer only when current behavior actually requires giveaway contact; otherwise mark them `needs_review` rather than guessing.
- Existing logo/theme/rating-face values seed the first generated template.
- Existing feedback links and destinations remain unchanged.
- Existing campaigns receive a generated `heard.flyer-design.v1` only when first opened or by an idempotent backfill.
- The old public survey remains readable during the rollout; publish switches a campaign to the versioned question contract.

## 7. qurl contract, auth, and registration plan

All work in this section belongs in `D:\\dev\\HaivenLabs\\qurl`. No qurl implementation file may be added to Heard.

### 7.1 Repair Heard-to-qurl compatibility

qurl remains the contract owner.

Heard must send this shape to qurl's existing direct URL contract:

```json
{
  "schemaVersion": "qurl.qr-project-config.v1",
  "name": "Heard feedback QR",
  "payload": {
    "schemaVersion": "qurl.qr-payload-config.v1",
    "kind": "url",
    "payload": {
      "destinationUrl": "https://heard.example/f/opaque-path",
      "normalize": false
    }
  },
  "design": {
    "schemaVersion": "qurl.qr-design-config.v1",
    "errorCorrectionLevel": "M",
    "quietZoneModules": 4,
    "foregroundColor": "#18211F",
    "backgroundColor": "#FFFFFF",
    "moduleStyle": "square",
    "eyeStyle": "square"
  },
  "export": {
    "schemaVersion": "qurl.qr-export-config.v1",
    "format": "svg",
    "fileName": "heard-feedback-qr.svg"
  }
}
```

Implementation order:

1. Add explicit non-2xx problem responses to qurl OpenAPI if they are missing.
2. Add/confirm qurl contract tests for direct destination preservation and bounded SVG/PNG output.
3. Update Heard's provider to call `/api/v1/direct-url/preview` for an isolated preview image.
4. Use `/api/v1/direct-url/export` when a print export needs the binary QR artifact.
5. Add Heard contract tests against an HTTP fake that implements qurl's real contract.
6. Add a cross-repo acceptance test or documented script that starts qurl and Heard and proves the decoded QR destination equals the Heard URL.

Stateless preview/export stays anonymous and rate-limited. Heard must not forward its `aud=heard` user token to qurl.

### 7.2 Passage auth in qurl

Use the existing Passage PKCE/JWKS contract. Do not invent qurl-local passwords or sessions.

Backend steps:

1. Add qurl config for Passage issuer, public URL, backend URL, client ID `qurl`, exact callback URI, secure-cookie mode, and bounded JWKS cache settings.
2. Use the qurl public HTTPS origin for local Docker and configure Passage's `PASSAGE_QURL_CALLBACK_URIS` to the same exact URI in the combined acceptance runtime. Do not weaken qurl to HTTP to match an old seed default.
3. Add provider-neutral qurl routes for auth start, callback, session, and logout.
4. Generate state, PKCE verifier/challenge, and return path server-side.
5. Store state and verifier only in short-lived HttpOnly SameSite cookies.
6. Exchange the one-time code server-side.
7. Verify EdDSA signature, issuer, `aud=qurl`, expiry, subject UUID, organization UUID, role, and qurl product grant.
8. Store the verified qurl product token only in a Secure/HttpOnly session cookie.
9. Cache JWKS with bounded refresh and force one refresh after key/signature mismatch.
10. Make local bypass impossible outside explicit local/test runtime and loopback origins.

Frontend steps:

1. Keep anonymous create/preview/export available.
2. Replace the Account placeholder with qurl-branded sign in, registration, recovery, authenticated account, and sign-out states.
3. Trigger registration or sign-in when an anonymous user chooses `Save`.
4. Preserve the canonical anonymous project in bounded browser storage during the handoff.
5. After callback, attach the draft to the verified organization through a protected qurl project API.
6. Remove the handoff draft after confirmed persistence; retain it on a retryable failure.
7. Never show Passage branding or provider names in qurl customer-facing copy except a user-selected external identity button such as Google.

### 7.3 qurl saved-project vertical slice

Registration without a useful saved-project outcome is incomplete.

Add, contract-first:

- tenant-scoped PostgreSQL project persistence;
- `POST /api/v1/projects`;
- `GET /api/v1/projects` with cursor pagination;
- `GET /api/v1/projects/{id}`;
- `PATCH /api/v1/projects/{id}` with revision checks;
- minimum owner/admin/member write and viewer read rules;
- anonymous-to-account draft import;
- audit events and `qr_project.saved.v1` outbox event;
- cross-tenant and unauthorized tests.

This qurl project capability is useful to qurl users. Heard does not depend on qurl saved projects for flyer rendering.

## 8. Rendering, proofing, and export plan

### 8.1 Renderer boundary

Define a Heard-owned `FlyerRenderer` interface. The design document is renderer-neutral.

Required outputs:

- front and back editor previews;
- front and back SVG;
- front and back 300 DPI PNG with bleed;
- two-page print PDF with exact physical dimensions;
- a provider-ready immutable print artifact for the proofed revision.

The implementation agent must perform a supply-chain review before adding a renderer dependency. The selected renderer must be pinned, deterministic, self-hostable, testable in CI, and unable to fetch arbitrary network resources during rendering.

Do not run untrusted HTML in a privileged browser. Images and fonts come only from validated, tenant-authorized stored assets and the curated font registry.

### 8.2 Proof checks

Proof returns errors and warnings separately.

Errors block publish/export:

- missing required element content;
- element outside the bleed box when it is not explicitly bleed art;
- essential text outside the safe area;
- QR missing when QR is selected;
- QR destination differs from the campaign feedback URL;
- broken/missing asset;
- response method shown but not verified;
- text overflow or non-renderable font;
- missing drawing terms/rules for `prize_drawing`.

Warnings require acknowledgement:

- low text contrast;
- QR smaller than the product's reviewed print threshold;
- QR quiet zone or styling risk reported by qurl;
- low effective image DPI;
- unusually small fine print;
- important content close to trim;
- back side is empty;
- SMS or email copy may wrap poorly.

Proofing never claims guaranteed scan performance. Export tests include decoding rendered QR images at representative print scales.

### 8.3 Export jobs

- Export is represented by a durable record with `queued`, `rendering`, `ready`, or `failed` status.
- State change and outbox insert are atomic.
- The worker retries safely with exponential backoff and idempotent artifact keys.
- Failure details are operator-visible but do not leak private asset URLs or guest data.
- Download URLs are short-lived and tenant-authorized.
- Old artifacts follow a documented retention policy; source designs and published revisions remain durable.

## 9. Permissions and tenant isolation

Minimum permissions:

- `campaign:read`: read campaign and flyer design.
- `campaign:write`: edit draft campaign and flyer design.
- `campaign:publish`: publish and replace the active revision.
- `campaign:export`: request/download print artifacts.
- `campaign:print`: request quotes and explicitly submit print orders.
- `asset:write`: upload/delete tenant design assets.

Rules:

- The backend derives tenant and actor from verified Passage claims.
- Client-supplied tenant IDs are never authority.
- Every campaign, design, revision, asset, response method, and export query is tenant-scoped.
- Background jobs persist and revalidate tenant context.
- An asset may be placed only when it belongs to the same tenant.
- Public survey responses resolve campaign context from the opaque feedback link, not submitted campaign metadata.

## 10. Events and audit

Versioned Heard events:

- `survey_campaign.configured.v1`
- `flyer_design.created.v1`
- `flyer_design.updated.v1`
- `flyer_design.published.v1`
- `flyer_export.requested.v1`
- `flyer_export.completed.v1`
- `flyer_export.failed.v1`
- `print_order.submitted.v1`
- `print_order.status_changed.v1`
- `response_method.verified.v1`

Audit entries record actor, tenant, entity, revision, action, correlation ID, and safe change summary. They do not store full uploaded assets, guest responses, tokens, or raw design JSON in log fields.

qurl owns its own auth/project audit and events. Heard does not write qurl audit tables.

## 11. Failure modes

- qurl unavailable during edit: keep the saved Heard design and destination; show a retryable QR preview error.
- qurl unavailable during export: fail the export job visibly and retry; never substitute a locally rendered QR engine.
- Passage unavailable in qurl save flow: preserve the anonymous draft and allow retry; anonymous preview/export still works.
- Browser loses connection during autosave: retain dirty state locally and surface it; do not report `saved`.
- Stale revision: return `409`, show the newer server revision, and offer explicit reload or duplicate; never silently overwrite.
- Asset upload fails validation: return field-specific structured errors and leave the design unchanged.
- Image later becomes unavailable: published revision continues using its immutable asset reference or export fails visibly.
- SMS/email provider unavailable: QR remains usable; unverified channels cannot be published.
- Renderer crashes: job becomes retryable/failed with an operator-visible reason.
- Print provider quote/order fails: Heard design and export remain intact; retry uses the same idempotency key and never duplicates an order.
- User closes the editor mid-change: next load offers the recoverable local draft only when it matches tenant, campaign, and base revision.

## 12. Test plan

Tests are written before implementation in each task.

### 12.1 Domain unit tests

- mechanic validation;
- two/three-question validation;
- offer rules;
- contact requirement by mechanic;
- drawing publication gate;
- element bounds, atomicity, and type validation;
- font/color validation;
- physical unit to pixel conversion;
- proof severity classification;
- qurl request data minimization;
- response-method verification rules;
- migration/backfill idempotency.

### 12.2 Contract tests

- every new Heard request/response against OpenAPI;
- `heard.flyer-design.v1` valid and invalid examples;
- qurl direct URL preview/export request and response;
- qurl Passage product-token validation;
- structured non-ad-hoc errors;
- generated clients/types refreshed where practical.

### 12.3 Integration and permission tests

- campaign/design/asset/export persistence;
- optimistic concurrency;
- publish snapshot atomicity;
- outbox atomicity and retry;
- cross-tenant campaign, asset, design, and export denial;
- role-based read/write/publish/export denial;
- qurl cross-tenant project denial;
- SMS/email webhook signature and replay rejection;
- qurl and provider outage behavior.

### 12.4 Browser and BDD scenarios

```gherkin
Scenario: Open and save an existing campaign without losing current work
  Given a campaign already has a location, restaurant and campaign name, logo, rating-face style, headline, prompt, incentive copy, SMS keyword and number, Google and Yelp review URLs, and an editable Heard survey path
  When an owner opens Guided Direct-Edit Studio C
  And switches between every settings section, Front, and Back
  And selects flyer items directly without changing them
  And autosave completes without deliberate value changes
  Then every existing value round-trips unchanged
  And copy link, test survey, real-time preview, and create/update behavior remain available
  And the new design state references the same campaign draft rather than a replacement campaign
```

```gherkin
Scenario: Show type-correct controls for atomic template elements
  Given a template contains separate offer rectangle, offer title, offer detail, QR, instruction, phone, email, and logo elements
  When the owner selects the offer title
  Then Heard shows copy and typography controls for only that text element
  When the owner selects the logo
  Then Heard shows replace, crop, fit, opacity, size, and position controls without text controls
  When the owner selects the QR code
  Then Heard shows qurl render controls and the read-only Heard destination without instruction-copy controls
  When the owner selects the offer rectangle
  Then Heard shows shape, fill, border, corner, size, and position controls without typography controls
```

```gherkin
Scenario: Autosave a draft
  When an owner changes copy, font size, or element position
  Then Heard immediately marks the draft dirty
  And a bounded autosave writes one new draft revision
  And the UI reports Saving then Saved
  But autosave never publishes, exports, places, or charges a print order
```

```gherkin
Scenario: Preview, download, and print have distinct outcomes
  Given the current draft passes proof
  When the owner chooses Preview & proof
  Then Heard shows the exact front and back onscreen without downloading a file
  When the owner chooses Download print files
  Then Heard downloads the authorized exact-size export package
  When the owner chooses Send to print service
  Then Heard starts quote and confirmation without submitting an order yet
```

```gherkin
Scenario: Rewrite and reposition response instructions without changing the SMS destination
  Given a flyer displays the generated copy "TEXT TASTE"
  And its configured SMS keyword is "TASTE" with a verified phone number
  When the owner selects that text directly on the flyer
  And changes it to "TEXT THE WORD \"TASTE\""
  And drags and resizes the text block
  Then the flyer renders the exact custom wording and placement
  And the SMS keyword and verified destination remain unchanged
  And undo restores the prior wording, position, and size
```

```gherkin
Scenario: Create an immediate discount flyer with two questions
  Given an owner has a restaurant, location, and verified QR and SMS methods
  When they choose an instant $5 offer and two questions
  And they customize both flyer sides
  Then Heard autosaves an atomic-element design draft
  And proof confirms the QR destination is the Heard feedback URL
  And the guest receives redemption instructions after answering two questions
```

```gherkin
Scenario: Create a three-question prize drawing
  Given an owner uploads licensed prize imagery
  When they choose a custom wireless-earbuds drawing
  And add three questions, contact collection, and official rules
  Then Heard publishes the selected design revision
  And a guest cannot enter without a valid contact method
  And marketing consent remains optional
```

```gherkin
Scenario: Create feedback without a reward
  When an operator selects feedback only
  Then the flyer contains no drawing claims
  And the guest can submit without contact details
  And a low rating still opens the recovery workflow
```

```gherkin
Scenario: qurl is unavailable during flyer editing
  Given a valid atomic-element Heard draft
  When qurl preview fails
  Then the draft remains saved
  And the operator sees a retryable QR error
  And Heard does not render a replacement QR locally
```

```gherkin
Scenario: Save an anonymous qurl project through Passage
  Given an anonymous user has a valid qurl project draft
  When they choose Save and register
  Then Passage returns a one-time PKCE-bound qurl authorization code
  And qurl verifies an aud=qurl token
  And the draft is saved under the verified organization exactly once
```

### 12.5 Visual and accessibility tests

- screenshot tests for the guided section flow, generated preview, direct-edit canvas, front/back, proof, and error states;
- 360, 768, 1024, and 1440 pixel widths;
- keyboard-only element selection, z-order actions, nudge, and property editing;
- visible focus, labels, error association, contrast, and reduced motion;
- zoomed canvas without clipped controls;
- no essential state communicated only by color;
- print snapshots at exact SVG/PDF bounds.

### 12.6 End-to-end QR and print tests

- export a real flyer through qurl;
- decode the rendered QR and compare its destination byte-for-byte with the Heard URL;
- rasterize both PDF pages and compare dimensions and critical-region snapshots;
- confirm bleed, safe area, and front/back order;
- test a low-resolution image warning and a missing-asset hard failure.

## 13. Observability

Metrics:

- builder started/completed by selected flow;
- drop-off and time per guided step;
- template selected;
- direct-edit canvas used;
- autosave latency/failure/conflict;
- proof errors/warnings by code;
- qurl preview/export latency and failure;
- export queue time, render time, retries, and failures;
- response method verification/failure;
- QR/SMS/email campaign entry and completion, without logging message bodies;
- qurl registration, auth failure, anonymous draft handoff, and project save.

Structured logs and traces carry correlation ID, tenant ID, actor ID, campaign ID, design revision, export ID, print order ID, and provider operation. Do not log auth tokens, QR SVG bodies, uploaded image bytes, guest answers, contact details, shipping addresses, payment tokens, or provider secrets.

## 14. Implementation task cards

Each task card is intentionally bounded so a `gpt-5.6-terra` agent at `medium` reasoning can execute it without inventing missing architecture.

### D0 — Approve the Guided Direct-Edit Studio C design

Repository: Heard docs only  
Model: `gpt-5.6-sol`, reasoning `high`  
Writes: this plan and design-decision notes only  
Depends on: none

Durable visual reference: [`docs/slice7-guided-direct-edit-studio-c-spec.md`](./slice7-guided-direct-edit-studio-c-spec.md)

Steps:

1. Read this plan and the complete reconstruction specification before revising the mockup.
2. Present the current Guided Direct-Edit Studio C mockup for review.
3. Collect specific requested changes and convert them into explicit visual and interaction requirements.
4. Update the mockup, this plan, the reconstruction specification, and the acceptance scenarios together so they remain consistent.
5. Verify that every preservation-critical current Heard setting is still present and that the front, back, direct-edit, autosave, proof, download, and print-service states remain coherent.
6. Obtain explicit approval and mark D0 approved.
7. Stop; do not implement product code in the same design-decision turn unless the user explicitly authorizes it.

Done when the current Studio C mockup is explicitly approved and its screenshot-ready behavior is consistent in both linked documents.

### Q1 — qurl Passage auth and authenticated session

Repository: qurl only  
Model: `gpt-5.6-terra`, reasoning `high`  
Owns: qurl auth config, middleware, routes, tests, Account flow, auth docs  
Depends on: D0 only for wording/visual direction; backend contract work may start earlier if explicitly authorized

Steps:

1. Read qurl AGENTS/Haiven docs and Passage integration docs.
2. Write failing tests for PKCE, state, callback, JWKS, wrong audience, expiry, tenant/role claims, logout, provider outage, and production bypass rejection.
3. Update qurl OpenAPI with provider-neutral auth/session routes and errors.
4. Add strict runtime configuration.
5. Implement start/callback/session/logout.
6. Replace the Account placeholder with qurl-branded states.
7. Add local Docker wiring for qurl's registered Passage callback.
8. Update qurl architecture/project-state/local-run docs.
9. Run qurl checks.

Commit boundary: `feat(auth): integrate qurl with Passage product sessions`.

### Q2 — qurl saved project after registration

Repository: qurl only  
Model: `gpt-5.6-terra`, reasoning `medium`  
Owns: qurl project OpenAPI/schema, migrations, project module, protected UI, tests  
Depends on: Q1

Steps:

1. Write BDD scenarios for anonymous draft handoff and cross-tenant denial.
2. Define project contracts and revision behavior.
3. Add PostgreSQL and versioned migrations without disturbing anonymous rendering.
4. Implement tenant-scoped project repository/service/handlers.
5. Implement bounded browser draft preservation and exactly-once import after auth.
6. Replace Library placeholders with saved-project states needed for this slice.
7. Add audit/outbox behavior and observability.
8. Run qurl checks.

Commit boundary: `feat(projects): save anonymous qurl drafts after Passage registration`.

### Q3 — qurl direct URL contract hardening

Repository: qurl only  
Model: `gpt-5.6-terra`, reasoning `medium`  
Owns: qurl QR OpenAPI/schema errors, handler contract tests, integration documentation  
Depends on: none; may run in parallel with Q1 if file ownership does not overlap

Steps:

1. Add failing tests for valid direct URL preview/export, bad schema, bounded body, response content type, and destination preservation.
2. Add documented structured error responses.
3. Keep anonymous endpoints anonymous and add production-grade rate-limit hooks.
4. Prove no redirect/tracking destination substitution.
5. Document the exact consumer request used by Heard.
6. Run qurl checks.

Commit boundary: `fix(api): harden direct URL QR contract for product consumers`.

### H1 — Heard campaign mechanic and question vertical slice

Repository: Heard only  
Model: `gpt-5.6-terra`, reasoning `medium`  
Owns: Heard OpenAPI campaign/question schemas, migrations, domain/store/server, public guest flow, tests, slice docs  
Depends on: D0

Steps:

1. Write the three mechanic BDD scenarios and domain validation tests.
2. Update OpenAPI and generate/refresh types where practical.
3. Add migrations and safe current-campaign compatibility.
4. Implement mechanic, structured offer, two/three questions, published revision, and answer validation.
5. Remove hard-coded `flyer_giveaway` authority from client metadata.
6. Update the public guest flow and recovery behavior.
7. Add permission, tenant, migration, and outage tests.
8. Update docs and run Heard checks.

Commit boundary: `feat(campaigns): add configurable flyer mechanics and questions`.

### H2 — Guided Studio shell and generated two-sided flyer

Repository: Heard only  
Model: `gpt-5.6-sol`, reasoning `high`  
Owns: selected builder UX, design-system components, generated templates, responsive/accessibility tests  
Depends on: D0 and H1 contracts

Steps:

1. Inventory and write regression tests for every current campaign control: location, restaurant name, campaign name, logo upload/remove, rating-face style, headline, prompt, incentive copy, SMS keyword/number, Google/Yelp URLs, Heard handle/path, copy link, test survey, preview, and create/update persistence behavior.
2. Write browser tests for the selected Guided Direct-Edit Studio C flow, resume, validation, front/back preview, direct item selection, and contextual editing without state loss.
3. Break the current campaign page into focused section components backed by one shared typed draft; migrate existing state rather than replacing it.
4. Implement Campaign, Survey, Offer, Responses, Brand, Design, and Link and reviews sections, preserving current labels and defaults where a domain change does not require new wording.
5. Add structured mechanics, two/three questions, response-method combinations, and brand/design controls as additive fields with explicit compatibility handling.
6. Generate a polished two-sided template from the shared draft and reflect form changes on the canvas immediately.
7. Implement bounded autosave with visible saving/saved/offline/failure states, local dirty-draft recovery, and explicit revision-conflict handling.
8. Verify 360/768/1024/1440 widths, keyboard navigation, screenshots, and existing-campaign round trips with no field loss.
9. Run web checks.

Commit boundary: `feat(builder): introduce guided two-sided flyer flow`.

### H3 — Heard atomic design contract and persistence

Repository: Heard only  
Model: `gpt-5.6-terra`, reasoning `medium`  
Owns: flyer design JSON schema, OpenAPI, migration, repository/service/handlers, autosave concurrency, tests  
Depends on: H1; contract may be prepared in parallel with H2 if files do not overlap

Steps:

1. Add valid/invalid schema fixtures and failing contract tests.
2. Define exact atomic element subtype schemas and physical bounds; reject mixed-style text, text inside image/QR/shape elements, and more than 40 elements per side.
3. Add design/revision persistence and tenant-scoped queries.
4. Implement GET/PUT with ETag/revision conflicts.
5. Implement publish snapshot atomicity and events.
6. Add audit, logs, metrics, traces, and cross-tenant tests.
7. Run Heard checks.

Commit boundary: `feat(flyers): persist versioned atomic designs`.

### H4 — Canvas editor interactions

Repository: Heard only  
Model: `gpt-5.6-sol`, reasoning `high`  
Owns: direct canvas UI components and interaction tests; must not change backend contracts without returning to H3  
Depends on: H2 and H3

Steps:

1. Write tests for direct item selection, inline/contextual copy editing, drag, resize, rotate, z-order actions, duplicate, delete, keyboard nudge, undo/redo, side switching, and autosave status.
2. Render the canonical physical-coordinate document as SVG.
3. Implement pointer and keyboard transforms without storing pixel-relative geometry.
4. Implement a contextual item editor that opens from direct canvas selection.
5. Implement exact displayed-copy fields separately from operational values such as SMS keyword, phone number, email address, and Heard survey URL.
6. Implement the fixed text/image/QR/rectangle/square/circle/squircle palette, z-order contextual actions, snapping, safe-area guides, and an optional exact position/size section.
7. Add template reset confirmation and local recovery draft.
8. Verify responsive stacked editor/canvas behavior and zoom behavior.
9. Run visual, accessibility, and web checks.

Commit boundary: `feat(flyers): add responsive direct-edit canvas studio`.

### H5 — Assets, fonts, colors, and decorative art

Repository: Heard only  
Model: `gpt-5.6-terra`, reasoning `medium`  
Owns: asset contracts/storage/validation, curated font registry, decorative primitives, editor integration, tests  
Depends on: H3; may run parallel with H4 under non-overlapping file ownership

Steps:

1. Write malicious/malformed/cross-tenant asset tests.
2. Add asset OpenAPI, storage adapter, migration, and tenant authorization.
3. Decode/re-encode allowed raster images and remove metadata.
4. Implement effective-DPI calculation.
5. Add reviewed fonts with license records.
6. Add code-owned decorative art primitives.
7. Wire assets and style controls into the typed design document.
8. Run security and Heard checks.

Commit boundary: `feat(flyers): add safe design assets and brand controls`.

### H6 — Heard/qurl live integration

Repository: Heard only  
Model: `gpt-5.6-terra`, reasoning `medium`  
Owns: Heard qurl adapter, fake, config, contract tests, campaign QR UI states, integration docs  
Depends on: Q3 contract and H3

Steps:

1. Replace the nonexistent `/api/v1/qr-codes` call with the canonical direct URL request.
2. Write a contract-accurate qurl fake and failure tests.
3. Consume preview SVG only as an isolated image source.
4. Request binary SVG for export.
5. Add timeout, retry classification, metrics, tracing, and safe logs.
6. Add the cross-repo destination-decoding acceptance test/script.
7. Run Heard and qurl contract checks.

Commit boundary: `fix(qr): integrate Heard with qurl direct URL API`.

### H7 — SMS and email campaign entry

Repository: Heard only  
Model: `gpt-5.6-terra`, reasoning `high`  
Owns: response-method contracts, provider interfaces/fakes, inbound adapters, verification, consent/opt-out behavior, tests  
Depends on: H1

Steps:

1. Define response-method and webhook contracts before provider code.
2. Write keyword/address uniqueness, signature, replay, STOP/HELP, and provider-outage tests.
3. Implement provider-neutral inbound SMS/email interfaces and local fakes.
4. Implement verification state and publication gates.
5. Map inbound keyword/address to an opaque campaign link and return the survey URL.
6. Keep transactional and marketing messages separate.
7. Add audit and delivery observability.
8. Run Heard checks.

Commit boundary: `feat(channels): add verified SMS and email flyer entry`.

### H8 — Proof, render, and print-file export

Repository: Heard only  
Model: `gpt-5.6-terra`, reasoning `high`  
Owns: renderer interface/adapter, proof service, export records/worker, artifact storage, tests  
Depends on: H3, H5, and H6

Steps:

1. Write physical-unit, bounds, overflow, QR, DPI, font, and export job tests.
2. Complete the dependency/supply-chain review and document the renderer decision.
3. Implement deterministic front/back SVG output.
4. Implement 300 DPI PNG and exact-size two-page PDF.
5. Implement proof errors/warnings and acknowledgement.
6. Implement durable idempotent export jobs and artifact authorization.
7. Make `Preview & proof` an onscreen state and `Download print files` the only file-download action.
8. Rasterize and visually verify exports in CI/local checks.
9. Run Heard checks.

Commit boundary: `feat(exports): proof and render print-ready flyer packages`.

### H9 — Print service quote and order vertical slice

Repository: Heard only  
Model: `gpt-5.6-terra`, reasoning `high`  
Owns: print-provider contract/adapter, quote/order persistence, checkout UI, provider webhooks, tests, operator docs  
Depends on: H8

Steps:

1. Compare currently viable print-provider APIs against the exact 3-by-4 duplex product, paper/finish options, shipping, quote, order, webhook, sandbox, and credential requirements; record the provider decision before code.
2. Define provider-neutral quote, order, status, structured-error, idempotency, and webhook contracts in OpenAPI before implementation.
3. Write tests for quote expiry, immutable export revision, explicit confirmation, duplicate submission, provider timeout, payment rejection, webhook signature/replay, cancelability, and cross-tenant denial.
4. Add print quote/order/status persistence and the provider adapter without putting provider fields into campaign domain logic.
5. Implement quantity, supported stock/finish, shipping, quote summary, proof acknowledgement, and final confirmation UI.
6. Submit only the exact proofed export revision with an idempotency key; autosave must never submit or charge an order.
7. Implement authenticated replay-safe provider status webhooks and useful operator status/history.
8. Add a local fake print provider and document sandbox/local operation.
9. Run Heard checks.

Commit boundary: `feat(print): quote and submit proofed flyer orders`.

### R1 — Cross-cutting security and architecture review

Repositories: read-only review of Heard and qurl  
Model: `gpt-5.6-sol`, reasoning `xhigh`  
Writes: review findings only; implementation agents fix findings in their owned repos  
Depends on: Q1-Q3 and H1-H8 complete

Review:

- Passage token verification and cookie safety;
- tenant/role/permission enforcement;
- anonymous qurl versus protected project boundaries;
- asset and renderer attack surface;
- webhook signatures/replay;
- qurl and print-provider data minimization;
- dependency and CI supply-chain changes;
- secrets/logging;
- contract compatibility and failure behavior.

All P0/P1 findings block completion. P2 findings require explicit disposition.

### V1 — Final acceptance and documentation

Repositories: Heard and qurl, each changed only in its own directory  
Model: `gpt-5.6-terra`, reasoning `medium`  
Owns: acceptance scripts, final docs/status, no feature redesign  
Depends on: R1 findings resolved

Steps:

1. Run all unit, integration, contract, browser, accessibility, and export tests.
2. Run the full local Passage + qurl + Heard journey.
3. Test the three campaign mechanic examples.
4. Test QR, SMS, and email entry combinations.
5. Decode QR from the final PDF/PNG.
6. Run `pnpm haiven:check` in qurl.
7. Run `pnpm haiven:check` in Heard.
8. Update README, architecture, OpenAPI examples, local-run instructions, operator docs, and this slice status.
9. Confirm separate, reviewable qurl and Heard commits.

## 15. Agent orchestration and model assignment

The main coordinator should use `gpt-5.6-terra` at `medium` reasoning. It owns the plan, dependency graph, task dispatch, status, integration decisions, and final checks. It should not take over a subagent's files while that agent is active.

Use no more than three subagents at once so the fourth concurrency slot remains with the coordinator.

### Wave 0 — design gate

- `design_gate`: `gpt-5.6-sol`, `high` — D0 only.

Stop after Wave 0 until the user approves the flow.

### Wave 1 — independent foundations

- `qurl_identity`: `gpt-5.6-terra`, `high` — Q1.
- `qurl_direct_contract`: `gpt-5.6-terra`, `medium` — Q3; coordinate OpenAPI file ownership with Q1.
- `heard_campaign_domain`: `gpt-5.6-terra`, `medium` — H1.

If Q1 and Q3 both need the same qurl OpenAPI file, run them sequentially. File ownership beats parallel speed.

### Wave 2 — usable first flyer flow

- `qurl_projects`: `gpt-5.6-terra`, `medium` — Q2.
- `heard_builder_experience`: `gpt-5.6-sol`, `high` — H2.
- `heard_design_contract`: `gpt-5.6-terra`, `medium` — H3.

H2 consumes H1's contract. H3 and H2 may overlap only after schema shapes are frozen and their file ownership is disjoint.

### Wave 3 — advanced composition and channels

- `heard_canvas`: `gpt-5.6-sol`, `high` — H4.
- `heard_assets`: `gpt-5.6-terra`, `medium` — H5.
- `heard_channels`: `gpt-5.6-terra`, `high` — H7.

### Wave 4 — integration and output

- `heard_qurl_adapter`: `gpt-5.6-terra`, `medium` — H6.
- `heard_export_renderer`: `gpt-5.6-terra`, `high` — H8.
- `heard_print_service`: `gpt-5.6-terra`, `high` — H9 after H8 fixes the immutable export contract.

### Wave 5 — independent review and acceptance

- `security_architecture_review`: `gpt-5.6-sol`, `xhigh` — R1, read-only.
- `acceptance_documentation`: `gpt-5.6-terra`, `medium` — V1 after findings are fixed.

Subagent dispatch template:

```text
Read the target repository AGENTS.md and every required Haiven document before acting.
Task: <task ID and exact task card text>.
Repository boundary: <heard or qurl only>.
Owned files/modules: <explicit list>.
Do not edit files owned by another active agent.
Preserve all existing uncommitted changes; never reset or checkout them away.
Work contract-first and test-first.
Implement one complete vertical slice; do not leave TODO scaffolding.
Run the task-specific checks and the repository Haiven check.
Return: files changed, contracts changed, tests run, failures/risks, commit-ready summary.
```

## 16. Dependency graph

```text
D0
├── Q1 ── Q2
├── Q3 ─────────────── H6 ──┐
└── H1 ── H2 ── H4 ────────┤
         └── H3 ── H5 ─────┼── H8 ── H9
              └── H7       │
                            └── R1 ── V1
```

H8 also requires H3, H5, and H6. H4 is required for the selected direct-edit studio experience but does not own export rendering.

## 17. Recommended commit sequence

qurl repository:

1. `fix(api): harden direct URL QR contract for product consumers`
2. `feat(auth): integrate qurl with Passage product sessions`
3. `feat(projects): save anonymous qurl drafts after Passage registration`
4. `test(integration): prove Passage and project tenant boundaries`

Heard repository:

1. `docs(flyers): record approved Guided Studio design`
2. `feat(campaigns): add configurable flyer mechanics and questions`
3. `feat(builder): introduce guided two-sided flyer flow`
4. `feat(flyers): persist versioned atomic designs`
5. `feat(flyers): add responsive atomic direct-edit studio`
6. `feat(flyers): add safe design assets and brand controls`
7. `fix(qr): integrate Heard with qurl direct URL API`
8. `feat(channels): add verified SMS and email flyer entry`
9. `feat(exports): proof and render print-ready flyer packages`
10. `feat(print): quote and submit proofed flyer orders`
11. `test(flyers): cover end-to-end design, print, and provider failures`

Do not combine qurl and Heard changes in one commit or copy code across repository boundaries.

## 18. Definition of done

- D0 is explicitly approved and recorded.
- Every current campaign-builder field and behavior remains available, editable, and round-trips through an existing campaign without data loss.
- Guided settings and direct canvas edits update one shared draft; selecting items, switching sides, or changing sections never resets or forks campaign state.
- A restaurant operator can create feedback-only, instant-offer, or prize-drawing campaigns.
- The operator can configure exactly two or three questions.
- QR, verified SMS, and verified email methods work in any enabled combination.
- The first 3-by-4-inch format supports directly editable front and back items backed by a z-ordered atomic element model.
- The operator can edit copy, fonts, colors, logo, offer/prize image, QR/contact block, and decorative corner art.
- Every differently styled text area is a separate atomic text element with one typography record; image, QR, and shape elements expose only type-correct controls.
- Templates pre-populate atomic placeholders and the operator can add text, image, QR, rectangle, square, circle, or squircle elements directly on the flyer.
- Templates are editable starting points: the operator can rewrite generated wording, drag items, resize text/images/shapes, and use contextual font, color, alignment, position, size, and z-order controls.
- The editor is polished and responsive, with autosave, recovery, undo/redo, contextual properties, and useful failures.
- qurl renders the actual QR through its documented direct URL API; Heard has no local QR engine.
- qurl has a qurl-branded Passage registration/session flow and useful saved-project result while anonymous QR creation remains available.
- Print proof catches blocking errors and useful warnings.
- SVG, 300 DPI PNG, and exact-size two-page PDF exports pass automated dimension and QR-decode checks.
- Preview is onscreen proofing, while Download print files is the file-export action.
- Every draft change autosaves with visible success/offline/failure states, and autosave can never submit a print order.
- Send to print service quotes and submits only the explicitly confirmed, proofed, immutable export revision and safely tracks provider status.
- Passage, qurl, messaging, rendering, and print-provider failures are safe and observable.
- Tenant, role, permission, asset, webhook, and cross-tenant tests pass.
- Contracts, migrations, generated types, events, docs, and local development are current.
- qurl and Heard each pass their own repository checks and `haiven check`.
