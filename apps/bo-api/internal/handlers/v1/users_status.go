package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"net/http"
	"time"
)

type V1UserStatusInputBody struct {
	CustomerSensitiveInput
	Status string `json:"status" enum:"active,disabled"`
}
type V1UserStatusBody struct {
	ID              string     `json:"id" format:"uuid"`
	Email           string     `json:"email"`
	DisplayName     string     `json:"displayName"`
	Status          string     `json:"status" enum:"active,disabled"`
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	Revision        string     `json:"revision"`
}
type V1UserStatusOutput struct{ Body V1UserStatusBody }

//vkit:handler
func NewSetUserStatus(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1SetUserStatus", Method: http.MethodPost, Path: "/v1/users/{id}/status", Errors: []int{401, 403, 404, 409, 422, 429, 503}}, func(ctx context.Context, in *struct {
		ItemPath
		Body V1UserStatusInputBody
	}) (*V1UserStatusOutput, error) {
		a, err := security.Actor(ctx)
		if err != nil {
			return nil, customerError(ctx, err)
		}
		u, err := customerCommands(b).SetUserStatus(ctx, a, customerUUID(in.ID), customers.UserStatus{Sensitive: in.Body.domain(), Status: in.Body.Status})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		return &V1UserStatusOutput{Body: V1UserStatusBody{ID: u.ID.String(), Email: u.Email, DisplayName: u.DisplayName, Status: u.Status, EmailVerifiedAt: u.EmailVerifiedAt, CreatedAt: u.CreatedAt.UTC(), UpdatedAt: u.UpdatedAt.UTC(), Revision: customers.UserRevision(u)}}, nil
	})
	return nil
}
