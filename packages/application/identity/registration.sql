-- Admission runs in a separate limiter-control → sorted bucket transaction.
-- Mutations hold user → credential → session/action descendants
-- → action rows, before any trigger can traverse that graph.

-- name: IdentityFindUserByEmail :one
SELECT * FROM users WHERE email_normalized=$1;

-- name: IdentityAdoptUserGeneration :exec
UPDATE users SET identity_version=1,updated_at=greatest($2,updated_at+interval '1 microsecond') WHERE id=$1 AND identity_version IS NULL;

-- name: IdentityAdoptCredentialGeneration :exec
UPDATE password_credentials SET generation=1 WHERE user_id=$1 AND generation IS NULL;

-- name: IdentityLockEmailActionByHash :one
SELECT * FROM email_action_tokens WHERE token_hash=$1 FOR UPDATE;

-- name: IdentityFindEmailActionByHash :one
SELECT * FROM email_action_tokens WHERE token_hash=$1;

-- name: IdentityFindSessionByHash :one
SELECT * FROM auth_sessions WHERE secret_hash=$1 AND user_id=$2;

-- name: IdentityLockEnrollmentReceipt :one
SELECT * FROM identity_enrollment_receipts WHERE secret_hash=$1 FOR UPDATE;

-- name: IdentityInsertUser :one
INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at,telephone,identity_version)
VALUES(gen_random_uuid(),$1,$1,$2,'active',$3,$3,$4,1)
ON CONFLICT(email_normalized) DO NOTHING
RETURNING *;

-- name: IdentityInsertCredential :exec
INSERT INTO password_credentials(user_id,password_hash,password_changed_at,generation)
VALUES($1,$2,$3,1);

-- name: IdentityInsertEnrollmentReceipt :exec
INSERT INTO identity_enrollment_receipts(secret_hash,user_id,identity_version,credential_generation,created_at,expires_at)
VALUES($1,$2,$3,$4,$5,$6);

-- name: IdentityInsertRestrictedSession :exec
INSERT INTO auth_sessions(id,user_id,secret_hash,authenticated_at,last_seen_at,expires_at,created_at,kind,identity_version,credential_generation,last_activity_at,browser_cookie_expires_at)
VALUES(gen_random_uuid(),$1,$2,$3,$3,$4,$3,'restricted',$5,$6,$3,$4);

-- name: IdentityInsertAction :exec
INSERT INTO email_action_tokens(id,user_id,purpose,target_email_normalized,token_hash,created_at,expires_at,identity_version,credential_generation,enrollment_hash)
VALUES(gen_random_uuid(),$1,$2,$3,$4,$5,$6,$7,$8,$9);

-- name: IdentityInsertAudit :exec
INSERT INTO audit_events(id,occurred_at,actor_kind,action,outcome,target_type,target_id,detail)
VALUES(gen_random_uuid(),$1,'anonymous',$2,'success',$3,$4,$5::jsonb);
