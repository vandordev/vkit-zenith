//go:build integration

package securitytest

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func TestBOLoginRejectsStaleCredentialAndDisable(t *testing.T) {
	for _, kind := range []string{"password", "disable"} {
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
				in := bootstrapInput("other@example.test")
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
				done := make(chan error, 1)
				go func() {
					_, e := svc.Login(ctx, staff.LoginInput{Email: root.Email, Password: fixturePassword, Source: "127.0.0.1"})
					done <- e
				}()
				if err = waitGovernanceBlocked(ctx, p, holder.Conn().PgConn().PID()); err != nil {
					_ = tx.Rollback(ctx)
					<-done
					return err
				}
				if kind == "password" {
					_, err = tx.Exec(ctx, `UPDATE platform_staff_credentials SET changed_at=changed_at+interval '1 second' WHERE staff_id=$1`, root.ID)
				} else {
					_, err = tx.Exec(ctx, `UPDATE platform_staff SET status='disabled' WHERE id=$1`, root.ID)
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
				if err = <-done; !errors.Is(err, staff.ErrUnauthenticated) {
					return errors.New("stale login issued session")
				}
				var n int
				if err = p.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE action='staff.login' AND outcome='success'`).Scan(&n); err != nil {
					return err
				}
				if n != 0 {
					return errors.New("stale login success audit leaked")
				}
				return nil
			})
		})
	}
}

func TestBOSessionCustomerFenceSamplesAfterWait(t *testing.T) {
	for _, kind := range []string{"expiry", "revoked"} {
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
					service := customers.NewService(p, func() time.Time { return time.Unix(now.Load(), 0) }, svc)
					_, e := service.CreateUser(ctx, a, customers.UserCreate{Email: "fenced@example.test", DisplayName: "Fenced"})
					done <- e
				}()
				if err = waitGovernanceBlocked(ctx, p, holder.Conn().PgConn().PID()); err != nil {
					_ = tx.Rollback(ctx)
					<-done
					return err
				}
				if kind == "expiry" {
					now.Store(fixtureNow.Add(720 * time.Hour).Unix())
				} else {
					if _, err = tx.Exec(ctx, `UPDATE platform_staff_sessions SET revoked_at=created_at WHERE staff_id=$1`, root.ID); err != nil {
						_ = tx.Rollback(ctx)
						<-done
						return err
					}
				}
				if err = tx.Commit(ctx); err != nil {
					<-done
					return err
				}
				if err = <-done; !errors.Is(err, staff.ErrUnauthenticated) {
					return errors.New("customer accepted stale session after wait")
				}
				var count int
				if err = p.QueryRow(ctx, `SELECT count(*) FROM users WHERE email_normalized='fenced@example.test'`).Scan(&count); err != nil {
					return err
				}
				if count != 0 {
					return errors.New("denied fenced customer write persisted")
				}
				return nil
			})
		})
	}
}
