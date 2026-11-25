package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/fixtures"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"net/http"
	"strings"
)

// DUMMY (ADR-0001 phase 2): returns fixtures, persists nothing.
//
//vkit:handler
func NewListJobs(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ListJobs", Method: http.MethodGet, Path: "/v1/operations/jobs", Description: "[DUMMY] Antrean job River.", Errors: []int{401, 403, 422}}, func(ctx context.Context, in *V1JobsListInput) (*V1OpsListOutput[fixtures.Job], error) {
		if _, e := b.Read(ctx, "any_staff"); e != nil {
			return nil, security.Error(ctx, e)
		}
		out := make([]fixtures.Job, 0)
		for _, x := range fixtures.Jobs() {
			if (in.Status == "" || x.State == in.Status) && (in.StateGroup != "failed" || isFailedJobState(x.State)) && (in.Q == "" || strings.Contains(strings.ToLower(x.ID+" "+x.Kind+" "+x.State), strings.ToLower(in.Q))) {
				out = append(out, x)
			}
		}
		items, p := paginate(out, in.Page, in.PageSize)
		return &V1OpsListOutput[fixtures.Job]{Body: V1OpsListBody[fixtures.Job]{Items: items, Pagination: p}}, nil
	})
	return nil
}
