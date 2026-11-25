# Embedded transactional email

Go-only rendering and bounded Resend delivery. `go:embed` loads the committed
templates exported from `packages/ui-email`; production needs no JS runtime,
template network request, separate email listener, queue or worker.

`NewRenderer` fixes the trusted Auth origin. `Render` accepts only the three
approved purposes, validates exact action paths and opaque query fields, rejects
unsafe/foreign destinations and incomplete inputs, and uses `html/template` plus
an independent `text/template` with `missingkey=error`. No trusted HTML/URL casts
or runtime placeholder replacement bypass contextual escaping.

`NewResend` uses pinned resend-go **v4.9.0**. Each send has at most a five-second
context/client deadline, a stable per-attempt idempotency key and one transport
attempt. Redirects cannot forward recipient/link/credential payloads. The
purpose-owned transport ignores ambient proxy selection, pins the canonical
Resend endpoint and bounds response size. The SDK's upstream `RESEND_BASE_URL`
package-init default is explicitly overwritten and never becomes destination
authority. No feature environment selector or provider URL override is exposed.

Errors distinguish not-sent, provider-rejected and acceptance-unknown using fixed
safe categories, never raw provider details. A nil send error means provider API
acceptance only, not mailbox delivery. Ephemeral inputs/messages/senders reject
generic JSON and redact debug formatting. Application commands must commit token
hashes and safe audit first, then pass an ephemeral delivery to the Web API
adapter; this module owns no database transaction and cannot roll one back.

No durable raw link, outbox, automatic retry or restart delivery guarantee exists.
The test-only capture transport lives in `packages/tools/authtest` and can be
injected only by test composition, never ordinary production configuration.
`task check:auth:unit` selects nonempty renderer/SDK/capture and actual owned-PG
post-commit-boundary cases. Account-specific send orchestration remains Tasks 9–11.
Real Resend sending, sender/domain verification and inbox placement require
separate user approval and are not claimed by fixture capture.
