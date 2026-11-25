package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/persistence"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
)

//vkit:handler
func NewListActivityStaffChoices(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ListActivityStaffChoices", Method: http.MethodGet, Path: "/v1/activity/staff-choices", Errors: []int{401, 403, 422, 503}}, func(ctx context.Context, in *ActivityChoicesInput) (*V1ActivityStaffChoicesOutput, error) {
		q, tx, err := activitySnapshot(ctx, b)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback(ctx)
		id, err := optionalID(in.SelectedID)
		if err != nil {
			return nil, security.Error(ctx, err)
		}
		params := persistence.BOActivityStaffChoicesParams{Actions: activityActions(), Search: in.Q, PageLimit: int32(in.PageSize), PageOffset: int32((in.Page - 1) * in.PageSize)}
		rows, err := q.BOActivityStaffChoices(ctx, params)
		if err != nil {
			return nil, catalogReadError(ctx, err)
		}
		total, err := q.BOActivityStaffChoiceCount(ctx, persistence.BOActivityStaffChoiceCountParams{Actions: params.Actions, Search: in.Q})
		if err != nil {
			return nil, catalogReadError(ctx, err)
		}
		out := &V1ActivityStaffChoicesOutput{Body: V1ActivityStaffChoicesBody{Items: make([]V1ActivityStaffChoice, 0, len(rows)), Pagination: V1ActivityPagination{in.Page, in.PageSize, total}}}
		for _, r := range rows {
			out.Body.Items = append(out.Body.Items, V1ActivityStaffChoice{ID: r.ID.String(), Name: r.DisplayName, Status: r.Status})
		}
		if id.Valid {
			selected, err := q.BOActivityStaffChoices(ctx, persistence.BOActivityStaffChoicesParams{Actions: params.Actions, SelectedID: id, PageLimit: 1})
			if err != nil {
				return nil, catalogReadError(ctx, err)
			}
			if len(selected) == 1 {
				r := selected[0]
				out.Body.Selected = &V1ActivityStaffChoice{ID: r.ID.String(), Name: r.DisplayName, Status: r.Status}
			}
		}
		if err = tx.Commit(ctx); err != nil {
			return nil, catalogReadError(ctx, err)
		}
		return out, nil
	})
	return nil
}
