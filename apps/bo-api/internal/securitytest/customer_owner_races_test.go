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

func eligibleOwners(ctx context.Context, p *pgxpool.Pool, id pgtype.UUID) (int, error) {
	var n int
	err := p.QueryRow(ctx, `SELECT count(*) FROM organization_memberships m JOIN users u ON u.id=m.user_id WHERE m.organization_id=$1 AND m.status='active' AND m.role = 'owner' AND u.status='active'`, id).Scan(&n)
	return n, err
}

func TestBOCustomerOwnerLastCannotBeRevoked(t *testing.T) {
	for _, status := range []string{"active", "suspended"} {
		t.Run(status, func(t *testing.T) {
			governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
				c, a, err := customerCommands(ctx, p, svc)
				if err != nil {
					return err
				}
				o, err := c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "Example", Slug: "example", NewOwner: &customers.UserCreate{Email: "owner@example.test", DisplayName: "Owner"}})
				if err != nil {
					return err
				}
				if _, err = p.Exec(ctx, `UPDATE organizations SET status=$2 WHERE id=$1`, o.Organization.ID, status); err != nil {
					return err
				}
				if _, err = c.SetRole(ctx, a, o.Organization.ID, o.Owner.ID, customers.RoleChange{Sensitive: sensitive(customers.MembershipRevision(o.Owner)), Role: customers.RoleMember}); !errors.Is(err, staff.ErrConflict) {
					return errors.New("last owner demotion not conflict")
				}
				n, err := eligibleOwners(ctx, p, o.Organization.ID)
				if err != nil {
					return err
				}
				if n != 1 {
					return errors.New("last-owner invariant lost")
				}
				return nil
			})
		})
	}
}

// Each competitor is held behind an observed PostgreSQL governance fence before
// release. Separate service/actor and pool connections, no scheduler sleeps.
func customerRace(ctx context.Context, p *pgxpool.Pool, first, second func() error) (error, error, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	holder, err := p.Acquire(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer holder.Release()
	tx, err := holder.Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `UPDATE platform_staff_auth_control SET mutation_epoch=mutation_epoch+1 WHERE id=1`); err != nil {
		return nil, nil, err
	}
	done := make(chan error, 2)
	start := make(chan struct{})
	go func() { <-start; done <- first() }()
	go func() { <-start; done <- second() }()
	close(start)
	bounded, stop := context.WithTimeout(ctx, 10*time.Second)
	defer stop()
	for {
		var n int
		if err = p.QueryRow(bounded, `SELECT count(*) FROM pg_stat_activity WHERE pid<>$1::integer AND datname=current_database() AND cardinality(pg_blocking_pids(pid))>0 AND query LIKE '%UPDATE platform_staff_auth_control SET mutation_epoch%'`, holder.Conn().PgConn().PID()).Scan(&n); err != nil {
			_ = tx.Rollback(ctx)
			<-done
			<-done
			return nil, nil, err
		}
		// Queue followers can be blocked by the first waiter as well as holder.
		if n >= 2 {
			break
		}
		if bounded.Err() != nil {
			_ = tx.Rollback(ctx)
			<-done
			<-done
			return nil, nil, bounded.Err()
		}
	}
	if err = tx.Commit(ctx); err != nil {
		<-done
		<-done
		return nil, nil, err
	}
	return <-done, <-done, nil
}

func TestBOCustomerOwnerCompetingDemotions(t *testing.T) {
	for _, status := range []string{"active", "suspended"} {
		t.Run(status, func(t *testing.T) {
			governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
				c, a, err := customerCommands(ctx, p, svc)
				if err != nil {
					return err
				}
				in := bootstrapInput("second-staff@example.test")
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
					_, e := c.SetRole(ctx, a, o.Organization.ID, o.Owner.ID, customers.RoleChange{Sensitive: sensitive(customers.MembershipRevision(o.Owner)), Role: customers.RoleMember})
					return e
				}, func() error {
					_, e := other.SetRole(ctx, b, o.Organization.ID, m.ID, customers.RoleChange{Sensitive: sensitive(customers.MembershipRevision(m)), Role: customers.RoleMember})
					return e
				})
				if err != nil {
					return err
				}
				if !((x == nil && errors.Is(y, staff.ErrConflict)) || (y == nil && errors.Is(x, staff.ErrConflict))) {
					return errors.New("competing demotions did not safely reject one")
				}
				n, err := eligibleOwners(ctx, p, o.Organization.ID)
				if err != nil {
					return err
				}
				if n != 1 {
					return errors.New("competing demotions lost owner")
				}
				return nil
			})
		})
	}
}

func TestBOCustomerOwnerCompetingAddRemoveAndIsolation(t *testing.T) {
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
		u, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "two@example.test", DisplayName: "Two"})
		if err != nil {
			return err
		}
		m, err := c.AddMember(ctx, a, o.Organization.ID, u.ID, customers.RoleMember)
		if err != nil {
			return err
		}
		x, y, err := customerRace(ctx, p, func() error {
			_, e := c.SetRole(ctx, a, o.Organization.ID, m.ID, customers.RoleChange{Sensitive: sensitive(customers.MembershipRevision(m)), Role: customers.RoleOwner})
			return e
		}, func() error {
			_, e := c.RemoveMember(ctx, a, o.Organization.ID, m.ID, customers.MembershipRevision(m))
			return e
		})
		if err != nil {
			return err
		}
		if !((x == nil && errors.Is(y, staff.ErrConflict)) || (y == nil && errors.Is(x, staff.ErrConflict))) {
			return errors.New("owner-add/member-remove both/neither won")
		}
		if _, err = c.SetRole(ctx, a, other.Organization.ID, m.ID, customers.RoleChange{Sensitive: sensitive(customers.MembershipRevision(m)), Role: customers.RoleOwner}); !errors.Is(err, staff.ErrNotFound) {
			return errors.New("cross-organization ownership allowed")
		}
		for _, id := range []pgtype.UUID{o.Organization.ID, other.Organization.ID} {
			n, err := eligibleOwners(ctx, p, id)
			if err != nil {
				return err
			}
			if n < 1 {
				return errors.New("cross-organization invariant lost")
			}
		}
		return nil
	})
}
