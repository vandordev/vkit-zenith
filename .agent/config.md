# Configuration

`config/config.yaml` is the sole YAML 1.2 runtime document, with sections `app`,
`web`, `webApi`, `api`, `bo`, `boApi`, `db`, `worker`, `storage`, `encryption`. Implementations live in `packages/config/{js,go}`; read
that package's README before changing grammar or selectors. Validate raw shape
and template syntax before selecting fields and interpolating original values
once. Reject unknown/duplicate/templated keys, aliases, anchors, explicit tags,
merge keys, includes and multiple documents. Preserve sequences and deliberate
YAML 1.2 scalar parity. Never rescan env substitutions.

Use lazy `@repo/config/server` getters in server-only TypeScript; Go bootstrap
calls `LoadWebAPI`, `LoadAPI` or `LoadBOAPI` with explicit options. Only consumer-selected
fields are resolved; one service must not require another's credentials. Enabled
docs Basic Auth requires username/password; disabled credentials are unused.
Diagnostics contain file/runtime/field and safe reasons, never resolved values.
`LoadWorker`, `LoadStorage` and `LoadEncryption` resolve only their own fields.
Storage permits HTTP only for loopback hosts in development/test. Encryption keys
use unique `<id>:<base64>` scalar entries containing canonical nonzero 32-byte keys.
Snapshots last until restart. Use Zod v4 and explicit Go validator instances.
Explicit Go DB consumers use `LoadDatabase`; other selectors never resolve DB
secrets. Map the five fields explicitly to DB Settings at the operational adapter.
No config→DB dependency, default DSN, ambient PG credentials or implicit passfile/
service discovery. Runtime accepts explicit PostgreSQL URLs/TLS; read owner docs.

Browser imports only `@repo/config/public` schema/types. `getPublicConfig` returns
exactly `{ apiBaseUrl }`, no-store before validation so failures are covered.
BO's isolated public config is exactly `{ boApiBaseUrl }`. Its server/Go selectors
resolve only BO fields; BO API DB selection remains separate. Production requires
canonical same-site HTTPS origins and injected CSRF/docs/DB secrets. Explicit
matching-loopback development is never a production fallback.
No env/build injection, real-config serialization or persistent browser config.
Web hosts customer auth; its public config contains only apiBaseUrl.
LoadIdentity selects Web API identity/email/HMAC settings, never DB credentials.
Web API webApi.identity.authOrigin is the Web origin.
Configuration failure leaves static foundation content visible and invents no URL.

Development adapters explicitly load root `.env`, injected values take precedence.
Deployment uses injected environment only. Trusted `--root /app` or bootstrap-only
`ZENITH_APPLICATION_ROOT` locates YAML independently of bundled source paths.
Vercel server bootstrap uses the function root; Nitro copies unresolved YAML into
its server directory. Narrow web tooling emits only listener/NODE_ENV values,
never flattened config. Feature code must not read env. Build/imports do not read
runtime YAML/secrets. Never read or modify real local `.env` automatically.

Verification: affected TS checks/Go compilation plus approved focused database
configuration/isolation fixtures under `.agent/verification.md`. No real `.env`,
existing database or unrelated parity suites. Deployment remains unverified.
