package v1

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/fixtures"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"strconv"
)

type dummyListInput struct {
	CustomerOrganizationPath
	CustomerPage
	Status string `query:"status"`
}
type dummyGlobalListInput struct {
	CustomerPage
	OrganizationID string `query:"organizationId" format:"uuid"`
}

func paginate[T any](items []T, page, size int) ([]T, V1OrganizationListPagination) {
	n := len(items)
	a := (page - 1) * size
	if a > n {
		a = n
	}
	z := a + size
	if z > n {
		z = n
	}
	return items[a:z], V1OrganizationListPagination{page, size, strconv.Itoa(n)}
}
func dummyOrg(ctx context.Context, b *security.Boundary, raw string) (fixtures.OrgData, error) {
	if _, e := b.Read(ctx, "any_staff"); e != nil {
		return fixtures.OrgData{}, security.Error(ctx, e)
	}
	id, e := uuid.Parse(raw)
	if e != nil {
		return fixtures.OrgData{}, &huma.ErrorModel{Status: 404, Title: "Not Found", Detail: "Data tidak tersedia."}
	}
	d, _ := fixtures.Org(id)
	if id == fixtures.ErrorOrgID {
		return d, huma.Error503ServiceUnavailable("Data dummy gagal dimuat")
	}
	return d, nil
}
func dummyValidation() error {
	return &huma.ErrorModel{Status: 422, Title: "Unprocessable Entity", Detail: "Isian tidak valid."}
}
func dummyNotFound() error {
	return &huma.ErrorModel{Status: 404, Title: "Not Found", Detail: "Data tidak tersedia."}
}
