package handlers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type HealthBody struct {
	Status string `json:"status" enum:"ok" doc:"Process health without infrastructure dependencies." example:"ok"`
}
type HealthOutput struct{ Body HealthBody }

//vkit:handler
func NewGetHealth(api huma.API) error {
	huma.Register(api, huma.Operation{
		OperationID: "getHealth", Method: http.MethodGet, Path: "/health",
		Summary: "Get process health", Description: "Returns process health without connecting to external infrastructure.",
		Tags: []string{"Operations"}, Errors: []int{http.StatusInternalServerError},
	}, func(ctx context.Context, input *struct{}) (*HealthOutput, error) {
		return &HealthOutput{Body: HealthBody{Status: "ok"}}, nil
	})
	return nil
}
