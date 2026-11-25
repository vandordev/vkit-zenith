# BO staff and customer surfaces

BO uses Indonesian presentation, English URLs, shared `@repo/ui` primitives and
feature-local stories. Its API and identity are independent from customer Auth.

- Login and account retain generic failures, pending states and transient password fields.
- The shell links to Beranda, organizations and users, plus staff/activity for Superadmin.
- Staff administration retains paginated search, consequential-write confirmation,
  credential replacement and server-enforced last-Superadmin protection.
- Customer administration retains organization/user details and membership/owner controls.
- Beranda shows recent activity for Superadmin. Activity retains date, actor,
  action and outcome filters, safe recorded detail and authoritative target links.

Presentation accepts typed data/callbacks; route adapters own fetching, auth,
URL state and imperative navigation. Server authorization is authoritative.
Storybook uses local fixtures only. Verify desktop/mobile appearance, keyboard
focus, overlays and overflow; runtime integration requires separate evidence.
