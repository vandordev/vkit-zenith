# PostgreSQL foundation

Native Go runtime owner for explicit caller-owned pgxpool pools and transactions.
Operational migrations live in `migrate`; container orchestration belongs only to
`packages/tools`. Imports, pool initialization and API startup never migrate.

Goose SQL in `migrations/00001_baseline.sql` owns the 21-table schema: customer
identity, organization/membership, independent staff and audit (preserving owner
and staff fences and bounded identity state), organization API keys and River
(ADR-0002, ADR-0003). See [baseline details and rewrite rule](migrations/README.md).
BO API and Web API explicitly adopt bounded pools. Membership roles are owner/admin/member.

## Runtime consumption

Supply `Settings` explicitly to `Open(ctx, settings)`; successful return is a native
`*pgxpool.Pool`. Each caller receives an independent pool and closes
it on shutdown. Initialization sets pool limits before a bounded Ping
and closes partial pools on failure. No global default, YAML/env credential
selection, automatic FX registration or startup migration. Config maps fields at
the bootstrap adapter, not through package aliases/imports.

`InTx(ctx, pool, options, callback)` supplies `pgx.Tx`, commits only nil callbacks,
rolls back errors/panics, preserves callback error identity with `errors.Is`, and
rethrows panics. Commit failures are errors. Rollback errors never replace primary
errors. Rollback uses an independent bounded 10-second context. No implicit retries,
nesting or savepoints. Only `migrate.Open` supplies the operational `*sql.DB` Goose
adapter. Future usecases decide atomic
boundaries; never hold a transaction around SMTP/non-rollbackable network effects.
Driver open/transaction and migration boundary errors are safe categories, not
raw driver text. Application callback errors remain application-owned.

## Configuration and TLS

Read [config contract](../config/README.md). `DB_DSN` has no fallback; only explicit
DB selectors resolve it. URLs require explicit user/host/database/TLS selection.
Prefer `sslmode=verify-full` across untrusted networks. Explicit `sslmode=disable`
is supported for operator-selected connections, including same-host private Docker
networks; it sends database traffic without TLS. There is no implicit plaintext
default or TLS fallback (`allow`/`prefer` remain rejected). No service/passfile
lookup or ambient nonempty `PG*` settings. Do not print DSNs or pass them
on command lines. Lifetime/idle zero means unlimited; limits are validated before
seconds-to-duration conversion. Zero idle duration maps to the maximum positive
Go duration (disabling practical idle retirement), because pgx zero retires
immediately; zero lifetime is natively nonexpiring. Minimum pool counts stay zero;
there is no max-idle-count field. Only explicit Web API/BO runtime and operational DB selectors resolve DB secrets;
health and offline export never do.

## Explicit operations (operator references, not verification authorization)

```sh
task db:status                   # injected environment only
task db:version
task db:up -- --development       # explicitly read root .env; injected values win
task db:down-to -- --target 1 --confirm-rollback # destructive; separately authorized
task gen:db                      # pinned offline sqlc; no Docker or configured DB
task install:dbml                # optional standalone documentation CLI
task dev:db-schema               # http://127.0.0.1:6007; no database required
task check:db-schema             # offline DBML syntax/reference validation
task check:db:unit
task check:db:integration         # isolated fixtures; Docker required, never skipped
```

Goose session advisory locking serializes independent runners (default Goose lock
ID); acquisition is bounded to 30 seconds and unlock has a fresh 10-second cleanup
context. Strict out-of-order defaults remain enabled; no repair/reset command.
Missing/unreadable/invalid sources fail; truly empty sources report version zero
and no work without fake SQL. SQL/annotations are parsed by Goose on execution.
Rollback requires a nonnegative explicit target plus confirmation, never defaults
to zero; inspect deployed history/data-loss consequences before authorizing it.
The agent's tests roll back only isolated fixture tables, never user/platform data.

## Generation ownership

### Visual schema documentation

[`schema.dbml`](schema.dbml) is a committed visual snapshot of the final Goose Up
schema, not an ORM, migration input, sqlc input or live database introspection.
Install the [DBML toolkit](https://github.com/jason-cairns/dbml-toolkit) with
`task install:dbml` (`go install github.com/jason-cairns/dbml-toolkit/cmd/dbml@latest`).
This optional standalone development CLI is not a Go workspace/runtime dependency;
the initial schema was checked with v0.0.19. Ensure `$(go env GOBIN)` (or
`$(go env GOPATH)/bin` when GOBIN is empty) is on `PATH`.

Run `task dev:db-schema`, then open `http://127.0.0.1:6007`. The upstream preview
binds loopback only, watches DBML changes and refreshes the diagram. It runs in the
foreground; Ctrl-C stops it. It does not auto-open a browser, load `.env`/runtime
config, run Goose/sqlc, connect to PostgreSQL or start application services.
Optional preview flags can be passed after `--`, e.g.
`task dev:db-schema -- --detail keys`. The viewer is development-only, not a
deployed runtime or part of `task dev`, `task gen` or database operations.

Every schema-affecting migration must update this file in the same change,
including table/column names, PostgreSQL types, nullability/defaults, primary and
unique keys, indexes and ordered composite foreign keys. Review the final Up
schema across the complete chain, not the rollback definitions. PostgreSQL checks,
partial indexes, exclusions and trigger/deferred invariants are documented in
notes where DBML cannot express their semantics; SQL remains authoritative.
Trigger-only changes also require reviewing/updating the relevant notes.

`task check:db-schema` uses `dbml parse` offline and rejects syntax/reference
diagnostics. It does **not** compare DBML against migrations; alignment requires
review of both artifacts. Keep the PostgreSQL array type `"text[]"` quoted:
v0.0.19's formatter emits an unquoted array type that its own parser rejects, so
avoid `dbml fmt -w` on this schema with that version. No generated diagrams are
committed, and neither migrations nor sqlc generation are changed.

### Runtime and query generation

Pinned sqlc v1.29.0 (MIT; exact provisioning/provenance in tools README), Goose
v3.29.0 (MIT), pgx v5.11.0 (MIT), Testcontainers core/PostgreSQL v0.44.0 (MIT).
PostgreSQL image is fixed in tooling by tag + digest. Runtime compiles with
`CGO_ENABLED=0`; neither runtime compilation nor generation needs Docker/credentials.

Generation reads reviewed offline sqlc configs and canonical Goose SQL, stages
explicit pgx/v5 query code and validates every output before publication. No live
introspection, cloud/plugin analyzer or runtime-config selector. Generated fixtures
belong to tooling testdata, not a shared runtime model catalog. Full-schema parsing
and native query round trips are covered by focused tooling acceptance. Goose bookkeeping is not
queried by fixtures.

Publication uses a persistent OS file lock (`generated.lock`, ignored), a sibling
staging directory and recoverable `generated.backup` rename. Concurrent runs wait
with cancellation. Handwritten files/symlinks are refused except the preserved
top-level README. Stale owned files are removed only after successful generation.
Interrupted backups cause refusal: an operator must inspect/restore the owned
backup rather than deleting it blindly. SIGINT/SIGTERM cancel work; independent
cleanup contexts close handles/terminate and confirm removal of the owned container.
No guarantee is made for SIGKILL/power loss; keep Ryuk enabled and recover only
verified owned resources. Never prune all containers/volumes or terminate others.

This package owns schema and infrastructure, not a default query catalog; do not
precreate `packages/db/queries`. Future API read SQL belongs beside flat versioned
handlers (e.g. `handlers/v1/users_list.sql`), generating one API-owned
`internal/persistence` package. Mutation SQL belongs beside its usecase and uses
the same pgx transaction. SQL basenames/query names must be unique across versions
within an output. Generated structs are internal persistence types, not HTTP DTOs.
Extract shared SQL only for proven identical reuse. Future consumers update native
module/image dependency closure; existing APIs remain DB-independent.
