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
func NewCancelJob(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1CancelJob", Method: http.MethodPost, Path: "/v1/operations/jobs/{id}/cancel", Description: "[DUMMY] Simulasi pembatalan job.", Errors: []int{401, 403, 404, 409, 422}}, func(ctx context.Context, in *V1OpsActionInput) (*V1OpsActionOutput[fixtures.Job], error) {
		if _, e := b.Read(ctx, "any_staff"); e != nil {
			return nil, security.Error(ctx, e)
		}
		if len([]rune(in.Body.Reason)) < 10 || len([]rune(in.Body.Reason)) > 1024 {
			return nil, dummyValidation()
		}
		x, _, ok := opsFindJob(in.ID)
		if !ok {
			return nil, dummyNotFound()
		}
		if x.State != "available" && x.State != "scheduled" && x.State != "retryable" {
			return nil, stateConflict()
		}
		x.State = "cancelled"
		return &V1OpsActionOutput[fixtures.Job]{Body: V1DummyResult[fixtures.Job]{Result: x, Dummy: true}}, nil
	})
	return nil
}
