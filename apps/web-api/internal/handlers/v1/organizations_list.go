package v1

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/orgaccess"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/persistence"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
)

type V1MyOrganization struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Status      string   `json:"status" enum:"active,suspended"`
	HasLogo     bool     `json:"hasLogo"`
	Role        string   `json:"role" enum:"owner,admin,member"`
	Permissions []string `json:"permissions" nullable:"false"`
}
type webOrganizationsOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         struct {
		Items []V1MyOrganization `json:"items" nullable:"false"`
	}
}

//vkit:handler
func NewListMyOrganizations(api huma.API, boundary *orgaccess.Boundary, q *persistence.Queries) error {
	huma.Register(api, huma.Operation{OperationID: "webV1ListMyOrganizations", Method: http.MethodGet, Path: "/v1/organizations", Errors: []int{401, 503}}, func(ctx context.Context, _ *struct{}) (*webOrganizationsOutput, error) {
		userID, err := boundary.Subject(ctx)
		if err != nil {
			return nil, err
		}
		rows, err := q.WebMembershipsForUser(ctx, pgtype.UUID{Bytes: [16]byte(userID), Valid: true})
		if err != nil {
			return nil, huma.Error503ServiceUnavailable("Service unavailable")
		}
		out := &webOrganizationsOutput{CacheControl: "no-store"}
		out.Body.Items = make([]V1MyOrganization, 0, len(rows))
		for _, row := range rows {
			role := customers.Role(row.Role)
			permissions := role.Permissions()
			item := V1MyOrganization{ID: webUUID(row.ID), Name: row.Name, Slug: row.Slug, Status: row.Status, HasLogo: row.HasLogo, Role: row.Role, Permissions: make([]string, 0, len(permissions))}
			for _, permission := range permissions {
				item.Permissions = append(item.Permissions, string(permission))
			}
			out.Body.Items = append(out.Body.Items, item)
		}
		return out, nil
	})
	return nil
}

func webUUID(value pgtype.UUID) string { return uuid.UUID(value.Bytes).String() }
