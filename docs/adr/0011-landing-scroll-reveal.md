# ADR-0011: Landing motion exception to the brand motion rule

- Status: Accepted
- Date: 2026-11-18

## Context

The brand motion rule (`packages/brand/README.md`) forbids bounce/overshoot,
scroll-triggered motion, parallax and hover-lift. That rule fits product surfaces.
The marketing landing aims for a calm, editorial feel where headings and sections
reveal slowly as they enter, and the CTO asked for slower text reveals.

## Decision

`apps/landing` may use:

- slow page-load entrance for the hero (opacity, small translate, blur);
- scroll-linked reveals for section heads, tiles, the FAQ list and the closing CTA,
  plus a one-time outline draw of the brand mark, using opacity, small translate,
  blur and stroke only;
- a timed "live" feed inside the decorative product mock.

Still forbidden on the landing: parallax, scroll-linked tilt or scale of large
surfaces, bounce/overshoot and hover-lift. Durations and easings use the brand
tokens' easing; durations may be longer than `--motion-slow`. Everything is
disabled under `prefers-reduced-motion`, and content stays visible without
JavaScript or scroll-timeline support. Product apps keep the brand rule unchanged.

## Alternatives considered

Following the brand rule strictly keeps one motion language but loses the
landing's reveal rhythm. A per-element JS reveal (motion `whileInView`) hides
content until hydration and when JavaScript is unavailable.

## Consequences

- Reveal motion is CSS (`animation-timeline: view()`); browsers without scroll
  timelines show content without reveals.
- Reviewers check landing motion against this list, not the product rule.

## Revisit when

The landing adds pages that are product-like, or brand motion guidance changes.
