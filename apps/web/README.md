# Web

TanStack Start owns the static foundation `/`, local `/health`, customer login,
registration, check-email, verify-email, forgot/reset/set-password and signed-in
`/w`. Browser identity requests call `apps/web-api` `/v1/auth/*` directly.

## Organization dashboard

`/w` resolves real memberships and opens `/w/$organizationSlug`. No-organization,
unknown membership and suspended-organization states have explicit presentation.
Ringkasan (organization, role, active members) and Pengaturan (profile, members,
API keys) call Web API directly, with generated consumer-owned types and no
business proxy. Private query keys include identity epoch and organization;
placeholders retain rows only within that scope. Permissions come from the current
membership, while server checks remain authoritative.

API keys and logo metadata are dummy (ADR-0001): results are labelled
`Simulasi (data contoh)` until navigation. The one-time API-key secret is an
unusable example shown only in local creation-dialog state, never a query or
mutation cache, URL, storage or logs. Logo upload sends content type and size only.

Component stories live beside dashboard features under
`w/$organizationSlug/**/-components`, grouped as `Web/Pages/Dashboard/*` and
`Web/Components/Dashboard/*`; `Web/Layouts/OrgShell` previews the shell. Stories
use local fixtures and the shared memory router, never production clients.

The no-store `getPublicConfig` returns exactly `{ apiBaseUrl }` from `web.apiBaseUrl`.
Web API `webApi.identity.authOrigin` carries this Web origin. The provider keeps CSRF in
memory, scopes private queries to the identity epoch and never gates the public
foundation. Email link tokens are scrubbed before router search state. Auth pages
retain noindex metadata and no-referrer responses. No SSR authentication fetch,
business proxy, persistent tokens or configured SSR-global client.

Routes are directory-first with English URLs. Presentation and stories stay
feature-local; global auth layout/boundaries remain under `-components`, dashboard
entry presentation under `w/-components`. Stories use `Web/Pages/...`,
`Web/Layouts/...` and `Web/Components/...`. Shared controls come from `@repo/ui/*`.

`task gen:routes` owns the route tree; `task gen:web-client` generates the SDK from
`contracts/openapi/web-api.json`. Never hand-edit generated files. `task dev:web` is
an operator launcher, not agent verification. Use typecheck, UI primitive checks
and Storybook on 6006; counted `task check:auth:*` and
`task check:web-api:security` gates provide focused in-process/fake-data coverage.
Written UI and fixture unit tests outside those gates are not executed under the
verification policy. Storybook evidence is finite presentation coverage, not proof
of live routes, browser login, cross-origin cookies/CORS, deployment or real mail.

## Deployment

`apps/web/vercel.json` and `Dockerfile.web` are source references. Set runtime
`API_BASE_URL` to the product's Web API origin; production requires same-site HTTPS.
