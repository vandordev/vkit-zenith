//go:build integration

package securitytest

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func TestBOCustomerHTTPConfirmationRetryAfter(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		u, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "limited@example.test", DisplayName: "Limited"})
		if err != nil {
			return err
		}
		e, _ := customerRouter(p, svc, func() time.Time { return fixtureNow })
		cookies, csrf, err := loginHTTP(e, "root@example.test")
		if err != nil {
			return err
		}
		body := customerJSON(map[string]any{"revision": customers.UserRevision(u), "reason": "Fixture confirmation limit", "currentPassword": "incorrect fixture password"})
		for i := 0; i < 11; i++ {
			w := perform(e, "POST", "/v1/users/"+u.ID.String()+"/email/verify", body, boOrigin, csrf, cookies)
			if i < 10 && w.Code != 403 {
				return fmt.Errorf("confirmation attempt %d: %d", i, w.Code)
			}
			if i == 10 && (w.Code != 429 || w.Header().Get("Retry-After") == "") {
				return fmt.Errorf("missing confirmation Retry-After: %d", w.Code)
			}
		}
		var unchanged bool
		if err = p.QueryRow(ctx, `SELECT email_verified_at IS NULL AND NOT EXISTS(SELECT 1 FROM audit_events WHERE action='customer.user.email_verify') FROM users WHERE id=$1`, u.ID).Scan(&unchanged); err != nil {
			return err
		}
		if !unchanged {
			return fmt.Errorf("failed confirmation verified customer")
		}
		return nil
	})
}

func TestBOCustomerHTTPAuditRollbackAndEmailVerificationRace(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		u, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "old@example.test", DisplayName: "HTTP atomic"})
		if err != nil {
			return err
		}
		if err = seedCustomerAuthorization(ctx, p, u.ID); err != nil {
			return err
		}
		e, _ := customerRouter(p, svc, func() time.Time { return fixtureNow })
		cookies, csrf, err := loginHTTP(e, "root@example.test")
		if err != nil {
			return err
		}
		path := "/v1/users/" + u.ID.String()
		sensitiveBody := func(extra map[string]any) map[string]any {
			out := map[string]any{"revision": customers.UserRevision(u), "reason": "HTTP atomic fixture", "currentPassword": fixturePassword}
			for k, v := range extra {
				out[k] = v
			}
			return out
		}
		before, err := customerAuthorizationSnapshot(ctx, p)
		if err != nil {
			return err
		}
		if err = failAudits(ctx, p); err != nil {
			return err
		}
		for _, test := range []struct {
			suffix string
			body   any
		}{{"/email", sensitiveBody(map[string]any{"email": "new@example.test"})}, {"/status", sensitiveBody(map[string]any{"status": "disabled"})}, {"/email/verify", sensitiveBody(nil)}} {
			w := perform(e, "POST", path+test.suffix, customerJSON(test.body), boOrigin, csrf, cookies)
			if w.Code != 503 {
				return fmt.Errorf("HTTP audit failure %s: %d %s", test.suffix, w.Code, w.Body.String())
			}
			after, err := customerAuthorizationSnapshot(ctx, p)
			if err != nil {
				return err
			}
			if before != after {
				return fmt.Errorf("HTTP audit failure persisted %s", test.suffix)
			}
		}
		// Dropping only the test-owned failure trigger restores the actual boundary.
		if _, err = p.Exec(ctx, `DROP TRIGGER fixture_reject_audit ON audit_events`); err != nil {
			return err
		}
		invoke := func(suffix string, body any) error {
			w := perform(e, "POST", path+suffix, customerJSON(body), boOrigin, csrf, cookies)
			switch w.Code {
			case 200:
				return nil
			case 409:
				return staff.ErrConflict
			default:
				return fmt.Errorf("race HTTP %s status %d: %s", suffix, w.Code, w.Body.String())
			}
		}
		x, y, err := customerRace(ctx, p, func() error { return invoke("/email", sensitiveBody(map[string]any{"email": "new@example.test"})) }, func() error { return invoke("/email/verify", sensitiveBody(nil)) })
		if err != nil {
			return err
		}
		if !((x == nil && errors.Is(y, staff.ErrConflict)) || (y == nil && errors.Is(x, staff.ErrConflict))) {
			return fmt.Errorf("HTTP email/verify race: %v / %v", x, y)
		}
		var valid bool
		if err = p.QueryRow(ctx, `SELECT (email='old@example.test' AND email_verified_at=$2) OR (email='new@example.test' AND email_verified_at IS NULL) FROM users WHERE id=$1`, u.ID, fixtureNow).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return fmt.Errorf("HTTP verified replacement using stale revision")
		}
		return nil
	})
}
