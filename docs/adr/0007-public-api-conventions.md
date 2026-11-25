# ADR-0007: Public API uses snake_case JSON; internal APIs stay camelCase

- Status: Accepted
- Date: 2026-11-25

## Context

`apps/api` is used by external developers; `apps/web-api` and `apps/bo-api` are used
only by our own apps through generated SDKs.

## Decision

- **Public contracts use snake_case** for JSON fields and enum values in
  `contracts/openapi/api.json`.
- **Internal APIs keep camelCase**.
- IDs are the schema UUIDs, without type prefixes. Timestamps are RFC 3339 strings.
- Errors are RFC 9457 problem+json with a stable snake_case `code` and never echo
  supplied values.
- Amounts, when a product adds them, are integer minor units with a `currency`.

## Alternatives considered

- **camelCase everywhere:** one convention, but unfamiliar for API consumers.
- **snake_case everywhere:** forces internal contracts and screens to change for no user benefit.

## Consequences

Each app owns its generated SDK, so the naming difference does not leak.

## Revisit when

An internal API becomes public.
