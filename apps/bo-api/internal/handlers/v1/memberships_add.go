package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"net/http"
	"time"
)

type V1OrganizationMemberAddInputBody struct {
	UserID string `json:"userId" format:"uuid"`
	Role   string `json:"role" enum:"admin,member"`
}
type V1OrganizationMemberAddBody struct {
	ID             string    `json:"id" format:"uuid"`
	OrganizationID string    `json:"organizationId" format:"uuid"`
	UserID         string    `json:"userId" format:"uuid"`
	Status         string    `json:"status" enum:"active,removed"`
	Role           string    `json:"role" enum:"owner,admin,member"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Revision       string    `json:"revision"`
}
type V1OrganizationMemberAddOutput struct{ Body V1OrganizationMemberAddBody }

//vkit:handler
func NewAddOrganizationMember(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1AddOrganizationMember", Method: http.MethodPost, Path: "/v1/organizations/{organizationId}/memberships", DefaultStatus: 201, Errors: []int{401, 403, 404, 409, 422, 503}}, func(ctx context.Context, in *struct {
		CustomerOrganizationPath
		Body V1OrganizationMemberAddInputBody
	}) (*V1OrganizationMemberAddOutput, error) {
		a, err := security.Actor(ctx)
		if err != nil {
			return nil, customerError(ctx, err)
		}
		m, err := customerCommands(b).AddMember(ctx, a, customerUUID(in.OrganizationID), customerUUID(in.Body.UserID), customers.Role(in.Body.Role))
		if err != nil {
			return nil, customerError(ctx, err)
		}
		return &V1OrganizationMemberAddOutput{Body: V1OrganizationMemberAddBody{ID: m.ID.String(), OrganizationID: m.OrganizationID.String(), UserID: m.UserID.String(), Status: m.Status, Role: string(m.Role), CreatedAt: m.CreatedAt.UTC(), UpdatedAt: m.UpdatedAt.UTC(), Revision: customers.MembershipRevision(m)}}, nil
	})
	return nil
}
