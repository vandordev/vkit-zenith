package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/packages/fixtures"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/orgaccess"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"net/http"
	"time"
)

type apiKeyCreateInput struct {
	OrganizationID string `path:"organizationId"`
	Origin         string `header:"Origin"`
	CSRF           string `header:"X-CSRF-Token"`
	Body           struct {
		Name      string  `json:"name" minLength:"1" maxLength:"120"`
		ExpiresAt *string `json:"expiresAt,omitempty" format:"date-time"`
	}
}

// DUMMY (ADR-0001 phase 2): returns fixtures, persists nothing.
//
//vkit:handler
func NewCreateAPIKey(api huma.API, b *orgaccess.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "webV1CreateApiKey", Method: http.MethodPost, Path: "/v1/organizations/{organizationId}/api-keys", DefaultStatus: http.StatusOK, Description: "[DUMMY] Simulate creating an API key; secret is shown once.", Errors: []int{401, 403, 404, 409, 422, 503}}, func(ctx context.Context, in *apiKeyCreateInput) (*apiKeyCreateOutput, error) {
		if err := settingsWrite(ctx, b, in.OrganizationID, customers.PermissionDevelopersManage, in.Origin, in.CSRF); err != nil {
			return nil, err
		}
		if in.Body.Name == "" {
			return nil, dummyInvalid()
		}
		if in.Body.ExpiresAt != nil {
			if _, err := time.Parse(time.RFC3339, *in.Body.ExpiresAt); err != nil {
				return nil, dummyInvalid()
			}
		}
		id, e := fixtureOrg(in.OrganizationID)
		if e != nil {
			return nil, dummyNotFound()
		}
		v := fixtures.APIKey{ID: fixtures.FixtureID(id, "simulated-api-key").String(), Name: in.Body.Name, Prefix: "zk_dummy", Status: "active", ExpiresAt: in.Body.ExpiresAt, CreatedAt: fixtures.BaseDate}
		return &apiKeyCreateOutput{CacheControl: "no-store", Body: V1DummyResult[V1CreateApiKeyResult]{Result: V1CreateApiKeyResult{Key: v, Secret: "zk_DUMMYNOTAKEY"}, Dummy: true}}, nil
	})
	return nil
}
