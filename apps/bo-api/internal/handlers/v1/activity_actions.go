package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
)

//vkit:handler
func NewListActivityActions(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ListActivityActions", Method: http.MethodGet, Path: "/v1/activity/actions", Errors: []int{401, 403, 503}}, func(ctx context.Context, _ *struct{}) (*V1ActivityActionsOutput, error) {
		if _, err := activityRead(ctx, b); err != nil {
			return nil, err
		}
		out := &V1ActivityActionsOutput{Body: V1ActivityActionsBody{Items: make([]V1ActivityAction, 0, len(activityPolicies))}}
		for _, p := range activityPolicies {
			out.Body.Items = append(out.Body.Items, V1ActivityAction{Value: p.action, Label: p.label})
		}
		return out, nil
	})
	return nil
}
