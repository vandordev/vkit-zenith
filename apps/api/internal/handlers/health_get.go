package handlers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type HealthOutput struct{ Body map[string]string }

//vkit:handler
func NewGetHealth(api huma.API) error {
	huma.Register(api, huma.Operation{
		OperationID: "getHealth", Method: http.MethodGet, Path: "/health",
		Hidden: true,
	}, func(ctx context.Context, input *struct{}) (*HealthOutput, error) {
		return &HealthOutput{Body: map[string]string{"status": "ok"}}, nil
	})
	return nil
}
