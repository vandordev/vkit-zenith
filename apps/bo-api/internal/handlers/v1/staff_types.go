package v1

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

type StaffWriteBody struct {
	Email           string `json:"email" maxLength:"254"`
	DisplayName     string `json:"displayName" minLength:"1" maxLength:"200"`
	Role            string `json:"role" enum:"superadmin,admin"`
	Status          string `json:"status" enum:"active,disabled"`
	CurrentPassword string `json:"currentPassword" maxLength:"512"`
}

func (b StaffWriteBody) input() staff.StaffWrite {
	return staff.StaffWrite{Email: b.Email, DisplayName: b.DisplayName, Role: b.Role, Status: b.Status, CurrentPassword: b.CurrentPassword}
}
func parseID(id string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(id); err != nil || !u.Valid {
		return u, staff.ErrValidation
	}
	return u, nil
}
func selfRevocation(ctx context.Context, b *security.Boundary, a staff.Actor) {
	if _, err := b.Service.Authenticate(ctx, a.SessionHash); err != nil {
		b.ClearSession(ctx)
	}
}

type PageQuery struct {
	Page     int    `query:"page" default:"1" minimum:"1" maximum:"1000000"`
	PageSize int    `query:"pageSize" default:"25" minimum:"1" maximum:"100"`
	Q        string `query:"q" maxLength:"100"`
	Status   string `query:"status" enum:",active,disabled"`
}

func catalogReadError(ctx context.Context, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return security.Error(ctx, staff.ErrConflict)
	}
	return security.Error(ctx, staff.ErrUnavailable)
}

type ItemPath struct {
	ID string `path:"id" format:"uuid"`
}
