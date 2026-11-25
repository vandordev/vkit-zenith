package customers

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/packages/application/internal/persistence"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
	"time"
)

func (s *Service) SetUserStatus(ctx context.Context, a staff.Actor, id pgtype.UUID, in UserStatus) (User, error) {
	if !id.Valid || (in.Status != "active" && in.Status != "disabled") {
		return User{}, staff.ErrValidation
	}
	var v User
	err := s.confirmed(ctx, a, in.Sensitive, func(q *persistence.Queries, at time.Time) error {
		old, err := lockUser(ctx, q, id)
		if err != nil {
			return err
		}
		if err = revision(UserRevision(old), in.Revision); err != nil {
			return err
		}
		if old.Status == in.Status {
			return staff.ErrConflict
		}
		// Holding the user prevents compatible membership writers changing scope.
		ids, err := q.CustomerUserOrganizationIDs(ctx, id)
		if err != nil {
			return err
		}
		for _, orgID := range ids {
			if _, err = lockOrganization(ctx, q, orgID); err != nil {
				return err
			}
		}
		if _, err = q.CustomerLockUserMemberships(ctx, id); err != nil {
			return err
		}
		if in.Status == "disabled" {
			for _, orgID := range ids {
				owns, err := q.CustomerUserOwnsOrganization(ctx, persistence.CustomerUserOwnsOrganizationParams{OrganizationID: orgID, UserID: id})
				if err != nil {
					return err
				}
				if !owns {
					continue
				}
				other, err := q.CustomerOtherUserOwnerExists(ctx, persistence.CustomerOtherUserOwnerExistsParams{OrganizationID: orgID, UserID: id})
				if err != nil {
					return err
				}
				if !other {
					return staff.ErrConflict
				}
			}
			if err = revoke(ctx, q, id, at); err != nil {
				return err
			}
		}
		r, err := q.CustomerSetUserStatus(ctx, persistence.CustomerSetUserStatusParams{ID: id, Status: in.Status, At: stamp(at)})
		if err != nil {
			return err
		}
		v = user(r)
		return audit(ctx, q, a, id, pgtype.UUID{}, "user", "customer.user.status", at, in.Reason, "status")
	})
	if err != nil {
		return User{}, err
	}
	return v, nil
}
func (s *Service) SetOrganizationStatus(ctx context.Context, a staff.Actor, id pgtype.UUID, in OrganizationStatus) (Organization, error) {
	if !id.Valid || (in.Status != "active" && in.Status != "suspended") {
		return Organization{}, staff.ErrValidation
	}
	var v Organization
	err := s.confirmed(ctx, a, in.Sensitive, func(q *persistence.Queries, at time.Time) error {
		old, err := lockOrganization(ctx, q, id)
		if err != nil {
			return err
		}
		if err = revision(OrganizationRevision(old), in.Revision); err != nil {
			return err
		}
		if old.Status == in.Status {
			return staff.ErrConflict
		}
		r, err := q.CustomerSetOrganizationStatus(ctx, persistence.CustomerSetOrganizationStatusParams{ID: id, Status: in.Status, At: stamp(at)})
		if err != nil {
			return err
		}
		v = organization(r)
		return audit(ctx, q, a, id, id, "organization", "customer.organization.status", at, in.Reason, "status")
	})
	if err != nil {
		return Organization{}, err
	}
	return v, nil
}
func (s *Service) ChangeUserEmail(ctx context.Context, a staff.Actor, id pgtype.UUID, in UserEmail) (User, error) {
	if !id.Valid || !validEmail(in.Email) {
		return User{}, staff.ErrValidation
	}
	var v User
	err := s.confirmed(ctx, a, in.Sensitive, func(q *persistence.Queries, at time.Time) error {
		old, err := lockUser(ctx, q, id)
		if err != nil {
			return err
		}
		if err = revision(UserRevision(old), in.Revision); err != nil {
			return err
		}
		same, err := q.CustomerEmailUnchanged(ctx, persistence.CustomerEmailUnchangedParams{ID: id, Email: in.Email})
		if err != nil {
			return err
		}
		if same {
			return staff.ErrConflict
		}
		if err = revoke(ctx, q, id, at); err != nil {
			return err
		}
		r, err := q.CustomerChangeUserEmail(ctx, persistence.CustomerChangeUserEmailParams{ID: id, Email: in.Email, At: stamp(at)})
		if err != nil {
			return err
		}
		v = user(r)
		fields := []string{"email"}
		if old.EmailVerifiedAt != nil {
			fields = append(fields, "email_verified_at")
		}
		return audit(ctx, q, a, id, pgtype.UUID{}, "user", "customer.user.email_change", at, in.Reason, fields...)
	})
	if err != nil {
		return User{}, err
	}
	return v, nil
}
func (s *Service) VerifyUserEmail(ctx context.Context, a staff.Actor, id pgtype.UUID, in Sensitive) (User, error) {
	if !id.Valid {
		return User{}, staff.ErrValidation
	}
	var v User
	err := s.confirmed(ctx, a, in, func(q *persistence.Queries, at time.Time) error {
		old, err := lockUser(ctx, q, id)
		if err != nil {
			return err
		}
		if err = revision(UserRevision(old), in.Revision); err != nil {
			return err
		}
		if old.EmailVerifiedAt != nil || at.Before(old.CreatedAt) {
			return staff.ErrConflict
		}
		r, err := q.CustomerVerifyUserEmail(ctx, persistence.CustomerVerifyUserEmailParams{ID: id, At: stamp(at)})
		if err != nil {
			return err
		}
		v = user(r)
		return audit(ctx, q, a, id, pgtype.UUID{}, "user", "customer.user.email_verify", at, in.Reason, "email_verified_at")
	})
	if err != nil {
		return User{}, err
	}
	return v, nil
}
