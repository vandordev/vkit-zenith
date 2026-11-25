package v1

import (
	"errors"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

func safeInputError(ctx huma.Context, status int, message string, errs ...error) huma.StatusError {
	if ctx != nil && strings.HasPrefix(ctx.URL().Path, "/v1/auth/") {
		return huma.NewError(status, message, errs...)
	}
	problem := &huma.ErrorModel{Type: "about:blank", Status: status, Title: http.StatusText(status), Detail: "Isian tidak valid."}
	for _, err := range errs {
		var detailer huma.ErrorDetailer
		if errors.As(err, &detailer) {
			detail := detailer.ErrorDetail()
			problem.Errors = append(problem.Errors, &huma.ErrorDetail{Location: safeInputLocation(detail.Location, detail.Message), Message: "Isian tidak valid."})
		}
	}
	return problem
}

func safeInputLocation(location, message string) string {
	const credentials = "body.credentials"
	if message == "unexpected property" {
		if strings.HasPrefix(location, credentials+".") {
			return credentials
		}
		if strings.HasPrefix(location, "body.") {
			return "body"
		}
	}
	if i := strings.Index(location, credentials+"."); i >= 0 {
		location = location[:i+len(credentials)]
	}
	return location
}

func init() {
	original := huma.NewErrorWithContext
	huma.NewErrorWithContext = func(ctx huma.Context, status int, message string, errs ...error) huma.StatusError {
		if (status == http.StatusBadRequest || status == http.StatusUnsupportedMediaType || status == http.StatusUnprocessableEntity) && ctx != nil && strings.HasPrefix(ctx.URL().Path, "/v1/") && !strings.HasPrefix(ctx.URL().Path, "/v1/auth/") {
			return safeInputError(ctx, status, message, errs...)
		}
		return original(ctx, status, message, errs...)
	}
}
