package customers

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/packages/application/internal/persistence"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
	"time"
)

func (s *Service) CreateOrganization(ctx context.Context, a staff.Actor, in OrganizationCreate) (OrganizationCreated, error) {
	if !validOrganization(in.Name, in.Slug) || in.OwnerUserID.Valid == (in.NewOwner != nil) || (in.NewOwner != nil && !validUser(*in.NewOwner)) {
		return OrganizationCreated{}, staff.ErrValidation
	}
	var v OrganizationCreated
	err := s.transaction(ctx, a, func(q *persistence.Queries, at time.Time) error {
		var err error
		if in.NewOwner != nil {
			v.User, err = insertUser(ctx, q, a, *in.NewOwner, at)
		} else {
			v.User, err = lockUser(ctx, q, in.OwnerUserID)
		}
		if err != nil {
			return err
		}
		if v.User.Status != "active" {
			return staff.ErrConflict
		}
		r, err := q.CustomerInsertOrganization(ctx, persistence.CustomerInsertOrganizationParams{Name: in.Name, Slug: in.Slug, At: stamp(at)})
		if err != nil {
			return err
		}
		v.Organization = organization(r)
		m, err := q.CustomerInsertInitialOwner(ctx, persistence.CustomerInsertInitialOwnerParams{OrganizationID: r.ID, UserID: v.User.ID, At: stamp(at)})
		if err != nil {
			return err
		}
		v.Owner = membership(m)
		if err = audit(ctx, q, a, r.ID, r.ID, "organization", "customer.organization.create", at, "", "name", "slug", "status"); err != nil {
			return err
		}
		if err = audit(ctx, q, a, m.ID, r.ID, "organization_membership", "customer.membership.add", at, "", "status"); err != nil {
			return err
		}
		return audit(ctx, q, a, m.ID, r.ID, "organization_membership", "customer.membership.role", at, "", "role")
	})
	if err != nil {
		return OrganizationCreated{}, err
	}
	return v, nil
}
func (s *Service) UpdateOrganization(ctx context.Context, a staff.Actor, id pgtype.UUID, in OrganizationProfile) (Organization, error) {
	if !id.Valid || !validOrganization(in.Name, in.Slug) {
		return Organization{}, staff.ErrValidation
	}
	var v Organization
	err := s.transaction(ctx, a, func(q *persistence.Queries, at time.Time) error {
		old, err := lockOrganization(ctx, q, id)
		if err != nil {
			return err
		}
		if err = revision(OrganizationRevision(old), in.Revision); err != nil {
			return err
		}
		fields := []string{}
		if old.Name != in.Name {
			fields = append(fields, "name")
		}
		if old.Slug != in.Slug {
			fields = append(fields, "slug")
		}
		if len(fields) == 0 {
			return staff.ErrConflict
		}
		r, err := q.CustomerUpdateOrganizationProfile(ctx, persistence.CustomerUpdateOrganizationProfileParams{ID: id, Name: in.Name, Slug: in.Slug, At: stamp(at)})
		if err != nil {
			return err
		}
		v = organization(r)
		return audit(ctx, q, a, id, id, "organization", "customer.organization.update", at, "", fields...)
	})
	if err != nil {
		return Organization{}, err
	}
	return v, nil
}
