package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
	"net/http"
)

type StaffPasswordBody struct {
	Password        string `json:"password" maxLength:"512"`
	CurrentPassword string `json:"currentPassword" maxLength:"512"`
	Revision        string `json:"revision" minLength:"64" maxLength:"64"`
}
type V1StaffPasswordResetOutput struct{}

//vkit:handler
func NewReplaceStaffPassword(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ReplaceStaffPassword", Method: http.MethodPost, Path: "/v1/staff/{id}/password", DefaultStatus: 204, Errors: []int{401, 403, 409, 422, 429, 503}}, func(ctx context.Context, in *struct {
		ID   string `path:"id" format:"uuid"`
		Body StaffPasswordBody
	}) (*V1StaffPasswordResetOutput, error) {
		a, err := security.Actor(ctx)
		if err != nil {
			return nil, security.Error(ctx, err)
		}
		id, err := parseID(in.ID)
		if err == nil {
			err = b.Service.ReplacePassword(ctx, a, id, staff.StaffWrite{Password: in.Body.Password, CurrentPassword: in.Body.CurrentPassword, Revision: in.Body.Revision})
		}
		if err != nil {
			return nil, security.Error(ctx, err)
		}
		selfRevocation(ctx, b, a)
		return nil, nil
	})
	return nil
}
