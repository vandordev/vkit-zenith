package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/fixtures"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"net/http"
	"strings"
)

type listAPIKeysInput struct {
	CustomerOrganizationPath
	CustomerPage
	Status string `query:"status" enum:",active,expired,revoked"`
}
type V1ListAPIKeysBody struct {
	Items      []fixtures.APIKey            `json:"items"`
	Pagination V1OrganizationListPagination `json:"pagination"`
}
type V1ListAPIKeysOutput struct{ Body V1ListAPIKeysBody }

// DUMMY (ADR-0001 phase 2): returns fixtures, persists nothing.
//
//vkit:handler
func NewListAPIKeys(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ListAPIKeys", Method: http.MethodGet, Path: "/v1/organizations/{organizationId}/api-keys", Description: "[DUMMY] Data contoh untuk Backoffice.", Errors: []int{401, 403, 404, 422, 503}}, func(ctx context.Context, in *listAPIKeysInput) (*V1ListAPIKeysOutput, error) {
		d, e := dummyOrg(ctx, b, in.OrganizationID)
		if e != nil {
			return nil, e
		}
		items := make([]fixtures.APIKey, 0)
		for _, x := range d.APIKeys {
			if (in.Status == "" || x.Status == in.Status) && (in.Q == "" || strings.Contains(strings.ToLower(x.Name), strings.ToLower(in.Q))) {
				items = append(items, x)
			}
		}
		items, pg := paginate(items, in.Page, in.PageSize)
		return &V1ListAPIKeysOutput{Body: V1ListAPIKeysBody{Items: items, Pagination: pg}}, nil
	})
	return nil
}
