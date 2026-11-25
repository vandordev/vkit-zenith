package customers

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/packages/application/internal/persistence"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
	"time"
)

func (s *Service) AddMember(ctx context.Context, a staff.Actor, orgID, userID pgtype.UUID, role Role) (Membership, error) {
	if !orgID.Valid || !userID.Valid || (role != RoleAdmin && role != RoleMember) {
		return Membership{}, staff.ErrValidation
	}
	var v Membership
	err := s.transaction(ctx, a, func(q *persistence.Queries, at time.Time) error {
		u, err := lockUser(ctx, q, userID)
		if err != nil {
			return err
		}
		if _, err = lockOrganization(ctx, q, orgID); err != nil {
			return err
		}
		if u.Status != "active" {
			return staff.ErrConflict
		}
		r, err := q.CustomerInsertMember(ctx, persistence.CustomerInsertMemberParams{OrganizationID: orgID, UserID: userID, Role: string(role), At: stamp(at)})
		if err != nil {
			return err
		}
		v = membership(r)
		return audit(ctx, q, a, v.ID, orgID, "organization_membership", "customer.membership.add", at, "", "status")
	})
	if err != nil {
		return Membership{}, err
	}
	return v, nil
}
func (s *Service) RemoveMember(ctx context.Context, a staff.Actor, orgID, id pgtype.UUID, expected string) (Membership, error) {
	return s.memberStatus(ctx, a, orgID, id, expected, "removed")
}
func (s *Service) RestoreMember(ctx context.Context, a staff.Actor, orgID, id pgtype.UUID, expected string) (Membership, error) {
	return s.memberStatus(ctx, a, orgID, id, expected, "active")
}
func (s *Service) memberStatus(ctx context.Context, a staff.Actor, orgID, id pgtype.UUID, expected, status string) (Membership, error) {
	if !orgID.Valid || !id.Valid {
		return Membership{}, staff.ErrValidation
	}
	var v Membership
	err := s.transaction(ctx, a, func(q *persistence.Queries, at time.Time) error {
		u, old, err := lockMembership(ctx, q, orgID, id)
		if err != nil {
			return err
		}
		if err = revision(MembershipRevision(old), expected); err != nil {
			return err
		}
		if old.Role == RoleOwner || old.Status == status || (status == "active" && u.Status != "active") {
			return staff.ErrConflict
		}
		r, err := q.CustomerSetMembershipStatus(ctx, persistence.CustomerSetMembershipStatusParams{OrganizationID: orgID, ID: id, Status: status, At: stamp(at)})
		if err != nil {
			return err
		}
		v = membership(r)
		action := "customer.membership.remove"
		if status == "active" {
			action = "customer.membership.restore"
		}
		return audit(ctx, q, a, id, orgID, "organization_membership", action, at, "", "status")
	})
	if err != nil {
		return Membership{}, err
	}
	return v, nil
}
func (s *Service) SetRole(ctx context.Context, a staff.Actor, orgID, id pgtype.UUID, in RoleChange) (Membership, error) {
	if !orgID.Valid || !id.Valid || (in.Role != RoleOwner && in.Role != RoleAdmin && in.Role != RoleMember) {
		return Membership{}, staff.ErrValidation
	}
	var v Membership
	err := s.confirmed(ctx, a, in.Sensitive, func(q *persistence.Queries, at time.Time) error {
		u, old, err := lockMembership(ctx, q, orgID, id)
		if err != nil {
			return err
		}
		if err = revision(MembershipRevision(old), in.Revision); err != nil {
			return err
		}
		if old.Status != "active" || old.Role == in.Role || (in.Role == RoleOwner && u.Status != "active") {
			return staff.ErrConflict
		}
		if old.Role == RoleOwner && in.Role != RoleOwner {
			other, err := q.CustomerOtherOwnerExists(ctx, persistence.CustomerOtherOwnerExistsParams{OrganizationID: orgID, ID: id})
			if err != nil {
				return err
			}
			if !other {
				return staff.ErrConflict
			}
		}
		r, err := q.CustomerSetMembershipRole(ctx, persistence.CustomerSetMembershipRoleParams{OrganizationID: orgID, ID: id, Role: string(in.Role), At: stamp(at)})
		if err != nil {
			return err
		}
		v = membership(r)
		return audit(ctx, q, a, id, orgID, "organization_membership", "customer.membership.role", at, in.Reason, "role")
	})
	if err != nil {
		return Membership{}, err
	}
	return v, nil
}
