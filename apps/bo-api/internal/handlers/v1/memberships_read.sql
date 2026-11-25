-- name: BOListCustomerMemberships :many
SELECT m.id,m.organization_id,m.user_id,m.status,m.role,m.created_at,m.updated_at,
 u.display_name AS user_name,u.email AS user_email,u.status AS user_status,
 o.name AS organization_name,o.slug AS organization_slug,o.status AS organization_status,
 EXISTS(SELECT 1 FROM organization_memberships r JOIN users ru ON ru.id=r.user_id WHERE r.organization_id=m.organization_id AND r.id<>m.id AND r.status='active' AND r.role = 'owner' AND ru.status='active') AS has_other_owner
FROM organization_memberships m JOIN users u ON u.id=m.user_id JOIN organizations o ON o.id=m.organization_id
WHERE (sqlc.narg(organization_id)::uuid IS NULL OR m.organization_id=sqlc.narg(organization_id))
 AND (sqlc.narg(user_id)::uuid IS NULL OR m.user_id=sqlc.narg(user_id))
 AND (sqlc.arg(search)::text='' OR u.display_name ILIKE sqlc.arg(search_pattern)::text OR u.email ILIKE sqlc.arg(search_pattern)::text OR o.name ILIKE sqlc.arg(search_pattern)::text)
 AND (sqlc.arg(state)::text='' OR m.status=sqlc.arg(state))
 AND (sqlc.arg(ownership)::text='' OR (m.role='owner')=(sqlc.arg(ownership)='owner'))
ORDER BY m.created_at DESC,m.id DESC LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: BOGetCustomerMembership :one
SELECT m.id,m.organization_id,m.user_id,m.status,m.role,m.created_at,m.updated_at,
 u.display_name AS user_name,u.email AS user_email,u.status AS user_status,
 o.name AS organization_name,o.slug AS organization_slug,o.status AS organization_status,
 EXISTS(SELECT 1 FROM organization_memberships r JOIN users ru ON ru.id=r.user_id WHERE r.organization_id=m.organization_id AND r.id<>m.id AND r.status='active' AND r.role = 'owner' AND ru.status='active') AS has_other_owner
FROM organization_memberships m JOIN users u ON u.id=m.user_id JOIN organizations o ON o.id=m.organization_id
WHERE m.id=sqlc.arg(id) AND m.organization_id=sqlc.arg(organization_id);

-- name: BOCountCustomerMemberships :one
SELECT count(*) FROM organization_memberships m JOIN users u ON u.id=m.user_id JOIN organizations o ON o.id=m.organization_id
WHERE (sqlc.narg(organization_id)::uuid IS NULL OR m.organization_id=sqlc.narg(organization_id))
 AND (sqlc.narg(user_id)::uuid IS NULL OR m.user_id=sqlc.narg(user_id))
 AND (sqlc.arg(search)::text='' OR u.display_name ILIKE sqlc.arg(search_pattern)::text OR u.email ILIKE sqlc.arg(search_pattern)::text OR o.name ILIKE sqlc.arg(search_pattern)::text)
 AND (sqlc.arg(state)::text='' OR m.status=sqlc.arg(state))
 AND (sqlc.arg(ownership)::text='' OR (m.role='owner')=(sqlc.arg(ownership)='owner'));

-- name: BOCustomerMembershipSummary :one
SELECT count(*)::bigint AS total,count(*) FILTER(WHERE status='active')::bigint AS active,
 count(*) FILTER(WHERE status='removed')::bigint AS removed,
 count(*) FILTER(WHERE status='active' AND role = 'owner')::bigint AS owners
FROM organization_memberships
WHERE (sqlc.narg(organization_id)::uuid IS NULL OR organization_id=sqlc.narg(organization_id))
 AND (sqlc.narg(user_id)::uuid IS NULL OR user_id=sqlc.narg(user_id));
