package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"net/http"
	"time"
)

type V1OrganizationMemberRestoreInputBody struct {
	Revision string `json:"revision" minLength:"64" maxLength:"64" pattern:"^[a-fA-F0-9]{64}$"`
}
type V1OrganizationMemberRestoreBody struct {
	ID             string    `json:"id" format:"uuid"`
	OrganizationID string    `json:"organizationId" format:"uuid"`
	UserID         string    `json:"userId" format:"uuid"`
	Status         string    `json:"status" enum:"active,removed"`
	Role           string    `json:"role" enum:"owner,admin,member"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	Revision       string    `json:"revision"`
}
type V1OrganizationMemberRestoreOutput struct {
	Body V1OrganizationMemberRestoreBody
}

//vkit:handler
func NewRestoreOrganizationMember(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1RestoreOrganizationMember", Method: http.MethodPost, Path: "/v1/organizations/{organizationId}/memberships/{id}/restore", Errors: []int{401, 403, 404, 409, 422, 503}}, func(ctx context.Context, in *struct {
		CustomerMembershipPath
		Body V1OrganizationMemberRestoreInputBody
	}) (*V1OrganizationMemberRestoreOutput, error) {
		a, err := security.Actor(ctx)
		if err != nil {
			return nil, customerError(ctx, err)
		}
		m, err := customerCommands(b).RestoreMember(ctx, a, customerUUID(in.OrganizationID), customerUUID(in.ID), in.Body.Revision)
		if err != nil {
			return nil, customerError(ctx, err)
		}
		return &V1OrganizationMemberRestoreOutput{Body: V1OrganizationMemberRestoreBody{ID: m.ID.String(), OrganizationID: m.OrganizationID.String(), UserID: m.UserID.String(), Status: m.Status, Role: string(m.Role), CreatedAt: m.CreatedAt.UTC(), UpdatedAt: m.UpdatedAt.UTC(), Revision: customers.MembershipRevision(m)}}, nil
	})
	return nil
}
