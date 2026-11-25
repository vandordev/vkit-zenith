-- name: StaffLoginCredential :one
SELECT s.id,s.email_normalized,s.display_name,s.role,s.status,s.updated_at,c.password_hash,c.changed_at
FROM platform_staff s JOIN platform_staff_credentials c ON c.staff_id=s.id WHERE s.email_normalized=$1;

-- name: StaffAuthenticate :one
SELECT s.id,s.email_normalized,s.display_name,s.role,s.status,s.updated_at
FROM platform_staff_sessions a JOIN platform_staff s ON s.id=a.staff_id
WHERE a.token_hash=$1 AND a.revoked_at IS NULL AND a.created_at<=$2 AND a.expires_at>$2 AND s.status='active';

-- name: StaffIssueSession :exec
INSERT INTO platform_staff_sessions(id,staff_id,token_hash,created_at,expires_at) VALUES($1,$2,$3,$4,$5);

-- name: StaffLogout :exec
UPDATE platform_staff_sessions SET revoked_at=greatest(created_at,$3::timestamptz) WHERE staff_id=$1 AND token_hash=$2 AND revoked_at IS NULL;

-- name: StaffLimiterLock :exec
SELECT pg_advisory_xact_lock(684731920156::bigint);

-- name: StaffLimiterCleanup :exec
DELETE FROM platform_staff_auth_buckets WHERE ctid IN
 (SELECT b.ctid FROM platform_staff_auth_buckets b WHERE b.expires_at<=$1 ORDER BY b.expires_at,b.purpose,b.dimension,b.key_hash LIMIT 256);

-- name: StaffLimiterCount :one
SELECT count(*) FROM platform_staff_auth_buckets;

-- name: StaffLimiterBucket :one
SELECT attempts FROM platform_staff_auth_buckets WHERE purpose=$1 AND dimension=$2 AND key_hash=$3 AND window_start=$4 FOR UPDATE;

-- name: StaffLimiterIncrement :exec
INSERT INTO platform_staff_auth_buckets(purpose,dimension,key_hash,window_start,attempts,expires_at) VALUES($1,$2,$3,$4,1,$5)
ON CONFLICT(purpose,dimension,key_hash,window_start) DO UPDATE SET attempts=least(platform_staff_auth_buckets.attempts+1,101);

-- name: StaffLimiterFailure :one
UPDATE platform_staff_auth_buckets SET failure_recorded=true WHERE purpose=$1 AND dimension=$2 AND key_hash=$3 AND window_start=$4 AND NOT failure_recorded RETURNING attempts;

-- name: StaffFailureAudit :exec
INSERT INTO audit_events(id,occurred_at,actor_kind,action,outcome,target_type,target_id,detail)
VALUES($1,$2,'anonymous',$3,'failure','auth_bucket',$4,'{}');
