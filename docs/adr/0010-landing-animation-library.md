# ADR-0010: Landing may use an animation library

- Status: Accepted
- Date: 2026-11-18

## Context

ADR-0008 said "Add no animation library at setup" for `apps/landing`. The landing
redesign aims for a precise, motion-rich product feel (scroll reveals, staged
product mocks, hover glows). CSS alone covers part of this but makes orchestrated
and interruptible animation costly to build and maintain. The CTO approved
lifting the restriction.

## Decision

This supersedes only the animation-library sentence of ADR-0008; the rest of
ADR-0008 stays Accepted.

`apps/landing` may depend on an animation library. Prefer `motion`
(`motion/react`), which the workspace already uses in `packages/ui`, so the
repository keeps one animation library. Adding a different library needs a short
justification in the change that introduces it.

Animated presentation components stay free of `next/*` imports so colocated
stories still render in the React/Vite Storybook. Components using the library
are client components (`"use client"`). The site remains a static export.
Every animation honours `prefers-reduced-motion`; content is visible and usable
without JavaScript-driven motion.

## Alternatives considered

CSS-only motion (`animation-timeline`, transitions, masks) needs no dependency
and stays the first choice for simple effects, but orchestrated sequences,
spring physics and gesture/scroll coordination are harder to build and keep
consistent. A second library (for example GSAP) would duplicate `motion`, which
is already in the workspace.

## Consequences

- Landing gains a client-side JavaScript dependency and some bundle cost; keep
  animated islands small and static content server-rendered.
- Animated components need `"use client"`, which is a no-op in Storybook.
- Reduced-motion behaviour must be checked in Storybook for each animated component.

## Revisit when

Bundle size or Core Web Vitals regress noticeably, or the landing moves away from
a static export.
