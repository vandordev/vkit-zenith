// Package customers owns customer administration, never customer authentication.
package customers

import (
	"github.com/jackc/pgx/v5/pgtype"
	"time"
)

type User struct {
	ID                         pgtype.UUID
	Email, DisplayName, Status string
	EmailVerifiedAt            *time.Time
	CreatedAt, UpdatedAt       time.Time
}
type Organization struct {
	ID                   pgtype.UUID
	Name, Slug, Status   string
	CreatedAt, UpdatedAt time.Time
}
type Membership struct {
	ID, OrganizationID, UserID pgtype.UUID
	Status                     string
	Role                       Role
	CreatedAt, UpdatedAt       time.Time
}
type UserCreate struct{ Email, DisplayName string }
type UserProfile struct{ DisplayName, Revision string }
type OrganizationCreate struct {
	Name, Slug  string
	OwnerUserID pgtype.UUID
	NewOwner    *UserCreate
}
type OrganizationProfile struct{ Name, Slug, Revision string }
type Sensitive struct{ Revision, Reason, CurrentPassword string }
type UserStatus struct {
	Sensitive
	Status string
}
type OrganizationStatus struct {
	Sensitive
	Status string
}
type UserEmail struct {
	Sensitive
	Email string
}
type RoleChange struct {
	Sensitive
	Role Role
}
type OrganizationCreated struct {
	Organization Organization
	Owner        Membership
	User         User
}
