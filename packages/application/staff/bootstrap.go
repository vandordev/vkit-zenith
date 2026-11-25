package staff

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/vandordev/vkit-zenith/packages/application/internal/persistence"
)

func (s *Service) Bootstrap(ctx context.Context, input StaffWrite) (Staff, error) {
	in, err := normalizeWrite(input)
	if err != nil {
		return Staff{}, err
	}
	if in.Role != "superadmin" || in.Status != "active" {
		return Staff{}, ErrValidation
	}
	hash, err := s.passwords.Hash(ctx, in.Password)
	if err != nil {
		return Staff{}, err
	}
	id, err := newID()
	if err != nil {
		return Staff{}, err
	}
	var result Staff
	err = s.transaction(ctx, func(q *persistence.Queries, _ pgx.Tx, at time.Time) error {
		latch, err := q.StaffFence(ctx)
		if err != nil {
			return err
		}
		if latch.Valid {
			return ErrConflict
		}
		exists, err := q.StaffBootstrapExists(ctx)
		if err != nil {
			return err
		}
		if exists {
			return ErrConflict
		}
		row, err := q.StaffInsert(ctx, persistence.StaffInsertParams{ID: id, EmailNormalized: in.Email, DisplayName: in.DisplayName, Role: in.Role, Status: in.Status, CreatedAt: stamp(at)})
		if err != nil {
			return err
		}
		if err = q.StaffSetCredential(ctx, persistence.StaffSetCredentialParams{StaffID: id, PasswordHash: hash, ChangedAt: stamp(at)}); err != nil {
			return err
		}
		if err = audit(ctx, q, id, id, "staff.bootstrap", at, map[string]any{"changed_fields": []string{"email_normalized", "display_name", "role", "status", "credentials"}}); err != nil {
			return err
		}
		result = Staff{ID: row.ID, Email: row.EmailNormalized, DisplayName: row.DisplayName, Role: row.Role, Status: row.Status, UpdatedAt: row.UpdatedAt.Time}
		return nil
	})
	if err != nil {
		return Staff{}, err
	}
	return result, nil
}
