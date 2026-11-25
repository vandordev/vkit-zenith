-- name: StaffFence :one
UPDATE platform_staff_auth_control SET mutation_epoch=mutation_epoch+1 WHERE id=1 RETURNING bootstrapped_at;

-- name: StaffBootstrapExists :one
SELECT EXISTS(SELECT 1 FROM platform_staff WHERE role='superadmin');

-- name: StaffReadCredential :one
SELECT c.staff_id,c.password_hash,c.changed_at FROM platform_staff_credentials c WHERE c.staff_id=$1;

-- name: StaffLockAccounts :many
SELECT id,email_normalized,display_name,role,status,updated_at FROM platform_staff WHERE id=ANY($1::uuid[]) ORDER BY id FOR UPDATE;

-- name: StaffLockCredentials :many
SELECT staff_id,password_hash,changed_at FROM platform_staff_credentials WHERE staff_id=ANY($1::uuid[]) ORDER BY staff_id FOR UPDATE;

-- name: StaffActiveSession :one
SELECT id FROM platform_staff_sessions WHERE staff_id=$1 AND token_hash=$2 AND revoked_at IS NULL AND created_at<=$3 AND expires_at>$3 FOR UPDATE;

-- name: StaffInsert :one
INSERT INTO platform_staff(id,email_normalized,display_name,role,status,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$6)
RETURNING id,email_normalized,display_name,role,status,updated_at;

-- name: StaffUpdate :one
UPDATE platform_staff SET email_normalized=$2,display_name=$3,role=$4,status=$5,updated_at=$6 WHERE id=$1
RETURNING id,email_normalized,display_name,role,status,updated_at;

-- name: StaffSetCredential :exec
INSERT INTO platform_staff_credentials(staff_id,password_hash,changed_at) VALUES($1,$2,$3)
ON CONFLICT(staff_id) DO UPDATE SET password_hash=excluded.password_hash,changed_at=excluded.changed_at;

-- name: StaffAudit :exec
INSERT INTO audit_events(id,occurred_at,actor_kind,actor_staff_id,action,outcome,target_type,target_id,detail)
VALUES($1,$2,'staff',$3,$4,'success','staff',$5,$6);

-- name: StaffRevokeSessions :exec
UPDATE platform_staff_sessions SET revoked_at=greatest(created_at,$2::timestamptz) WHERE staff_id=$1 AND revoked_at IS NULL;
