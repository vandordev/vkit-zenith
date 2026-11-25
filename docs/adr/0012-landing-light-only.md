# ADR-0012: Landing uses the light theme only

- Status: Accepted
- Date: 2026-11-18

## Context

The shared brand tokens follow the system light/dark preference. The CTO wants the
landing to present one consistent look, while product apps keep following the system.

## Decision

Product apps keep following the system preference.

`apps/landing` renders the brand light palette only: no `.dark` class, no theme script and no
dark Storybook variant. Product apps keep their own handling; BO defaults to the system
preference with an explicit light/dark/system choice.

## Alternatives considered

Following the system adapts to user settings but splits the marketing look, the
share image and screenshots across two themes.

## Consequences

- No theme script runs on the landing; it renders identically with or without JavaScript.
- Dark-only landing decoration (the hero glow) is removed.

## Revisit when

Marketing wants a dark presentation, or landing analytics show demand for one.
