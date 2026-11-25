package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"net/http"
	"time"
)

type V1UserCreateInputBody struct {
	Email       string `json:"email" minLength:"1" maxLength:"254"`
	DisplayName string `json:"displayName" minLength:"1" maxLength:"200"`
}
type V1UserCreateBody struct {
	ID              string     `json:"id" format:"uuid"`
	Email           string     `json:"email"`
	DisplayName     string     `json:"displayName"`
	Status          string     `json:"status" enum:"active,disabled"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	Revision        string     `json:"revision"`
}
type V1UserCreateOutput struct{ Body V1UserCreateBody }

//vkit:handler
func NewCreateUser(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1CreateUser", Method: http.MethodPost, Path: "/v1/users", DefaultStatus: 201, Errors: []int{401, 403, 409, 422, 503}}, func(ctx context.Context, in *struct{ Body V1UserCreateInputBody }) (*V1UserCreateOutput, error) {
		a, err := security.Actor(ctx)
		if err != nil {
			return nil, customerError(ctx, err)
		}
		u, err := customerCommands(b).CreateUser(ctx, a, customers.UserCreate{Email: in.Body.Email, DisplayName: in.Body.DisplayName})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		return &V1UserCreateOutput{Body: V1UserCreateBody{ID: u.ID.String(), Email: u.Email, DisplayName: u.DisplayName, Status: u.Status, EmailVerifiedAt: u.EmailVerifiedAt, CreatedAt: u.CreatedAt.UTC(), UpdatedAt: u.UpdatedAt.UTC(), Revision: customers.UserRevision(u)}}, nil
	})
	return nil
}
