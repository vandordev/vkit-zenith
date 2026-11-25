package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/packages/fixtures"
)

type Organization struct {
	ID        string `json:"id" format:"uuid" doc:"Organization UUID."`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	Status    string `json:"status" enum:"active,suspended"`
	CreatedAt string `json:"createdAt" format:"date-time"`
}
type organizationGetInput struct {
	Authorization string `header:"Authorization" hidden:"true"`
}
type organizationGetOutput struct {
	CacheControl string `header:"Cache-Control" doc:"The response must not be stored by caches." example:"no-store"`
	Body         Organization
}

// DUMMY (ADR-0001 phase 2): returns the fixture organization for the key.
//
//vkit:handler
func NewGetOrganization(api huma.API) error {
	huma.Register(api, huma.Operation{
		OperationID: "getOrganization", Method: http.MethodGet, Path: "/v1/organization",
		Summary:     "Retrieve the key's organization",
		Description: "Returns the organization that owns the API key in the Authorization header.",
		Tags:        []string{"Organization"}, Security: []map[string][]string{{"bearerAuth": {}}},
		Errors: []int{http.StatusUnauthorized},
	}, func(ctx context.Context, in *organizationGetInput) (*organizationGetOutput, error) {
		orgID, err := bearerOrg(in.Authorization)
		if err != nil {
			return nil, err
		}
		o := fixtures.OrganizationFor(orgID)
		return &organizationGetOutput{CacheControl: "no-store", Body: Organization{ID: o.ID, Name: o.Name, Slug: o.Slug, Status: o.Status, CreatedAt: o.CreatedAt}}, nil
	})
	return nil
}
