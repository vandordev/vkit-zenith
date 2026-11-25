# Architecture

`apps/web` owns the public foundation, local health and customer auth/dashboard.
Browser identity and dashboard calls go directly to independent `apps/web-api`;
public `apps/api` exposes organization API-key (Bearer) operations. `apps/bo` calls
independent `apps/bo-api` for staff/customer administration (ADR-0006). The Go APIs
use Echo/Huma, manual listener lifecycle, generated FX wiring and committed offline
schemas. Business operations use `/v1/...`.

Web API and BO explicitly own bounded pgx pools. Application usecases own mutation
authorization, ordered locks, transactions and audit; API-owned SQL provides scoped
reads. Runtime code never imports tools, migrations or containers. The single
21-table Goose baseline holds customer identity/admission, organizations,
memberships, independent staff, append-only audit, API keys and River. DBML documents the Up schema.

Membership role is owner/admin/member with a text CHECK; owner fences protect every
organization, including suspended ones. Permissions are a fixed Go map with
organization.read for all, members.manage and developers.manage for owner/admin and
organization.manage for owner (ADR-0004). No registered jobs, realtime runtime or
custom role system is installed.

Bun owns JS workspaces; go.work owns native Go modules. Task is the command entrypoint.
Offline task gen exports four schemas, the consumer SDKs and email templates;
route generation and pinned offline sqlc remain explicit. Email is a library,
never a separate runtime. Do not revive historical architecture as authority.

## Frontend-oriented backend contracts

All backends serving a frontend must adapt their data contracts to that frontend's
component and workflow needs, not force the UI to fit an existing DTO or database
row shape. `apps/bo-api` serves this Backend for Frontend (BFF) role for `apps/bo`;
`apps/web-api` serves customer identity and dashboard data for `apps/web`. BFF describes responsibility, not an extra
proxy/service layer. Other future public operations require their own approved contracts.

Use `packages/db/schema.dbml` to understand available data, relationships and
constraints before component-driven design (CDD). Goose migrations remain the
canonical schema and generation input. A documented table does not authorize a
new feature, database adoption or exposure of every column.

For reads, the frontend-serving backend prepares authorized, scoped, bounded data
ready for the intended presentation: joins, relationship assembly, aggregates,
business calculations and required display-ready values belong there. The frontend
renders and manages interaction; it must not reconstruct domain relationships,
derive business rules or combine raw API results just to obtain a component's
required data. Local interaction state is not a backend responsibility.

Browser-dependent presentation is also frontend-owned: format timestamp instants
in the user's browser timezone/locale using `@repo/ui-utils/date` (SSR-safe React
adapter: `@repo/ui-utils/react`, semantic component: `@repo/ui/date-time`). APIs
retain timezone-qualified ISO timestamps; do not preformat them in the server's
timezone or treat locale formatting as a business-data calculation.

Queries may read the database directly through API-owned persistence, without a
pass-through usecase. Mutations must call application-owned usecases for business
validation, authorization, transaction boundaries and audit; BFF handlers must not
write business data directly. Read paths must not hide business mutations.

Keep explicit typed HTTP/OpenAPI contracts and generated consumer clients, but do
not add redundant DTO layers or arbitrarily narrow responses to today's widgets.
Response shapes follow actual component needs, preserve meaningful nullability
and numeric precision, and never expose credentials, token hashes, internal
security state or data outside the caller's authority. This policy governs future
work; it does not claim existing contracts have already been refactored.

## Runtime configuration

One YAML document has isolated TypeScript/Zod v4 and Go/validator loaders. Builds
never resolve runtime secrets. `getPublicConfig` returns only `{ apiBaseUrl }`
with no-store; BO's isolated public function returns only `{ boApiBaseUrl }` and
initializes a credentialed memory-CSRF browser-local Hey API client. No tRPC, business
proxy, browser env or mutable SSR-global configured client. Presentation and shared
primitives remain independent of fetching/auth/navigation. See `config.md`,
`backend/README.md`, `frontend/README.md`, and `verification.md`.
