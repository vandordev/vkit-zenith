package customers

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/packages/application/internal/persistence"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func lockUser(ctx context.Context, q *persistence.Queries, id pgtype.UUID) (User, error) {
	r, err := q.CustomerLockUser(ctx, id)
	return user(r), err
}
func lockOrganization(ctx context.Context, q *persistence.Queries, id pgtype.UUID) (Organization, error) {
	r, err := q.CustomerLockOrganization(ctx, id)
	return organization(r), err
}

// Immutable membership scope is resolved without locking; after user-first locks
// it is re-read and validated. No organization→user lock inversion is allowed.
func lockMembership(ctx context.Context, q *persistence.Queries, orgID, id pgtype.UUID) (User, Membership, error) {
	r, err := q.CustomerResolveMembership(ctx, persistence.CustomerResolveMembershipParams{OrganizationID: orgID, ID: id})
	if err != nil {
		return User{}, Membership{}, err
	}
	u, err := lockUser(ctx, q, r.UserID)
	if err != nil {
		return User{}, Membership{}, err
	}
	if _, err = lockOrganization(ctx, q, orgID); err != nil {
		return User{}, Membership{}, err
	}
	locked, err := q.CustomerLockMembership(ctx, persistence.CustomerLockMembershipParams{OrganizationID: orgID, ID: id})
	if err != nil {
		return User{}, Membership{}, err
	}
	if locked.UserID != u.ID || locked.OrganizationID != orgID {
		return User{}, Membership{}, staff.ErrNotFound
	}
	return u, membership(locked), nil
}
