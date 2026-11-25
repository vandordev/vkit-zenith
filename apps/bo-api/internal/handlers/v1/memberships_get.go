package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/persistence"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"net/http"
	"time"
)

type V1OrganizationMembershipGetBody struct {
	ID                   string    `json:"id" format:"uuid"`
	OrganizationID       string    `json:"organizationId" format:"uuid"`
	UserID               string    `json:"userId" format:"uuid"`
	Status               string    `json:"status" enum:"active,removed"`
	Role                 string    `json:"role" enum:"owner,admin,member"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
	Revision             string    `json:"revision"`
	UserName             string    `json:"userName"`
	UserEmail            string    `json:"userEmail"`
	UserStatus           string    `json:"userStatus" enum:"active,disabled"`
	OrganizationName     string    `json:"organizationName"`
	OrganizationSlug     string    `json:"organizationSlug"`
	OrganizationStatus   string    `json:"organizationStatus" enum:"active,suspended"`
	CanRemove            bool      `json:"canRemove"`
	RemoveBlockedReason  string    `json:"removeBlockedReason"`
	CanRestore           bool      `json:"canRestore"`
	RestoreBlockedReason string    `json:"restoreBlockedReason"`
	CanChangeRole        bool      `json:"canChangeRole"`
	RoleBlockedReason    string    `json:"roleBlockedReason"`
}
type V1OrganizationMembershipGetOutput struct {
	Body V1OrganizationMembershipGetBody
}

//vkit:handler
func NewGetOrganizationMembership(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1GetOrganizationMembership", Method: http.MethodGet, Path: "/v1/organizations/{organizationId}/memberships/{id}", Errors: []int{401, 403, 404, 422, 503}}, func(ctx context.Context, in *CustomerMembershipPath) (*V1OrganizationMembershipGetOutput, error) {
		q, tx, actor, err := customerSnapshot(ctx, b)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback(ctx)
		r, err := q.BOGetCustomerMembership(ctx, persistence.BOGetCustomerMembershipParams{ID: customerUUID(in.ID), OrganizationID: customerUUID(in.OrganizationID)})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		out := &V1OrganizationMembershipGetOutput{Body: V1OrganizationMembershipGetBody(customerMembershipListItem(persistence.BOListCustomerMembershipsRow(r), actor.Role == "superadmin"))}
		if err = tx.Commit(ctx); err != nil {
			return nil, customerError(ctx, err)
		}
		return out, nil
	})
	return nil
}
