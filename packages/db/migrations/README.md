# Platform baseline

`00001_baseline.sql` is the canonical 21-table schema: customer identity,
organizations and memberships, independent staff authentication/governance,
append-only audit, enrollment receipts, bounded identity admission, organization
API keys and River.

Owner write fences and deferred checks preserve an active owner for every
organization. Staff fences retain an active Superadmin after bootstrap, immutable
sessions and exactly 720 elapsed hours per staff session. Membership role is
owner/admin/member with a text CHECK; owner fences use the owner role. Customer identity
keeps nullable legacy generations, bounded email actions and admission capacity.
Only the staff and identity limiter control singletons are seeded.

No `btree_gist` extension. Audit detail accepts safe structural fields only; callers
still own action allowlists and secret exclusion. `schema.dbml` documents the Up
schema; Goose SQL remains authoritative.

Down drops all owned objects in reverse dependency order. This baseline is for a
fresh database; it is not an upgrade path for the retired migration chain. Never
run operational migrations as agent verification or against an existing database.

## API keys and River

`api_keys` belongs to one organization (ADR-0002), stores only a 32-byte token hash
and a `zk_` display prefix (`api_keys_prefix_check`; a product picks its own prefix).
River v0.51.2 adds five tables (migration, job, leader, queue and notification); its
canonical SQL is pinned in the tools module and was obtained with
`river migrate-get --all --up`. Goose owns the River enum, helper function, unlogged
leader table and partial indexes in this baseline; the final schema has no
notification trigger. Future River upgrades are new Goose migrations from that pinned version.

## When the baseline may be rewritten

Folding changes into `00001_baseline.sql` is allowed **only while no persistent
database (shared development, staging or production) must keep its data**. Once a
product's first persistent database exists, the baseline is frozen and every schema
change is a new numbered migration (`task create:migration name=<snake>`). Record
that cutoff date and commit here when it happens.
