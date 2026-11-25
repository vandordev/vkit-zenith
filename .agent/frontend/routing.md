# TanStack app routing

## Scope and URL vocabulary

These conventions apply to **every TanStack Start application** in this repository,
including `apps/web`, `apps/bo`, and future consumers. Paths below are relative to
the owning application, not a shared cross-application route tree.

- Use English for URL segments, route directories, dynamic parameter names and
  code identifiers: `products`, `plans`, `offers`, `staff`, `account`, `roles`,
  `versions`, `$productId`. UI language is separate: `/products` can display
  **Produk**. Neither Indonesian UI nor chat language implies translated URLs.
- Use directory-first routes: `resource/index.tsx`, nested resources under their
  owner, and `route.tsx` only for an actual shared layout/policy. `_` is pathless;
  `-` is ignored code, including fixtures and helpers outside `-components`.
- A route rename updates typed links/navigation, literal hrefs, return/fallback
  destinations, fixtures, relevant tests and current docs together. Regenerate
  the owning app's tree; never edit generated output. Preserve query/selection
  context and safe-return validation. Compatibility comes from approved requirements,
  not invented redirects or parallel route implementations.

## Route tree

Each app's TanStack Start Vite plugin reads its own `src/app`, with the generated
tree in its own `src/routeTree.gen.ts`. The web foundation starts with:

```text
src/app/
├── __root.tsx
├── index.tsx
├── -components/                 # global pending/error/not-found UI
└── health/index.ts
```

`__root.tsx` owns the document shell, providers, head, scripts, and global
route boundaries. The index renders the prop-driven Zenith foundation page at `/`.

## Native naming tokens

| Token         | Meaning                                                |
| ------------- | ------------------------------------------------------ |
| `__root.tsx`  | Required root route                                    |
| `index.tsx`   | Directory index route                                  |
| `route.tsx`   | Directory layout route, usually rendering `<Outlet />` |
| `$userId`     | Dynamic URL segment                                    |
| `$`           | Final splat segment                                    |
| `_public`     | Pathless layout; the name is absent from the URL       |
| `(admin)`     | Organization-only route group                          |
| `-components` | Ignored colocated code; never a route                  |

Use `route.tsx` only when a directory needs shared UI, a loader, a guard, or a
route policy. Do not invent `layout.tsx`, `page.tsx`, `loading.tsx`, or
`error.tsx` conventions. `_` is reserved for pathless layouts; use `-` for
ignored colocation.

Examples:

```text
app/_public/index.tsx                 # /
app/users/$userId/index.tsx           # /users/:userId
app/dashboard/route.tsx               # nested dashboard layout
app/api/$.ts                           # server adapter for /api/*
app/health/index.ts                    # server adapter for /health
app/_authenticated/route.tsx           # future beforeLoad auth guard
```

## Server adapter isolation

`health/index.ts` provides local web process health without a backend connection.
It has no UI component and does not proxy another service. No `/api` or `/trpc`
business adapter is installed. `getPublicConfig` is the only runtime-config
server function; it is not a business proxy.

## UI data and navigation

UI uses consumer-owned Hey API SDK/Query output and direct Web API access.
Route adapters own data, auth, and imperative navigation; presentation components accept
typed props and callbacks without importing real clients. UI search parameters must use
`validateSearch` with Zod. Internal navigation uses typed `<Link>` or
`navigate`, including the future `_authenticated` `beforeLoad` guard.
BO and Web dashboard presentation (`dashboard/` routes) may import declarative `Link` directly from `@tanstack/react-router`
for internal destinations, without adding props solely to substitute links in Storybook.
Shared `packages/ui` primitives remain router-agnostic. BO and Web dashboard stories receive the shared
in-memory preview router, never the production route tree or authenticated providers.

## Component-driven pages

New or changed UI pages/layouts own typed presentation components and colocated
stories under `-components`. Develop and verify these in Storybook before route
integration. Place them at the **nearest route or layout owning the feature**,
not an application-root catalog of unrelated pages:

```text
src/app/
├── -components/                   # genuinely app-wide compositions only
├── login/
│   ├── index.tsx                  # data/auth/navigation adapter
│   └── -components/
│       ├── login-form.tsx
│       └── login-form.stories.tsx
└── _authenticated/
    ├── route.tsx                  # session/layout adapter
    ├── -components/
    │   ├── bo-shell.tsx
    │   └── bo-shell.stories.tsx
    └── plans/$planId/versions/$versionId/
        ├── index.tsx
        └── -components/
            ├── version-editor.tsx
            └── version-editor.stories.tsx
```

Adapters own fetching, auth integration, URL state and imperative navigation; presentation
owns typed props/callbacks and local interaction state. Promote a composition to
the nearest common owner only for real reuse. Do not combine unrelated pages in
app-root `catalog.tsx` or `access.tsx`; small primitives still belong in `packages/ui`.
Follow [story organization](stories.md) for component-accurate CSF metadata,
sidebar hierarchy and fixture/state coverage. Installing Storybook alone does
not automatically migrate existing pages.

## Metadata and route boundaries

Brand identity comes from `packages/brand`. Use `appConfig` from `src/lib/config.ts`
for web defaults derived from that identity. Its
`appName`, `defaultTitle`, `defaultDescription`, `favicon`, and `repositoryUrl`
fields are the web configuration adapter, not an independent brand source. Use `createMetadata` from
`src/lib/metadata.ts` in `head: () => ...`; nested TanStack head values merge
with the most-specific title/meta winning. Route metadata may override title or
description for a specific page. The root defaults come from `appConfig`. The
`tanstack-seo` skill and [SEO rules](seo.md) govern future SEO extensions, including
absolute canonical/social URLs, structured data, and public discovery endpoints.
The existing helper is a partial foundation, not full SEO acceptance. The
root `pendingComponent` uses the shared `@repo/ui/zenith-loader` through
`GlobalPending`; Auth uses the same screen loader through `RoutePending`.
Route-local pending states may
wrap the same component. `GlobalError` calls the supplied `reset()` and offers
a typed home link without exposing raw errors. `GlobalNotFound` handles missing
UI resources. Server health responses remain outside all three UI boundaries.

Motion is progressive only: every animated route has a static equivalent and
honors `prefers-reduced-motion`. Content and actions never require motion.

`appConfig` is static and server-safe. Do not add credentials, database URLs,
or `process.env` reads to it; runtime secrets continue through the YAML config
loader and server-only runtime modules.

For a future searchable UI route, define a Zod validator explicitly:

```tsx
export const Route = createFileRoute("/users/")({
	validateSearch: z.object({ q: z.string().optional() }),
});
```

For a dynamic resource, use `$userId/index.tsx` and call `notFound()` for a
missing UI resource; do not put Prisma or credentials in the route. A future
authenticated tree belongs under `_authenticated/route.tsx` and performs its
guard in `beforeLoad`.

## Generated ownership and tests

Never edit `src/routeTree.gen.ts` manually. The TanStack plugin regenerates it
from `src/app` using its supported generator; inspect generated
imports afterward. Tests and stories are excluded by `routeFileIgnorePattern`; `-` directories
are ignored by native routing. Route generation is authorized without repeated
permission. Prefer standalone codegen; use narrowly scoped dev/build only if
necessary for regeneration, not general verification. Inspect prerequisites,
never start dependencies, and stop agent-started dev processes afterward.
Follow [verification.md](../verification.md): UI uses Storybook
plus typecheck; server adapters/functions use typecheck only. Runtime auth/navigation
integration remains unverified. The native web typecheck is:

```bash
bun --no-env-file run --cwd apps/web check-types
```
