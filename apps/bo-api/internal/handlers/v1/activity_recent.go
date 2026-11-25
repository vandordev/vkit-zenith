package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/persistence"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
)

//vkit:handler
func NewGetRecentActivity(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1GetRecentActivity", Method: http.MethodGet, Path: "/v1/activity/recent", Errors: []int{401, 403, 503}}, func(ctx context.Context, _ *struct{}) (*V1ActivityRecentOutput, error) {
		q, err := activityRead(ctx, b)
		if err != nil {
			return nil, err
		}
		rows, err := q.BOActivityRecords(ctx, persistence.BOActivityRecordsParams{Actions: activityActions(), PageLimit: 5})
		if err != nil {
			return nil, catalogReadError(ctx, err)
		}
		out := &V1ActivityRecentOutput{Body: V1ActivityRecentBody{Items: make([]V1ActivityRow, 0, len(rows))}}
		for _, r := range rows {
			out.Body.Items = append(out.Body.Items, projectActivity(activityFromRow(r)))
		}
		return out, nil
	})
	return nil
}
