# Recipes

Copy the shape of a real file, not a template. Every path below exists; if two
examples disagree, the real code wins. After any API change run `task sync`, then
`task check:quick` (see [verification.md](verification.md)). Never hand-edit generated
files (FX `app.go`, `contracts/openapi`, `generated/`, route trees, sqlc output).

## Backend

**Query endpoint, BO** (reads go through API-owned persistence, no usecase)

- Copy `apps/bo-api/internal/handlers/v1/organizations_get.go` + `organizations_read.sql`.
- `task create:handler app=bo-api name=<resource>_<action> kind=query method=GET path=/v1/...`
- Write the SQL beside the handler with `-- name: BOxxx :one|:many`, run `task gen:db`, scope by actor.
- Real handlers use an inline input struct with `ItemPath` and `customerError`/`security.Error`;
  the scaffold's named `...Input` type is only a starting point.

**Mutation endpoint, BO** (writes always go through a usecase)

- Copy `apps/bo-api/internal/handlers/v1/organizations_update.go`; the usecase is
  `packages/application/customers/organizations.go` (`UpdateOrganization`).
- `task create:handler ... kind=mutation`, then `task create:usecase domain=<d> name=<Verb>`.
- Usecase order: validate, open transaction, authorize, lock rows in fixed order, write, audit.
  The handler only maps input/errors. The scaffold writes one file per verb; existing domains
  group verbs per resource, either is fine.

**Query endpoint, Web API** (organization-scoped)

- Copy `apps/web-api/internal/handlers/v1/organizations_list.go` + `organizations.sql`.
- Path must contain `{organizationId}`; go through `orgaccess.Boundary`, never skip tenancy (ADR-0002).

**Public API** (`apps/api`, Bearer `zk_` keys, snake_case JSON, ADR-0007)

- Copy `apps/api/internal/handlers/v1/organization_get.go` (fixture-backed until the real-functionality phase).

**Dummy-phase endpoint** (ADR-0001, UI first)

- Copy `apps/web-api/internal/handlers/v1/api_keys_list.go`: data from `packages/fixtures`,
  list/page helpers in `dummy_page.go`. No real persistence in this phase.

**New domain** — `task create:usecase` registers it in `packages/application/sqlc.yaml`.
Write queries in the generated `.sql`, then `task gen:db`.

**Migration**

- `task create:migration name=<snake_case>`; edit the SQL (Up is the schema).
- Update `packages/db/schema.dbml` in the same change (tables, types, nullability,
  keys, indexes; PostgreSQL-only constraints as notes), then `task check:db-schema`.
- Behaviour changes: `task check:db:integration`.

## Frontend

**New screen (BO or Web)** — presentation first (CDD), then wiring.

1. Fixtures + typed presentation component + colocated story, checked in Storybook.
   Golden: `apps/bo/src/app/_authenticated/organizations/-components/organization-collection.tsx`
   - `.stories.tsx` + `-fixtures.ts`.
2. Route file that fetches with the generated SDK and passes props down. Golden:
   `apps/bo/src/app/_authenticated/organizations/index.tsx`. Routes are directory-first,
   English segments (`.agent/frontend/routing.md`); run `task sync` (or `task gen:bo-routes` /
   `task gen:routes`).
3. **Register the page in navigation**, the step most often forgotten:
   BO `apps/bo/src/app/_authenticated/-navigation.ts` (+ icons in `-components/navigation-links.tsx`,
   test `-navigation.test.ts`); Web `apps/web/src/app/w/$organizationSlug/-components/navigation.ts`
   (test `navigation.test.ts` beside it).
4. Shared primitives come from `@repo/ui/<name>`; add missing ones to `packages/ui` first.

## Check before saying done

`task check:quick`; focused tests when logic changed; stories reviewed in the browser for UI;
`task check:go` and `task check-types` at the end of a batch. State what was not verified.
