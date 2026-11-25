package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
	"net/http"
	"time"
)

type V1AuthMeBody struct {
	ID          string    `json:"id" format:"uuid"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Role        string    `json:"role" enum:"superadmin,admin"`
	Status      string    `json:"status" enum:"active,disabled"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Revision    string    `json:"revision"`
}
type V1AuthMeOutput struct{ Body V1AuthMeBody }

//vkit:handler
func NewGetMe(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1GetMe", Method: http.MethodGet, Path: "/v1/auth/me", Errors: []int{401, 503}}, func(ctx context.Context, _ *struct{}) (*V1AuthMeOutput, error) {
		s, err := b.Read(ctx, "any_staff")
		if err != nil {
			return nil, security.Error(ctx, err)
		}
		return &V1AuthMeOutput{Body: V1AuthMeBody{s.ID.String(), s.Email, s.DisplayName, s.Role, s.Status, s.UpdatedAt.UTC(), staff.Revision(s)}}, nil
	})
	return nil
}
