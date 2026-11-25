-- Customer session queries run only after the application has selected and
-- locked the user/credential/session fences in canonical order.

-- name: IdentityFindUserForSession :one
SELECT * FROM users WHERE email_normalized=$1;

-- name: IdentityFindPasswordCredential :one
SELECT * FROM password_credentials WHERE user_id=$1;

-- name: IdentityFindSessionOwnerByHash :one
SELECT id,user_id FROM auth_sessions WHERE secret_hash=$1;

-- name: IdentityRotateLoginSession :execrows
UPDATE auth_sessions SET secret_hash=$2,authenticated_at=$3,last_seen_at=$3,
 expires_at=$4,kind=$5,identity_version=$6,credential_generation=$7,
 last_activity_at=$3,browser_cookie_expires_at=$4
WHERE id=$1 AND revoked_at IS NULL AND secret_hash=$8;

-- name: IdentityInsertLoginSession :exec
INSERT INTO auth_sessions(id,user_id,secret_hash,authenticated_at,last_seen_at,expires_at,created_at,kind,identity_version,credential_generation,last_activity_at,browser_cookie_expires_at)
VALUES(gen_random_uuid(),$1,$2,$3,$3,$4,$3,$5,$6,$7,$3,$4);

-- name: IdentityRevokeSession :execrows
UPDATE auth_sessions SET revoked_at=greatest(created_at,$2)
WHERE id=$1 AND revoked_at IS NULL;

-- name: IdentityTouchFullSession :execrows
UPDATE auth_sessions SET last_seen_at=greatest(last_seen_at,$2),last_activity_at=$2,
 expires_at=$2+interval '1176 hours',browser_cookie_expires_at=$2+interval '1176 hours'
WHERE id=$1 AND kind='full' AND revoked_at IS NULL AND expires_at>$2
 AND last_activity_at<=$2-interval '1 minute';
