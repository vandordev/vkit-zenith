# BO API runtime

Independent Echo/Huma module with staff auth/governance endpoints. Runtime selects
BO config and explicit DB settings, opens a bounded caller-owned pool on FX
start, closes it on listener failure/shutdown and never migrates. No request
bodies, credentials, DSNs or tokens are logged. Service startup was not executed.

Offline `go run ./apps/bo-api --export-openapi contracts/openapi/bo-api.json`
constructs ExportModule and the same generated handlerOptions without Start/Run,
config selection, secret resolution or pool creation. The export-only Echo refuses
all requests (503); nil business dependencies cannot serve runtime requests.

Read SQL belongs beside flat handlers/v1 operations, generated into independent
internal/persistence via reviewed `db-gen --owner bo-api`.
Application mutation SQL belongs to its own module/target, not here or tools.

## Development hot reload

### Customer operation contracts

Flat operation-owned `/v1/users`, `/v1/organizations`, organization memberships and
bounded user choices are staff-authenticated, no-store and DB-backed. Read lists,
summaries, counts and selected context use repeatable-read snapshots where combined;
joined status/ownership/verification and decimal-string counts are authoritative.
Membership selected/detail/commands bind both organization and membership IDs.
Missing/mismatched activity membership metadata remains unavailable, never guessed.

Ordinary mutations call application customer usecases. Subsequent owner, user status,
organization status, email and manual-verification commands use per-action reason,
current password and revision under current Superadmin authority. Profile DTOs
cannot carry these privileged fields. JSON/Origin/CSRF/no-replay and existing
confirmation admission/rate limits still apply. Generated OpenAPI and BO-owned
SDK/Zod contracts are regenerated offline with `task gen`, not a server probe.
Customer authentication belongs to Web API; BO owns independent staff and customer administration.
Restoring a membership revives neither ownership nor revoked
authorization. Focused acceptance below is recorder→owned PG, not deployed behavior.

From the repository root:

```bash
task --parallel dev:bo dev:bo-api
```

`dev:bo-api` runs pinned Air v1.67.4 through `go run`; no global Air installation
is needed. The first invocation requires network access to download Air and its
dependencies into the Go cache. `apps/bo-api/.air.toml` rebuilds and restarts the
API for Go/module changes in BO API, application, Go config and DB packages,
workspace metadata, and `config/config.yaml`. Frontend changes do not rebuild
the API. SQL changes still require explicit code generation; Air does not run
generators or migrations. Test files/fixtures and `.env` are not watched; restart
the task manually after changing `.env` or Air's configuration.

The binary retains `--root . --development`, including the explicit root `.env`
opt-in and existing BO/DB configuration requirements. Artifacts are placed in
ignored `.air/bo-api` and cleaned on exit. Air sends an interrupt before restart,
with a five-second grace period before forced termination, and stops the old
binary on compilation failure. It does not clear parallel frontend output.
This is an operator development launcher, not a verification command.

`building...` followed by `running...` means compilation finished, not that the
HTTP listener is ready. Air v1.61.7 could omit child-process exit messages on
Linux; v1.67.4 uses direct stdout/stderr and reports process exit reliably in the
isolated fast-exit/rebuild fixture. An API startup failure still needs its own
diagnosis. Air does not preload `.env`; the application's explicit development
adapter remains responsible for it. Restart the task after upgrading Air or
editing `.air.toml`.

The runtime emits safe error messages when its database cannot be opened or its
HTTP listener cannot bind, and `BO API ready` after the listener is established
(subject to the configured log level). FX lifecycle logging remains disabled;
raw database errors, DSNs and credentials are never printed. Compilation alone
does not verify startup or connectivity.

## First Superadmin (operator command, not verification)

Build/distribute `cmd/staff-bootstrap` through the normal internal channel. Run
`staff-bootstrap --root PATH` (or explicitly `--development` for root .env opt-in).
Requires a real terminal; password is entered twice without echo. No password
arguments, stdin pipeline, reset, force or recovery override exists. It selects
only DB config, never migrates or opens an HTTP listener, and closes its bounded
pool. A previous Superadmin, even disabled, permanently closes first bootstrap.
Deliver the chosen password through an approved protected internal channel; the
tool prints only the created UUID. Existing-account disaster recovery needs new
explicit authorization, not rerunning bootstrap.

## Source packaging and recovery

`Dockerfile.bo-api` copies BO API plus application, db and config Go module source;
test-only tools module metadata resolves local replacements but no tools/
Testcontainers package is linked by either executable. Final distroless nonroot
image declares only API/bootstrap executables and unresolved YAML, never source,
fixtures or `.env`. Source review and native compilation are not an image build.
Compose exposes BO/BO API on loopback4103/4104 for operator-owned same-site HTTPS
proxying, requires injected origins/CSRF/docs/DB secrets and never starts DB or
automigration. Do not enable insecure development cookies on production origins.

Apply migrations through separately authorized operator operations before startup.
Recovery of a lost Superadmin is a new
approved operational procedure, never bootstrap/reset/SQL guidance implied here.
No public reset, invitation acceptance, email delivery or mandatory rotation exists.

Membership roles are owner/admin/member. Add accepts admin/member only; owner
requires atomic initial-owner creation or the existing password/reason confirmed
role endpoint. All role edits, including admin/member, require confirmation.
