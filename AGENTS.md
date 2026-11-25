# AGENTS.md

Map for agents. Read only what your task needs; detail lives in the linked files.

## Start here

Run `git status --short` and preserve unrelated changes. Product scope work: read
`PRODUCT.md`, `docs/adr/README.md` (Accepted ADRs are binding, Proposed ones need user
approval) and the current spec/plan in `docs/superpowers/{specs,plans}`. Dispatched
workers take their brief as the plan and read only the files it names.

- Do not silently change agreed scope; discuss anything outside the spec first.
- A decision that contradicts an Accepted ADR needs a new superseding ADR.
- Delivery follows ADR-0001: schema, then UI with dummy-data endpoints, then real functionality.

## Which file to read

| Task                         | Read                                                                                                                |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| Endpoint, usecase, migration | `.agent/recipes.md`, `.agent/backend/README.md`                                                                     |
| BO or Web screen             | `.agent/recipes.md`, `.agent/frontend/README.md` (`routing.md`, `stories.md`), `docs/frontend/FRONTEND_CONTRACT.md` |
| Shared UI primitive          | `.agent/components.md`, `packages/ui/README.md`                                                                     |
| Config / env                 | `.agent/config.md`                                                                                                  |
| Landing / SEO                | ADR-0008..0012, `.agent/frontend/seo.md`, `apps/landing/README.md`                                                  |
| Brand                        | `packages/brand/README.md`                                                                                          |
| Architecture overview        | `.agent/architecture.md`                                                                                            |
| Verifying anything           | `.agent/verification.md`                                                                                            |

## Repository shape

- `apps/web` customer auth + organization dashboard (TanStack Start); `apps/bo` staff UI;
  `apps/landing` static marketing (Next.js export, no API/auth/env); `apps/storybook` previews.
- `apps/web-api`, `apps/bo-api`, `apps/api`: independent Go Echo/Huma services (one API per
  consumer, ADR-0006); `apps/worker` fails fast until a product registers a job (ADR-0003).
- `packages/application` usecases (authorization, transactions, audit); `packages/db` Goose
  migrations + `schema.dbml`; `packages/fixtures` dummy data (ADR-0001); `packages/ui` shared
  UI; `packages/brand` identity; `packages/config/{js,go}` typed config; `packages/tools`
  generators and scaffolds; `contracts/openapi` committed schemas.

## Rules the code relies on

- **UI primitives:** applications import React primitives from `@repo/ui/<name>` and never render
  native `button`, `input`, `select`, `textarea`, `label`, `dialog`, `details`, table elements,
  `progress`/`meter`; a missing primitive is added to `packages/ui` first
  (`.agent/components.md`, checked by `task check:ui-primitives`).
- **Handlers:** exact `//vkit:handler` constructors, one operation per file, no `routes.go`.
  Never hand-edit generated files; regenerate with `task sync` (or `task gen`).
- **Layering (ADR-0004):** queries use API-owned persistence; every business mutation goes
  through an application usecase. Handlers never write business data.
- **Schema:** Goose migrations are canonical; update `packages/db/schema.dbml` in the same
  change and run `task check:db-schema`. Never automigrate.
- **Config:** one `config/config.yaml`, typed loaders, no feature-level env reads, browser reads
  only `@repo/config/public`.
- **UI work:** component-driven; typed presentation components with colocated stories, verified in
  Storybook, then routes and generated clients. English URL segments.

## Working

Prefer `task` commands (`task --list`). Verify by risk per `.agent/verification.md` and report
exactly what ran and what is unverified. Commit finished work with a conventional message;
never merge or push automatically.
