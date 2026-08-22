# Security and Privacy Standards

Trust is a product requirement.

Security is always on. It applies to source, dependencies, developer installs,
CI, artifacts, publication, deployment, runtime behavior, and incident recovery.
Controls fail closed when trust cannot be established.

## Software supply chain

Every Haiven JavaScript repository must:

- use the ecosystem-pinned Node and pnpm versions
- pin container base images to reviewed immutable digests
- use one package manager and one committed lockfile
- install with a frozen lockfile in CI
- delay newly published dependency versions
- reject publisher-trust or provenance downgrades where supported
- block exotic transitive dependency sources
- deny dependency lifecycle/build scripts unless the package is explicitly reviewed and allowlisted
- audit the complete install-time graph, including development dependencies
- review dependency changes before merge
- generate a machine-readable SBOM for release artifacts
- retain enough release evidence to reproduce, investigate, and roll back a compromised build

Lockfile integrity proves that installed bytes match the reviewed lockfile. It
does not prove that the reviewed package version is benign, so release-age,
provenance, malware detection, review, and script isolation remain mandatory.

## CI and delivery

GitHub Actions must be pinned to immutable full commit SHAs. Workflows must set
explicit least-privilege permissions, disable persisted checkout credentials,
use bounded timeouts, and keep untrusted pull-request code away from secrets,
write tokens, deployment credentials, and self-hosted runners.

Container builds must use digest-pinned base images and emit SBOM and provenance
attestations when the target registry supports OCI referrers.

Do not use `pull_request_target` to build or execute pull-request code. A workflow
that genuinely needs elevated privileges must consume a previously reviewed,
immutable artifact and must not check out or execute attacker-controlled content.

Enable dependency review, CodeQL or equivalent static analysis, Dependabot
security and malware alerts, secret scanning, push protection, required status
checks, protected branches, and reviewed deployment environments wherever the
hosting provider supports them.

## Package publication

Publish packages through OIDC trusted publishing with provenance. Do not keep a
long-lived npm or other package-registry write token when trusted publishing is
available. Scope every package to its owning platform namespace and require
review of publication workflow changes.

Consumption through a registry proxy may add quarantine, retention, and rollback,
but it does not replace package-manager controls or credential isolation.

## Supply-chain incident response

For a suspected malicious dependency or workflow compromise:

1. Stop publication and deployment and isolate affected runners and workstations.
2. Revoke package, source-control, cloud, signing, and deployment credentials that may have been exposed.
3. Inspect workflow changes, repository persistence, new runners, releases, hooks, and package versions.
4. Restore a reviewed lockfile and rebuild on clean ephemeral infrastructure.
5. Verify SBOMs, checksums, attestations, and downstream consumers before resuming delivery.
6. Record the incident, affected artifacts, rotations, customer impact, and preventive control changes.

Security exceptions require an owner, rationale, exact scope, compensating
controls, expiry date, and removal issue. Expired exceptions fail the build.

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
- dependency and workflow policy checks pass
- the complete install-time graph passes vulnerability and malware gates
- release artifacts have an SBOM, checksums, and provenance where supported
- untrusted code cannot reach privileged credentials
- rollback and credential-revocation procedures remain actionable
