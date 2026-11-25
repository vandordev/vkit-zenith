# BO customer operations

## Context and journey

Indonesian staff manage customer organizations and global users. Both staff roles
may browse/create/edit ordinary profiles and non-owner memberships. Only confirmed
Superadmins change lifecycle, email verification or ownership. Organizations are
the primary connected workspace; Staf remains a separate identity system.
Authority: approved spec,
generated BO operation contracts and application-owned eligibility/revisions.

Staff scan these operational records on laptops in a normally lit office. Retain
the existing light Zenith semantic tokens, typography, floating controls and compact
density. No new palette, fonts, images, redesign or shared primitive is needed.
Use real generated string counts and current joined labels, never client-side
relationship joins or imagined login/product access.

## Containers and hierarchy

- Collection pages: PageContainer/PageHeader, exact StatsCard structural counts,
  search/status controls, linked identity, scoped table overflow and fixed25 paging.
- Full organization/user detail pages support connected profile, lifecycle and
  paginated memberships. Owners continue through the authoritative membership
  ownership filter, not a supposedly complete truncated roster.
- Short create/profile tasks use the existing EditorDrawer. Atomic organization
  creation is one task with explicit existing-owner/new-owner choice, not stacked
  customer and organization dialogs. Inline creation inside a collection would
  crowd filters; a separate full create page adds unnecessary context switching.
- Sensitive actions use focused confirmations beside their target, not a reusable
  unlock. Owner removal remains explicit demotion then ordinary removal. Backend
  availability/reasons are authoritative, including disabled-owner demotion with
  an eligible replacement (BOCO-CUST-002).

## Navigation and editing

Adapters later own English directory-first URLs, validated search/return context,
identity-scoped queries, cancellation, no-store/CSRF and explicit mutation requests.
Presentation accepts generated type-only contracts, link slots and callbacks;
no auth, fetch, router, SDK runtime or duplicate transport entities.

Create mode switches clear incompatible owner inputs. New customers start active
and unverified; no password/verification checkbox or promise of login readiness.
Profile edits expose name/slug only. Retain safe drafts in component memory on
failure; explicit revision reload never silently rebases a stale write. Pending
writes block duplicate submit, dismissal and mode switching. Existing EditorDrawer
guards unsaved dismissal, traps/restores focus and uses accessible title/close.
No promise of draft persistence across reload or identity change.

Single active overlay; scrollable body keeps actions reachable at360px and zoom.
Form labels persist; errors focus the existing Notice summary. Long identity/email
wraps; wide tables scroll locally. Links are real anchors supplied by adapters.
Empty records offer create; no matches offers filter reset. Initial failure/loading
never becomes zero/empty success, and background failures retain loaded content.

## Inspected ownership

Shared direct exports: page-container, page-header, stats-card, table, Input,
Select, Autocomplete (server filtering, disabled items, selected value), Button,
DateTime, Badge, Drawer and existing ContentState/Notice/EditorDrawer compositions.
Existing ProductCollection/ProductEditor/PlanEditor and their stories supply the
local visual and dismissal patterns. No application-local primitives are added.

## Evidence

Task9 and later state/keyboard/mobile/Storybook evidence is recorded in the
execution ledger.
Required states include both owner modes, ineligible selected owner, duplicate and
revision conflicts, retained safe input, pending/dismissal guard, no records/no
matches, long labels/email and refresh failure. Until actual browser inspection,
visual acceptance is pending, not inferred from types.

Fixture stories and typechecks do not prove live frontend/API navigation or
deployed cookies/proxy. Focused owned PG tests are separate evidence. No customer
login/SSO/issuer, password provisioning, remote logout or product/billing controls.
