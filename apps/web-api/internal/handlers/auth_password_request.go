package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/identityhttp"
	"github.com/vandordev/vkit-zenith/packages/application/identity"
)

type authPasswordRequestInput struct {
	Origin string `header:"Origin"`
	CSRF   string `header:"X-CSRF-Token"`
	Body   struct {
		Email string `json:"email" minLength:"3" maxLength:"254"`
	}
}
type authPasswordRequestOutput struct {
	Status         int    `status:"202"`
	CacheControl   string `header:"Cache-Control"`
	Pragma         string `header:"Pragma"`
	ReferrerPolicy string `header:"Referrer-Policy"`
	Body           identity.PublicAccepted
}

//vkit:handler
func NewRequestPasswordAction(api huma.API, runtime *identityhttp.Runtime) error {
	huma.Register(api, huma.Operation{OperationID: "authV1RequestPasswordAction", Method: http.MethodPost, Path: "/v1/auth/password/request", DefaultStatus: http.StatusAccepted, MaxBodyBytes: 16 * 1024, Errors: []int{http.StatusForbidden, http.StatusTooManyRequests, http.StatusUnprocessableEntity, http.StatusServiceUnavailable}}, func(ctx context.Context, in *authPasswordRequestInput) (*authPasswordRequestOutput, error) {
		headers := http.Header{"Cache-Control": {"no-store"}, "Pragma": {"no-cache"}, "Referrer-Policy": {"no-referrer"}}
		fail := func(status int) error {
			var err error
			switch status {
			case 403:
				err = huma.Error403Forbidden(http.StatusText(status))
			case 422:
				err = huma.Error422UnprocessableEntity(http.StatusText(status))
			case 429:
				err = huma.Error429TooManyRequests(http.StatusText(status))
			default:
				err = huma.Error503ServiceUnavailable(http.StatusText(status))
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
		accepted, err := runtime.Service.RequestPasswordAction(ctx, in.Body.Email, identity.RequestMeta{Source: peer})
		if errors.Is(err, identity.ErrOverloaded) {
			return nil, fail(429)
		}
		if errors.Is(err, identity.ErrValidation) {
			return nil, fail(422)
		}
		if err != nil {
			return nil, fail(503)
		}
		return &authPasswordRequestOutput{Status: http.StatusAccepted, CacheControl: "no-store", Pragma: "no-cache", ReferrerPolicy: "no-referrer", Body: accepted}, nil
	})
	return nil
}
