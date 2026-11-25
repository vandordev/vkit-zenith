# Verification

Verify by risk: run the cheapest check that can actually fail for the change, report
exactly what ran. Decision record: [ADR-0013](../docs/adr/0013-risk-based-verification.md).

## What to run

| Change                                       | Run                                                                                                                          |
| -------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| Docs, comments, copy                         | `bunx prettier --check <files>`                                                                                              |
| Go handler/usecase, plumbing                 | `task check:quick`; add a focused `go test` when logic, authorization, money or concurrency changed                          |
| SQL query / migration                        | `task gen:db`, `task check:db-schema`, then `task check:db:integration` (migration or constraint changes)                    |
| TypeScript logic, adapters, server functions | `task check:quick`; focused `bun test <file>` for branching/parsing/state                                                    |
| UI presentation                              | `task check:quick` + review the affected stories in Storybook (below)                                                        |
| API contract change                          | `task sync`, then `task check:quick`                                                                                         |
| Scaffold / generators                        | `task check:scaffold`                                                                                                        |
| Auth, customers, org boundary, BO security   | `task check:auth:*`, `task check:bo:customers`, `task check:bo:security`, `task check:web-api:security` for the area touched |
| Before finishing a batch                     | `task check:go`, `task check-types`                                                                                          |

`task check:quick` checks only what `git status` shows changed, plus its consumers;
it prints one line on success. Counted gates (`check:auth:*`, `check:infra`) fail on
missing or skipped cases; their supervisor is described in `packages/tools/authtest/README.md`.

## Allowed without asking

Compile, vet, typecheck; `go test` and `bun test` for affected owners; the `task check:*`
suites above including their owned Testcontainers PostgreSQL (bounded timeouts, Ryuk,
fake data); offline generators; TanStack route generation; the repository Storybook.
Run only the affected suite, one Testcontainers suite at a time.

## Never without explicit instruction

An existing or user database, real `.env` or secrets, operational or remote migrations
(`task db:*`), deployment, the Compose stack, real mail, starting backend listeners
or workers against real configuration.

## Storybook

Use `http://localhost:6006`. If absent or stale, identify this repository's old process,
stop it, confirm exit, then `task dev:storybook`. Exactly one instance, never another
port; ask if ownership is unclear and never kill unrelated processes. Stories use local
fixtures only. Typecheck cannot show appearance or interaction: look at the stories.

## Reporting

Blocked or skipped is not a pass. Name the commands run and what they cannot prove
(typecheck does not prove authorization, SQL constraints, transactions or live
integrations). Preserve unrelated changes.

## Keeping this file small

No feature- or milestone-specific exceptions here. A new kind of check adds one row to
the table above.
