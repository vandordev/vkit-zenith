package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/orgaccess"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"net/http"
)

type logoUploadInput struct {
	OrganizationID string `path:"organizationId"`
	Origin         string `header:"Origin"`
	CSRF           string `header:"X-CSRF-Token"`
	Body           struct {
		ContentType string `json:"contentType" enum:"image/png,image/jpeg,image/svg+xml"`
		Size        int    `json:"size" minimum:"1" maximum:"1048576"`
	}
}
type logoUploadOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         V1DummyResult[logoResult]
}

// DUMMY (ADR-0001 phase 2): returns metadata, accepts no file and persists nothing.
//
//vkit:handler
func NewUploadLogo(api huma.API, b *orgaccess.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "webV1UploadLogo", Method: http.MethodPost, Path: "/v1/organizations/{organizationId}/profile/logo", Description: "[DUMMY] Simulate uploading organization logo metadata.", Errors: []int{401, 403, 404, 409, 422, 503}}, func(ctx context.Context, in *logoUploadInput) (*logoUploadOutput, error) {
		if err := settingsWrite(ctx, b, in.OrganizationID, customers.PermissionOrganizationManage, in.Origin, in.CSRF); err != nil {
			return nil, err
		}
		if (in.Body.ContentType != "image/png" && in.Body.ContentType != "image/jpeg" && in.Body.ContentType != "image/svg+xml") || in.Body.Size < 1 || in.Body.Size > 1048576 {
			return nil, dummyInvalid()
		}
		return &logoUploadOutput{CacheControl: "no-store", Body: V1DummyResult[logoResult]{Result: logoResult{ContentType: in.Body.ContentType, Size: in.Body.Size}, Dummy: true}}, nil
	})
	return nil
}
