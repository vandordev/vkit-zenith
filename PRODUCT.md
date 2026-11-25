# Product brief

Template: replace every section with the product built on Zenith. Agents read this
before changing product scope (see `AGENTS.md`).

## What it is

One paragraph: the problem, who has it, and what the product does about it.

## Who it is for

- **Primary users:** who signs up and why.
- **Not for (v1):** who the product deliberately does not serve yet.

## Surfaces

- **Web dashboard** (`apps/web`): what an organization's members do here.
- **Public API** (`apps/api`): what customers' servers automate.
- **Backoffice** (`apps/bo`): what staff support and adjust.
- **Landing** (`apps/landing`): the story the marketing page tells.

## Glossary

| Term | Meaning |
| --- | --- |
| Organization | The tenant. Every customer record belongs to exactly one (ADR-0002). |
| Member | A user's membership in an organization, with role owner, admin or member. |
| Staff | Backoffice operator; a separate identity from customer users. |
| API key | Organization-scoped credential for the public API, prefixed `zk_`. |

## Status

Foundation only: identity, organizations, memberships, Backoffice staff
administration, audit, API keys (dummy phase) and infrastructure foundations.
