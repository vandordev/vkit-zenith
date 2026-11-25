# Transactional email authoring

Build-time React Email source for the three approved customer message families:
verification, first-password setup and password reset. This is not a runtime
email service and contains no Resend key, live-send button or account policy.

`src/templates.tsx` consumes `@repo/brand` identity. The export-only color adapter
reads the brand's actual light CSS tokens, converting achromatic OKLCH to sRGB
for mail-client compatibility; there is no independently maintained palette.
The text wordmark avoids an invented deployment URL for a local logo asset.

Run `task gen:email` from the repository. Its Node launcher starts Bun in an owned
empty directory with an allowlisted environment, preventing repository-root dotenv
discovery and lifecycle hooks. It produces committed HTML/plain-text Go templates
and a field/source-hash manifest in `packages/email/templates`. Placeholders use
delimiters `[[` / `]]`, remain dynamic in text and link attributes, and are rendered
by Go's contextual template engine at runtime, never manual string substitution.
Do not hand-edit generated artifacts.

`previews` contain only static `.test` links explicitly marked non-credentials.
Colocated `Email/Templates/Transactional` stories display those actual generated
previews in sandboxed, non-interactive frames. No story fetches a provider or
application API. These previews are not inbox-placement/mail-client evidence.
The user approved one alternate Zenith-only Storybook loopback port (6008) while
another repository owns 6006; this is not general alternate-port permission.

`task check:auth:unit` counts export/determinism, dependency-backport, renderer,
transport/capture and API post-commit boundary tests. Owner typechecking uses
`node node_modules/typescript/bin/tsc --noEmit -p packages/ui-email/tsconfig.json`.
Actual account commands and four-runtime customer journeys remain later gates.

Exact reviewed pins are React Email components **1.0.12**, explicit render
**2.1.1**, and existing React/React DOM **19.2.7**. Components' nested render
**2.0.6** remains separately selected; the exporter imports 2.1.1 directly.
Preserve dependency license/notices on distribution. The low-severity deepmerge
advisory has a bounded native backport described in `../../patches/README.md`;
registry audit output is not falsely relabeled clean.
