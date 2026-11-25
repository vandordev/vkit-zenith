package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
	"net/http"
)

type PasswordChangeBody struct {
	CurrentPassword string `json:"currentPassword" maxLength:"512"`
	NewPassword     string `json:"newPassword" maxLength:"512"`
}

type V1AuthPasswordChangeOutput struct{}

//vkit:handler
func NewChangePassword(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ChangePassword", Method: http.MethodPost, Path: "/v1/auth/password", DefaultStatus: 204, Errors: []int{401, 403, 422, 429, 503}}, func(ctx context.Context, in *struct{ Body PasswordChangeBody }) (*V1AuthPasswordChangeOutput, error) {
		a, err := security.Actor(ctx)
		if err == nil {
			err = b.Service.ChangePassword(ctx, a, staff.PasswordChange{CurrentPassword: in.Body.CurrentPassword, NewPassword: in.Body.NewPassword})
		}
		if err != nil {
			return nil, security.Error(ctx, err)
		}
		b.ClearSession(ctx)
		return nil, nil
	})
	return nil
}
