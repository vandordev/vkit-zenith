package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/packages/fixtures"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/orgaccess"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"net/http"
)

type apiKeyRevokeInput struct {
	OrganizationID string `path:"organizationId"`
	KeyID          string `path:"keyId"`
	Origin         string `header:"Origin"`
	CSRF           string `header:"X-CSRF-Token"`
}

// DUMMY (ADR-0001 phase 2): returns fixtures, persists nothing.
//
//vkit:handler
func NewRevokeAPIKey(api huma.API, b *orgaccess.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "webV1RevokeApiKey", Method: http.MethodPost, Path: "/v1/organizations/{organizationId}/api-keys/{keyId}/revoke", Description: "[DUMMY] Simulate revoking an API key.", Errors: []int{401, 403, 404, 409, 422, 503}}, func(ctx context.Context, in *apiKeyRevokeInput) (*settingsWriteOutput[fixtures.APIKey], error) {
		if err := settingsWrite(ctx, b, in.OrganizationID, customers.PermissionDevelopersManage, in.Origin, in.CSRF); err != nil {
			return nil, err
		}
		id, e := fixtureOrg(in.OrganizationID)
		if e != nil {
			return nil, dummyNotFound()
		}
		for _, v := range fixtures.Org(id).APIKeys {
			if v.ID == in.KeyID {
				v.Status = "revoked"
				revokedAt := fixtures.BaseDate
				v.RevokedAt = &revokedAt
				return &settingsWriteOutput[fixtures.APIKey]{CacheControl: "no-store", Body: V1DummyResult[fixtures.APIKey]{Result: v, Dummy: true}}, nil
			}
		}
		return nil, dummyNotFound()
	})
	return nil
}
