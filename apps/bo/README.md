# Backoffice

TanStack Start owns staff/customer routes with Indonesian UI and actual presentation
stories. Browser calls independent BO API directly; no BFF, cookie relay, SSO or
SSR authentication fetch. Client route checks are UX only. The public config
server function returns exactly `{boApiBaseUrl}`, no-store, selected from the
strict unresolved runtime YAML. Web remains exactly `{apiBaseUrl}`.

## Ownership

Each route/layout owns its typed presentation and colocated stories in `-components`;
only genuinely shared compositions remain in `src/app/-components`. Storybook uses
`BO/Pages/<Area>/<Page>`, `BO/Layouts/<Layout>` and `BO/Components/<Composition>`.
URLs use English: `/activity`, `/staff`, `/account`, `/organizations`, `/users`,
with UUID detail routes for retained customer records.
Indonesian labels and feedback are independent of URL vocabulary. The development
route rename replaces the earlier Indonesian paths without compatibility aliases.
Follow [routing](../../.agent/frontend/routing.md) and
[stories](../../.agent/frontend/stories.md) for all future TanStack Start work.
Routes own current identity, pagination/validated URL state,
generated SDK calls, request-body Zod validation and invalidation.
BO presentation may use declarative `Link` from `@tanstack/react-router` directly;
imperative navigation stays in route adapters. All `BO/` stories receive the host's
shared in-memory router, without production routes, auth providers or API clients.
Shared UI primitives remain router-agnostic. Shared controls
come directly from `@repo/ui/*`; identity/tokens from `@repo/brand`. Published
versions are inspection-only, feature omission is explicit, no subscription
operations exist. Customer administration is described below. Money and bigint
limits remain strings end-to-end.

The explicit browser transport uses credentials include/no-store, memory-only
CSRF and identity-generation fencing. All private Query keys include identity and
epoch; cancellation/clear on401/logout/self-revocation prevents prior-user data
reuse. No auth/localStorage, durable password draft or mutable default SDK client.
Failed unsafe mutations are never automatically replayed.401 clears server
cookies; the next explicit login fetches fresh anonymous CSRF.403 invalidates
cached CSRF but remains a truthful denial. Conflicts require explicit data reload,
preserving safe local input. Currency selection traverses all paginated metadata.

## Commands and evidence

### Customer administration

Both roles have `/organizations`, `/organizations/$organizationId`, `/users` and
`/users/$userId`, labeled Organisasi and Pengguna pelanggan, distinct from Staf.
Lists use exact backend decimal-string summaries, bounded filters and pagination.
The browser calls generated BO SDK operations directly. Route adapters own identity-
scoped cancellable queries, strict generated request schemas, explicit writes and
cross-context allowlisted return URLs; passwords/reasons never enter route state.
Conflicts preserve safe local drafts but require explicit reload and fresh password.

Admin can create an organization and its initial owner atomically (existing active
customer or new active/unverified customer), edit profiles, add admin/member roles,
remove non-owners and explicitly restore the same removed membership. Restoration
does not restore revoked authorization or ownership. Superadmin alone, with a reason
and current password for each action, can change subsequent roles (owner/admin/member), suspend or
reactivate organizations, disable/reactivate customers, replace email or manually
verify the current email. Last eligible owner protection includes suspended
organizations. Demotion and removal are separate actions.

Customer creation supplies no password, session, SSO or login-ready guarantee.
Email replacement clears verification and revokes applicable persisted authorization;
reactivation/manual verification never revive revoked material. No external product logout guarantee or product/billing access is supplied. Customer activity
is productless and uses current labels, safe changed-field keys, no email history
or raw reasons. Membership links use authoritative organization+membership IDs.

See [customer brief](../../docs/frontend/pages/bo-customer-operations.md) and
ledger.
`task check:bo:customers` permits focused owned isolated PostgreSQL/in-process
acceptance only; it does not prove live browser→API, issuer, TLS/cookie or deployment
integration. Storybook fixtures and owner typechecks are the frontend gate.

`task dev:bo` is an operator command on4103, not agent verification. Development
API transport additionally needs the explicit matching-loopback configuration
documented by [BO API](../bo-api/README.md). `task gen:bo-routes` and
`task gen:bo-client` are offline. `bun --no-env-file run --cwd apps/bo check-types`
checks owner/stories; Storybook6006 supplies actual desktop/mobile/keyboard UI
acceptance without live API calls. See [brief](../../docs/frontend/pages/bo-runtime.md)
and ledger.

`Dockerfile.bo` is source-reviewed only: nonroot Bun, committed SDK/routes,
unresolved YAML copied server-side, no build-time URL/secret injection. No image
build or deployed TLS/cookie/proxy/browser integration is claimed. Production
requires operator-provided same-site HTTPS origins, not third-party cookies.

## Deployment

`apps/bo/vercel.json`, `Dockerfile.bo` and `Dockerfile.bo-api` are source references.
Set runtime `BO_API_BASE_URL` to the product's BO API origin. Create the first
Superadmin from an operator terminal after migrating:

```sh
docker compose exec bo-api /app/staff-bootstrap --root /app
```

Enter the email, display name and a 15–128 character password interactively.
Bootstrap is one-time only; never substitute direct SQL or pass the password as an argument.
