-- name: BOStaffList :many
SELECT id,email_normalized,display_name,role,status,updated_at FROM platform_staff
WHERE ($1::text='' OR email_normalized ILIKE '%'||$1||'%' OR display_name ILIKE '%'||$1||'%')
AND ($2::text='' OR status=$2) ORDER BY created_at,id LIMIT $3 OFFSET $4;

-- name: BOStaffCount :one
SELECT count(*) FROM platform_staff WHERE ($1::text='' OR email_normalized ILIKE '%'||$1||'%' OR display_name ILIKE '%'||$1||'%') AND ($2::text='' OR status=$2);

-- name: BOStaffGet :one
SELECT id,email_normalized,display_name,role,status,updated_at FROM platform_staff WHERE id=$1;

-- name: BOStaffSummary :one
SELECT count(*) AS total, count(*) FILTER (WHERE status='active') AS active, count(*) FILTER (WHERE status='disabled') AS disabled, count(*) FILTER (WHERE role='superadmin') AS superadmin, count(*) FILTER (WHERE role='admin') AS admin FROM platform_staff;
