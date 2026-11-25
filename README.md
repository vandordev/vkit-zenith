# vkit-zenith

Vandor's foundation for multi-tenant SaaS products. It ships customer identity,
organizations with memberships, an independent staff Backoffice with audit, a public
API keyed by organization API keys, and the infrastructure foundations (River worker,
private R2 storage, envelope encryption, typed config). A product starts from
Zenith and adds its own domain on top.

Reference implementation: Valcentra (a payment platform built on the same patterns).

## Ownership

- `apps/web`: TanStack Start customer app: `/` foundation page, local `/health`,
  login/register/email/password pages and the signed-in organization dashboard `/w/{slug}`.
- `apps/web-api`: Go Echo/Huma API for `apps/web`: customer auth, session cookie, CSRF
  and dashboard operations.
- `apps/bo`, `apps/bo-api`: independent Indonesian staff Backoffice UI and API.
- `apps/api`: Go Echo/Huma public API for customers' servers (Bearer API keys).
- `apps/worker`: River worker skeleton; fails fast until a product registers a job.
- `apps/landing`: static Next.js marketing page (placeholder copy).
- `apps/storybook`: repository-wide presentation previews on port 6006.
- `packages/application`: staff, customer-administration and identity usecases:
  authorization, transactions and audit.
- `packages/db`: explicit pgx pools/transactions and one 21-table Goose baseline
  (identity, organizations, staff, audit, API keys, River).
- `packages/config/{js,go}`: isolated typed loaders for `config/config.yaml`.
- `packages/fixtures`: deterministic dummy data for phase-2 endpoints (ADR-0001);
  its API-key values are fake fixtures, never credentials.
- `packages/ui`, `packages/ui-utils`, `packages/brand`: shared controls, display helpers and identity.
- `packages/ui-email`, `packages/email`: offline email templates and Go delivery library.
- `packages/storage`, `packages/secrets`: R2 client and envelope encryption.
- `packages/tools`: offline generators, explicit migration tools and isolated test fixtures.

Browser calls target each app's own Go API directly; no business proxy and no SSO
between staff and customers. Go APIs expose `/health`, `/docs` and `/openapi.json`.
See [architecture](.agent/architecture.md), [configuration](packages/config/README.md),
[verification](.agent/verification.md) and the [ADRs](docs/adr/README.md).

Membership roles are `owner`, `admin`, `member`, stored as text with a CHECK. Every
organization, including suspended ones, retains an active owner. Permissions are a
fixed Go map (ADR-0004): `organization.read` for all, `members.manage` and
`developers.manage` for owner/admin, `organization.manage` for owner.

## Development and configuration

Use pinned toolchains and Task commands: `task install`, `task dev:web`, `task dev:bo`,
`task dev:landing`, `task dev:storybook`. Backend launchers are operator commands,
not agent verification. Dockerfiles and Compose are source references only. Never
automatically read or modify an existing `.env`.

YAML sections are `db`, `worker`, `storage`, `encryption`, `app`, `web`, `bo`, `boApi`,
`api`, `webApi`. Selected fields resolve once after strict document validation; no
feature-level env reads. Web's no-store `getPublicConfig` returns exactly
`{ apiBaseUrl }`; BO returns exactly `{ boApiBaseUrl }`. Browser code imports only
`@repo/config/public`.

## Offline generation

```sh
task gen:routes       # committed Web TanStack route tree
task gen:bo-routes    # committed BO route tree
task gen:handlers     # exact //vkit:handler constructors to FX app.go
task gen:openapi      # three committed offline schemas, no listener/config/DB
task gen:web-client   # Web API schema to Web-owned SDK
task gen:bo-client    # BO schema to BO-owned SDK
task gen:email        # deterministic local templates
task gen              # ordered API/SDK/email generation
task gen:db           # pinned offline sqlc, no DB or Docker
```

Commit generated wiring, schemas, route trees and SDKs; never hand-edit them.

## Scaffolding

```sh
task create:handler app=web-api name=projects_list kind=query method=GET path=/v1/organizations/{organizationId}/projects
task create:handler app=bo-api name=projects_archive kind=mutation method=POST path=/v1/projects/{projectId}/archive
task create:usecase domain=projects name=Create
task create:migration name=add_projects
task sync:web   # or sync:bo / sync:api: regenerate wiring, schemas, SDK and routes for one pair
```

Generators only create files and refuse to overwrite. A handler gets typed
input/output, the app's authorization check and a 501 until you connect it: queries
read through the API's sqlc persistence, mutations call a usecase (ADR-0004). `kind`
is required; queries use GET, mutations never do; web-api paths must carry
`{organizationId}`. Operation IDs derive from the name (`webV1ListProjects`,
`boV1ArchiveProjects`, `listProjects` on the public API); rename before the first
commit if the resource should read singular. `task check:scaffold` runs the golden,
validation and real compile tests.

## Presentation and verification

Application controls and stories import shared `@repo/ui/<name>` primitives directly.
Brand name/logo/color tokens belong to `@repo/brand`. Presentation is prop-driven;
route adapters own fetching, authorization and imperative navigation. Follow the
[component policy](.agent/components.md) and [verification policy](.agent/verification.md).

System gates: `task check-types`, `task check:go`, `task check:ui-primitives`,
`task check:db-schema`. UI changes also need Storybook inspection at desktop and
mobile sizes. Focused persistence/BO/auth gates (`task check:db:integration`,
`check:bo:customers`, `check:bo:security`, `check:web-api:security`,
`check:auth:*`) use owned isolated PostgreSQL with fake data; no skipped required cases.

Goose migrations are canonical; `packages/db/schema.dbml` is visual documentation
and is updated in the same change. `task db:status|db:version|db:up|db:down-to` are
operator commands, never authorization to touch an existing database.

## Starting a product from Zenith

1. Copy the tree without history: `git archive --remote` or
   `git -C vkit-zenith archive HEAD | tar -x -C my-product`, then `git init`.
2. Rename the Go module root `github.com/vandordev/vkit-zenith` in every `go.mod`,
   import and `go.work` (one search-and-replace), and the TS package names if needed.
3. Set the brand in `packages/brand` (name, logo, tokens) and regenerate icons
   as described in its README.
4. Pick your API key prefix: change `zk_` in the `api_keys_prefix_check` constraint,
   `schema.dbml` and the fixtures.
5. Fill `config/config.yaml` defaults and `.env.example` for your environments.
6. `task install && task gen && task check-types && task check:go`.
7. Write your first spec: schema first, then dummy UI/API, then functionality (ADR-0001).
