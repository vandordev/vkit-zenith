# Frontend

This directory groups frontend guidance by responsibility, not by application
name. It applies across `apps/web`, `apps/bo`, and future frontend applications;
it is not specific to `apps/web`. Technology-specific guidance identifies its
own scope, and runtime-specific ownership remains in each application's README.

Before broad UI/UX changes, read [design direction](../../docs/frontend/DESIGN.md)
and [the UX contract](../../docs/frontend/FRONTEND_CONTRACT.md), including the
drawer/dialog/page decision rules. For new or substantially refactored surfaces,
use [the page brief](../../docs/frontend/PAGE_BRIEF.md) and
[review checklist](../../docs/frontend/FRONTEND_REVIEW.md). These are design norms,
not evidence that every existing screen already complies or new verification authority.

Follow [verification.md](../verification.md): UI uses Storybook plus typecheck;
server functions/adapters use TypeScript checks; Go uses compile-only checks. Do not start
the product stack, run system tests, or add lint/build/full-repository gates.
New or changed UI follows component-driven design: check persisted-data feasibility
and runtime support, develop typed presentation components with colocated stories,
verify them in Storybook, then integrate routes/generated clients. Presentation components do
not fetch data or import real API clients, route adapters, or server modules.
BO and Web dashboard presentation may use declarative TanStack Router `Link` directly; imperative
navigation stays in adapters and shared UI primitives remain router-agnostic.

Apply [frontend-oriented backend contracts](../architecture.md#frontend-oriented-backend-contracts)
to all frontend/backend pairs. Start CDD from available data and relationships in
`packages/db/schema.dbml` (Goose SQL remains canonical), then define the data each
component and workflow needs. Do not restrict a design to an existing DTO; adapt
the dedicated backend contract within approved feature scope. The backend prepares
joins, aggregates, business calculations and display-ready values; frontend
adapters must not compensate for missing contracts by reconstructing relationships
or business rules from raw responses. Presentation owns rendering and interaction,
including local draft/selection state, not domain data assembly. Backend queries
may read DB directly with authorization/scoping; mutations always use application
usecases, never direct business-data writes from BFF handlers. Typed contracts,
generated SDKs, data privacy and precision remain required.

Browser-local date/locale formatting is presentation, not domain data assembly.
Shared helpers belong in `packages/ui-utils`; use `@repo/ui-utils/date` for pure
formatting, `@repo/ui-utils/react` for hydration-safe formatting, or the semantic
`DateTime` primitive from `@repo/ui/date-time`. Do not format using the server's
timezone during SSR and assume it represents the user.

Read [the Storybook guidance](../../apps/storybook/README.md) for discovery,
fixtures, styling, and the ownership-safe server lifecycle on port 6006; starting
or restarting this repository's instance needs no repeated permission.

`apps/web` owns the public foundation, local health, customer auth pages and
/dashboard, its committed route tree and Web API-owned SDK consumer. Auth pages
use Indonesian UI, memory-only CSRF and identity-scoped private queries. The public
foundation stays visible if configuration fails. See the Web README.

Independent `apps/bo` owns staff/customer routes, its own SDK and exact
{ boApiBaseUrl } config. Both frontends call their Go APIs directly; no business
proxy, SSR auth fetch or shared private query cache. All primitive/Storybook rules apply.

Use TanStack Start with Tailwind CSS and shared UI primitives as the UI
baseline. All application controls, including existing UI and stories, must import
`@repo/ui/<name>` directly, not local `@/components/ui` copies or Base UI/Radix.
Do not render native controls or bypass the policy with intrinsic aliases, motion
wrappers, or JSX factories. Implement/install missing primitives in the shared
package first; native implementations belong only there. Structural HTML, forms,
headings, lists, links, and SVG remain allowed. Feature compositions stay local.
Run `task check:ui-primitives`; root and web typechecks run it too.
Read [component rules](../components.md) and
[package guidance](../../packages/ui/README.md) for actual installed APIs.
Keep accessible labels, focus states, responsive layouts, and typed
generated SDK/query calls with explicit runtime client instances. Do not
reintroduce Next.js or Mantine as defaults.

See [routing.md](routing.md) for the complete directory-first route
convention, English URL vocabulary, feature-local ownership, adapter isolation rules
and native typecheck command. These apply to **all TanStack Start apps**.
Read [stories.md](stories.md) for per-component CSF metadata, sidebar hierarchy,
fixture ownership and state coverage before creating or moving stories.
See [client guidance](../../apps/web/README.md) for runtime setup,
offline generation and Zod request-body reuse in future forms.
For TanStack Start SEO, use the `tanstack-seo` skill and [SEO rules](seo.md).
Adapt its examples to `src/app` and `appConfig`; do not invent a public domain or
index private content. SEO server/head changes follow system typecheck-only verification.

The generated tree is owned by the TanStack plugin. Keep route-local helpers
under `-`-prefixed files or directories; `_` is reserved for pathless layouts.

Brand name, logo, color tokens, and shared radius scale are owned by `packages/brand`; see its
[README](../../packages/brand/README.md). Import `@repo/brand` and its asset/CSS
entrypoints instead of duplicating values. `apps/web/src/lib/config.ts` is the web
adapter: `appName` and `defaultTitle` derive from brand, while `defaultDescription`,
`favicon`, and `repositoryUrl` remain app configuration. Metadata and public UI
may override route-specific copy, but must not repeat brand or application defaults.
Secrets and runtime environment values remain in the single YAML configuration
and lazy server-only typed getters; `appConfig` never reads credentials or
`process.env`. Browser config is exactly `{ apiBaseUrl }` from no-store
`getPublicConfig` at runtime. Pending/error/ready initialization never gates the
static page or invents a fallback URL. No deployment env or build injection.
Public schema/types use `@repo/config/public` only.
