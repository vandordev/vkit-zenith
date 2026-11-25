//go:build integration

package securitytest

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"testing"
)

func TestWebAPISecretsOnlyOnCreate(t *testing.T) {
	r := newOrgRecorder(t, orgFixture)
	for _, path := range []string{"/api-keys"} {
		resp := r.request(http.MethodGet, "/v1/organizations/"+r.orgID+path, orgTestOrigin, "")
		if resp.Code != http.StatusOK {
			t.Fatalf("GET %s returned %d", path, resp.Code)
		}
		assertNoSecretKeys(t, resp.Body.Bytes())
	}
	for _, tc := range []struct {
		path  string
		total string
	}{{"/api-keys?status=revoked&q=lama&page=1&size=1", "1"}} {
		resp := r.request(http.MethodGet, "/v1/organizations/"+r.orgID+tc.path, orgTestOrigin, "")
		var body struct {
			Items      []json.RawMessage `json:"items"`
			Pagination struct {
				Total string `json:"total"`
				Page  int    `json:"page"`
				Size  int    `json:"size"`
			} `json:"pagination"`
		}
		if resp.Code != http.StatusOK || json.Unmarshal(resp.Body.Bytes(), &body) != nil || len(body.Items) != 1 || body.Pagination.Total != tc.total || body.Pagination.Page != 1 || body.Pagination.Size != 1 {
			t.Errorf("settings list filter/pagination failed for %s", tc.path)
		}
	}
	paged := r.request(http.MethodGet, "/v1/organizations/"+r.orgID+"/api-keys?page=2&size=1", orgTestOrigin, "")
	var pageBody struct {
		Items      []json.RawMessage `json:"items"`
		Pagination struct {
			Total string `json:"total"`
			Page  int    `json:"page"`
			Size  int    `json:"size"`
		} `json:"pagination"`
	}
	if paged.Code != http.StatusOK || json.Unmarshal(paged.Body.Bytes(), &pageBody) != nil || len(pageBody.Items) != 1 || pageBody.Pagination.Total != "3" || pageBody.Pagination.Page != 2 || pageBody.Pagination.Size != 1 {
		t.Error("settings list page 2 was not applied")
	}
	{
		resp := r.requestBody(http.MethodPost, "/v1/organizations/"+r.orgID+"/api-keys", orgTestOrigin, r.csrf, `{"name":"Integrasi"}`)
		if resp.Code != http.StatusOK || resp.Header().Get("Cache-Control") != "no-store" || !strings.Contains(resp.Body.String(), "zk_DUMMYNOTAKEY") {
			t.Fatalf("API key create returned %d without expected one-time secret", resp.Code)
		}
	}
	for _, expiresAt := range []string{"2026-12-01T12:00:00Z", "2026-12-01T12:00:00+07:00", "2026-12-01T12:00:00-05:00"} {
		body, _ := json.Marshal(map[string]any{"name": "Expiry", "expiresAt": expiresAt})
		resp := r.requestBody(http.MethodPost, "/v1/organizations/"+r.orgID+"/api-keys", orgTestOrigin, r.csrf, string(body))
		if resp.Code != http.StatusOK || !strings.Contains(resp.Body.String(), expiresAt) {
			t.Errorf("RFC3339 expiry %q returned %d", expiresAt, resp.Code)
		}
	}
}

func TestWebAPIAPIKeyValidation(t *testing.T) {
	r := newOrgRecorder(t, orgFixture)
	resp := r.requestBody(http.MethodPost, "/v1/organizations/"+r.orgID+"/api-keys", orgTestOrigin, r.csrf, `{"name":"Key","expiresAt":"tomorrow"}`)
	if resp.Code != http.StatusUnprocessableEntity {
		t.Errorf("invalid API key expiry returned %d", resp.Code)
	}
}

func TestWebAPILogoOwnerOnly(t *testing.T) {
	r := newOrgRecorder(t, orgFixture)
	seedSecondOwner(t, r)
	path := "/v1/organizations/" + r.orgID + "/profile/logo"
	body := `{"contentType":"image/png","size":1024}`
	if owner := r.requestBody(http.MethodPost, path, orgTestOrigin, r.csrf, body); owner.Code != http.StatusOK {
		t.Fatalf("owner logo upload returned %d", owner.Code)
	}
	if _, err := r.pool.Exec(context.Background(), `UPDATE organization_memberships SET role='admin' WHERE organization_id=$1 AND user_id=$2`, r.orgID, r.userID); err != nil {
		t.Fatal("admin fixture update failed")
	}
	if admin := r.requestBody(http.MethodPost, path, orgTestOrigin, r.csrf, body); admin.Code != http.StatusForbidden {
		t.Fatalf("admin logo upload returned %d", admin.Code)
	}
}

func TestWebAPIAdminManagesDevelopers(t *testing.T) {
	r := newOrgRecorder(t, orgFixture)
	seedSecondOwner(t, r)
	if _, err := r.pool.Exec(context.Background(), `UPDATE organization_memberships SET role='admin' WHERE organization_id=$1 AND user_id=$2`, r.orgID, r.userID); err != nil {
		t.Fatal("admin fixture update failed")
	}
	path := "/v1/organizations/" + r.orgID + "/api-keys"
	admin := r.requestBody(http.MethodPost, path, orgTestOrigin, r.csrf, `{"name":"Admin key"}`)
	if admin.Code != http.StatusOK {
		t.Fatalf("admin API key create returned %d", admin.Code)
	}
	if _, err := r.pool.Exec(context.Background(), `UPDATE organization_memberships SET role='member' WHERE organization_id=$1 AND user_id=$2`, r.orgID, r.userID); err != nil {
		t.Fatal("member fixture update failed")
	}
	member := r.requestBody(http.MethodPost, path, orgTestOrigin, r.csrf, `{"name":"Member key"}`)
	if member.Code != http.StatusForbidden {
		t.Fatalf("member API key create returned %d", member.Code)
	}
	var body map[string]any
	if json.Unmarshal(admin.Body.Bytes(), &body) != nil || body["dummy"] != true {
		t.Fatal("API key response contract mismatch")
	}
}

func seedSecondOwner(t *testing.T, r *orgRecorder) {
	t.Helper()
	owner := uuid.NewString()
	if _, err := r.pool.Exec(context.Background(), `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at) VALUES($1,$2,$2,'Second Owner','active',now(),now())`, owner, "owner-"+owner+"@example.test"); err != nil {
		t.Fatal("second owner seed failed")
	}
	if _, err := r.pool.Exec(context.Background(), `INSERT INTO organization_memberships(id,organization_id,user_id,status,role,created_at,updated_at) VALUES($1,$2,$3,'active','owner',now(),now())`, uuid.NewString(), r.orgID, owner); err != nil {
		t.Fatal("second owner membership seed failed")
	}
}
