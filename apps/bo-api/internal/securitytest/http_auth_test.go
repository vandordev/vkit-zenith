//go:build integration

package securitytest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	v1 "github.com/vandordev/vkit-zenith/apps/bo-api/internal/handlers/v1"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/server"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

const boOrigin = "https://bo.example.test"

func httpRouter(p *pgxpool.Pool, s *staff.Service, clock func() time.Time) (*echo.Echo, *security.Boundary) {
	return httpRouterLogger(p, s, clock, slog.New(slog.NewJSONHandler(io.Discard, nil)))
}
func httpRouterLogger(p *pgxpool.Pool, s *staff.Service, clock func() time.Time, logger *slog.Logger, extra ...func(huma.API, *security.Boundary) error) (*echo.Echo, *security.Boundary) {
	e := server.NewHTTP(logger)
	b := &security.Boundary{Pool: p, Service: s, Clock: clock, Origin: boOrigin, Secret: []byte("fixture-only-csrf-secret-at-least-32-bytes")}
	e.Pre(b.Middleware)
	c := huma.DefaultConfig("BO fixture", "1")
	c.OpenAPIPath = ""
	c.DocsPath = ""
	c.SchemasPath = ""
	c.CreateHooks = nil
	api := humaecho.NewV4(e, c)
	registrations := []func(huma.API, *security.Boundary) error{v1.NewGetCSRF, v1.NewLogin, v1.NewLogout, v1.NewGetMe, v1.NewChangePassword, v1.NewCreateStaff, v1.NewUpdateStaff, v1.NewReplaceStaffPassword, v1.NewGetStaff, v1.NewListStaff}
	for _, register := range append(registrations, extra...) {
		if err := register(api, b); err != nil {
			panic(err)
		}
	}
	return e, b
}

func perform(e *echo.Echo, method, path, body, origin, csrf string, cookies []*http.Cookie) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.RemoteAddr = "192.0.2.1:1234"
	r.Header.Set("Origin", origin)
	r.Header.Set("Content-Type", "application/json")
	if csrf != "" {
		r.Header.Set("X-CSRF-Token", csrf)
	}
	for _, c := range cookies {
		r.AddCookie(c)
	}
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)
	return w
}
func token(e *echo.Echo, cookies []*http.Cookie) (string, []*http.Cookie, error) {
	w := perform(e, "GET", "/v1/auth/csrf", "", boOrigin, "", cookies)
	if w.Code != 200 {
		return "", nil, fmt.Errorf("csrf status %d: %s", w.Code, w.Body.String())
	}
	var body struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		return "", nil, err
	}
	updated := append([]*http.Cookie{}, cookies...)
	for _, fresh := range w.Result().Cookies() {
		filtered := updated[:0]
		for _, old := range updated {
			if old.Name != fresh.Name {
				filtered = append(filtered, old)
			}
		}
		updated = append(filtered, fresh)
	}
	return body.Token, updated, nil
}
func loginHTTP(e *echo.Echo, email string) ([]*http.Cookie, string, error) {
	csrf, cookies, err := token(e, nil)
	if err != nil {
		return nil, "", err
	}
	data, _ := json.Marshal(map[string]string{"email": email, "password": fixturePassword})
	w := perform(e, "POST", "/v1/auth/login", string(data), boOrigin, csrf, cookies)
	if w.Code != 200 {
		return nil, "", fmt.Errorf("login status %d: %s", w.Code, w.Body.String())
	}
	var sessions []*http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.MaxAge > 0 {
			sessions = append(sessions, c)
		}
	}
	csrf, sessions, err = token(e, sessions)
	return sessions, csrf, err
}

func TestBOHTTPActualFlowAndRevocation(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		e, b := httpRouter(p, svc, func() time.Time { return fixtureNow })
		bad := perform(e, "POST", "/v1/auth/login", `{"email":"root@example.test","password":"isolated fixture password 123"}`, "https://evil.example.test", "", nil)
		if bad.Code != 403 {
			return errors.New("foreign origin login allowed")
		}
		var n int
		if err = p.QueryRow(ctx, `SELECT count(*) FROM platform_staff_sessions`).Scan(&n); err != nil {
			return err
		}
		if n != 0 {
			return errors.New("denied login persisted session")
		}
		cookies, csrf, err := loginHTTP(e, root.Email)
		if err != nil {
			return err
		}
		for _, c := range cookies {
			if !c.HttpOnly || !c.Secure || c.Path != "/" || c.Domain != "" || c.SameSite != http.SameSiteStrictMode {
				return errors.New("unsafe cookie flags")
			}
			if c.Name == b.SessionName() && (c.MaxAge != 2592000 || !c.Expires.Equal(fixtureNow.Add(staff.SessionLifetime))) {
				return errors.New("incorrect cookie expiry")
			}
		}
		me := perform(e, "GET", "/v1/auth/me", "", boOrigin, "", cookies)
		if me.Code != 200 || !strings.Contains(me.Body.String(), root.ID.String()) || me.Header().Get("Cache-Control") != "no-store" || !strings.Contains(me.Header().Get("Vary"), "Origin") || me.Header().Get("Access-Control-Allow-Origin") != boOrigin {
			return fmt.Errorf("me boundary failed: %s", me.Body.String())
		}
		for _, secret := range []string{fixturePassword, "passwordHash", "tokenHash", "sessionHash", "argon2"} {
			if strings.Contains(me.Body.String(), secret) {
				return errors.New("secret in staff DTO")
			}
		}
		w := perform(e, "POST", "/v1/auth/password", `{"currentPassword":"isolated fixture password 123","newPassword":"isolated replacement password 456"}`, boOrigin, csrf, cookies)
		if w.Code != 204 {
			return fmt.Errorf("password change status %d: %s", w.Code, w.Body.String())
		}
		if len(w.Result().Cookies()) != 2 {
			return errors.New("password change did not clear both cookies")
		}
		old := perform(e, "GET", "/v1/auth/me", "", boOrigin, "", cookies)
		if old.Code != 401 {
			return errors.New("old session accepted after HTTP password change")
		}
		return nil
	})
}
