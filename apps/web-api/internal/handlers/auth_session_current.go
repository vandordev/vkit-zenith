package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/identityhttp"
	"github.com/vandordev/vkit-zenith/packages/application/identity"
)

type authSessionCurrentInput struct {
	Origin string `header:"Origin"`
}
type authSessionCurrentOutput struct {
	CacheControl   string `header:"Cache-Control"`
	Pragma         string `header:"Pragma"`
	ReferrerPolicy string `header:"Referrer-Policy"`
	Body           struct {
		Kind            identity.SessionKind `json:"kind" enum:"restricted,full"`
		DisplayName     string               `json:"displayName"`
		Email           string               `json:"email"`
		EmailVerified   bool                 `json:"emailVerified"`
		AuthenticatedAt string               `json:"authenticatedAt"`
		ExpiresAt       string               `json:"expiresAt"`
	}
}

//vkit:handler
func NewGetCurrentCustomerSession(api huma.API, runtime *identityhttp.Runtime) error {
	huma.Register(api, huma.Operation{OperationID: "authV1CurrentSession", Method: http.MethodGet, Path: "/v1/auth/session/current", DefaultStatus: http.StatusOK, Errors: []int{http.StatusForbidden, http.StatusUnauthorized, http.StatusServiceUnavailable}}, func(ctx context.Context, in *authSessionCurrentInput) (*authSessionCurrentOutput, error) {
		headers := http.Header{"Cache-Control": {"no-store"}, "Pragma": {"no-cache"}, "Referrer-Policy": {"no-referrer"}}
		fail := func(code int) error {
			var e error
			if code == 403 {
				e = huma.Error403Forbidden(http.StatusText(code))
			} else if code == 401 {
				e = huma.Error401Unauthorized("Session tidak tersedia")
			} else {
				e = huma.Error503ServiceUnavailable("Service unavailable")
			}
			return huma.ErrorWithHeaders(e, headers)
		}
		if runtime == nil || runtime.Service == nil {
			return nil, fail(503)
		}
		if in.Origin == "" || in.Origin != runtime.Origin {
			return nil, fail(403)
		}
		bindings := identityhttp.RequestBrowserBindings(ctx)
		if bindings.Session == ([32]byte{}) {
			return nil, fail(401)
		}
		view, err := runtime.Service.Current(ctx, bindings.Session)
		if errors.Is(err, identity.ErrUnauthenticated) || errors.Is(err, identity.ErrDenied) {
			return nil, fail(401)
		}
		if err != nil {
			return nil, fail(503)
		}
		return &authSessionCurrentOutput{CacheControl: "no-store", Pragma: "no-cache", ReferrerPolicy: "no-referrer", Body: struct {
			Kind            identity.SessionKind `json:"kind" enum:"restricted,full"`
			DisplayName     string               `json:"displayName"`
			Email           string               `json:"email"`
			EmailVerified   bool                 `json:"emailVerified"`
			AuthenticatedAt string               `json:"authenticatedAt"`
			ExpiresAt       string               `json:"expiresAt"`
		}{Kind: view.Kind, DisplayName: view.DisplayName, Email: view.Email, EmailVerified: view.EmailVerified, AuthenticatedAt: view.AuthenticatedAt.UTC().Format(time.RFC3339Nano), ExpiresAt: view.IdleExpiresAt.UTC().Format(time.RFC3339Nano)}}, nil
	})
	return nil
}
