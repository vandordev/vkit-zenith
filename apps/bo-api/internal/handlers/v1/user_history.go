package v1

import (
	"context"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/fixtures"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
)

type UserHistoryInput struct {
	UserID string `path:"userId" format:"uuid"`
	CustomerPage
}
type UserHistoryBody struct {
	Items      []V1HistoryEntry             `json:"items" nullable:"false"`
	Pagination V1OrganizationListPagination `json:"pagination"`
}
type UserHistoryOutput struct{ Body UserHistoryBody }

// DUMMY (ADR-0001 phase 2): returns fixtures, persists nothing.
//
//vkit:handler
func NewListUserHistory(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ListUserHistory", Method: http.MethodGet, Path: "/v1/users/{userId}/history", Description: "[DUMMY] Riwayat contoh Pengguna.", Errors: []int{401, 403, 422}}, func(ctx context.Context, in *UserHistoryInput) (*UserHistoryOutput, error) {
		if _, err := b.Read(ctx, "any_staff"); err != nil {
			return nil, security.Error(ctx, err)
		}
		id, err := uuid.Parse(in.UserID)
		if err != nil {
			return nil, dummyNotFound()
		}
		items := userHistoryEntries(fixtures.UserHistory(id))
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
		return &UserHistoryOutput{Body: UserHistoryBody{Items: page, Pagination: pg}}, nil
	})
	return nil
}

func userHistoryEntries(history []fixtures.History) []V1HistoryEntry {
	items := make([]V1HistoryEntry, 0, len(history))
	for _, h := range history {
		at, _ := time.Parse(time.RFC3339, h.OccurredAt)
		actor := V1Actor{Kind: h.ActorKind, Name: strPtr(h.ActorName), ID: h.ActorID, Role: h.Role, APIKeyPrefix: h.APIKeyPrefix, System: h.System}
		items = append(items, V1HistoryEntry{ID: h.ID, OccurredAt: at, Actor: actor, Action: h.Action, Outcome: h.Outcome, TargetType: h.TargetType, TargetID: h.TargetID, Reason: h.Reason})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].OccurredAt.After(items[j].OccurredAt) })
	return items
}
