-- name: BOListCustomerUsers :many
SELECT u.id,u.email,u.display_name,u.status,u.email_verified_at,u.created_at,u.updated_at,
 (SELECT count(*) FROM organization_memberships m WHERE m.user_id=u.id)::bigint AS membership_count,
 (SELECT count(*) FROM organization_memberships m WHERE m.user_id=u.id AND m.status='active')::bigint AS active_membership_count,
 (SELECT count(*) FROM organization_memberships m WHERE m.user_id=u.id AND m.status='active' AND m.role = 'owner')::bigint AS owner_membership_count,
 NOT EXISTS(SELECT 1 FROM organization_memberships m WHERE m.user_id=u.id AND m.status='active' AND m.role = 'owner' AND NOT EXISTS(SELECT 1 FROM organization_memberships r JOIN users ru ON ru.id=r.user_id WHERE r.organization_id=m.organization_id AND r.id<>m.id AND r.status='active' AND r.role = 'owner' AND ru.status='active')) AS can_disable
FROM users u
WHERE (sqlc.arg(search)::text='' OR u.display_name ILIKE sqlc.arg(search_pattern)::text OR u.email ILIKE sqlc.arg(search_pattern)::text)
 AND (sqlc.arg(state)::text='' OR u.status=sqlc.arg(state))
 AND (sqlc.arg(verification)::text='' OR (u.email_verified_at IS NOT NULL)=(sqlc.arg(verification)='verified'))
ORDER BY u.created_at DESC,u.id DESC LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: BOGetCustomerUser :one
SELECT u.id,u.email,u.display_name,u.status,u.email_verified_at,u.created_at,u.updated_at,
 (SELECT count(*) FROM organization_memberships m WHERE m.user_id=u.id)::bigint AS membership_count,
 (SELECT count(*) FROM organization_memberships m WHERE m.user_id=u.id AND m.status='active')::bigint AS active_membership_count,
 (SELECT count(*) FROM organization_memberships m WHERE m.user_id=u.id AND m.status='active' AND m.role = 'owner')::bigint AS owner_membership_count,
 NOT EXISTS(SELECT 1 FROM organization_memberships m WHERE m.user_id=u.id AND m.status='active' AND m.role = 'owner' AND NOT EXISTS(SELECT 1 FROM organization_memberships r JOIN users ru ON ru.id=r.user_id WHERE r.organization_id=m.organization_id AND r.id<>m.id AND r.status='active' AND r.role = 'owner' AND ru.status='active')) AS can_disable
FROM users u WHERE u.id=$1;

-- name: BOCountCustomerUsers :one
SELECT count(*) FROM users u
WHERE (sqlc.arg(search)::text='' OR u.display_name ILIKE sqlc.arg(search_pattern)::text OR u.email ILIKE sqlc.arg(search_pattern)::text)
 AND (sqlc.arg(state)::text='' OR u.status=sqlc.arg(state))
 AND (sqlc.arg(verification)::text='' OR (u.email_verified_at IS NOT NULL)=(sqlc.arg(verification)='verified'));

-- name: BOCustomerUserSummary :one
SELECT count(*)::bigint AS total,count(*) FILTER(WHERE status='active')::bigint AS active,
 count(*) FILTER(WHERE status='disabled')::bigint AS disabled,
 count(*) FILTER(WHERE email_verified_at IS NOT NULL)::bigint AS verified,
 count(*) FILTER(WHERE email_verified_at IS NULL)::bigint AS unverified FROM users;

-- name: BOListCustomerUserChoices :many
SELECT u.id,u.email,u.display_name,u.status,m.status AS membership_status
FROM users u LEFT JOIN organization_memberships m ON m.user_id=u.id AND m.organization_id=sqlc.narg(organization_id)::uuid
WHERE (sqlc.arg(search)::text='' OR u.display_name ILIKE sqlc.arg(search_pattern)::text OR u.email ILIKE sqlc.arg(search_pattern)::text)
ORDER BY u.created_at DESC,u.id DESC LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: BOGetCustomerUserChoice :one
SELECT u.id,u.email,u.display_name,u.status,m.status AS membership_status
FROM users u LEFT JOIN organization_memberships m ON m.user_id=u.id AND m.organization_id=sqlc.narg(organization_id)::uuid
WHERE u.id=sqlc.arg(id);
