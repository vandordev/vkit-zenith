# TanStack Start presentation and stories

Applies to **all current and future TanStack Start apps**. Read [routing](routing.md),
[component ownership](../components.md), and [host guidance](../../apps/storybook/README.md).
Follow the feature-local organization demonstrated by the sibling Universe project,
not its branding, backend architecture, dependencies or verification commands.

## Ownership

- Colocate each presentation component and its `*.stories.tsx` in its owning
  route's `-components`. Layout presentation belongs beside its layout. Use English
  kebab-case filenames such as `version-editor.tsx` / `version-editor.stories.tsx`.
- `index.tsx` and `route.tsx` are adapters, not Storybook entrypoints. Stories render
  the same typed prop-driven presentation as the app, not production routes/providers
  or a duplicate page implementation. Keep fetching, auth and imperative navigation
  outside it.
  BO and Web dashboard presentation may use declarative TanStack Router `Link` directly; their stories
  get the shared in-memory router from the host, not a production route/provider.
- No catch-all application-root `catalog.tsx`, `access.tsx`, `bo.stories.tsx` or
  `runtime.stories.tsx` combining unrelated features. Genuinely shared compositions
  live at their nearest common owner; small primitives stay in `packages/ui`.
- Keep feature fixtures local in `-fixtures.ts` or inside `-components`. Only shared
  deterministic facts belong at app root. Fixtures/helpers never import another
  story or enter production adapters. Prefix ignored route-tree helpers with `-`.

## Sidebar and CSF

- One coherent component/task per CSF file, with unique English titles:
  `<App>/Pages/<Area>/<Page>`, `<App>/Layouts/<Layout>` or
  `<App>/Components/<Composition>`. Current prefixes: `Web`, `BO`.
  Examples: `BO/Pages/Auth/Login`, `BO/Pages/Staff/Collection`,
  `BO/Pages/Plans/VersionEditor`, `BO/Layouts/Shell`.
- `component` identifies what the story actually renders. Use
  `satisfies Meta<typeof Component>` and `StoryObj<typeof meta>`; never type every
  unrelated preview as `LoginForm`. Shared forms/collections may have feature-local
  stories with their actual component and that feature's props.
- Put common props in `args` and state differences in small overrides. Export
  `Default`/`Populated`, `Pending`, `Empty`, `NoMatches`, `Error`, `Conflict`,
  `Denied`, `ReadOnly` or meaningful domain states; avoid repeating page names.
  Story titles/identifiers are English; UI copy retains its approved language.
- Disable controls for callbacks, JSX/provider values and complex fixture objects
  with `argTypes` when controls would mislead. Render functions/decorators only add
  local interaction/context and forward typed args to the actual component.

## Fixtures and acceptance

- Inspect actual DTOs/runtime support first. Fixtures use deterministic IDs/dates,
  fake non-sensitive values and local state only, never live APIs or credentials.
  Keep string money/bigint exact. Never persist or display submitted passwords.
- Preserve meaningful existing regression states when reorganizing. Cover relevant
  populated/pending/error/empty/no-match/denied/conflict/readonly/long-content states.
- Give local observable interaction feedback when useful without claiming API
  persistence/login. Preview links must not enter real application/API flows.
- Restart repository Storybook safely after discovery changes. Inspect sidebar,
  desktop/mobile previews, keyboard/focus, overlay behavior and overflow; run native
  affected typechecks. Follow [verification policy](../verification.md); these rules
  do not authorize product servers, backend stacks or broad build/lint/test gates.
