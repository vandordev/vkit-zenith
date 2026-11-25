# ADR-0002: Tenancy by organization, enforced in the database

- Status: Accepted
- Date: 2026-11-25

## Context

Every customer record belongs to one organization. A bug that shows one organization's
data to another is a serious incident, and retrofitting tenant keys after data exists
is very expensive.

## Decision

- Every tenant-owned table has `organization_id uuid NOT NULL` with an FK to
  `organizations(id) ON DELETE RESTRICT`.
- When a child row must belong to the same organization as its parent, the parent
  exposes `UNIQUE (organization_id, id)` and the child references it with a composite
  FK `(organization_id, parent_id)`, so the database rejects cross-tenant links.
- Reads take the organization from the authenticated membership or API key, never
  from unchecked client input.
- Baseline conventions continue: application-generated UUIDs, text + CHECK enums,
  named constraints, `ON DELETE RESTRICT`, finite timestamps, and an index backing
  every foreign key (enforced by the schema manifest tests).

## Alternatives considered

- **Application-only tenant checks:** one missed `WHERE` clause leaks data.
- **Schema or database per tenant:** multiplies operations for no v1 benefit.

## Consequences

- Wider keys and more indexes; queries filter by `organization_id` first.
- Moving a record between organizations is impossible by design.

## Revisit when

Sharding by tenant is required.
