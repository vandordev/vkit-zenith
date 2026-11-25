-- Callers hold user → credential → sessions → actions. Never acquire an earlier fence later.

-- name: IdentitySetVerification :exec
UPDATE users SET email_verified_at=coalesce(email_verified_at,sqlc.arg(at)::timestamptz),
 updated_at=greatest(sqlc.arg(at)::timestamptz,updated_at+interval '1 microsecond')
WHERE id=$1 AND email_normalized=$2 AND status='active';

-- name: IdentitySetTelephone :exec
UPDATE users SET telephone=$2,identity_version=coalesce(identity_version,1),
 email_verified_at=coalesce(email_verified_at,sqlc.arg(at)::timestamptz),
 updated_at=greatest(sqlc.arg(at)::timestamptz,updated_at+interval '1 microsecond')
WHERE id=$1 AND status='active';

-- name: IdentitySetCredential :exec
INSERT INTO password_credentials(user_id,password_hash,password_changed_at,generation)
VALUES($1,$2,$3,$4)
ON CONFLICT(user_id) DO UPDATE SET password_hash=excluded.password_hash,
 password_changed_at=excluded.password_changed_at,generation=excluded.generation;

-- name: IdentityConsumeAction :execrows
UPDATE email_action_tokens SET consumed_at=$2
WHERE id=$1 AND consumed_at IS NULL AND revoked_at IS NULL AND expires_at>$2;

-- name: IdentityRevokeOtherActions :exec
UPDATE email_action_tokens SET revoked_at=greatest(created_at,$2)
WHERE user_id=$1 AND id<>$3 AND consumed_at IS NULL AND revoked_at IS NULL;

-- name: IdentityRevokeOtherPurposeActions :exec
UPDATE email_action_tokens SET revoked_at=greatest(created_at,$3)
WHERE user_id=$1 AND purpose=$2 AND id<>$4 AND consumed_at IS NULL AND revoked_at IS NULL;

-- name: IdentityRevokeCredentialActions :exec
UPDATE email_action_tokens SET revoked_at=greatest(created_at,$2)
WHERE user_id=$1 AND purpose IN ('setup_password','reset_password')
 AND consumed_at IS NULL AND revoked_at IS NULL;

-- name: IdentityRevokeAllSessions :exec
UPDATE auth_sessions SET revoked_at=greatest(created_at,$2) WHERE user_id=$1 AND revoked_at IS NULL;

-- name: IdentityRevokeRestrictedSessions :exec
UPDATE auth_sessions SET revoked_at=greatest(created_at,$2)
WHERE user_id=$1 AND kind='restricted' AND revoked_at IS NULL;

-- name: IdentityRotateVerifiedSession :execrows
UPDATE auth_sessions SET secret_hash=$2,kind='full',last_seen_at=greatest(last_seen_at,$3),
 last_activity_at=$3,expires_at=$3+interval '1176 hours',browser_cookie_expires_at=$3+interval '1176 hours'
WHERE id=$1 AND kind='restricted' AND revoked_at IS NULL AND expires_at>$3;
