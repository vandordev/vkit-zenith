# Offline API contracts

Three committed schemas: customer authentication and dashboard (`web-api.json`),
the public organization API-key API (`api.json`) and staff/customer administration
(`bo-api.json`). The public API serves its reference at `/docs` and `/openapi.json`;
its operational `/health` probe is excluded from the reference.
`task gen:openapi` orders AST handler generation before offline FX export; export
never starts listeners, resolves runtime config/secrets or opens a database.

Web generates its credentialed identity/dashboard SDK from `web-api.json`; BO
generates from `bo-api.json`. DTOs use Huma validation and RFC 9457 errors. Internal
APIs use camelCase, the public API snake_case (ADR-0007). Membership reads expose
role owner/admin/member; add accepts admin/member; the confirmed role endpoint
accepts all three. Lists use bounded `{ items, pagination }` with empty arrays.

Use `task gen`, commit generated outputs and confirm reproducibility. Never
hand-edit SDKs, wiring or schemas. Export/typechecks do not prove live HTTP,
cookies/CORS, deployment or mail.
