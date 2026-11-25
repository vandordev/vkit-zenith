# Customer Auth presentation

## Context and journey

Approved 2026-11-11: adapt the two-column customer login composition from
`meta-flow-hub` to Zenith. Customer authentication is an Operate surface: entering
credentials and recovering access take priority over decoration. Installed flows
and contracts remain those documented in `apps/web/README.md`; no new data,
identity methods or API operations are introduced.

## Container and navigation

Keep standalone auth pages for direct entry. Extend the
existing `AuthLayout` so login, registration, password recovery and email-link
outcomes share the composition. Customer auth URLs, CSRF handling, secret
clearing and navigation are unchanged. No overlay or dismissal contract applies.
Forms remain in document flow, including when the viewport is short or zoomed.

Desktop uses two equal columns: a white task surface with the brand at the top
and a bounded form, beside an animated grain-gradient identity panel. Below `lg`, omit the decorative
panel entirely. Use brand-owned tokens and mark geometry, existing typography,
and shared controls. The effect uses the same Paper Shaders `GrainGradient`
version (0.0.80) and corner/noise/motion settings as the reference. Resolve colors
from brand CSS tokens rather than copying a palette. A dark brand surface and
top/bottom overlays keep text readable; no font is added. Reduced motion freezes
the shader, and mobile unmounts it. Without WebGL2, the dark panel remains usable.

## Information and states

Prioritize the page title and explanation, fields and feedback, primary submit,
then recovery and registration links. Login fields have more height; footer
links wrap on narrow screens. Preserve existing labels, password visibility,
autocomplete, required fields, disabled/loading states and accessible errors.
Decorative SVGs are hidden from assistive technology; the visible brand name
identifies Zenith. The identity panel refers only to installed connected-product
account access, without metrics or service/security claims.

## Verification record

- PASS: `bun --no-env-file run --cwd apps/web check-types`, including the
  application UI primitive policy check.
- PASS: `bun --no-env-file run --cwd apps/storybook check-types`.
- PASS: Impeccable detector on the changed layout and login form: no findings.
- Storybook reviewed in headless Chromium at 1280×900 and 360×800. The embedded
  browser could inspect the DOM but screenshot capture timed out; Chromium
  supplied the visual evidence instead.
- Reviewed login default, long email, pending, invalid input and credentials;
  registration; reset password; layout long title and absent footer. No horizontal
  document overflow in the inspected cases; the identity panel is hidden at 360px.
- Confirmed keyboard focus from email to password, visible focus, password
  visibility toggling and disabled login controls during pending.
- Shader follow-up: fresh Auth/Storybook typechecks and detector pass. In
  Storybook Chromium, assertions confirmed advancing animation frames, a frozen
  frame at reduced-motion speed zero, mobile canvas removal, desktop remount,
  and usable form/password visibility with WebGL2 disabled. Desktop, mobile,
  reduced-motion and no-WebGL screenshots were visually inspected.
- Screenshots retained locally under `/tmp/opencode/auth-layout-review`.
- Live authentication, API calls and deployment were not exercised. These are
  presentation checks with deterministic stories, not integration evidence.

## Decisions

The smaller centered-card adjustment would preserve the original composition but
would not meet the approved two-column direction. Reuse the shared layout rather
than introducing separate page shells. Social login, extra copy
variants and backend changes are outside this change. Consistency across the auth
journey is built now because the existing shared layout already owns it.

Customer auth lives in Web alongside the public foundation; the signed-in
home is /dashboard. Current story groups use Web/Pages and Web/Layouts.
