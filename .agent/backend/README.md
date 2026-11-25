# Backend

This directory groups backend guidance by responsibility, not by application
name. It applies across `apps/web-api`, `apps/api`, `apps/bo-api`, and future
backends. Runtime-specific ownership and
exceptions remain explicit below and in each application's README.

`apps/web-api`, `apps/api` and `apps/bo-api` own independent Echo/Huma runtimes and Go
modules. They never import each other's `internal` packages or proxy one another.
Independent `apps/bo-api` follows the same handler/module conventions and owns
staff/customer `/v1/...` operations. Web API and BO API opt into bounded pgx
pools/application usecases on runtime start, never on health/offline export. See
its README for independent identity/bootstrap/security and source packaging.
Manual `main.go` owns dispatch; `internal/server` owns config, logger, docs,
timeouts, middleware, bind-error handling and graceful FX lifecycle shutdown.
Generated root `app.go` owns `handlerOptions()`/`newApp()`; never edit manually.

One endpoint per handler file. Constructors use exact `//vkit:handler`, exported
package-level non-generic `func NewName(dependencies...) error`, register one
explicit stable unique Huma operation ID/path, and return an error before listener
startup on failure. Flat resource-first versioned files under `internal/handlers/v1`
are for future features; foundation health is unversioned. No `routes.go`.
Read `packages/tools/README.md` before changing generation conventions. Start new
endpoints with `task create:handler ... kind=query|mutation` and new writes with
`task create:usecase`; the scaffolds encode this layering.

Follow [frontend-oriented backend contracts](../architecture.md#frontend-oriented-backend-contracts)
for every frontend-serving backend. Shape typed responses around the consuming
components and workflows, not restrictive preselected DTOs or raw persistence
models. Prepare required joins, aggregates, business calculations and display-ready
values in the backend so the frontend does not reconstruct them. BFF is a role,
not a requirement to add a proxy or another service.

Timestamp instants stay timezone-qualified ISO values. Browser-local timezone and
locale formatting belong to the frontend's shared `packages/ui-utils`, not
server-preformatted display strings (see the architecture policy).

Queries may directly use API-owned persistence with authorization, scoping,
bounded pagination and consistent snapshots where needed. Business mutations must
go through application usecases, which own business validation, authorization,
transactions and audit; never bypass them with handler-owned writes. Keep query
paths free of hidden business mutations. Existing narrow infrastructure/security
bookkeeping is not permission to move business mutations into handlers.

Use explicit typed HTTP/OpenAPI contracts without redundant DTO-only layers;
never serialize arbitrary DB rows or secrets. Use direct detail objects and lists
`{ items, pagination }` with `[]` for empty items. Choose bounded pagination with the first collection,
not an unused generic framework now. Creation uses 201/Location where applicable,
bodyless success 204. Errors use Huma RFC 9457 Problem Details and proper statuses;
never expose raw errors/submitted secrets. Huma schema constraints own transport
validation/OpenAPI; Go validator complements them, not an automatic tag translator.

Serve Scalar `/docs` and `/openapi.json` on the public, Web and BO APIs; docs Basic Auth is configurable,
default off for Web API and public API, on for BO. Protect both routes when enabled; credentials
never enter HTML/schema/public config. Web auth uses exact configured Web-origin
credentialed CORS and memory-only CSRF; BO has separate staff identity.

`task gen` orders handlers, three offline schemas and the Web/BO SDKs. Export never
reads runtime config or opens a pool/listener. Web API identity and email usecases
remain installed; account/session tests use the counted Node wrappers and owned
isolated PostgreSQL as allowed by verification.md. Test capture/clock/fault tools
never enter runtime dependency closures. Other verification follows
`verification.md`: no service startup, probes or existing databases.
