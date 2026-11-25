package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/fixtures"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"net/http"
)

type V1RevokeAPIKeyBody struct{ V1Adjustment }
type V1RevokeAPIKeyOutput struct {
	Body V1DummyResult[fixtures.APIKey]
}

// DUMMY (ADR-0001 phase 2): returns fixtures, persists nothing.
//
//vkit:handler
func NewRevokeAPIKey(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1RevokeAPIKey", Method: http.MethodPost, Path: "/v1/organizations/{organizationId}/api-keys/{apiKeyId}/revoke", Description: "[DUMMY] Pencabutan kunci contoh Backoffice.", Errors: []int{401, 403, 404, 422, 503}}, func(ctx context.Context, in *struct {
		CustomerOrganizationPath
		ID   string `path:"apiKeyId" format:"uuid"`
		Body V1RevokeAPIKeyBody
	}) (*V1RevokeAPIKeyOutput, error) {
		d, e := dummyOrg(ctx, b, in.OrganizationID)
		if e != nil {
			return nil, e
		}
		for _, x := range d.APIKeys {
			if x.ID == in.ID {
				x.Status = "revoked"
				revokedAt := "2026-11-01T00:00:00Z"
				x.RevokedAt = &revokedAt
				return &V1RevokeAPIKeyOutput{Body: V1DummyResult[fixtures.APIKey]{Result: x, Dummy: true}}, nil
			}
		}
		return nil, dummyNotFound()
	})
	return nil
}
