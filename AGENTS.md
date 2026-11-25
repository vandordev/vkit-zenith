# AGENTS.md

## Start here

Read `README.md`, the relevant `.agent/**/*.md` files for the runtime being
changed, the plan/spec, and run
`git status --short` before changing code. Preserve unrelated changes.
For any product-scope work, first read the foundation documents below.

Dispatched workers (an orchestrator gave you a task brief with excerpts of the
plan, spec and ADRs) skip the full reading above: the brief is your plan/spec.
Read the `.agent/**/*.md` files for your runtime and the files the brief names;
open `README.md`, the foundation documents or the full plan/spec only for a
question the brief doesn't answer, and then only the section you need. Run
`git status --short` in your own checkout only.

## Foundation documents (product direction)

These are the active authority for what the product is and how it is built.
Read them at the start of every session that changes product scope:

1. `PRODUCT.md`: vision, users and glossary of the product built on Zenith.
2. `docs/adr/README.md`: index of Architecture Decision Records. Accepted ADRs are
   binding; Proposed ADRs need user approval before implementation.
3. The current spec/plan under `docs/superpowers/{specs,plans}`.

Rules:

- Do not silently change agreed scope. Anything outside the current spec is
  discussed with the user first.
- A decision that contradicts an Accepted ADR needs a new ADR that supersedes it.
  New durable decisions get an ADR.
- Delivery follows ADR-0001: schema first, then UI with dummy-data endpoints for
  every screen, then real functionality. Do not wire real persistence in a dummy phase.

## Repository shape

- `apps/web`: TanStack Start, shared UI, runtime public config and consumer-owned
  Hey API client calling Web API directly; customer auth and the organization dashboard.
- `apps/landing`: Next.js App Router static Indonesian marketing site (`output: "export"`);
  no API, auth, runtime config or env reads. See ADR-0008 and its README.
- `apps/bo`, `apps/bo-api`: independent Indonesian staff/customer UI and Go API.
  Browser calls BO API directly with credentials and memory-only CSRF. Independent
  staff identity, no customer SSO or business proxy. See their READMEs.
- `apps/storybook`: repository-wide presentation previews; stories stay colocated
  with their owning components. Development tooling only, not a deployed runtime.
- `packages/ui`: shared UI source, imported through explicit
  `@repo/ui/<name>` entrypoints. All new small UI primitives belong here,
  never in application-local copies. This includes existing application controls
  and Storybook fixtures; structural/feature presentation stays application-owned.
- `packages/brand`: source of truth for brand name, logo, and color tokens.
  Consumers use `@repo/brand` entrypoints; do not duplicate identity or palette.
- `apps/web-api`, `apps/api`, `apps/bo-api`: independent Go Echo/Huma modules/listeners, manual
  server lifecycle/entrypoints, generated FX `app.go`, service-local handlers.
- `packages/config/{js,go}`: isolated typed loaders for one YAML document;
  nested JS workspace retains `@repo/config` identity.
- `packages/db`: opt-in Go pools/transactions, explicit Goose helpers, SQL source
  and native pgx transactions. Never automigrate; Web API and BO API opt in explicitly.
  Goose migrations remain the schema source of truth; `packages/db/schema.dbml`
  is its committed visual documentation only, never input to migrations or sqlc.
  Every schema-affecting migration must update DBML in the same change: tables,
  columns/types/nullability/defaults, keys, indexes and composite foreign keys.
  Document PostgreSQL-only constraints/triggers/partial indexes in DBML notes
  rather than inventing equivalent DBML enforcement. Review the final Up schema,
  not Down statements; trigger-only changes must align relevant notes as well.
  Run `task check:db-schema` for offline syntax/reference validation; this does
  not establish migration parity. `task dev:db-schema` is a local documentation
  viewer only (loopback port 6007), never a database connection or replacement
  for Goose/sqlc. Do not launch it as a backend verification workaround.
- `packages/application`: native staff/security, customer and identity usecases;
  owns authorization, transaction boundaries and audit, not HTTP transports.
- `packages/tools`: AST composition and pinned offline sqlc generation plus
  explicit operational migration CLI; not a runtime app.
- `contracts/openapi`: three committed offline schemas; Web and BO own their SDKs.
- `packages/fixtures` owns deterministic dummy data for phase-2 endpoints (ADR-0001);
  the public API's Bearer keys and the dashboard's API keys are fake fixtures until a
  product persists API keys. `apps/worker` fails fast until a product registers its
  first job (ADR-0003).

## Architecture rules

### React UI primitives

Application controls must import shared React primitives directly from
`@repo/ui/<component>`, including existing UI and Storybook fixtures.
Do not render native `button`, `input`, `select`/`option`/`optgroup`, `textarea`,
`label`, `dialog`, `details`/`summary`, table elements (including `caption`),
or `progress`/`meter` in applications. Do not bypass this with intrinsic aliases,
motion wrappers, `createElement`/JSX factories, application-local primitive copies,
or direct Base UI/Radix imports. Native implementations belong only in
`packages/ui`. If a primitive is missing, implement or install it there
and expose its public entrypoint before consuming it. The full Vandor catalog is
installed by explicit user request; this is not a reason to bulk-install other catalogs.

Structural/semantic HTML (containers, headings, forms, lists, navigation links)
and SVG remain allowed. Router links are not primitive violations.
Run `task check:ui-primitives` (or `bun run check:ui-primitives`); the static checker
also runs with root and web typechecks. It scans `apps`, including stories and
legacy UI, excluding test/spec files and generated/dependency outputs. It detects
common bypasses, not every possible dynamic expression; the policy still applies.
See `.agent/components.md` for ownership and `.agent/verification.md` for gates.

Web hosts the foundation page, local health and runtime public config. Browser
identity calls target Web API directly; no tRPC/business proxy. Go APIs expose
`/health`, `/docs`, `/openapi.json`; future business operations use `/v1/...`.
BO API serves approved staff/customer `/v1/...` operations. Each service owns its
module/listener and root Dockerfile. Web API adopts a bounded DB pool and usecases for customer identity.

Runtime configuration uses only `config/config.yaml` and lazy typed
`@repo/config/server` getters with explicit field isolation and Zod 4. Browser
imports are limited to `@repo/config/public`; deployment values arrive through
the no-store `getPublicConfig` server function, exactly `{ apiBaseUrl }`, never
Vite/public env or build injection. BO's isolated function returns exactly
`{ boApiBaseUrl }`. Go uses isolated typed loaders and validator;
Huma retains native transport validation. No feature-level env reads. Only narrow
bootstrap/tool adapters may emit env settings; see `.agent/config.md`.

Use exact `//vkit:handler` constructors, one operation per file, no `routes.go`.
Regenerate FX wiring, offline schemas and consumer SDK via `task gen`; never
hand-edit generated files. Go discovery is native experimental Turbo support,
not package.json shims. New business scope still requires an approved real feature.

## Workflow

Follow `.agent/verification.md` for the two verification tracks: UI uses Storybook
plus typecheck; system (all backend/service work, including web server functions)
uses TypeScript typecheck or Go compile-only verification. Go compilation is
authorized without repeated permission; do not execute the resulting binaries.
Never start Compose, containers, databases, backend services, or execute system
tests, migrations, general builds, or full quality gates for verification.
The approved database-foundation exception in `.agent/verification.md` overrides
that default only for focused DB/config-isolation, pool/transaction/migration/
codegen tests and owned isolated Testcontainers PostgreSQL. The approved BO
exception additionally permits focused crypto/governance and in-process HTTP→PG
security acceptance via `task check:bo:security`, never backend listeners. Fake credentials,
bounded timeouts and cleanup are required. No existing user database, operational
migration target, SSO checks, unrelated suite or real `.env` access in tests.
The customer auth/email gates additionally permit focused fake-data
account/session/config/email tests and owned isolated PostgreSQL with allowlisted
child environments, bounded deadlines and confirmed cleanup. It does not authorize
backend listeners, real mail, existing databases, deployment or general builds.
Auth verification uses the Node-native counted `task check:auth:*` wrappers;
absent or skipped required cases fail. Do not invoke root Bun hooks/audits for
these gates; Bun tests run from an owned empty directory with an allowlist.
Narrow TanStack route generation (including dev/build only when necessary for
codegen) and safe repository-owned Storybook start/restart on port 6006 are also
authorized without repeated permission; follow `.agent/verification.md` safeguards.
Risk does not authorize extra runtime checks. Report runtime integrations as
unverified, not proven by typecheck. Exceptions require new explicit user instruction.

UI follows component-driven design: check schema and runtime feasibility, build
typed presentation components with colocated stories, verify them in Storybook,
then integrate routes and generated clients. Keep fetching, authorization, and imperative
navigation out of presentation components. BO and Web dashboard presentation may use declarative
TanStack Router `Link` directly; shared `packages/ui` primitives remain router-agnostic.
Read `.agent/frontend/README.md` and the Storybook README.
For shared UI, read `.agent/components.md` and `packages/ui/README.md`.
Every TanStack Start app follows `.agent/frontend/routing.md` and `.agent/frontend/stories.md`:
English URL segments, directory-first routes, feature-local presentation/stories,
and component-accurate CSF groups. UI language does not determine URL language.
For brand changes, read `packages/brand/README.md`; web config derives brand identity.
For TanStack Start SEO, use the `tanstack-seo` skill and `.agent/frontend/seo.md`;
adapt examples to `src/app`, existing brand/config ownership, and typecheck-only
system verification. Confirm public origin/indexing scope before implementation.
Prefer `task` commands. Preserve unrelated changes and report exactly what was
verified and what remains unresolved. Commit completed tasks with an appropriate
conventional message; no automatic merge or push.
