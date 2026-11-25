package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"net/http"
	"time"
)

type V1UserUpdateInputBody struct {
	DisplayName string `json:"displayName" minLength:"1" maxLength:"200"`
	Revision    string `json:"revision" minLength:"64" maxLength:"64" pattern:"^[a-fA-F0-9]{64}$"`
}
type V1UserUpdateBody struct {
	ID              string     `json:"id" format:"uuid"`
	Email           string     `json:"email"`
	DisplayName     string     `json:"displayName"`
	Status          string     `json:"status" enum:"active,disabled"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	Revision        string     `json:"revision"`
}
type V1UserUpdateOutput struct{ Body V1UserUpdateBody }

//vkit:handler
func NewUpdateUser(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1UpdateUser", Method: http.MethodPatch, Path: "/v1/users/{id}", Errors: []int{401, 403, 404, 409, 422, 503}}, func(ctx context.Context, in *struct {
		ItemPath
		Body V1UserUpdateInputBody
	}) (*V1UserUpdateOutput, error) {
		a, err := security.Actor(ctx)
		if err != nil {
			return nil, customerError(ctx, err)
		}
		u, err := customerCommands(b).UpdateUser(ctx, a, customerUUID(in.ID), customers.UserProfile{DisplayName: in.Body.DisplayName, Revision: in.Body.Revision})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		return &V1UserUpdateOutput{Body: V1UserUpdateBody{ID: u.ID.String(), Email: u.Email, DisplayName: u.DisplayName, Status: u.Status, EmailVerifiedAt: u.EmailVerifiedAt, CreatedAt: u.CreatedAt.UTC(), UpdatedAt: u.UpdatedAt.UTC(), Revision: customers.UserRevision(u)}}, nil
	})
	return nil
}
