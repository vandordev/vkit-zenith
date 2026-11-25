package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/identityhttp"
	"github.com/vandordev/vkit-zenith/packages/application/identity"
)

type authRegisterInput struct {
	Origin string `header:"Origin"`
	CSRF   string `header:"X-CSRF-Token"`
	Body   struct {
		DisplayName string `json:"displayName" minLength:"1" maxLength:"200"`
		Email       string `json:"email" minLength:"3" maxLength:"254"`
		Telephone   string `json:"telephone" minLength:"3" maxLength:"64"`
		Password    string `json:"password" minLength:"15" maxLength:"512"`
	}
}
type authRegisterResponse struct {
	Status         int           `status:"202"`
	CacheControl   string        `header:"Cache-Control"`
	Pragma         string        `header:"Pragma"`
	ReferrerPolicy string        `header:"Referrer-Policy"`
	Cookies        []http.Cookie `header:"Set-Cookie"`
	Body           struct {
		Accepted bool `json:"accepted"`
	}
}

//vkit:handler
func NewRegisterCustomer(api huma.API, runtime *identityhttp.Runtime) error {
	huma.Register(api, huma.Operation{OperationID: "authV1Register", Method: http.MethodPost, Path: "/v1/auth/register", DefaultStatus: http.StatusAccepted, MaxBodyBytes: 16 * 1024, Errors: []int{http.StatusForbidden, http.StatusRequestEntityTooLarge, http.StatusTooManyRequests, http.StatusUnprocessableEntity, http.StatusServiceUnavailable}}, func(ctx context.Context, in *authRegisterInput) (*authRegisterResponse, error) {
		headers := http.Header{"Cache-Control": {"no-store"}, "Pragma": {"no-cache"}, "Referrer-Policy": {"no-referrer"}}
		fail := func(status int) error {
			var err error
			switch status {
			case http.StatusForbidden:
				err = huma.Error403Forbidden(http.StatusText(status))
			case http.StatusTooManyRequests:
				err = huma.Error429TooManyRequests(http.StatusText(status))
			case http.StatusUnprocessableEntity:
				err = huma.Error422UnprocessableEntity(http.StatusText(status))
			default:
				err = huma.Error503ServiceUnavailable(http.StatusText(status))
			}
			return huma.ErrorWithHeaders(err, headers)
		}
		if runtime == nil || runtime.Service == nil {
			return nil, fail(http.StatusServiceUnavailable)
		}
		if in.Origin == "" || in.Origin != runtime.Origin {
			return nil, fail(http.StatusForbidden)
		}
		bindings := identityhttp.RequestBrowserBindings(ctx)
		if !runtime.Service.ValidateCSRF(bindings.Session, in.CSRF) {
			return nil, fail(http.StatusForbidden)
		}
		peer := identityhttp.RequestSource(ctx)
		if peer == "" {
			return nil, fail(http.StatusForbidden)
		}
		result, err := runtime.Service.Register(ctx, identity.RegisterInput{DisplayName: in.Body.DisplayName, Email: in.Body.Email, Telephone: in.Body.Telephone, Password: in.Body.Password}, identity.RequestMeta{Source: peer})
		if errors.Is(err, identity.ErrOverloaded) {
			return nil, fail(http.StatusTooManyRequests)
		}
		if errors.Is(err, identity.ErrValidation) {
			return nil, fail(http.StatusUnprocessableEntity)
		}
		if err != nil {
			return nil, fail(http.StatusServiceUnavailable)
		}
		secure := !runtime.Development
		base := http.Cookie{Path: "/", HttpOnly: true, Secure: secure, SameSite: http.SameSiteLaxMode, MaxAge: int(identity.RestrictedSessionLifetime.Seconds())}
		enrollment := base
		enrollment.Name = "__Host-zenith_enrollment"
		if runtime.Development {
			enrollment.Name = "zenith_enrollment_dev"
		}
		enrollment.Value = result.EnrollmentSecret
		session := base
		session.Name = "__Host-zenith_auth"
		if runtime.Development {
			session.Name = "zenith_auth_dev"
		}
		session.Value = result.SessionSecret
		// Both cookie shapes are identical on new/duplicate normalized addresses;
		// an unbound duplicate receipt/session has no account authority.
		return &authRegisterResponse{Status: http.StatusAccepted, CacheControl: "no-store", Pragma: "no-cache", ReferrerPolicy: "no-referrer", Cookies: []http.Cookie{enrollment, session}, Body: struct {
			Accepted bool `json:"accepted"`
		}{Accepted: true}}, nil
	})
	return nil
}
