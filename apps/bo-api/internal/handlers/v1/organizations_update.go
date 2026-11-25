package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"net/http"
	"time"
)

type V1OrganizationUpdateInputBody struct {
	Name     string `json:"name" minLength:"1" maxLength:"200"`
	Slug     string `json:"slug" minLength:"1" maxLength:"64" pattern:"^[a-z0-9]+(-[a-z0-9]+)*$"`
	Revision string `json:"revision" minLength:"64" maxLength:"64" pattern:"^[a-fA-F0-9]{64}$"`
}
type V1OrganizationUpdateBody struct {
	ID        string    `json:"id" format:"uuid"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Status    string    `json:"status" enum:"active,suspended"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Revision  string    `json:"revision"`
}
type V1OrganizationUpdateOutput struct{ Body V1OrganizationUpdateBody }

//vkit:handler
func NewUpdateOrganization(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1UpdateOrganization", Method: http.MethodPatch, Path: "/v1/organizations/{id}", Errors: []int{401, 403, 404, 409, 422, 503}}, func(ctx context.Context, in *struct {
		ItemPath
		Body V1OrganizationUpdateInputBody
	}) (*V1OrganizationUpdateOutput, error) {
		a, err := security.Actor(ctx)
		if err != nil {
			return nil, customerError(ctx, err)
		}
		o, err := customerCommands(b).UpdateOrganization(ctx, a, customerUUID(in.ID), customers.OrganizationProfile{Name: in.Body.Name, Slug: in.Body.Slug, Revision: in.Body.Revision})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		return &V1OrganizationUpdateOutput{Body: V1OrganizationUpdateBody{ID: o.ID.String(), Name: o.Name, Slug: o.Slug, Status: o.Status, CreatedAt: o.CreatedAt.UTC(), UpdatedAt: o.UpdatedAt.UTC(), Revision: customers.OrganizationRevision(o)}}, nil
	})
	return nil
}
