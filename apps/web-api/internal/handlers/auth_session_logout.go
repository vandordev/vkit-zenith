package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/identityhttp"
	"github.com/vandordev/vkit-zenith/packages/application/identity"
)

type authLogoutInput struct {
	Origin string `header:"Origin"`
	CSRF   string `header:"X-CSRF-Token"`
}
type authLogoutOutput struct {
	CacheControl   string        `header:"Cache-Control"`
	Pragma         string        `header:"Pragma"`
	ReferrerPolicy string        `header:"Referrer-Policy"`
	Cookies        []http.Cookie `header:"Set-Cookie"`
	Body           struct {
		LoggedOut bool `json:"loggedOut"`
	}
}

//vkit:handler
func NewLogoutCustomer(api huma.API, runtime *identityhttp.Runtime) error {
	huma.Register(api, huma.Operation{OperationID: "authV1Logout", Method: http.MethodPost, Path: "/v1/auth/session/logout", DefaultStatus: http.StatusOK, Errors: []int{http.StatusForbidden, http.StatusServiceUnavailable}}, func(ctx context.Context, in *authLogoutInput) (*authLogoutOutput, error) {
		headers := http.Header{"Cache-Control": {"no-store"}, "Pragma": {"no-cache"}, "Referrer-Policy": {"no-referrer"}}
		fail := func(code int) error {
			var e error
			if code == 403 {
				e = huma.Error403Forbidden(http.StatusText(code))
			} else {
				e = huma.Error503ServiceUnavailable("Service unavailable")
			}
			return huma.ErrorWithHeaders(e, headers)
		}
		if runtime == nil || runtime.Service == nil {
			return nil, fail(503)
		}
		bindings := identityhttp.RequestBrowserBindings(ctx)
		if in.Origin == "" || in.Origin != runtime.Origin || !runtime.Service.ValidateSessionCSRF(bindings.Session, in.CSRF) {
			return nil, fail(403)
		}
		if err := runtime.Service.Logout(ctx, bindings.Session); err != nil && !errors.Is(err, identity.ErrUnauthenticated) {
			return nil, fail(503)
		}
		return &authLogoutOutput{CacheControl: "no-store", Pragma: "no-cache", ReferrerPolicy: "no-referrer", Cookies: []http.Cookie{clearingAuthCookie(runtime.Development), clearingEnrollmentCookie(runtime.Development)}, Body: struct {
			LoggedOut bool `json:"loggedOut"`
		}{LoggedOut: true}}, nil
	})
	return nil
}
