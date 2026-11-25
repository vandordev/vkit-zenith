# ADR-0004: Backend layering: usecase writes, handler-owned reads

- Status: Accepted
- Date: 2026-11-25

## Context

Several callers (web dashboard, Backoffice, public API, worker) touch the same data.
Authorization, transactions and audit must not be re-implemented per caller.

## Decision

1. **Every business write goes through a usecase** in `packages/application/<domain>`
   (one flat package per domain with prefixed files, following `staff/` and
   `customers/`), whoever the caller is. Usecases own validation, authorization,
   ordered locks, the transaction and the audit event.
2. **Reusable queries live with the usecases**: anything used by a mutation or by more
   than one runtime, generated into `packages/application/internal/persistence`.
3. **Endpoint-only reads** are sqlc queries beside the handler
   (`apps/<api>/internal/handlers/v1/*.sql`), generated into that API's
   `internal/persistence` and registered in `scripts/generate-db.ts`. Reads are
   scoped by the authenticated organization (ADR-0002).
4. **No external call inside a DB transaction.** Commit the intent, call outside any
   transaction, then record the result in a new transaction with an idempotency key.
5. **Permissions** are a fixed Go map in `packages/application/customers/roles.go`:

   | Permission                                | owner | admin | member |
   | ----------------------------------------- | ----- | ----- | ------ |
   | `organization.read`                       | ✓     | ✓     | ✓      |
   | `members.manage`                          | ✓     | ✓     | –      |
   | `developers.manage` (API keys)            | ✓     | ✓     | –      |
   | `organization.manage` (profile, branding) | ✓     | –     | –      |

   Products add permissions to this map with their features.

## Alternatives considered

- **All SQL in usecases, including list/detail reads:** a pass-through layer with no
  authorization or transaction value.
- **Handlers writing directly with sqlc:** spreads authorization, transactions and
  audit across callers.

## Consequences

- `task create:usecase` and `task create:handler kind=query|mutation` scaffold exactly
  this split.
- Usecases get focused tests against owned temporary PostgreSQL.

## Revisit when

A second API needs the same reads (move them into `application`), or a domain package
grows past easy navigation (split by subdomain).
