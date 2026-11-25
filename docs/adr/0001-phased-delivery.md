# ADR-0001: Phased delivery: schema → dummy UI/API → functionality

- Status: Accepted
- Date: 2026-11-25

## Context

A product built on Zenith has several surfaces (web dashboard, public API, Backoffice).
The database is the core contract: every screen and endpoint is a view over it. Work
spans many agent sessions, so direction must live in the repository, not in
conversation memory.

## Decision

Every major capability is delivered in three phases:

1. **Schema.** Goose migration + `schema.dbml` + schema contract tests. Decide the data shape first.
2. **UI + endpoints with dummy responses** for every screen. Real `/v1/...` operations
   with typed OpenAPI contracts return fixtures that cover every state the schema
   allows: each status, nullable field and empty/error state. No real persistence or
   external calls in this phase. Dummy handlers say so in a `// DUMMY (ADR-0001 phase 2)`
   comment and an OpenAPI description starting with `[DUMMY]`.
3. **Functionality.** Replace fixtures with persistence and integrations behind the
   same contracts.

## Alternatives considered

- **Vertical slices** (one feature end-to-end at a time): faster first demo, but each
  slice reshapes the schema and contracts that later slices depend on.
- **Backend first, UI last:** contracts get validated by real users too late.

## Consequences

- UI/UX and API contracts are reviewed against realistic data before persistence work.
- Fixture data must follow the schema faithfully, or phase 3 breaks the UI.
- Zenith itself ships API keys and the Backoffice operations screens in phase 2.

## Revisit when

A feature is small enough that the three phases add more overhead than value.
