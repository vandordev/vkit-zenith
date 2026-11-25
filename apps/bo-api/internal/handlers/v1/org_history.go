package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"net/http"
	"sort"
	"strings"
	"time"
)

type dummyHistoryInput struct {
	CustomerOrganizationPath
	CustomerPage
	TargetType string `query:"targetType"`
	TargetID   string `query:"targetId"`
}
type V1HistoryListBody struct {
	Items      []V1HistoryEntry             `json:"items"`
	Pagination V1OrganizationListPagination `json:"pagination"`
}
type V1HistoryListOutput struct{ Body V1HistoryListBody }

// DUMMY (ADR-0001 phase 2): returns fixtures, persists nothing.
//
//vkit:handler
func NewListHistory(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ListHistory", Method: http.MethodGet, Path: "/v1/organizations/{organizationId}/history", Description: "[DUMMY] Riwayat contoh Backoffice.", Errors: []int{401, 403, 404, 503}}, func(ctx context.Context, in *dummyHistoryInput) (*V1HistoryListOutput, error) {
		d, e := dummyOrg(ctx, b, in.OrganizationID)
		if e != nil {
			return nil, e
		}
		items := make([]V1HistoryEntry, 0, len(d.History))
		for _, h := range d.History {
			at, _ := time.Parse(time.RFC3339, h.OccurredAt)
			name := h.ActorName
			actor := V1Actor{Kind: h.ActorKind, Name: &name, System: h.System}
			if h.ActorID != nil {
				actor.ID = h.ActorID
			}
			if h.Role != nil {
				actor.Role = h.Role
			}
			if h.APIKeyPrefix != nil {
				actor.APIKeyPrefix = h.APIKeyPrefix
			}
			items = append(items, V1HistoryEntry{ID: h.ID, OccurredAt: at, Actor: actor, Action: h.Action, Outcome: h.Outcome, TargetType: h.TargetType, TargetID: h.TargetID, Reason: h.Reason})
		}

		sort.Slice(items, func(i, j int) bool { return items[i].OccurredAt.After(items[j].OccurredAt) })
		filtered := items[:0]
		for _, v := range items {
			if (in.TargetType == "" || (v.TargetType != nil && *v.TargetType == in.TargetType)) && (in.TargetID == "" || (v.TargetID != nil && *v.TargetID == in.TargetID)) && (in.Q == "" || strings.Contains(strings.ToLower(v.Action+" "+stringValue(v.Actor.Name)+" "+stringValue(v.Actor.System)+" "+stringValue(v.TargetID)), strings.ToLower(in.Q))) {
				filtered = append(filtered, v)
			}
		}
		items = filtered
		page, pg := paginate(items, in.Page, in.PageSize)
		return &V1HistoryListOutput{Body: V1HistoryListBody{page, pg}}, nil
	})
	return nil
}
func strPtr(s string) *string { return &s }
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
