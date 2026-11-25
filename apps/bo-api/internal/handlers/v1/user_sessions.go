package v1

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/fixtures"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
)

type UserSessionsInput struct {
	UserID string `path:"userId" format:"uuid"`
	CustomerPage
}
type V1UserSession struct {
	ID              string     `json:"id" format:"uuid"`
	AuthenticatedAt time.Time  `json:"authenticatedAt"`
	LastSeenAt      time.Time  `json:"lastSeenAt"`
	ExpiresAt       time.Time  `json:"expiresAt"`
	RevokedAt       *time.Time `json:"revokedAt"`
	CreatedAt       time.Time  `json:"createdAt"`
}
type UserSessionsBody struct {
	Items      []V1UserSession              `json:"items" nullable:"false"`
	Pagination V1OrganizationListPagination `json:"pagination"`
}
type UserSessionsOutput struct{ Body UserSessionsBody }

// DUMMY (ADR-0001 phase 2): returns fixtures, persists nothing.
//
//vkit:handler
func NewListUserSessions(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ListUserSessions", Method: http.MethodGet, Path: "/v1/users/{userId}/sessions", Description: "[DUMMY] Sesi masuk contoh Pengguna.", Errors: []int{401, 403, 422}}, func(ctx context.Context, in *UserSessionsInput) (*UserSessionsOutput, error) {
		if _, err := b.Read(ctx, "any_staff"); err != nil {
			return nil, security.Error(ctx, err)
		}
		id, err := uuid.Parse(in.UserID)
		if err != nil {
			return nil, dummyNotFound()
		}
		fixtureItems := filterUserSessions(fixtures.UserSessions(id), in.Q)
		items := make([]V1UserSession, 0, len(fixtureItems))
		for _, item := range fixtureItems {
			authenticated, _ := time.Parse(time.RFC3339, item.AuthenticatedAt)
			lastSeen, _ := time.Parse(time.RFC3339, item.LastSeenAt)
			expires, _ := time.Parse(time.RFC3339, item.ExpiresAt)
			created, _ := time.Parse(time.RFC3339, item.CreatedAt)
			var revoked *time.Time
			if item.RevokedAt != nil {
				parsed, _ := time.Parse(time.RFC3339, *item.RevokedAt)
				revoked = &parsed
			}
			items = append(items, V1UserSession{ID: item.ID, AuthenticatedAt: authenticated, LastSeenAt: lastSeen, ExpiresAt: expires, RevokedAt: revoked, CreatedAt: created})
		}
		page, pg := paginate(items, in.Page, in.PageSize)
		return &UserSessionsOutput{Body: UserSessionsBody{Items: page, Pagination: pg}}, nil
	})
	return nil
}

func filterUserSessions(items []fixtures.UserSession, q string) []fixtures.UserSession {
	if q == "" {
		return items
	}
	q = strings.ToLower(q)
	filtered := make([]fixtures.UserSession, 0, len(items))
	for _, item := range items {
		revokedAt := ""
		if item.RevokedAt != nil {
			revokedAt = *item.RevokedAt
		}
		if strings.Contains(strings.ToLower(item.ID+" "+item.AuthenticatedAt+" "+item.LastSeenAt+" "+item.ExpiresAt+" "+revokedAt), q) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
