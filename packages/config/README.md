# Runtime configuration

`config/config.yaml` is the sole YAML 1.2 document. Sections: app, web, webApi, api, bo,
boApi, db, worker, storage, encryption. JS `@repo/config/server` and Go isolated loaders validate raw shape and
template syntax before selecting fields and interpolating once. Unknown/duplicate/
templated keys, aliases, anchors, explicit tags, merge keys and multiple documents
are rejected. Zod 4 and Go validator own typed validation; errors omit values.

`${NAME}` requires nonempty input; `${NAME:-fallback}` accepts missing/empty input;
`$${` escapes interpolation. Substituted strings are never rescanned. YAML 1.1
booleans/timestamps remain strings, leading zeros are decimal, nonfinite values
and YAML 1.1 directives are rejected. Process getters cache one snapshot until restart.

`getWebConfig` and `getWebToolConfig` are lazy; Web's `getPublicWebConfig` returns
exactly `{ apiBaseUrl }` from web.apiBaseUrl. The no-store public server function
returns only this schema. BO independently exposes exactly `{ boApiBaseUrl }`.
Browser imports only `@repo/config/public`, with no server/env/build injection.
Tool adapters emit only listener/NODE_ENV settings; public origins are runtime data.

Go LoadAPI selects public service settings, LoadWebAPI selects customer auth/dashboard
settings, LoadWebOrigin selects webApi.identity.authOrigin and its non-secret origin-validation
context for trusted public-API dashboard links, and LoadBOAPI selects staff settings; enabled
docs protection requires both
credentials, disabled credentials are never resolved. LoadIdentity selects
webApi.identity authOrigin, development, trustedProxyCIDRs, csrfKey, limiterKey and
email. authOrigin is the Web origin. CSRF/limiter keys are canonical base64 nonzero
32-byte values and distinct. Production origins are canonical same-site HTTPS;
explicit matching-loopback development is restricted to development/test.

LoadDatabase alone selects db.dsn, with no default/ambient credentials. Pool defaults:
10 max open (1–10000), lifetime 1800/idle 300 seconds (0–604800), connect timeout 10
seconds (1–300). No maxIdleConnections setting. DB module maps these explicitly;
no loader imports DB/tools. Use explicit PostgreSQL URL/TLS; no passfile/service discovery.

`LoadWorker`, `LoadStorage` and `LoadEncryption` select only their own fields.
Worker shutdown timeout is 1–300 seconds and max workers is 1–1000. Storage
presign TTL is 60–3600 seconds; endpoints require HTTPS and a host only, except
loopback HTTP in development/test. Encryption uses a scalar sequence of
`<id>:<base64>` entries with unique lowercase ids and canonical base64 nonzero
32-byte keys; the first key is active. These selectors do not resolve DB,
identity, API or each other's secrets.

Development launchers alone opt into root .env; injected environment wins. Production
uses injected configuration. Trusted --root /app or bootstrap ZENITH_APPLICATION_ROOT
locates YAML; Nitro copies unresolved config into its output. Never automatically
read or modify real .env. Restart to refresh snapshots.

`task check:auth:config` counts focused Go/JS isolation tests from owned temporary
roots and typechecks config. Required missing/skipped cases fail. See the verification
policy for permitted checks; no live runtime, existing DB or real credentials.
