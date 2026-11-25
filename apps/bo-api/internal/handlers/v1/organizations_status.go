package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"net/http"
	"time"
)

type V1OrganizationStatusInputBody struct {
	CustomerSensitiveInput
	Status string `json:"status" enum:"active,suspended"`
}
type V1OrganizationStatusBody struct {
	ID        string    `json:"id" format:"uuid"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Status    string    `json:"status" enum:"active,suspended"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Revision  string    `json:"revision"`
}
type V1OrganizationStatusOutput struct{ Body V1OrganizationStatusBody }

//vkit:handler
func NewSetOrganizationStatus(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1SetOrganizationStatus", Method: http.MethodPost, Path: "/v1/organizations/{id}/status", Errors: []int{401, 403, 404, 409, 422, 429, 503}}, func(ctx context.Context, in *struct {
		ItemPath
		Body V1OrganizationStatusInputBody
	}) (*V1OrganizationStatusOutput, error) {
		a, err := security.Actor(ctx)
		if err != nil {
			return nil, customerError(ctx, err)
		}
		o, err := customerCommands(b).SetOrganizationStatus(ctx, a, customerUUID(in.ID), customers.OrganizationStatus{Sensitive: in.Body.domain(), Status: in.Body.Status})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		return &V1OrganizationStatusOutput{Body: V1OrganizationStatusBody{ID: o.ID.String(), Name: o.Name, Slug: o.Slug, Status: o.Status, CreatedAt: o.CreatedAt.UTC(), UpdatedAt: o.UpdatedAt.UTC(), Revision: customers.OrganizationRevision(o)}}, nil
	})
	return nil
}
