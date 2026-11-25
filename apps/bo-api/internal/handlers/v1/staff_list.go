package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/persistence"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
	"net/http"
	"strconv"
	"time"
)

type V1StaffListItem struct {
	ID          string    `json:"id" format:"uuid"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Role        string    `json:"role" enum:"superadmin,admin"`
	Status      string    `json:"status" enum:"active,disabled"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Revision    string    `json:"revision"`
}

func v1StaffListItem(s staff.Staff) V1StaffListItem {
	return V1StaffListItem{s.ID.String(), s.Email, s.DisplayName, s.Role, s.Status, s.UpdatedAt.UTC(), staff.Revision(s)}
}

type V1StaffListSummary struct {
	Total      string `json:"total"`
	Active     string `json:"active"`
	Disabled   string `json:"disabled"`
	Superadmin string `json:"superadmin"`
	Admin      string `json:"admin"`
}
type V1StaffListPagination struct {
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
	Total    int64 `json:"total"`
}
type V1StaffListBody struct {
	Items      []V1StaffListItem     `json:"items" nullable:"false"`
	Summary    V1StaffListSummary    `json:"summary"`
	Pagination V1StaffListPagination `json:"pagination"`
}
type V1StaffListOutput struct{ Body V1StaffListBody }

//vkit:handler
func NewListStaff(api huma.API, b *security.Boundary) error {
	huma.Register(api, huma.Operation{OperationID: "boV1ListStaff", Method: http.MethodGet, Path: "/v1/staff", Errors: []int{401, 403, 422, 503}}, func(ctx context.Context, in *PageQuery) (*V1StaffListOutput, error) {
		if _, err := b.Read(ctx, "superadmin"); err != nil {
			return nil, security.Error(ctx, err)
		}
		tx, err := b.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
		if err != nil {
			return nil, catalogReadError(ctx, err)
		}
		defer tx.Rollback(ctx)
		q := persistence.New(tx)
		rows, err := q.BOStaffList(ctx, persistence.BOStaffListParams{Column1: in.Q, Column2: in.Status, Limit: int32(in.PageSize), Offset: int32((in.Page - 1) * in.PageSize)})
		if err != nil {
			return nil, security.Error(ctx, err)
		}
		total, err := q.BOStaffCount(ctx, persistence.BOStaffCountParams{Column1: in.Q, Column2: in.Status})
		if err != nil {
			return nil, security.Error(ctx, err)
		}
		s, err := q.BOStaffSummary(ctx)
		if err != nil {
			return nil, catalogReadError(ctx, err)
		}
		out := &V1StaffListOutput{}
		out.Body.Items = []V1StaffListItem{}
		out.Body.Pagination = V1StaffListPagination{Page: in.Page, PageSize: in.PageSize, Total: total}
		out.Body.Summary = V1StaffListSummary{strconv.FormatInt(s.Total, 10), strconv.FormatInt(s.Active, 10), strconv.FormatInt(s.Disabled, 10), strconv.FormatInt(s.Superadmin, 10), strconv.FormatInt(s.Admin, 10)}
		for _, r := range rows {
			out.Body.Items = append(out.Body.Items, v1StaffListItem(staff.Staff{ID: r.ID, Email: r.EmailNormalized, DisplayName: r.DisplayName, Role: r.Role, Status: r.Status, UpdatedAt: r.UpdatedAt.Time}))
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, catalogReadError(ctx, err)
		}
		return out, nil
	})
	return nil
}
