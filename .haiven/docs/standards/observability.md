# Observability Standards

Every critical Haiven workflow must be observable.

## Required signals

Critical workflows should include:

- structured logs
- metrics
- traces or trace hooks
- correlation IDs
- health checks
- readiness checks
- audit logs where applicable
- event processing metrics where applicable
- queue health metrics where applicable
- provider delivery logs where applicable

## Structured logs

Logs should include:

- correlation ID
- tenant ID where applicable
- user ID where safe and appropriate
- operation name
- outcome
- latency where useful
- error code
- provider name where applicable

Do not log sensitive data.

## Metrics

Useful metrics may include:

- request count
- request latency
- error count
- provider call count
- provider failure count
- queue depth
- event publish count
- event retry count
- job success/failure count
- workflow completion latency

## Tracing

Use OpenTelemetry where practical.

Trace important boundaries:

- API request
- database call
- provider call
- event publish
- job execution
- background worker step

## Health checks

Products should expose:

- liveness check
- readiness check
- dependency checks where practical

## Definition of done

Observability is complete only when:

- critical workflows log useful structured events
- errors include correlation IDs
- key metrics exist
- traces or hooks exist
- provider/job/event failures are visible
- sensitive data is not logged
