//go:build integration

package securitytest

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func TestBOLimiterSourceAndCanonicalIPv6(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		for i := 0; i < 101; i++ {
			source := fmt.Sprintf("2001:db8:abcd:1234::%x", i+1)
			_, err := svc.Login(ctx, staff.LoginInput{Email: "malformed", Password: "short", Source: source})
			if i < 100 && !errors.Is(err, staff.ErrValidation) {
				return fmt.Errorf("source attempt %d: %w", i, err)
			}
			if i == 100 && !errors.Is(err, staff.ErrRateLimited) {
				return errors.New("101st /64 attempt admitted")
			}
		}
		if _, err := svc.Login(ctx, staff.LoginInput{Email: "bad", Source: "2001:db8:abcd:1235::1"}); !errors.Is(err, staff.ErrValidation) {
			return errors.New("unrelated /64 throttled")
		}
		for _, pair := range [][2]string{{"::ffff:192.0.2.1", "192.0.2.1"}, {"2001:db8::dead", "2001:db8::/64"}} {
			v, e := staff.CanonicalSource(pair[0])
			if e != nil || v != pair[1] {
				return errors.New("source canonicalization mismatch")
			}
		}
		for _, bad := range []string{"", "1.2.3.4, 5.6.7.8", "fe80::1%eth0", "not-an-ip"} {
			if _, e := staff.CanonicalSource(bad); !errors.Is(e, staff.ErrValidation) {
				return errors.New("invalid source accepted")
			}
		}
		var count int
		if err := p.QueryRow(ctx, `SELECT count(*) FROM platform_staff_auth_buckets`).Scan(&count); err != nil {
			return err
		}
		if count != 2 {
			return errors.New("IPv6 source generated unbounded counters")
		}
		return nil
	})
}

func TestBOLimiterParallelCapCleanup(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		_, err := p.Exec(ctx, `INSERT INTO platform_staff_auth_buckets(purpose,dimension,key_hash,window_start,attempts,expires_at)
 SELECT 'login','account',sha256(i::text::bytea),$1,1,$2 FROM generate_series(1,19999)i`, fixtureNow, fixtureNow.Add(15*time.Minute))
		if err != nil {
			return err
		}
		done := make(chan error, 4)
		start := make(chan struct{})
		for i := 0; i < 4; i++ {
			go func(i int) {
				<-start
				other := staff.NewService(p, func() time.Time { return fixtureNow }, staff.NewPasswords())
				_, e := other.Login(ctx, staff.LoginInput{Email: "malformed", Source: fmt.Sprintf("192.0.2.%d", i+1)})
				done <- e
			}(i)
		}
		close(start)
		allowed, limited := 0, 0
		for i := 0; i < 4; i++ {
			e := <-done
			if errors.Is(e, staff.ErrValidation) {
				allowed++
			} else if errors.Is(e, staff.ErrRateLimited) {
				limited++
			} else {
				return e
			}
		}
		if allowed != 1 || limited != 3 {
			return errors.New("replicas exceeded row cap")
		}
		var n int
		if err = p.QueryRow(ctx, `SELECT count(*) FROM platform_staff_auth_buckets`).Scan(&n); err != nil {
			return err
		}
		if n != 20000 {
			return errors.New("cap mismatch")
		}
		next := staff.NewService(p, func() time.Time { return fixtureNow.Add(15 * time.Minute) }, staff.NewPasswords())
		if _, err = next.Login(ctx, staff.LoginInput{Email: "malformed", Source: "192.0.2.100"}); !errors.Is(err, staff.ErrValidation) {
			return errors.New("cleanup failed to admit new key")
		}
		if err = p.QueryRow(ctx, `SELECT count(*) FROM platform_staff_auth_buckets`).Scan(&n); err != nil {
			return err
		}
		if n != 19745 {
			return errors.New("cleanup chunk not bounded at 256")
		}
		return nil
	})
}

func TestBOLimiterConfirmationAndAuditRollback(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		a, err := fixtureActor(ctx, p, root)
		if err != nil {
			return err
		}
		for i := 0; i < 11; i++ {
			err = svc.ChangePassword(ctx, a, staff.PasswordChange{CurrentPassword: replacementPassword, NewPassword: fixturePassword})
			if i < 10 && !errors.Is(err, staff.ErrDenied) {
				return err
			}
			if i == 10 && !errors.Is(err, staff.ErrRateLimited) {
				return errors.New("confirmation not limited")
			}
		}
		if err = failAudits(ctx, p); err != nil {
			return err
		}
		if _, err = svc.Login(ctx, staff.LoginInput{Email: "unknown@example.test", Password: fixturePassword, Source: "192.0.2.3"}); !errors.Is(err, staff.ErrUnavailable) {
			return errors.New("audit failure unsafe")
		}
		var ok bool
		if err = p.QueryRow(ctx, `SELECT NOT EXISTS(SELECT 1 FROM platform_staff_auth_buckets WHERE purpose='login' AND failure_recorded) AND (SELECT count(*) FROM platform_staff_sessions)=1 AND (SELECT count(*) FROM audit_events WHERE outcome='success')=1`).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("failed audit marker or session partially committed")
		}
		return nil
	})
}
