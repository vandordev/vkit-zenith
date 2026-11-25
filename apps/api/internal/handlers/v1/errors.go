package v1

import (
	"net/http"
	"reflect"
	"slices"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

type problem struct {
	*huma.ErrorModel
	Code string `json:"code"`
}

func (p *problem) Error() string  { return p.Detail }
func (p *problem) GetStatus() int { return p.Status }

func Problem(status int, code string) huma.StatusError {
	return &problem{ErrorModel: &huma.ErrorModel{Type: "about:blank", Title: http.StatusText(status), Status: status, Detail: http.StatusText(status)}, Code: code}
}

func safeInputError(_ huma.Context, status int, _ ...error) huma.StatusError {
	base := huma.NewError(status, "Invalid input.")
	model := base.(*huma.ErrorModel)
	model.Detail = "Invalid input."
	model.Errors = nil
	return &problem{ErrorModel: model, Code: "validation_failed"}
}

func init() {
	original := huma.NewErrorWithContext
	huma.NewErrorWithContext = func(ctx huma.Context, status int, message string, errs ...error) huma.StatusError {
		if ctx != nil && strings.HasPrefix(ctx.URL().Path, "/v1/") {
			if status == http.StatusBadRequest || status == http.StatusUnsupportedMediaType || status == http.StatusUnprocessableEntity {
				return safeInputError(ctx, status, errs...)
			}
			if status >= 500 {
				return Problem(status, "internal_error")
			}
		}
		return original(ctx, status, message, errs...)
	}
}

func registerProblemSchema(api huma.API) {
	registry := api.OpenAPI().Components.Schemas
	registry.Schema(reflect.TypeOf(huma.ErrorModel{}), true, "ErrorModel")
	registry.Schema(reflect.TypeOf(huma.ErrorDetail{}), true, "ErrorDetail")
	errorSchema := registry.Map()["ErrorModel"]
	if errorSchema != nil {
		if errorSchema.Properties == nil {
			errorSchema.Properties = map[string]*huma.Schema{}
		}
		errorSchema.Properties["code"] = &huma.Schema{
			Type: "string", Description: "Stable machine-readable error code for programmatic error handling.",
			Examples: []any{"validation_failed"},
		}
		if !slices.Contains(errorSchema.Required, "code") {
			errorSchema.Required = append(errorSchema.Required, "code")
		}
		if detail := registry.Map()["ErrorDetail"]; detail != nil {
			delete(detail.Properties, "value")
		}
	}
}

func problemResponse(status int, description, code string) *huma.Response {
	return &huma.Response{
		Description: description,
		Content: map[string]*huma.MediaType{"application/problem+json": {
			Schema:  &huma.Schema{Ref: "#/components/schemas/ErrorModel"},
			Example: Problem(status, code),
		}},
	}
}

func transportProblemResponse(api huma.API, status int, description, detail string) *huma.Response {
	return &huma.Response{
		Description: description,
		Content: map[string]*huma.MediaType{"application/problem+json": {
			Schema:  huma.SchemaFromType(api.OpenAPI().Components.Schemas, reflect.TypeFor[huma.ErrorModel]()),
			Example: huma.NewError(status, detail),
		}},
	}
}
