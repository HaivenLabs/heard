# Security Policy

Heard handles guest feedback, contact information, operational data, and communication history. Security issues should be reported responsibly.

## Supported Versions

Heard is early-stage software. Security fixes will target the active development branch until stable releases are established.

## Reporting a Vulnerability

Please do not open a public GitHub issue for a suspected vulnerability.

Report security concerns through the official maintainer contact channel.

Include:

- Description of the issue
- Steps to reproduce
- Affected component
- Potential impact
- Suggested fix, if known

## Security Expectations

Heard should follow these security principles:

- Tenant isolation by default
- Least-privilege access
- Role-based permissions
- Sensitive data masking
- No sensitive guest data in logs
- Signed webhooks
- Secure provider credential storage
- Rate limiting on public endpoints
- Abuse protection on guest-facing flows
- Audit logging for sensitive actions
- Safe handling of SMS and email opt-outs

## Sensitive Data

Do not commit:

- API keys
- Database credentials
- Provider credentials
- Private certificates
- Real guest data
- Real phone numbers or emails
- Production logs

Use fake local providers and seed data for development.

## Dependency Security

Dependencies should be reviewed before introduction.

Avoid adding large dependencies for small problems.

### Install-time supply-chain controls

Heard treats dependency installation as an untrusted execution boundary. The
web application belongs to the root pnpm workspace, uses the repository's one
committed lockfile, installs with `--frozen-lockfile`, and audits the complete
development and production dependency graph. pnpm dependency build scripts are
denied by default; any exception must be reviewed, documented, version-scoped,
time-bounded, and added to the explicit `allowBuilds` policy.

Pull requests must not weaken lockfile reproducibility, lifecycle-script
blocking, least-privilege workflow permissions, or immutable GitHub Action
references. Credentials must not be available to untrusted pull-request code.
