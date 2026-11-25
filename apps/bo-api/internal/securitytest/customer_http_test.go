//go:build integration

package securitytest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func customerJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// Catches omitted authority/confirmation/CSRF checks or privileged profile extras
// on every mounted mutation, and verifies denied requests leave persisted state.
func TestBOCustomerHTTPAuthorityAndForgedBodies(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		c, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		org, err := c.CreateOrganization(ctx, a, customers.OrganizationCreate{Name: "One", Slug: "one", NewOwner: &customers.UserCreate{Email: "owner@example.test", DisplayName: "Owner"}})
		if err != nil {
			return err
		}
		user, err := c.CreateUser(ctx, a, customers.UserCreate{Email: "user@example.test", DisplayName: "User"})
		if err != nil {
			return err
		}
		member, err := c.AddMember(ctx, a, org.Organization.ID, user.ID, customers.RoleMember)
		if err != nil {
			return err
		}
		adminIn := bootstrapInput("admin-http@example.test")
		adminIn.Role = "admin"
		adminIn.CurrentPassword = fixturePassword
		admin, err := svc.CreateStaff(ctx, a, adminIn)
		if err != nil {
			return err
		}
		e, _ := customerRouter(p, svc, func() time.Time { return fixtureNow })
		cookies, csrf, err := loginHTTP(e, "root@example.test")
		if err != nil {
			return err
		}
		adminCookies, adminCSRF, err := loginHTTP(e, admin.Email)
		if err != nil {
			return err
		}
		base := "/v1/organizations/" + org.Organization.ID.String()
		u := "/v1/users/" + user.ID.String()
		m := base + "/memberships/" + member.ID.String()
		revision := customers.UserRevision(user)
		mr := customers.MembershipRevision(member)
		or := customers.OrganizationRevision(org.Organization)
		sensitiveBody := func(rev string, extra map[string]any) map[string]any {
			out := map[string]any{"revision": rev, "reason": "Fixture maintenance", "currentPassword": fixturePassword}
			for k, v := range extra {
				out[k] = v
			}
			return out
		}
		requests := []struct {
			method, path string
			body         any
			sensitive    bool
		}{
			{"POST", "/v1/users", map[string]any{"email": "create@example.test", "displayName": "Created"}, false},
			{"PATCH", u, map[string]any{"displayName": "Changed", "revision": revision}, false},
			{"POST", u + "/status", sensitiveBody(revision, map[string]any{"status": "disabled"}), true},
			{"POST", u + "/email", sensitiveBody(revision, map[string]any{"email": "changed@example.test"}), true},
			{"POST", u + "/email/verify", sensitiveBody(revision, nil), true},
			{"POST", "/v1/organizations", map[string]any{"name": "New", "slug": "new", "ownerUserId": org.User.ID.String()}, false},
			{"PATCH", base, map[string]any{"name": "Changed", "slug": "changed", "revision": or}, false},
			{"POST", base + "/status", sensitiveBody(or, map[string]any{"status": "suspended"}), true},
			{"POST", base + "/memberships", map[string]any{"userId": org.User.ID.String()}, false},
			{"POST", m + "/remove", map[string]any{"revision": mr}, false},
			{"POST", m + "/restore", map[string]any{"revision": mr}, false},
			{"POST", m + "/role", sensitiveBody(mr, map[string]any{"role": "owner"}), true},
		}
		snapshot := func() (string, error) {
			var s string
			err := p.QueryRow(ctx, `SELECT jsonb_build_array((SELECT jsonb_agg(to_jsonb(u) ORDER BY id) FROM users u),(SELECT jsonb_agg(to_jsonb(o) ORDER BY id) FROM organizations o),(SELECT jsonb_agg(to_jsonb(m) ORDER BY id) FROM organization_memberships m),(SELECT count(*) FROM audit_events WHERE action LIKE 'customer.%'))::text`).Scan(&s)
			return s, err
		}
		before, err := snapshot()
		if err != nil {
			return err
		}
		for _, r := range requests {
			body := customerJSON(r.body)
			for _, denied := range []struct {
				cookies      []*http.Cookie
				csrf, origin string
				want         int
			}{{nil, "", boOrigin, 401}, {cookies, "", boOrigin, 403}, {cookies, csrf, "https://evil.example.test", 403}} {
				w := perform(e, r.method, r.path, body, denied.origin, denied.csrf, denied.cookies)
				if w.Code != denied.want {
					return fmt.Errorf("%s %s denied status %d want %d: %s", r.method, r.path, w.Code, denied.want, w.Body.String())
				}
			}
			if r.sensitive {
				w := perform(e, r.method, r.path, body, boOrigin, adminCSRF, adminCookies)
				if w.Code != 403 {
					return fmt.Errorf("Admin sensitive %s: %d %s", r.path, w.Code, w.Body.String())
				}
				missing := map[string]any{}
				if err := json.Unmarshal([]byte(body), &missing); err != nil {
					return err
				}
				delete(missing, "currentPassword")
				w = perform(e, r.method, r.path, customerJSON(missing), boOrigin, csrf, cookies)
				if w.Code != 422 {
					return fmt.Errorf("missing confirmation %s: %d", r.path, w.Code)
				}
			}
		}
		for _, test := range []struct {
			method, path string
			body         any
			want         int
		}{
			{"PATCH", u, map[string]any{"displayName": "Forged", "revision": revision, "status": "disabled", "email": "forged@example.test"}, 422},
			{"POST", "/v1/users", map[string]any{"email": "forged@example.test", "displayName": "Forged", "emailVerifiedAt": fixtureNow, "password": fixturePassword}, 422},
			{"POST", base + "/memberships", map[string]any{"userId": user.ID.String(), "role": "owner"}, 422},
			{"POST", u + "/status", sensitiveBody(revision, map[string]any{"status": "disabled", "currentPassword": "not the fixture password"}), 403},
			{"POST", u + "/email", sensitiveBody(revision, map[string]any{"email": "new@example.test", "reason": ""}), 422},
			{"POST", u + "/email/verify", sensitiveBody(revision, map[string]any{"reason": strings.Repeat("界", 1025)}), 422},
			{"PATCH", u, map[string]any{"displayName": "Stale", "revision": strings.Repeat("0", 64)}, 409},
		} {
			w := perform(e, test.method, test.path, customerJSON(test.body), boOrigin, csrf, cookies)
			if w.Code != test.want {
				return fmt.Errorf("rejected %s status %d want %d: %s", test.path, w.Code, test.want, w.Body.String())
			}
			if strings.Contains(w.Body.String(), fixturePassword) || strings.Contains(w.Body.String(), "not the fixture password") {
				return fmt.Errorf("secret reflected in error")
			}
		}
		after, err := snapshot()
		if err != nil {
			return err
		}
		if before != after {
			return fmt.Errorf("denied/forged request changed customer state or audit")
		}
		return nil
	})
}

func TestBOCustomerHTTPCommandsAndRevisions(t *testing.T) {
	governanceDatabase(t, func(ctx context.Context, p *pgxpool.Pool, svc *staff.Service) error {
		_, a, err := customerCommands(ctx, p, svc)
		if err != nil {
			return err
		}
		in := bootstrapInput("admin-commands@example.test")
		in.Role = "admin"
		in.CurrentPassword = fixturePassword
		admin, err := svc.CreateStaff(ctx, a, in)
		if err != nil {
			return err
		}
		e, _ := customerRouter(p, svc, func() time.Time { return fixtureNow })
		rootCookies, rootCSRF, err := loginHTTP(e, "root@example.test")
		if err != nil {
			return err
		}
		adminCookies, adminCSRF, err := loginHTTP(e, admin.Email)
		if err != nil {
			return err
		}
		call := func(method, path string, body any, super bool, want int) (map[string]any, error) {
			cookies, csrf := adminCookies, adminCSRF
			if super {
				cookies, csrf = rootCookies, rootCSRF
			}
			w := perform(e, method, path, customerJSON(body), boOrigin, csrf, cookies)
			if w.Code != want {
				return nil, fmt.Errorf("%s %s status %d want %d: %s", method, path, w.Code, want, w.Body.String())
			}
			if w.Header().Get("Cache-Control") != "no-store" {
				return nil, fmt.Errorf("cacheable mutation")
			}
			var out map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
				return nil, err
			}
			return out, nil
		}
		user, err := call("POST", "/v1/users", map[string]any{"email": "member@example.test", "displayName": "Member"}, false, 201)
		if err != nil {
			return err
		}
		uid, ok := user["id"].(string)
		if !ok {
			return fmt.Errorf("missing user id")
		}
		upath := "/v1/users/" + uid
		rev, ok := user["revision"].(string)
		if !ok {
			return fmt.Errorf("missing revision")
		}
		user, err = call("PATCH", upath, map[string]any{"displayName": "Edited", "revision": rev}, false, 200)
		if err != nil {
			return err
		}
		sensitiveBody := func(entity map[string]any, extra map[string]any) map[string]any {
			out := map[string]any{"revision": entity["revision"], "reason": "Isolated HTTP maintenance", "currentPassword": fixturePassword}
			for k, v := range extra {
				out[k] = v
			}
			return out
		}
		user, err = call("POST", upath+"/email/verify", sensitiveBody(user, nil), true, 200)
		if err != nil {
			return err
		}
		if user["emailVerifiedAt"] == nil {
			return fmt.Errorf("manual verification absent")
		}
		old := user
		user, err = call("POST", upath+"/email", sensitiveBody(user, map[string]any{"email": "new@example.test"}), true, 200)
		if err != nil {
			return err
		}
		if user["id"] != uid || user["emailVerifiedAt"] != nil {
			return fmt.Errorf("email identity/reset violation")
		}
		if _, err = call("POST", upath+"/email/verify", sensitiveBody(old, nil), true, 409); err != nil {
			return err
		}
		user, err = call("POST", upath+"/status", sensitiveBody(user, map[string]any{"status": "disabled"}), true, 200)
		if err != nil {
			return err
		}
		user, err = call("POST", upath+"/status", sensitiveBody(user, map[string]any{"status": "active"}), true, 200)
		if err != nil {
			return err
		}
		first, err := call("POST", "/v1/organizations", map[string]any{"name": "First", "slug": "first", "newOwner": map[string]any{"email": "owner@example.test", "displayName": "Owner"}}, false, 201)
		if err != nil {
			return err
		}
		oid, ok := first["id"].(string)
		if !ok {
			return fmt.Errorf("missing organization id")
		}
		opath := "/v1/organizations/" + oid
		ownerID, ok := first["ownerUserId"].(string)
		if !ok {
			return fmt.Errorf("missing initial owner user id")
		}
		second, err := call("POST", "/v1/organizations", map[string]any{"name": "Second", "slug": "second", "ownerUserId": ownerID}, false, 201)
		if err != nil {
			return err
		}
		org, err := call("PATCH", opath, map[string]any{"name": "Edited org", "slug": "edited-org", "revision": first["revision"]}, false, 200)
		if err != nil {
			return err
		}
		org, err = call("POST", opath+"/status", sensitiveBody(org, map[string]any{"status": "suspended"}), true, 200)
		if err != nil {
			return err
		}
		_, err = call("POST", opath+"/status", sensitiveBody(org, map[string]any{"status": "active"}), true, 200)
		if err != nil {
			return err
		}
		member, err := call("POST", opath+"/memberships", map[string]any{"userId": uid, "role": "member"}, false, 201)
		if err != nil {
			return err
		}
		mid, ok := member["id"].(string)
		if !ok {
			return fmt.Errorf("missing member id")
		}
		mpath := opath + "/memberships/" + mid
		if _, err = call("POST", "/v1/organizations/"+second["id"].(string)+"/memberships/"+mid+"/remove", map[string]any{"revision": member["revision"]}, false, 404); err != nil {
			return err
		}
		member, err = call("POST", mpath+"/remove", map[string]any{"revision": member["revision"]}, false, 200)
		if err != nil {
			return err
		}
		member, err = call("POST", mpath+"/restore", map[string]any{"revision": member["revision"]}, false, 200)
		if err != nil {
			return err
		}
		if member["id"] != mid || member["role"] != "member" {
			return fmt.Errorf("restore identity/owner violation")
		}
		member, err = call("POST", mpath+"/role", sensitiveBody(member, map[string]any{"role": "owner"}), true, 200)
		if err != nil {
			return err
		}
		if _, err = call("POST", mpath+"/remove", map[string]any{"revision": member["revision"]}, false, 409); err != nil {
			return err
		}
		_, err = call("POST", mpath+"/role", sensitiveBody(member, map[string]any{"role": "member"}), true, 200)
		return err
	})
}
