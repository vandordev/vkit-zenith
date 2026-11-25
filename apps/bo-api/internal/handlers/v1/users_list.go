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

type V1UserListItem struct {
	ID                        string     `json:"id" format:"uuid"`
	Email                     string     `json:"email"`
	DisplayName               string     `json:"displayName"`
	Status                    string     `json:"status" enum:"active,disabled"`
	EmailVerifiedAt           *time.Time `json:"emailVerifiedAt"`
	CreatedAt                 time.Time  `json:"createdAt"`
	UpdatedAt                 time.Time  `json:"updatedAt"`
	Revision                  string     `json:"revision"`
	MembershipCount           string     `json:"membershipCount"`
	ActiveMembershipCount     string     `json:"activeMembershipCount"`
	OwnerMembershipCount      string     `json:"ownerMembershipCount"`
	CanChangeStatus           bool       `json:"canChangeStatus"`
	StatusBlockedReason       string     `json:"statusBlockedReason"`
	CanChangeEmail            bool       `json:"canChangeEmail"`
	EmailBlockedReason        string     `json:"emailBlockedReason"`
	CanVerifyEmail            bool       `json:"canVerifyEmail"`
	VerificationBlockedReason string     `json:"verificationBlockedReason"`
}
type V1UserListSummary struct {
	Total      string `json:"total"`
	Active     string `json:"active"`
	Disabled   string `json:"disabled"`
	Verified   string `json:"verified"`
	Unverified string `json:"unverified"`
}
type V1UserListPagination struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Total    string `json:"total"`
}
type V1UserListBody struct {
	Items      []V1UserListItem     `json:"items" nullable:"false"`
	Pagination V1UserListPagination `json:"pagination"`
	Summary    V1UserListSummary    `json:"summary"`
}
type V1UserListOutput struct{ Body V1UserListBody }

func customerUserListItem(r persistence.BOListCustomerUsersRow, super bool) V1UserListItem {
	u := customerUserDomain(r)
	statusReason := customerDisableReason(super, r.CanDisable, u.Status)
	verifyReason := customerVerifyReason(super, u.EmailVerifiedAt != nil)
	return V1UserListItem{ID: u.ID.String(), Email: u.Email, DisplayName: u.DisplayName, Status: u.Status, EmailVerifiedAt: u.EmailVerifiedAt, CreatedAt: u.CreatedAt.UTC(), UpdatedAt: u.UpdatedAt.UTC(), Revision: customers.UserRevision(u), MembershipCount: customerCount(r.MembershipCount), ActiveMembershipCount: customerCount(r.ActiveMembershipCount), OwnerMembershipCount: customerCount(r.OwnerMembershipCount), CanChangeStatus: statusReason == "", StatusBlockedReason: statusReason, CanChangeEmail: super, EmailBlockedReason: customerSensitiveReason(super), CanVerifyEmail: verifyReason == "", VerificationBlockedReason: verifyReason}
}

//vkit:handler
func NewListUsers(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ListUsers", Method: http.MethodGet, Path: "/v1/users", Errors: []int{401, 403, 422, 503}}, func(ctx context.Context, in *customerUserPage) (*V1UserListOutput, error) {
		q, tx, actor, err := customerSnapshot(ctx, b)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback(ctx)
		rows, err := q.BOListCustomerUsers(ctx, persistence.BOListCustomerUsersParams{Search: in.Q, SearchPattern: customerPattern(in.Q), State: in.Status, Verification: in.Verification, PageLimit: int32(in.PageSize), PageOffset: int32((in.Page - 1) * in.PageSize)})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		n, err := q.BOCountCustomerUsers(ctx, persistence.BOCountCustomerUsersParams{Search: in.Q, SearchPattern: customerPattern(in.Q), State: in.Status, Verification: in.Verification})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		s, err := q.BOCustomerUserSummary(ctx)
		if err != nil {
			return nil, customerError(ctx, err)
		}
		out := &V1UserListOutput{Body: V1UserListBody{Items: make([]V1UserListItem, 0, len(rows)), Pagination: V1UserListPagination{in.Page, in.PageSize, customerCount(n)}, Summary: V1UserListSummary{customerCount(s.Total), customerCount(s.Active), customerCount(s.Disabled), customerCount(s.Verified), customerCount(s.Unverified)}}}
		for _, r := range rows {
			out.Body.Items = append(out.Body.Items, customerUserListItem(r, actor.Role == "superadmin"))
		}
		if err = tx.Commit(ctx); err != nil {
			return nil, customerError(ctx, err)
		}
		return out, nil
	})
	return nil
}
