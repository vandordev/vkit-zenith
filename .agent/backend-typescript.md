# Server TypeScript boundary

TypeScript server code is limited to TanStack Start's web hosting, local health,
safe logging, and runtime public-config adapter. There is no TypeScript business
API, tRPC router, ORM/client, worker, scheduler or realtime runtime.

Use `@repo/config/server` only in server-only modules; public schemas/types come
from `/public`. `getPublicConfig` allowlists `{ apiBaseUrl }`, sets no-store
before validation, and sanitizes failures. No business proxy or serialization of
full config. Browser calls use consumer-owned generated Hey API and explicit
runtime client instances, never a configured SSR-global singleton.

Feature business logic and future persistence architecture require an approved
design, not revival of the removed product layers. Follow `verification.md`:
fresh native TypeScript checks only for this boundary, no tests, backend servers,
general builds or live API probes. Report runtime behavior as unverified.
