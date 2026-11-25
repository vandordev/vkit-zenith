package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/identityhttp"
	"github.com/vandordev/vkit-zenith/packages/application/identity"
)

type authActionInspectInput struct {
	Token string `query:"token" minLength:"1" maxLength:"128"`
}
type authActionInspectOutput struct {
	CacheControl   string `header:"Cache-Control"`
	Pragma         string `header:"Pragma"`
	ReferrerPolicy string `header:"Referrer-Policy"`
	Body           struct {
		Purpose       identity.ActionPurpose `json:"purpose" enum:"verify_email,setup_password,reset_password"`
		ExpiresAt     string                 `json:"expiresAt"`
		RequiresLogin bool                   `json:"requiresLogin"`
		DisplayName   string                 `json:"displayName,omitempty"`
	}
}

//vkit:handler
func NewInspectEmailAction(api huma.API, runtime *identityhttp.Runtime) error {
	huma.Register(api, huma.Operation{OperationID: "authV1InspectEmailAction", Method: http.MethodGet, Path: "/v1/auth/action/inspect", DefaultStatus: http.StatusOK, Errors: []int{http.StatusNotFound, http.StatusServiceUnavailable}}, func(ctx context.Context, in *authActionInspectInput) (*authActionInspectOutput, error) {
		headers := http.Header{"Cache-Control": {"no-store"}, "Pragma": {"no-cache"}, "Referrer-Policy": {"no-referrer"}}
		if runtime == nil || runtime.Service == nil {
			return nil, huma.ErrorWithHeaders(huma.Error503ServiceUnavailable("Service unavailable"), headers)
		}
		bindings := identityhttp.RequestBrowserBindings(ctx)
		view, err := runtime.Service.InspectAction(ctx, identity.RedeemActionInput{Token: in.Token, SessionHash: bindings.Session, EnrollmentHash: bindings.Enrollment})
		if err != nil {
			return nil, huma.ErrorWithHeaders(huma.Error404NotFound("Link tidak tersedia"), headers)
		}
		return &authActionInspectOutput{CacheControl: "no-store", Pragma: "no-cache", ReferrerPolicy: "no-referrer", Body: struct {
			Purpose       identity.ActionPurpose `json:"purpose" enum:"verify_email,setup_password,reset_password"`
			ExpiresAt     string                 `json:"expiresAt"`
			RequiresLogin bool                   `json:"requiresLogin"`
			DisplayName   string                 `json:"displayName,omitempty"`
		}{Purpose: view.Purpose, ExpiresAt: view.ExpiresAt.UTC().Format(time.RFC3339), RequiresLogin: view.RequiresLogin, DisplayName: view.DisplayName}}, nil
	})
	return nil
}
