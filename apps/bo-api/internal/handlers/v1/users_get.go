package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/persistence"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"net/http"
	"time"
)

type V1UserGetBody struct {
	ID                        string     `json:"id" format:"uuid"`
	Email                     string     `json:"email"`
	DisplayName               string     `json:"displayName"`
	Status                    string     `json:"status" enum:"active,disabled"`
	EmailVerifiedAt           *time.Time `json:"emailVerifiedAt"`
	CreatedAt                 time.Time  `json:"createdAt"`
	UpdatedAt                 time.Time  `json:"updatedAt"`
	Revision                  string     `json:"revision"`
	MembershipCount           string     `json:"membershipCount"`
	ActiveMembershipCount     string     `json:"activeMembershipCount"`
	OwnerMembershipCount      string     `json:"ownerMembershipCount"`
	CanChangeStatus           bool       `json:"canChangeStatus"`
	StatusBlockedReason       string     `json:"statusBlockedReason"`
	CanChangeEmail            bool       `json:"canChangeEmail"`
	EmailBlockedReason        string     `json:"emailBlockedReason"`
	CanVerifyEmail            bool       `json:"canVerifyEmail"`
	VerificationBlockedReason string     `json:"verificationBlockedReason"`
}
type V1UserGetOutput struct{ Body V1UserGetBody }

//vkit:handler
func NewGetUser(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1GetUser", Method: http.MethodGet, Path: "/v1/users/{id}", Errors: []int{401, 403, 404, 422, 503}}, func(ctx context.Context, in *ItemPath) (*V1UserGetOutput, error) {
		q, tx, actor, err := customerSnapshot(ctx, b)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback(ctx)
		r, err := q.BOGetCustomerUser(ctx, customerUUID(in.ID))
		if err != nil {
			return nil, customerError(ctx, err)
		}
		out := &V1UserGetOutput{Body: V1UserGetBody(customerUserListItem(persistence.BOListCustomerUsersRow(r), actor.Role == "superadmin"))}
		if err = tx.Commit(ctx); err != nil {
			return nil, customerError(ctx, err)
		}
		return out, nil
	})
	return nil
}
