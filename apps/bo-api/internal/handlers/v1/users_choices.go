package v1

import (
	"context"
	"errors"
	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/persistence"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"net/http"
)

type V1UserChoiceListItem struct {
	ID             string `json:"id" format:"uuid"`
	Email          string `json:"email"`
	DisplayName    string `json:"displayName"`
	Status         string `json:"status" enum:"active,disabled"`
	Eligible       bool   `json:"eligible"`
	DisabledReason string `json:"disabledReason"`
}
type V1UserChoiceListPagination struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Total    string `json:"total"`
}
type V1UserChoiceListBody struct {
	Items      []V1UserChoiceListItem     `json:"items" nullable:"false"`
	Selected   *V1UserChoiceListItem      `json:"selected"`
	Pagination V1UserChoiceListPagination `json:"pagination"`
}
type V1UserChoiceListOutput struct{ Body V1UserChoiceListBody }

// Huma represents pointer-to-object fields by reference; express nullability
// explicitly, matching the installed product-choice contract convention.
func (V1UserChoiceListBody) TransformSchema(_ huma.Registry, s *huma.Schema) *huma.Schema {
	s.Properties["selected"] = &huma.Schema{AnyOf: []*huma.Schema{s.Properties["selected"], {Type: "null"}}}
	return s
}

type customerUserChoiceInput struct {
	CustomerPage
	SelectedID     string `query:"selectedId" format:"uuid"`
	OrganizationID string `query:"organizationId" format:"uuid"`
}

func customerUserChoice(r persistence.BOListCustomerUserChoicesRow) V1UserChoiceListItem {
	reason := ""
	if r.Status != "active" {
		reason = "Pengguna dinonaktifkan."
	} else if r.MembershipStatus.Valid {
		if r.MembershipStatus.String == "removed" {
			reason = "Keanggotaan dihapus; gunakan pemulihan eksplisit."
		} else {
			reason = "Pengguna sudah menjadi anggota."
		}
	}
	return V1UserChoiceListItem{ID: r.ID.String(), Email: r.Email, DisplayName: r.DisplayName, Status: r.Status, Eligible: reason == "", DisabledReason: reason}
}

//vkit:handler
func NewListUserChoices(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ListUserChoices", Method: http.MethodGet, Path: "/v1/users/choices", Errors: []int{401, 403, 404, 422, 503}}, func(ctx context.Context, in *customerUserChoiceInput) (*V1UserChoiceListOutput, error) {
		q, tx, _, err := customerSnapshot(ctx, b)
		if err != nil {
			return nil, err
		}
		defer tx.Rollback(ctx)
		var org pgtype.UUID
		if in.OrganizationID != "" {
			org = customerUUID(in.OrganizationID)
			if _, err = q.BOGetCustomerOrganization(ctx, org); err != nil {
				return nil, customerError(ctx, err)
			}
		}
		rows, err := q.BOListCustomerUserChoices(ctx, persistence.BOListCustomerUserChoicesParams{OrganizationID: org, Search: in.Q, SearchPattern: customerPattern(in.Q), PageLimit: int32(in.PageSize), PageOffset: int32((in.Page - 1) * in.PageSize)})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		n, err := q.BOCountCustomerUsers(ctx, persistence.BOCountCustomerUsersParams{Search: in.Q, SearchPattern: customerPattern(in.Q)})
		if err != nil {
			return nil, customerError(ctx, err)
		}
		out := &V1UserChoiceListOutput{Body: V1UserChoiceListBody{Items: make([]V1UserChoiceListItem, 0, len(rows)), Pagination: V1UserChoiceListPagination{in.Page, in.PageSize, customerCount(n)}}}
		for _, r := range rows {
			out.Body.Items = append(out.Body.Items, customerUserChoice(r))
		}
		if in.SelectedID != "" {
			r, e := q.BOGetCustomerUserChoice(ctx, persistence.BOGetCustomerUserChoiceParams{ID: customerUUID(in.SelectedID), OrganizationID: org})
			if e != nil && !errors.Is(e, pgx.ErrNoRows) {
				return nil, customerError(ctx, e)
			}
			if e == nil {
				selected := customerUserChoice(persistence.BOListCustomerUserChoicesRow(r))
				out.Body.Selected = &selected
			}
		}
		if err = tx.Commit(ctx); err != nil {
			return nil, customerError(ctx, err)
		}
		return out, nil
	})
	return nil
}
