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

type authSessionTouchInput struct {
	Origin string `header:"Origin"`
	CSRF   string `header:"X-CSRF-Token"`
}
type authSessionTouchOutput struct {
	CacheControl   string      `header:"Cache-Control"`
	Pragma         string      `header:"Pragma"`
	ReferrerPolicy string      `header:"Referrer-Policy"`
	Cookie         http.Cookie `header:"Set-Cookie"`
	Body           struct {
		ExpiresAt string `json:"expiresAt"`
	}
}

//vkit:handler
func NewTouchCustomerSession(api huma.API, runtime *identityhttp.Runtime) error {
	huma.Register(api, huma.Operation{OperationID: "authV1TouchSession", Method: http.MethodPost, Path: "/v1/auth/session/touch", DefaultStatus: http.StatusOK, Errors: []int{http.StatusForbidden, http.StatusTooManyRequests, http.StatusUnauthorized, http.StatusServiceUnavailable}}, func(ctx context.Context, in *authSessionTouchInput) (*authSessionTouchOutput, error) {
		headers := http.Header{"Cache-Control": {"no-store"}, "Pragma": {"no-cache"}, "Referrer-Policy": {"no-referrer"}}
		fail := func(code int) error {
			var e error
			switch code {
			case 403:
				e = huma.Error403Forbidden(http.StatusText(code))
			case 401:
				e = huma.Error401Unauthorized("Session tidak tersedia")
			case 429:
				e = huma.Error429TooManyRequests(http.StatusText(code))
			default:
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
		deadline, maxAge, err := runtime.Service.TouchFullSession(ctx, bindings.Session)
		if errors.Is(err, identity.ErrUnauthenticated) {
			return nil, fail(401)
		}
		if errors.Is(err, identity.ErrOverloaded) {
			return nil, fail(429)
		}
		if err != nil {
			return nil, fail(503)
		}
		name := "__Host-zenith_auth"
		if runtime.Development {
			name = "zenith_auth_dev"
		}
		if bindings.SessionSecret == "" {
			return nil, fail(401)
		}
		return &authSessionTouchOutput{CacheControl: "no-store", Pragma: "no-cache", ReferrerPolicy: "no-referrer", Cookie: http.Cookie{Name: name, Value: bindings.SessionSecret, Path: "/", HttpOnly: true, Secure: !runtime.Development, SameSite: http.SameSiteLaxMode, Expires: deadline, MaxAge: maxAge}, Body: struct {
			ExpiresAt string `json:"expiresAt"`
		}{ExpiresAt: deadline.UTC().Format(time.RFC3339Nano)}}, nil
	})
	return nil
}
