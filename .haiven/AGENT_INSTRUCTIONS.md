# Agent Instructions

These instructions apply to Codex, Cursor, Claude, ChatGPT, and any other AI coding agent working on Haiven repositories.

Agents must follow the Haiven Constitution.

Do not optimize for speed by bypassing architecture.

Do not generate broad scaffolding that cannot be tested.

Do not silently invent product rules.

---

## Before writing code

Read:

1. `HAIVEN_CONSTITUTION.md`
2. `PRODUCT_REGISTRY.md`
3. `SHARED_SERVICES.md`
4. The product-specific docs for the target repo
5. The active slice or implementation plan

If the target repo has its own constitution or plan, follow the stricter rule.

If rules conflict, stop and document the conflict.

---

## Required task shape

Every implementation task should include:

- Goal
- User story
- Scope
- Non-goals
- API contract
- Data model
- Events emitted
- Permissions
- Failure modes
- Observability
- Tests required
- Definition of done

If the task does not include these, infer the smallest safe version and document the assumptions.

---

## Build vertical slices

Do not build disconnected layers.

A valid slice should include enough of the following to prove value:

- data model
- migration
- API contract
- API implementation
- domain logic
- permissions
- events where applicable
- tests
- UI where applicable
- observability
- documentation
- local development support

Avoid:

- building all tables first
- building all APIs first
- building all UIs first
- creating unused abstractions
- creating fake completeness

---

## Contract-first behavior

When changing an API:

1. Update OpenAPI first.
2. Update JSON Schema contracts where applicable.
3. Regenerate types/clients/stubs where practical.
4. Update implementation.
5. Update tests.
6. Update docs.

Do not add undocumented API behavior.

Do not return ad hoc error shapes.

Do not bypass generated types when generated types exist.

---

## Shared service rules

Use Passage for shared identity.

Use qurl for QR generation.

Shared-service ownership is an implementation boundary, not permission to expose the service as part of a consumer product's UX.

Consumer products must own their complete branded customer journey. Do not put Passage, qurl, provider, sandbox, adapter, hosted-service, or other infrastructure names in consumer-facing copy, controls, errors, screens, or URLs unless the user explicitly requires that disclosure. Use provider-neutral interfaces and routes in consumer code; keep provider names inside adapters, configuration, internal contracts, and operator documentation.

Do not implement product-local replacements for shared services unless:

- the shared service does not exist yet
- the local replacement is a fake/test adapter
- the adapter contract matches the intended shared service
- the temporary decision is documented

---

## Testing rules

Every meaningful change needs tests.

Required tests depend on the change, but may include:

- unit tests
- integration tests
- contract tests
- permission tests
- tenant isolation tests
- failure-mode tests
- event emission tests
- regression tests
- end-to-end tests for critical flows

Do not leave tests as TODOs.

If a test cannot be written yet, document exactly why and what must change.

---

## Observability rules

Critical workflows need:

- structured logs
- correlation IDs
- metrics
- tracing or trace hooks
- health/readiness checks where applicable
- provider failure visibility
- event failure visibility

Do not add invisible background behavior.

If a job can fail, operators need a way to know.

---

## Security rules

Never commit secrets.

Every Haiven repository must ignore real environment files using `.env`, `.env.*`, and `!.env.example` in `.gitignore`. Commit only sanitized placeholders in `.env.example`; never add an actual environment file to Git. If one was tracked, remove it from tracking and rotate any exposed credentials.

Never log sensitive data.

Never expose raw internal IDs in public URLs when opaque tokens or slugs are more appropriate.

Always enforce tenant and permission boundaries.

For browser identity work, model one canonical public origin per environment. Redirect aliases before issuing OAuth or session cookies; derive or validate callback, CORS, and branded return URLs against that origin; and test the local plus deployed-environment contract. Do not rely on the incoming Host header to choose an OAuth callback.

Local bypasses must be impossible to enable in production mode.

Treat software supply-chain security as part of every change. Preserve the
repository's pinned package manager, committed lockfile, lifecycle-script
allowlist, release-age and provenance policy, frozen CI installs, immutable
GitHub Action pins, least-privilege permissions, dependency gates, SBOMs, and
artifact provenance. Do not add a dependency, action, registry, install script,
publishing token, or CI secret exposure without reviewing its trust boundary.

Untrusted pull-request code must never run with secrets, write-capable tokens,
deployment credentials, package publication credentials, or self-hosted runners.
If a required tool cannot meet the baseline, document a narrow, owned,
time-bounded exception and a removal date; never weaken a control silently.

---

## UX rules

Do not ship developer-looking production UI.

User-facing and admin-facing surfaces need:

- responsive layouts
- polished spacing and typography
- useful empty states
- useful loading states
- useful error states
- accessible labels and contrast
- clear interaction states

Registration, login, recovery, and account-management screens must read as native parts of the consumer product, even when a shared Haiven service performs the underlying operation.

The first usable slice should be screenshot-ready.

---

## Documentation rules

Durable decisions must be written into docs.

If the user gives a new instruction that changes product, UX, architecture, quality, or implementation expectations, update one of:

- Haiven Constitution
- product-specific plan
- active slice doc
- API contract
- README
- architecture doc

Do not leave durable product decisions only in chat history.

---

## Commit behavior

Keep commits focused.

Prefer small, reviewable changes.

Commit messages should clearly state the product capability or standard changed.

Generated code should be separated from hand-written logic where practical.

---

## Done means done

A task is not complete unless:

- implementation works
- contracts are updated
- tests pass
- docs are updated
- local development still works
- failure modes are handled
- observability exists where needed
- UI is polished where user-facing
