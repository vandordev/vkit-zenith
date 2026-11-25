package v1

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/vandordev/vkit-zenith/packages/fixtures"
)

func TestGetOrganizationReturnsKeyOrganization(t *testing.T) {
	_, api := humatest.New(t)
	if err := NewGetOrganization(api); err != nil {
		t.Fatal(err)
	}
	resp := api.Get("/v1/organization", "Authorization: Bearer "+fixtures.DemoAPIKey)
	var body Organization
	if resp.Code != http.StatusOK || resp.Header().Get("Cache-Control") != "no-store" || json.Unmarshal(resp.Body.Bytes(), &body) != nil {
		t.Fatalf("status=%d body=%s", resp.Code, resp.Body.String())
	}
	if body.ID != fixtures.DemoOrgID || body.Status != "active" || body.Name == "" || body.Slug == "" || body.CreatedAt == "" {
		t.Fatalf("unexpected organization: %+v", body)
	}
	resp = api.Get("/v1/organization", "Authorization: Bearer "+fixtures.SuspendedTestAPIKey)
	if resp.Code != http.StatusOK || json.Unmarshal(resp.Body.Bytes(), &body) != nil || body.ID != fixtures.SuspendedTestOrgID || body.Status != "suspended" {
		t.Fatalf("suspended key: status=%d body=%s", resp.Code, resp.Body.String())
	}
}

func TestGetOrganizationRejectsMissingOrUnknownKey(t *testing.T) {
	_, api := humatest.New(t)
	if err := NewGetOrganization(api); err != nil {
		t.Fatal(err)
	}
	for _, header := range []string{"", "Authorization: Bearer zk_unknownkey", "Authorization: Basic " + fixtures.DemoAPIKey, "Authorization: Bearer"} {
		args := []any{}
		if header != "" {
			args = append(args, header)
		}
		if resp := api.Get("/v1/organization", args...); resp.Code != http.StatusUnauthorized {
			t.Errorf("%q returned %d", header, resp.Code)
		}
	}
}
