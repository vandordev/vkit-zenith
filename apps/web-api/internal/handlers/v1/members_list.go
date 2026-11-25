package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/orgaccess"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/persistence"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
)

type V1OrganizationMember struct {
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Role     string    `json:"role" enum:"owner,admin,member"`
	JoinedAt time.Time `json:"joinedAt" format:"date-time"`
}
type webOrganizationMembersOutput struct {
	CacheControl string `header:"Cache-Control"`
	Body         struct {
		Items []V1OrganizationMember `json:"items" nullable:"false"`
	}
}

//vkit:handler
func NewListOrganizationMembers(api huma.API, boundary *orgaccess.Boundary, q *persistence.Queries) error {
	huma.Register(api, huma.Operation{OperationID: "webV1ListOrganizationMembers", Method: http.MethodGet, Path: "/v1/organizations/{organizationId}/members", Errors: []int{401, 403, 404, 503}}, func(ctx context.Context, in *webOrganizationPath) (*webOrganizationMembersOutput, error) {
		access, err := boundary.Read(ctx, in.OrganizationID, customers.PermissionOrganizationRead)
		if err != nil {
			return nil, err
		}
		rows, err := q.WebOrganizationMembers(ctx, pgtype.UUID{Bytes: [16]byte(access.OrganizationID), Valid: true})
		if err != nil {
			return nil, huma.Error503ServiceUnavailable("Service unavailable")
		}
		out := &webOrganizationMembersOutput{CacheControl: "no-store"}
		out.Body.Items = make([]V1OrganizationMember, 0, len(rows))
		for _, row := range rows {
			out.Body.Items = append(out.Body.Items, V1OrganizationMember{Name: row.DisplayName, Email: row.Email, Role: row.Role, JoinedAt: row.JoinedAt.Time.UTC()})
		}
		return out, nil
	})
	return nil
}
