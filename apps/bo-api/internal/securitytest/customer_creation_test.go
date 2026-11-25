//go:build integration

package securitytest

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func customerCommands(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) (*customers.Service, staff.Actor, error) {
	root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
	if err != nil {
		return nil, staff.Actor{}, err
	}
	a, err := fixtureActor(ctx, p, root)
	return customers.NewService(p, func() time.Time { return fixtureNow }, svc), a, err
}

func TestBOCustomerCreationUnverifiedWithoutCredentials(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		u, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "  PERSON@EXAMPLE.TEST  ", DisplayName: "Person"})
		if err != nil {
			return err
		}
		if u.Status != "active" || u.EmailVerifiedAt != nil || u.Email != "person@example.test" {
			return errors.New("incorrect initial customer state")
		}
		var ok bool
		if err = p.QueryRow(ctx, `SELECT NOT EXISTS(SELECT 1 FROM password_credentials) AND NOT EXISTS(SELECT 1 FROM auth_sessions) AND EXISTS(SELECT 1 FROM audit_events WHERE action='customer.user.create' AND actor_staff_id=$1 AND organization_id IS NULL AND target_id=$2 AND detail='{"changed_fields":["email","display_name","status"]}'::jsonb)`, a.StaffID, u.ID.String()).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("credential/product or audit boundary violated")
		}
		if _, err = c.CreateUser(ctx, a, customers.UserCreate{Email: "PERSON@example.test", DisplayName: "Other"}); !errors.Is(err, staff.ErrConflict) {
			return errors.New("duplicate normalized email accepted")
		}
		original := customers.UserRevision(u)
		for _, name := range []string{"Changed", "Person"} {
			u, err = c.UpdateUser(ctx, a, u.ID, customers.UserProfile{DisplayName: name, Revision: customers.UserRevision(u)})
			if err != nil {
				return err
			}
		}
		if customers.UserRevision(u) == original || !u.UpdatedAt.After(fixtureNow) || u.Email != "person@example.test" || u.Status != "active" || u.EmailVerifiedAt != nil {
			return errors.New("profile ABA or privileged mutation")
		}
		if _, err = c.UpdateUser(ctx, a, u.ID, customers.UserProfile{DisplayName: "Stale", Revision: original}); !errors.Is(err, staff.ErrConflict) {
			return errors.New("stale profile accepted")
		}
		for _, bad := range []customers.UserCreate{{Email: "invalid", DisplayName: "Name"}, {Email: "x@example.test", DisplayName: " "}, {Email: "x@example.test", DisplayName: strings.Repeat("界", 201)}, {Email: strings.Repeat("x", 250) + "@t.test", DisplayName: "Name"}, {Email: "x@example.test", DisplayName: string([]byte{255})}} {
			if _, err = c.CreateUser(ctx, a, bad); !errors.Is(err, staff.ErrValidation) {
				return errors.New("invalid customer accepted")
			}
		}
		return nil
	})
}

func TestBOCustomerCreationOrganizationsAndAtomicAudit(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		u, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "owner@example.test", DisplayName: "Owner"})
		if err != nil {
			return err
		}
		for _, in := range []customers.OrganizationCreate{
			{Name: "Neither", Slug: "neither"},
			{Name: "Both", Slug: "both", OwnerUserID: u.ID, NewOwner: &customers.UserCreate{Email: "both@example.test", DisplayName: "Both"}},
			{Name: "Bad", Slug: "Bad_slug", OwnerUserID: u.ID},
		} {
			if _, err = c.CreateOrganization(ctx, a, in); !errors.Is(err, staff.ErrValidation) {
				return errors.New("invalid owner mode/slug accepted")
			}
		}
		if _, err = p.Exec(ctx, `UPDATE users SET status='disabled' WHERE id=$1`, u.ID); err != nil {
			return err
		}
		if _, err = c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "Disabled", Slug: "disabled", OwnerUserID: u.ID}); !errors.Is(err, staff.ErrConflict) {
			return errors.New("disabled owner accepted")
		}
		if _, err = p.Exec(ctx, `UPDATE users SET status='active' WHERE id=$1`, u.ID); err != nil {
			return err
		}
		created, err := c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "Existing", Slug: "existing", OwnerUserID: u.ID})
		if err != nil {
			return err
		}
		if created.User.ID != u.ID || created.User.EmailVerifiedAt != nil || created.Owner.Role != customers.RoleOwner || created.Owner.Status != "active" || created.Organization.Status != "active" {
			return errors.New("incorrect existing owner creation")
		}
		newOrg, err := c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "New", Slug: "new", NewOwner: &customers.UserCreate{Email: "new@example.test", DisplayName: "New"}})
		if err != nil {
			return err
		}
		var ok bool
		if err = p.QueryRow(ctx, `SELECT (SELECT count(*) FROM audit_events WHERE organization_id=$1 AND action IN ('customer.organization.create','customer.membership.add','customer.membership.role') AND actor_staff_id=$2)=3 AND (SELECT count(*) FROM users)=2 AND (SELECT count(*) FROM organization_memberships)=2 AND NOT EXISTS(SELECT 1 FROM password_credentials)`, newOrg.Organization.ID, a.StaffID).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("creation events/state missing")
		}
		before := customers.OrganizationRevision(created.Organization)
		for _, name := range []string{"Changed", "Existing"} {
			created.Organization, err = c.UpdateOrganization(ctx, a, created.Organization.ID, customers.OrganizationProfile{Name: name, Slug: "existing", Revision: customers.OrganizationRevision(created.Organization)})
			if err != nil {
				return err
			}
		}
		if customers.OrganizationRevision(created.Organization) == before {
			return errors.New("organization ABA revision")
		}
		if _, err = c.UpdateOrganization(ctx, a, created.Organization.ID, customers.OrganizationProfile{Name: "Stale", Slug: "existing", Revision: before}); !errors.Is(err, staff.ErrConflict) {
			return errors.New("stale organization accepted")
		}
		if _, err = c.UpdateOrganization(ctx, a, created.Organization.ID, customers.OrganizationProfile{Name: "Collision", Slug: "new", Revision: customers.OrganizationRevision(created.Organization)}); !errors.Is(err, staff.ErrConflict) {
			return errors.New("slug collision accepted")
		}
		for _, in := range []customers.OrganizationCreate{
			{Name: "Duplicate slug", Slug: "new", NewOwner: &customers.UserCreate{Email: "orphan@example.test", DisplayName: "Orphan"}},
			{Name: "Duplicate email", Slug: "duplicate", NewOwner: &customers.UserCreate{Email: "NEW@example.test", DisplayName: "Duplicate"}},
		} {
			if _, err = c.CreateOrganization(ctx, a, in); !errors.Is(err, staff.ErrConflict) {
				return errors.New("duplicate creation accepted")
			}
		}
		var countsBefore, countsAfter string
		counts := `SELECT jsonb_build_array((SELECT count(*) FROM users),(SELECT count(*) FROM organizations),(SELECT count(*) FROM organization_memberships),(SELECT count(*) FROM audit_events))::text`
		if err = p.QueryRow(ctx, counts).Scan(&countsBefore); err != nil {
			return err
		}
		if err = failAudits(ctx, p); err != nil {
			return err
		}
		for _, in := range []customers.OrganizationCreate{{Name: "Fail existing", Slug: "fail-existing", OwnerUserID: u.ID}, {Name: "Fail new", Slug: "fail-new", NewOwner: &customers.UserCreate{Email: "fail@example.test", DisplayName: "Fail"}}} {
			if _, err = c.CreateOrganization(ctx, a, in); !errors.Is(err, staff.ErrUnavailable) {
				return errors.New("audit failure not safe")
			}
		}
		if _, err = c.CreateUser(ctx, a, customers.UserCreate{Email: "fail-user@example.test", DisplayName: "Fail"}); !errors.Is(err, staff.ErrUnavailable) {
			return errors.New("user audit failure not safe")
		}
		if _, err = c.UpdateUser(ctx, a, u.ID, customers.UserProfile{DisplayName: "Fail", Revision: customers.UserRevision(u)}); !errors.Is(err, staff.ErrUnavailable) {
			return errors.New("profile audit failure not safe")
		}
		if err = p.QueryRow(ctx, counts).Scan(&countsAfter); err != nil {
			return err
		}
		if countsBefore != countsAfter {
			return errors.New("audit failure partially committed")
		}
		if err = p.QueryRow(ctx, `SELECT display_name='Owner' AND email_verified_at IS NULL FROM users WHERE id=$1`, u.ID).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("selected user changed")
		}
		return nil
	})
}

func TestBOCustomerCreationAdminAndLiveAuthority(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		in := bootstrapInput("admin@example.test")
		in.Role = "admin"
		in.CurrentPassword = fixturePassword
		admin, err := svc.CreateStaff(ctx, a, in)
		if err != nil {
			return err
		}
		b, err := fixtureActor(ctx, p, admin)
		if err != nil {
			return err
		}
		created, err := c.CreateOrganization(ctx, b, customers.OrganizationCreate{Name: "Admin", Slug: "admin", NewOwner: &customers.UserCreate{Email: "admin-owner@example.test", DisplayName: "Owner"}})
		if err != nil {
			return err
		}
		if _, err = c.UpdateUser(ctx, b, created.User.ID, customers.UserProfile{DisplayName: "Edited", Revision: customers.UserRevision(created.User)}); err != nil {
			return err
		}
		if _, err = p.Exec(ctx, `UPDATE platform_staff_sessions SET revoked_at=$1 WHERE token_hash=$2`, fixtureNow, b.SessionHash[:]); err != nil {
			return err
		}
		if _, err = c.CreateUser(ctx, b, customers.UserCreate{Email: "denied@example.test", DisplayName: "Denied"}); !errors.Is(err, staff.ErrUnauthenticated) {
			return errors.New("revoked actor admitted")
		}
		if _, err = c.UpdateUser(ctx, a, pgtype.UUID{}, customers.UserProfile{}); !errors.Is(err, staff.ErrValidation) {
			return errors.New("invalid id accepted")
		}
		return nil
	})
}
