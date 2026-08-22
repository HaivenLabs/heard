# The Haiven Constitution

This document defines the shared product, engineering, architecture, quality, security, design, and AI-agent contribution rules for every Haiven product.

These are not preferences.

They are the baseline operating rules for the Haiven ecosystem.

Product repositories may add stricter product-specific rules, but they may not weaken these rules.

Heard, qurl, Passage, and every future Haiven product must be built as part of one coherent ecosystem, not as random one-off apps.

---

## 1. Ecosystem-first product development

Every Haiven product must be designed as part of the broader Haiven ecosystem.

Products should have clear boundaries, but they must not duplicate shared platform capabilities.

Required behavior:

- Shared identity belongs to Passage.
- Shared QR generation belongs to qurl.
- Shared contracts should be reused where practical.
- Shared design language should be reused where practical.
- Product-specific code should stay product-specific.
- Shared services should expose clean APIs and SDKs.
- Products should integrate through contracts, not private implementation details.

Definition of done:

- Product ownership is clear.
- Shared capabilities are not reimplemented inside product repos.
- Integration contracts are documented.
- Product-specific exceptions are documented.

---

## 2. Shared platform capabilities must be centralized

Haiven products must not repeatedly implement common platform capabilities.

If a capability is needed by more than one product, or is likely to become a reusable ecosystem primitive, it must be evaluated as a shared Haiven service before being implemented product-locally.

Shared platform candidates include:

- registration
- authentication
- identity
- account management
- tenant or workspace membership
- roles and permissions
- billing
- subscriptions
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

Product repos may create temporary local implementations only when:

- the shared service does not exist yet
- the implementation is explicitly marked as temporary
- the implementation is wrapped behind a provider/client interface
- the contract matches the future shared service
- migration to the shared service is documented

No product may create a permanent local replacement for an existing Haiven shared service without a documented architecture exception.

Definition of done:

- Shared-service evaluation is documented for reusable capabilities.
- Temporary local implementations are behind interfaces.
- Migration path to the shared service is documented.
- Existing shared services are consumed instead of duplicated.

---

## 3. API-first

Every major capability must be accessible through stable APIs before it is tightly coupled to a UI.

The product UI, admin surfaces, mobile apps, integrations, automations, CLI tools, and future AI agents should operate on the same APIs.

Requirements:

- Public APIs use OpenAPI specifications unless a documented exception exists.
- OpenAPI contracts are the source of truth.
- Code generation is required where practical.
- Generated types, validation, clients, and server stubs should be derived from contracts.
- AI agents must modify contracts before implementation code when changing APIs.
- Internal service APIs use explicit contracts.
- APIs must support versioning.
- APIs must support pagination where listing is supported.
- APIs must support idempotency where duplicate creation or retries are plausible.
- APIs must return structured machine-readable errors.
- APIs must enforce tenant isolation and permissions.
- APIs must never expose unsafe internal implementation details.

Definition of done:

- OpenAPI spec updated.
- Generated code refreshed where practical.
- Contract compatibility validated.
- Contract tests included.
- API examples documented.
- Authentication and authorization enforced.
- Failure modes documented.

---

## 4. Contract-first

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
- open-source extensions

Requirements:

- All major entities use explicit versioned schemas.
- Schemas follow JSON Schema standards unless a documented exception exists.
- Breaking changes require schema version changes.
- Shared contracts must live in a dedicated contracts package or clearly named contract location.
- Arbitrary metadata must be intentionally scoped.
- Event payloads must be documented.
- Provider payload mappings must be documented.

Definition of done:

- Schema added or updated.
- Schema version documented.
- Validation tests included.
- Migration strategy documented where needed.

---

## 5. Modular monolith first

Haiven products should start as modular monoliths unless there is a clear operational reason not to.

The codebase must still use clear domain boundaries, contracts, events, providers, and package separation, but the first production system should be simple to run, debug, and self-host.

Preferred initial deployment shape:

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

- Local development runs with a small number of processes.
- Module boundaries are documented.
- Future service extraction points are clear but not prematurely implemented.

---

## 6. Test-first development

Tests are a first-class deliverable.

AI agents and human contributors must not treat tests as optional cleanup work.

Requirements:

- Unit tests required for domain logic.
- Integration tests required for APIs.
- Contract tests required for providers and generated clients.
- Permission tests required for protected workflows.
- Multi-tenant isolation tests required for tenant-scoped products.
- Event emission tests required for event-driven behavior.
- Failure-mode tests required for critical workflows.
- Regression tests required for bug fixes.
- Critical user flows should be covered end-to-end.

Definition of done:

- Required tests pass.
- CI validates tests automatically.
- New functionality cannot merge without relevant tests.
- Important failure cases are tested, not merely happy paths.

---

## 7. Behavior-driven workflow validation

BDD-style scenarios should primarily drive integration and end-to-end workflow validation.

Unit tests should continue handling deterministic domain logic, validation, transformation, utility behavior, and isolated business rules.

BDD should focus on:

- multi-step workflows
- permission-sensitive operations
- event-driven orchestration
- guest or customer interaction flows
- account and registration flows
- provider integration behavior
- automation behavior

Preferred structure:

```gherkin
Given a known system state
When the user or system performs an action
Then the expected business outcome occurs
And the expected events, audit entries, and permissions are enforced
```

Requirements:

- High-value workflows include behavioral scenarios.
- Acceptance criteria must be testable.
- Behavioral tests should cover business-critical flows.
- Avoid heavyweight BDD for simple deterministic utility logic.

---

## 8. Event-driven architecture

Haiven products should communicate internally through explicit events instead of hidden coupling where the workflow benefits from asynchronous orchestration, auditability, retries, or downstream integrations.

Requirements:

- Events must be versioned.
- Events must be idempotent or safely de-duplicated.
- Events must support retries.
- Event contracts must be documented.
- Failed event delivery must be observable.
- Event consumers must be resilient to duplicate delivery.
- Event payloads must avoid unnecessary sensitive data.

Preferred infrastructure:

- Internal domain event dispatch plus transactional outbox initially.
- NATS adapter when async scale or deployment topology justifies it.
- Kafka-compatible future path only if enterprise scale requires it.

Definition of done:

- Event contract documented.
- Event emitted in the correct transaction or workflow boundary.
- Event tests included.
- Retry behavior validated.
- Failure behavior observable.

---

## 9. Transactional outbox and consistency

Haiven products must guarantee consistency between durable database state changes and emitted domain events when events are used to drive downstream behavior.

Preferred pattern:

- Transactional outbox.

Requirements:

- Domain events are persisted in the same transaction as the originating state change.
- Background workers publish queued outbox events.
- Event publication must support retries.
- Event publication failures must be observable.
- Event consumers must support idempotency.
- Outbox replay tooling must exist for production systems.
- Outbox records must include enough metadata for debugging and safe replay.

Definition of done:

- State change and outbox insert occur atomically.
- Event publication is observable.
- Replay and retry behavior are tested.

---

## 10. Multi-tenant by default where applicable

Every product that supports organizations, businesses, teams, customers, workspaces, brands, or accounts must be tenant-aware from the beginning.

Tenant isolation cannot be retrofitted later without risk.

Requirements:

- Tenant-scoped records include tenant identifiers.
- Authorization checks enforce tenant boundaries.
- Cross-tenant access is prohibited by default.
- Queries must be tenant-scoped.
- Exports must be tenant-scoped.
- Audit logs must preserve tenant context.
- Background workers and async jobs must preserve tenant context.
- Cross-tenant queries require explicit privileged service behavior.

Preferred strategy:

- PostgreSQL Row-Level Security for tenant-scoped production tables where practical.
- Application-level tenant context enforcement.
- Repository/query-layer tenant injection.
- Service-level authorization checks.

Definition of done:

- Tenant isolation tests included.
- Permission tests included.
- Unsafe queries blocked by linting, review, or repository patterns.
- RLS coverage documented where used.

---

## 11. Canonical identity and identifiers

Core platform identifiers must use UUIDs unless there is a deliberate, documented exception.

Requirements:

- Tenant IDs must be UUIDs.
- Internal entity IDs should be UUIDs.
- External provider IDs must be stored separately from internal IDs.
- External IDs must never be treated as globally unique without source-system and tenant/location/context.
- Database primary keys should use UUIDs for tenant-scoped domain entities.
- Public URLs should use opaque tokens or slugs instead of exposing raw UUIDs when appropriate.

Definition of done:

- Internal ID fields use UUIDs unless documented otherwise.
- External ID mappings are stored separately.
- ID semantics are documented in schemas.

---

## 12. Passage owns shared identity

Passage is the shared Haiven registration, authentication, identity, account, session, tenant, role, and permission platform.

Haiven products must not each invent their own auth stack.

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

Product repos own:

- product-specific authorization decisions
- product-specific permission checks
- the complete branded registration, login, recovery, and account-management experience presented to their customers
- product-specific user experience after identity is resolved
- product-specific account onboarding where Passage delegates context

Requirements:

- Production admin surfaces require authentication.
- Passage is invisible infrastructure to consumer-product users. Consumer-facing copy, controls, errors, URLs, and screens must use the consumer product's brand and provider-neutral language unless a legal or compliance requirement explicitly requires disclosure.
- Consumer applications must call provider-neutral identity interfaces and routes; Passage-specific names belong only in adapter implementations, internal contracts, configuration, and operator documentation.
- Third-party identity providers should be pluggable through provider abstractions.
- Role and tenant scoping must be enforced after authentication.
- Local development may use a clearly labeled auth bypass only in non-production environments.
- Product repos must integrate with Passage through documented contracts.

Definition of done:

- Product auth paths use Passage or a temporary documented adapter.
- Registration, login, recovery, and account-management flows remain visually and verbally owned by the consumer product.
- Automated tests prevent shared-service or provider names from leaking into consumer-facing auth surfaces.
- Role and tenant checks are enforced for authenticated requests.
- Local auth bypass cannot run in production mode.
- Product-specific permissions are tested.

---

## 13. qurl owns shared QR generation

qurl is the shared QR generation product in the Haiven ecosystem.

Haiven products must not implement their own QR rendering engines unless a temporary local fake is required for tests or offline development.

qurl owns:

- QR code rendering
- QR export formats
- QR styling
- QR asset generation
- QR payload encoding rules
- QR design studio
- QR templates
- hosted dynamic QR behavior when explicitly selected
- scan analytics when explicitly selected

Product repos own:

- product context
- destination URL creation
- opaque token generation
- permissions
- product-specific analytics
- product workflow tied to the destination URL

Integration rule:

- The product generates the destination URL.
- The product sends that destination URL to qurl.
- qurl returns QR assets or downloadable QR representations.
- The product stores the qurl asset reference if needed.
- The product should not send unnecessary tenant, customer, guest, order, or operational metadata to qurl.

Direct QR rule:

If a user creates a direct QR code for `https://example.com`, the QR code points to `https://example.com`.

No surprise short links.

No forced redirect domains.

No tracking unless the user explicitly chooses it.

Definition of done:

- Product has a qurl provider/client interface.
- Local development has a fake qurl provider if qurl is unavailable.
- Production configuration can point to a real qurl API.
- QR generation tests verify minimal data sharing.
- Product does not add a separate QR rendering library except for local fakes or tests.

---

## 14. Provider abstraction

Third-party providers must be replaceable.

Requirements:

- Identity providers use interfaces.
- SMS providers use interfaces.
- Email providers use interfaces.
- AI providers use interfaces.
- Storage providers use interfaces.
- Translation providers use interfaces.
- Payment providers use interfaces where applicable.
- POS integrations use adapters where applicable.
- Fake local providers exist for development and tests.
- Provider-specific behavior must not leak into core domain logic.

Definition of done:

- Provider interface documented.
- Fake provider included where practical.
- Contract tests included.
- Provider failure behavior documented.

---

## 15. Schema migration strategy

Schema evolution must be explicit, versioned, testable, and safe for self-hosted deployments.

Preferred tooling:

- Atlas or golang-migrate for Go-backed services.
- Equivalent versioned migration tooling for other stacks.

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
- Destructive behavior explicitly called out.

---

## 16. Scalability and performance

Haiven products should be designed to scale without prematurely overcomplicating the first usable slice.

Requirements:

- Stateless application services where practical.
- Horizontally scalable workers.
- Async processing where appropriate.
- Queue-backed provider delivery where appropriate.
- Efficient indexing strategies.
- Pagination for large result sets.
- Rate limiting.
- Idempotent APIs.
- Background job retries.
- Bulk operation safeguards.
- Timeouts on network calls.
- Backpressure handling for high-volume workflows.

Definition of done:

- Performance considerations documented.
- Expensive queries identified.
- Backpressure handling defined.
- Timeouts and retries configured.
- Critical list APIs support pagination.

---

## 17. Reliability and failure recovery

The platform must fail gracefully.

Users should not lose important work because a provider, queue, browser, export, or integration is temporarily unavailable.

Requirements:

- Transactional outbox where appropriate.
- Retry with exponential backoff.
- Dead-letter handling.
- Graceful provider degradation.
- Circuit breakers for unstable providers.
- Safe retry semantics.
- Idempotent event handling.
- Recovery jobs for failed deliveries.
- Clear user-facing failure messages.
- Observable internal failure state.

Definition of done:

- Failure modes documented.
- Retry behavior tested.
- Recovery path validated.
- Provider outages handled safely.

---

## 18. Maintainability and readability

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
- Functions and methods should be named so a bot with zero prior context can understand the intent.

Definition of done:

- Module responsibilities documented.
- Public interfaces documented.
- Complex logic explained.
- Dead code avoided.
- Generated code does not obscure product intent.

---

## 19. Observability by default

Every critical workflow must be observable.

Requirements:

- Structured logs.
- OpenTelemetry tracing.
- Correlation IDs.
- Metrics for critical workflows.
- Health checks.
- Readiness checks.
- Provider delivery logs.
- Audit logs where applicable.
- Event processing metrics.
- Queue health metrics where applicable.
- SLA or workflow latency monitoring where applicable.

Definition of done:

- Logs added.
- Metrics added.
- Traces emitted or hooks included.
- Critical failures are searchable.
- Dashboards updated where applicable.

---

## 20. Security and privacy

Trust is a platform requirement.

Security is continuous, not a release-phase checklist. Every Haiven repository,
dependency, build, artifact, deployment, and shared service must remain secure by
default throughout its lifecycle. When a security control cannot make a safe
decision, it must fail closed.

Requirements:

- RBAC enforcement.
- Principle of least privilege.
- Secure secret handling.
- Signed webhooks where applicable.
- PII masking.
- Secure defaults.
- Rate limiting.
- Abuse protection.
- Sensitive data exclusion from logs.
- Audit logging for sensitive actions.
- Privacy retention configurable where applicable.
- Consent records stored where applicable.
- One supported JavaScript package manager and one committed lockfile per repository.
- Exact runtime and package-manager versions pinned in source and CI.
- Container base images pinned to reviewed immutable digests, with automated refresh and rescanning.
- Frozen, integrity-checked dependency installation in CI.
- Dependency lifecycle scripts denied by default and enabled only through a reviewed allowlist.
- Newly published dependency versions delayed long enough for ecosystem detection and removal of common malware campaigns.
- Package provenance and publisher-trust downgrades rejected where the package manager supports it.
- Transitive Git, direct-tarball, and other exotic dependency sources blocked by default.
- Runtime and development dependency graphs scanned for known vulnerabilities and malware before code executes with credentials.
- GitHub Actions pinned to immutable full commit SHAs.
- CI tokens explicitly least-privileged and checkout credentials not persisted.
- Untrusted pull-request code never executes with repository, package-registry, cloud, deployment, or self-hosted-runner credentials.
- Dependency review, static analysis, secret scanning, push protection, SBOM generation, and artifact provenance enabled where supported.
- Package publication uses short-lived OIDC trusted publishing and provenance; long-lived registry write tokens are prohibited when trusted publishing is available.
- Phishing-resistant MFA, protected branches, required review, and environment approval enforced for privileged repositories and releases.
- Security exceptions are narrowly scoped, owned, documented, time-bounded, and tested; they may never silently weaken the ecosystem baseline.

Definition of done:

- Permission checks tested.
- Sensitive fields protected.
- Security-sensitive flows audited.
- Secrets are not committed.
- Logs do not expose sensitive data.
- The complete install-time dependency graph passes the configured vulnerability and malware gates.
- Lockfiles, package-manager policy, CI action pins, token permissions, SBOMs, and provenance are validated automatically.
- Publishing and deployment credentials are short-lived and job-scoped where provider support exists.
- A documented containment, credential-revocation, dependency-rollback, and artifact-rebuild path exists for supply-chain incidents.

---

## 21. Compliance-aware communication

Messaging systems must distinguish between transactional and marketing communication.

Requirements:

- Transactional and marketing messaging are separated.
- SMS opt-outs are enforced.
- Email unsubscribe behavior is enforced where applicable.
- Consent records are stored.
- Review-gating is avoided.
- Communication history is auditable.
- Privacy retention is configurable where applicable.

Definition of done:

- Consent captured where needed.
- Opt-outs enforced.
- Communication audit entries recorded.
- Transactional and marketing flows are not mixed.

---

## 22. Unified responsive experience standard

Haiven products must ship one premium responsive interface system across guest-facing, customer-facing, and admin-facing surfaces instead of maintaining separate device-specific variants.

Requirements:

- UI work must be mobile-first and scale cleanly to tablet, laptop, and desktop layouts.
- Separate iOS, Android, tablet, or desktop variants are discouraged unless a platform constraint makes them unavoidable.
- Shared layout primitives, spacing rules, typography, and component patterns should support all supported screen sizes.
- Production UX must meet a high visual bar.
- Interfaces must be screenshot-ready from the first usable slice.
- Responsive behavior is part of the feature, not deferred polish.
- Production UX should not be conflated with internal test harness UX.
- Dev/test tools must live in clearly labeled internal testing surfaces.

Definition of done:

- Feature works on phone, tablet, and desktop breakpoints.
- Interface feels polished, not merely functional.
- No device class requires a separate bespoke implementation for the same workflow unless documented.

---

## 23. Design quality bar

Haiven products should feel sharp, useful, calm, modern, and trustworthy.

Requirements:

- No ugly default forms.
- No developer-demo interfaces in production paths.
- Useful empty states.
- Useful loading states.
- Useful error states.
- Clear interaction states.
- Accessible contrast and form labels.
- Screenshot-ready layout from the first usable slice.
- Clear hierarchy, spacing, and typography.
- Product UI should feel intentional even before it is feature-complete.

Definition of done:

- The feature is usable and polished.
- The interface is responsive.
- Empty, loading, error, and populated states exist where applicable.
- The design supports public demos or screenshots.

---

## 24. Open-source credibility

Haiven products should operate as legitimate open-source products unless explicitly private.

Requirements:

- Local development works without paid vendors.
- Seed data included where useful.
- Contributor docs included.
- CI works locally or has a clear local equivalent.
- Docker-first development where practical.
- Self-hosting supported where appropriate.
- Clear extension points.
- Stable contracts.
- Fake providers for local development.
- No hidden commercial dependency required for the first usable slice.

Definition of done:

- Local startup documented.
- Fake providers functional.
- Seed data validated where applicable.
- Contributor workflow documented.

---

## 25. Vertical slice delivery

Features should be built as complete end-to-end vertical slices.

Avoid:

- building all database tables first
- building all APIs first
- building all UIs first
- giant incomplete architectural layers
- disconnected scaffolding that cannot prove user value

Preferred approach:

- build one complete capability end-to-end
- validate operational behavior
- validate observability
- validate permissions
- validate scalability assumptions
- make the UI feel real

Definition of done for every slice:

- API implemented.
- Contract documented.
- Tests passing.
- Observability added.
- Permissions enforced.
- Events emitted where applicable.
- Local development supported.
- Docs updated.
- UI is usable and polished.

---

## 26. AI-agent contribution standards

Haiven is intentionally designed for agentic software development.

Every task given to Codex, Cursor, Claude, ChatGPT, or any other coding agent must preserve the architecture and quality bar.

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
- Generated code must not invent architecture that conflicts with this constitution.
- New durable directives that change product, UX, architecture, quality, or implementation expectations must be written back into the relevant docs instead of living only in chat context.
- When user feedback identifies UX confusion or workflow friction, the plan should be updated with concrete interaction requirements before subsequent implementation continues.

Preferred task structure:

- Goal
- User story
- Scope
- API contract
- Data model
- Events emitted
- Permissions
- Failure modes
- Observability
- Tests required
- Definition of done

---

## 27. Preferred default technical stack

This is the default stack unless a product-specific document chooses otherwise.

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

- TypeScript
- React or Next.js where appropriate
- Tailwind where appropriate
- Shared UI tokens where practical

Mobile:

- Expo React Native where mobile apps are needed

API:

- REST externally
- OpenAPI contracts
- Avoid gRPC until a domain is extracted into a separately deployed service

Testing:

- Unit tests
- Integration tests
- Contract tests
- BDD-style scenario tests for critical workflows

Infrastructure:

- Docker-first local development
- CI/CD from day one
- GitHub Actions initially

---

## 28. Engineering enforcement strategy

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
- required definition of done checklists
- product-specific implementation plans

---

## 29. Global definition of done

Every production-ready capability in a Haiven product must include:

- Working implementation
- API contract
- Schema validation
- Tests
- Permission enforcement
- Tenant isolation where applicable
- Event emission where applicable
- Logging
- Metrics
- Tracing or trace hooks
- Audit coverage where applicable
- Documentation
- Local development support
- Failure-mode handling
- CI validation
- Responsive, polished UX where user-facing

No feature is complete without these.
