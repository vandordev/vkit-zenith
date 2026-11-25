//go:build integration

package securitytest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func TestBOHTTPConcurrentLastSuperadmin(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		a, err := fixtureActor(ctx, p, root)
		if err != nil {
			return err
		}
		in := bootstrapInput("second@example.test")
		in.CurrentPassword = fixturePassword
		second, err := svc.CreateStaff(ctx, a, in)
		if err != nil {
			return err
		}
		e, _ := httpRouter(p, svc, func() time.Time { return fixtureNow })
		start := make(chan struct{})
		done := make(chan int, 2)
		for _, s := range []staff.Staff{root, second} {
			cookies, csrf, err := loginHTTP(e, s.Email)
			if err != nil {
				return err
			}
			body, _ := json.Marshal(map[string]string{"email": s.Email, "displayName": s.DisplayName, "role": "admin", "status": "active", "currentPassword": fixturePassword, "revision": staff.Revision(s)})
			go func(s staff.Staff, c []*http.Cookie, csrf, body string) {
				<-start
				w := perform(e, "PATCH", "/v1/staff/"+s.ID.String(), body, boOrigin, csrf, c)
				done <- w.Code
			}(s, cookies, csrf, string(body))
		}
		close(start)
		wins, conflicts := 0, 0
		for i := 0; i < 2; i++ {
			switch <-done {
			case 200:
				wins++
			case 409:
				conflicts++
			default:
				return errors.New("unexpected last-admin HTTP status")
			}
		}
		if wins != 1 || conflicts != 1 {
			return errors.New("both/neither superadmin demotions accepted")
		}
		var ok bool
		if err = p.QueryRow(ctx, `SELECT (SELECT count(*) FROM platform_staff WHERE role='superadmin' AND status='active')=1 AND (SELECT count(*) FROM audit_events WHERE action='staff.update' AND outcome='success')=1`).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("last-admin HTTP race partially committed")
		}
		return nil
	})
}

func TestBOHTTPRedactsLoggingAndValidation(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		if _, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test")); err != nil {
			return err
		}
		var logs bytes.Buffer
		e, _ := httpRouterLogger(p, svc, func() time.Time { return fixtureNow }, slog.New(slog.NewJSONHandler(&logs, nil)))
		csrf, cookies, err := token(e, nil)
		if err != nil {
			return err
		}
		raw := `{"email":"root@example.test","password":"` + strings.Repeat("SECRET", 100) + `"}`
		w := perform(e, "POST", "/v1/auth/login?password=SECRET", raw, boOrigin, csrf, cookies)
		if w.Code != 422 {
			return errors.New("oversized password not validated")
		}
		all := logs.String() + w.Body.String()
		for _, s := range []string{"SECRET", csrf, "password=", fixturePassword, "argon2"} {
			if strings.Contains(all, s) {
				return errors.New("sensitive input leaked in logger/validation")
			}
		}
		if !strings.Contains(logs.String(), "request") {
			return errors.New("real logger not exercised")
		}
		return nil
	})
}
