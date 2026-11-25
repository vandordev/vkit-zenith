package v1

import (
	"context"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/orgaccess"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
)

func settingsWrite(ctx context.Context, b *orgaccess.Boundary, org string, permission customers.Permission, origin, csrf string) error {
	_, err := b.Write(ctx, org, permission, orgaccess.WriteHeaders{Origin: origin, CSRF: csrf})
	return err
}
func fixtureOrg(raw string) (uuid.UUID, error) { return uuid.Parse(raw) }
func dummyRead(ctx context.Context, b *orgaccess.Boundary, org string) error {
	_, err := b.Read(ctx, org, customers.PermissionOrganizationRead)
	return err
}
func dummyNotFound() error { return huma.Error404NotFound("Not found") }
func dummyInvalid() error  { return huma.Error422UnprocessableEntity("Isian tidak valid.") }
func contains(q string, values ...string) bool {
	q = strings.ToLower(q)
	if q == "" {
		return true
	}
	for _, v := range values {
		if strings.Contains(strings.ToLower(v), q) {
			return true
		}
	}
	return false
}
func validStatus(status string, allowed ...string) bool {
	if status == "" {
		return true
	}
	for _, candidate := range allowed {
		if status == candidate {
			return true
		}
	}
	return false
}
func listOutput[T any](items []T, page, size int) V1List[T] {
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 20
	}
	selected, pagination := paginate(items, page, size)
	if selected == nil {
		selected = []T{}
	}
	return V1List[T]{Items: selected, Pagination: pagination}
}
