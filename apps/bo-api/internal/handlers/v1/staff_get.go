package v1

import (
	"context"
	"errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/persistence"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
	"net/http"
	"time"
)

type V1StaffDetailBody struct {
	ID          string    `json:"id" format:"uuid"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Role        string    `json:"role" enum:"superadmin,admin"`
	Status      string    `json:"status" enum:"active,disabled"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Revision    string    `json:"revision"`
}

func v1StaffDetailBody(s staff.Staff) V1StaffDetailBody {
	return V1StaffDetailBody{s.ID.String(), s.Email, s.DisplayName, s.Role, s.Status, s.UpdatedAt.UTC(), staff.Revision(s)}
}

type V1StaffDetailOutput struct{ Body V1StaffDetailBody }

//vkit:handler
func NewGetStaff(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1GetStaff", Method: http.MethodGet, Path: "/v1/staff/{id}", Errors: []int{401, 403, 409, 422, 503}}, func(ctx context.Context, in *struct {
		ID string `path:"id" format:"uuid"`
	}) (*V1StaffDetailOutput, error) {
		if _, err := b.Read(ctx, "superadmin"); err != nil {
			return nil, security.Error(ctx, err)
		}
		id, err := parseID(in.ID)
		if err != nil {
			return nil, security.Error(ctx, err)
		}
		r, err := persistence.New(b.Pool).BOStaffGet(ctx, id)
		if errors.Is(err, pgx.ErrNoRows) {
			err = staff.ErrConflict
		}
		if err != nil {
			return nil, security.Error(ctx, err)
		}
		return &V1StaffDetailOutput{Body: v1StaffDetailBody(staff.Staff{ID: r.ID, Email: r.EmailNormalized, DisplayName: r.DisplayName, Role: r.Role, Status: r.Status, UpdatedAt: r.UpdatedAt.Time})}, nil
	})
	return nil
}
