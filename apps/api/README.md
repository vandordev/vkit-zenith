# Public API

Independent Go Echo/Huma service for customers' servers (ADR-0006). It authenticates
with organization API keys (`Authorization: Bearer <key>`), never cookie sessions,
and publishes snake_case contracts (ADR-0007) in `contracts/openapi/api.json`.
`/docs` serves the Scalar reference; `/openapi.json` the schema. Regenerate the
embedded documentation favicon after the shared logo changes (`go generate`).

## Operations

| Operation | Path | Notes |
| --- | --- | --- |
| `getOrganization` | `GET /v1/organization` | The organization that owns the key. |

## Dummy phase (ADR-0001)

API keys are not persisted yet: `internal/handlers/v1/bearer_auth.go` resolves the
fixture keys below from `packages/fixtures`. They are fake values, never credentials.

| Key | Resolves to |
| --- | --- |
| `zk_demofixturekey` | Demo organization (active) |
| `zk_suspendedfixturekey` | Suspended demo organization |

Replacing the fixture lookup with a hashed `api_keys` lookup (SHA-256 of the token,
`last_used_at` bookkeeping, revoked/expired rejection) keeps every contract unchanged.

## Offline contract generation

`task gen` exports this schema without loading runtime configuration, connecting to
a database or starting a listener. Compile-only checks do not prove live HTTP, CORS or deployment.
