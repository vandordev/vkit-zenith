package v1

import (
	"time"

	"github.com/danielgtaylor/huma/v2"
)

type V1ActivityActor struct {
	Kind string  `json:"kind" enum:"staff,anonymous,system"`
	ID   *string `json:"id" format:"uuid"`
	Name *string `json:"name"`
}
type V1ActivityTarget struct {
	Type           *string `json:"type"`
	ID             *string `json:"id" format:"uuid"`
	OrganizationID *string `json:"organizationId" format:"uuid"`
	Name           *string `json:"name"`
	Available      bool    `json:"available"`
}

type V1ActivityRow struct {
	ID         string           `json:"id" format:"uuid"`
	OccurredAt time.Time        `json:"occurredAt"`
	Action     string           `json:"action"`
	Outcome    string           `json:"outcome" enum:"success,failure,denied"`
	Actor      V1ActivityActor  `json:"actor"`
	Target     V1ActivityTarget `json:"target"`
}
type V1ActivityStaffState struct {
	Role   *string `json:"role" enum:"superadmin,admin"`
	Status *string `json:"status" enum:"active,disabled"`
}

func (V1ActivityStaffState) TransformSchema(_ huma.Registry, s *huma.Schema) *huma.Schema {
	return activityNullable(s, "role", "status")
}

type V1ActivityDetail struct {
	V1ActivityRow
	DetailAvailable bool                  `json:"detailAvailable"`
	ChangedFields   []string              `json:"changedFields" nullable:"false"`
	Previous        *V1ActivityStaffState `json:"previous"`
	Current         *V1ActivityStaffState `json:"current"`
}

// Current joined names are labels, never recorded historical values. Explicit
// referenced-object nullability is needed by Huma and the generated consumer.
func activityNullable(s *huma.Schema, names ...string) *huma.Schema {
	for _, name := range names {
		s.Properties[name] = &huma.Schema{AnyOf: []*huma.Schema{s.Properties[name], {Type: "null"}}}
	}
	return s
}

func (V1ActivityDetail) TransformSchema(_ huma.Registry, s *huma.Schema) *huma.Schema {
	return activityNullable(s, "previous", "current")
}

type V1ActivityPagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
	Total    int64 `json:"total"`
}
type V1ActivityListBody struct {
	Items      []V1ActivityRow      `json:"items" nullable:"false"`
	Pagination V1ActivityPagination `json:"pagination"`
}
type V1ActivityListOutput struct{ Body V1ActivityListBody }
type V1ActivityDetailOutput struct{ Body V1ActivityDetail }
type V1ActivityRecentBody struct {
	Items []V1ActivityRow `json:"items" nullable:"false"`
}
type V1ActivityRecentOutput struct{ Body V1ActivityRecentBody }

type V1ActivityStaffChoice struct {
	ID     string `json:"id" format:"uuid"`
	Name   string `json:"name"`
	Status string `json:"status" enum:"active,disabled"`
}
type V1ActivityStaffChoicesBody struct {
	Items      []V1ActivityStaffChoice `json:"items" nullable:"false"`
	Selected   *V1ActivityStaffChoice  `json:"selected"`
	Pagination V1ActivityPagination    `json:"pagination"`
}

func (V1ActivityStaffChoicesBody) TransformSchema(_ huma.Registry, s *huma.Schema) *huma.Schema {
	return activityNullable(s, "selected")
}

type V1ActivityStaffChoicesOutput struct{ Body V1ActivityStaffChoicesBody }
type V1ActivityAction struct {
	Value string `json:"value"`
	Label string `json:"label"`
}
type V1ActivityActionsBody struct {
	Items []V1ActivityAction `json:"items" nullable:"false"`
}
type V1ActivityActionsOutput struct{ Body V1ActivityActionsBody }
