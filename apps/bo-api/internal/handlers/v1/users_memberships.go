package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/persistence"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"net/http"
	"time"
)

type V1UserMembershipListItem struct {
	ID                   string    `json:"id" format:"uuid"`
	OrganizationID       string    `json:"organizationId" format:"uuid"`
	UserID               string    `json:"userId" format:"uuid"`
	Status               string    `json:"status" enum:"active,removed"`
	Role                 string    `json:"role" enum:"owner,admin,member"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
	Revision             string    `json:"revision"`
	UserName             string    `json:"userName"`
	UserEmail            string    `json:"userEmail"`
	UserStatus           string    `json:"userStatus" enum:"active,disabled"`
	OrganizationName     string    `json:"organizationName"`
	OrganizationSlug     string    `json:"organizationSlug"`
	OrganizationStatus   string    `json:"organizationStatus" enum:"active,suspended"`
	CanRemove            bool      `json:"canRemove"`
	RemoveBlockedReason  string    `json:"removeBlockedReason"`
	CanRestore           bool      `json:"canRestore"`
	RestoreBlockedReason string    `json:"restoreBlockedReason"`
	CanChangeRole        bool      `json:"canChangeRole"`
	RoleBlockedReason    string    `json:"roleBlockedReason"`
}
type V1UserMembershipListSummary struct {
	Total   string `json:"total"`
	Active  string `json:"active"`
	Removed string `json:"removed"`
	Owners  string `json:"owners"`
}
type V1UserMembershipListPagination struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Total    string `json:"total"`
}
type V1UserMembershipListBody struct {
	Items      []V1UserMembershipListItem     `json:"items" nullable:"false"`
	Pagination V1UserMembershipListPagination `json:"pagination"`
	Summary    V1UserMembershipListSummary    `json:"summary"`
}
type V1UserMembershipListOutput struct{ Body V1UserMembershipListBody }
type customerUserMembersInput struct {
	ItemPath
	CustomerMembershipPage
}

//vkit:handler
func NewListUserMemberships(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ListUserMemberships", Method: http.MethodGet, Path: "/v1/users/{id}/memberships", Errors: []int{401, 403, 404, 422, 503}}, func(ctx context.Context, in *customerUserMembersInput) (*V1UserMembershipListOutput, error) {
		q, tx, actor, err := customerSnapshot(ctx, b)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback(ctx)
		id := customerUUID(in.ID)
		if _, err = q.BOGetCustomerUser(ctx, id); err != nil {
			return nil, customerError(ctx, err)
		}
		rows, err := q.BOListCustomerMemberships(ctx, persistence.BOListCustomerMembershipsParams{UserID: id, Search: in.Q, SearchPattern: customerPattern(in.Q), State: in.Status, Ownership: in.Ownership, PageLimit: int32(in.PageSize), PageOffset: int32((in.Page - 1) * in.PageSize)})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		n, err := q.BOCountCustomerMemberships(ctx, persistence.BOCountCustomerMembershipsParams{UserID: id, Search: in.Q, SearchPattern: customerPattern(in.Q), State: in.Status, Ownership: in.Ownership})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		s, err := q.BOCustomerMembershipSummary(ctx, persistence.BOCustomerMembershipSummaryParams{UserID: id})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		out := &V1UserMembershipListOutput{Body: V1UserMembershipListBody{Items: make([]V1UserMembershipListItem, 0, len(rows)), Pagination: V1UserMembershipListPagination{in.Page, in.PageSize, customerCount(n)}, Summary: V1UserMembershipListSummary{customerCount(s.Total), customerCount(s.Active), customerCount(s.Removed), customerCount(s.Owners)}}}
		for _, r := range rows {
			out.Body.Items = append(out.Body.Items, V1UserMembershipListItem(customerMembershipListItem(r, actor.Role == "superadmin")))
		}
		if err = tx.Commit(ctx); err != nil {
			return nil, customerError(ctx, err)
		}
		return out, nil
	})
	return nil
}
