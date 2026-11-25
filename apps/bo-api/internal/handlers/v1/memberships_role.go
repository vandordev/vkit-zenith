package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"net/http"
	"time"
)

type V1OrganizationMembershipRoleInputBody struct {
	CustomerSensitiveInput
	Role string `json:"role" enum:"owner,admin,member"`
}
type V1OrganizationMembershipRoleBody struct {
	ID             string    `json:"id" format:"uuid"`
	OrganizationID string    `json:"organizationId" format:"uuid"`
	UserID         string    `json:"userId" format:"uuid"`
	Status         string    `json:"status" enum:"active,removed"`
	Role           string    `json:"role" enum:"owner,admin,member"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Revision       string    `json:"revision"`
}
type V1OrganizationMembershipRoleOutput struct {
	Body V1OrganizationMembershipRoleBody
}

//vkit:handler
func NewSetOrganizationMembershipRole(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1SetOrganizationMembershipRole", Method: http.MethodPost, Path: "/v1/organizations/{organizationId}/memberships/{id}/role", Errors: []int{401, 403, 404, 409, 422, 429, 503}}, func(ctx context.Context, in *struct {
		CustomerMembershipPath
		Body V1OrganizationMembershipRoleInputBody
	}) (*V1OrganizationMembershipRoleOutput, error) {
		a, err := security.Actor(ctx)
		if err != nil {
			return nil, customerError(ctx, err)
		}
		m, err := customerCommands(b).SetRole(ctx, a, customerUUID(in.OrganizationID), customerUUID(in.ID), customers.RoleChange{Sensitive: in.Body.domain(), Role: customers.Role(in.Body.Role)})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		return &V1OrganizationMembershipRoleOutput{Body: V1OrganizationMembershipRoleBody{ID: m.ID.String(), OrganizationID: m.OrganizationID.String(), UserID: m.UserID.String(), Status: m.Status, Role: string(m.Role), CreatedAt: m.CreatedAt.UTC(), UpdatedAt: m.UpdatedAt.UTC(), Revision: customers.MembershipRevision(m)}}, nil
	})
	return nil
}
