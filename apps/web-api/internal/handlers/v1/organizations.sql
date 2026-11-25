-- name: WebMembershipsForUser :many
SELECT o.id, o.name, o.slug, o.status, (o.logo_object_key IS NOT NULL)::boolean AS has_logo, m.role
FROM organization_memberships m
JOIN organizations o ON o.id=m.organization_id
JOIN users u ON u.id=m.user_id
WHERE m.user_id=$1 AND m.status='active' AND u.status='active'
ORDER BY o.name, o.id;

-- name: WebMembershipForOrganization :one
SELECT o.id, o.name, o.slug, o.status, (o.logo_object_key IS NOT NULL)::boolean AS has_logo, m.role
FROM organization_memberships m
JOIN organizations o ON o.id=m.organization_id
JOIN users u ON u.id=m.user_id
WHERE m.user_id=$1 AND m.organization_id=$2 AND m.status='active' AND u.status='active';

-- name: WebOrganizationMembers :many
SELECT u.display_name, u.email, m.role, m.created_at AS joined_at
FROM organization_memberships m
JOIN users u ON u.id=m.user_id
WHERE m.organization_id=$1 AND m.status='active' AND u.status='active'
ORDER BY CASE m.role WHEN 'owner' THEN 0 WHEN 'admin' THEN 1 ELSE 2 END, u.display_name, u.id;
