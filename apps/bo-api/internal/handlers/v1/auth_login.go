package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
	"net/http"
	"time"
)

type LoginBody struct {
	Email    string `json:"email" maxLength:"254"`
	Password string `json:"password" maxLength:"512"`
}
type V1AuthLoginStaff struct {
	ID          string    `json:"id" format:"uuid"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Role        string    `json:"role" enum:"superadmin,admin"`
	Status      string    `json:"status" enum:"active,disabled"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Revision    string    `json:"revision"`
}
type V1AuthLoginBody struct {
	Staff     V1AuthLoginStaff `json:"staff"`
	ExpiresAt time.Time        `json:"expiresAt"`
}
type V1AuthLoginOutput struct{ Body V1AuthLoginBody }

//vkit:handler
func NewLogin(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1Login", Method: http.MethodPost, Path: "/v1/auth/login", Errors: []int{401, 403, 409, 422, 429, 503}}, func(ctx context.Context, in *struct{ Body LoginBody }) (*V1AuthLoginOutput, error) {
		r, err := b.Service.Login(ctx, staff.LoginInput{Email: in.Body.Email, Password: in.Body.Password, Source: security.Source(ctx)})
		if err != nil {
			return nil, security.Error(ctx, err)
		}
		if err = b.SetSession(ctx, r.Token, r.ExpiresAt); err != nil {
			return nil, security.Error(ctx, err)
		}
		out := &V1AuthLoginOutput{}
		s := r.Staff
		out.Body.Staff = V1AuthLoginStaff{s.ID.String(), s.Email, s.DisplayName, s.Role, s.Status, s.UpdatedAt.UTC(), staff.Revision(s)}
		out.Body.ExpiresAt = r.ExpiresAt
		return out, nil
	})
	return nil
}
