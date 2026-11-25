//go:build integration

package securitytest

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func TestBOCustomerLifecycleStaleVerificationCannotVerifyReplacement(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		u, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "old@example.test", DisplayName: "Person"})
		if err != nil {
			return err
		}
		old := sensitive(customers.UserRevision(u))
		updated, err := c.ChangeUserEmail(ctx, a, u.ID, customers.UserEmail{Sensitive: old, Email: "  NEW@example.test  "})
		if err != nil {
			return err
		}
		if updated.EmailVerifiedAt != nil || updated.ID != u.ID || updated.Email != "new@example.test" {
			return errors.New("replacement changed identity or verified email")
		}
		if _, err = c.VerifyUserEmail(ctx, a, u.ID, old); !errors.Is(err, staff.ErrConflict) {
			return errors.New("stale verification accepted")
		}
		verified, err := c.VerifyUserEmail(ctx, a, u.ID, sensitive(customers.UserRevision(updated)))
		if err != nil {
			return err
		}
		if verified.EmailVerifiedAt == nil || !verified.EmailVerifiedAt.Equal(fixtureNow) {
			return errors.New("manual verification not operation instant")
		}
		if _, err = c.VerifyUserEmail(ctx, a, u.ID, sensitive(customers.UserRevision(verified))); !errors.Is(err, staff.ErrConflict) {
			return errors.New("repeated verification accepted")
		}
		if _, err = c.ChangeUserEmail(ctx, a, u.ID, customers.UserEmail{Sensitive: sensitive(customers.UserRevision(verified)), Email: "NEW@example.test"}); !errors.Is(err, staff.ErrConflict) {
			return errors.New("unchanged normalized address accepted")
		}
		if _, err = c.CreateUser(ctx, a, customers.UserCreate{Email: "taken@example.test", DisplayName: "Taken"}); err != nil {
			return err
		}
		if _, err = c.ChangeUserEmail(ctx, a, u.ID, customers.UserEmail{Sensitive: sensitive(customers.UserRevision(verified)), Email: "TAKEN@example.test"}); !errors.Is(err, staff.ErrConflict) {
			return errors.New("duplicate replacement accepted")
		}
		updated, err = c.ChangeUserEmail(ctx, a, u.ID, customers.UserEmail{Sensitive: sensitive(customers.UserRevision(verified)), Email: "final@example.test"})
		if err != nil {
			return err
		}
		if updated.EmailVerifiedAt != nil || updated.ID != u.ID {
			return errors.New("real replacement did not reset verification")
		}
		return nil
	})
}

func TestBOCustomerLifecycleOwnerDisableAndReactivation(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		o, err := c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "One", Slug: "one", NewOwner: &customers.UserCreate{Email: "owner@example.test", DisplayName: "Owner"}})
		if err != nil {
			return err
		}
		other, err := c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "Other", Slug: "other", OwnerUserID: o.User.ID})
		if err != nil {
			return err
		}
		other.Organization, err = c.SetOrganizationStatus(ctx, a, other.Organization.ID, customers.OrganizationStatus{Sensitive: sensitive(customers.OrganizationRevision(other.Organization)), Status: "suspended"})
		if err != nil {
			return err
		}
		if err = seedCustomerAuthorization(ctx, p, o.User.ID); err != nil {
			return err
		}
		beforeDenied, err := customerAuthorizationSnapshot(ctx, p)
		if err != nil {
			return err
		}
		if _, err = c.SetUserStatus(ctx, a, o.User.ID, customers.UserStatus{Sensitive: sensitive(customers.UserRevision(o.User)), Status: "disabled"}); !errors.Is(err, staff.ErrConflict) {
			return errors.New("multi-organization last owner disabled")
		}
		afterDenied, err := customerAuthorizationSnapshot(ctx, p)
		if err != nil {
			return err
		}
		if beforeDenied != afterDenied {
			return errors.New("last-owner rejection changed authorization/history")
		}
		u, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "replacement@example.test", DisplayName: "Replacement"})
		if err != nil {
			return err
		}
		for _, org := range []customers.Organization{o.Organization, other.Organization} {
			m, err := c.AddMember(ctx, a, org.ID, u.ID, customers.RoleMember)
			if err != nil {
				return err
			}
			if _, err = c.SetRole(ctx, a, org.ID, m.ID, customers.RoleChange{Sensitive: sensitive(customers.MembershipRevision(m)), Role: customers.RoleOwner}); err != nil {
				return err
			}
			if org.ID == o.Organization.ID {
				if _, err = c.SetUserStatus(ctx, a, o.User.ID, customers.UserStatus{Sensitive: sensitive(customers.UserRevision(o.User)), Status: "disabled"}); !errors.Is(err, staff.ErrConflict) {
					return errors.New("suspended organization's last owner disabled after active organization repaired")
				}
			}
		}
		disabled, err := c.SetUserStatus(ctx, a, o.User.ID, customers.UserStatus{Sensitive: sensitive(customers.UserRevision(o.User)), Status: "disabled"})
		if err != nil {
			return err
		}
		if disabled.ID != o.User.ID || disabled.Status != "disabled" {
			return errors.New("disable identity mismatch")
		}
		if err = assertCustomerRevocation(ctx, p, o.User.ID); err != nil {
			return err
		}
		if _, err = c.SetUserStatus(ctx, a, disabled.ID, customers.UserStatus{Sensitive: sensitive(customers.UserRevision(disabled)), Status: "disabled"}); !errors.Is(err, staff.ErrConflict) {
			return errors.New("unchanged status accepted")
		}
		verified, err := c.VerifyUserEmail(ctx, a, disabled.ID, sensitive(customers.UserRevision(disabled)))
		if err != nil {
			return err
		}
		if verified.Status != "disabled" {
			return errors.New("verification activated disabled user")
		}
		active, err := c.SetUserStatus(ctx, a, verified.ID, customers.UserStatus{Sensitive: sensitive(customers.UserRevision(verified)), Status: "active"})
		if err != nil {
			return err
		}
		if active.Status != "active" {
			return errors.New("reactivation failed")
		}
		var ok bool
		if err = p.QueryRow(ctx, `SELECT (SELECT count(*) FROM organization_memberships WHERE user_id=$1 AND status='active' AND role = 'owner')=2 AND (SELECT status='suspended' FROM organizations WHERE id=$2)`, o.User.ID, other.Organization.ID).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("lifecycle changed membership/product state")
		}
		return nil
	})
}

func TestBOCustomerLifecycleAuthorityReasonAndPassword(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		u, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "person@example.test", DisplayName: "Person"})
		if err != nil {
			return err
		}
		o, err := c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "One", Slug: "one", OwnerUserID: u.ID})
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
		for _, fn := range []func() error{
			func() error {
				_, e := c.SetUserStatus(ctx, b, u.ID, customers.UserStatus{Sensitive: sensitive(customers.UserRevision(u)), Status: "disabled"})
				return e
			},
			func() error {
				_, e := c.ChangeUserEmail(ctx, b, u.ID, customers.UserEmail{Sensitive: sensitive(customers.UserRevision(u)), Email: "denied@example.test"})
				return e
			},
			func() error { _, e := c.VerifyUserEmail(ctx, b, u.ID, sensitive(customers.UserRevision(u))); return e },
			func() error {
				_, e := c.SetOrganizationStatus(ctx, b, o.Organization.ID, customers.OrganizationStatus{Sensitive: sensitive(customers.OrganizationRevision(o.Organization)), Status: "suspended"})
				return e
			},
		} {
			if err = fn(); !errors.Is(err, staff.ErrDenied) {
				return errors.New("Admin sensitive action admitted")
			}
		}
		for _, reason := range []string{"", " ", strings.Repeat("界", 1025), "secret " + fixturePassword, "$argon2id$fake-credential", string([]byte{255})} {
			s := sensitive(customers.UserRevision(u))
			s.Reason = reason
			if _, err = c.VerifyUserEmail(ctx, a, u.ID, s); !errors.Is(err, staff.ErrValidation) {
				return errors.New("invalid/credential reason admitted")
			}
		}
		for _, password := range []string{"", "wrong fixture password"} {
			s := sensitive(customers.UserRevision(u))
			s.CurrentPassword = password
			if _, err = c.VerifyUserEmail(ctx, a, u.ID, s); !errors.Is(err, staff.ErrDenied) {
				return errors.New("bad confirmation admitted")
			}
		}
		var ok bool
		if err = p.QueryRow(ctx, `SELECT email='person@example.test' AND status='active' AND email_verified_at IS NULL AND NOT EXISTS(SELECT 1 FROM audit_events WHERE action IN ('customer.user.status','customer.user.email_change','customer.user.email_verify','customer.organization.status')) FROM users WHERE id=$1`, u.ID).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("denied action persisted")
		}
		return nil
	})
}
