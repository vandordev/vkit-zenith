//go:build integration

package securitytest

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

// Missing live authority/credential rechecks would admit the callback here.
func TestBOCustomerConfirmationAuthorityAndRollback(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		a, err := fixtureActor(ctx, p, root)
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
		calls := 0
		callback := func(_ context.Context, _ pgx.Tx, at time.Time) error {
			calls++
			if !at.Equal(fixtureNow) {
				return errors.New("wrong operation instant")
			}
			return nil
		}
		if err = svc.ConfirmedMutation(ctx, a, fixturePassword, callback); err != nil {
			return err
		}
		if calls != 1 {
			return errors.New("confirmed callback not invoked once")
		}
		for _, tc := range []struct {
			actor    staff.Actor
			password string
			service  *staff.Service
		}{
			{b, fixturePassword, svc}, {a, "", svc}, {a, "wrong fixture password", svc},
			{a, fixturePassword, staff.NewService(p, func() time.Time { return fixtureNow.Add(30 * 24 * time.Hour) }, staff.NewPasswords())},
		} {
			err = tc.service.ConfirmedMutation(ctx, tc.actor, tc.password, callback)
			if !errors.Is(err, staff.ErrDenied) && !errors.Is(err, staff.ErrUnauthenticated) {
				return errors.New("invalid confirmation accepted")
			}
			if strings.Contains(err.Error(), fixturePassword) || strings.Contains(err.Error(), "wrong fixture") {
				return errors.New("secret leaked")
			}
		}
		if calls != 1 {
			return errors.New("denied callback invoked")
		}
		err = svc.ConfirmedMutation(ctx, a, fixturePassword, func(ctx context.Context, tx pgx.Tx, at time.Time) error {
			if _, err := tx.Exec(ctx, `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at) VALUES(gen_random_uuid(),'rollback@example.test','rollback@example.test','Rollback','active',$1,$1)`, at); err != nil {
				return err
			}
			return staff.ErrConflict
		})
		if !errors.Is(err, staff.ErrConflict) {
			return errors.New("callback conflict not preserved")
		}
		var n int
		if err = p.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&n); err != nil {
			return err
		}
		if n != 0 {
			return errors.New("callback did not roll back")
		}
		if _, err = p.Exec(ctx, `UPDATE platform_staff_sessions SET revoked_at=$1 WHERE token_hash=$2`, fixtureNow, a.SessionHash[:]); err != nil {
			return err
		}
		if err = svc.ConfirmedMutation(ctx, a, fixturePassword, callback); !errors.Is(err, staff.ErrUnauthenticated) {
			return errors.New("revoked confirmation accepted")
		}
		if calls != 1 {
			return errors.New("revoked callback invoked")
		}
		return nil
	})
}

func TestBOCustomerConfirmationRechecksAfterAdmission(t *testing.T) {
	for _, kind := range []string{"password", "role", "disabled", "revoked", "expired"} {
		t.Run(kind, func(t *testing.T) {
			governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
				root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
				if err != nil {
					return err
				}
				a, err := fixtureActor(ctx, p, root)
				if err != nil {
					return err
				}
				var now atomic.Int64
				now.Store(fixtureNow.Unix())
				waiting := staff.NewService(p, func() time.Time { return time.Unix(now.Load(), 0).UTC() }, staff.NewPasswords())
				in := bootstrapInput("second@example.test")
				in.CurrentPassword = fixturePassword
				if _, err = svc.CreateStaff(ctx, a, in); err != nil {
					return err
				}
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
				called := false
				done := make(chan error, 1)
				go func() {
					done <- waiting.ConfirmedMutation(ctx, a, fixturePassword, func(context.Context, pgx.Tx, time.Time) error { called = true; return nil })
				}()
				if err = waitGovernanceBlocked(ctx, p, holder.Conn().PgConn().PID()); err != nil {
					_ = tx.Rollback(ctx)
					<-done
					return err
				}
				switch kind {
				case "password":
					_, err = tx.Exec(ctx, `UPDATE platform_staff_credentials SET changed_at=changed_at+interval '1 second' WHERE staff_id=$1`, root.ID)
				case "role":
					_, err = tx.Exec(ctx, `UPDATE platform_staff SET role='admin' WHERE id=$1`, root.ID)
				case "disabled":
					_, err = tx.Exec(ctx, `UPDATE platform_staff SET status='disabled' WHERE id=$1`, root.ID)
				case "revoked":
					_, err = tx.Exec(ctx, `UPDATE platform_staff_sessions SET revoked_at=$1 WHERE token_hash=$2`, fixtureNow, a.SessionHash[:])
				case "expired":
					now.Store(fixtureNow.Add(30 * 24 * time.Hour).Unix())
				}
				if err != nil {
					_ = tx.Rollback(ctx)
					<-done
					return err
				}
				if err = tx.Commit(ctx); err != nil {
					<-done
					return err
				}
				err = <-done
				if !errors.Is(err, staff.ErrDenied) && !errors.Is(err, staff.ErrUnauthenticated) {
					return errors.New("stale confirmation survived fence")
				}
				if called {
					return errors.New("stale callback invoked")
				}
				return nil
			})
		})
	}
}
