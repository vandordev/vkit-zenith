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

func sensitive(revision string) customers.Sensitive {
	return customers.Sensitive{Revision: revision, Reason: "Isolated customer maintenance", CurrentPassword: fixturePassword}
}

func TestBOCustomerMembershipLifecycleAndScope(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		o, err := c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "One", Slug: "one", NewOwner: &customers.UserCreate{Email: "one@example.test", DisplayName: "One"}})
		if err != nil {
			return err
		}
		other, err := c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "Other", Slug: "other", OwnerUserID: o.User.ID})
		if err != nil {
			return err
		}
		u, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "member@example.test", DisplayName: "Member"})
		if err != nil {
			return err
		}
		m, err := c.AddMember(ctx, a, o.Organization.ID, u.ID, customers.RoleAdmin)
		if err != nil {
			return err
		}
		if m.Role != customers.RoleAdmin || m.Status != "active" || m.UserID != u.ID || m.OrganizationID != o.Organization.ID {
			return errors.New("invalid added member")
		}
		if _, err = c.AddMember(ctx, a, o.Organization.ID, u.ID, customers.RoleMember); !errors.Is(err, staff.ErrConflict) {
			return errors.New("duplicate active add accepted")
		}
		if _, err = c.RemoveMember(ctx, a, other.Organization.ID, m.ID, customers.MembershipRevision(m)); !errors.Is(err, staff.ErrNotFound) {
			return errors.New("cross-scope removal admitted")
		}
		original := customers.MembershipRevision(m)
		removed, err := c.RemoveMember(ctx, a, o.Organization.ID, m.ID, original)
		if err != nil {
			return err
		}
		if removed.ID != m.ID || removed.Role == customers.RoleOwner || removed.Status != "removed" {
			return errors.New("removal lost identity/state")
		}
		if _, err = c.AddMember(ctx, a, o.Organization.ID, u.ID, customers.RoleMember); !errors.Is(err, staff.ErrConflict) {
			return errors.New("historical membership add admitted")
		}
		if _, err = c.RestoreMember(ctx, a, o.Organization.ID, m.ID, original); !errors.Is(err, staff.ErrConflict) {
			return errors.New("stale restore admitted")
		}
		if _, err = p.Exec(ctx, `UPDATE users SET status='disabled' WHERE id=$1`, u.ID); err != nil {
			return err
		}
		if _, err = c.RestoreMember(ctx, a, o.Organization.ID, m.ID, customers.MembershipRevision(removed)); !errors.Is(err, staff.ErrConflict) {
			return errors.New("disabled restore admitted")
		}
		if _, err = c.AddMember(ctx, a, other.Organization.ID, u.ID, customers.RoleMember); !errors.Is(err, staff.ErrConflict) {
			return errors.New("disabled add admitted")
		}
		if _, err = p.Exec(ctx, `UPDATE users SET status='active' WHERE id=$1`, u.ID); err != nil {
			return err
		}
		restored, err := c.RestoreMember(ctx, a, o.Organization.ID, m.ID, customers.MembershipRevision(removed))
		if err != nil {
			return err
		}
		if restored.ID != m.ID || restored.Role != customers.RoleAdmin || restored.Status != "active" || customers.MembershipRevision(restored) == original {
			return errors.New("restore identity/ownership/ABA violation")
		}
		if _, err = c.RemoveMember(ctx, a, o.Organization.ID, o.Owner.ID, customers.MembershipRevision(o.Owner)); !errors.Is(err, staff.ErrConflict) {
			return errors.New("ordinary owner removal allowed")
		}
		var ok bool
		if err = p.QueryRow(ctx, `SELECT (SELECT count(*) FROM organization_memberships)=3 AND (SELECT count(*) FROM audit_events WHERE action='customer.membership.remove')=1 AND (SELECT count(*) FROM audit_events WHERE action='customer.membership.restore')=1`).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("membership audit/product boundary violated")
		}
		return nil
	})
}

func TestBOCustomerMembershipOwnershipAndRollback(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		o, err := c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "One", Slug: "one", NewOwner: &customers.UserCreate{Email: "one@example.test", DisplayName: "One"}})
		if err != nil {
			return err
		}
		u, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "two@example.test", DisplayName: "Two"})
		if err != nil {
			return err
		}
		m, err := c.AddMember(ctx, a, o.Organization.ID, u.ID, customers.RoleMember)
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
		if _, err = c.SetRole(ctx, b, o.Organization.ID, m.ID, customers.RoleChange{Sensitive: sensitive(customers.MembershipRevision(m)), Role: customers.RoleOwner}); !errors.Is(err, staff.ErrDenied) {
			return errors.New("Admin ownership allowed")
		}
		promoted, err := c.SetRole(ctx, a, o.Organization.ID, m.ID, customers.RoleChange{Sensitive: sensitive(customers.MembershipRevision(m)), Role: customers.RoleOwner})
		if err != nil {
			return err
		}
		if promoted.Role != customers.RoleOwner {
			return errors.New("promotion missing")
		}
		if _, err = c.RemoveMember(ctx, a, o.Organization.ID, promoted.ID, customers.MembershipRevision(promoted)); !errors.Is(err, staff.ErrConflict) {
			return errors.New("owner removed with replacement present")
		}
		if _, err = c.SetRole(ctx, a, o.Organization.ID, m.ID, customers.RoleChange{Sensitive: sensitive(customers.MembershipRevision(m)), Role: customers.RoleMember}); !errors.Is(err, staff.ErrConflict) {
			return errors.New("stale owner write admitted")
		}
		demoted, err := c.SetRole(ctx, a, o.Organization.ID, m.ID, customers.RoleChange{Sensitive: sensitive(customers.MembershipRevision(promoted)), Role: customers.RoleMember})
		if err != nil {
			return err
		}
		if demoted.Role == customers.RoleOwner || demoted.Status != "active" {
			return errors.New("demotion removed ordinary membership")
		}
		removed, err := c.RemoveMember(ctx, a, o.Organization.ID, m.ID, customers.MembershipRevision(demoted))
		if err != nil {
			return err
		}
		if _, err = c.SetRole(ctx, a, o.Organization.ID, m.ID, customers.RoleChange{Sensitive: sensitive(customers.MembershipRevision(removed)), Role: customers.RoleOwner}); !errors.Is(err, staff.ErrConflict) {
			return errors.New("removed member promoted")
		}
		m, err = c.RestoreMember(ctx, a, o.Organization.ID, m.ID, customers.MembershipRevision(removed))
		if err != nil {
			return err
		}
		if m.Role == customers.RoleOwner {
			return errors.New("restore reinstated ownership")
		}
		if _, err = p.Exec(ctx, `UPDATE users SET status='disabled' WHERE id=$1`, u.ID); err != nil {
			return err
		}
		if _, err = c.SetRole(ctx, a, o.Organization.ID, m.ID, customers.RoleChange{Sensitive: sensitive(customers.MembershipRevision(m)), Role: customers.RoleOwner}); !errors.Is(err, staff.ErrConflict) {
			return errors.New("disabled member promoted")
		}
		if _, err = p.Exec(ctx, `UPDATE users SET status='active' WHERE id=$1`, u.ID); err != nil {
			return err
		}
		if err = failAudits(ctx, p); err != nil {
			return err
		}
		if _, err = c.SetRole(ctx, a, o.Organization.ID, m.ID, customers.RoleChange{Sensitive: sensitive(customers.MembershipRevision(m)), Role: customers.RoleOwner}); !errors.Is(err, staff.ErrUnavailable) {
			return errors.New("owner audit error not safe")
		}
		if _, err = c.RemoveMember(ctx, a, o.Organization.ID, m.ID, customers.MembershipRevision(m)); !errors.Is(err, staff.ErrUnavailable) {
			return errors.New("member audit error not safe")
		}
		var ok bool
		if err = p.QueryRow(ctx, `SELECT role <> 'owner' AND status='active' AND updated_at=$2 FROM organization_memberships WHERE id=$1`, m.ID, m.UpdatedAt).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("audit failure committed membership")
		}
		return nil
	})
}
