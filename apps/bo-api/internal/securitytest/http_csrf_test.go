//go:build integration

package securitytest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func TestBOHTTPCSRFAndOriginMatrix(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		var now atomic.Int64
		now.Store(fixtureNow.Unix())
		clock := func() time.Time { return time.Unix(now.Load(), 0) }
		e, b := httpRouter(p, svc, clock)
		anonymous, anonCookies, err := token(e, nil)
		if err != nil {
			return err
		}
		for _, origin := range []string{"", "null", "https://evil.example.test", "https://bo.example.test.evil.test"} {
			w := perform(e, "POST", "/v1/auth/login", `{}`, origin, anonymous, anonCookies)
			if w.Code != 403 {
				return fmt.Errorf("untrusted origin status %d", w.Code)
			}
		}
		for _, value := range []string{"", anonymous + "x", "1.0.invalid.invalid.invalid"} {
			w := perform(e, "POST", "/v1/auth/login", `{}`, boOrigin, value, anonCookies)
			if w.Code != 403 {
				return errors.New("invalid CSRF accepted")
			}
		}
		for _, content := range []string{"text/plain", "application/x-www-form-urlencoded", ""} {
			r := httptest.NewRequest("POST", "/v1/auth/login", strings.NewReader(`{}`))
			r.RemoteAddr = "192.0.2.1:1"
			r.Header.Set("Origin", boOrigin)
			r.Header.Set("Content-Type", content)
			r.Header.Set("X-CSRF-Token", anonymous)
			for _, c := range anonCookies {
				r.AddCookie(c)
			}
			w := httptest.NewRecorder()
			e.ServeHTTP(w, r)
			if w.Code != 403 {
				return errors.New("non-JSON mutation accepted")
			}
		}
		now.Store(fixtureNow.Add(30 * time.Minute).Unix())
		if w := perform(e, "POST", "/v1/auth/login", `{}`, boOrigin, anonymous, anonCookies); w.Code != 403 {
			return errors.New("expired anonymous CSRF accepted")
		}
		now.Store(fixtureNow.Unix())
		cookies, csrf, err := loginHTTP(e, root.Email)
		if err != nil {
			return err
		}
		second, csrf2, err := loginHTTP(e, root.Email)
		if err != nil {
			return err
		}
		for _, attack := range []struct {
			token  string
			cookie string
		}{{anonymous, anonymous}, {csrf2, csrf2}, {csrf, csrf2}} {
			swapped := []*http.Cookie{}
			for _, c := range cookies {
				if c.Name != b.CSRFName() {
					swapped = append(swapped, c)
				}
			}
			swapped = append(swapped, &http.Cookie{Name: b.CSRFName(), Value: attack.cookie})
			if w := perform(e, "POST", "/v1/auth/logout", `{}`, boOrigin, attack.token, swapped); w.Code != 403 {
				return errors.New("cross-session/anonymous substitution accepted")
			}
		}
		_ = second
		if w := perform(e, "GET", "/v1/auth/csrf", "", "null", "", cookies); w.Code != 403 {
			return errors.New("foreign token fetch allowed")
		}
		for _, origin := range []string{boOrigin, "https://evil.example.test", ""} {
			r := httptest.NewRequest("OPTIONS", "/v1/staff", nil)
			r.Header.Set("Origin", origin)
			r.Header.Set("Access-Control-Request-Method", "PATCH")
			r.Header.Set("Access-Control-Request-Headers", "content-type,x-csrf-token")
			w := httptest.NewRecorder()
			e.ServeHTTP(w, r)
			if origin == boOrigin {
				if w.Code != 204 || w.Header().Get("Access-Control-Allow-Credentials") != "true" || w.Header().Get("Access-Control-Allow-Origin") != boOrigin {
					return errors.New("allowed preflight failed")
				}
			} else if w.Code != 403 || w.Header().Get("Access-Control-Allow-Origin") != "" {
				return errors.New("foreign preflight allowed")
			}
		}
		if w := perform(e, "POST", "/v1/auth/logout", `{}`, boOrigin, csrf, cookies); w.Code != 204 {
			return fmt.Errorf("logout failed %d", w.Code)
		}
		if w := perform(e, "GET", "/v1/auth/me", "", boOrigin, "", cookies); w.Code != 401 {
			return errors.New("logged out cookie accepted")
		}
		return nil
	})
}

func TestBOHTTPGenericLoginRateAndForwardSpoof(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		if _, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test")); err != nil {
			return err
		}
		e, _ := httpRouter(p, svc, func() time.Time { return fixtureNow })
		csrf, cookies, err := token(e, nil)
		if err != nil {
			return err
		}
		var previous string
		for _, email := range []string{"unknown@example.test", "root@example.test"} {
			w := perform(e, "POST", "/v1/auth/login", fmt.Sprintf(`{"email":%q,"password":%q}`, email, replacementPassword), boOrigin, csrf, cookies)
			if w.Code != 401 {
				return errors.New("wrong/unknown login not generic")
			}
			if previous != "" && w.Body.String() != previous {
				return errors.New("account enumeration response")
			}
			previous = w.Body.String()
		}
		for i := 0; i < 101; i++ {
			r := httptest.NewRequest("POST", "/v1/auth/login", strings.NewReader(`{"email":"bad","password":"short"}`))
			r.RemoteAddr = "198.51.100.1:123"
			r.Header.Set("Origin", boOrigin)
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("X-CSRF-Token", csrf)
			r.Header.Set("X-Forwarded-For", fmt.Sprintf("192.0.2.%d", i))
			for _, c := range cookies {
				r.AddCookie(c)
			}
			w := httptest.NewRecorder()
			e.ServeHTTP(w, r)
			if i < 100 && w.Code != 422 {
				return fmt.Errorf("malformed status %d", w.Code)
			}
			if i == 100 && (w.Code != 429 || w.Header().Get("Retry-After") != "900") {
				return errors.New("spoof bypassed global source limiter")
			}
			if strings.Contains(w.Body.String(), "short") {
				return errors.New("validation echoes secret")
			}
		}
		var ok bool
		if err = p.QueryRow(ctx, `SELECT NOT EXISTS(SELECT 1 FROM platform_staff_sessions) AND NOT EXISTS(SELECT 1 FROM audit_events WHERE action='staff.login' AND outcome='success')`).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return errors.New("denied logins left session/success audit")
		}
		return nil
	})
}
