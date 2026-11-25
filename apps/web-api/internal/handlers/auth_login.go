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

type authLoginInput struct {
	Origin string `header:"Origin"`
	CSRF   string `header:"X-CSRF-Token"`
	Body   struct {
		Email    string `json:"email" minLength:"3" maxLength:"254"`
		Password string `json:"password" minLength:"15" maxLength:"512"`
	}
}
type authLoginOutput struct {
	CacheControl   string        `header:"Cache-Control"`
	Pragma         string        `header:"Pragma"`
	ReferrerPolicy string        `header:"Referrer-Policy"`
	Cookies        []http.Cookie `header:"Set-Cookie"`
	Body           struct {
		CSRFToken       string               `json:"csrfToken"`
		Kind            identity.SessionKind `json:"kind" enum:"restricted,full"`
		DisplayName     string               `json:"displayName"`
		Email           string               `json:"email"`
		EmailVerified   bool                 `json:"emailVerified"`
		AuthenticatedAt string               `json:"authenticatedAt"`
		ExpiresAt       string               `json:"expiresAt"`
	}
}

//vkit:handler
func NewLoginCustomer(api huma.API, runtime *identityhttp.Runtime) error {
	huma.Register(api, huma.Operation{OperationID: "authV1Login", Method: http.MethodPost, Path: "/v1/auth/login", DefaultStatus: http.StatusOK, MaxBodyBytes: 16 * 1024, Errors: []int{http.StatusForbidden, http.StatusTooManyRequests, http.StatusUnauthorized, http.StatusUnprocessableEntity, http.StatusServiceUnavailable}}, func(ctx context.Context, in *authLoginInput) (*authLoginOutput, error) {
		headers := http.Header{"Cache-Control": {"no-store"}, "Pragma": {"no-cache"}, "Referrer-Policy": {"no-referrer"}}
		fail := func(code int) error {
			var err error
			switch code {
			case 403:
				err = huma.Error403Forbidden(http.StatusText(code))
			case 401:
				err = huma.Error401Unauthorized("Kredensial tidak valid")
			case 422:
				err = huma.Error422UnprocessableEntity(http.StatusText(code))
			case 429:
				err = huma.Error429TooManyRequests(http.StatusText(code))
			default:
				err = huma.Error503ServiceUnavailable(http.StatusText(code))
			}
			return huma.ErrorWithHeaders(err, headers)
		}
		if runtime == nil || runtime.Service == nil {
			return nil, fail(503)
		}
		if in.Origin == "" || in.Origin != runtime.Origin {
			return nil, fail(403)
		}
		bindings := identityhttp.RequestBrowserBindings(ctx)
		if !runtime.Service.ValidateCSRF(bindings.Session, in.CSRF) {
			return nil, fail(403)
		}
		peer := identityhttp.RequestSource(ctx)
		if peer == "" {
			return nil, fail(403)
		}
		issued, err := runtime.Service.Login(ctx, identity.LoginInput{Email: in.Body.Email, Password: in.Body.Password}, identity.RequestMeta{Source: peer, SessionHash: bindings.Session})
		if errors.Is(err, identity.ErrValidation) {
			return nil, fail(422)
		}
		if errors.Is(err, identity.ErrUnauthenticated) || errors.Is(err, identity.ErrDenied) {
			return nil, fail(401)
		}
		if errors.Is(err, identity.ErrOverloaded) {
			return nil, fail(429)
		}
		if err != nil {
			return nil, fail(503)
		}
		name := "__Host-zenith_auth"
		enrollmentName := "__Host-zenith_enrollment"
		if runtime.Development {
			name = "zenith_auth_dev"
			enrollmentName = "zenith_enrollment_dev"
		}
		cookie := http.Cookie{Name: name, Value: issued.Secret, Path: "/", HttpOnly: true, Secure: !runtime.Development, SameSite: http.SameSiteLaxMode, Expires: issued.ExpiresAt, MaxAge: issued.CookieMaxAge}
		cookies := []http.Cookie{cookie}
		if issued.View.Kind == identity.SessionRestricted && issued.EnrollmentSecret != "" {
			cookies = append(cookies, http.Cookie{Name: enrollmentName, Value: issued.EnrollmentSecret, Path: "/", HttpOnly: true, Secure: !runtime.Development, SameSite: http.SameSiteLaxMode, MaxAge: int(identity.RestrictedSessionLifetime.Seconds())})
		} else {
			cookies = append(cookies, clearingEnrollmentCookie(runtime.Development))
		}
		return &authLoginOutput{CacheControl: "no-store", Pragma: "no-cache", ReferrerPolicy: "no-referrer", Cookies: cookies, Body: struct {
			CSRFToken       string               `json:"csrfToken"`
			Kind            identity.SessionKind `json:"kind" enum:"restricted,full"`
			DisplayName     string               `json:"displayName"`
			Email           string               `json:"email"`
			EmailVerified   bool                 `json:"emailVerified"`
			AuthenticatedAt string               `json:"authenticatedAt"`
			ExpiresAt       string               `json:"expiresAt"`
		}{CSRFToken: issued.CSRFToken, Kind: issued.View.Kind, DisplayName: issued.View.DisplayName, Email: issued.View.Email, EmailVerified: issued.View.EmailVerified, AuthenticatedAt: issued.View.AuthenticatedAt.UTC().Format(time.RFC3339Nano), ExpiresAt: issued.ExpiresAt.UTC().Format(time.RFC3339Nano)}}, nil
	})
	return nil
}
