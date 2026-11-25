# Security

Do not commit `.env`, credentials, API keys, tokens or private certificates.
Use `.env.example` as an operator template. Report suspected vulnerabilities
privately to the repository owner with impact and reproduction details.

Deployment injects environment into unresolved `config/config.yaml`; images never
copy `.env`. Docs Basic Auth is service-local; enabled protection requires both
credentials and HTTPS. Docs protection does not authorize business operations.

Web public configuration is exactly `{ apiBaseUrl }` through no-store
`getPublicConfig`; BO uses `{ boApiBaseUrl }`. Browser imports only
`@repo/config/public`. No Vite/public env, build injection or full server config
serialization. Logs/errors omit passwords, tokens, request URLs/headers/bodies
and submitted secret values. One-time email tokens are scrubbed before router state.

Customer sessions use Web API cookies and memory-only CSRF. Web API allows the exact
configured Web origin; production requires same-site HTTPS. BO has independent
staff identity. All role edits require existing Superadmin password/reason
confirmation; ordinary membership creation accepts admin/member only. Owner
fences and transaction-atomic audit protect the last eligible owner.

Use focused counted fake-data gates and owned isolated PostgreSQL only as
per `.agent/verification.md`. Live browser login, deployed TLS/cookies/proxies
and real email are separate unverified boundaries.
