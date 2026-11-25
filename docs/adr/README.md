# Architecture Decision Records

One file per durable decision. ADRs explain **why**; specs under
`docs/superpowers/specs` say **exactly how**.

## Status rules

- **Accepted:** binding. Implement as written.
- **Proposed:** direction agreed, details open. Get user approval before building.
- **Superseded by ADR-XXXX:** kept for history; follow the superseding ADR.

Never edit the Decision of an Accepted ADR to mean something new. Write a new ADR
that supersedes it.

## Index

| ADR                                          | Title                                                         | Status                               |
| -------------------------------------------- | ------------------------------------------------------------- | ------------------------------------ |
| [0001](0001-phased-delivery.md)              | Phased delivery: schema → dummy UI/API → functionality        | Accepted                             |
| [0002](0002-tenancy.md)                      | Tenancy by organization, enforced in the database             | Accepted                             |
| [0003](0003-infrastructure.md)               | Infrastructure: worker, R2 storage, envelope encryption       | Accepted                             |
| [0004](0004-backend-layering.md)             | Backend layering: usecase writes, handler-owned reads         | Accepted                             |
| [0005](0005-backoffice-operations-center.md) | Backoffice is the operations center                           | Accepted                             |
| [0006](0006-api-per-consumer.md)             | One API per consumer                                          | Accepted                             |
| [0007](0007-public-api-conventions.md)       | Public API uses snake_case JSON; internal APIs stay camelCase | Accepted                             |
| [0008](0008-landing-nextjs.md)               | Next.js for the static marketing landing app                  | Accepted (partly superseded by 0010) |
| [0009](0009-landing-seo.md)                  | Landing canonical origin and static SEO                       | Accepted                             |
| [0010](0010-landing-animation-library.md)    | Landing may use an animation library                          | Accepted                             |
| [0011](0011-landing-scroll-reveal.md)        | Landing motion exception to the brand motion rule             | Accepted                             |
| [0012](0012-landing-light-only.md)           | Landing uses the light theme only                             | Accepted                             |
| [0013](0013-risk-based-verification.md)      | Verification is risk-based                                    | Accepted                             |

## Template

```markdown
# ADR-NNNN: Title

- Status: Proposed | Accepted | Superseded by ADR-XXXX
- Date: YYYY-MM-DD

## Context

## Decision

## Alternatives considered

## Consequences

## Revisit when
```
