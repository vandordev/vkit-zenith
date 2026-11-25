-- Lock order: user (caller) → sessions → email tokens.
-- Future issuers must participate in this user-first protocol. These queries do
-- not install an issuer or fabricate completion/consumption of requests/codes.

-- name: CustomerLockSessions :many
SELECT id FROM auth_sessions WHERE user_id=$1 ORDER BY id FOR UPDATE;

-- name: CustomerLockEmailTokens :many
SELECT id FROM email_action_tokens WHERE user_id=$1 AND consumed_at IS NULL ORDER BY id FOR UPDATE;

-- name: CustomerRevokeSessions :exec
UPDATE auth_sessions SET revoked_at=greatest(created_at,sqlc.arg(at)::timestamptz) WHERE user_id=$1 AND revoked_at IS NULL;

-- name: CustomerRevokeEmailTokens :exec
UPDATE email_action_tokens SET revoked_at=greatest(created_at,sqlc.arg(at)::timestamptz) WHERE user_id=$1 AND consumed_at IS NULL AND revoked_at IS NULL;
