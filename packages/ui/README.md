# Shared UI

`@repo/ui` owns editable React source installed from
[Vandor UI](https://vandor-ui.vercel.app/llms.txt). The complete current catalog
(36 components), all 36 portable story items, optional Query helpers, and a local
semantic Table are installed by explicit user request. This package was renamed
from `packages/components` / `@repo/components`; active consumers use the new name.

## Usage

Branded loading: `ZenithLoader` from `@repo/ui/zenith-loader` uses CSS keyframes
to draw the two brand-owned SVG outlines from `@repo/brand/logo-mark`, then
reveal the theme-aware `ZenithMark`, hold and fade out over a repeating
2-second cycle. There is no scaling, rotation or Motion dependency in this component.
Reduced-motion users get the complete static logo without outlines. The
`label` is persistent accessible status text (default `Memuat…`), visually hidden
by default; set `hideLabel={false}` only when visible helper text is needed.
`mode="screen"` centers in a full viewport in normal flow; default `"container"`
centers within its parent's height. `size` defaults to 72 pixels. It performs no
fetching, authorization or simulated progress.

```tsx
import { ZenithLoader } from "@repo/ui/zenith-loader";

<ZenithLoader mode="screen" />;
```

For localized timestamps use `DateTime` from `@repo/ui/date-time`. It delegates to
`@repo/ui-utils`, accepts Intl `format`, `locale`, `timeZone` and `fallback`
overrides, preserves the ISO instant, and waits until hydration for browser-local
display. See [shared utilities](../ui-utils/README.md).

Declare `"@repo/ui": "*"` in the consuming workspace and import explicit entrypoints:

```tsx
import { Button } from "@repo/ui/button";
import { Input } from "@repo/ui/input";
import { Select } from "@repo/ui/select";

<Button type="submit" isLoading={pending}>
	Save changes
</Button>;
<Input label="Email" labelStyle="static" type="email" />;
```

```css
@import "tailwindcss";
@import "@repo/brand/tokens.css";
@import "@repo/ui/styles.css";
```

React 19/React DOM 19 are peers; Vite consumes TypeScript source directly. There
is no root barrel importing the catalog. Internal imports are relative, independent
of web's `@` alias. The installer-only alias lives in `components.json`.

## Public entrypoints

- Layout: `app-shell`, `page-container`, `page-header`, `stats-card`.
- Controls: `button`, `input`, `textarea`, `checkbox`, `radio-group`, `switch`, `select`,
  `autocomplete`, `input-amount`, `input-group`, `input-otp`, `input-password`,
  `input-phone`, `input-search`, `input-secret`.
- Dates: `calendar`, `date-picker`, `date-range-picker`.
- Navigation and overlays: `accordion`, `tabs`, `dialog`, `drawer`, `dropdown`,
  `popover`, `tooltip`, `command-palette`.
- Feedback and identity: `badge`, `blobatar`, `boring-avatar`, `empty`,
  `error-state`, `loading`, `toast`.
- Data: `collection`, `filter-bar`, `table`, `data-grid`, `data-grid/schema`, `money`, `copyable-id`, `json-view`.
- Display: `qr-code` (local SVG via `qrcode-generator`; no network, fixed black-on-white for scanners).
- Display type: `display-heading` (auth layouts and landing titles).
- Opt-in integrations: `autocomplete/query`, `toast/tanstack-query`.

Use `@repo/ui/<entrypoint>`. `data-grid/schema` is a separate server-safe Zod v4
entrypoint, not a database schema or an installed API. Helpers and the licensed
`loading-ui` snapshot stay internal. Query adapters are not mounted in the web app;
providers, backend operations, routing and authorization remain consumer-owned.

All type uses the sans family; headings in `PageHeader`, `StatsCard`, `Empty`,
the `AppShell` `sidebarContext` title and `DisplayHeading` are semibold with tight tracking.
`task check:ui-primitives` rejects serif utilities. Motion durations and easings come from `@repo/brand/motion` (JS)
or `--motion-*`/`--ease-out` (CSS); `Money` takes integer minor-unit strings and never floats.
`exponentOf(currency)` from `@repo/ui/money` exposes the same Intl currency exponent
for exact nominal-input conversion; invalid currency syntax returns `null`.
Currency-scale refund input remains an application concern.

## Contracts and migration notes

- **Select/Autocomplete labels:** floating labels use a transparent surface with a
  real outline notch, not an opaque background patch. Focus/invalid borders and
  gradient control surfaces remain shared defaults.
- **BO remote choices:** presentations use `Autocomplete` directly; the BO adapter
  connects the optional infinite-query helper with session-scoped private keys.
  Pagination accumulates deduplicated choices inside the dropdown, including
  loading/retry feedback. `UI/Autocomplete/Infinite Scroll` and `Infinite Scroll Retry`
  use local fixtures only, not a live BO transport.

- **Button:** preserves `isLoading`, `asChild`, `render`, `nativeButton` and
  `whileTap`. Loading disables actions, blocks link clicks and sets `aria-busy`.
  Use explicit `type="submit"` for submission, real links for navigation, and
  accessible labels for icon-only controls. A visual link variant is still a button.
  Native `disabled` alone does not disable an anchor.
- **Input:** now uses the full registry implementation, including floating/static
  labels and icons. The default label style is floating; use `labelStyle="static"`
  for the previous static treatment. Native props and React 19 refs pass through.
  File inputs must remain uncontrolled.
- **AppShell:** now uses the complete upstream API. `brand` is required. Main
  content is edge-to-edge: the consumer's `PageContainer` owns page padding
  (16px below `sm`, 24px from `sm`, 48px from `lg`), maximum width and centering.
  Do not add another padding layer to the shell. The optional topbar retains its
  own compact horizontal spacing.
  Navigation uses `href`, `onClick`, or `renderControl` rather than the retired local `link`
  property. Router integrations supply real links through `renderControl`.
  Context animation and optional consumer-owned search are available. Demo
  workspace/account state is local, not installed application functionality.
  `showHeader={false}` removes the topbar (used by BO): the desktop sidebar owns
  its collapse control. In the compact rail the brand changes to the expand icon
  on hover or keyboard focus, with an accessible label and tooltip. Mobile uses
  separate fixed controls: a top-attached brand tab with concave upper transitions
  and rounded bottom corners on
  the left and a floating navigation button on the right, without a header or
  full-width background. Safe-area padding/offsets and initial
  spacing keep controls clear of page titles; navigation opens a right-side drawer.
  The drawer retains Escape dismissal and focus return. `showHeader` defaults to
  `true` for consumers with header/actions/search content.
- **Drawer:** now supports directional panels, snap points, handles and compound
  composition.
  Supply an accessible title. Use `swipeDirection="left"` for left navigation;
  the registry defaults to `"right"`. The old left-navigation-only wrapper is retired.
- **Tooltip:** uses the full Base UI/Motion wrapper, including provider, trigger,
  collision-aware content and reduced-motion handling.
- **Loading:** exposes all 47 upstream variants with typed `variantProps`, labeled
  status and static reduced-motion fallback. Button imports only its internal Arc,
  not the full animation catalog. Preserve `src/components/loading-ui/LICENSE.md`.
- **Table:** remains the local semantic wrapper set (`Table`, `TableCaption`,
  `TableHeader`, `TableBody`, `TableFooter`, `TableRow`, `TableHead`, `TableCell`).
  Native props pass through; `TableHead` defaults to `scope="col"`. Consumers own
  overflow, data and navigation. This is separate from server-driven DataGrid.
- **DataGrid:** `useDataGrid` owns its Query integration; `useDataGridSource` accepts
  consumer-owned paginated rows, loading/error state and reload callback without
  creating a second Query cache or requiring a QueryClient provider. Both use the
  same controller, typed columns, pagination and presentation. BO uses the source
  adapter so its identity-scoped private Query lifecycle remains authoritative.
  Pass a single `pageSizes` entry to pagination for a fixed-size API; the redundant
  page-size selector is hidden. Sorting requires an actual backend sort contract.
- **Select:** the standard trigger accepts `label` and `labelStyle="floating" | "static"`
  (floating by default when a label is supplied). The visible label names the
  trigger accessibly; custom triggers retain ownership of their own label layout.
  Unlabeled existing consumers retain their previous appearance.
- **DateRangePicker:** reexports `dateRangeIndonesianLocale` from the already installed
  DayPicker locale, including Indonesian calendar accessibility labels. Consumers
  do not need a new date dependency. Apply/cancel/clear copy stays caller-owned;
  optional English shortcuts can remain disabled with the default empty features.

All application controls and application stories import shared primitives directly
from `@repo/ui/<name>`, not native controls, vendor primitives or application-local
copies. Structural HTML, links and SVG remain allowed outside this package.

## Local adaptations

- StatsCard balances items across full-width rows (five metrics: 3 + 2; six:
  3 + 3; seven: 4 + 3), with at most four columns on desktop. Tablet uses at
  most two columns (three for exactly three metrics); mobile stacks one per row.
  Separators follow each row's actual size, including partially populated rows.
- Explicit package exports, repository formatting, relative imports, and Tailwind 4
  source registration; no app routes or live API clients are installed.
- Corrected registry sibling imports in Autocomplete/DataGrid to reference `../ui`.
- Autocomplete positions its isolated portal at `z-50`, matching Select, so nested
  choice popups are not painted behind a modal drawer's viewport. Consumer popup
  z-index overrides cannot escape an unlayered isolated positioner.
- Preserved strict TypeScript: guarded indexed fixture/sort access and Tabs mode
  detection; an explicit InputAmount story meta type avoids an unnameable inferred
  third-party type. No strictness flags were disabled.
- Removed portable `autodocs` tags because the host has no Docs addon.
- Dark utility variants use `vandor-dark`, activated only by an explicit `.dark`
  ancestor, never automatically by the OS against light tokens. Brand provides
  the Zenith dark palette; no theme switch is installed.
- Zenith light/dark colors copied from `packages/brand` and additional catalog status
  roles are owned by `packages/brand/src/tokens.css`; shared CSS maps utilities
  without defining a duplicate palette. Button's
  destructive foreground uses the brand semantic token.
- Brand also owns `--radius: 0.8rem` and the shared Tailwind `sm`–`4xl` radius
  mappings. Existing component utilities follow that scale in web, BO and
  Storybook. Preserve circular shapes, explicit small geometry and square joined
  edges; do not repeat application-local radius tokens.
- `collection` and `filter-bar` are local: the collection pattern for web and BO (rules in
  `docs/frontend/FRONTEND_CONTRACT.md`). `Collection` renders a table or cards from one
  keyed column definition; a table stacks into rows by its own width (`@container`, `@2xl`,
  then a ResizeObserver keeps one layout mounted). The record link wraps the title and stretches
  over its row via `::after`. The plain `table` primitive stays for non-collection uses.
- `command-palette` is local, not from the Vandor catalog (it has none): Base UI's
  Autocomplete command-palette composition (inline Autocomplete in a Dialog). The consumer
  owns the items, the open state and what selecting an item does (normally navigation).
- `AppShell` `inset` puts the content in a rounded panel on the sidebar surface. It moves
  the desktop collapse toggle and search into the sidebar, and on desktop the panel, not
  the window, scrolls. The scroll element carries `data-scroll-restoration-id="app-shell-main"`;
  a consumer router resets it through `scrollToTopSelectors`. Inside the panel
  `--background` resolves to `--card`, so the panel is lighter than the chrome in both themes.
- Existing compatible Motion 14 and Base UI 1.9 dependency ranges are retained
  rather than downgraded to the registry's minimums.

## Install and update

Run from `packages/ui` using the configured `@vandor` registry:

```bash
bunx --bun shadcn@latest add @vandor/select --dry-run
bunx --bun shadcn@latest add @vandor/select --diff
bunx --bun shadcn@latest add @vandor/select
bunx --bun shadcn@latest add @vandor/select-stories
```

Review upstream changes before overwrite. Preserve local adaptations, license
notices, dependency closure and explicit exports. Future additions should be
need-driven; this full-catalog installation is the requested exception, not a
policy to bulk-install other registries.

## Preview and verification

```bash
task dev:storybook
bun --no-env-file run --cwd packages/ui check-types
bun --no-env-file run --cwd apps/web check-types
bun --no-env-file run --cwd apps/storybook check-types
task check:ui-primitives
```

Colocated stories appear under `UI/*`; brand token previews are under
`Brand/Identity`. Story fixtures use deterministic local data, not production APIs.
Follow [component ownership](../../.agent/components.md) and
[verification policy](../../.agent/verification.md): owner/consumer typechecks plus
desktop/mobile Storybook inspection, including relevant keyboard/focus and local
interaction checks. Use only the ownership-safe server on port 6006. Typecheck
does not establish visual quality, live integrations, authentication or persistence.
