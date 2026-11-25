package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/persistence"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"net/http"
	"time"
)

type V1OrganizationMembershipListItem struct {
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
type V1OrganizationMembershipListSummary struct {
	Total   string `json:"total"`
	Active  string `json:"active"`
	Removed string `json:"removed"`
	Owners  string `json:"owners"`
}
type V1OrganizationMembershipListPagination struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Total    string `json:"total"`
}
type V1OrganizationMembershipListBody struct {
	Items      []V1OrganizationMembershipListItem     `json:"items" nullable:"false"`
	Pagination V1OrganizationMembershipListPagination `json:"pagination"`
	Summary    V1OrganizationMembershipListSummary    `json:"summary"`
}
type V1OrganizationMembershipListOutput struct {
	Body V1OrganizationMembershipListBody
}
type customerOrganizationMembersInput struct {
	CustomerOrganizationPath
	CustomerMembershipPage
}

func customerMembershipListItem(r persistence.BOListCustomerMembershipsRow, super bool) V1OrganizationMembershipListItem {
	m := customerMembershipDomain(r)
	remove, restore, owner := customerMemberReasons(r, super)
	return V1OrganizationMembershipListItem{ID: m.ID.String(), OrganizationID: m.OrganizationID.String(), UserID: m.UserID.String(), Status: m.Status, Role: string(m.Role), CreatedAt: m.CreatedAt.UTC(), UpdatedAt: m.UpdatedAt.UTC(), Revision: customers.MembershipRevision(m), UserName: r.UserName, UserEmail: r.UserEmail, UserStatus: r.UserStatus, OrganizationName: r.OrganizationName, OrganizationSlug: r.OrganizationSlug, OrganizationStatus: r.OrganizationStatus, CanRemove: remove == "", RemoveBlockedReason: remove, CanRestore: restore == "", RestoreBlockedReason: restore, CanChangeRole: owner == "", RoleBlockedReason: owner}
}

//vkit:handler
func NewListOrganizationMemberships(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ListOrganizationMemberships", Method: http.MethodGet, Path: "/v1/organizations/{organizationId}/memberships", Errors: []int{401, 403, 404, 422, 503}}, func(ctx context.Context, in *customerOrganizationMembersInput) (*V1OrganizationMembershipListOutput, error) {
		q, tx, actor, err := customerSnapshot(ctx, b)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback(ctx)
		id := customerUUID(in.OrganizationID)
		if _, err = q.BOGetCustomerOrganization(ctx, id); err != nil {
			return nil, customerError(ctx, err)
		}
		rows, err := q.BOListCustomerMemberships(ctx, persistence.BOListCustomerMembershipsParams{OrganizationID: id, Search: in.Q, SearchPattern: customerPattern(in.Q), State: in.Status, Ownership: in.Ownership, PageLimit: int32(in.PageSize), PageOffset: int32((in.Page - 1) * in.PageSize)})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		n, err := q.BOCountCustomerMemberships(ctx, persistence.BOCountCustomerMembershipsParams{OrganizationID: id, Search: in.Q, SearchPattern: customerPattern(in.Q), State: in.Status, Ownership: in.Ownership})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		s, err := q.BOCustomerMembershipSummary(ctx, persistence.BOCustomerMembershipSummaryParams{OrganizationID: id})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		out := &V1OrganizationMembershipListOutput{Body: V1OrganizationMembershipListBody{Items: make([]V1OrganizationMembershipListItem, 0, len(rows)), Pagination: V1OrganizationMembershipListPagination{in.Page, in.PageSize, customerCount(n)}, Summary: V1OrganizationMembershipListSummary{customerCount(s.Total), customerCount(s.Active), customerCount(s.Removed), customerCount(s.Owners)}}}
		for _, r := range rows {
			out.Body.Items = append(out.Body.Items, customerMembershipListItem(r, actor.Role == "superadmin"))
		}
		if err = tx.Commit(ctx); err != nil {
			return nil, customerError(ctx, err)
		}
		return out, nil
	})
	return nil
}
