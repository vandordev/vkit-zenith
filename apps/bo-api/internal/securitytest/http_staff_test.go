//go:build integration

package securitytest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	v1 "github.com/vandordev/vkit-zenith/apps/bo-api/internal/handlers/v1"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func TestBOHTTPStaffAuthorityAndRevision(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		e, _ := httpRouter(p, svc, func() time.Time { return fixtureNow })
		cookies, csrf, err := loginHTTP(e, root.Email)
		if err != nil {
			return err
		}
		body := map[string]string{"email": "admin@example.test", "displayName": "Fixture Admin", "role": "admin", "status": "active", "password": fixturePassword, "currentPassword": replacementPassword}
		raw, _ := json.Marshal(body)
		if w := perform(e, "POST", "/v1/staff", string(raw), boOrigin, csrf, cookies); w.Code != 403 {
			return errors.New("wrong current password allowed")
		}
		body["currentPassword"] = fixturePassword
		raw, _ = json.Marshal(body)
		w := perform(e, "POST", "/v1/staff", string(raw), boOrigin, csrf, cookies)
		if w.Code != 201 {
			return fmt.Errorf("create staff %d %s", w.Code, w.Body.String())
		}
		var target v1.V1StaffCreateBody
		if err = json.Unmarshal(w.Body.Bytes(), &target); err != nil {
			return err
		}
		admin, adminCSRF, err := loginHTTP(e, target.Email)
		if err != nil {
			return err
		}
		for _, path := range []string{"/v1/staff", "/v1/staff/" + target.ID} {
			if w = perform(e, "GET", path, "", boOrigin, "", admin); w.Code != 403 {
				return errors.New("admin staff read allowed")
			}
		}
		body["email"] = "denied@example.test"
		raw, _ = json.Marshal(body)
		if w = perform(e, "POST", "/v1/staff", string(raw), boOrigin, adminCSRF, admin); w.Code != 403 {
			return errors.New("admin staff creation allowed")
		}
		update := map[string]string{"email": target.Email, "displayName": "Edited", "role": "superadmin", "status": "active", "currentPassword": fixturePassword, "revision": target.Revision}
		raw, _ = json.Marshal(update)
		if w = perform(e, "PATCH", "/v1/staff/"+target.ID, string(raw), boOrigin, adminCSRF, admin); w.Code != 403 {
			return errors.New("admin self-escalation allowed")
		}
		if w = perform(e, "PATCH", "/v1/staff/"+target.ID, string(raw), boOrigin, csrf, cookies); w.Code != 200 {
			return fmt.Errorf("superadmin edit %d %s", w.Code, w.Body.String())
		}
		if w = perform(e, "PATCH", "/v1/staff/"+target.ID, string(raw), boOrigin, csrf, cookies); w.Code != 409 {
			return errors.New("stale revision accepted")
		}
		if w = perform(e, "GET", "/v1/auth/me", "", boOrigin, "", admin); w.Code != 401 {
			return errors.New("role change retained HTTP session")
		}
		var n int
		if err = p.QueryRow(ctx, `SELECT count(*) FROM platform_staff`).Scan(&n); err != nil {
			return err
		}
		if n != 2 {
			return errors.New("denied staff operation committed")
		}
		return nil
	})
}

func TestBOHTTPAuditFailureAndExpiry(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		root, err := svc.Bootstrap(ctx, bootstrapInput("root@example.test"))
		if err != nil {
			return err
		}
		e, _ := httpRouter(p, svc, func() time.Time { return fixtureNow })
		cookies, csrf, err := loginHTTP(e, root.Email)
		if err != nil {
			return err
		}
		if err = failAudits(ctx, p); err != nil {
			return err
		}
		body := fmt.Sprintf(`{"email":"failed@example.test","displayName":"Failure","role":"admin","status":"active","password":%q,"currentPassword":%q}`, fixturePassword, fixturePassword)
		if w := perform(e, "POST", "/v1/staff", body, boOrigin, csrf, cookies); w.Code != 503 {
			return fmt.Errorf("audit fault %d", w.Code)
		}
		var n int
		if err = p.QueryRow(ctx, `SELECT count(*) FROM platform_staff`).Scan(&n); err != nil {
			return err
		}
		if n != 1 {
			return errors.New("HTTP audit fault partially committed")
		}
		expired := staff.NewService(p, func() time.Time { return fixtureNow.Add(720 * time.Hour) }, staff.NewPasswords())
		e, _ = httpRouter(p, expired, func() time.Time { return fixtureNow.Add(720 * time.Hour) })
		if w := perform(e, "GET", "/v1/auth/me", "", boOrigin, "", cookies); w.Code != 401 {
			return errors.New("exact expiry accepted through HTTP")
		}
		return nil
	})
}
