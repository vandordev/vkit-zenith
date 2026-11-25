# Brand source of truth

`@repo/brand` owns the Zenith name, the Vandor logo mark (two hexagons), neutral
light/dark color tokens, fonts, motion tokens and the shared radius scale. The palette
is neutral: primary is near-black in light mode and near-white in dark mode, and color
is reserved for status. A product replaces the name, logo and tokens here.

## Public entrypoints

```ts
import { brand } from "@repo/brand";
import { motion } from "@repo/brand/motion";
```

The root export is static, browser/server-safe TypeScript (no React, env or asset loader).

- `@repo/brand/logo.svg`: source (near-black `#1b1c1f` square, white `#ffffff` mark, viewBox 643).
- `@repo/brand/logo.png`: 512×512 export of `logo.svg` (`rsvg-convert -w 512 -h 512`),
  for non-themed consumers such as email. Regenerate from the SVG; never edit by hand.
- `@repo/brand/logo-mark`: `zenithMark` (`viewBox` + two mark paths, no square) for
  stroke animation and transparent interface logos. Themed rendering:
  `@repo/ui/zenith-mark` (`ZenithMark`) has no background; its paths use
  `--primary` (ink in light mode, cream in dark mode). The boxed SVG/PNG assets
  remain for application icons, favicons and email.
  Animation lives in `@repo/ui/zenith-loader`.
- `@repo/brand/motion`: `motion` (seconds, for Motion JS), mirroring the CSS tokens.
- `@repo/brand/tokens.css` (imports `fonts.css`) and `@repo/brand/fonts.css`.
- Favicons (`favicon.svg`, 32px/180px PNG, web `.ico`) are generated from `logo.svg` into
  `apps/bo/public` and `apps/web/public`.

```css
@import "tailwindcss";
@import "@repo/brand/tokens.css";
@import "@repo/ui/styles.css";
```

Color values live only in `src/tokens.css`. Status roles (`--success`, `--warning`,
`--info`, `--destructive`, `--test`) hold the _text_ color; badge `-light` variants derive
surfaces at /10. Do not duplicate the palette elsewhere or add a second TypeScript catalog.
The email exporter parses the first `:root {}` block: keep it plain hex with no nested
braces; fonts, motion and `--test*` live in later blocks.

Data visualization uses `--viz-1`..`--viz-5` (blue, orange, aqua, yellow, magenta), a categorical
palette in a fixed order with light and dark steps, validated with the dataviz palette script
(adjacent CVD ΔE >= 8, normal-vision ΔE >= 19). Keep the order, never cycle past five (fold
into a neutral "Lainnya"), and always label marks; aqua, yellow and magenta sit under 3:1 on the
light surface, so a legend or table is required. `--chart-1..5` remain the neutral ramp.

### Contrast (WCAG AA)

Checked light / dark (2026-11-18): body text 16.5 / 18.7 on the page; muted text ≥ 5.3
on page, card and secondary; primary button 17.0 / 18.7; focus ring ≥ 5.3 / 6.3; `--input`
(edges of fields and outline/secondary buttons, so interactive controls stay visible)
≥ 3.06 / 3.29 on page, card and secondary; every status text on its /10 surface ≥ 5.0.
`--border` is a decorative hairline for structure, never the only cue for an interactive
control. Re-check these when any palette value changes.

### Surfaces

Near-white page (`--background`) with white cards (`--card`); in dark mode a near-black page with slightly lighter cards: structure is
carried by visible hairlines (`--border`), not by card colour, so a card always keeps its
border. Wells, sidebar and hover use the `--muted`/`--secondary` step. `--input` is the
field-edge token and holds ≥ 3:1 against card, background and secondary (see Contrast);
it also draws checkbox borders and the off switch track. Fields are recessed (a faint
`--input` tint and a soft inner shadow, `fieldSurfaceClassName` in `@repo/ui/input`);
clickable controls are solid or carry the single `shadow-xs` drop. No gradients on controls.
Dark mode elevates by lightness (card lighter than background), never by shadow.

### Typography

Self-hosted `@fontsource` packages (no Google Fonts at runtime): Plus Jakarta Sans
(variable) for all text, headings and the wordmark; JetBrains Mono (variable) only for IDs,
codes and API samples. There is no serif or display family: headings use the sans
family, semibold, with tight tracking. `task check:ui-primitives` rejects `font-display`,
`font-serif` and serif `font-[...]` utilities.
Tabular figures: Plus Jakarta Sans ships `tnum` (verified in the browser: with `tabular-nums`, "1111111" and "8888888" render at equal width, 84px at 20px, vs 56px proportional), so `Money` and table numbers use the sans family with `tabular-nums`; no mono fallback is needed.

### Motion

CSS: `--motion-fast` 120ms, `--motion-base` 200ms, `--motion-slow` 320ms (all 0ms under
`prefers-reduced-motion: reduce`), `--ease-out` `cubic-bezier(.22,1,.36,1)`, `--ease-in-out`
`cubic-bezier(.65,0,.35,1)`. No bounce/overshoot, scroll-triggered, parallax or hover-lift.
Exception: the marketing landing (`apps/landing`) may use slow entrance and scroll-linked
reveals per [ADR-0011](../../docs/adr/0011-landing-scroll-reveal.md); never parallax or hover-lift.

### Shared radius

`src/tokens.css` owns `--radius: 0.8rem` and the Tailwind inline radius mappings:
`sm` subtracts 4px, `md` subtracts 2px, `lg` uses the base, and `xl`, `2xl`,
`3xl`, `4xl` add 4px, 8px, 12px, 16px respectively. Web, BO and Storybook
import this same scale without application-local radius overrides. Dark mode uses
the same geometry. Components retain their size hierarchy, circular/pill shapes,
explicit small geometry (such as checkbox and tooltip arrows), and square joined
edges. `rounded-xs` remains Tailwind's default, not part of this shared scale.
Use the existing `rounded-*` utilities rather than repeating numeric radii.

Web's `appConfig` derives `appName` and `defaultTitle` from `brand.name`; it still
owns product description, favicon configuration, and repository URL. Its favicon path stays `/favicon.ico`.
SEO helpers and future OG work consume brand identity rather than redefining it.

## Preview and verification

Colocated stories appear under `Brand/Identity`, covering logo/name, color roles (light/dark,
status and test), typography, motion tokens and the shared radius scale. The package's React/Storybook/Vite
dependencies support these previews; production identity itself has no runtime dependency.

Follow [verification policy](../../.agent/verification.md): UI uses Storybook plus
affected native typechecks. Server/config wiring uses typecheck only. Do not run
builds, system tests, Compose, or backend services for verification. Starting or
restarting repository-owned Storybook on port 6006 is authorized without repeated permission,
subject to ownership/process safeguards in the verification policy.

```bash
bun --no-env-file run --cwd packages/brand check-types
bun --no-env-file run --cwd apps/web check-types
```
