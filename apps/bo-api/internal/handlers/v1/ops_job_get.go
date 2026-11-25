package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/fixtures"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"net/http"
)

// DUMMY (ADR-0001 phase 2): returns fixtures, persists nothing.
//
//vkit:handler
func NewGetJob(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1GetJob", Method: http.MethodGet, Path: "/v1/operations/jobs/{id}", Description: "[DUMMY] Detail job River.", Errors: []int{401, 403, 404}}, func(ctx context.Context, in *V1OpsGetInput) (*V1OpsDetailOutput[fixtures.Job], error) {
		if _, e := b.Read(ctx, "any_staff"); e != nil {
			return nil, security.Error(ctx, e)
		}
		x, h, ok := opsFindJob(in.ID)
		if !ok {
			return nil, dummyNotFound()
		}
		return &V1OpsDetailOutput[fixtures.Job]{Body: V1OpsDetail[fixtures.Job]{Item: x, History: h}}, nil
	})
	return nil
}
