package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/packages/fixtures"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/orgaccess"
	"net/http"
)

//vkit:handler
func NewListAPIKeys(api huma.API, b *orgaccess.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "webV1ListApiKeys", Method: http.MethodGet, Path: "/v1/organizations/{organizationId}/api-keys", Description: "[DUMMY] List API key metadata.", Errors: []int{401, 403, 404, 422, 503}}, func(ctx context.Context, in *apiKeysListInput) (*settingsList[fixtures.APIKey], error) {
		if err := dummyRead(ctx, b, in.OrganizationID); err != nil {
			return nil, err
		}
		if !validStatus(in.Status, "active", "revoked", "expired") {
			return nil, dummyInvalid()
		}
		id, e := fixtureOrg(in.OrganizationID)
		if e != nil {
			return nil, dummyNotFound()
		}
		out := []fixtures.APIKey{}
		for _, v := range fixtures.Org(id).APIKeys {
			if (in.Status == "" || v.Status == in.Status) && contains(in.Q, v.Name, v.Prefix) {
				out = append(out, v)
			}
		}
		return &settingsList[fixtures.APIKey]{CacheControl: "no-store", Body: listOutput(out, in.Page, in.Size)}, nil
	})
	return nil
}
