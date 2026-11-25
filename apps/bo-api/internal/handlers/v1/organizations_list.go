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

type V1OrganizationListItem struct {
	ID                  string    `json:"id" format:"uuid"`
	Name                string    `json:"name"`
	Slug                string    `json:"slug"`
	Status              string    `json:"status" enum:"active,suspended"`
	CreatedAt           time.Time `json:"createdAt"`
	UpdatedAt           time.Time `json:"updatedAt"`
	Revision            string    `json:"revision"`
	MemberCount         string    `json:"memberCount"`
	ActiveMemberCount   string    `json:"activeMemberCount"`
	ActiveOwnerCount    string    `json:"activeOwnerCount"`
	CanChangeStatus     bool      `json:"canChangeStatus"`
	StatusBlockedReason string    `json:"statusBlockedReason"`
}
type V1OrganizationListSummary struct {
	Total     string `json:"total"`
	Active    string `json:"active"`
	Suspended string `json:"suspended"`
}
type V1OrganizationListPagination struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Total    string `json:"total"`
}
type V1OrganizationListBody struct {
	Items      []V1OrganizationListItem     `json:"items" nullable:"false"`
	Pagination V1OrganizationListPagination `json:"pagination"`
	Summary    V1OrganizationListSummary    `json:"summary"`
}
type V1OrganizationListOutput struct{ Body V1OrganizationListBody }

func customerOrganizationListItem(r persistence.BOListCustomerOrganizationsRow, super bool) V1OrganizationListItem {
	o := customerOrganizationDomain(r)
	return V1OrganizationListItem{ID: o.ID.String(), Name: o.Name, Slug: o.Slug, Status: o.Status, CreatedAt: o.CreatedAt.UTC(), UpdatedAt: o.UpdatedAt.UTC(), Revision: customers.OrganizationRevision(o), MemberCount: customerCount(r.MemberCount), ActiveMemberCount: customerCount(r.ActiveMemberCount), ActiveOwnerCount: customerCount(r.ActiveOwnerCount), CanChangeStatus: super, StatusBlockedReason: customerSensitiveReason(super)}
}

//vkit:handler
func NewListOrganizations(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ListOrganizations", Method: http.MethodGet, Path: "/v1/organizations", Errors: []int{401, 403, 422, 503}}, func(ctx context.Context, in *customerOrganizationPage) (*V1OrganizationListOutput, error) {
		q, tx, actor, err := customerSnapshot(ctx, b)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback(ctx)
		rows, err := q.BOListCustomerOrganizations(ctx, persistence.BOListCustomerOrganizationsParams{Search: in.Q, SearchPattern: customerPattern(in.Q), State: in.Status, PageLimit: int32(in.PageSize), PageOffset: int32((in.Page - 1) * in.PageSize)})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		n, err := q.BOCountCustomerOrganizations(ctx, persistence.BOCountCustomerOrganizationsParams{Search: in.Q, SearchPattern: customerPattern(in.Q), State: in.Status})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		s, err := q.BOCustomerOrganizationSummary(ctx)
		if err != nil {
			return nil, customerError(ctx, err)
		}
		out := &V1OrganizationListOutput{Body: V1OrganizationListBody{Items: make([]V1OrganizationListItem, 0, len(rows)), Pagination: V1OrganizationListPagination{in.Page, in.PageSize, customerCount(n)}, Summary: V1OrganizationListSummary{customerCount(s.Total), customerCount(s.Active), customerCount(s.Suspended)}}}
		for _, r := range rows {
			out.Body.Items = append(out.Body.Items, customerOrganizationListItem(r, actor.Role == "superadmin"))
		}
		if err = tx.Commit(ctx); err != nil {
			return nil, customerError(ctx, err)
		}
		return out, nil
	})
	return nil
}
