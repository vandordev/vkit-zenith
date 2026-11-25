//go:build integration

package securitytest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/orgaccess"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
)

func TestWebAPIListMyOrganizations(t *testing.T) {
	r := newOrgRecorder(t, orgFixture)
	response := r.request(http.MethodGet, "/v1/organizations", orgTestOrigin, "")
	if response.Code != http.StatusOK || response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("unexpected response: %d", response.Code)
	}
	var body struct {
		Items []struct {
			ID          string   `json:"id"`
			Role        string   `json:"role"`
			Permissions []string `json:"permissions"`
			HasLogo     bool     `json:"hasLogo"`
		} `json:"items"`
	}
	if json.Unmarshal(response.Body.Bytes(), &body) != nil || len(body.Items) != 1 || body.Items[0].ID != r.orgID || body.Items[0].Role != "owner" || body.Items[0].HasLogo || strings.Join(body.Items[0].Permissions, ",") != "developers.manage,members.manage,organization.manage,organization.read" {
		t.Fatal("membership list contract mismatch")
	}
}

func TestWebAPIRemovedMembership(t *testing.T) {
	r := newOrgRecorder(t, orgFixture)
	ownerID := uuid.NewString()
	if _, err := r.pool.Exec(context.Background(), `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at) VALUES($1,$2,$2,'Second Owner','active',now(),now())`, ownerID, "owner-"+ownerID+"@example.test"); err != nil {
		t.Fatal("owner fixture seed failed")
	}
	if _, err := r.pool.Exec(context.Background(), `INSERT INTO organization_memberships(id,organization_id,user_id,status,role,created_at,updated_at) VALUES($1,$2,$3,'active','owner',now(),now())`, uuid.NewString(), r.orgID, ownerID); err != nil {
		t.Fatal("owner membership seed failed")
	}
	if _, err := r.pool.Exec(context.Background(), `UPDATE organization_memberships SET role='member' WHERE organization_id=$1 AND user_id=$2`, r.orgID, r.userID); err != nil {
		t.Fatal("membership fixture role update failed")
	}
	_, err := r.pool.Exec(context.Background(), `UPDATE organization_memberships SET status='removed' WHERE organization_id=$1 AND user_id=$2`, r.orgID, r.userID)
	if err != nil {
		t.Fatal("membership fixture update failed")
	}
	removed := r.request(http.MethodGet, "/v1/organizations/"+r.orgID+"/profile", orgTestOrigin, "")
	unknown := r.request(http.MethodGet, "/v1/organizations/"+uuid.NewString()+"/profile", orgTestOrigin, "")
	if removed.Code != http.StatusNotFound || removed.Body.String() != unknown.Body.String() {
		t.Fatal("removed membership did not match unknown organization")
	}
}

func TestWebAPIInactiveUser(t *testing.T) {
	r := newOrgRecorder(t, orgFixture)
	ownerID := uuid.NewString()
	if _, err := r.pool.Exec(context.Background(), `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at) VALUES($1,$2,$2,'Second Owner','active',now(),now())`, ownerID, "owner-"+ownerID+"@example.test"); err != nil {
		t.Fatal("owner fixture seed failed")
	}
	if _, err := r.pool.Exec(context.Background(), `INSERT INTO organization_memberships(id,organization_id,user_id,status,role,created_at,updated_at) VALUES($1,$2,$3,'active','owner',now(),now())`, uuid.NewString(), r.orgID, ownerID); err != nil {
		t.Fatal("owner membership seed failed")
	}
	if _, err := r.pool.Exec(context.Background(), `UPDATE organization_memberships SET role='member' WHERE organization_id=$1 AND user_id=$2`, r.orgID, r.userID); err != nil {
		t.Fatal("membership role fixture update failed")
	}
	if _, err := r.pool.Exec(context.Background(), `UPDATE users SET status='disabled' WHERE id=$1`, r.userID); err != nil {
		t.Fatal("user fixture update failed")
	}
	if response := r.request(http.MethodGet, "/v1/organizations", orgTestOrigin, ""); response.Code != http.StatusUnauthorized {
		t.Fatalf("inactive user returned %d", response.Code)
	}
}

func TestWebAPIRestrictedSession(t *testing.T) {
	r := newOrgRecorder(t, orgFixture)
	if _, err := r.pool.Exec(context.Background(), `UPDATE auth_sessions SET kind='restricted',last_activity_at=authenticated_at,expires_at=authenticated_at+interval '30 minutes',browser_cookie_expires_at=authenticated_at+interval '30 minutes' WHERE user_id=$1`, r.userID); err != nil {
		t.Fatal("restricted session fixture update failed")
	}
	if response := r.request(http.MethodGet, "/v1/organizations", orgTestOrigin, ""); response.Code != http.StatusUnauthorized {
		t.Fatalf("restricted session returned %d", response.Code)
	}
}

func TestWebAPIForeignOrganization(t *testing.T) {
	r := newOrgRecorder(t, orgFixture)
	removed := uuid.NewString()
	foreignUser := uuid.NewString()
	if _, err := r.pool.Exec(context.Background(), `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at) VALUES($1,$2,$2,'Foreign Owner','active',now(),now())`, foreignUser, "foreign-"+foreignUser+"@example.test"); err != nil {
		t.Fatal("foreign user fixture failed")
	}
	tx, err := r.pool.Begin(context.Background())
	if err != nil {
		t.Fatal("foreign organization transaction unavailable")
	}
	if _, err = tx.Exec(context.Background(), `INSERT INTO organizations(id,name,slug,status,created_at,updated_at) VALUES($1,'Foreign Org',$2,'active',now(),now())`, removed, "foreign-"+strings.ReplaceAll(removed, "-", "")); err == nil {
		_, err = tx.Exec(context.Background(), `INSERT INTO organization_memberships(id,organization_id,user_id,status,role,created_at,updated_at) VALUES($1,$2,$3,'active','owner',now(),now())`, uuid.NewString(), removed, foreignUser)
	}
	if err == nil {
		err = tx.Commit(context.Background())
	} else {
		_ = tx.Rollback(context.Background())
	}
	if err != nil {
		t.Fatal("foreign organization fixture failed")
	}
	foreign := r.request(http.MethodGet, "/v1/organizations/"+removed+"/profile", orgTestOrigin, "")
	unknown := r.request(http.MethodGet, "/v1/organizations/"+uuid.NewString()+"/profile", orgTestOrigin, "")
	if foreign.Code != http.StatusNotFound || foreign.Body.String() != unknown.Body.String() {
		t.Fatal("foreign organization did not match unknown organization")
	}
}

func TestWebAPIMalformedOrganizationID(t *testing.T) {
	r := newOrgRecorder(t, orgFixture)
	malformed := r.request(http.MethodGet, "/v1/organizations/not-a-uuid/profile", orgTestOrigin, "")
	unknown := r.request(http.MethodGet, "/v1/organizations/"+uuid.NewString()+"/profile", orgTestOrigin, "")
	if malformed.Code != http.StatusNotFound || malformed.Body.String() != unknown.Body.String() {
		t.Fatalf("malformed organization id returned %d", malformed.Code)
	}
	for _, boundary := range []*orgaccess.Boundary{orgaccess.NewBoundary(nil, nil)} {
		assertUnavailable(t, func() error { _, err := boundary.Subject(context.Background()); return err }())
		assertUnavailable(t, func() error {
			_, err := boundary.Read(context.Background(), "bad-id", customers.PermissionOrganizationRead)
			return err
		}())
	}
	assertUnavailable(t, func() error { _, err := orgaccess.NewBoundary(r.runtime, nil).Subject(context.Background()); return err }())
}

func assertUnavailable(t *testing.T, err error) {
	t.Helper()
	var status huma.StatusError
	if !errors.As(err, &status) || status.GetStatus() != http.StatusServiceUnavailable {
		t.Fatal("missing boundary dependency was not unavailable")
	}
}

func TestWebAPIPermissionDenials(t *testing.T) {
	r := newOrgRecorder(t, orgFixture)
	ownerID := uuid.NewString()
	if _, err := r.pool.Exec(context.Background(), `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at) VALUES($1,$2,$2,'Second Owner','active',now(),now())`, ownerID, "owner-"+ownerID+"@example.test"); err != nil {
		t.Fatal("owner fixture seed failed")
	}
	if _, err := r.pool.Exec(context.Background(), `INSERT INTO organization_memberships(id,organization_id,user_id,status,role,created_at,updated_at) VALUES($1,$2,$3,'active','owner',now(),now())`, uuid.NewString(), r.orgID, ownerID); err != nil {
		t.Fatal("owner membership seed failed")
	}
	if _, err := r.pool.Exec(context.Background(), `UPDATE organization_memberships SET role='member' WHERE organization_id=$1 AND user_id=$2`, r.orgID, r.userID); err != nil {
		t.Fatal("member role fixture update failed")
	}
	for _, permission := range []string{"developers", "manage", "members"} {
		response := r.request(http.MethodGet, "/v1/organizations/"+r.orgID+"/permission/"+permission, orgTestOrigin, "")
		if response.Code != http.StatusForbidden {
			t.Errorf("member %s denied incorrectly: %d", permission, response.Code)
		}
	}
	if _, err := r.pool.Exec(context.Background(), `UPDATE organization_memberships SET role='admin' WHERE organization_id=$1 AND user_id=$2`, r.orgID, r.userID); err != nil {
		t.Fatal("admin role fixture update failed")
	}
	if response := r.request(http.MethodGet, "/v1/organizations/"+r.orgID+"/permission/manage", orgTestOrigin, ""); response.Code != http.StatusForbidden {
		t.Errorf("admin organization manage permission returned %d", response.Code)
	}
}

func TestWebAPISuspendedWrite(t *testing.T) {
	r := newOrgRecorder(t, orgFixture)
	if _, err := r.pool.Exec(context.Background(), `UPDATE organizations SET status='suspended' WHERE id=$1`, r.orgID); err != nil {
		t.Fatal("suspended organization fixture failed")
	}
	response := r.request(http.MethodPost, "/v1/organizations/"+r.orgID+"/write", orgTestOrigin, r.csrf)
	if response.Code != http.StatusConflict {
		t.Fatalf("suspended write returned %d", response.Code)
	}
}

func TestWebAPIWriteOriginCSRF(t *testing.T) {
	r := newOrgRecorder(t, orgFixture)
	for _, tc := range []struct{ name, origin, csrf string }{
		{"missing-origin", "", r.csrf}, {"foreign-origin", "https://foreign.example.test", r.csrf},
		{"missing-csrf", orgTestOrigin, ""}, {"invalid-csrf", orgTestOrigin, "not-a-token"},
		{"other-session-csrf", orgTestOrigin, func() string { token, _ := r.service.NewSessionCSRF([32]byte{9}); return token }()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			response := r.request(http.MethodPost, "/v1/organizations/"+r.orgID+"/write", tc.origin, tc.csrf)
			if response.Code != http.StatusForbidden {
				t.Fatalf("invalid write binding returned %d", response.Code)
			}
		})
	}
}

func TestWebAPICORSNoStore(t *testing.T) {
	r := newOrgRecorder(t, orgFixture)
	response := r.request(http.MethodGet, "/v1/organizations/"+r.orgID+"/profile", orgTestOrigin, "")
	if response.Code != http.StatusOK || response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("Access-Control-Allow-Origin") != orgTestOrigin || response.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Fatal("v1 response missing CORS/no-store policy")
	}
	var profile struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		HasLogo bool   `json:"hasLogo"`
	}
	if json.Unmarshal(response.Body.Bytes(), &profile) != nil || profile.ID != r.orgID || profile.Name != "Organization Example" || profile.HasLogo {
		t.Fatal("profile response contract mismatch")
	}
	preflight := r.request(http.MethodOptions, "/v1/organizations/"+r.orgID+"/profile", orgTestOrigin, "")
	if preflight.Code != http.StatusNoContent || preflight.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT, PATCH, DELETE, OPTIONS" {
		t.Fatal("v1 preflight policy mismatch")
	}
	foreign := r.request(http.MethodGet, "/v1/organizations/"+r.orgID+"/profile", "https://foreign.example.test", "")
	if foreign.Code != http.StatusForbidden {
		t.Fatalf("foreign origin returned %d", foreign.Code)
	}
	if response := r.requestBody(http.MethodPost, "/v1/auth/body-policy", "", "", `"`+strings.Repeat("x", 16*1024)+`"`); response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("Auth limit returned %d", response.Code)
	}
	if response := r.requestBody(http.MethodPost, "/v1/body-policy", "", "", `"`+strings.Repeat("x", 16*1024)+`"`); response.Code != http.StatusOK {
		t.Fatalf("organization request below its limit returned %d", response.Code)
	}
	exactBody := `"` + strings.Repeat("x", 64*1024-2) + `"`
	overBody := `"` + strings.Repeat("x", 64*1024-1) + `"`
	for _, tc := range []struct {
		name    string
		request func(string) *httptest.ResponseRecorder
	}{
		{"body-operation-content-length", func(body string) *httptest.ResponseRecorder {
			return r.requestBody(http.MethodPost, "/v1/body-policy", orgTestOrigin, "", body)
		}},
		{"body-operation-stream", func(body string) *httptest.ResponseRecorder {
			return r.requestStream(http.MethodPost, "/v1/body-policy", orgTestOrigin, body)
		}},
		{"bodyless-profile-content-length", func(body string) *httptest.ResponseRecorder {
			return r.requestBody(http.MethodGet, "/v1/organizations/"+r.orgID+"/profile", orgTestOrigin, "", body)
		}},
		{"bodyless-profile-stream", func(body string) *httptest.ResponseRecorder {
			return r.requestStream(http.MethodGet, "/v1/organizations/"+r.orgID+"/profile", orgTestOrigin, body)
		}},
	} {
		if response := tc.request(exactBody); response.Code != http.StatusOK {
			t.Errorf("%s exact 65536 bytes returned %d", tc.name, response.Code)
		}
		if response := tc.request(overBody); response.Code != http.StatusRequestEntityTooLarge || response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("Access-Control-Allow-Origin") != orgTestOrigin {
			t.Errorf("%s 65537 bytes missed 413/CORS/no-store: %d", tc.name, response.Code)
		}
	}
	for _, tc := range []struct {
		name, method, path string
		max                time.Duration
	}{
		{"auth-default", http.MethodPost, "/v1/auth/body-policy", 5 * time.Second},
		{"dashboard-default", http.MethodPost, "/v1/body-policy", 5 * time.Second},
		{"bodyless-default", http.MethodGet, "/v1/organizations/" + r.orgID + "/profile", 5 * time.Second},
		{"dashboard-short-operation", http.MethodPost, "/v1/body-policy-short", 250 * time.Millisecond},
	} {
		response, deadlines := r.requestDeadlineTimeout(tc.method, tc.path, orgTestOrigin)
		if response.Code != http.StatusRequestTimeout || response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("Access-Control-Allow-Origin") != orgTestOrigin || !strings.Contains(response.Body.String(), "Request Timeout") || strings.Contains(response.Body.String(), "deadline was not installed") {
			t.Errorf("%s timeout response missing safe 408/CORS/no-store: %d %s", tc.name, response.Code, response.Body.String())
		}
		if len(deadlines) < 2 || deadlines[0].IsZero() || !deadlines[len(deadlines)-1].IsZero() {
			t.Errorf("%s deadline was not installed and cleared: %v", tc.name, deadlines)
		} else {
			remaining := time.Until(deadlines[0])
			if remaining <= 0 || remaining > tc.max || (tc.max == 5*time.Second && remaining < 4*time.Second) {
				t.Errorf("%s deadline exceeded its operation timeout: %s", tc.name, remaining)
			}
		}
	}
	response = r.requestBody(http.MethodPost, "/v1/body-policy", orgTestOrigin, "", `"`+strings.Repeat("x", 1024*1024+1)+`"`)
	if response.Code != http.StatusRequestEntityTooLarge || response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("Access-Control-Allow-Origin") != orgTestOrigin {
		t.Fatal("global body limit error missed CORS/no-store policy")
	}
	secret := "fake-sensitive-organization-value"
	response = r.requestBody(http.MethodPost, "/v1/organizations/"+r.orgID+"/validation", orgTestOrigin, "", `{"password":"fake-sensitive-organization-value"}`)
	if response.Code != http.StatusUnprocessableEntity || strings.Contains(response.Body.String(), secret) || !strings.Contains(response.Body.String(), "body.password") || !strings.Contains(response.Body.String(), "Isian tidak valid.") {
		t.Fatal("validation response leaked or lost its field location")
	}
	secret = "fake-malformed-sensitive-value"
	response = r.requestBody(http.MethodPost, "/v1/organizations/"+r.orgID+"/validation", orgTestOrigin, "", `{"password":"`+secret+`",`)
	if response.Code != http.StatusBadRequest || strings.Contains(response.Body.String(), secret) || !strings.Contains(response.Body.String(), `"location":"body"`) || !strings.Contains(response.Body.String(), "Isian tidak valid.") {
		t.Fatal("malformed request response leaked input or lost its safe error")
	}
	secret = `{"password":"fake-unsupported-sensitive-value"}`
	response = r.requestBodyContentType(http.MethodPost, "/v1/organizations/"+r.orgID+"/validation", orgTestOrigin, "", secret, "text/plain")
	if response.Code != http.StatusUnsupportedMediaType || strings.Contains(response.Body.String(), secret) || !strings.Contains(response.Body.String(), `"location":"body"`) || !strings.Contains(response.Body.String(), "Isian tidak valid.") {
		t.Fatal("unsupported request response leaked input or lost its safe location")
	}
}

func TestWebAPIMembersScopedAndNoSecrets(t *testing.T) {
	r := newOrgRecorder(t, orgFixture)
	otherUser, otherOrg := uuid.NewString(), uuid.NewString()
	otherEmail := "other-" + otherUser + "@example.test"
	_, err := r.pool.Exec(context.Background(), `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at) VALUES($1,$2,$2,'Other Org Member','active',now(),now())`, otherUser, otherEmail)
	if err != nil {
		t.Fatal("foreign user fixture failed")
	}
	tx, err := r.pool.Begin(context.Background())
	if err != nil {
		t.Fatal("foreign organization transaction unavailable")
	}
	if _, err = tx.Exec(context.Background(), `INSERT INTO organizations(id,name,slug,status,created_at,updated_at) VALUES($1,'Other Organization',$2,'active',now(),now())`, otherOrg, "other-"+strings.ReplaceAll(otherOrg, "-", "")); err == nil {
		_, err = tx.Exec(context.Background(), `INSERT INTO organization_memberships(id,organization_id,user_id,status,role,created_at,updated_at) VALUES($1,$2,$3,'active','owner',now(),now())`, uuid.NewString(), otherOrg, otherUser)
	}
	if err == nil {
		err = tx.Commit(context.Background())
	} else {
		_ = tx.Rollback(context.Background())
	}
	if err != nil {
		t.Fatal("foreign organization fixture failed")
	}
	memberID := uuid.NewString()
	_, err = r.pool.Exec(context.Background(), `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at) VALUES($1,$2,$2,'Second Member','active',now(),now())`, memberID, "second-"+memberID+"@example.test")
	if err != nil {
		t.Fatal("member fixture failed")
	}
	_, err = r.pool.Exec(context.Background(), `INSERT INTO organization_memberships(id,organization_id,user_id,status,role,created_at,updated_at) VALUES($1,$2,$3,'active','member',now(),now())`, uuid.NewString(), r.orgID, memberID)
	if err != nil {
		t.Fatal("membership fixture failed")
	}
	inactiveID, removedID := uuid.NewString(), uuid.NewString()
	for id, status := range map[string]string{inactiveID: "disabled", removedID: "active"} {
		email := "inactive-" + id + "@example.test"
		if status == "active" {
			email = "removed-" + id + "@example.test"
		}
		if _, err = r.pool.Exec(context.Background(), `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at) VALUES($1,$2,$2,'Filtered Member',$3,now(),now())`, id, email, status); err != nil {
			t.Fatal("filtered member user seed failed")
		}
		membershipStatus := "active"
		if status == "active" {
			membershipStatus = "removed"
		}
		if _, err = r.pool.Exec(context.Background(), `INSERT INTO organization_memberships(id,organization_id,user_id,status,role,created_at,updated_at) VALUES($1,$2,$3,$4,'member',now(),now())`, uuid.NewString(), r.orgID, id, membershipStatus); err != nil {
			t.Fatal("filtered member membership seed failed")
		}
	}
	response := r.request(http.MethodGet, "/v1/organizations/"+r.orgID+"/members", orgTestOrigin, "")
	if response.Code != http.StatusOK {
		t.Fatalf("member list returned %d", response.Code)
	}
	var body struct {
		Items []struct {
			Email string `json:"email"`
		} `json:"items"`
	}
	if json.Unmarshal(response.Body.Bytes(), &body) != nil || len(body.Items) != 2 || !strings.Contains(response.Body.String(), "member-"+r.userID) || strings.Contains(response.Body.String(), otherEmail) {
		t.Fatal("member list escaped tenant scope")
	}
	assertNoSecretKeys(t, response.Body.Bytes())
}
