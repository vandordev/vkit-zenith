package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"net/http"
	"time"
)

type V1OrganizationCreateNewOwner struct {
	Email       string `json:"email" minLength:"1" maxLength:"254"`
	DisplayName string `json:"displayName" minLength:"1" maxLength:"200"`
}
type V1OrganizationCreateInputBody struct {
	Name        string                        `json:"name" minLength:"1" maxLength:"200"`
	Slug        string                        `json:"slug" minLength:"1" maxLength:"64" pattern:"^[a-z0-9]+(-[a-z0-9]+)*$"`
	OwnerUserID string                        `json:"ownerUserId,omitempty" format:"uuid"`
	NewOwner    *V1OrganizationCreateNewOwner `json:"newOwner,omitempty"`
}
type V1OrganizationCreateBody struct {
	ID                string    `json:"id" format:"uuid"`
	Name              string    `json:"name"`
	Slug              string    `json:"slug"`
	Status            string    `json:"status" enum:"active,suspended"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
	Revision          string    `json:"revision"`
	OwnerUserID       string    `json:"ownerUserId" format:"uuid"`
	OwnerMembershipID string    `json:"ownerMembershipId" format:"uuid"`
}
type V1OrganizationCreateOutput struct{ Body V1OrganizationCreateBody }

//vkit:handler
func NewCreateOrganization(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1CreateOrganization", Method: http.MethodPost, Path: "/v1/organizations", DefaultStatus: 201, Errors: []int{401, 403, 404, 409, 422, 503}}, func(ctx context.Context, in *struct{ Body V1OrganizationCreateInputBody }) (*V1OrganizationCreateOutput, error) {
		a, err := security.Actor(ctx)
		if err != nil {
			return nil, customerError(ctx, err)
		}
		input := customers.OrganizationCreate{Name: in.Body.Name, Slug: in.Body.Slug}
		if in.Body.OwnerUserID != "" {
			input.OwnerUserID = customerUUID(in.Body.OwnerUserID)
		}
		if in.Body.NewOwner != nil {
			input.NewOwner = &customers.UserCreate{Email: in.Body.NewOwner.Email, DisplayName: in.Body.NewOwner.DisplayName}
		}
		created, err := customerCommands(b).CreateOrganization(ctx, a, input)
		if err != nil {
			return nil, customerError(ctx, err)
		}
		o := created.Organization
		return &V1OrganizationCreateOutput{Body: V1OrganizationCreateBody{ID: o.ID.String(), Name: o.Name, Slug: o.Slug, Status: o.Status, CreatedAt: o.CreatedAt.UTC(), UpdatedAt: o.UpdatedAt.UTC(), Revision: customers.OrganizationRevision(o), OwnerUserID: created.User.ID.String(), OwnerMembershipID: created.Owner.ID.String()}}, nil
	})
	return nil
}
