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

type authActionConfirmInput struct {
	Origin string `header:"Origin"`
	CSRF   string `header:"X-CSRF-Token"`
	Body   struct {
		Token     string  `json:"token" minLength:"1" maxLength:"128"`
		Password  *string `json:"password,omitempty" maxLength:"512"`
		Telephone *string `json:"telephone,omitempty" maxLength:"64"`
	}
}
type authActionConfirmOutput struct {
	CacheControl   string        `header:"Cache-Control"`
	Pragma         string        `header:"Pragma"`
	ReferrerPolicy string        `header:"Referrer-Policy"`
	Cookies        []http.Cookie `header:"Set-Cookie"`
	Body           struct {
		Completed          bool   `json:"completed"`
		RequiresFreshLogin bool   `json:"requiresFreshLogin"`
		CSRFToken          string `json:"csrfToken,omitempty"`
	}
}

//vkit:handler
func NewConfirmEmailAction(api huma.API, runtime *identityhttp.Runtime) error {
	huma.Register(api, huma.Operation{OperationID: "authV1ConfirmEmailAction", Method: http.MethodPost, Path: "/v1/auth/action/confirm", DefaultStatus: http.StatusOK, MaxBodyBytes: 16 * 1024, Errors: []int{http.StatusForbidden, http.StatusNotFound, http.StatusUnprocessableEntity, http.StatusServiceUnavailable}}, func(ctx context.Context, in *authActionConfirmInput) (*authActionConfirmOutput, error) {
		headers := http.Header{"Cache-Control": {"no-store"}, "Pragma": {"no-cache"}, "Referrer-Policy": {"no-referrer"}}
		fail := func(code int) error {
			var err error
			switch code {
			case 403:
				err = huma.Error403Forbidden(http.StatusText(code))
			case 404:
				err = huma.Error404NotFound("Link tidak tersedia")
			case 422:
				err = huma.Error422UnprocessableEntity(http.StatusText(code))
			default:
				err = huma.Error503ServiceUnavailable(http.StatusText(code))
			}
			return huma.ErrorWithHeaders(err, headers)
		}
		if runtime == nil || runtime.Service == nil {
			return nil, fail(503)
		}
		bindings := identityhttp.RequestBrowserBindings(ctx)
		if in.Origin == "" || in.Origin != runtime.Origin || !runtime.Service.ValidateCSRF(bindings.Session, in.CSRF) {
			return nil, fail(403)
		}
		password, telephone := "", ""
		if in.Body.Password != nil {
			password = *in.Body.Password
		}
		if in.Body.Telephone != nil {
			telephone = *in.Body.Telephone
		}
		result, err := runtime.Service.RedeemAction(ctx, identity.RedeemActionInput{Token: in.Body.Token, Password: password, Telephone: telephone, SessionHash: bindings.Session, EnrollmentHash: bindings.Enrollment})
		if errors.Is(err, identity.ErrValidation) {
			return nil, fail(422)
		}
		if errors.Is(err, identity.ErrDenied) || errors.Is(err, identity.ErrNotFound) {
			return nil, fail(404)
		}
		if err != nil {
			return nil, fail(503)
		}
		out := &authActionConfirmOutput{CacheControl: "no-store", Pragma: "no-cache", ReferrerPolicy: "no-referrer"}
		if result.RequiresFreshLogin {
			out.Cookies = []http.Cookie{clearingAuthCookie(runtime.Development), clearingEnrollmentCookie(runtime.Development)}
		} else {
			out.Cookies = []http.Cookie{authSessionCookie(result.SessionSecret, result.SessionExpiresAt, result.CookieMaxAge, runtime.Development), clearingEnrollmentCookie(runtime.Development)}
			out.Body.CSRFToken = result.CSRFToken
		}
		out.Body.Completed = true
		out.Body.RequiresFreshLogin = result.RequiresFreshLogin
		return out, nil
	})
	return nil
}

func authSessionCookie(secret string, expires time.Time, maxAge int, development bool) http.Cookie {
	name := "__Host-zenith_auth"
	if development {
		name = "zenith_auth_dev"
	}
	return http.Cookie{Name: name, Value: secret, Path: "/", HttpOnly: true, Secure: !development, SameSite: http.SameSiteLaxMode, Expires: expires, MaxAge: maxAge}
}
func clearingAuthCookie(development bool) http.Cookie {
	c := authSessionCookie("", time.Unix(1, 0), -1, development)
	c.MaxAge = -1
	return c
}
func clearingEnrollmentCookie(development bool) http.Cookie {
	name := "__Host-zenith_enrollment"
	if development {
		name = "zenith_enrollment_dev"
	}
	return http.Cookie{Name: name, Value: "", Path: "/", HttpOnly: true, Secure: !development, SameSite: http.SameSiteLaxMode, MaxAge: -1}
}
