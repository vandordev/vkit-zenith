# Web API

Independent Go Echo/Huma service for `apps/web` (ADR-0006). It owns customer
authentication, session cookies, CSRF and dashboard `/v1/organizations/...`
operations. The public `apps/api` never shares cookie endpoints.

The typed `webApi` configuration uses `WEB_API_*` listener/docs variables (default
4101), existing `IDENTITY_*`, and an explicitly selected bounded database pool.
Public `api` uses `PUBLIC_API_*` (default 4102). Web's `API_BASE_URL` is the
browser-reachable Web API origin. Agents never read the operator's real `.env`.

Authentication and identity/organization reads are real: current full verified
session, fresh active user and membership, role permissions, organization profile
and active members. Endpoint-only SQL is beside its handler in
`internal/handlers/v1/organizations.sql`, generated to `internal/persistence` by
the `web-api` sqlc owner. Runtime shares one caller-owned pool and never migrates.

API keys and logo metadata are `[DUMMY]` (ADR-0001 phase 2): they return
deterministic fixtures or simulated responses; nothing is persisted, audited or
uploaded. Writes still enforce actual permission, Origin, session-bound CSRF and
suspension checks. Sensitive metadata is omitted; the only creation secret is the
unusable dummy constant `zk_DUMMYNOTAKEY`.

`/v1/` has exact-origin credentialed CORS and no-store responses. Dashboard bodies
have an inclusive 64 KiB limit; authentication preserves its 16 KiB policy. Body reads
have bounded operation deadlines, safe 408 timeout errors and safe parse/validation
locations that do not reflect credential values or dynamic keys. Authorization
distinguishes unauthenticated 401, scoped not-found 404, denied 403, suspended 409 and
unavailable 503 responses.

`task gen` generates FX handlers, offline `contracts/openapi/web-api.json` and the
Web-owned SDK; generated files are not hand-edited. The export graph constructs
contracts without runtime configuration, database access or listener startup.
`task gen:db` regenerates SQL bindings offline. The root `Dockerfile.web-api`
packages this service independently; native compilation does not verify an image.

Verification uses scoped compile-only/typechecks, fixture-only Storybook in the
Web owner, and counted `task check:web-api:security` plus `task check:auth:*` gates.
The organization gate requires the role matrix and 15 in-process HTTP cases against
owned isolated PostgreSQL, fake data, bounded supervision and confirmed cleanup;
absent/skipped cases fail. No backend listener, existing database, real mail,
or operational migration is used. Fixture/DTO reflection and
other unit tests outside counted gates remain written-only. These checks do not
prove deployed cookies/CORS/TLS, browser login, or live API routing.
