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

type UserEmailActionsInput struct {
	UserID string `path:"userId" format:"uuid"`
	CustomerPage
}
type V1UserEmailAction struct {
	ID         string     `json:"id" format:"uuid"`
	Purpose    string     `json:"purpose" enum:"verify_email,setup_password,reset_password"`
	CreatedAt  time.Time  `json:"createdAt"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	ConsumedAt *time.Time `json:"consumedAt"`
	RevokedAt  *time.Time `json:"revokedAt"`
}
type UserEmailActionsBody struct {
	Items      []V1UserEmailAction          `json:"items" nullable:"false"`
	Pagination V1OrganizationListPagination `json:"pagination"`
}
type UserEmailActionsOutput struct{ Body UserEmailActionsBody }

// DUMMY (ADR-0001 phase 2): returns fixtures, persists nothing.
//
//vkit:handler
func NewListUserEmailActions(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ListUserEmailActions", Method: http.MethodGet, Path: "/v1/users/{userId}/email-actions", Description: "[DUMMY] Metadata tindakan email contoh Pengguna.", Errors: []int{401, 403, 422}}, func(ctx context.Context, in *UserEmailActionsInput) (*UserEmailActionsOutput, error) {
		if _, err := b.Read(ctx, "any_staff"); err != nil {
			return nil, security.Error(ctx, err)
		}
		id, err := uuid.Parse(in.UserID)
		if err != nil {
			return nil, dummyNotFound()
		}
		fixtureItems := filterUserEmailActions(fixtures.UserEmailActions(id), in.Q)
		items := make([]V1UserEmailAction, 0, len(fixtureItems))
		for _, item := range fixtureItems {
			created, _ := time.Parse(time.RFC3339, item.CreatedAt)
			expires, _ := time.Parse(time.RFC3339, item.ExpiresAt)
			var consumed, revoked *time.Time
			if item.ConsumedAt != nil {
				parsed, _ := time.Parse(time.RFC3339, *item.ConsumedAt)
				consumed = &parsed
			}
			if item.RevokedAt != nil {
				parsed, _ := time.Parse(time.RFC3339, *item.RevokedAt)
				revoked = &parsed
			}
			items = append(items, V1UserEmailAction{ID: item.ID, Purpose: item.Purpose, CreatedAt: created, ExpiresAt: expires, ConsumedAt: consumed, RevokedAt: revoked})
		}
		page, pg := paginate(items, in.Page, in.PageSize)
		return &UserEmailActionsOutput{Body: UserEmailActionsBody{Items: page, Pagination: pg}}, nil
	})
	return nil
}

func filterUserEmailActions(items []fixtures.EmailAction, q string) []fixtures.EmailAction {
	if q == "" {
		return items
	}
	q = strings.ToLower(q)
	filtered := make([]fixtures.EmailAction, 0, len(items))
	for _, item := range items {
		consumedAt, revokedAt, status := "", "", "pending"
		if item.ConsumedAt != nil {
			consumedAt, status = *item.ConsumedAt, "consumed"
		}
		if item.RevokedAt != nil {
			revokedAt, status = *item.RevokedAt, "revoked"
		}
		if strings.Contains(strings.ToLower(item.ID+" "+item.Purpose+" "+item.CreatedAt+" "+item.ExpiresAt+" "+consumedAt+" "+revokedAt+" "+status), q) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
