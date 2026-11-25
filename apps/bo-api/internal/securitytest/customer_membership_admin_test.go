//go:build integration

package securitytest

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func TestBOCustomerMembershipAdminOrdinaryAndAuditRollback(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		in := bootstrapInput("admin@example.test")
		in.Role = "admin"
		in.CurrentPassword = fixturePassword
		s, err := svc.CreateStaff(ctx, a, in)
		if err != nil {
			return err
		}
		b, err := fixtureActor(ctx, p, s)
		if err != nil {
			return err
		}
		o, err := c.CreateOrganization(ctx, b, customers.OrganizationCreate{Name: "Admin", Slug: "admin", NewOwner: &customers.UserCreate{Email: "owner@example.test", DisplayName: "Owner"}})
		if err != nil {
			return err
		}
		u, err := c.CreateUser(ctx, b, customers.UserCreate{Email: "member@example.test", DisplayName: "Member"})
		if err != nil {
			return err
		}
		m, err := c.AddMember(ctx, b, o.Organization.ID, u.ID, customers.RoleMember)
		if err != nil {
			return err
		}
		m, err = c.RemoveMember(ctx, b, o.Organization.ID, m.ID, customers.MembershipRevision(m))
		if err != nil {
			return err
		}
		m, err = c.RestoreMember(ctx, b, o.Organization.ID, m.ID, customers.MembershipRevision(m))
		if err != nil {
			return err
		}
		v, err := c.CreateUser(ctx, b, customers.UserCreate{Email: "rollback@example.test", DisplayName: "Rollback"})
		if err != nil {
			return err
		}
		if err = failAudits(ctx, p); err != nil {
			return err
		}
		if _, err = c.AddMember(ctx, b, o.Organization.ID, v.ID, customers.RoleMember); !errors.Is(err, staff.ErrUnavailable) {
			return errors.New("member-add audit failure not safe")
		}
		var n int
		if err = p.QueryRow(ctx, `SELECT count(*) FROM organization_memberships WHERE user_id=$1`, v.ID).Scan(&n); err != nil {
			return err
		}
		if n != 0 {
			return errors.New("failed add persisted")
		}
		return nil
	})
}
