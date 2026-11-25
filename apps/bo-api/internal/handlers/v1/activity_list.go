package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
)

//vkit:handler
func NewListActivity(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ListActivity", Method: http.MethodGet, Path: "/v1/activity", Errors: []int{401, 403, 422, 503}}, func(ctx context.Context, in *ActivityInput) (*V1ActivityListOutput, error) {
		q, tx, err := activitySnapshot(ctx, b)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback(ctx)
		params, err := activityFilters(in)
		if err != nil {
			return nil, security.Error(ctx, err)
		}
		rows, err := q.BOActivityRecords(ctx, params)
		if err != nil {
			return nil, catalogReadError(ctx, err)
		}
		total, err := q.BOActivityCount(ctx, activityCountFilters(params))
		if err != nil {
			return nil, catalogReadError(ctx, err)
		}
		out := &V1ActivityListOutput{Body: V1ActivityListBody{Items: make([]V1ActivityRow, 0, len(rows)), Pagination: V1ActivityPagination{in.Page, in.PageSize, total}}}
		for _, r := range rows {
			out.Body.Items = append(out.Body.Items, projectActivity(activityFromRow(r)))
		}
		if err = tx.Commit(ctx); err != nil {
			return nil, catalogReadError(ctx, err)
		}
		return out, nil
	})
	return nil
}
