# Open Restaurant Guest Experience Platform

## Heard Engineering Constitution

This section defines the non-negotiable engineering, architecture, quality, scalability, maintainability, and contributor principles for Heard.

These are not preferences.
They are foundational platform requirements.

Every epic, story, API, schema, service, provider, UI flow, integration, and AI-generated contribution must follow these principles.

---

# Core Engineering Principles

## 0. Foundational Architectural Enforcement

The following implementation standards are mandatory platform-wide enforcement mechanisms.

### Multi-Tenancy Enforcement Strategy

Tenant isolation must be enforced at multiple layers.

Preferred strategy:
- PostgreSQL Row-Level Security (RLS) for tenant-scoped tables.
- Application-level tenant context enforcement.
- Repository/query-layer tenant injection.
- Service-level authorization checks.

Requirements:
- Tenant-scoped tables must include tenant identifiers.
- RLS policies must be enabled for protected production tables where practical.
- Application queries must include explicit tenant context.
- Unsafe raw queries are prohibited in production code paths.
- Background workers and async jobs must preserve tenant context.
- Cross-tenant queries require explicit privileged service behavior.

Definition of done:
- Tenant isolation tests included.
- RLS policy coverage documented.
- Unsafe queries blocked by linting/review standards.

### Schema Migration Strategy

Schema evolution must be explicit, versioned, testable, and safe for self-hosted deployments.

Preferred tooling:
- Atlas or golang-migrate.

Requirements:
- All schema changes use versioned migrations.
- Forward and rollback behavior must be documented.
- Migrations must be deterministic.
- Destructive migrations require explicit approval.
- Seed data migrations must be isolated from structural migrations.
- CI must validate schema migrations against clean databases.
- Schema drift detection should be automated where possible.

Definition of done:
- Migration added.
- Migration validated in CI.
- Rollback behavior documented.

### Canonical Identity and UUID Strategy

Core platform identifiers must use UUIDs unless there is a deliberate, documented exception.

Requirements:
- Tenant IDs must be UUIDs.
- Internal entity IDs must be UUIDs.
- External provider IDs must be stored separately from internal IDs.
- External IDs must never be treated as globally unique without source-system and tenant/location context.
- Database primary keys should use UUIDs for tenant-scoped domain entities.
- Public URLs should use opaque tokens or slugs instead of exposing raw UUIDs when appropriate.

Definition of done:
- Internal ID fields use UUIDs.
- External ID mappings are stored separately.
- ID semantics documented in schemas.

### Experience Identity and Attribution

Heard must support a minimal canonical Experience entity representing a uniquely identifiable guest interaction at a specific restaurant location.

The Experience entity is the durable internal bridge between feedback, recovery, reviews, offers, refunds, analytics, and integrations.

Core identity fields:
- id: Heard-owned UUID
- tenant_id: UUID
- location_id: UUID
- source_system: external system name such as toast, square, olo, manual, qr, or imported
- source_entity_id: unique identifier from the source system

Rules:
- The identity tuple is tenant_id + location_id + source_system + source_entity_id.
- Heard should not require order details, table, server, subtotal, items, fulfillment type, or guest identity to create an Experience.
- Additional provider data must be treated as optional enrichment.
- Enrichment must not be required for feedback submission, recovery case creation, or core reporting.
- Enrichment should be additive, nullable, and provider-specific where needed.
- Integrations should prefer stable provider identifiers over visible receipt numbers when available.

Definition of done:
- Experience schema uses UUID internal identity.
- Source-system mapping is explicit.
- Optional enrichment is separated from core identity.
- Feedback sessions can attach to an Experience when available.
- Feedback sessions can still operate without an Experience when only generic QR/link context exists.

### Transactional Outbox and Event Consistency

Heard must guarantee consistency between durable database state changes and emitted domain events.

Preferred pattern:
- Transactional Outbox.

Requirements:
- Domain events are persisted in the same transaction as the originating state change.
- Background workers publish queued outbox events to NATS.
- Event publication must support retries.
- Event publication failures must be observable.
- Event consumers must support idempotency.
- Outbox replay tooling must exist.

Definition of done:
- State change and outbox insert occur atomically.
- Event publication observable.
- Replay and retry behavior tested.

---

## 0A. Architecture Shape: Modular Monolith First

Heard should start as a modular Go monolith, not as premature microservices.

The codebase must still use clear domain boundaries, contracts, events, providers, and package separation, but the first production system should deploy as a small number of processes:

- API server
- background worker
- web/admin app
- database
- optional local provider simulators

Rationale:
- Open-source contributors need a simple local development experience.
- Early deployments should be easy to run, debug, and self-host.
- Domain boundaries can be preserved without network boundaries.
- Services can be extracted later only when operational pressure justifies it.

Requirements:
- Domain packages must avoid circular dependencies.
- Cross-domain behavior should use interfaces, events, or application-layer orchestration.
- Database boundaries should be explicit even inside the monolith.
- Internal package contracts should be documented.
- No feature should require Kubernetes, service mesh, distributed tracing infrastructure, or paid cloud services to run locally.

Definition of done:
- Local development runs with one backend process and one worker process.
- Module boundaries are documented.
- Future service extraction points are clear but not prematurely implemented.

---

## 1. API-First

Every major capability in Heard must be accessible through stable APIs before it is tightly coupled to UI implementations.

The admin UI, mobile apps, integrations, automations, and future AI agents should all operate on the same APIs.

Requirements:
- Public APIs use OpenAPI specifications.
- OpenAPI contracts are the source of truth.
- Code generation is required where practical.
- Generated types, validation, clients, and server stubs should be derived from contracts.
- AI agents must modify contracts before implementation code when changing APIs.
- Internal service APIs use explicit contracts.
- APIs must support versioning.
- APIs must support pagination where applicable.
- APIs must support idempotency where appropriate.
- APIs must return structured machine-readable errors.
- APIs must enforce tenant isolation and permissions.
- APIs must never expose unsafe internal implementation details.

Definition of done:
- OpenAPI spec updated.
- Generated code refreshed.
- Contract compatibility validated.
- Contract tests included.
- API examples documented.
- Authentication and authorization enforced.
- Failure modes documented.

---

## 2. Contract-First

Canonical schemas and contracts must exist before implementation begins.

Schemas are the backbone of:
- APIs
- events
- persistence
- provider integrations
- exports
- webhooks
- AI workflows
- analytics
- OSS extensions

Requirements:
- All major entities use explicit versioned schemas.
- Schemas follow JSON Schema standards.
- Breaking changes require schema version changes.
- Shared contracts must live in a dedicated contracts package.
- Arbitrary metadata must be intentionally scoped.
- Event payloads must be documented.

Canonical entities include:
- tenant
- brand
- location
- user
- role
- permission
- guest
- feedback-session
- feedback-response
- survey
- survey-version
- survey-question
- recovery-case
- recovery-message
- offer
- webhook-subscription
- audit-entry
- provider-config
- integration-account
- notification-event

Definition of done:
- Schema added or updated.
- Schema version documented.
- Validation tests included.
- Migration strategy documented where needed.

---

## 3. Test-First Development

Heard should be developed with tests as a first-class deliverable.

AI agents and contributors must not treat tests as optional cleanup work.

Requirements:
- Unit tests required for domain logic.
- Integration tests required for APIs.
- Contract tests required for providers.
- Permission tests required for protected workflows.
- Multi-tenant isolation tests required.
- Event emission tests required.
- Failure-mode tests required.
- Regression tests required for bug fixes.

Critical guest flows must be covered end-to-end.

Definition of done:
- All required tests pass.
- Coverage meets module expectations.
- CI validates tests automatically.
- New functionality cannot merge without tests.

---

## 4. Behavior-Driven Development

BDD scenarios should primarily drive integration and end-to-end workflow validation.

Unit tests should continue handling deterministic domain logic, validation, transformation, utility behavior, and isolated business rules.

BDD should focus on:
- multi-step workflows
- permission-sensitive operations
- recovery flows
- event-driven orchestration
- guest interaction flows
- automation behavior
- provider integration behavior

Avoid using heavyweight BDD flows for simple deterministic utility logic.

Product behavior should drive implementation.

Stories should be translatable into executable behavioral scenarios.

Preferred structure:
- Given
- When
- Then

Example:

Given a negative feedback submission
When the guest submits feedback
Then a recovery case is automatically created
And the case is visible to authorized managers
And an event is emitted
And an audit entry is written

Requirements:
- High-value workflows include behavioral scenarios.
- Acceptance criteria must be testable.
- Behavioral tests should cover business-critical flows.

---

## 5. Event-Driven Architecture

Heard should operate as an event-driven platform.

Core modules must communicate through explicit events instead of hidden internal coupling.

Core events include:
- feedback-submitted
- feedback-updated
- recovery-case-created
- recovery-case-assigned
- recovery-message-sent
- recovery-message-received
- offer-issued
- offer-redeemed
- review-imported
- SLA-breached
- webhook-delivery-failed

Requirements:
- Events must be versioned.
- Events must be idempotent or safely de-duplicated.
- Events must support retries.
- Event contracts must be documented.
- Failed event delivery must be observable.

Preferred infrastructure:
- Internal domain event dispatch plus transactional outbox initially.
- NATS adapter when async scale or deployment topology justifies it.
- Kafka-compatible future path only if enterprise scale requires it.

Definition of done:
- Event contract documented.
- Event emitted.
- Event tests included.
- Retry behavior validated.

---

## 6. Multi-Tenant by Default

Every major system component must be tenant-aware from the beginning.

Tenant isolation cannot be retrofitted later.

Requirements:
- All tenant-scoped records include tenant identifiers.
- Authorization checks enforce tenant boundaries.
- Cross-tenant access is prohibited by default.
- Queries must be tenant-scoped.
- Exports must be tenant-scoped.
- Audit logs must preserve tenant context.

Definition of done:
- Tenant isolation tests included.
- Permission tests included.
- Unsafe queries prohibited.

---

## 7. Scalability and Performance

Heard should be designed to scale from a single-location restaurant to enterprise multi-brand deployments.

Requirements:
- Stateless application services.
- Horizontally scalable workers.
- Async processing where appropriate.
- Queue-backed provider delivery.
- Efficient indexing strategies.
- Pagination for large result sets.
- Rate limiting.
- Idempotent APIs.
- Background job retries.
- Bulk operation safeguards.

Preferred infrastructure:
- Go services
- PostgreSQL
- Redis
- NATS
- Kubernetes-ready deployment

Definition of done:
- Performance considerations documented.
- Expensive queries identified.
- Backpressure handling defined.
- Timeouts and retries configured.

---

## 8. Reliability and Failure Recovery

The platform must fail gracefully.

Restaurants should not lose guest feedback because a provider is temporarily unavailable.

Requirements:
- Transactional outbox pattern where appropriate.
- Retry with exponential backoff.
- Dead-letter handling.
- Graceful provider degradation.
- Circuit breakers for unstable providers.
- Safe retry semantics.
- Idempotent event handling.
- Recovery jobs for failed deliveries.

Definition of done:
- Failure modes documented.
- Retry behavior tested.
- Recovery path validated.
- Provider outages handled safely.

---

## 9. Maintainability and Readability

The codebase must remain understandable by humans and AI agents.

Requirements:
- Clear package boundaries.
- Small focused functions.
- Explicit naming.
- Minimal framework magic.
- Limited abstraction layers.
- Comments explain why, not what.
- Shared utilities must justify existence.
- Technical debt shortcuts must be explicitly documented.

Definition of done:
- Module responsibilities documented.
- Public interfaces documented.
- Complex logic explained.
- Dead code avoided.

---

## 10. Observability by Default

Every critical workflow must be observable.

Requirements:
- Structured logs.
- OpenTelemetry tracing.
- Correlation IDs.
- Metrics for critical workflows.
- Health checks.
- Readiness checks.
- Provider delivery logs.
- Audit logs.
- Event processing metrics.
- Queue health metrics.
- SLA monitoring.

Definition of done:
- Logs added.
- Metrics added.
- Traces emitted.
- Dashboards updated where applicable.

---

## 10A. Unified Responsive Experience Standard

Heard must ship one premium responsive interface system across guest-facing and admin-facing surfaces instead of maintaining separate device-specific variants.

Requirements:
- UI work must be mobile-first and scale cleanly to tablet, laptop, and desktop layouts.
- Separate iOS, Android, tablet, or desktop variants are discouraged unless a platform constraint makes them unavoidable.
- Shared layout primitives, spacing rules, typography, and component patterns should support all supported screen sizes.
- Guest and admin experiences must both meet a high visual bar and feel like the same intentional product family.
- Interfaces must be screenshot-ready from the first usable slice.
- Responsive behavior must be considered part of the feature, not deferred polish.
- Admin and guest production UX should not be conflated with internal test harness UX; dev/test tools must live in clearly labeled internal testing surfaces.

Definition of done:
- The feature works on phone, tablet, and desktop breakpoints.
- The interface feels polished, not merely functional.
- No device class requires a separate bespoke implementation for the same workflow.

---

## 11. Security and Privacy

Guest trust is critical.

Requirements:
- RBAC enforcement.
- Principle of least privilege.
- Secure secret handling.
- Signed webhooks.
- PII masking.
- Secure defaults.
- Rate limiting.
- Abuse protection.
- Sensitive data exclusion from logs.
- Audit logging for sensitive actions.

Definition of done:
- Permission checks tested.
- Sensitive fields protected.
- Security-sensitive flows audited.

---

## 11A. Admin Authentication Baseline

Admin surfaces must be protected by authentication.

Requirements:
- No production admin surface should be publicly accessible without login.
- Initial auth should support third-party identity providers.
- Priority providers for initial implementation are Google, Apple, and Facebook.
- Additional providers should be pluggable through a provider abstraction.
- Role and tenant scoping must be enforced after authentication.
- Local development may use a clearly labeled auth bypass only in non-production environments.

Definition of done:
- Admin access requires authenticated identity in production mode.
- Google, Apple, and Facebook provider paths are implemented or explicitly scaffolded with integration contracts.
- Role and tenant checks are enforced for authenticated admin requests.

---

## 12. Compliance-Aware Communication

Messaging systems must support compliant restaurant communication.

Requirements:
- Transactional and marketing messaging separated.
- SMS opt-outs enforced.
- Consent records stored.
- Review-gating avoided.
- Guest communication history auditable.
- Privacy retention configurable.

Definition of done:
- Consent captured.
- Opt-outs enforced.
- Communication audit entries recorded.

---

## 12A. qurl as the QR Generation System

Heard must use qurl for QR code generation.

qurl is the QR generation product in the Haiven Labs ecosystem. Heard should not implement its own QR rendering engine unless a temporary local fake is required for tests or offline development.

Responsibilities:
- Heard owns restaurant context, tenant context, feedback links, opaque tokens, permissions, analytics, and feedback workflows.
- qurl owns QR code rendering, QR export formats, QR styling, and QR asset generation.

Integration rule:
- Heard generates the feedback destination URL.
- Heard sends that destination URL to qurl.
- qurl returns QR assets or downloadable QR representations.
- Heard stores the qurl asset reference and associates it with the FeedbackLink.

Security rule:
- Heard should not send unnecessary restaurant, guest, tenant, or operational metadata to qurl.
- qurl should only need the destination URL, rendering options, and export format.
- The destination URL should use Heard's opaque tokenized feedback link.

Codex instruction:
- When implementing QR generation in Heard, use a qurl client/provider interface.
- If qurl does not yet expose the API Heard needs, Codex should update qurl or define the required qurl API contract before implementing the Heard integration.
- Do not add a separate QR rendering library directly into Heard core unless it is a local fake provider used only for development or tests.

Definition of done:
- Heard has a qurl QR provider interface.
- Local development has a fake qurl provider.
- Production configuration can point to a real qurl API.
- QR generation tests verify that Heard sends only the destination URL and rendering options.
- QR assets remain tied to Heard FeedbackLink records.

---

## 13. Provider Abstraction

Third-party providers must be replaceable.

Requirements:
- SMS providers use interfaces.
- Email providers use interfaces.
- AI providers use interfaces.
- Storage providers use interfaces.
- Translation providers use interfaces.
- POS integrations use adapters.
- Fake local providers exist for development.

Definition of done:
- Provider interface documented.
- Fake provider included.
- Contract tests included.

---

## 14. Open-Source Credibility

Heard must operate as a legitimate open-source platform.

Requirements:
- Local development works without paid vendors.
- Seed data included.
- Contributor docs included.
- CI works locally.
- Docker-first development.
- Self-hosting supported.
- Clear extension points.
- Stable contracts.

Definition of done:
- Local startup documented.
- Fake providers functional.
- Seed data validated.
- Contributor workflow documented.

---

## 15. Vertical Slice Delivery

Features should be built as complete end-to-end vertical slices.

Avoid:
- building all database tables first
- building all APIs first
- building all UIs first
- giant incomplete architectural layers

Preferred approach:
- build one complete capability end-to-end
- validate operational behavior
- validate observability
- validate permissions
- validate scalability assumptions

Example MVP slice:
- create QR code
- submit feedback
- create recovery case
- manager replies by SMS
- audit log created
- event emitted
- metrics recorded

Definition of done for every slice:
- API implemented
- contract documented
- tests passing
- observability added
- permissions enforced
- events emitted
- local development supported
- docs updated

---

## 16. AI-Agent Contribution Standards

Heard should be intentionally designed for agentic software development.

Requirements:
- Every task includes clear scope.
- Every task includes acceptance criteria.
- Every task includes API expectations.
- Every task includes testing expectations.
- Every task includes observability requirements.
- Every task includes permission expectations.
- Every task includes failure-mode handling.
- Generated code must remain readable.
- Generated code must not bypass contracts.
- New user directives that change product, UX, architecture, quality, or implementation expectations must be written back into the constitution, implementation plan, or active slice docs instead of living only in chat context.
- When user feedback identifies UX confusion or workflow friction, the plan should be updated with concrete interaction requirements before subsequent implementation continues.

Preferred task structure:
- Goal
- User story
- API contract
- Data model
- Events emitted
- Permissions
- Failure modes
- Observability
- Tests required
- Definition of done

---

## Preferred Technical Stack

Backend:
- Go
- PostgreSQL
- Redis only when caching, rate limiting, or ephemeral state requires it
- Internal domain events plus transactional outbox first
- NATS adapter later when deployment scale justifies it
- OpenTelemetry
- Docker-first local development
- Kubernetes-compatible, but not Kubernetes-required

Frontend:
- React
- Next.js
- TypeScript
- Tailwind

Mobile:
- Expo React Native

API:
- REST externally
- OpenAPI contracts
- Avoid gRPC until a domain is extracted into a separately deployed service

Testing:
- Unit tests
- Integration tests
- Contract tests
- BDD-style scenario tests

Infrastructure:
- Docker-first local development
- CI/CD from day one
- GitHub Actions initially

---

## Engineering Enforcement Strategy

These principles are enforced through:
- pull request templates
- contributor guidelines
- CI gates
- required tests
- schema validation
- contract validation
- linting
- code generation standards
- architecture reviews
- automated observability checks where possible
- required Definition of Done checklists

---

## Definition of Done (Global)

Every production-ready capability in Heard must include:

- Working implementation
- API contract
- Schema validation
- Tests
- Permission enforcement
- Tenant isolation
- Event emission where applicable
- Logging
- Metrics
- Tracing
- Audit coverage where applicable
- Documentation
- Local development support
- Failure-mode handling
- CI validation

No feature is complete without these.

---

# Open Restaurant Guest Experience Platform

## Build Kickoff Plan

Codex should use this document as the source of truth for the first implementation wave.

The goal is not to build every feature at once. The goal is to create a clean, working, end-to-end foundation that proves the product loop:

1. A restaurant can be created.
2. A location can be created.
3. A feedback link or QR token can be generated.
4. A guest can open a hosted feedback form.
5. A guest can submit feedback.
6. Heard stores the feedback.
7. Heard emits a feedback-submitted event through the transactional outbox.
8. Negative feedback creates a recovery case.
9. A manager can see the case in a recovery inbox.
10. The whole thing is observable, tested, tenant-safe, and beautiful.

This first implementation wave should prioritize one complete vertical slice over broad incomplete scaffolding.

---

# Slice 1: Frictionless Feedback to Recovery Case

## Slice Goal

Build the first usable Heard product loop: create a restaurant/location, generate a secure feedback link, collect guest feedback, and automatically create a recovery case when the feedback is negative.

This slice should feel real. It should not look like a developer demo. The guest-facing form and manager inbox must look sexy as hell: clean, modern, fast, mobile-first, polished, and trustworthy.

## Slice Scope

Slice 1 includes:

- Modular Go backend foundation.
- PostgreSQL schema and migrations.
- Local Docker-based development environment.
- OpenAPI contract for the first APIs.
- Generated API types and handlers where practical.
- Tenant and location creation.
- Opaque feedback link token generation.
- QR code generation through qurl.
- Hosted guest feedback form.
- Feedback session creation.
- Feedback response submission.
- Negative sentiment/rating classification using deterministic rules.
- Transactional outbox event for feedback-submitted.
- Automatic recovery case creation for negative feedback.
- Manager recovery inbox.
- Basic admin web shell.
- Audit entries for key actions.
- Structured logs, traces, and basic metrics hooks.
- Tests for contracts, domain behavior, tenant isolation, permissions, persistence, and the end-to-end slice.

Slice 1 does not include:

- POS integrations.
- Automated post-order triggers.
- Advanced survey builder.
- Public review management.
- Listings management.
- Inbound SMS routing.
- AI-generated replies.
- Offer redemption.
- Advanced analytics.
- Multi-brand enterprise hierarchy beyond tenant and location basics.
- Table, server, order-item, subtotal, or floor-plan enrichment.

## Slice 1 Domain Entities

Codex should start with only the domain entities required for the slice.

Required entities:

- Tenant
- Location
- User
- FeedbackLink
- FeedbackSession
- FeedbackResponse
- Experience
- RecoveryCase
- AuditEntry
- OutboxEvent

The Experience entity must remain minimal:

- id: UUID
- tenant_id: UUID
- location_id: UUID
- source_system: string
- source_entity_id: string

Feedback must be able to work without an Experience when only a generic QR or feedback link exists.

## Slice 1 qurl Integration

Heard must use qurl for Slice 1 QR generation.

Slice 1 should include a qurl provider interface with two implementations:

- fake/local qurl provider for tests and local development
- real qurl API client when qurl API configuration is present

Minimum qurl capability needed:

- Accept a destination URL.
- Accept basic rendering options.
- Return a QR asset URL, binary payload, SVG string, or asset reference.

Heard must not pass raw tenant IDs, location IDs, guest data, order IDs, or restaurant operational context to qurl. Heard passes the opaque feedback URL only.

If qurl does not currently support the required API, Codex should create or update the qurl API contract first, then wire Heard to it.

Do not implement QR rendering directly inside Heard except for a fake test provider.

---

## Slice 1 API Surface

Codex should define the OpenAPI contract before implementation.

Required APIs:

- Create tenant
- Get tenant
- Create location
- Get location
- Create feedback link
- Generate QR for feedback link through qurl
- Resolve feedback link token
- Create feedback session
- Submit feedback response
- List feedback responses
- Get feedback response
- List recovery cases
- Get recovery case
- Update recovery case status

APIs must include:

- UUID internal IDs.
- Tenant context.
- Structured errors.
- Pagination where listing is supported.
- Idempotency where duplicate creation is plausible.
- Permission enforcement hooks, even if the first implementation uses a simple local auth stub.

## Slice 1 Events

Required event:

- feedback-submitted

The feedback-submitted event must be written through the transactional outbox in the same database transaction as the feedback response.

The event payload should include:

- event_id
- event_type
- event_version
- tenant_id
- location_id
- feedback_response_id
- feedback_session_id
- feedback_link_id when available
- experience_id when available
- sentiment
- rating
- occurred_at

Recovery case creation may initially be performed by a local worker that consumes pending outbox events.

## Slice 1 Rules

Negative feedback creates a recovery case.

Initial deterministic rule:

- If the rating maps to negative, create a recovery case.
- If the feedback form uses a 1 to 5 scale, ratings 1 and 2 are negative by default.
- If the feedback form uses positive, neutral, negative options, negative creates a recovery case.

The rule must be explicit and testable. Do not use AI for sentiment in Slice 1.

## Slice 1 UI Requirements

The UI must look polished from the first slice.

Guest-facing feedback form:

- Mobile-first.
- Fast-loading.
- Beautiful spacing and typography.
- Clear rating interaction.
- Frictionless open-text input.
- Optional contact capture.
- Clear submit state.
- Friendly confirmation page.
- Accessible contrast and form labels.
- No clutter.
- No default ugly form controls.
- Must adapt beautifully across phones, tablets, laptops, and desktops without a separate device-specific build.
- Form controls should use appropriate control types: selections must use dropdowns/pickers rather than raw free-text IDs where avoidable.
- Human-readable labels should explain internal fields clearly and avoid ambiguous naming.

Manager experience:

- Clean admin shell.
- Recovery inbox with useful empty, loading, error, and populated states.
- Case rows should show status, rating, sentiment, location, guest/contact state, created time, and short feedback preview.
- Feedback detail should be easy to scan.
- The visual standard should feel like a serious modern SaaS product, not an internal admin panel.
- Must use the same responsive product language as guest flows and stay strong on tablet and smaller laptop widths.

Design direction:

- Simple.
- Premium.
- Calm.
- Restaurant-friendly.
- Trustworthy.
- Fast.
- Sharp enough for public screenshots.
- Beautiful as fuck across form factors.

Do not ship a developer-looking interface.

## Slice 1 Testing Requirements

Codex must include tests with the implementation.

Required tests:

- Tenant creation test.
- Location creation test.
- Feedback link token creation test.
- qurl provider integration test using fake provider.
- Feedback link token resolution test.
- Feedback session creation test.
- Feedback submission validation test.
- Feedback-submitted outbox event test.
- Negative feedback creates recovery case test.
- Neutral or positive feedback does not create recovery case test.
- Tenant isolation test for feedback reads.
- Tenant isolation test for recovery case reads.
- Recovery case status transition test.
- API contract test where practical.

## Slice 1 Definition of Done

Slice 1 is complete only when:

- Local development starts cleanly from documented commands.
- Database migrations run from empty database.
- OpenAPI contract exists.
- Generated API code is refreshed where practical.
- Backend APIs work.
- Guest feedback form works.
- Feedback QR generation goes through the qurl provider interface.
- Manager recovery inbox works.
- Negative feedback creates a recovery case.
- Outbox event is written transactionally.
- Audit entries are written for key actions.
- Tests pass.
- Tenant isolation is tested.
- Structured logs exist for critical workflows.
- Basic traces or trace hooks exist.
- README or docs explain how to run the slice locally.
- UI looks sexy as hell and is ready for screenshots.

---

# Slice 2: Printed Flyer Giveaway Survey Campaign

## Slice Goal

Build the first product feature around the restaurant takeout flyer workflow.

A Heard customer should be able to create a QR-backed survey campaign for printed flyers placed in takeout bags, delivery bags, catering boxes, receipts, or counter handouts without any POS integration or purchase context.

The flyer can offer an incentive such as:
- Complete this survey for a chance to win a $100 gift card.
- Text WIN to the configured phone number.
- Scan the QR code to answer the survey.

## Slice 2 Workflow

Restaurant admin flow:
- Select a restaurant location.
- Configure restaurant display name.
- Configure flyer headline, survey prompt, incentive copy, SMS keyword, SMS phone number, Google review URL, and Yelp review URL.
- Generate a Heard feedback link with an opaque token.
- Generate a QR asset through the qurl provider.
- Preview the printed flyer experience.

Guest flow:
- Guest receives a printed flyer with no known customer, purchase, order, POS, table, staff, or receipt metadata.
- Guest scans the QR code or texts the configured keyword.
- Guest lands on the Heard-hosted survey.
- Guest answers the first question by selecting one of five faces, mapping to a 1 to 5 rating.
- Ratings below 5 collect open-text feedback and contact information before completion.
- Ratings below 5 require immediate restaurant follow-up and create a recovery case.
- Rating 5 asks whether the guest wants to continue to Google Maps and/or Yelp to leave a public review.
- After the public-review prompt, rating 5 collects contact information for the giveaway entry.
- Contact capture must support name, phone, and email.
- Phone or email is required for giveaway entry.
- Marketing consent is optional and separate from transactional follow-up consent.

## Slice 2 Rules

- No POS integration is required.
- No guest identity is required before the flyer is scanned or texted.
- No purchase/order context is required.
- The campaign itself is the attribution context.
- QR and text entry must resolve to the same survey campaign.
- Rating 5 is public-review eligible.
- Ratings 1 through 4 are follow-up required.
- Ratings 1 through 4 create recovery cases through the same outbox-backed workflow.
- The restaurant should be able to act quickly on any rating below 5.

## Slice 2 Definition of Done

- Admin can create a flyer survey campaign.
- Admin can generate a QR-backed campaign link.
- Heard must not render QR codes itself; campaign QR assets must come from qurl when qurl is configured.
- If qurl is not configured, Heard must create the survey link and clearly report that QR asset generation is unavailable.
- Admin can preview flyer copy and QR code.
- Guest can complete the campaign survey from QR link.
- Guest flow branches correctly for ratings 1 through 4 versus rating 5.
- Giveaway contact capture is enforced.
- Ratings below 5 create recovery cases.
- Rating 5 stores public-review prompt metadata without creating a recovery case.
- Epic 1 user stories continue building toward this printed flyer survey product.

---

# Codex Execution Instructions

Codex should not improvise broad architecture outside this plan.

Codex should build Slice 1 as a working vertical slice, not as disconnected layers.

For every implementation task, Codex must include:

- Goal
- Scope
- Contracts
- Data model
- API changes
- Events emitted
- Permissions
- Failure modes
- Observability
- Tests
- Definition of done

Codex should prefer boring, explicit, maintainable Go code. Avoid clever abstractions, premature microservices, hidden global state, provider lock-in, and incomplete scaffolding.

If a requirement is ambiguous, Codex should choose the simplest implementation that preserves the architecture principles in this document.

If the user provides new instructions that materially affect UX standards, architecture, process, delivery expectations, or quality bars, Codex must persist those instructions into the relevant governing documents as part of the work.

If a temporary admin/testing setup is included in a slice for velocity, Codex must clearly label it as non-production and include the production-grade path in the plan.

---

## Product Backlog

This document defines the product epics and user stories for an open-source restaurant guest experience platform intended to replace and improve on Ovation-style products. The product should support independent restaurants, multi-location restaurant groups, ghost kitchens, food halls, quick-serve, fast-casual, full-service, delivery-heavy concepts, and enterprise restaurant brands.

The product should be modular, extensible, API-oriented, and designed for self-hosting, managed hosting, and future commercial add-ons. Product requirements should be written clearly enough for Codex or another agentic coding tool to convert into implementation plans, technical specs, schemas, APIs, and code.

This first section focuses only on the Frictionless Feedback product surface.

---

# Epic Set 1: Frictionless Feedback

## Product Intent

Frictionless Feedback is the primary guest-facing entry point into the platform. It gives restaurants a simple, high-response way to hear from guests immediately after an experience, without forcing them through long forms, account creation, app downloads, or awkward public complaint channels.

The experience should feel fast, human, branded, and effortless. A guest should be able to scan a QR code, click a text link, or tap a post-order link and answer the core question in seconds: “How was your experience?”

From there, the platform should collect just enough structured and unstructured feedback to help the restaurant understand what happened, identify whether the guest needs recovery, capture permission to follow up, and connect that feedback to the right location, order channel, staff context, and operational category.

This product must support both short-form feedback and deeper survey flows. The short-form flow should optimize for response volume and recovery speed. The long-form survey flow should support deeper guest research, menu testing, LTO feedback, brand studies, and operational investigations.

The product should be designed better than legacy restaurant feedback tools by being self-serve, open-source, privacy-conscious, compliant, extensible, transparent, and friendly to single-location operators as well as multi-unit brands.

---

## Epic 1.1: Guest Feedback Entry Points

### Goal
Allow guests to start a feedback flow from any relevant restaurant touchpoint, including QR codes, SMS links, email links, receipts, table tents, delivery inserts, kiosk screens, order confirmation pages, post-order pages, loyalty messages, call-to-text flows, and embedded web links.

### User Stories

#### Story 1.1.1: Start feedback from a QR code
As a restaurant guest, I want to scan a QR code and immediately open a feedback form so that I can share my experience without downloading an app or logging in.

Acceptance criteria:
- The QR code opens a mobile-friendly feedback page.
- The page loads quickly on common mobile browsers.
- The guest is not required to create an account.
- The feedback flow is associated with the correct restaurant account.
- If the QR code is location-specific, the feedback is associated with the correct location.
- If the QR code is context-specific, the feedback can attach to an Experience when available, or to opaque QR/link context tags when no Experience exists.
- If the QR code is expired, disabled, or invalid, the guest sees a graceful error page with restaurant-safe language.

#### Story 1.1.2: Start feedback from an SMS link
As a restaurant guest, I want to tap a feedback link from a text message so that I can quickly respond after placing or receiving an order.

Acceptance criteria:
- The feedback link opens the correct branded feedback flow.
- The feedback session can be tied to the phone number that received the message when legally and technically available.
- The guest can provide feedback without re-entering order information if the platform already has it.
- The system records the source as SMS.
- The system records campaign, trigger, and integration metadata when available.

#### Story 1.1.3: Start feedback from an email link
As a restaurant guest, I want to click a feedback link from an email so that I can share feedback after an online order, catering order, reservation, or visit.

Acceptance criteria:
- The feedback link opens the correct survey.
- The system records source as email.
- The system can associate the session with email address, order, reservation, location, or campaign when available.
- The guest can still respond if email identity data is unavailable.

#### Story 1.1.4: Start feedback from a receipt link
As a restaurant guest, I want to use a printed or digital receipt link to provide feedback so that I can respond after my transaction.

Acceptance criteria:
- The link or QR code can be printed on physical receipts.
- The link or QR code can be included in digital receipts.
- Receipt-specific metadata should prioritize source system, location, and the source system's stable unique entity ID when available.
- The guest is not exposed to internal IDs in a confusing or unsafe way.

#### Story 1.1.5: Start feedback from a delivery or takeout insert
As a restaurant guest, I want to scan a QR code on packaging or an insert so that I can give feedback about a delivery or takeout experience.

Acceptance criteria:
- Operators can create QR codes for delivery, takeout, catering, curbside, drive-thru, dine-in, or custom channels.
- The feedback flow can ask channel-specific follow-up questions.
- The system records order channel when known or provided.
- The system supports QR codes that are not tied to a specific order.

#### Story 1.1.6: Start feedback from an embedded website or order confirmation page
As a restaurant guest, I want to provide feedback from a restaurant website or order confirmation page so that I can respond while the experience is fresh.

Acceptance criteria:
- The platform provides embeddable links or widgets.
- Embedded feedback can inherit metadata from the host page when configured.
- The guest experience works on mobile and desktop.
- The restaurant can choose whether the feedback opens inline, in a modal, or on a hosted feedback page.

#### Story 1.1.6A: Secure QR and feedback link tokenization
As a platform operator, I want QR and feedback links to use opaque tokenized identifiers so that malicious users cannot infer or manipulate internal restaurant context.

Acceptance criteria:
- QR and feedback links use opaque identifiers or signed tokens.
- Internal identifiers such as table IDs, location IDs, staff IDs, or order IDs are never directly exposed when avoidable.
- Expired or tampered tokens are rejected safely.
- Token verification failures are logged.
- Short links remain mobile-friendly and printable.

#### Story 1.1.7: Preserve entry context across the feedback session
As a restaurant operator, I want feedback to retain its original source and context so that I can understand where the response came from and what it relates to.

Acceptance criteria:
- Feedback sessions capture source, location, channel, campaign, QR code, link, integration, Experience reference, timestamp, and scoped custom metadata when available.
- Metadata persists through the full feedback flow.
- Metadata is visible in admin views and available through exports and APIs.
- Invalid or missing metadata does not prevent the guest from submitting feedback.

---

## Epic 1.2: Short-Form Feedback Flow

### Goal
Create a high-conversion feedback experience that starts with one simple rating question and allows guests to provide meaningful detail only when needed.

### User Stories

#### Story 1.2.1: Answer the primary experience question
As a restaurant guest, I want to answer a simple first question about my experience so that I can give feedback quickly.

Acceptance criteria:
- The first screen asks a clear primary question such as “How was your experience?”
- The restaurant can configure the exact wording.
- The response mechanism can support rating buttons, emoji, stars, thumbs, NPS-style scale, smile/frown scale, or custom options.
- The selected answer is saved immediately or safely retained before final submission.
- The experience is optimized for one-handed mobile use.

#### Story 1.2.2: Branch based on guest sentiment
As a restaurant operator, I want the feedback flow to branch based on the guest’s first answer so that happy, neutral, and unhappy guests receive appropriate next steps.

Acceptance criteria:
- The restaurant can define which ratings count as positive, neutral, or negative.
- Positive, neutral, and negative paths can have different follow-up questions.
- The platform supports default branching rules out of the box.
- Branching decisions are stored with the feedback record.
- Branching can trigger downstream workflows such as recovery alerts, review prompts, tags, or marketing opt-ins.

#### Story 1.2.3: Collect open-text feedback
As a restaurant guest, I want to explain what happened in my own words so that the restaurant understands my experience.

Acceptance criteria:
- The guest can enter open-text feedback.
- The open-text field can be optional or required based on rating path.
- The restaurant can configure placeholder text and helper copy.
- The system supports multilingual text input.
- The system handles profanity, emojis, long messages, and special characters without failure.

#### Story 1.2.4: Ask lightweight category follow-ups
As a restaurant guest, I want to select what my feedback is about so that I do not need to type everything manually.

Acceptance criteria:
- The platform can present quick-select categories such as food quality, speed, service, order accuracy, cleanliness, value, atmosphere, delivery, packaging, app/website, loyalty, catering, or other.
- Categories can be customized by restaurant or brand.
- Categories can vary by sentiment path, channel, or survey type.
- Guests can select one or multiple categories.
- Selected categories are stored as structured feedback tags.

#### Story 1.2.5: Collect contact information for follow-up
As a restaurant guest, I want to provide my contact information when I want a response so that the restaurant can follow up with me.

Acceptance criteria:
- The flow can collect name, phone number, and email address.
- Contact fields can be optional, required, or conditional based on rating path.
- The platform can explain why contact information is being requested.
- The guest can submit anonymous feedback when the restaurant allows it.
- Contact information is stored securely and associated with the guest profile when applicable.

#### Story 1.2.6: Request permission to contact the guest
As a restaurant operator, I want guests to explicitly consent to follow-up communication so that we can respond appropriately and comply with messaging rules.

Acceptance criteria:
- The feedback flow includes clear consent language when phone or email follow-up may occur.
- Consent is recorded with timestamp, source, consent language version, and contact method.
- The guest can submit feedback without granting marketing consent unless restaurant policy requires contact for a specific flow.
- Transactional follow-up and marketing consent are captured separately.
- Consent records are auditable.

#### Story 1.2.7: Submit feedback successfully
As a restaurant guest, I want to know my feedback was received so that I trust the restaurant heard me.

Acceptance criteria:
- The guest sees a confirmation state after submission.
- The confirmation message can vary based on sentiment path.
- The restaurant can configure thank-you language.
- The system prevents accidental duplicate submissions where reasonable.
- If submission fails, the guest sees a recoverable error and can retry.

---

## Epic 1.3: Long-Form Survey Builder

### Goal
Allow restaurants and brands to create deeper surveys for research, operations, product testing, catering feedback, employee interaction feedback, and brand studies without needing custom development.

### User Stories

#### Story 1.3.1: Create a custom survey
As a restaurant admin, I want to create a custom survey so that I can ask guests questions beyond the default feedback flow.

Acceptance criteria:
- Admins can create, name, edit, duplicate, archive, and delete surveys.
- Surveys can be assigned to one location, multiple locations, or an entire brand.
- Surveys can have draft, published, paused, archived, and deleted states.
- A survey cannot collect responses until published.
- Published survey edits preserve historical response integrity.

#### Story 1.3.2: Add multiple question types
As a restaurant admin, I want to add different question types so that I can collect the right kind of guest input.

Acceptance criteria:
- Supported question types include short text, long text, single choice, multiple choice, rating scale, star rating, emoji rating, yes/no, NPS-style score, date, number, email, phone, name, dropdown, ranking, matrix, and consent checkbox.
- Each question can be marked required or optional.
- Each question can include helper text.
- Each question can include internal reporting labels.
- The survey builder validates unsupported or incomplete question configurations.

#### Story 1.3.3: Reorder survey questions
As a restaurant admin, I want to reorder questions so that I can optimize the guest experience.

Acceptance criteria:
- Admins can reorder questions before publishing.
- Reordering is reflected in the guest survey flow.
- Reordering does not corrupt existing draft responses.
- Reordering published surveys creates a safe version change when required.

#### Story 1.3.4: Configure conditional logic
As a restaurant admin, I want survey questions to appear based on previous answers so that guests only see relevant questions.

Acceptance criteria:
- Admins can define branching rules based on prior answers.
- Branching can show, hide, skip, or end the survey.
- Branching can route guests to different final screens.
- The builder warns admins about unreachable questions or circular logic.
- The system stores the path each guest took through the survey.

#### Story 1.3.5: Preview a survey
As a restaurant admin, I want to preview a survey before publishing so that I can confirm the guest experience.

Acceptance criteria:
- Admins can preview surveys on desktop and mobile layouts.
- Preview mode does not create real guest feedback.
- Admins can test conditional paths.
- Preview clearly indicates that it is not live.

#### Story 1.3.6: Version published surveys
As a restaurant admin, I want published surveys to be versioned so that changes do not break reporting or historical comparisons.

Acceptance criteria:
- Publishing creates an immutable survey version.
- Changes to published surveys create a new version when they affect response structure.
- Reports can distinguish between survey versions.
- Admins can view version history.
- Admins can restore or duplicate a prior version into a new draft.

---

## Epic 1.4: Feedback Configuration and Branding

### Goal
Allow restaurants to make feedback experiences feel native to their brand while preserving usability, accessibility, and high conversion.

### User Stories

#### Story 1.4.1: Configure brand appearance
As a restaurant admin, I want to customize feedback pages with my brand colors, logo, and styling so that guests recognize the experience as ours.

Acceptance criteria:
- Admins can upload a logo.
- Admins can configure primary color, accent color, background color, button style, and font preferences where supported.
- The system validates color contrast for readability.
- The platform provides a clean default theme if no branding is configured.
- Branding can be configured at brand and location levels.

#### Story 1.4.2: Configure guest-facing copy
As a restaurant admin, I want to customize the wording in the feedback flow so that it matches our voice.

Acceptance criteria:
- Admins can edit primary question text, helper text, button labels, category labels, confirmation messages, error messages, consent copy, and follow-up prompts.
- The system provides safe default copy.
- Copy can be configured by sentiment path.
- Copy can be configured by language when localization is enabled.
- Required legal or compliance copy cannot be accidentally removed when enabled.

#### Story 1.4.3: Configure rating scales
As a restaurant admin, I want to choose the rating method that fits my brand so that guests respond naturally.

Acceptance criteria:
- Admins can choose from emoji, stars, thumbs up/down, numeric scale, NPS-style scale, smile/frown, or custom labeled options.
- Admins can map each rating option to positive, neutral, or negative sentiment.
- The system provides recommended defaults.
- Changes to rating scales are versioned when they affect reporting.

#### Story 1.4.4: Configure channel-specific flows
As a restaurant admin, I want different feedback flows for dine-in, takeout, delivery, catering, drive-thru, and digital orders so that questions are relevant.

Acceptance criteria:
- Admins can create channel-specific variants of a feedback flow.
- Each variant can have unique copy, categories, branching, and follow-up questions.
- Reports can compare feedback across channels.
- A default flow is used when the channel is unknown.

#### Story 1.4.5: Configure location-specific overrides
As a multi-location operator, I want individual locations to customize limited parts of the feedback experience so that local details are accurate without breaking brand standards.

Acceptance criteria:
- Brand admins can define which settings are locked globally and which can be overridden locally.
- Location admins can only edit allowed fields.
- Location overrides are visible to brand admins.
- Global updates do not accidentally overwrite intentional local overrides without warning.

---

## Epic 1.5: QR Code and Feedback Link Management

### Goal
Allow restaurants to create, manage, track, and safely retire QR codes and feedback links for different locations, channels, campaigns, and operational contexts.

### User Stories

#### Story 1.5.1: Generate a feedback QR code
As a restaurant admin, I want to generate a QR code for a feedback flow so that guests can easily access it from physical or digital touchpoints.

Acceptance criteria:
- Admins can generate QR codes from the platform.
- Each QR code points to a feedback entry URL.
- QR codes can be associated with brand, location, survey, channel, campaign, Experience, or opaque custom context metadata.
- QR codes can be downloaded as PNG and SVG.
- QR codes remain scannable after download.

#### Story 1.5.2: Generate feedback links
As a restaurant admin, I want to generate shareable feedback links so that I can place them in emails, receipts, websites, social posts, and order flows.

Acceptance criteria:
- Admins can create shareable links for any published feedback flow or survey.
- Links can include UTM-like metadata.
- Links can be copied from the admin UI.
- Links can be disabled or archived.
- Links support short, readable URLs when configured.

#### Story 1.5.3: Manage QR code and link inventory
As a restaurant admin, I want to manage all QR codes and feedback links in one place so that I know what is active and where each one is used.

Acceptance criteria:
- Admins can view all QR codes and links by brand, location, survey, channel, campaign, status, created date, and last used date.
- Admins can search and filter QR codes and links.
- Admins can rename codes and links for internal clarity.
- Admins can archive or disable codes and links.
- The system warns admins before disabling a code or link that has recent activity.

#### Story 1.5.4: Track QR and link performance
As a restaurant operator, I want to see how each QR code or feedback link performs so that I can understand which touchpoints generate feedback.

Acceptance criteria:
- The platform tracks scans, opens, starts, submissions, conversion rate, and last activity.
- Performance can be viewed by QR code, link, channel, campaign, location, and survey.
- Bot or suspicious traffic can be filtered when identifiable.
- Reports distinguish between scan/open events and completed feedback.

#### Story 1.5.5: Preserve static QR codes while changing destinations
As a restaurant admin, I want to update what a QR code opens without reprinting physical materials so that I can safely change surveys or campaigns.

Acceptance criteria:
- QR codes can resolve through a platform-managed redirect.
- Admins can change the destination survey or flow for a QR code.
- Destination changes are logged.
- Admins can preview the current destination.
- Disabled QR codes show a safe fallback page.

---

## Epic 1.6: Feedback Triggers and Delivery Rules

### Goal
Automatically request feedback from guests at the right moment using configured rules, integrations, and communication channels.

### User Stories

#### Story 1.6.1: Trigger feedback after an order
As a restaurant operator, I want feedback requests to be sent automatically after orders so that we collect timely responses without manual work.

Acceptance criteria:
- The platform can trigger feedback requests after completed orders when order events are available.
- Trigger rules can vary by order channel.
- Admins can configure delay timing after order completion.
- The system prevents duplicate feedback requests for the same order when configured.
- Failed trigger attempts are logged.

#### Story 1.6.2: Trigger feedback after a reservation or visit
As a restaurant operator, I want feedback requests to be sent after reservations or guest visits so that dine-in guests can respond after their meal.

Acceptance criteria:
- The platform can receive reservation or visit events from integrations when available.
- Admins can configure timing rules.
- The system can associate the feedback request with reservation metadata.
- Duplicate suppression is supported.

#### Story 1.6.3: Trigger feedback after catering orders or events
As a catering manager, I want feedback requests after catering orders so that I can understand event quality and identify follow-up opportunities.

Acceptance criteria:
- Admins can create catering-specific feedback triggers.
- Catering feedback can include event date, event size, delivery/setup notes, and customer contact metadata when available.
- Catering-specific surveys can be used.
- Reports can separate catering feedback from regular orders.

#### Story 1.6.4: Configure delivery channel rules
As a restaurant admin, I want to choose whether feedback requests are sent by SMS, email, both, or neither so that communication matches guest permissions and channel availability.

Acceptance criteria:
- Trigger rules can choose SMS, email, or both.
- The system respects opt-out and consent rules.
- The system falls back to email when SMS is unavailable if configured.
- The system does not send marketing messages under a transactional feedback rule.

#### Story 1.6.5: Configure suppression rules
As a restaurant admin, I want to prevent guests from being over-messaged so that feedback collection does not annoy customers.

Acceptance criteria:
- Admins can configure suppression windows by guest, phone number, email, location, brand, or order.
- Admins can suppress requests after refunds, cancellations, failed orders, test orders, staff orders, or low-value transactions when data is available.
- Suppression decisions are logged.
- Suppression rules can be tested before activation.

#### Story 1.6.6: Manually send a feedback request
As a restaurant manager, I want to manually send a feedback request to a guest so that I can follow up after a specific interaction.

Acceptance criteria:
- Authorized users can send a feedback request by phone number or email.
- The manager can choose the location, flow, and optional context.
- The system checks opt-out status before sending.
- The send action is logged with user, timestamp, recipient, and feedback flow.

---

## Epic 1.7: Guest Identity and Feedback Attribution

### Goal
Connect feedback to guest, order, location, channel, and visit context when available while still allowing anonymous and low-friction feedback.

### User Stories

#### Story 1.7.1: Match feedback to an existing guest profile
As a restaurant operator, I want feedback to connect to an existing guest profile when possible so that I can understand guest history and recovery context.

Acceptance criteria:
- Feedback can be matched by phone number, email, customer ID, loyalty ID, order ID, reservation ID, or integration-specific identifiers.
- Matching confidence is stored.
- The system avoids unsafe automatic merging when confidence is low.
- Admins can review or correct guest matching where permitted.

#### Story 1.7.2: Create a new guest profile from feedback
As a restaurant operator, I want a new guest profile to be created when a new guest provides contact information so that future interactions can be connected.

Acceptance criteria:
- A guest profile can be created from submitted feedback.
- The profile stores contact information, consent status, feedback history, and source metadata.
- Duplicate detection is applied.
- The platform respects privacy and retention settings.

#### Story 1.7.3: Allow anonymous feedback
As a restaurant guest, I want to provide feedback anonymously when allowed so that I can share honestly without personal follow-up.

Acceptance criteria:
- Restaurants can allow or disallow anonymous feedback by flow.
- Anonymous feedback does not require name, phone, or email.
- Anonymous feedback still captures non-personal metadata such as location, source, channel, and timestamp when available.
- Anonymous feedback is clearly marked in admin views.

#### Story 1.7.4: Attribute feedback to an Experience
As a restaurant operator, I want feedback connected to a specific guest Experience when available so that I can investigate problems without forcing every integration to model orders the same way.

Acceptance criteria:
- Feedback can attach to a Heard-owned Experience UUID.
- An Experience is identified by tenant_id, location_id, source_system, and source_entity_id.
- Provider-specific order, ticket, check, receipt, reservation, delivery, or catering identifiers are stored as external references, not as Heard primary keys.
- Feedback can still be submitted without an Experience when only generic QR/link context exists.
- Optional provider enrichment does not block feedback submission or recovery case creation.

#### Story 1.7.5: Store optional operational enrichment separately from identity
As a restaurant operator, I want optional context stored separately from core identity so that Heard can support richer analytics without bloating the required feedback model.

Acceptance criteria:
- Core feedback does not require table, staff, subtotal, items, modifiers, or fulfillment details.
- Optional enrichment can be attached to an Experience when provided by an integration.
- Enrichment fields are nullable and provider-aware.
- Enrichment must not be required for MVP flows.
- Analytics can use enrichment when available and degrade gracefully when it is missing.

---

## Epic 1.8: Feedback Data Capture and Validation

### Goal
Capture complete, reliable, secure feedback records that can support recovery, analytics, reporting, automation, and integrations.

### User Stories

#### Story 1.8.1: Store a complete feedback record
As a platform operator, I want every submitted feedback response to create a durable feedback record so that downstream workflows can rely on it.

Acceptance criteria:
- Each feedback record has a unique ID.
- Each record stores survey/flow ID, version, location, brand, source, submitted timestamp, sentiment, rating, answers, metadata, guest identity reference when available, and consent records when applicable.
- Records are queryable by location, guest, source, channel, sentiment, category, survey, date range, and status.
- Records are available through internal services and APIs.

#### Story 1.8.2: Validate feedback submissions
As a platform operator, I want submissions validated before storage so that bad data does not corrupt reporting or workflows.

Acceptance criteria:
- Required fields are enforced.
- Field formats are validated for phone, email, number, date, and configured question types.
- Invalid submissions return useful guest-safe errors.
- Suspicious payloads are rejected or sanitized.
- Validation errors are logged without exposing sensitive data.

#### Story 1.8.3: Save partial feedback progress when appropriate
As a restaurant operator, I want partial feedback progress captured when possible so that we can understand abandonment and improve conversion.

Acceptance criteria:
- The platform can record feedback session starts, question progress, abandonment, and completion.
- Partial records do not appear as completed feedback.
- Partial data respects privacy and consent settings.
- Abandonment reporting is available by flow, source, question, and device type.

#### Story 1.8.3A: Support privacy-preserving abuse prevention
As a platform operator, I want anonymous feedback flows protected against automated abuse so that guests can remain anonymous without allowing large-scale spam.

Acceptance criteria:
- Anonymous flows support lightweight abuse mitigation.
- The system can support proof-of-work, invisible CAPTCHA, behavioral detection, signed challenge tokens, or equivalent privacy-preserving mechanisms.
- Abuse protection should minimize guest friction.
- Abuse prevention should avoid unnecessary PII collection.
- Suspicious submissions can be rate-limited or quarantined.

#### Story 1.8.4: Prevent abuse and spam
As a restaurant operator, I want protection against spam and abusive submissions so that feedback remains useful.

Acceptance criteria:
- The platform supports rate limiting by IP, device fingerprint where permitted, phone, email, QR code, link, and location.
- The platform can flag suspicious repeated submissions.
- Basic bot protection is available without creating guest friction.
- Admins can mark feedback as spam or not spam.
- Spam status is excluded from default reporting but can be audited.

#### Story 1.8.5: Handle offline or unreliable guest connections
As a restaurant guest, I want the feedback flow to tolerate weak mobile connections so that my response is not lost.

Acceptance criteria:
- The UI handles slow or failed network requests gracefully.
- The guest can retry submission.
- The platform avoids duplicate records from retry attempts.
- The guest sees clear status messages.

---

## Epic 1.9: Feedback Alerts and Routing Hooks

### Goal
Prepare feedback events for downstream workflows such as recovery alerts, review prompts, manager notifications, automations, and analytics without fully defining those later product areas yet.

### User Stories

#### Story 1.9.1: Emit feedback submitted events
As a platform operator, I want every feedback submission to emit an event so that other modules can react to it.

Acceptance criteria:
- A feedback submitted event is emitted after successful storage.
- The event includes feedback ID, brand ID, location ID, sentiment, source, channel, survey ID, version, guest reference, and key metadata.
- Events are idempotent or safely de-duplicated by downstream consumers.
- Failed event emission is retried or logged for recovery.

#### Story 1.9.2: Flag feedback requiring recovery
As a restaurant operator, I want negative or high-risk feedback to be marked for recovery so that managers can respond quickly.

Acceptance criteria:
- Feedback can be flagged for recovery based on rating, sentiment, category, text content, selected issue, VIP guest status, order value, or custom rules.
- Default recovery rules are provided.
- Recovery flagging does not require the recovery module to be fully enabled.
- Recovery status is visible on the feedback record.

#### Story 1.9.3: Flag feedback eligible for public review prompt
As a restaurant operator, I want positive feedback to be marked as eligible for a review prompt so that happy guests can be invited to share publicly in a compliant way.

Acceptance criteria:
- Feedback can be marked review-prompt eligible based on rating, sentiment, consent, source, and restaurant configuration.
- The system supports compliant, neutral review invitation language.
- The platform does not require review gating as a default behavior.
- Review eligibility is stored separately from actual review request delivery.

#### Story 1.9.4: Trigger internal notifications
As a restaurant manager, I want important feedback to trigger notifications so that I can act quickly.

Acceptance criteria:
- Feedback can trigger notifications by sentiment, category, location, channel, rating, or custom rule.
- Notification delivery is abstracted so future channels can include email, SMS, push, Slack, Teams, or webhooks.
- Notification failures are logged.
- Notification preferences are configurable by role and location.

---

## Epic 1.10: Guest Experience Quality, Accessibility, and Localization

### Goal
Make the feedback experience fast, accessible, inclusive, and trustworthy for real restaurant guests across devices, languages, and abilities.

### User Stories

#### Story 1.10.1: Optimize for mobile speed
As a restaurant guest, I want the feedback page to load quickly so that I do not abandon it.

Acceptance criteria:
- The feedback page is optimized for mobile-first performance.
- Critical content renders quickly on average mobile connections.
- Large assets are optimized or deferred.
- The platform tracks load time, start rate, and completion rate.

#### Story 1.10.2: Support accessible feedback flows
As a restaurant guest using assistive technology, I want the feedback form to be accessible so that I can provide feedback independently.

Acceptance criteria:
- Guest-facing forms follow WCAG-aligned accessibility practices.
- Interactive controls are keyboard accessible.
- Form controls have labels and accessible names.
- Color is not the only way to communicate status or meaning.
- Error messages are accessible to screen readers.

#### Story 1.10.3: Support multiple languages
As a restaurant guest, I want to provide feedback in my preferred language so that I can clearly explain my experience.

Acceptance criteria:
- The platform supports multiple configured languages per brand or location.
- The guest can select a language when multiple languages are enabled.
- The system can infer language from browser settings when configured.
- Admin-configured copy supports localization.
- Guest-entered feedback can be stored in the original language.

#### Story 1.10.4: Translate feedback for operators
As a restaurant manager, I want guest feedback translated when needed so that I can understand and respond.

Acceptance criteria:
- The platform can detect feedback language.
- The platform can store original text and translated text separately.
- Translations are clearly marked as machine-generated when applicable.
- Translation can be disabled for self-hosted deployments without translation providers.

#### Story 1.10.5: Build guest trust with privacy-safe language
As a restaurant guest, I want to understand how my feedback and contact information will be used so that I feel safe responding.

Acceptance criteria:
- Feedback pages include clear, concise privacy language.
- Contact and consent prompts distinguish follow-up from marketing.
- The platform can link to restaurant-specific privacy policies.
- The platform can link to platform-level privacy disclosures when hosted.

---

## Epic 1.11: Feedback Admin Views

### Goal
Give restaurant operators a simple way to view, search, filter, and understand incoming feedback before deeper analytics and recovery modules are introduced.

### User Stories

#### Story 1.11.1: View feedback inbox
As a restaurant manager, I want to see recent feedback in one place so that I can understand what guests are saying.

Acceptance criteria:
- Managers can view a chronological list of submitted feedback.
- Each row shows rating, sentiment, location, source, channel, guest name or anonymous status, submitted time, categories, and recovery flag.
- Managers can filter by location, date, sentiment, rating, source, channel, category, survey, and status.
- Managers can search feedback text and guest information where permitted.

#### Story 1.11.2: View feedback details
As a restaurant manager, I want to open a feedback record so that I can see the full guest response and context.

Acceptance criteria:
- The detail view shows all submitted answers.
- The detail view shows guest contact information when available and permitted.
- The detail view shows metadata such as source, channel, order, location, survey version, and timestamps.
- The detail view shows consent records related to the feedback.
- The detail view shows recovery and review eligibility flags.

#### Story 1.11.3: Mark feedback status
As a restaurant manager, I want to mark feedback as new, reviewed, ignored, spam, or needs follow-up so that my team can manage the queue.

Acceptance criteria:
- Managers can update feedback status.
- Status changes are logged with user and timestamp.
- Status can be filtered in the feedback inbox.
- Permissions control who can change status.

#### Story 1.11.4: Export feedback
As a restaurant admin, I want to export feedback so that I can analyze it outside the platform.

Acceptance criteria:
- Admins can export feedback as CSV.
- Exports can be filtered by date, location, survey, source, channel, sentiment, rating, category, and status.
- Exports include structured answers and metadata.
- Export permissions are role-controlled.
- Export activity is logged.

---

## Epic 1.12: Feedback APIs and Webhooks

### Goal
Expose feedback capabilities through stable APIs and webhooks so the open-source ecosystem can integrate with POS systems, ordering platforms, loyalty systems, CRMs, data warehouses, and custom restaurant tools.

### User Stories

#### Story 1.12.1: Create feedback sessions by API
As an integration developer, I want to create feedback sessions through an API so that external systems can generate guest-specific feedback links.

Acceptance criteria:
- The API can create a feedback session for a specific brand, location, survey, flow, guest, order, channel, and metadata payload.
- The API returns a feedback URL.
- The API supports idempotency keys.
- The API validates permissions and tenant boundaries.
- API errors are clear and machine-readable.

#### Story 1.12.2: Submit feedback by API
As an integration developer, I want to submit feedback through an API so that custom front ends or embedded experiences can use the platform backend.

Acceptance criteria:
- The API accepts structured answers and metadata.
- The API enforces survey version validation.
- The API returns a durable feedback ID after successful submission.
- The API rejects invalid payloads with clear errors.
- API submissions trigger the same downstream events as hosted form submissions.

#### Story 1.12.3: Retrieve feedback by API
As an integration developer, I want to retrieve feedback through an API so that external systems can display, analyze, or sync feedback data.

Acceptance criteria:
- The API supports listing and retrieving feedback records.
- The API supports filtering by brand, location, date range, sentiment, rating, source, channel, status, guest, and survey.
- The API enforces role, token, and tenant permissions.
- The API supports pagination.
- Sensitive fields are omitted unless the caller has permission.

#### Story 1.12.4: Subscribe to feedback webhooks
As an integration developer, I want webhooks for feedback events so that external systems can react in real time.

Acceptance criteria:
- Developers can subscribe to feedback submitted, feedback updated, feedback flagged for recovery, feedback marked spam, and feedback status changed events.
- Webhook endpoints can be configured per brand or integration.
- Webhooks include signed payloads.
- Failed deliveries are retried.
- Delivery attempts are logged and inspectable.

---

## Epic 1.13: Open-Source Readiness for Frictionless Feedback

### Goal
Ensure the feedback product is designed as a credible open-source system, not a brittle SaaS clone, so contributors can understand, extend, self-host, and safely integrate with it.

### User Stories

#### Story 1.13.1: Provide clear module boundaries
As an open-source contributor, I want the feedback product organized into clear modules so that I can contribute without understanding the entire platform.

Acceptance criteria:
- Feedback entry, survey rendering, survey builder, response storage, QR/link management, trigger rules, guest identity matching, admin views, APIs, and events are separable modules or packages.
- Module responsibilities are documented.
- Cross-module dependencies are explicit.
- Shared contracts are stable and versioned.

#### Story 1.13.2: Support provider abstractions
As a self-hosting operator, I want provider abstractions for SMS, email, storage, translation, analytics, and event delivery so that I can choose vendors or run locally.

Acceptance criteria:
- SMS delivery is abstracted behind a provider interface.
- Email delivery is abstracted behind a provider interface.
- Translation is abstracted behind a provider interface.
- Object storage for logos/assets/exports is abstracted behind a provider interface.
- Event delivery supports local and external implementations.

#### Story 1.13.3: Include local development defaults
As an open-source developer, I want the feedback module to run locally with safe defaults so that I can build and test without paid third-party services.

Acceptance criteria:
- Local development can run without real SMS or email providers.
- Fake provider implementations are available for local testing.
- Seed data includes sample brand, locations, surveys, QR links, feedback, and guests.
- Documentation explains how to run the feedback module locally.

#### Story 1.13.4: Provide extensible schemas
As an integration developer, I want feedback-related schemas to be explicit and extensible so that I can safely build integrations.

Acceptance criteria:
- Feedback session, survey, question, answer, guest, consent, QR/link, trigger, and webhook schemas are documented.
- Schemas support custom metadata without breaking core fields.
- Schema versions are explicit.
- Breaking schema changes are documented.

#### Story 1.13.5: Provide test coverage expectations
As a maintainer, I want clear test expectations for the feedback module so that contributions do not break critical guest flows.

Acceptance criteria:
- Critical guest feedback flows have automated tests.
- Survey branching logic has automated tests.
- API validation has automated tests.
- QR/link routing has automated tests.
- Permission and tenant isolation tests are required for admin and API workflows.

---

---

# Epic Set 2: Guest Recovery

## Product Intent

Guest Recovery turns negative, neutral, high-risk, or operationally important feedback into a structured follow-up workflow. The goal is to help restaurants respond quickly, personally, and consistently when a guest has a poor experience, while giving operators visibility into response quality, root causes, team ownership, recovery outcomes, and repeat guest behavior.

This module should feel like a restaurant-native command center, not a generic support ticketing system. A manager should be able to see what happened, who the guest is, whether the issue is urgent, what order or visit it relates to, what the guest already said, whether the guest has complained before, and what action should happen next.

The product should support both lightweight independent-restaurant recovery and multi-location brand operations. For small restaurants, the experience should be simple: “A guest is unhappy. Reply fast. Make it right.” For larger brands, it should support assignment, escalation, templates, approval rules, offers, SLAs, permissions, audit logs, and performance reporting.

The system should improve on legacy guest recovery tools by being transparent, configurable, open-source, API-oriented, AI-assisted but human-controlled, compliant, and deeply integrated with the feedback record.

---

## Epic 2.1: Recovery Case Creation

### Goal
Automatically or manually create recovery cases from feedback, reviews, calls, imported complaints, integrations, or manager-entered incidents.

### User Stories

#### Story 2.1.1: Automatically create a recovery case from negative feedback
As a restaurant manager, I want negative feedback to automatically create a recovery case so that unhappy guests are not missed.

Acceptance criteria:
- A recovery case can be created automatically when feedback meets configured criteria.
- Default criteria include negative sentiment, low rating, selected issue category, high-risk keywords, request for contact, VIP guest status, high order value, or repeat complaint.
- The recovery case links back to the original feedback record.
- The case includes guest identity, contact information, consent status, location, channel, order context, feedback text, categories, rating, and submitted timestamp when available.
- Duplicate prevention avoids creating multiple open cases for the same feedback record.

#### Story 2.1.2: Create a recovery case from neutral or mixed feedback
As a restaurant operator, I want neutral or mixed feedback to create a case when it indicates a fixable problem so that we can recover guests before they churn.

Acceptance criteria:
- Admins can define rules that create cases from neutral or mixed sentiment feedback.
- Rules can use rating, category, text keywords, guest value, guest frequency, channel, order total, or custom metadata.
- The case clearly identifies why it was created.
- Neutral cases can have a lower default priority than negative cases.

#### Story 2.1.3: Manually create a recovery case
As a restaurant manager, I want to manually create a recovery case so that I can track guest issues that come from phone calls, walk-ins, emails, social messages, staff reports, or delivery provider complaints.

Acceptance criteria:
- Authorized users can manually create a recovery case.
- The manager can enter guest name, phone, email, location, issue category, description, order details, channel, priority, and internal notes.
- The manager can link the case to an existing guest profile, order, feedback record, or review when available.
- Manual cases are marked with source as manual.
- Manual case creation is logged with user and timestamp.

#### Story 2.1.4: Create a recovery case from a public review
As a restaurant operator, I want negative public reviews to create recovery cases so that we can respond publicly and privately when possible.

Acceptance criteria:
- A recovery case can be created from a public review imported through the reputation module.
- The case links to the original review.
- The case includes review platform, rating, review text, review date, reviewer display name, location, and response status when available.
- If guest contact information is unavailable, the case supports public-response-only workflows.
- Review-created cases are distinguishable from private feedback cases.

#### Story 2.1.5: Prevent duplicate recovery cases
As a restaurant manager, I want duplicate complaints to be grouped or flagged so that my team does not work the same issue multiple times.

Acceptance criteria:
- The platform detects possible duplicate cases using feedback ID, guest identity, order ID, review ID, phone, email, timestamp, location, and similar text.
- Potential duplicates are flagged for review.
- Exact duplicates are suppressed when safe.
- Managers can merge duplicate cases when permitted.
- Merging preserves all original records, notes, events, and audit history.

---

## Epic 2.2: Recovery Inbox and Case Queue

### Goal
Give managers and operators a focused queue of guest recovery work, prioritized by urgency, ownership, and operational impact.

### User Stories

#### Story 2.2.1: View recovery inbox
As a restaurant manager, I want to see all open recovery cases in one place so that I can quickly decide what needs attention.

Acceptance criteria:
- The inbox shows open recovery cases by default.
- Each case row shows priority, status, guest name or anonymous label, location, source, rating or sentiment, issue category, submitted time, age, assignee, SLA state, and latest message state.
- Managers can filter by brand, region, location, source, sentiment, rating, category, status, priority, assignee, SLA state, and date range.
- Managers can search by guest name, phone, email, order ID, feedback text, review text, and case ID where permitted.
- The inbox is optimized for daily restaurant manager use on desktop and mobile.

#### Story 2.2.2: Prioritize urgent cases
As a restaurant manager, I want urgent cases to rise to the top so that severe guest issues get fast attention.

Acceptance criteria:
- Cases have priority levels such as low, normal, high, urgent, and critical.
- Priority can be assigned automatically by rules or manually by authorized users.
- Priority rules can use rating, category, keywords, guest value, repeat issue, food safety terms, refund request, delivery failure, VIP status, or custom metadata.
- Priority is visible in the inbox and detail view.
- Priority changes are logged.

#### Story 2.2.3: View cases by location
As a multi-location operator, I want to view recovery cases by location so that each team can manage its own guest issues.

Acceptance criteria:
- Users only see locations they are authorized to access.
- Brand users can view all locations.
- Region or district users can view assigned location groups.
- Location users can view only their location unless granted broader access.
- Location filters are available across recovery views.

#### Story 2.2.4: Save inbox views
As an operator, I want to save common recovery inbox filters so that I can quickly return to the views I use every day.

Acceptance criteria:
- Users can create saved views from filters and sorting.
- Saved views can be private or shared based on permissions.
- Default saved views include My Open Cases, Unassigned Cases, SLA At Risk, Urgent Cases, Food Safety Mentions, and Recently Closed.
- Users can set a default view.

#### Story 2.2.5: Bulk update recovery cases
As a manager, I want to update multiple recovery cases at once so that I can manage the queue efficiently.

Acceptance criteria:
- Authorized users can bulk assign, change status, change priority, add tags, mark spam, or archive selected cases.
- Bulk actions require confirmation.
- Bulk changes are logged per case.
- Bulk actions respect permissions and tenant boundaries.

---

## Epic 2.3: Recovery Case Detail View

### Goal
Provide a complete, usable view of the guest issue, guest history, order context, conversation, internal notes, recovery actions, and audit trail.

### User Stories

#### Story 2.3.1: View full case context
As a restaurant manager, I want to open a recovery case and see everything relevant so that I can respond intelligently.

Acceptance criteria:
- The case detail view shows original feedback or review text, rating, sentiment, categories, source, location, channel, order context, guest profile, contact methods, consent status, submitted timestamp, assigned user, priority, status, and SLA state.
- The case detail view shows any related feedback, reviews, conversations, offers, refunds, notes, and prior cases when available.
- Sensitive fields are hidden from users without permission.
- The view is readable on desktop and usable on mobile.

#### Story 2.3.2: View guest history
As a manager, I want to see a guest’s history so that I know whether this is a loyal guest, first-time guest, repeat complainer, or high-value customer.

Acceptance criteria:
- The case detail view shows prior feedback, prior cases, prior offers, prior conversations, order history summary, visit count, loyalty status, and consent status when available.
- The system distinguishes confirmed guest matches from possible matches.
- Users can open related records where permitted.
- Guest history can be disabled or limited based on privacy settings.

#### Story 2.3.3: View Experience context
As a manager, I want to see the source-system reference for the guest Experience when available so that I can investigate the issue in the restaurant's system of record.

Acceptance criteria:
- The case can display source system, source entity ID, location, and linked Experience UUID when available.
- The case can display optional provider enrichment only when it exists.
- The system does not require items, modifiers, totals, server, table, or payment details for recovery.
- Missing Experience context is handled gracefully.
- Sensitive payment data is not stored.

#### Story 2.3.4: Add internal notes
As a manager, I want to add internal notes to a case so that my team can coordinate without exposing those notes to the guest.

Acceptance criteria:
- Authorized users can add internal notes.
- Notes show author and timestamp.
- Notes are never sent to the guest.
- Notes can mention users or teams when mention support is enabled.
- Notes are included in audit/export views based on permission.

#### Story 2.3.5: View case activity timeline
As an operator, I want a chronological timeline of case activity so that I can understand what has happened.

Acceptance criteria:
- The timeline includes case creation, assignments, status changes, priority changes, notes, messages sent/received, offers issued, offer redemptions, escalations, merges, and closure.
- Each timeline event includes timestamp, actor, event type, and relevant details.
- System-generated events are distinguishable from user actions.
- Timeline data is immutable except for allowed redaction workflows.

---

## Epic 2.4: Guest Messaging and Conversation Management

### Goal
Allow restaurant teams to respond to guests through SMS, email, or other supported channels from inside the recovery workflow.

### User Stories

#### Story 2.4.1: Reply to guest by SMS
As a restaurant manager, I want to reply to a guest by text so that I can recover the guest quickly using the channel they are most likely to read.

Acceptance criteria:
- Managers can send SMS replies from the case detail view when guest phone number and consent/transactional basis allow it.
- The conversation thread shows inbound and outbound messages.
- Messages are associated with the case and guest profile.
- SMS opt-out keywords are respected.
- Failed messages show clear error states and retry options where appropriate.

#### Story 2.4.2: Reply to guest by email
As a restaurant manager, I want to reply by email when SMS is unavailable or inappropriate so that I can still follow up.

Acceptance criteria:
- Managers can send email replies from the case detail view when a guest email is available.
- Emails are associated with the case and guest profile.
- Email subject and body can be customized.
- Email delivery status is tracked when the provider supports it.
- Failed emails show clear error states.

#### Story 2.4.3: Receive guest replies
As a manager, I want guest replies to appear in the case so that the full conversation stays in one place.

Acceptance criteria:
- Inbound SMS replies are routed to the correct case when possible.
- Inbound email replies are routed to the correct case when possible.
- If a reply cannot be matched confidently, it is routed to an unmatched conversation queue.
- New replies update case unread state and latest activity timestamp.
- New replies can trigger notifications.

#### Story 2.4.4: Manage multi-channel conversations
As a manager, I want to see SMS and email messages together so that I understand the full guest conversation.

Acceptance criteria:
- The case shows all guest-facing messages in chronological order.
- Each message indicates channel, sender, recipient, timestamp, and delivery status.
- Users can filter the thread by channel when needed.
- Internal notes are visually distinct from guest-facing messages.

#### Story 2.4.5: Handle guest opt-outs
As a platform operator, I want SMS and email opt-outs handled correctly so that restaurants stay compliant.

Acceptance criteria:
- SMS STOP, UNSUBSCRIBE, CANCEL, END, and QUIT are recognized according to provider capabilities and regulatory expectations.
- Opt-out status is stored on the guest contact method.
- Users are prevented from sending prohibited messages to opted-out contact methods.
- Opt-out events are visible in the guest and case timeline.
- Re-subscribe behavior is provider-aware and auditable.

---

## Epic 2.5: AI-Assisted Recovery Replies

### Goal
Help managers respond faster and better with AI-generated draft replies that are empathetic, brand-aligned, specific, compliant, and editable before sending.

### User Stories

#### Story 2.5.1: Generate a suggested reply
As a restaurant manager, I want the system to suggest a reply so that I can respond quickly without writing from scratch.

Acceptance criteria:
- The platform can generate a draft reply based on feedback text, issue category, rating, guest history, order context, brand voice, and case status when available.
- AI replies are drafts only by default.
- The manager can edit the reply before sending.
- The draft avoids unsupported promises, unsafe admissions, or fabricated details.
- The system indicates that the reply was AI-generated.

#### Story 2.5.2: Apply brand voice to AI replies
As a brand admin, I want AI replies to match our voice so that guest recovery feels consistent across locations.

Acceptance criteria:
- Admins can define brand voice guidelines.
- Admins can define words, phrases, claims, or offers to avoid.
- Admins can define preferred sign-offs and tone.
- Location-specific voice overrides can be allowed or locked by brand admins.
- AI drafts use the configured brand voice where available.

#### Story 2.5.3: Generate multiple reply options
As a manager, I want multiple suggested replies so that I can choose the one that best fits the situation.

Acceptance criteria:
- The platform can generate concise, warm, formal, apologetic, direct, or offer-focused reply variants.
- Users can choose a variant and edit it.
- Variant generation respects brand and compliance rules.
- The system avoids generating manipulative or review-gating language.

#### Story 2.5.4: Summarize long guest conversations
As a manager, I want the system to summarize the case so that I can quickly understand long or transferred conversations.

Acceptance criteria:
- The system can produce a short case summary from feedback, messages, notes, and timeline events.
- Summaries identify the issue, guest sentiment, promised actions, open questions, and recommended next step.
- Summaries are clearly marked as AI-generated.
- Users can refresh summaries as the case changes.

#### Story 2.5.5: Disable AI features
As a self-hosting operator or privacy-sensitive brand, I want to disable AI features so that the system can run without external AI providers.

Acceptance criteria:
- AI reply generation can be disabled globally, by brand, or by environment.
- The product remains usable without AI.
- Provider configuration is abstracted.
- Local development can use a fake AI provider.

---

## Epic 2.6: Response Templates and Playbooks

### Goal
Give restaurants reusable response templates and recovery playbooks so teams can act consistently without losing the human touch.

### User Stories

#### Story 2.6.1: Create response templates
As a restaurant admin, I want to create response templates so that managers can reply faster and more consistently.

Acceptance criteria:
- Admins can create, edit, archive, and delete templates.
- Templates can be scoped to brand, region, location, source, channel, sentiment, category, or case type.
- Templates support variables such as guest name, location name, manager name, issue category, order date, and offer details.
- Templates can be previewed before use.
- Template usage is tracked.

#### Story 2.6.2: Recommend templates by case type
As a manager, I want relevant templates suggested automatically so that I do not have to search for the right one.

Acceptance criteria:
- Templates can be suggested based on category, sentiment, source, channel, priority, location, and guest history.
- Suggested templates appear in the reply composer.
- Users can search and browse all allowed templates.
- Users can edit a template response before sending.

#### Story 2.6.3: Define recovery playbooks
As a brand operator, I want to define playbooks for common issue types so that teams know what steps to take.

Acceptance criteria:
- Admins can create playbooks for issue types such as wrong order, cold food, late delivery, rude service, food safety concern, missing item, refund request, app problem, catering issue, or cleanliness concern.
- Playbooks can include recommended response tone, required steps, optional offer guidance, escalation rules, and closure criteria.
- Playbooks can be shown in the case detail view.
- Playbooks can vary by brand, location, channel, or category.

#### Story 2.6.4: Require playbook steps for severe cases
As a brand admin, I want severe cases to require specific steps before closure so that important issues are handled correctly.

Acceptance criteria:
- Admins can mark playbook steps as required.
- Required steps must be completed before closure unless a user has override permission.
- Overrides require a reason.
- Completion and override events are logged.

---

## Epic 2.7: Offers, Coupons, Refund References, and Make-It-Right Actions

### Goal
Allow managers to issue appropriate recovery gestures, track them, and understand whether they were redeemed or effective.

### User Stories

#### Story 2.7.1: Issue a recovery offer
As a restaurant manager, I want to send a guest an offer so that I can encourage them to give us another chance.

Acceptance criteria:
- Authorized users can create or select a recovery offer from a case.
- Offer types can include percentage discount, fixed amount, free item, free entree, free dessert, free drink, catering credit, or custom apology offer.
- Offers can include expiration dates, redemption limits, location restrictions, and channel restrictions.
- Offer issuance is logged on the case timeline.
- The offer can be sent by SMS or email when allowed.

#### Story 2.7.2: Use approved offer templates
As a brand admin, I want managers to use approved offer templates so that recovery gestures stay within policy.

Acceptance criteria:
- Admins can create offer templates.
- Templates can be scoped by brand, region, location, issue category, priority, channel, and role.
- Templates can define max value, expiration, redemption rules, and approval requirements.
- Managers only see offers they are allowed to issue.

#### Story 2.7.3: Require approval for high-value offers
As a brand operator, I want high-value offers to require approval so that managers do not over-discount.

Acceptance criteria:
- Admins can define approval thresholds by role, location, issue type, and offer value.
- Offers requiring approval cannot be sent until approved.
- Approvers receive notifications.
- Approval, rejection, and comments are logged.
- The guest does not see internal approval status.

#### Story 2.7.4: Track offer redemption
As a restaurant operator, I want to know whether recovery offers were redeemed so that I can measure recovery effectiveness.

Acceptance criteria:
- Offers have unique redemption codes or provider-backed identifiers.
- The platform can record redeemed, expired, cancelled, and unused states.
- Redemption can be updated manually or through integrations when available.
- Redemption events appear on the case and guest timeline.
- Reports can analyze redemption by location, issue type, offer type, and guest segment.

#### Story 2.7.5: Record refund references
As a manager, I want to record refund actions or references so that the case reflects what we did to resolve the issue.

Acceptance criteria:
- Managers can record that a refund was requested, approved, denied, issued, or completed.
- Refund records can include amount, reason, order ID, provider reference, user, and timestamp.
- The platform can link to external refund systems when integrations support it.
- Sensitive payment data is not stored.
- Refund information is permission-controlled.

---

## Epic 2.8: Assignment, Ownership, Collaboration, and Escalation

### Goal
Support clear ownership and collaboration so recovery cases do not fall through the cracks.

### User Stories

#### Story 2.8.1: Assign a case to a user
As a manager, I want to assign a case to myself or a teammate so that ownership is clear.

Acceptance criteria:
- Authorized users can assign and reassign cases.
- Cases can be assigned to individual users.
- Assignment changes are logged.
- Assignees receive notifications when configured.
- The inbox supports filtering by assignee.

#### Story 2.8.2: Assign a case to a team or role
As a brand operator, I want cases assigned to a team or role so that the right group can handle them even when individuals change.

Acceptance criteria:
- Cases can be assigned to teams, roles, or queues.
- Team members can claim cases.
- Assignment rules can route cases by location, category, source, channel, priority, or escalation state.
- Team assignment history is logged.

#### Story 2.8.3: Mention teammates in notes
As a manager, I want to mention teammates in internal notes so that I can pull them into a case.

Acceptance criteria:
- Users can mention allowed users in internal notes.
- Mentioned users receive notifications when configured.
- Mentions respect location and tenant permissions.
- Mention events appear in the timeline.

#### Story 2.8.4: Escalate severe cases
As a restaurant manager, I want to escalate severe cases so that brand leadership or specialists can help.

Acceptance criteria:
- Cases can be escalated manually or automatically.
- Escalation rules can use priority, category, keywords, SLA risk, guest value, repeat complaints, food safety terms, legal terms, media threats, or custom metadata.
- Escalated cases display escalation status.
- Escalation notifications are sent to configured users or teams.
- Escalation and de-escalation events are logged.

#### Story 2.8.5: Transfer a case between locations
As a brand operator, I want to transfer a case to another location when it was misattributed so that the right team owns it.

Acceptance criteria:
- Authorized users can change the case location.
- Location changes preserve original attribution history.
- The receiving location can see the case if permissions allow.
- Location transfer events are logged.
- Reports can optionally use original location or current owner location.

---

## Epic 2.9: SLAs, Reminders, and Notifications

### Goal
Ensure recovery cases receive timely attention and give operators visibility into response discipline.

### User Stories

#### Story 2.9.1: Define recovery SLAs
As a brand admin, I want to define response-time expectations so that teams know how quickly to respond.

Acceptance criteria:
- Admins can define SLA targets by priority, issue category, location, region, source, channel, and business hours.
- SLA targets can include time to first response, time to assignment, time to resolution, and time since last guest reply.
- Default SLA rules are provided.
- SLA rules can be enabled, disabled, and versioned.

#### Story 2.9.2: Track SLA state
As a manager, I want to see whether a case is within SLA, at risk, or overdue so that I can act before it gets worse.

Acceptance criteria:
- Cases display SLA state.
- SLA state updates as time passes and case activity changes.
- SLA timers respect configured business hours when enabled.
- SLA state is filterable in the inbox.
- SLA calculations are auditable.

#### Story 2.9.3: Send recovery notifications
As a manager, I want notifications for cases that need attention so that I do not have to stare at the inbox.

Acceptance criteria:
- Notifications can be triggered by new case, new guest reply, assignment, mention, escalation, SLA at risk, SLA overdue, approval request, or case reopened.
- Notification channels are abstracted and can include email, SMS, push, Slack, Teams, webhook, or in-app notifications in future implementations.
- Users can configure notification preferences where allowed.
- Notifications include enough context to act but avoid exposing sensitive data unnecessarily.
- Notification failures are logged.

#### Story 2.9.4: Remind assignees about stale cases
As a brand operator, I want assignees reminded about stale cases so that cases do not sit unresolved.

Acceptance criteria:
- Admins can configure stale-case reminder rules.
- Reminders can trigger based on no first response, no update, waiting on internal owner, or no resolution.
- Reminders can be sent to assignee, manager, location owner, or escalation team.
- Reminder events are logged.

---

## Epic 2.9A: Recovery Case State Machine and Concurrency Controls

### Goal
Prevent invalid recovery lifecycle transitions, race conditions, and conflicting updates across managers, automations, and AI-assisted workflows.

### User Stories

#### Story 2.9A.1: Enforce valid recovery state transitions
As a platform operator, I want recovery cases to follow a strict lifecycle so that invalid transitions cannot corrupt workflow state.

Acceptance criteria:
- Recovery cases use a formal state machine.
- Allowed transitions are explicitly defined.
- Invalid transitions are rejected.
- Closed or archived cases cannot receive invalid automated updates.
- Reopen behavior follows configured transition rules.
- State transition failures are observable and audited.

#### Story 2.9A.2: Prevent conflicting concurrent updates
As a restaurant manager, I want conflicting updates prevented so that multiple actors do not overwrite each other accidentally.

Acceptance criteria:
- Recovery cases use optimistic locking or equivalent version control.
- Conflicting updates return safe retryable errors.
- Bulk updates validate record versions where appropriate.
- Automation workflows respect version conflicts.
- Timeline and audit history preserve attempted conflicting actions when appropriate.

---

## Epic 2.10: Case Status, Resolution, Closure, and Reopening

### Goal
Give teams a consistent recovery lifecycle from new issue to resolved guest outcome.

### User Stories

#### Story 2.10.1: Manage case status
As a manager, I want to update a case’s status so that the team knows where it stands.

Acceptance criteria:
- Supported default statuses include new, open, assigned, waiting on guest, waiting on internal action, offer pending, resolved, closed, spam, duplicate, and archived.
- Status names can be configurable by brand where appropriate.
- Status changes are logged.
- Status changes can trigger notifications, SLA changes, and reporting updates.

#### Story 2.10.2: Mark a case resolved
As a manager, I want to mark a case resolved when the guest issue has been addressed so that the queue stays clean.

Acceptance criteria:
- Resolved status can require a resolution reason.
- Resolution reasons can include apologized, offer sent, refund issued, issue explained, guest satisfied, no response, invalid complaint, duplicate, spam, or custom reason.
- Required playbook steps must be complete before resolution when configured.
- Resolution timestamp and resolver are stored.

#### Story 2.10.3: Close a case
As a manager, I want to close a case after resolution so that it is removed from active work.

Acceptance criteria:
- Closed cases are hidden from default active views.
- Closure can require resolution reason, internal note, or manager approval based on configuration.
- Closure is logged.
- Closed cases remain searchable and reportable.

#### Story 2.10.4: Reopen a case
As a manager, I want cases to reopen when a guest replies or when new information appears so that unresolved issues are not ignored.

Acceptance criteria:
- A closed or resolved case can reopen automatically when a guest replies within a configured window.
- Authorized users can manually reopen cases.
- Reopening restores the case to an active status.
- Reopening events are logged.
- Reopened cases can trigger notifications and SLA timers.

#### Story 2.10.5: Capture guest outcome
As a brand operator, I want to capture whether recovery worked so that we can improve retention and operations.

Acceptance criteria:
- Managers can record outcome such as saved guest, guest satisfied, guest unresolved, no response, guest declined offer, guest returned, or unknown.
- Outcome can be updated manually or inferred from future orders, offer redemption, or guest response when integrations support it.
- Outcome is reportable by location, category, priority, assignee, and recovery action.

---

## Epic 2.11: Recovery Rules Engine

### Goal
Allow brands to configure how cases are created, prioritized, assigned, escalated, notified, and closed without custom development.

### User Stories

#### Story 2.11.1: Create recovery automation rules
As a brand admin, I want to create recovery automation rules so that case workflows match our operating model.

Acceptance criteria:
- Admins can create rules with triggers, conditions, and actions.
- Supported triggers include feedback submitted, review imported, case created, case updated, message received, SLA at risk, SLA overdue, offer issued, and case closed.
- Supported conditions include location, region, source, channel, rating, sentiment, category, keywords, guest segment, order value, case age, priority, and status.
- Supported actions include create case, set priority, assign, escalate, notify, tag, recommend template, recommend offer, and require playbook.
- Rules can be enabled, disabled, tested, and versioned.

#### Story 2.11.2: Test recovery rules
As a brand admin, I want to test rules before activating them so that I do not accidentally spam managers or misroute cases.

Acceptance criteria:
- Admins can run a rule against sample or historical cases.
- The system shows what actions would have occurred.
- Test runs do not modify production data unless explicitly applied.
- Rule validation identifies invalid conditions, missing users, missing teams, or conflicting actions.

#### Story 2.11.3: Audit rule execution
As a platform operator, I want to know which rules acted on a case so that automation is explainable.

Acceptance criteria:
- Rule executions are logged with rule ID, version, trigger, matched conditions, actions taken, timestamp, and outcome.
- Case timelines can show relevant automation events.
- Failed rule actions are logged and retryable where appropriate.

---

## Epic 2.12: Recovery Reporting and Performance Metrics

### Goal
Measure whether restaurant teams respond quickly, resolve issues well, recover guests, and reduce repeat problems.

### User Stories

#### Story 2.12.1: View recovery performance dashboard
As a brand operator, I want a dashboard of recovery performance so that I can understand how well teams are handling guest issues.

Acceptance criteria:
- Dashboard includes open cases, new cases, closed cases, overdue cases, average first response time, median first response time, average resolution time, SLA compliance, recovery rate, offer redemption rate, reopened cases, and guest no-response rate.
- Metrics can be filtered by date range, brand, region, location, category, source, channel, priority, assignee, and case status.
- Metrics support location comparison.
- Dashboard handles small data volumes gracefully.

#### Story 2.12.2: Measure response speed
As an operator, I want to measure response speed so that I can improve manager follow-up discipline.

Acceptance criteria:
- The platform calculates time to first human response.
- The platform can distinguish automated messages from human messages.
- Response speed can be reported by location, assignee, category, priority, source, and channel.
- Outliers can be inspected.

#### Story 2.12.3: Measure resolution effectiveness
As an operator, I want to measure whether cases are resolved effectively so that recovery is not just fast but useful.

Acceptance criteria:
- The platform tracks resolution reason, outcome, offer issued, offer redeemed, guest replied, guest sentiment after reply, return visit, and reopened status when available.
- Reports can compare outcomes by action type.
- Reports can identify issue categories with poor recovery outcomes.

#### Story 2.12.4: Identify repeat issues
As a brand operator, I want to see recurring recovery categories so that I can fix root causes operationally.

Acceptance criteria:
- Reports show top recovery categories by count, severity, trend, location, source, and channel.
- Reports distinguish one-off issues from recurring issues.
- Reports can show changes over time.
- Reports can link back to example cases.

#### Story 2.12.5: Export recovery data
As a brand admin, I want to export recovery data so that I can analyze it outside the platform.

Acceptance criteria:
- Admins can export cases, messages, notes, statuses, offers, outcomes, and timeline events based on permissions.
- Exports can be filtered by date, location, category, status, priority, source, channel, and assignee.
- Sensitive fields are permission-controlled.
- Export activity is logged.

---

## Epic 2.13: Recovery Permissions, Privacy, and Auditability

### Goal
Protect guest data, enforce proper access, and maintain a reliable audit trail for recovery actions.

### User Stories

#### Story 2.13.1: Enforce role-based access to recovery cases
As a brand operator, I want recovery access controlled by role and location so that guest information is only visible to authorized users.

Acceptance criteria:
- Permissions control who can view, create, assign, message, issue offers, approve offers, close, export, merge, delete, or redact cases.
- Permissions can be scoped by brand, region, location, team, and role.
- Users cannot access cases outside their permitted tenant or location scope.
- Permission denials are safe and logged where appropriate.

#### Story 2.13.2: Protect sensitive guest information
As a guest, I want my contact information and complaint details protected so that my data is not misused.

Acceptance criteria:
- Phone, email, name, order context, refund references, and conversation content are protected by permission controls.
- Sensitive fields can be masked in inbox views.
- Exports respect masking and permission rules.
- Data retention settings can be configured.

#### Story 2.13.3: Maintain recovery audit log
As a platform operator, I want every important recovery action audited so that actions are traceable.

Acceptance criteria:
- Audit log includes case creation, assignment, status changes, priority changes, notes, messages, offers, approvals, escalations, merges, exports, redactions, and deletions.
- Audit entries include actor, timestamp, tenant, location, action, and relevant metadata.
- Audit logs are immutable except for compliant retention/redaction workflows.
- Audit logs can be exported by authorized users.

#### Story 2.13.4: Redact guest data when required
As a privacy admin, I want to redact guest data when legally or operationally required so that we can honor privacy obligations.

Acceptance criteria:
- Authorized users can redact guest contact information and message content according to configured policies.
- Redaction preserves operational metrics where possible.
- Redacted data is not recoverable through normal application access.
- Redaction events are audited without preserving redacted sensitive content.

---

## Epic 2.14: Recovery APIs and Webhooks

### Goal
Expose recovery workflows through stable APIs and event hooks so the open-source ecosystem can integrate recovery with POS, CRM, loyalty, BI, call centers, customer support, and custom restaurant tools.

### User Stories

#### Story 2.14.1: Create recovery cases by API
As an integration developer, I want to create recovery cases through an API so that external systems can push guest issues into the platform.

Acceptance criteria:
- The API can create a case with source, guest, location, issue category, description, priority, metadata, and linked records.
- The API supports idempotency keys.
- The API validates tenant boundaries and permissions.
- API-created cases trigger the same automation rules as UI-created cases unless disabled.
- API errors are clear and machine-readable.

#### Story 2.14.2: Retrieve recovery cases by API
As an integration developer, I want to retrieve recovery cases so that external systems can display or sync recovery work.

Acceptance criteria:
- The API supports listing and retrieving cases.
- The API supports filtering by location, status, priority, category, source, assignee, guest, linked feedback, linked review, and date range.
- The API supports pagination.
- Sensitive fields are omitted unless the caller has permission.

#### Story 2.14.3: Update recovery cases by API
As an integration developer, I want to update recovery cases so that external workflows can participate in case management.

Acceptance criteria:
- The API can update status, priority, assignment, tags, internal notes, outcome, and linked metadata where permitted.
- Updates are audited.
- Invalid state transitions are rejected.
- API updates trigger configured rules and webhooks when appropriate.

#### Story 2.14.4: Send recovery messages by API
As an integration developer, I want to send recovery messages through the platform so that external tools can use the same compliance, logging, and conversation infrastructure.

Acceptance criteria:
- The API can send SMS or email messages from an authorized case context.
- The API enforces opt-out, consent, permission, and tenant rules.
- Sent messages are logged in the case conversation.
- Failed messages return provider-aware error details.

#### Story 2.14.5: Subscribe to recovery webhooks
As an integration developer, I want webhooks for recovery events so that other systems can react in real time.

Acceptance criteria:
- Webhooks support case created, case updated, case assigned, case escalated, message received, message sent, offer issued, offer redeemed, SLA at risk, SLA overdue, case resolved, case closed, and case reopened.
- Webhook payloads are signed.
- Delivery attempts are logged.
- Failed deliveries are retried.
- Webhook subscriptions are scoped by tenant and permissions.

---

## Epic 2.15: Open-Source Readiness for Guest Recovery

### Goal
Design Guest Recovery as a clean, extensible open-source module that can run locally, self-host, or operate as part of a managed SaaS offering.

### User Stories

#### Story 2.15.1: Provide clear recovery module boundaries
As an open-source contributor, I want recovery concerns separated clearly so that I can understand and modify the system safely.

Acceptance criteria:
- Recovery case lifecycle, messaging, templates, offers, assignment, SLA, rules, reporting, permissions, and APIs are organized into clear packages or modules.
- Each module has a documented purpose.
- Cross-module contracts are explicit and versioned.
- Recovery depends on feedback and guest identity through stable interfaces, not tight coupling to implementation details.

#### Story 2.15.2: Abstract communication providers
As a self-hosting operator, I want SMS and email providers abstracted so that I can use my preferred vendors or local fakes.

Acceptance criteria:
- SMS messaging uses a provider interface.
- Email messaging uses a provider interface.
- Local fake providers are available for development and tests.
- Provider failures are represented consistently.
- Provider-specific metadata can be stored without polluting core domain models.

#### Story 2.15.3: Abstract AI providers
As a self-hosting operator, I want AI features abstracted so that I can use OpenAI, Azure OpenAI, local models, or no AI at all.

Acceptance criteria:
- AI reply generation and summarization use provider interfaces.
- AI can be disabled without breaking recovery workflows.
- Local fake AI providers are available for development and tests.
- AI prompts and guardrails are configurable and versioned.
- AI-generated content is never automatically sent unless a future explicit automation feature is enabled.

#### Story 2.15.4: Provide seed recovery data
As an open-source developer, I want realistic seed data so that I can test recovery workflows locally.

Acceptance criteria:
- Seed data includes sample brands, locations, guests, feedback, recovery cases, messages, templates, offers, SLAs, rules, and users.
- Seed cases include positive, neutral, negative, urgent, duplicate, reopened, and closed examples.
- Seed data avoids real personal information.

#### Story 2.15.5: Define recovery test expectations
As a maintainer, I want clear test expectations so that recovery behavior remains stable.

Acceptance criteria:
- Automated tests cover case creation, duplicate detection, assignment, status transitions, messaging, opt-out enforcement, offer approval, SLA calculations, rule execution, permissions, and webhooks.
- Tests include multi-location and multi-tenant isolation scenarios.
- Provider interfaces have contract tests.
- Critical guest messaging paths have failure-mode tests.

---

# Guest Recovery MVP Cut

The MVP for Guest Recovery should include:

1. Automatic case creation from negative feedback.
2. Manual case creation.
3. Recovery inbox with filtering and search.
4. Case detail view with feedback, guest, location, source, and optional Experience context.
5. Internal notes.
6. Outbound recovery messaging through abstract SMS and email provider interfaces.
7. Guest-accessible recovery link for continued conversation.
8. Simple AI-suggested reply provider interface with fake local provider.
9. Response templates.
10. Basic offer recording without POS redemption integration.
11. Assignment to a user.
12. Strict recovery case state machine.
13. Optimistic locking on case updates.
14. Status lifecycle from new to closed.
15. Simple SLA tracking for time to first response.
16. Recovery created/updated webhooks.
17. Role and location-based permissions.
18. Audit log for core case actions.
19. Basic recovery performance dashboard with core metrics.

Defer from MVP:

1. Public review-created recovery cases.
2. Advanced inbound multi-channel message routing.
3. Advanced duplicate merging.
4. Team queues and complex routing.
5. Approval workflows for offers.
6. POS-backed coupon issuance and redemption.
7. Refund execution integrations.
8. Advanced automation rules engine.
9. Full playbook enforcement.
10. Advanced AI case summarization and multi-variant replies.
11. Slack, Teams, push, and advanced notification channels.
12. Complex business-hour SLA calendars.
13. Deep recovery outcome inference from future guest behavior.
14. Structured order-item, table, server, and ticket enrichment.

---

# Product Principles for Guest Recovery

1. Recovery must be fast enough to save the guest before frustration becomes permanent.
2. Managers should never have to hunt across systems to understand what happened.
3. AI should help humans respond better, not impersonate them without control.
4. Guest-facing messages must be logged, permissioned, and compliant.
5. Internal notes must never leak to guests.
6. Offers should be generous enough to recover trust but governed enough to prevent abuse.
7. Every case should have a clear owner, status, and next step.
8. SLA metrics should drive better behavior, not create fake activity.
9. Multi-location access control must be designed from the start.
10. The system should be usable without paid providers in local open-source development.

# Frictionless Feedback MVP Cut

The MVP for Frictionless Feedback should include:

1. Mobile-friendly hosted feedback page.
2. Static and manually generated feedback links and QR codes.
3. Opaque QR/link tokens.
4. One default short-form feedback flow.
5. Configurable rating scale, open-text feedback, categories, and optional contact fields.
6. Basic consent capture for follow-up.
7. Location and source attribution.
8. Optional Experience attachment using tenant_id, location_id, source_system, and source_entity_id when available.
9. Feedback storage and admin inbox.
10. Feedback detail view.
11. Basic status management.
12. Feedback-submitted event written through the transactional outbox.
13. Negative feedback recovery flag.
14. CSV export.
15. Public APIs for session creation, submission, and retrieval.
16. Local development mode with fake SMS/email providers.

Defer from MVP:

1. Automated post-order integrations.
2. Full long-form survey builder.
3. Advanced conditional survey logic.
4. Translation.
5. Deep analytics.
6. Review prompting delivery.
7. Full guest CRM.
8. Advanced anti-spam.
9. Multi-provider production SMS/email configuration UI.
10. Embedded widget variants beyond simple links.
11. Structured table, server, item, subtotal, and floor-plan analytics.
12. POS-specific enrichment beyond minimal Experience identity.

---

# Product Principles for This Epic Set

1. The guest experience must be faster than leaving a public review.
2. The platform must collect enough context to be useful without making guests work.
3. Anonymous feedback must remain possible where restaurants allow it.
4. Consent must be explicit, auditable, and separated by purpose.
5. Feedback should be operationally actionable, not just stored.
6. Every feedback submission should be event-driven so future modules can react.
7. Multi-location support should be designed from the beginning.
8. Self-hosted and local development modes must not require paid vendors.
9. Open-source contributors should be able to extend the system without rewriting the core.
10. The system should avoid review-gating patterns and support compliant reputation workflows later.
