# ADR-0006: One API per consumer

- Status: Accepted
- Date: 2026-11-25

## Context

Browser sessions (cookie + CSRF), staff sessions and server-to-server API keys have
different authentication, CORS, rate-limit and stability needs. Mixing them in one
listener couples their policies and release cadence.

## Decision

- **`apps/web-api`** serves only `apps/web`: customer authentication, the session
  cookie, CSRF and every `/v1/...` operation the dashboard calls.
- **`apps/bo-api`** serves only `apps/bo`: independent staff identity and operations.
- **`apps/api`** is the public API for customers' servers: Bearer API keys only,
  never cookie sessions.
- Each is an independent Go Echo/Huma module and listener with its own config
  section, Dockerfile and committed OpenAPI schema (`contracts/openapi/*.json`).
  Only consumers own SDKs: `apps/web` generates from `web-api.json`, `apps/bo` from
  `bo-api.json`.
- APIs never import each other's `internal` packages or proxy one another.

## Alternatives considered

- **One API with route groups:** one deployable less, but session and API-key traffic
  share CORS policy, rate limits and release risk.

## Consequences

- Three backend deployables. A read needed by two APIs moves into `packages/application`.

## Revisit when

A second consumer needs session access to an existing API.
