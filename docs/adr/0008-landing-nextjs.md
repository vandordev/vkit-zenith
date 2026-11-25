# ADR-0008: Next.js for the static marketing landing app

- Status: Accepted (animation-library clause superseded by [ADR-0010](0010-landing-animation-library.md))
- Date: 2026-11-25

## Context

A product needs an independent marketing page at `/`, separate from authentication
and the dashboard. Product apps use TanStack Start.

## Decision

`apps/landing` uses Next.js App Router, TypeScript and React 19 matching the
repository, with `output: "export"` for a static site. It uses Tailwind v4, the shared
`@repo/brand` identity/tokens/self-hosted fonts and `@repo/ui/<name>` primitives.
Presentation components stay independent of Next.js so colocated stories work in the
repository Storybook.

## Alternatives considered

Extending `apps/web` would couple marketing delivery to authentication and runtime
configuration.

## Consequences

- A second frontend framework with its own route and metadata conventions.
- Landing has no API, auth, runtime config or environment reads.
- Shared brand, primitive ownership, Storybook and verification rules still apply.

## Revisit when

The marketing site needs dynamic server features.
