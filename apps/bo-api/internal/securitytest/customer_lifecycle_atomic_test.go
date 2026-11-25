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

func TestBOCustomerLifecycleRevisionABAAndAuditRollback(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		u, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "person@example.test", DisplayName: "Person"})
		if err != nil {
			return err
		}
		initial := customers.UserRevision(u)
		for _, status := range []string{"disabled", "active"} {
			u, err = c.SetUserStatus(ctx, a, u.ID, customers.UserStatus{Sensitive: sensitive(customers.UserRevision(u)), Status: status})
			if err != nil {
				return err
			}
		}
		if customers.UserRevision(u) == initial {
			return errors.New("status ABA recreated revision")
		}
		if _, err = c.VerifyUserEmail(ctx, a, u.ID, sensitive(initial)); !errors.Is(err, staff.ErrConflict) {
			return errors.New("ABA stale user revision admitted")
		}
		o, err := c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "Example", Slug: "example", OwnerUserID: u.ID})
		if err != nil {
			return err
		}
		orgInitial := customers.OrganizationRevision(o.Organization)
		for _, status := range []string{"suspended", "active"} {
			o.Organization, err = c.SetOrganizationStatus(ctx, a, o.Organization.ID, customers.OrganizationStatus{Sensitive: sensitive(customers.OrganizationRevision(o.Organization)), Status: status})
			if err != nil {
				return err
			}
		}
		if customers.OrganizationRevision(o.Organization) == orgInitial {
			return errors.New("organization status ABA recreated revision")
		}
		if _, err = c.SetOrganizationStatus(ctx, a, o.Organization.ID, customers.OrganizationStatus{Sensitive: sensitive(orgInitial), Status: "suspended"}); !errors.Is(err, staff.ErrConflict) {
			return errors.New("ABA stale organization revision admitted")
		}
		if err = failAudits(ctx, p); err != nil {
			return err
		}
		if _, err = c.VerifyUserEmail(ctx, a, u.ID, sensitive(customers.UserRevision(u))); !errors.Is(err, staff.ErrUnavailable) {
			return errors.New("verification audit failure unsafe")
		}
		if _, err = c.SetOrganizationStatus(ctx, a, o.Organization.ID, customers.OrganizationStatus{Sensitive: sensitive(customers.OrganizationRevision(o.Organization)), Status: "suspended"}); !errors.Is(err, staff.ErrUnavailable) {
			return errors.New("organization audit failure unsafe")
		}
		var ok bool
		if err = p.QueryRow(ctx, `SELECT u.status='active' AND u.email_verified_at IS NULL AND u.updated_at=$2 AND o.status='active' AND o.updated_at=$4 FROM users u CROSS JOIN organizations o WHERE u.id=$1 AND o.id=$3`, u.ID, u.UpdatedAt, o.Organization.ID, o.Organization.UpdatedAt).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("audit failure changed lifecycle state")
		}
		return nil
	})
}
