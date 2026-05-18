# Slice 1

## Goal

Deliver a Docker-runnable vertical slice that proves the loop:

1. Create tenant
2. Create location
3. Generate feedback link and QR payload
4. Start a hosted guest feedback session
5. Submit feedback
6. Persist feedback and transactional outbox event
7. Let the worker create a recovery case for negative feedback
8. Review and update the case from the manager inbox

## Scope

- Modular Go backend with one API process and one worker process
- PostgreSQL persistence and SQL migrations
- Next.js web app for guest and admin flows
- Fake qurl provider for local QR generation
- Transactional outbox with polling worker
- Audit entries for key actions
- OpenAPI contract for Slice 1
- One responsive UI system for guest and admin flows across phone, tablet, laptop, and desktop breakpoints

## Contracts

- OpenAPI: [openapi.slice1.yaml](/D:/dev/HaivenLabs/heard/docs/openapi.slice1.yaml)
- Event: `feedback-submitted` version `1`

## Data model

- `tenants`
- `locations`
- `experiences`
- `feedback_links`
- `feedback_sessions`
- `feedback_responses`
- `recovery_cases`
- `audit_entries`
- `outbox_events`

## Rules

- Ratings `1` and `2` map to `negative`
- Rating `3` maps to `neutral`
- Ratings `4` and `5` map to `positive`
- Negative feedback creates a recovery case in the worker
- UI changes must stay mobile-first, fully responsive, and premium across form factors without separate device-specific builds
- New user directives that change implementation expectations must be written back into the governing docs

## Local startup

1. Run `docker compose up --build`
2. Open `http://localhost:3010`
3. Use `/admin/campaigns` to create a flyer survey campaign, or use the seeded token `demo-heard`
4. Submit a low rating from `/f/demo-heard`
5. Open `/admin/recovery` to inspect the generated case

## Failure modes covered

- Invalid feedback link token returns `404`
- Duplicate feedback session submission is rejected
- Invalid recovery status transitions are rejected
- Worker keeps polling if outbox processing fails
- Tenant header mismatches are rejected for admin writes
