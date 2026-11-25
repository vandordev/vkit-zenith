package v1

import (
	"context"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/fixtures"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
)

type StaffHistoryInput struct {
	StaffID string `path:"staffId" format:"uuid"`
	CustomerPage
}
type StaffHistoryBody struct {
	Items      []V1HistoryEntry             `json:"items" nullable:"false"`
	Pagination V1OrganizationListPagination `json:"pagination"`
}
type StaffHistoryOutput struct{ Body StaffHistoryBody }

// DUMMY (ADR-0001 phase 2): returns fixtures, persists nothing.
//
//vkit:handler
func NewListStaffHistory(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ListStaffHistory", Method: http.MethodGet, Path: "/v1/staff/{staffId}/history", Description: "[DUMMY] Riwayat staf contoh Backoffice.", Errors: []int{401, 403, 422}}, func(ctx context.Context, in *StaffHistoryInput) (*StaffHistoryOutput, error) {
		if _, err := b.Read(ctx, "superadmin"); err != nil {
			return nil, security.Error(ctx, err)
		}
		id, err := uuid.Parse(in.StaffID)
		if err != nil {
			return nil, dummyNotFound()
		}
		items := userHistoryEntries(fixtures.StaffHistory(id))
		if in.Q != "" {
			q := strings.ToLower(in.Q)
			filtered := items[:0]
			for _, item := range items {
				if strings.Contains(strings.ToLower(item.Action+" "+stringValue(item.Actor.Name)+" "+stringValue(item.Actor.System)), q) {
					filtered = append(filtered, item)
				}
			}
			items = filtered
		}
		page, pg := paginate(items, in.Page, in.PageSize)
		return &StaffHistoryOutput{Body: StaffHistoryBody{Items: page, Pagination: pg}}, nil
	})
	return nil
}
