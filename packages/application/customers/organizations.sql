-- name: CustomerLockOrganization :one
SELECT * FROM organizations WHERE id=$1 FOR UPDATE;

-- name: CustomerInsertOrganization :one
INSERT INTO organizations(id,name,slug,status,created_at,updated_at)
VALUES(gen_random_uuid(),$1,$2,'active',sqlc.arg(at)::timestamptz,sqlc.arg(at)::timestamptz) RETURNING *;

-- name: CustomerInsertInitialOwner :one
INSERT INTO organization_memberships(id,organization_id,user_id,status,role,created_at,updated_at)
VALUES(gen_random_uuid(),$1,$2,'active','owner',sqlc.arg(at)::timestamptz,sqlc.arg(at)::timestamptz) RETURNING *;

-- name: CustomerUpdateOrganizationProfile :one
UPDATE organizations SET name=$2,slug=$3,updated_at=greatest(sqlc.arg(at)::timestamptz,updated_at+interval '1 microsecond') WHERE id=$1 RETURNING *;
