package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
	"net/http"
	"time"
)

type V1StaffUpdateBody struct {
	ID          string    `json:"id" format:"uuid"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Role        string    `json:"role" enum:"superadmin,admin"`
	Status      string    `json:"status" enum:"active,disabled"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Revision    string    `json:"revision"`
}

func v1StaffUpdateBody(s staff.Staff) V1StaffUpdateBody {
	return V1StaffUpdateBody{s.ID.String(), s.Email, s.DisplayName, s.Role, s.Status, s.UpdatedAt.UTC(), staff.Revision(s)}
}

type V1StaffUpdateOutput struct{ Body V1StaffUpdateBody }

type StaffUpdateBody struct {
	StaffWriteBody
	Revision string `json:"revision" minLength:"64" maxLength:"64"`
}

//vkit:handler
func NewUpdateStaff(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1UpdateStaff", Method: http.MethodPatch, Path: "/v1/staff/{id}", Errors: []int{401, 403, 409, 422, 429, 503}}, func(ctx context.Context, in *struct {
		ID   string `path:"id" format:"uuid"`
		Body StaffUpdateBody
	}) (*V1StaffUpdateOutput, error) {
		a, err := security.Actor(ctx)
		if err != nil {
			return nil, security.Error(ctx, err)
		}
		id, err := parseID(in.ID)
		if err != nil {
			return nil, security.Error(ctx, err)
		}
		input := in.Body.input()
		input.Revision = in.Body.Revision
		s, err := b.Service.UpdateStaff(ctx, a, id, input)
		if err != nil {
			return nil, security.Error(ctx, err)
		}
		selfRevocation(ctx, b, a)
		return &V1StaffUpdateOutput{Body: v1StaffUpdateBody(s)}, nil
	})
	return nil
}
