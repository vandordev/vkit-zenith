# Application ownership

Native Go owner of staff, customer-administration and the approved Core
Auth identity usecases. Staff owns its separate password/token/bootstrap/session
and governance authority. Colocated SQL generates native pgx under `internal/persistence`.

`identity` owns registration, bounded customer password KDF/login/change, generic
setup/reset requests, email verification, restricted/full sessions, CSRF, admission,
and credential-revocation transactions. API in-process HTTP→owned-PG acceptance is
`task check:auth:accounts` and `task check:auth:sessions`; the pure/source/renderer
subset is `task check:auth:unit`. These do not prove live browser/deployment integration.
Password work remains bounded and outside database transactions.

Mutations revalidate actor/session under the singleton governance fence, use
explicit ordered row locks and one DB transaction including safe success audit.
Reads belong to BO API's independent SQL target. Password work is bounded and
never held inside a DB transaction; final mutation rechecks credentials/state.
Crypto and actual HTTP→PG/security evidence belongs to `task check:bo:security`;
schema persistence evidence to `task check:db:integration`. Neither proves live
HTTP integration, production proxies, replicas or deployment lifecycle.

Future consumers must update their native module and Docker dependency closures.
BO API imports this module for staff operations and its actual customer
commands; those command races are exercised against the Web API identity usecases in
the narrowly selected Auth gate. The approved Core Auth increment adds bounded
customer identity persistence to `apps/web-api`. Runtime code must
never import tools, Testcontainers or migration orchestration.

## Customer mutation foundation

`customers` owns approved user/organization profiles, atomic initial ownership,
membership lifecycle and protected lifecycle commands. Ordinary writes recheck
live staff authority; sensitive actions use staff's callback-only confirmed
transaction (same runtime pool), never reusable proof or exposed credential hashes.
Expected revisions canonicalize UTC/microseconds; real updates monotonically
advance entity `updated_at` even with a frozen clock. Audit is transaction-atomic
and stores changed-field keys, not email/name snapshots or credentials.

Lock order: staff governance before operation clock, target user before existing
organizations (UUID order), then membership rows (UUID order); organization-only
writes never acquire users afterward. Disable/email replacement then lock existing
sessions → outstanding email tokens in ID order before revoking applicable live
records. Consumed email tokens and previous revocation instants remain intact.
Reactivation/manual verification revive no authorization. Identity commands use
user → credential → session → email-action fences and recheck current eligibility.
`task check:bo:customers` uses owned isolated PostgreSQL, never a backend listener.

Membership role is `Role` with owner/admin/member constants. `Role.Can(Permission)`
uses a fixed Go map (ADR-0004): organization.read for every role, members.manage and developers.manage for owner/admin, organization.manage for owner;
unknown roles/permissions deny. BO add accepts admin/member only; every subsequent
role edit uses the existing confirmed Superadmin mutation and records the role key.

The initial-owner create is the explicit ordinary-staff exception: organization,
optional new unverified customer and first owner membership commit atomically.
Subsequent owner changes are separately protected; owner removal cannot silently
demote. An active membership's disabled owner can be explicitly demoted only when
another eligible active owner remains; promotion still requires an active user.
Suspension does not waive the last-owner invariant. Non-owner remove/restore keeps
stable membership identity and restoration grants neither ownership nor tokens.

BO API owns joined reads and exact safe activity projection. Customer audit emits
twelve explicit actions, no raw reason/email history in
read output. It records reason through the approved audit path only, transactionally
with each successful protected command. HTTP and frontend callers must not infer
authorization from a stale availability flag or substitute revisions across targets.
