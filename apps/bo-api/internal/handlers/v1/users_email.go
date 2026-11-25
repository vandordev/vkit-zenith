package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"net/http"
	"time"
)

type V1UserEmailInputBody struct {
	CustomerSensitiveInput
	Email string `json:"email" minLength:"1" maxLength:"254"`
}
type V1UserEmailBody struct {
	ID              string     `json:"id" format:"uuid"`
	Email           string     `json:"email"`
	DisplayName     string     `json:"displayName"`
	Status          string     `json:"status" enum:"active,disabled"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	Revision        string     `json:"revision"`
}
type V1UserEmailOutput struct{ Body V1UserEmailBody }

//vkit:handler
func NewChangeUserEmail(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ChangeUserEmail", Method: http.MethodPost, Path: "/v1/users/{id}/email", Errors: []int{401, 403, 404, 409, 422, 429, 503}}, func(ctx context.Context, in *struct {
		ItemPath
		Body V1UserEmailInputBody
	}) (*V1UserEmailOutput, error) {
		a, err := security.Actor(ctx)
		if err != nil {
			return nil, customerError(ctx, err)
		}
		u, err := customerCommands(b).ChangeUserEmail(ctx, a, customerUUID(in.ID), customers.UserEmail{Sensitive: in.Body.domain(), Email: in.Body.Email})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		return &V1UserEmailOutput{Body: V1UserEmailBody{ID: u.ID.String(), Email: u.Email, DisplayName: u.DisplayName, Status: u.Status, EmailVerifiedAt: u.EmailVerifiedAt, CreatedAt: u.CreatedAt.UTC(), UpdatedAt: u.UpdatedAt.UTC(), Revision: customers.UserRevision(u)}}, nil
	})
	return nil
}
