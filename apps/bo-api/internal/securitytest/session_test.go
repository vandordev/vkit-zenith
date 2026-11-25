//go:build integration

package securitytest

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func TestBOLoginAndSessionBoundary(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		var now atomic.Int64
		now.Store(fixtureNow.UnixMicro())
		svc = staff.NewService(p, func() time.Time { return time.UnixMicro(now.Load()) }, staff.NewPasswords())
		for _, in := range []staff.LoginInput{{Email: "unknown@example.test", Password: fixturePassword, Source: "127.0.0.1"}, {Email: root.Email, Password: replacementPassword, Source: "127.0.0.1"}} {
			if _, err = svc.Login(ctx, in); !errors.Is(err, staff.ErrUnauthenticated) {
				return errors.New("unsafe login failure")
			}
		}
		login, err := svc.Login(ctx, staff.LoginInput{Email: " ROOT@EXAMPLE.TEST ", Password: fixturePassword, Source: "127.0.0.1"})
		if err != nil {
			return err
		}
		if login.ExpiresAt.Sub(fixtureNow) != 720*time.Hour || login.Staff.ID != root.ID {
			return errors.New("invalid login result")
		}
		hash, err := staff.HashSessionToken(login.Token)
		if err != nil {
			return err
		}
		for _, elapsed := range []time.Duration{0, 24 * time.Hour, 720*time.Hour - time.Microsecond, 720 * time.Hour, 721 * time.Hour} {
			now.Store(fixtureNow.Add(elapsed).UnixMicro())
			_, err = svc.Authenticate(ctx, hash)
			if elapsed < 720*time.Hour && err != nil {
				return err
			}
			if elapsed >= 720*time.Hour && !errors.Is(err, staff.ErrUnauthenticated) {
				return errors.New("expired session accepted")
			}
		}
		var expires time.Time
		if err = p.QueryRow(ctx, `SELECT expires_at FROM platform_staff_sessions WHERE token_hash=$1`, hash[:]).Scan(&expires); err != nil {
			return err
		}
		if !expires.Equal(login.ExpiresAt) {
			return errors.New("session extended")
		}
		return nil
	})
}

func TestBOSessionPasswordChangeAndLogout(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		var actors []staff.Actor
		for i := 0; i < 2; i++ {
			r, e := svc.Login(ctx, staff.LoginInput{Email: root.Email, Password: fixturePassword, Source: "127.0.0.1"})
			if e != nil {
				return e
			}
			h, e := staff.HashSessionToken(r.Token)
			if e != nil {
				return e
			}
			actors = append(actors, staff.Actor{StaffID: root.ID, SessionHash: h, Source: "127.0.0.1"})
		}
		if err = svc.ChangePassword(ctx, actors[0], staff.PasswordChange{CurrentPassword: replacementPassword, NewPassword: replacementPassword}); !errors.Is(err, staff.ErrDenied) {
			return errors.New("wrong current password accepted")
		}
		if err = svc.ChangePassword(ctx, actors[0], staff.PasswordChange{CurrentPassword: fixturePassword, NewPassword: fixturePassword}); !errors.Is(err, staff.ErrValidation) {
			return errors.New("same password accepted")
		}
		if err = svc.ChangePassword(ctx, actors[0], staff.PasswordChange{CurrentPassword: fixturePassword, NewPassword: replacementPassword}); err != nil {
			return err
		}
		for _, a := range actors {
			if _, err = svc.Authenticate(ctx, a.SessionHash); !errors.Is(err, staff.ErrUnauthenticated) {
				return errors.New("password change retained session")
			}
		}
		r, err := svc.Login(ctx, staff.LoginInput{Email: root.Email, Password: replacementPassword, Source: "127.0.0.1"})
		if err != nil {
			return err
		}
		h, err := staff.HashSessionToken(r.Token)
		if err != nil {
			return err
		}
		if err = svc.Logout(ctx, staff.Actor{StaffID: root.ID, SessionHash: h, Source: "127.0.0.1"}); err != nil {
			return err
		}
		if _, err = svc.Authenticate(ctx, h); !errors.Is(err, staff.ErrUnauthenticated) {
			return errors.New("logout retained session")
		}
		return nil
	})
}

func TestBOLimiterSharedAccountWindow(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		other := staff.NewService(p, func() time.Time { return fixtureNow }, staff.NewPasswords())
		for i := 0; i < 11; i++ {
			use := svc
			if i%2 == 1 {
				use = other
			}
			_, err := use.Login(ctx, staff.LoginInput{Email: "unknown@example.test", Password: fixturePassword, Source: "127.0.0.1"})
			if i < 10 && !errors.Is(err, staff.ErrUnauthenticated) {
				return err
			}
			if i == 10 {
				var rate *staff.RateLimitError
				if !errors.As(err, &rate) || rate.RetryAfter != 900 {
					return errors.New("shared account limit missing")
				}
			}
		}
		next := staff.NewService(p, func() time.Time { return fixtureNow.Add(15 * time.Minute) }, staff.NewPasswords())
		if _, err := next.Login(ctx, staff.LoginInput{Email: "unknown@example.test", Password: fixturePassword, Source: "127.0.0.1"}); !errors.Is(err, staff.ErrUnauthenticated) {
			return errors.New("next window did not recover")
		}
		var n int
		if err := p.QueryRow(ctx, `SELECT count(*) FROM audit_events WHERE action='staff.login' AND outcome='failure'`).Scan(&n); err != nil {
			return err
		}
		if n != 4 {
			return errors.New("failure audit not bounded per dimension/window")
		}
		return nil
	})
}

func TestBOSessionElapsedDurationAcrossDST(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		conn, err := p.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()
		for _, zone := range []string{"UTC", "America/New_York"} {
			if _, err = conn.Exec(ctx, `SELECT set_config('TimeZone',$1,false)`, zone); err != nil {
				return err
			}
			for _, start := range []time.Time{time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC), time.Date(2026, 10, 25, 12, 0, 0, 0, time.UTC)} {
				_, hash, e := staff.NewSessionToken()
				if e != nil {
					return e
				}
				if _, e = conn.Exec(ctx, `INSERT INTO platform_staff_sessions(id,staff_id,token_hash,created_at,expires_at) VALUES(gen_random_uuid(),$1,$2,$3,$4)`, root.ID, hash[:], start, start.Add(720*time.Hour)); e != nil {
					return e
				}
				_, bad, e := staff.NewSessionToken()
				if e != nil {
					return e
				}
				if _, e = conn.Exec(ctx, `INSERT INTO platform_staff_sessions(id,staff_id,token_hash,created_at,expires_at) VALUES(gen_random_uuid(),$1,$2,$3,$4)`, root.ID, bad[:], start, start.Add(720*time.Hour-time.Second)); e == nil {
					return errors.New("non-exact session duration accepted")
				}
			}
		}
		return nil
	})
}

func TestBOSessionAuditFailureIsAtomic(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		r, err := svc.Login(ctx, staff.LoginInput{Email: root.Email, Password: fixturePassword, Source: "127.0.0.1"})
		if err != nil {
			return err
		}
		hash, err := staff.HashSessionToken(r.Token)
		if err != nil {
			return err
		}
		a := staff.Actor{StaffID: root.ID, SessionHash: hash, Source: "127.0.0.1"}
		if err = failAudits(ctx, p); err != nil {
			return err
		}
		if _, err = svc.Login(ctx, staff.LoginInput{Email: root.Email, Password: fixturePassword, Source: "127.0.0.1"}); !errors.Is(err, staff.ErrUnavailable) {
			return errors.New("login audit failure unsafe")
		}
		if err = svc.Logout(ctx, a); !errors.Is(err, staff.ErrUnavailable) {
			return errors.New("logout audit failure unsafe")
		}
		if err = svc.ChangePassword(ctx, a, staff.PasswordChange{CurrentPassword: fixturePassword, NewPassword: replacementPassword}); !errors.Is(err, staff.ErrUnavailable) {
			return errors.New("password audit failure unsafe")
		}
		if _, err = svc.Authenticate(ctx, hash); err != nil {
			return err
		}
		var ok bool
		if err = p.QueryRow(ctx, `SELECT (SELECT count(*) FROM platform_staff_sessions)=1 AND (SELECT count(*) FROM audit_events)=2 AND NOT EXISTS(SELECT 1 FROM platform_staff_sessions WHERE revoked_at IS NOT NULL)`).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("audit failure left partial mutation")
		}
		return nil
	})
}

func TestBOLoginDisabledAndRoleRevocation(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		a, err := fixtureActor(ctx, p, root)
		if err != nil {
			return err
		}
		in := bootstrapInput("target@example.test")
		in.Role = "admin"
		in.CurrentPassword = fixturePassword
		target, err := svc.CreateStaff(ctx, a, in)
		if err != nil {
			return err
		}
		for _, change := range []string{"role", "disabled", "active"} {
			r, err := svc.Login(ctx, staff.LoginInput{Email: target.Email, Password: fixturePassword, Source: "127.0.0.1"})
			if change == "active" {
				if !errors.Is(err, staff.ErrUnauthenticated) {
					return errors.New("disabled account authenticated")
				}
			} else if err != nil {
				return err
			}
			hash, _ := staff.HashSessionToken(r.Token)
			update := staff.StaffWrite{Email: target.Email, DisplayName: target.DisplayName, Role: target.Role, Status: target.Status, CurrentPassword: fixturePassword, Revision: staff.Revision(target)}
			if change == "role" {
				update.Role = "superadmin"
			} else {
				update.Status = change
			}
			target, err = svc.UpdateStaff(ctx, a, target.ID, update)
			if err != nil {
				return err
			}
			if _, err = svc.Authenticate(ctx, hash); !errors.Is(err, staff.ErrUnauthenticated) {
				return errors.New("role/disable revived session")
			}
		}
		return nil
	})
}
