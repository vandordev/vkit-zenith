# ADR-0005: Backoffice is the operations center

- Status: Accepted
- Date: 2026-11-25

## Context

Staff are the first line of support. When something is stuck or misconfigured they
must find the cause and fix it. Without that in the Backoffice, the only path is
direct database access, which is unaudited and unsafe.

## Decision

- **Every table is reachable in the Backoffice**, except secrets. Rate-limiter and
  queue-internal tables are excluded; their effects show on related screens.
- **Secrets are never returned or rendered**: credential and token hashes, encrypted
  values. They appear only as metadata (present, last changed).
- **Staff can adjust, not only view.** Each adjustment requires a written reason and
  writes an audit event with the staff actor.
- **Every change shows its actor.** Detail screens show history from `audit_events`
  with the actor kind: staff, organization user, API key, a named system subsystem,
  or anonymous.
- Adjustments go through the same usecases as customer actions (ADR-0004); there is
  no Backoffice-only write path.
- Staff identity is independent of customer identity: no SSO between them.
- User-visible text uses "Backoffice", never "BO".

## Alternatives considered

- **Read-only Backoffice, fixes by engineers in SQL:** every incident needs an
  engineer and leaves no audit trail.
- **Show secrets to superadmins:** one compromised staff session exposes every customer.

## Consequences

- Each feature adds its Backoffice screens in phase 2 (ADR-0001).
- Usecases accept a staff actor and reason alongside customer actors; tests assert it.

## Revisit when

Audit findings call for role-gated actions, or support moves to a separate tool.
