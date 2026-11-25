-- name: CustomerLockUser :one
SELECT * FROM users WHERE id=$1 FOR UPDATE;

-- name: CustomerInsertUser :one
INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at)
VALUES(gen_random_uuid(),lower(btrim(sqlc.arg(email)::text)),lower(btrim(sqlc.arg(email)::text)),sqlc.arg(display_name),'active',sqlc.arg(at)::timestamptz,sqlc.arg(at)::timestamptz)
RETURNING *;

-- name: CustomerUpdateUserProfile :one
UPDATE users SET display_name=$2,updated_at=greatest(sqlc.arg(at)::timestamptz,updated_at+interval '1 microsecond') WHERE id=$1 RETURNING *;

-- name: CustomerAudit :exec
INSERT INTO audit_events(id,occurred_at,actor_kind,actor_staff_id,organization_id,action,outcome,target_type,target_id,detail)
VALUES(gen_random_uuid(),$1,'staff',$2,$3,$4,'success',$5,$6,$7);
