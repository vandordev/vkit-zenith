# ADR-0003: Infrastructure: worker, R2 storage, envelope encryption

- Status: Accepted
- Date: 2026-11-25

## Context

SaaS products need background jobs, private object storage and encryption for stored
secrets. These must exist as foundations before the first feature needs them.

## Decision

- **Jobs:** River `v0.51.2` with its pgx v5 driver in `apps/worker`, with no HTTP
  listener or Redis. River tables are folded into Goose's `00001_baseline.sql`;
  River upgrades use reviewed `migrate-get` SQL in a new Goose migration. Until a
  product registers its first job, the worker fails fast with
  `worker has no registered jobs`; no noop job is registered.
- **Storage:** one private Cloudflare R2 bucket per environment, using the S3 API and
  presigned URLs. Object keys are scoped by organization, e.g.
  `orgs/{organization_id}/logo/{uuid}.{ext}`. Credentials come from injected typed
  config; agents never read `.env`.
- **Encryption:** AES-256-GCM envelope encryption. Each record receives a random
  32-byte DEK wrapped by the active KEK from the injected keyring. Ciphertext uses
  version `0x01`, a 12-byte wrap nonce, wrapped DEK, a 12-byte data nonce, and
  ciphertext with tag. The KEK id is stored with the record; caller-provided AAD
  binds ciphertext to its record. Rotation adds an active key while retaining old
  keys for reads. KMS is deferred.
- **Configuration:** `worker`, `storage` and `encryption` are isolated selectors in
  the Go loader; JS validates their document shape only.
- **Upload limits:** `Head` reports stored size and declared content type, not
  verified file bytes. Byte-level type sniffing is the uploading feature's job.

## Alternatives considered

- **Cron inside an API:** couples request serving to background reliability.
- **Asynq/Redis:** needs an outbox and relay for transactional enqueue.
- **Encrypting records directly with the KEK:** rotation would rewrite every record.
- **Public object bucket:** mixes public and private access models.

## Consequences

Jobs can be enqueued in the same PostgreSQL transaction as business data. R2 and the
worker are code foundations; neither is verified against live resources.

## Revisit when

A product registers its first job (add the worker lifecycle), or key custody needs KMS.
