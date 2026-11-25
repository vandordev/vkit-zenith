# Owned identity test resources

Tooling only; runtime dependency closures must never import this package.
WithFixture creates fresh pinned Testcontainers PostgreSQL, applies canonical
migrations and owns the pool/temp directory/container. No DSN override; fake
credentials only. Bounded contexts, enabled Ryuk and confirmed cleanup are required.
Callback error/cancellation/panic still release owned resources.

Retained counted gates are contracts/config/unit/schema/accounts/sessions/adapters
under task check:auth:. Missing, zero, failed or skipped required cases fail.
Node constructs allowlisted child environments; Bun runs from an empty owned root,
Go uses explicit workspace/GOENV=off. Never run root Bun hooks/audits or read real .env.

Every counted Go run uses scripts/auth-e2e/supervise.py, the retained non-browser
Linux Python/pidfd subreaper. Acquisition intent/identities are fsynced in a private
journal outside disposable HOME, before acquiring PostgreSQL/Ryuk resources.
Raw logs remain private. Nonzero/timeout stays failed even when cleanup succeeds;
cleanup-unknown retains HOME/evidence and stops execution. Never replace it with
an unsupervised outer timeout or use an empty Docker snapshot as cleanup proof.

ReserveLoopback/StartOwned remain test-only process-ownership utilities for the
supervisor regressions. Retained receipt/cleanup controls never authorize product
listeners. Schema gates exercise real SQL fences, manifests and affected BO
regressions; account/session gates exercise actual application/in-process HTTP
with owned PostgreSQL. They do not establish browser login, deployed TLS/cookies/
CORS/proxies or real mail. See .agent/verification.md for exact authority.
