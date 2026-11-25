# ADR-0009: Landing canonical origin and static SEO

- Status: Accepted
- Date: 2026-11-25

## Decision

Keep ADR-0008's static export and no runtime configuration. Landing owns its public
SEO constants (`src/lib/seo.ts`, placeholder `https://example.com` until a product sets
its domain) and typed Metadata API composition. Each page supplies its canonical,
title, description and social metadata. A build-time sitemap lists static page routes,
excluding the shared noindex list; noindex pages stay crawlable. Do not add fake
dates, ratings, offers or company details to structured data.

## Consequences

New static pages enter the sitemap automatically and need their own metadata review.
Search Console, real crawler access and redirects need deployment evidence.

## Revisit when

The production domain changes or content becomes dynamic or localized.
