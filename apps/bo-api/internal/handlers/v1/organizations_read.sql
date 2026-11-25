-- name: BOListCustomerOrganizations :many
SELECT o.id,o.name,o.slug,o.status,o.created_at,o.updated_at,
 (SELECT count(*) FROM organization_memberships m WHERE m.organization_id=o.id)::bigint AS member_count,
 (SELECT count(*) FROM organization_memberships m WHERE m.organization_id=o.id AND m.status='active')::bigint AS active_member_count,
 (SELECT count(*) FROM organization_memberships m JOIN users u ON u.id=m.user_id WHERE m.organization_id=o.id AND m.status='active' AND m.role = 'owner' AND u.status='active')::bigint AS active_owner_count
FROM organizations o
WHERE (sqlc.arg(search)::text='' OR o.name ILIKE sqlc.arg(search_pattern)::text OR o.slug ILIKE sqlc.arg(search_pattern)::text)
 AND (sqlc.arg(state)::text='' OR o.status=sqlc.arg(state))
ORDER BY o.created_at DESC,o.id DESC LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: BOGetCustomerOrganization :one
SELECT o.id,o.name,o.slug,o.status,o.created_at,o.updated_at,
 (SELECT count(*) FROM organization_memberships m WHERE m.organization_id=o.id)::bigint AS member_count,
 (SELECT count(*) FROM organization_memberships m WHERE m.organization_id=o.id AND m.status='active')::bigint AS active_member_count,
 (SELECT count(*) FROM organization_memberships m JOIN users u ON u.id=m.user_id WHERE m.organization_id=o.id AND m.status='active' AND m.role = 'owner' AND u.status='active')::bigint AS active_owner_count
FROM organizations o WHERE o.id=$1;

-- name: BOCountCustomerOrganizations :one
SELECT count(*) FROM organizations o
WHERE (sqlc.arg(search)::text='' OR o.name ILIKE sqlc.arg(search_pattern)::text OR o.slug ILIKE sqlc.arg(search_pattern)::text)
 AND (sqlc.arg(state)::text='' OR o.status=sqlc.arg(state));

-- name: BOCustomerOrganizationSummary :one
SELECT count(*)::bigint AS total,count(*) FILTER(WHERE status='active')::bigint AS active,
 count(*) FILTER(WHERE status='suspended')::bigint AS suspended FROM organizations;
