package handlers

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"net/http"
)

type HealthBody struct {
	Status string `json:"status" enum:"ok" doc:"Process health, not database readiness."`
}
type HealthOutput struct{ Body HealthBody }

//vkit:handler
func NewGetHealth(api huma.API) error {
	huma.Register(api, huma.Operation{OperationID: "getHealth", Method: http.MethodGet, Path: "/health", Summary: "Get process health", Tags: []string{"Operations"}}, func(_ context.Context, _ *struct{}) (*HealthOutput, error) {
		return &HealthOutput{Body: HealthBody{Status: "ok"}}, nil
	})
	return nil
}
