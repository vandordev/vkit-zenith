package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/fixtures"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"net/http"
	"strconv"
)

// DUMMY (ADR-0001 phase 2): returns fixtures, persists nothing.
//
//vkit:handler
func NewGetHomeCounters(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1GetHomeCounters", Method: http.MethodGet, Path: "/v1/home/counters", Description: "[DUMMY] Penghitung operasional Beranda.", Errors: []int{401, 403}}, func(ctx context.Context, _ *struct{}) (*V1HomeCountersOutput, error) {
		if _, e := b.Read(ctx, "any_staff"); e != nil {
			return nil, security.Error(ctx, e)
		}
		jobs := 0
		for _, j := range fixtures.Jobs() {
			if isFailedJobState(j.State) {
				jobs++
			}
		}
		counter := func(n int, href string) V1HomeCounter { return V1HomeCounter{Count: strconv.Itoa(n), Href: href} }
		return &V1HomeCountersOutput{Body: V1HomeCountersBody{FailedJobs: counter(jobs, "/operations/jobs?stateGroup=failed")}}, nil
	})
	return nil
}
