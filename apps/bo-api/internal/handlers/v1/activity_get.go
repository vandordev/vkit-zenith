package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/persistence"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

//vkit:handler
func NewGetActivity(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1GetActivity", Method: http.MethodGet, Path: "/v1/activity/{id}", Errors: []int{401, 403, 409, 422, 503}}, func(ctx context.Context, in *ItemPath) (*V1ActivityDetailOutput, error) {
		q, err := activityRead(ctx, b)
		if err != nil {
			return nil, err
		}
		id, err := parseID(in.ID)
		if err != nil {
			return nil, security.Error(ctx, err)
		}
		rows, err := q.BOActivityRecords(ctx, persistence.BOActivityRecordsParams{Actions: activityActions(), EventID: id, PageLimit: 1, IncludeDetail: true})
		if err != nil {
			return nil, catalogReadError(ctx, err)
		}
		if len(rows) != 1 {
			return nil, security.Error(ctx, staff.ErrConflict)
		}
		return &V1ActivityDetailOutput{Body: projectActivityDetail(activityFromRow(rows[0]))}, nil
	})
	return nil
}
