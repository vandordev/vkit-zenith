-- Caller resolves immutable scope without locks, then rereads under this order:
-- sorted users → credentials → sessions → email actions. BO alone has its staff/owner prefix.
-- A successful lock is NOT authorization: callers recheck all live bindings.

-- name: IdentityLockUser :one
SELECT * FROM users WHERE id=$1 FOR UPDATE;

-- name: IdentityLockCredential :one
SELECT * FROM password_credentials WHERE user_id=$1 FOR UPDATE;

-- name: IdentityLockSession :one
SELECT * FROM auth_sessions WHERE id=$1 FOR UPDATE;

-- name: IdentityLockSessions :many
SELECT id FROM auth_sessions WHERE user_id=$1 ORDER BY id FOR UPDATE;

-- name: IdentityLockEmailActions :many
SELECT * FROM email_action_tokens WHERE user_id=$1 AND consumed_at IS NULL ORDER BY id FOR UPDATE;

-- name: IdentityCountLiveActions :one
SELECT count(*) FROM email_action_tokens WHERE user_id=$1 AND purpose=$2
    AND consumed_at IS NULL AND revoked_at IS NULL AND expires_at>$3;

-- Admission is a SEPARATE bounded transaction, before entering the user chain.
-- Its singleton must be acquired before any sorted bucket key. Never evict live
-- buckets at capacity. Bucket cleanup occurs in chunks <=100 under this fence.
-- name: IdentityLockLimiterControl :one
SELECT bucket_count FROM identity_limiter_control WHERE id=1 FOR UPDATE;
