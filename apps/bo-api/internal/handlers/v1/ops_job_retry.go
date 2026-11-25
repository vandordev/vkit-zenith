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
func NewRetryJob(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1RetryJob", Method: http.MethodPost, Path: "/v1/operations/jobs/{id}/retry", Description: "[DUMMY] Simulasi pengulangan job.", Errors: []int{401, 403, 404, 409, 422}}, func(ctx context.Context, in *V1OpsActionInput) (*V1OpsActionOutput[fixtures.Job], error) {
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
		if x.State != "retryable" && x.State != "discarded" {
			return nil, stateConflict()
		}
		x.State = "available"
		return &V1OpsActionOutput[fixtures.Job]{Body: V1DummyResult[fixtures.Job]{Result: x, Dummy: true}}, nil
	})
	return nil
}
