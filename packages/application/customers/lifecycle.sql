-- name: CustomerUserOrganizationIDs :many
SELECT organization_id FROM organization_memberships WHERE user_id=$1 ORDER BY organization_id;

-- name: CustomerLockUserMemberships :many
SELECT id FROM organization_memberships WHERE user_id=$1 ORDER BY id FOR UPDATE;

-- name: CustomerOtherUserOwnerExists :one
SELECT EXISTS (
 SELECT 1 FROM organization_memberships m JOIN users u ON u.id=m.user_id
 WHERE m.organization_id=$1 AND m.status='active' AND m.role = 'owner'
 AND u.status='active' AND m.user_id<>$2
);

-- name: CustomerUserOwnsOrganization :one
SELECT EXISTS(SELECT 1 FROM organization_memberships WHERE organization_id=$1 AND user_id=$2 AND status='active' AND role = 'owner');

-- name: CustomerSetUserStatus :one
UPDATE users SET status=$2,updated_at=greatest(sqlc.arg(at)::timestamptz,updated_at+interval '1 microsecond') WHERE id=$1 RETURNING *;

-- name: CustomerSetOrganizationStatus :one
UPDATE organizations SET status=$2,updated_at=greatest(sqlc.arg(at)::timestamptz,updated_at+interval '1 microsecond') WHERE id=$1 RETURNING *;

-- name: CustomerEmailUnchanged :one
SELECT email_normalized=lower(btrim(sqlc.arg(email)::text)) FROM users WHERE id=$1;

-- name: CustomerChangeUserEmail :one
UPDATE users SET email=lower(btrim(sqlc.arg(email)::text)),email_normalized=lower(btrim(sqlc.arg(email)::text)),email_verified_at=NULL,
updated_at=greatest(sqlc.arg(at)::timestamptz,updated_at+interval '1 microsecond') WHERE id=$1 RETURNING *;

-- name: CustomerVerifyUserEmail :one
UPDATE users SET email_verified_at=sqlc.arg(at)::timestamptz,
updated_at=greatest(sqlc.arg(at)::timestamptz,updated_at+interval '1 microsecond') WHERE id=$1 AND email_verified_at IS NULL RETURNING *;
