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

type StaffSessionsInput struct {
	StaffID string `path:"staffId" format:"uuid"`
	CustomerPage
}
type V1StaffSession struct {
	ID        string     `json:"id" format:"uuid"`
	CreatedAt time.Time  `json:"createdAt"`
	ExpiresAt time.Time  `json:"expiresAt"`
	RevokedAt *time.Time `json:"revokedAt"`
}
type StaffSessionsBody struct {
	Items      []V1StaffSession             `json:"items" nullable:"false"`
	Pagination V1OrganizationListPagination `json:"pagination"`
}
type StaffSessionsOutput struct{ Body StaffSessionsBody }

// DUMMY (ADR-0001 phase 2): returns fixtures, persists nothing.
//
//vkit:handler
func NewListStaffSessions(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ListStaffSessions", Method: http.MethodGet, Path: "/v1/staff/{staffId}/sessions", Description: "[DUMMY] Sesi staf contoh Backoffice.", Errors: []int{401, 403, 422}}, func(ctx context.Context, in *StaffSessionsInput) (*StaffSessionsOutput, error) {
		if _, err := b.Read(ctx, "superadmin"); err != nil {
			return nil, security.Error(ctx, err)
		}
		id, err := uuid.Parse(in.StaffID)
		if err != nil {
			return nil, dummyNotFound()
		}
		fixturesList := filterStaffSessions(fixtures.StaffSessions(id), in.Q)
		items := make([]V1StaffSession, 0, len(fixturesList))
		for _, item := range fixturesList {
			created, _ := time.Parse(time.RFC3339, item.CreatedAt)
			expires, _ := time.Parse(time.RFC3339, item.ExpiresAt)
			var revoked *time.Time
			if item.RevokedAt != nil {
				parsed, _ := time.Parse(time.RFC3339, *item.RevokedAt)
				revoked = &parsed
			}
			items = append(items, V1StaffSession{ID: item.ID, CreatedAt: created, ExpiresAt: expires, RevokedAt: revoked})
		}
		page, pg := paginate(items, in.Page, in.PageSize)
		return &StaffSessionsOutput{Body: StaffSessionsBody{Items: page, Pagination: pg}}, nil
	})
	return nil
}

func filterStaffSessions(items []fixtures.StaffSession, q string) []fixtures.StaffSession {
	if q == "" {
		return items
	}
	q = strings.ToLower(q)
	filtered := make([]fixtures.StaffSession, 0, len(items))
	for _, item := range items {
		revokedAt := ""
		if item.RevokedAt != nil {
			revokedAt = *item.RevokedAt
		}
		if strings.Contains(strings.ToLower(item.ID+" "+item.CreatedAt+" "+item.ExpiresAt+" "+revokedAt), q) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
