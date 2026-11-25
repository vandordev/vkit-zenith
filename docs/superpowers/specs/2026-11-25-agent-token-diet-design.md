# Agent token diet — design

Status: draft for approval · 2026-11-25

## Intent

Zenith is developed entirely by AI agents. Today every session pays for ~11 KB of
always-loaded `AGENTS.md` that mandates reading README, `.agent/**` and ADRs
(~100 KB, roughly 25k tokens) before work starts. `.agent/verification.md` (14 KB)
is mostly six dated valcentra-era exceptions to a "never run tests" policy, which
also makes agents stop and ask for permission. Generators save hundreds of tokens
per file; context reading costs tens of thousands per session.

Goal: an agent starting a typical task reads ≤ ~6 KB of instructions plus the one
recipe and example files it needs, verifies with one quiet command, and never has
to hand-edit wiring that a script can do. Real-world variance stays with the agent
(business logic, queries, error mapping), not with more templates.

Anti-goals: losing a safety rule (no existing/user DB, no real `.env`, no deploy,
no remote migrations); losing an architectural rule that the code relies on
(handler/usecase layering, UI primitives, DBML parity, config isolation, ADR-0001).

## Decisions (user)

- Verification becomes **risk-based** (user's global policy): focused tests and
  owned Testcontainers PostgreSQL may run without asking; compile/typecheck remain
  the floor; Storybook for UI. Safety boundaries stay.
- Approach **A**: docs first, minimal tooling. No `create:screen` generator yet.
- Work on `main`, one commit per batch.

## Approaches

- **A (picked)** — shrink docs into a map + recipes pointing at real golden code;
  add only mechanical tooling: sqlc registration, `task sync`, `task check:quick`.
- **B** — A plus `create:screen`. Rejected for now: frontend patterns vary more than
  handlers and templates already drift from real code (template uses a named
  `Input` struct; 14 BO handlers use inline struct + `ItemPath`). Revisit once the
  BO screen recipe is stable.
- **C** — docs only. Rejected: `sqlc.yaml` registration and the generate chain are
  exactly the steps agents forget.

## Design

### 1. `AGENTS.md` → map (target ≤ 4.5 KB; landed 4.3 KB, down from 11 KB)

Keep, compressed: start-here (git status, preserve unrelated changes), foundation
documents (PRODUCT.md, ADR index, spec/plan), ADR-0001 phasing, repository shape
as a one-line-per-owner table, and the binding rules that code relies on (UI
primitives, `//vkit:handler` + no `routes.go`, never hand-edit generated files,
queries via API-owned persistence / mutations via usecases, DBML parity with
migrations, config isolation, no feature-level env reads). Then a routing table:

| Task                                   | Read                                                                             |
| -------------------------------------- | -------------------------------------------------------------------------------- |
| Backend endpoint / usecase / migration | `.agent/recipes.md` + `.agent/backend/README.md`                                 |
| BO or Web screen                       | `.agent/recipes.md` + `.agent/frontend/README.md` (+ `routing.md`, `stories.md`) |
| Shared UI primitive                    | `.agent/components.md`, `packages/ui/README.md`                                  |
| Config                                 | `.agent/config.md`                                                               |
| Landing / SEO                          | ADR-0008..0012, `.agent/frontend/seo.md`                                         |
| Brand                                  | `packages/brand/README.md`                                                       |
| Verifying anything                     | `.agent/verification.md`                                                         |

Detail currently duplicated in AGENTS.md that already lives in a `.agent` file or
README is removed from AGENTS.md, not rewritten. `CLAUDE.md` stays a pointer.

### 2. `.agent/verification.md` → risk table (target ≤ 4 KB; landed 3.7 KB)

- Table: change type → required check → command (docs, Go handler/usecase, SQL/
  migration, TS logic, UI presentation, generators/tooling, auth/security).
- Allowed without asking: compile/typecheck, `go test`/`bun test` for affected
  owners, the existing `task check:*` suites (incl. Testcontainers), offline
  generators, route generation, the repository Storybook on 6006 (keep the
  single-instance safeguard in one short paragraph).
- Never without explicit instruction: existing/user databases, real `.env` or
  secrets, operational/remote migrations, deploys, Compose stack, real mail,
  running backend listeners against real config.
- Reporting rule kept: blocked/skipped ≠ pass; name runtime-unverified boundaries.
- Resource rule: run only the affected Testcontainers suite, never several at once
  (valcentra ran up to two parallel workers on one machine).
- Anti-accretion rule: no feature- or milestone-specific exceptions; a new kind of
  check adds one table row. (Valcentra milestones each appended their own
  exception, which is how the file reached 14 KB.)
- The six dated exceptions are deleted (git history keeps them). Supervisor/
  counted-gate mechanics stay documented where they live (`packages/tools/authtest/README.md`,
  the scripts), not here.
- New **ADR-0013 "Risk-based verification"** records the change (repo rule: new
  durable decisions get an ADR); ADR index updated.

### 3. `.agent/recipes.md` (new, target ≤ 5 KB)

One section per task type, each: golden files to copy, ordered steps, check command.

- BO query endpoint → `apps/bo-api/internal/handlers/v1/organizations_get.go` +
  `organizations_read.sql`.
- BO mutation endpoint → `organizations_update.go` + usecase in
  `packages/application/customers/organizations.go`.
- Web API query → `apps/web-api/internal/handlers/v1/organizations_list.go` + `organizations.sql`.
- Public API → `apps/api/internal/handlers/v1/organization_get.go`.
- Dummy-phase endpoint (ADR-0001) → `apps/web-api/internal/handlers/v1/api_keys_list.go`
  (data from `packages/fixtures`, helpers in `dummy_page.go`).
- Screen recipes end with registering navigation: BO
  `apps/bo/src/app/_authenticated/-navigation.ts`, Web
  `apps/web/src/app/w/$organizationSlug/-components/navigation.ts` (valcentra
  reworked these 11× and 6×).
- New domain usecase → `task create:usecase`, then the customers example for shape.
- Migration → `task create:migration` + DBML update + `task check:db-schema`.
- BO screen → `apps/bo/src/app/_authenticated/organizations/` (`index.tsx`,
  `-components/organization-collection.tsx` + story, `-fixtures.ts`).

Steps reference `task create:*`, `task sync`, `task check:quick`. Exact example
files may be swapped by the builder if a better-shaped one exists; each must exist.

### 4. `create:usecase` registers the domain in `sqlc.yaml`

When `Usecase` creates a new domain's `service.go`, it also appends the domain to
`packages/application/sqlc.yaml` `queries:` (text insertion after the last list
entry, idempotent: skip if already present). Template comments that tell the user
to register manually are updated. Covered by a scaffold unit test (append, already
present, no `queries` key → error). `gen:db` is **not** auto-run: it succeeds on
the stub (verified 2026-11-25 in a temp copy, sqlc exit 0) but generates nothing
useful until queries exist; the recipe says to run it after writing them.

### 5. `task sync`

`task sync` = `gen:handlers` → `gen:openapi` → `gen:bo-client` + `gen:web-client` →
`gen:routes` + `gen:bo-routes`. Existing `sync:web|bo|api` and `gen` stay.
`scripts/taskfile.test.ts` list updated.

### 6. `task check:quick`

`scripts/check-quick.mjs` (Node, like the other check scripts):

1. Changed files = `git status --porcelain=v1 -z -uall` (staged, unstaged and
   every untracked file — without `-uall` a new directory shows only as `dir/`).
2. Go files → nearest `go.mod` dir → `go vet ./...` there (vet type-checks every
   package including tests, so a separate `go build` adds nothing; baseline clean
   on all 12 modules). A change in a `packages/*` Go module also vets its
   consumers `apps/{web-api,bo-api,api,worker}`.
3. TS/TSX files → owning workspace → that workspace's `check-types` (baseline
   clean, 2–21 s each). A change in a shared `packages/*` workspace also checks
   its consumers `apps/web` and `apps/bo`. Their `check-types` already include
   `check:ui-primitives`; `apps/landing|storybook` changes add it as a root step.
4. Each command's output is captured; on success print one line
   `ok: <n> checks (<owners>)`; on failure print only that command and its output,
   exit non-zero. No changes → `ok: nothing changed`.

Owner mapping is pure and covered by a small test (`scripts/check-quick.test.mjs`).
It does not replace focused tests; the verification table says when to add those.

## Simplest alternative considered

Only deleting text from AGENTS.md/verification.md. Cheapest, and it delivers most
of the token saving, but leaves agents re-discovering which file to copy and the
forgotten `sqlc.yaml`/sync steps, which cost retries. Recipes + two small tools are
the minimum that removes those loops.

## Not doing

- `create:screen` or any new generator beyond the sqlc registration.
- Changing the handler/usecase templates to match the inline-struct style (note
  the drift in recipes; fix later if it bites).
- Rewriting `docs/frontend/*`, package READMEs or ADR bodies (only routed to).
- Fixing the known pre-existing failing bun tests.
- Changing CI, hooks or the Go/TS check implementations themselves.

## Built now on purpose

- ADR-0013, because verification policy is binding and later agents must find why
  the exceptions vanished.
- Idempotent `sqlc.yaml` registration with a test, because a duplicated or
  misplaced entry silently breaks `gen:db` for every domain.

## Left to the builder

Exact wording and section order of the shrunk docs; which existing check suites
appear in the verification table; turbo filter syntax for `check:quick`.
