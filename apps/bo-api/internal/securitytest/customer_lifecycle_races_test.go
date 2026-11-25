//go:build integration

package securitytest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func TestBOCustomerLifecycleCredentialReplacementRace(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		u, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "person@example.test", DisplayName: "Person"})
		if err != nil {
			return err
		}
		if err = seedCustomerAuthorization(ctx, p, u.ID); err != nil {
			return err
		}
		var before string
		holder, err := p.Acquire(ctx)
		if err != nil {
			return err
		}
		defer holder.Release()
		tx, err := holder.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
		if _, err = tx.Exec(ctx, `UPDATE platform_staff_auth_control SET mutation_epoch=mutation_epoch+1 WHERE id=1`); err != nil {
			return err
		}
		done := make(chan error, 1)
		go func() {
			_, e := c.ChangeUserEmail(ctx, a, u.ID, customers.UserEmail{Sensitive: sensitive(customers.UserRevision(u)), Email: "stale@example.test"})
			done <- e
		}()
		if err = waitGovernanceBlocked(ctx, p, holder.Conn().PgConn().PID()); err != nil {
			_ = tx.Rollback(ctx)
			<-done
			return err
		}
		if _, err = tx.Exec(ctx, `UPDATE platform_staff_credentials SET changed_at=changed_at+interval '1 second' WHERE staff_id=$1`, a.StaffID); err != nil {
			_ = tx.Rollback(ctx)
			<-done
			return err
		}
		// The schema legitimately revokes staff sessions on credential change.
		// Capture that intentional external effect, not a false customer delta.
		if err = tx.QueryRow(ctx, customerAuthorizationSnapshotSQL).Scan(&before); err != nil {
			_ = tx.Rollback(ctx)
			<-done
			return err
		}
		if err = tx.Commit(ctx); err != nil {
			<-done
			return err
		}
		if err = <-done; !errors.Is(err, staff.ErrDenied) && !errors.Is(err, staff.ErrUnauthenticated) {
			return errors.New("stale credential changed customer")
		}
		after, err := customerAuthorizationSnapshot(ctx, p)
		if err != nil {
			return err
		}
		if before != after {
			return errors.New("stale confirmation committed customer/revocation")
		}
		return nil
	})
}

func TestBOCustomerLifecycleConcurrentEmailVerification(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		u, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "old@example.test", DisplayName: "Person"})
		if err != nil {
			return err
		}
		x, y, err := customerRace(ctx, p, func() error {
			_, e := c.ChangeUserEmail(ctx, a, u.ID, customers.UserEmail{Sensitive: sensitive(customers.UserRevision(u)), Email: "new@example.test"})
			return e
		}, func() error { _, e := c.VerifyUserEmail(ctx, a, u.ID, sensitive(customers.UserRevision(u))); return e })
		if err != nil {
			return err
		}
		if !((x == nil && errors.Is(y, staff.ErrConflict)) || (y == nil && errors.Is(x, staff.ErrConflict))) {
			return errors.New("email/verification admitted stale revision")
		}
		var ok bool
		if err = p.QueryRow(ctx, `SELECT (email='new@example.test' AND email_verified_at IS NULL) OR (email='old@example.test' AND email_verified_at=$2) FROM users WHERE id=$1`, u.ID, fixtureNow).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("stale verification verified replacement")
		}
		return nil
	})
}

func TestBOCustomerOwnerCompetingDisableDemotion(t *testing.T) {
	for _, op := range []string{"demotion", "disable"} {
		t.Run(op, func(t *testing.T) {
			for _, status := range []string{"active", "suspended"} {
				t.Run(status, func(t *testing.T) {
					governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
						c, a, err := customerCommands(ctx, p, svc)
						if err != nil {
							return err
						}
						in := bootstrapInput("other-staff@example.test")
						in.CurrentPassword = fixturePassword
						second, err := svc.CreateStaff(ctx, a, in)
						if err != nil {
							return err
						}
						b, err := fixtureActor(ctx, p, second)
						if err != nil {
							return err
						}
						other := customers.NewService(p, func() time.Time { return fixtureNow }, staff.NewService(p, func() time.Time { return fixtureNow }, staff.NewPasswords()))
						o, err := c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "Example", Slug: "example", NewOwner: &customers.UserCreate{Email: "one@example.test", DisplayName: "One"}})
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
						m, err = c.SetRole(ctx, a, o.Organization.ID, m.ID, customers.RoleChange{Sensitive: sensitive(customers.MembershipRevision(m)), Role: customers.RoleOwner})
						if err != nil {
							return err
						}
						if _, err = p.Exec(ctx, `UPDATE organizations SET status=$2 WHERE id=$1`, o.Organization.ID, status); err != nil {
							return err
						}
						x, y, err := customerRace(ctx, p, func() error {
							_, e := c.SetUserStatus(ctx, a, o.User.ID, customers.UserStatus{Sensitive: sensitive(customers.UserRevision(o.User)), Status: "disabled"})
							return e
						}, func() error {
							if op == "disable" {
								_, e := other.SetUserStatus(ctx, b, u.ID, customers.UserStatus{Sensitive: sensitive(customers.UserRevision(u)), Status: "disabled"})
								return e
							}
							_, e := other.SetRole(ctx, b, o.Organization.ID, m.ID, customers.RoleChange{Sensitive: sensitive(customers.MembershipRevision(m)), Role: customers.RoleMember})
							return e
						})
						if err != nil {
							return err
						}
						if !((x == nil && errors.Is(y, staff.ErrConflict)) || (y == nil && errors.Is(x, staff.ErrConflict))) {
							return errors.New("disable/demotion failed safe competition")
						}
						n, err := eligibleOwners(ctx, p, o.Organization.ID)
						if err != nil {
							return err
						}
						if n != 1 {
							return errors.New("disable/demotion lost final owner")
						}
						return nil
					})
				})
			}
		})
	}
}

func TestBOCustomerOwnerDisableAgainstAddRemove(t *testing.T) {
	for _, action := range []string{"owner-add", "member-remove"} {
		t.Run(action, func(t *testing.T) {
			governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
				c, a, err := customerCommands(ctx, p, svc)
				if err != nil {
					return err
				}
				o, err := c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "Example", Slug: "example", NewOwner: &customers.UserCreate{Email: "owner@example.test", DisplayName: "Owner"}})
				if err != nil {
					return err
				}
				u, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "member@example.test", DisplayName: "Member"})
				if err != nil {
					return err
				}
				m, err := c.AddMember(ctx, a, o.Organization.ID, u.ID, customers.RoleMember)
				if err != nil {
					return err
				}
				x, y, err := customerRace(ctx, p, func() error {
					_, e := c.SetUserStatus(ctx, a, u.ID, customers.UserStatus{Sensitive: sensitive(customers.UserRevision(u)), Status: "disabled"})
					return e
				}, func() error {
					if action == "owner-add" {
						_, e := c.SetRole(ctx, a, o.Organization.ID, m.ID, customers.RoleChange{Sensitive: sensitive(customers.MembershipRevision(m)), Role: customers.RoleOwner})
						return e
					}
					_, e := c.RemoveMember(ctx, a, o.Organization.ID, m.ID, customers.MembershipRevision(m))
					return e
				})
				if err != nil {
					return err
				}
				if x != nil {
					return x
				}
				if y != nil && !errors.Is(y, staff.ErrConflict) {
					return y
				}
				var eligible bool
				if err = p.QueryRow(ctx, `SELECT u.status='disabled' AND (m.status='removed' OR m.status='active') FROM users u JOIN organization_memberships m ON m.user_id=u.id WHERE u.id=$1 AND m.id=$2`, u.ID, m.ID).Scan(&eligible); err != nil {
					return err
				}
				if !eligible {
					return errors.New("disable race state invalid")
				}
				for _, id := range []pgtype.UUID{o.Organization.ID} {
					n, err := eligibleOwners(ctx, p, id)
					if err != nil {
						return err
					}
					if n < 1 {
						return errors.New("disable race lost owner")
					}
				}
				return nil
			})
		})
	}
}
