# Repository Storybook

`@repo/storybook` is the shared React/Vite preview host. Stories stay next to
the components they demonstrate in `apps` and `packages`, not copied here.

```bash
task dev:storybook
task build:storybook
```

Development uses `http://127.0.0.1:6006`. CI mode refuses an occupied port rather
than selecting a fallback. Static output is ignored at `apps/storybook/storybook-static`.
Storybook is development tooling, not a Compose/deployed product runtime.
Commands do not load `.env` or provision PostgreSQL/Redis.

## Discovery and styling

`.storybook/discovery.ts` discovers repository-wide `*.stories.tsx`, excluding
dependencies and generated outputs. Each file is registered individually. New
story files require a server restart to refresh discovery; existing edits use HMR.
Use unique CSF titles. This host supports React only.

Preview imports the actual `apps/web/src/styles.css`, with explicit Tailwind
sources for web and packages. It retains the existing light theme, without
inventing dark tokens or redesigning components. The `@` alias resolves to web's
`src` for existing components. Future shared packages should use their own public
entrypoints/relative internal imports, not the web alias. If source/styles move,
update preview paths and Tailwind sources accordingly.

Landing presentation under `apps/landing/src/app/_components` is grouped under
`Landing/Pages/*`. Discovery includes its colocated stories and excludes Next.js
`.next` and static `out` outputs. The preview registers landing Tailwind sources;
presentation components remain free of Next.js imports.

BO presentation is discovered from feature-local `-components` throughout
`apps/bo/src/app`, grouped under `BO/Pages/<Area>/<Page>`, `BO/Layouts/<Layout>`
and `BO/Components/<Composition>`, not a catch-all `BO/Runtime` group.
The separate `@bo` alias and explicit BO Tailwind source coexist with web's `@`
alias. Generated BO types may be imported type-only; no runtime SDK/provider,
credentials or live API requests belong in a presentation story.

Customer auth presentation under `apps/web/src/app` is grouped `Web/Pages/<Page>`,
`Web/Layouts/AuthLayout` and `Web/Components/LinkOutcome`; its anchors use Auth paths
and stay inside the preview (`withAuthLinks`), with no providers or API clients.

All `BO/` stories and the Web organization dashboard titles (`Web/Pages/Dashboard/`, `Web/Layouts/OrgShell`, `Web/Components/Dashboard/`) receive one shared TanStack Router decorator with memory history; external `http(s)` anchors stay blocked there.
BO presentation can import `Link` directly without link-only substitution props or
per-story router setup. Each story gets a fresh router; destinations keep rendering
that story and never mount application routes, loaders or auth/API providers.
Use `parameters.routerInitialPath` for a non-root starting URL. Ordinary clicks on
`Link` stay in memory; existing dummy anchors remain local fixtures (not converted
automatically). Keep fixture anchor navigation prevented; do not open real app/API
flows. Web and shared UI stories are unaffected; shared UI stays router-agnostic.

Stories are typechecked/linted by their owning workspace; the host checks its
configuration and discovery tests. `task build` builds web only; Storybook has an
explicit `task build:storybook`. Its Turbo build cache is disabled because discovery spans workspaces beyond
its declared dependency graph.

## Component-driven workflow

Follow [verification policy](../../.agent/verification.md): check schema and runtime
feasibility, build typed presentation components with stories, verify them here,
then integrate fetching/auth/navigation in route adapters. Pages/layouts use
`Web/Pages/<Area>/<Page>` and `Web/Layouts/<Layout>` titles under route-local
`-components`. Shared UI Button and its portable stories live in
`packages/ui`, under `UI/Button`. The obsolete web-local Button
and `Web/UI/Button` stories are removed; error/not-found boundaries use the shared
Button and are previewed under `Web/Components/RouteBoundaries`.
The full UI catalog has colocated portable stories. Loading previews under `UI/Loading`,
including its full variant catalog, sizing, text, variant options, and Button usage.
See [component ownership rules](../../.agent/components.md) and
[TanStack story organization](../../.agent/frontend/stories.md), which apply to all apps.

Fixtures are deterministic and local, respecting actual schema/runtime contracts.
All fixture controls must use direct `@repo/ui/<name>` imports, just like
application controls; no native controls, local copies or direct Base UI/Radix
imports in `apps`. Native implementations belong in `packages/ui` only.
`task check:ui-primitives` scans application stories as part of root/web typechecks.
Stories compose their own providers when needed, never production credentials
or real authenticated APIs. Review relevant states, copy, keyboard/focus,
interactions, and mobile overflow. Build/typecheck alone are not visual acceptance.

## Agent server lifecycle

Check existing `http://localhost:6006` before UI verification. If absent or stale,
start/restart of this repository's instance is authorized without repeated permission.
Before start/restart, inspect all Storybook processes/listeners,
identify the old project instance, stop it and confirm exit before starting
`task dev:storybook`. Verify exactly one server, including possible duplicates
on other ports. Never broadly kill unrelated processes, choose another port, or
substitute an alternate dev/static preview. Ask if ownership is unclear.

## Focused checks

```bash
bun --no-env-file test apps/storybook/test
bun --no-env-file run --cwd apps/storybook check-types
bun --no-env-file run --cwd apps/web check-types
task build:storybook
```

The commands above are tooling references, not automatic verification gates.
Follow [verification policy](../../.agent/verification.md): UI uses affected native
typechecks plus browser Storybook inspection; no default test/lint/build gate.
System/backend/service changes use typecheck only; never start Compose or product
services for previews. Storybook/browser checks do not establish production auth
or backend integration behavior. Report those runtime boundaries as unverified.
