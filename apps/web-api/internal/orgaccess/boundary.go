package orgaccess

import (
	"context"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/identityhttp"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/persistence"
	"github.com/vandordev/vkit-zenith/packages/application/customers"
	"github.com/vandordev/vkit-zenith/packages/application/identity"
)

type Access struct {
	UserID, OrganizationID uuid.UUID
	Role                   customers.Role
	Suspended, HasLogo     bool
	Name, Slug             string
}

type WriteHeaders struct {
	Origin string `header:"Origin"`
	CSRF   string `header:"X-CSRF-Token"`
}

type Boundary struct {
	runtime *identityhttp.Runtime
	queries *persistence.Queries
}

func NewBoundary(runtime *identityhttp.Runtime, queries *persistence.Queries) *Boundary {
	return &Boundary{runtime: runtime, queries: queries}
}

func (b *Boundary) Subject(ctx context.Context) (uuid.UUID, error) {
	if b == nil || b.runtime == nil || b.runtime.Service == nil || b.queries == nil {
		return uuid.Nil, huma.Error503ServiceUnavailable("Service unavailable")
	}
	bindings := identityhttp.RequestBrowserBindings(ctx)
	if bindings.Session == ([32]byte{}) {
		return uuid.Nil, huma.Error401Unauthorized("Session tidak tersedia")
	}
	view, err := b.runtime.Service.Current(ctx, bindings.Session)
	if errors.Is(err, identity.ErrUnauthenticated) || errors.Is(err, identity.ErrDenied) {
		return uuid.Nil, huma.Error401Unauthorized("Session tidak tersedia")
	}
	if err != nil {
		return uuid.Nil, huma.Error503ServiceUnavailable("Service unavailable")
	}
	if view.Kind != identity.SessionFull {
		return uuid.Nil, huma.Error401Unauthorized("Session tidak tersedia")
	}
	userID, err := uuid.Parse(view.Subject)
	if err != nil {
		return uuid.Nil, huma.Error401Unauthorized("Session tidak tersedia")
	}
	return userID, nil
}

func (b *Boundary) Read(ctx context.Context, organizationID string, permission customers.Permission) (Access, error) {
	return b.access(ctx, organizationID, permission, nil)
}

func (b *Boundary) Write(ctx context.Context, organizationID string, permission customers.Permission, headers WriteHeaders) (Access, error) {
	return b.access(ctx, organizationID, permission, &headers)
}

func (b *Boundary) access(ctx context.Context, rawOrganizationID string, permission customers.Permission, headers *WriteHeaders) (Access, error) {
	var empty Access
	if b == nil || b.queries == nil {
		return empty, huma.Error503ServiceUnavailable("Service unavailable")
	}
	userID, err := b.Subject(ctx)
	if err != nil {
		return empty, err
	}
	organizationID, err := uuid.Parse(rawOrganizationID)
	if err != nil {
		return empty, huma.Error404NotFound("Not found")
	}
	row, err := b.queries.WebMembershipForOrganization(ctx, persistence.WebMembershipForOrganizationParams{
		UserID: pgtypeUUID(userID), OrganizationID: pgtypeUUID(organizationID),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return empty, huma.Error404NotFound("Not found")
	}
	if err != nil {
		return empty, huma.Error503ServiceUnavailable("Service unavailable")
	}
	role := customers.Role(row.Role)
	if !role.Can(permission) {
		return empty, huma.Error403Forbidden("Forbidden")
	}
	if headers != nil {
		if b.runtime == nil || headers.Origin == "" || headers.Origin != b.runtime.Origin || !b.runtime.Service.ValidateSessionCSRF(identityhttp.RequestBrowserBindings(ctx).Session, headers.CSRF) {
			return empty, huma.Error403Forbidden("Forbidden")
		}
	}
	access := Access{UserID: userID, OrganizationID: organizationID, Role: role, Suspended: row.Status == "suspended", HasLogo: row.HasLogo, Name: row.Name, Slug: row.Slug}
	if headers != nil && access.Suspended {
		return empty, huma.Error409Conflict("Organization is suspended")
	}
	return access, nil
}

func pgtypeUUID(id uuid.UUID) pgtype.UUID { return pgtype.UUID{Bytes: [16]byte(id), Valid: true} }
