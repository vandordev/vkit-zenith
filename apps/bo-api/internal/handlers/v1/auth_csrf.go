package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"net/http"
)

type V1AuthCsrfBody struct {
	Token string `json:"token"`
}
type V1AuthCsrfOutput struct{ Body V1AuthCsrfBody }

//vkit:handler
func NewGetCSRF(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1GetCSRF", Method: http.MethodGet, Path: "/v1/auth/csrf", Errors: []int{401, 403, 503}}, func(ctx context.Context, _ *struct{}) (*V1AuthCsrfOutput, error) {
		token, err := b.CSRF(ctx)
		if err != nil {
			return nil, security.Error(ctx, err)
		}
		out := &V1AuthCsrfOutput{}
		out.Body.Token = token
		return out, nil
	})
	return nil
}
