# Local Development Standards

Haiven products should be easy to run locally.

## Principles

Local development should be:

- boring
- documented
- repeatable
- Docker-first where practical
- free of paid vendor requirements
- compatible with fake providers
- fast enough for daily use

## Required docs

Every product repo should document:

- prerequisites
- bootstrap command
- run command
- test command
- check/lint command
- database setup
- migration command
- seed data command where applicable
- fake provider behavior
- environment variables
- troubleshooting notes

## Fake providers

Use fake providers for local development when real providers are unavailable or expensive.

Fake providers should exist for:

- Passage/auth when Passage is unavailable
- qurl when product is not testing QR rendering itself
- SMS
- email
- AI
- storage
- payments
- POS integrations

Fake providers must:

- be clearly labeled
- not run in production mode
- preserve the real provider contract
- be predictable in tests

## Docker

Docker-first does not mean Docker-only.

Products should support:

- Docker Compose for full local runtime
- direct app/backend runs where useful
- clear commands for both paths

## Canonical browser origins

Every browser-facing product declares one canonical external application origin for each environment. This is an identity and deployment contract, not a UI preference.

| Environment | Canonical origin rule |
| --- | --- |
| Local | Use one documented loopback origin, such as `http://localhost:3010`; redirect aliases such as `127.0.0.1` before cookies are created. |
| Integration | Configure that environment's explicit HTTPS product hostname. |
| Staging | Configure that environment's explicit HTTPS product hostname. |
| Production | Configure the production HTTPS product hostname. |

The application must derive or reject drift among its public URL, CORS allow-origin, OAuth callback, branded recovery redirect, and cookie scope. OAuth providers must register the exact callback for each environment's canonical origin. Never derive an OAuth callback from an untrusted request host. Promote the same source through environments by changing deployment configuration and registered provider callbacks, not code.

## CI parity

Local checks should approximate CI.

Preferred commands:

```bash
pnpm run check
pnpm run test
pnpm run lint
```

or product-equivalent commands.

## Definition of done

Local development is complete only when:

- repo boots from documented commands
- migrations run from an empty database
- tests run locally
- fake providers work
- required environment variables are documented
- first-time contributor path is clear
