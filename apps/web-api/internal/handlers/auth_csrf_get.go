package handlers

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/identityhttp"
)

type authCSRFCookieOutput struct {
	CacheControl   string `header:"Cache-Control"`
	Pragma         string `header:"Pragma"`
	ReferrerPolicy string `header:"Referrer-Policy"`
	Body           struct {
		Token string `json:"token"`
	}
}

//vkit:handler
func NewGetAuthCSRF(api huma.API, runtime *identityhttp.Runtime) error {
	huma.Register(api, huma.Operation{OperationID: "authV1GetCSRF", Method: http.MethodGet, Path: "/v1/auth/csrf", DefaultStatus: http.StatusOK, Errors: []int{http.StatusForbidden, http.StatusServiceUnavailable}}, func(ctx context.Context, in *struct {
		Origin string `header:"Origin"`
	}) (*authCSRFCookieOutput, error) {
		if runtime == nil || runtime.Service == nil {
			return nil, huma.Error503ServiceUnavailable("Service unavailable")
		}
		if in.Origin == "" || in.Origin != runtime.Origin {
			return nil, huma.Error403Forbidden("Origin forbidden")
		}
		bindings := identityhttp.RequestBrowserBindings(ctx)
		token, err := runtime.Service.NewAnonymousCSRF()
		if bindings.Session != ([32]byte{}) {
			token, err = runtime.Service.NewSessionCSRF(bindings.Session)
		}
		if err != nil {
			return nil, huma.Error503ServiceUnavailable("Service unavailable")
		}
		return &authCSRFCookieOutput{CacheControl: "no-store", Pragma: "no-cache", ReferrerPolicy: "no-referrer", Body: struct {
			Token string `json:"token"`
		}{Token: token}}, nil
	})
	return nil
}
