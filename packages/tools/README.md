# Repository tools

BO runtime increment: `db-gen --owner tools|application|web-api|bo-api` chooses only
reviewed repository configs (default tools), never an arbitrary external config.
Root `task gen:db` selects installed nonempty query owners; dormant configs are
not selected and no dummy production queries are generated. Runtime outputs stay
in each owner's internal/persistence. Tool/test packages never enter runtime
dependency closures. Handler/OpenAPI orchestration now includes BO API; Turbo
schema inputs include application DTO/usecase sources. BO route/client ownership
now exists in `apps/bo`; `task gen` generates both web and BO consumer SDKs.

Run from the repository root:

```sh
task gen:handlers
# or go run ./packages/tools/cmd/handler-gen --app apps/web-api
```

`handler-gen` parses only the selected service's `internal/handlers` Go AST. It
never imports/executes the service or its constructors. The supported convention:

```go
//vkit:handler
func NewGetHealth(api huma.API) error { /* register one operation; return nil */ }
```

Exact standalone doc marker; exported, implemented, package-level, non-generic,
non-variadic constructor; injected parameters (zero allowed); sole return `error`.
Each file implements one endpoint. Versioned handlers use flat resource-first
filenames inside `internal/handlers/v1`; health is unversioned. No `routes.go`.
Paths and stable operation IDs are explicit, not derived from filenames.

Tests and generated files are excluded. Build-tagged/platform-specific handlers,
methods, detached markers, duplicates, invalid signatures, mixed packages, nested
modules and symlinks fail with file-local diagnostics. Constructors sort by
import path/name, with collision-safe numbered aliases. Only root `app.go` is
owned; it must be absent or carry the exact generated notice. Atomic, gofmt'd
writes preserve identical output. Export uses the same `handlerOptions()` with
an export-safe infrastructure graph, never a second maintained handler list.

Native Turbo Go support is experimental, pinned to 2.11.10. Installed schema
supports discovery, not the newer online guide's command-override flag. Task
owns explicit codegen orchestration. Do not add JS package wrappers to Go modules.

## Scaffolding

`cmd/scaffold` (via `task create:handler|usecase|migration`) renders embedded templates
in `internal/scaffold/templates`: one per app (`web-api`, `bo-api`, `api`) and kind
(`query`, `mutation`), plus the usecase `Service`, usecase method, SQL stub and Goose
migration. It validates names, methods, versioned paths and lowerCamel path parameters,
gofmt's Go output and creates files with `O_EXCL`, so nothing is overwritten.
`go test ./internal/scaffold -update` rewrites the golden files after an intended
template change; review the diff. `task check:scaffold` also copies the tracked tree to
a temp directory, scaffolds every app and kind, runs `handler-gen` and `go vet`, so a
template that stops compiling against the real APIs fails the gate.

## Database tools

### Offline sqlc pin

Install exactly `go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0` into a
trusted tool directory on PATH. The generator checks `sqlc version` plus Go binary
build metadata: module `github.com/sqlc-dev/sqlc` v1.29.0, module checksum
`h1:HQctoD7y/i29Bao53qXO7CZ/BV9NcvpGpsJWvz9nKWs=`, with no replacement.
Provision through the Go checksum database; never use `@latest` or an arbitrary
unverified PATH binary. sqlc is a CLI tool, not a runtime Go dependency.

`task gen:db` reads `packages/tools/sqlc.yaml`, executes offline with empty temporary
HOME and no inherited credentials, stages native pgx/v5 files, validates all
outputs, then publishes under per-output locks. Multiple targets publish
sequentially; an error reports how many already published. No live analyzer,
plugins or cloud configuration is accepted. Full-schema parser compatibility and
native round trips are tested directly against canonical migrations, including
functions/triggers/exclusion; no derived parallel DDL is used. Both minimal and
platform explicit-query outputs are tooling-fixture-owned. Run `task gen:db`
twice and compare hashes for determinism; offline generation is not SQL runtime
constraint proof. Fixtures cover nullable native pgtype values, arrays, bytea,
JSONB and bigint; generated structs are not HTTP DTOs.

`dbtool` and `cmd/db-{migrate,gen}` own explicit operational adapters and isolated
generation, not API startup hooks. See [DB guide](../db/README.md) for commands,
settings, locks, rollback authorization and output recovery. Codegen has no DSN
override or runtime-config selector. Focused live acceptance uses the pinned image
`postgres:17.11-bookworm@sha256:858bb5e98eed02eb8cafbcdc007f9a65d1b613972028944a8ec2133ccaf80d27`,
Testcontainers core/PostgreSQL v0.44.0 and Goose v3.29.0. Runtime pgx
is v5.11.0. All five Go projects use MIT licenses; upstream PostgreSQL image
distribution includes its own notices. MVS-selected transitive dependencies are
recorded in module/sum files; no `@latest` invocation.

Bob introspection and its dependencies are retired. Offline fixture generation
must be deterministic and compile against native pgx. Live PostgreSQL constraints
and query round trips are separate acceptance, never inferred from generation.
No skipped Go test counts as acceptance. Fixtures never enter runtime query owners.

The canonical platform fixture uses the single 21-table baseline. Generated
queries cover retained identity and audit round trips. Read Committed and
Repeatable Read races use observed locks for owner and staff invariants.

Only focused DB tests and temporary containers are authorized under the active
verification exception. Tests do not load real `.env`, select existing databases,
or stop unrelated processes. Cleanup failures are errors. No app image build is
authorized as verification; BO API packaging explicitly includes DB/application/config source.
Offline
handler/OpenAPI/client `task gen` remains container-free.
