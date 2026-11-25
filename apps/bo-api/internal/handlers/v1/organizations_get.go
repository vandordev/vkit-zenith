package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/persistence"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"net/http"
	"time"
)

type V1OrganizationGetBody struct {
	ID                  string    `json:"id" format:"uuid"`
	Name                string    `json:"name"`
	Slug                string    `json:"slug"`
	Status              string    `json:"status" enum:"active,suspended"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
	Revision            string    `json:"revision"`
	MemberCount         string    `json:"memberCount"`
	ActiveMemberCount   string    `json:"activeMemberCount"`
	ActiveOwnerCount    string    `json:"activeOwnerCount"`
	CanChangeStatus     bool      `json:"canChangeStatus"`
	StatusBlockedReason string    `json:"statusBlockedReason"`
}
type V1OrganizationGetOutput struct{ Body V1OrganizationGetBody }

//vkit:handler
func NewGetOrganization(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1GetOrganization", Method: http.MethodGet, Path: "/v1/organizations/{id}", Errors: []int{401, 403, 404, 422, 503}}, func(ctx context.Context, in *ItemPath) (*V1OrganizationGetOutput, error) {
		q, tx, actor, err := customerSnapshot(ctx, b)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback(ctx)
		r, err := q.BOGetCustomerOrganization(ctx, customerUUID(in.ID))
		if err != nil {
			return nil, customerError(ctx, err)
		}
		out := &V1OrganizationGetOutput{Body: V1OrganizationGetBody(customerOrganizationListItem(persistence.BOListCustomerOrganizationsRow(r), actor.Role == "superadmin"))}
		if err = tx.Commit(ctx); err != nil {
			return nil, customerError(ctx, err)
		}
		return out, nil
	})
	return nil
}
