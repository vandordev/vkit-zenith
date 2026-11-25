# Marketing landing app

`landing` is the Indonesian, static one-page marketing site at `/`, built with
Next.js App Router, TypeScript and React 19. Its copy is a neutral placeholder that a
product replaces; the contact CTA uses `contactHref` in `landing-content.ts`. The static `/privacy` page, linked from the footer,
states that the site uses no cookies, analytics or forms; have it reviewed before relying on it legally. The page content lives in `src/app/_components` (styles in `src/app/landing.css`).

- `task dev:landing` (or root `bun --no-env-file run dev:landing`) serves
  `http://127.0.0.1:3002`, separate from Web 4100, Web API 4101, public API 4102, BO 4103,
  BO API 4104 and Storybook 6006.
- `bun --no-env-file run --cwd apps/landing check-types` works offline after
  dependency installation, without starting Next.js or generating route types.
- `bun --no-env-file run --cwd apps/landing build` exports static files to `out/`.
- `bun --no-env-file run --cwd apps/landing lint` uses the shared ESLint config.

Brand identity, self-hosted fonts and tokens come from `@repo/brand`; import
controls directly from `@repo/ui/<name>`. Tailwind v4 uses the PostCSS plugin;
Next.js transpiles both shared source packages. Presentation components and
colocated stories live in `src/app/_components`, use relative imports and stay
free of `next/*` imports so the repository's React/Vite Storybook can render them.
Route files own Next.js metadata and routing integration.

There is no API, auth, runtime config, `getPublicConfig` or environment access.
The landing uses the light theme only ([ADR-0012](../../docs/adr/0012-landing-light-only.md));
product apps keep their own theme handling.
Animation may use `motion` (`motion/react`) per
[ADR-0010](../../docs/adr/0010-landing-animation-library.md); honour
`prefers-reduced-motion`. No public origin is assumed in application metadata.
The TanStack routing/SEO instructions do not apply as-is: see
[ADR-0008](../../docs/adr/0008-landing-nextjs.md). Follow the repository verification
policy; builds and product dev servers are not agent verification gates.

## Deployment

`apps/landing/vercel.json` is a source reference: the static export needs no
environment variables or backend configuration.

## SEO

The production origin is the placeholder `https://example.com`; set the product domain in `src/lib/seo.ts`.
`src/lib/seo.ts` owns public SEO values (no environment reads), shared page
metadata and the indexing exclusion list. Each new route must call `pageMetadata`
with its own title, description and canonical path. Root metadata provides the
base URL and shared icons; page metadata includes full OG/Twitter objects to
avoid losing nested defaults through Next.js metadata inheritance.

`/sitemap.xml` is generated at every build from static App Router page files.
Route groups are flattened; private folders, parallel slots and parameterized
routes are excluded. `/privacy` and descendants are excluded through the shared
indexing policy and receive `noindex, follow` metadata. Keep them crawlable so
search engines can read that directive. Dynamic routes will need explicit URL
generation when introduced; CMS publication requires a new static deployment.
No fabricated last-modified timestamps, priorities or update frequencies are emitted.

`/robots.txt` allows crawling and points to the canonical sitemap.
Homepage JSON-LD describes the visible brand identity and website; it adds no
parent company, reviews, ratings, prices or unsupported rich-result markup.

`public/images/social-preview.png` is the 1200×630 share image, composed from
light brand tokens, the shared symbol and Plus Jakarta Sans.
Icons derive from `packages/brand/assets/logo.svg`; update copied assets together
when brand artwork changes.

After deployment, inspect the served homepage/privacy canonical and robots tags,
OG image and icons, `/robots.txt` and `/sitemap.xml`.
Submit the sitemap in Google Search Console and inspect indexing/canonical status;
account verification and field Core Web Vitals are external checks. Repository
policy permits typechecks here; focused regression tests are committed but not
executed, and static export/hosting behavior remains unverified until deployment.

Guidance used: [nextjs-seo](https://skills.sh/laguagu/claude-code-nextjs-skills/nextjs-seo)
and [Vercel next-best-practices](https://skills.sh/vercel-labs/openreview/next-best-practices),
checked against the installed Next.js documentation.
