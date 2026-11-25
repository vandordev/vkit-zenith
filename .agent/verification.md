# Verification and component-driven design

This is the active verification policy, linked from `AGENTS.md`. Verification
has exactly two tracks: UI uses Storybook plus typecheck; system uses TypeScript
typecheck or Go compile-only checks. Historical plans, risk-based testing guidance,
and command examples do not override these gates. Documentation-only changes use
document/formatting review.

## System changes

### Approved Core Auth/email exception

Focused fake-data configuration, account/session, crypto, renderer, transport,
parser and in-process API security tests remain authorized. Owned isolated
Testcontainers PostgreSQL initializes only canonical migrations, with bounded
contexts, enabled Ryuk and verified owned cleanup. No existing/user database,
real `.env`, credentials, mail, deployment, Compose, backend listeners or general
builds. Tests use fake keys, clocks, capture and explicit configuration.

Retained counted commands are `task check:auth:contracts`, `config`, `unit`,
`schema`, `accounts`, `sessions` and `adapters`. Every selected required case must
exist and pass; zero/skip/failure is a failed gate. Run the Node entrypoint or
Task, never root Bun hooks/audits. Bun runs from an owned empty directory with
an allowlisted environment; Go uses the explicit workspace and GOENV=off.

Every counted Go run uses the non-browser Python/pidfd subreaper supervisor at
`scripts/auth-e2e/supervise.py`. A private fsynced acquisition journal lives outside
disposable HOME. Child nonzero/timeout stays failed even if cleanup succeeds;
cleanup-unknown retains evidence and stops execution. Never substitute an outer
unsupervised timeout for resource-acquiring counted checks. See
`packages/tools/authtest/README.md`. Source/in-process evidence does not establish
live browser login, deployed cookies/CORS/TLS/proxies or real mail.

### Approved BO customer operations exception (2026-10-26)

The approved customer operations spec authorizes focused application and in-process
Echo/Huma recorder → owned isolated Testcontainers PostgreSQL acceptance for customer
authority/confirmation, transactions/audit, revisions, last-owner concurrency and
persisted authorization revocation. Run `task check:bo:customers` (or its precise
task prefix during red/green), with fake credentials/data, bounded contexts,
enabled Ryuk and confirmed cleanup. Explicit integration fails rather than skips
without Docker. Existing `task check:bo:security` remains unchanged in scope/selection.
No listeners, customer login/SSO/issuer, Compose, existing/user DB, real `.env` or
secrets, operational migration target, unrelated suites/builds or deployment.
Persisted fixtures do not prove absent issuer/redemption or remote product logout.

### Approved BO security exception (2026-10-25)

The approved BO runtime spec/plan additionally authorize focused actual password/
token primitive, bootstrap/governance, login, cookie/CSRF, expiry/revocation,
limiter, authorization and last-Superadmin tests. Use in-process application or
Echo/Huma requests (Recorder, never httptest.NewServer), owned isolated
Testcontainers PostgreSQL where required, fake credentials, injected clocks,
bounded timeouts, enabled Ryuk and confirmed cleanup. Explicit integration fails,
never skips, without Docker. This includes real transaction/concurrency acceptance,
not merely mocked authorization. Preserve the persistence exception below.
No live product/backend listeners, compiled service execution, app stack, Compose,
Redis, real .env, existing/user DB, operational migration target, customer SSO,
unrelated suites, images/deployments or broad build/lint/
quality gates are authorized. Offline generators and affected Go compile/TS
typechecks remain authorized. Fixture proof is not deployed runtime proof.

Commands: `task check:bo:security` selects bounded crypto/terminal/security units
and actual in-process HTTP→PG acceptance; `task check:db:integration` retains
schema persistence and deterministic offline sqlc fixtures. Focused BO
browser transport/redirect/form-value adapter tests may run with injected fetch
and local data only, no listening server or real credentials. Frontend appearance
still requires actual Storybook review; adapter tests do not prove deployed cookies.

### Approved organization dashboard security exception

`task check:web-api:security` counts the organization role matrix and in-process HTTP
recorders against an owned isolated Testcontainers PostgreSQL database with fake
data. Its supervisor uses bounded deadlines, Ryuk, an allowlisted child environment
and confirmed cleanup receipts; missing or skipped cases, unavailable Docker,
timeouts, failures or uncertain cleanup fail the gate. It starts no backend listener
and uses no existing database, real credentials, mail, Compose or operational
migrations.

### Approved database foundation and platform-schema exception

The user-approved database foundation spec and platform decision log section 10
authorize focused settings/config-isolation, connection, transaction, Goose
migration and offline sqlc codegen tests, plus approved platform-schema constraint,
scope, audit, lifecycle, concurrency and generated-query round-trip tests. These
include owned isolated Testcontainers PostgreSQL, not application services.
Use fake credentials, bounded timeouts, temporary
directories under `/tmp/opencode`, and owned-resource cleanup. Explicit integration
selection must fail rather than skip when Docker is unavailable. Record
PASS/FAIL/BLOCKED/SKIPPED honestly. This exception overrides the static-only rules
below only for this persistence scope; it does not authorize unrelated tests, SSO checks,
general Compose, existing user databases, remote/production migrations, deployment,
service startup, or broad quality/build gates. Never execute operational migrations
against the configured user target as verification or read real `.env` in tests.

- System includes backend integration and all service-related work: Go APIs,
  config/tooling, web server functions, and external service adapters. Deferred
  future persistence/jobs/auth changes remain system work if approved.
  Server-side code in `apps/web` belongs to this track, not the UI track.
- Run fresh repository-native TypeScript typechecks or Go compile-only checks for
  affected owners and consumers. Stop after successful static checks. Inspect the
  command first: it must not start
  services, containers, or migrations through prerequisites or lifecycle hooks.
- Go compile-only checks such as `go build ./...` are authorized without repeated
  permission. Inspect commands first, scope them to affected modules, and direct
  generated executables outside source trees where practical. Never execute the
  resulting binaries as verification. Do not substitute `go test` or `go vet`,
  invoke runtime lifecycle hooks, or turn this into an image/deployment build gate.
- Trust typed contracts, inferred types, and explicit domain types. Do not weaken
  types with `any`, unsafe casts, or non-null assertions to make checks pass.
  Use `unknown` with validation at untyped boundaries.
- Do not run system tests (including unit, integration, smoke, or end-to-end),
  lint/general build gates (except Go compile-only checks above), `task quality`,
  `task build`, Compose, Docker containers,
  databases, Redis, backend dev servers, workers, schedulers, realtime servers,
  migration commands, or live service probes for verification. Do not substitute
  another launcher or tool to bypass this resource limit. A perceived risk or
  integration need is not permission to run them.
- Preserve existing tests, but do not execute them as part of this track.
  These commands may remain in the repository for other purposes; their presence
  does not authorize execution. Any exception needs a new explicit user instruction.
- Typecheck cannot establish authorization, validation behavior, SQL constraints,
  migrations, transactions, concurrency, or live integrations. Report these as
  runtime-unverified where relevant, never as proven by typecheck. If there is no
  safe typecheck for a changed artifact, report the gap rather than substituting
  a runtime check. Never apply remote migrations without deployment authorization.

### Approved M2 infrastructure exception (2026-11-13)

`task check:infra` counts only the focused `packages/secrets`, `packages/storage`,
and `packages/config/go` cases named in `scripts/check-infra.mjs`, plus both JS
document tests. Missing, zero, skipped, failed, timed-out, or nonzero runs fail.
Go uses the explicit workspace and `GOENV=off`; Bun runs from an owned empty
directory with `--no-env-file`. The gate uses no real credentials and accesses
neither `.env`, R2, databases, service listeners nor the worker binary.

The approved `task check:db:integration` scope includes `TestSchemaRiver` on its
owned Testcontainers PostgreSQL database: a test-local job is inserted with
`InsertTx` and processed by an in-process River client within a bounded context.
The client stops before container cleanup. This does not authorize a worker
binary or listener; cleanup uncertainty remains unresolved.

## UI changes

UI includes presentation primitives, layouts, pages, and local interactions.
The gate is fresh typecheck of affected owners/consumers plus browser inspection
of affected Storybook stories. Stories use local fixtures/providers only; do not
start or call a backend/database/service to verify UI. Do not add automatic lint,
build, full-repository, or system test gates. Standalone Storybook is the only
preview runtime allowed by this track, subject to the lifecycle safeguards below.

1. Before designing data-bearing UI or fixtures, inspect the relevant OpenAPI
   artifact, explicit handler DTOs and runtime implementation. Canonical platform
   schema source exists but has no API consumers. Map filters, sorting and aggregates
   to real supported fields/derivations. Respect nullability and any approved
   authorization boundaries. Schema support does not prove runtime availability.
   Report unsupported data before designing dependent UI. Do not invent fixture
   fields, assume future APIs, or change the schema without approval. Static copy
   and local UI state are not persisted product data.
2. Develop typed presentation components and colocated `*.stories.tsx` first.
   Stories render the actual components used by the app, not a duplicate demo.
   Separate fetching, auth, imperative navigation, and server dependencies; accept typed props
   and callbacks. Use deterministic fixtures and local providers, never production
   credentials or authenticated APIs.
   BO and Web dashboard presentation may use declarative TanStack Router `Link` with the shared
   Storybook memory router; never import production routes/providers into stories.
3. Add/update stories for changed visible UI and interactions, including pages and
   layouts, not just primitives. Pure infrastructure needs no artificial stories.
   Cover relevant populated, pending, error, empty, disabled, and long-content
   states; responsive layouts and themes where supported. Do not invent themes
   or variants the application does not support.
4. Verify affected stories in the browser: appearance, realistic copy, interactions,
   keyboard/focus, responsive overflow, and accessible feedback. Typecheck alone
   cannot establish visual or interaction quality. A static Storybook build is
   not required and does not replace browser inspection.
5. Check the existing Storybook at `http://localhost:6006` first. If absent or stale,
   starting/restarting this repository's instance is authorized without repeated
   permission. Inspect processes/listeners, identify the old project instance,
   stop it and confirm exit before starting its replacement
   with `task dev:storybook`. Verify exactly one server, including duplicates on
   other ports. Never fall back to another port or an alternate preview server.
   Ask if ownership is unclear; do not broadly kill unrelated processes.
6. Integrate verified components into routes afterward. Derive operation inputs
   and response shapes from verified UI needs, regenerating consumer-owned clients.
   Validation, authorization, business rules, and secret filtering remain server-owned.
   Real auth, persisted data, and backend integration belong to the system track:
   TypeScript typecheck or Go compile-only checks, with runtime coverage explicitly
   reported as unverified.

## Mixed changes

### TanStack route code generation

Route-tree generation is authorized without repeated permission. Prefer a
supported standalone generator. A narrowly scoped web dev/build command is allowed
only when necessary to regenerate routes, not as a general verification gate.
Inspect prerequisites first; do not start backend services, containers, databases,
or migrations. Stop agent-started dev processes after generation, never hand-edit
the generated tree, and inspect the resulting imports. This exception does not
authorize product browsing, live API probes, or unrelated build/test work.

Apply each gate to its own boundary: presentation uses Storybook plus typecheck;
system wiring uses TypeScript typecheck or Go compile-only checks. A UI change does
not authorize starting the product stack. Schema/runtime feasibility inspection is read-only source review,
not permission to connect to a database or run a service.

## Reporting

Preserve unrelated changes; do not reformat unrelated files for global checks.
Report exact commands/evidence and unresolved boundaries. Blocked/skipped checks
are not passes. Distinguish UI browser evidence from system type safety and list
runtime-unverified boundaries without claiming integration success. Stop after
the applicable gate; risk does not expand execution authority. Do not treat
historical verification counts as current results.
