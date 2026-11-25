package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/orgaccess"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
)

type webOrganizationPath struct {
	OrganizationID string `path:"organizationId"`
}
type V1OrganizationProfile struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Slug    string `json:"slug"`
	Status  string `json:"status" enum:"active,suspended"`
	HasLogo bool   `json:"hasLogo"`
}
type webOrganizationProfileOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         V1OrganizationProfile
}

//vkit:handler
func NewGetOrganizationProfile(api huma.API, boundary *orgaccess.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "webV1GetOrganizationProfile", Method: http.MethodGet, Path: "/v1/organizations/{organizationId}/profile", Errors: []int{401, 403, 404, 503}}, func(ctx context.Context, in *webOrganizationPath) (*webOrganizationProfileOutput, error) {
		access, err := boundary.Read(ctx, in.OrganizationID, customers.PermissionOrganizationRead)
		if err != nil {
			return nil, err
		}
		return &webOrganizationProfileOutput{CacheControl: "no-store", Body: V1OrganizationProfile{ID: access.OrganizationID.String(), Name: access.Name, Slug: access.Slug, Status: organizationStatus(access.Suspended), HasLogo: access.HasLogo}}, nil
	})
	return nil
}

func organizationStatus(suspended bool) string {
	if suspended {
		return "suspended"
	}
	return "active"
}
