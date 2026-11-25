-- name: PlatformUserByNormalizedEmail :one
SELECT id,email_normalized,email_verified_at,status FROM users WHERE email_normalized=$1;

-- name: PlatformEmailToken :one
SELECT id,token_hash,expires_at,consumed_at,revoked_at FROM email_action_tokens WHERE id=$1;

-- name: PlatformInsertAudit :exec
INSERT INTO audit_events(id,occurred_at,actor_kind,action,outcome,detail)
VALUES($1,$2,'system',$3,'success',$4);

-- name: PlatformReadAudit :one
SELECT id,occurred_at,detail FROM audit_events WHERE id=$1;

-- name: PlatformRiverMigrationCatalog :many
SELECT * FROM river_migration ORDER BY line,version;
