package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"net/http"
)

type V1AuthLogoutOutput struct{}

//vkit:handler
func NewLogout(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1Logout", Method: http.MethodPost, Path: "/v1/auth/logout", DefaultStatus: 204, Errors: []int{401, 403, 503}}, func(ctx context.Context, _ *struct{}) (*V1AuthLogoutOutput, error) {
		a, err := security.Actor(ctx)
		if err == nil {
			err = b.Service.Logout(ctx, a)
		}
		if err != nil {
			return nil, security.Error(ctx, err)
		}
		b.ClearSession(ctx)
		return nil, nil
	})
	return nil
}
