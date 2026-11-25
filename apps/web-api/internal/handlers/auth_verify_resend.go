package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/identityhttp"
	"github.com/vandordev/vkit-zenith/packages/application/identity"
)

type authVerifyResendInput struct {
	Origin string `header:"Origin"`
	CSRF   string `header:"X-CSRF-Token"`
}
type authVerifyResendOutput struct {
	Status         int    `status:"202"`
	CacheControl   string `header:"Cache-Control"`
	Pragma         string `header:"Pragma"`
	ReferrerPolicy string `header:"Referrer-Policy"`
	Body           identity.PublicAccepted
}

//vkit:handler
func NewResendVerification(api huma.API, runtime *identityhttp.Runtime) error {
	huma.Register(api, huma.Operation{OperationID: "authV1ResendVerification", Method: http.MethodPost, Path: "/v1/auth/verification/resend", DefaultStatus: http.StatusAccepted, Errors: []int{http.StatusForbidden, http.StatusTooManyRequests, http.StatusServiceUnavailable}}, func(ctx context.Context, in *authVerifyResendInput) (*authVerifyResendOutput, error) {
		headers := http.Header{"Cache-Control": {"no-store"}, "Pragma": {"no-cache"}, "Referrer-Policy": {"no-referrer"}}
		fail := func(code int) error {
			var err error
			switch code {
			case 403:
				err = huma.Error403Forbidden(http.StatusText(code))
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
		bindings := identityhttp.RequestBrowserBindings(ctx)
		if in.Origin == "" || in.Origin != runtime.Origin || !runtime.Service.ValidateSessionCSRF(bindings.Session, in.CSRF) {
			return nil, fail(403)
		}
		peer := identityhttp.RequestSource(ctx)
		if peer == "" {
			return nil, fail(403)
		}
		sessionHash, enrollmentHash := bindings.Session, bindings.Enrollment
		if sessionHash == ([32]byte{}) || enrollmentHash == ([32]byte{}) {
			return nil, fail(403)
		}
		accepted, err := runtime.Service.RequestVerification(ctx, identity.RequestMeta{Source: peer, SessionHash: sessionHash, EnrollmentHash: enrollmentHash})
		if errors.Is(err, identity.ErrOverloaded) {
			return nil, fail(429)
		}
		if err != nil {
			return nil, fail(503)
		}
		return &authVerifyResendOutput{Status: http.StatusAccepted, CacheControl: "no-store", Pragma: "no-cache", ReferrerPolicy: "no-referrer", Body: accepted}, nil
	})
	return nil
}
