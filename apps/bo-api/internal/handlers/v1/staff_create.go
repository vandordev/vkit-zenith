package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
	"net/http"
	"time"
)

type V1StaffCreateBody struct {
	ID          string    `json:"id" format:"uuid"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Role        string    `json:"role" enum:"superadmin,admin"`
	Status      string    `json:"status" enum:"active,disabled"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Revision    string    `json:"revision"`
}

func v1StaffCreateBody(s staff.Staff) V1StaffCreateBody {
	return V1StaffCreateBody{s.ID.String(), s.Email, s.DisplayName, s.Role, s.Status, s.UpdatedAt.UTC(), staff.Revision(s)}
}

type V1StaffCreateOutput struct{ Body V1StaffCreateBody }

type StaffCreateBody struct {
	StaffWriteBody
	Password string `json:"password" maxLength:"512"`
}

//vkit:handler
func NewCreateStaff(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1CreateStaff", Method: http.MethodPost, Path: "/v1/staff", DefaultStatus: 201, Errors: []int{401, 403, 409, 422, 429, 503}}, func(ctx context.Context, in *struct{ Body StaffCreateBody }) (*V1StaffCreateOutput, error) {
		a, err := security.Actor(ctx)
		if err != nil {
			return nil, security.Error(ctx, err)
		}
		input := in.Body.input()
		input.Password = in.Body.Password
		s, err := b.Service.CreateStaff(ctx, a, input)
		if err != nil {
			return nil, security.Error(ctx, err)
		}
		return &V1StaffCreateOutput{Body: v1StaffCreateBody(s)}, nil
	})
	return nil
}
