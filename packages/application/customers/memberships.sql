-- name: CustomerResolveMembership :one
SELECT * FROM organization_memberships WHERE organization_id=$1 AND id=$2;

-- name: CustomerLockMembership :one
SELECT * FROM organization_memberships WHERE organization_id=$1 AND id=$2 FOR UPDATE;

-- name: CustomerInsertMember :one
INSERT INTO organization_memberships(id,organization_id,user_id,status,role,created_at,updated_at)
VALUES(gen_random_uuid(),$1,$2,'active',sqlc.arg(role)::text,sqlc.arg(at)::timestamptz,sqlc.arg(at)::timestamptz) RETURNING *;

-- name: CustomerSetMembershipStatus :one
UPDATE organization_memberships SET status=$3,updated_at=greatest(sqlc.arg(at)::timestamptz,updated_at+interval '1 microsecond')
WHERE organization_id=$1 AND id=$2 AND role <> 'owner' RETURNING *;

-- name: CustomerSetMembershipRole :one
UPDATE organization_memberships SET role=$3,updated_at=greatest(sqlc.arg(at)::timestamptz,updated_at+interval '1 microsecond')
WHERE organization_id=$1 AND id=$2 RETURNING *;

-- name: CustomerOtherOwnerExists :one
SELECT EXISTS (
 SELECT 1 FROM organization_memberships m JOIN users u ON u.id=m.user_id
 WHERE m.organization_id=$1 AND m.status='active' AND m.role = 'owner'
 AND u.status='active' AND m.id<>$2
);
