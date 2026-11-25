# ADR-0013: Verification is risk-based

- Status: Accepted
- Date: 2026-11-25

## Context

Zenith inherited a verification policy from valcentra that forbade running tests,
containers and builds, then grew six dated exceptions (Core Auth, BO security, BO
customers, organization dashboard, database foundation, M2 infrastructure). Every
milestone appended another one, the file reached 14 KB and was read at the start of
each session, and agents stopped to ask permission for checks that the exceptions
already allowed.

## Decision

Verify by risk: the cheapest check that can actually fail for the change. Compile,
vet, typecheck, focused tests and the existing `task check:*` suites (including their
owned, bounded Testcontainers PostgreSQL) run without asking. UI is reviewed in the
repository Storybook. Safety boundaries stay: no existing or user database, no real
`.env` or secrets, no operational or remote migrations, no deployment, no Compose
stack, no real mail, no backend listeners against real configuration.

`.agent/verification.md` is a table of change type to command. It takes no feature- or
milestone-specific exceptions; a new kind of check adds a row.

## Alternatives considered

Keep the strict compile-only policy and consolidate the exceptions into one list.
Smaller than before, but agents still could not run the tests that prove behavior.

## Consequences

The six exceptions are removed; git history keeps their text. Counted-gate mechanics
live with their scripts (`packages/tools/authtest/README.md`). Typecheck is still not
evidence for authorization, SQL constraints or live integrations, and reports must say so.
