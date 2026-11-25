package staff

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/packages/application/internal/persistence"
)

func (s *Service) CreateStaff(ctx context.Context, actor Actor, input StaffWrite) (Staff, error) {
	in, err := normalizeWrite(input)
	if err != nil {
		return Staff{}, err
	}
	confirmed, err := s.confirmation(ctx, actor, in.CurrentPassword)
	if err != nil {
		return Staff{}, err
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
		locked, err := lockGovernance(ctx, q, actor, pgtype.UUID{}, "superadmin", at)
		if err != nil {
			return err
		}
		if !credentialMatches(confirmed, locked.credentials[actor.StaffID]) {
			return ErrDenied
		}
		row, err := q.StaffInsert(ctx, persistence.StaffInsertParams{ID: id, EmailNormalized: in.Email, DisplayName: in.DisplayName, Role: in.Role, Status: in.Status, CreatedAt: stamp(at)})
		if err != nil {
			return err
		}
		if err = q.StaffSetCredential(ctx, persistence.StaffSetCredentialParams{StaffID: id, PasswordHash: hash, ChangedAt: stamp(at)}); err != nil {
			return err
		}
		if err = audit(ctx, q, actor.StaffID, id, "staff.create", at, map[string]any{"changed_fields": []string{"email_normalized", "display_name", "role", "status", "credentials"}}); err != nil {
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

func (s *Service) UpdateStaff(ctx context.Context, actor Actor, id pgtype.UUID, input StaffWrite) (Staff, error) {
	if !id.Valid || input.Password != "" {
		return Staff{}, ErrValidation
	}
	in, err := normalizeWrite(input)
	if err != nil {
		return Staff{}, err
	}
	confirmed, err := s.confirmation(ctx, actor, in.CurrentPassword)
	if err != nil {
		return Staff{}, err
	}
	var result Staff
	err = s.transaction(ctx, func(q *persistence.Queries, _ pgx.Tx, at time.Time) error {
		locked, err := lockGovernance(ctx, q, actor, id, "superadmin", at)
		if err != nil {
			return err
		}
		if !credentialMatches(confirmed, locked.credentials[actor.StaffID]) {
			return ErrDenied
		}
		if !locked.target.ID.Valid || !matchRevision(locked.target, in.Revision) {
			return ErrConflict
		}
		if locked.target.Role != in.Role || locked.target.Status != in.Status {
			if err = q.StaffRevokeSessions(ctx, persistence.StaffRevokeSessionsParams{StaffID: id, Column2: stamp(at)}); err != nil {
				return err
			}
		}
		row, err := q.StaffUpdate(ctx, persistence.StaffUpdateParams{ID: id, EmailNormalized: in.Email, DisplayName: in.DisplayName, Role: in.Role, Status: in.Status, UpdatedAt: stamp(at)})
		if err != nil {
			return err
		}
		detail := map[string]any{"changed_fields": []string{"email_normalized", "display_name", "role", "status"}, "previous": map[string]string{"role": locked.target.Role, "status": locked.target.Status}, "current": map[string]string{"role": in.Role, "status": in.Status}}
		if err = audit(ctx, q, actor.StaffID, id, "staff.update", at, detail); err != nil {
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

func (s *Service) ReplacePassword(ctx context.Context, actor Actor, id pgtype.UUID, input StaffWrite) error {
	if !id.Valid {
		return ErrValidation
	}
	confirmed, err := s.confirmation(ctx, actor, input.CurrentPassword)
	if err != nil {
		return err
	}
	if !validPassword(input.Password) {
		return ErrValidation
	}
	target, err := persistence.New(s.pool).StaffReadCredential(ctx, id)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return safeError(err)
	}
	if err == nil {
		same, e := s.passwords.Verify(ctx, target.PasswordHash, input.Password)
		if e != nil {
			return e
		}
		if same {
			return ErrValidation
		}
	}
	hash, err := s.passwords.Hash(ctx, input.Password)
	if err != nil {
		return err
	}
	return s.transaction(ctx, func(q *persistence.Queries, _ pgx.Tx, at time.Time) error {
		locked, err := lockGovernance(ctx, q, actor, id, "superadmin", at)
		if err != nil {
			return err
		}
		if !credentialMatches(confirmed, locked.credentials[actor.StaffID]) {
			return ErrDenied
		}
		if !locked.target.ID.Valid || !matchRevision(locked.target, input.Revision) {
			return ErrConflict
		}
		current, exists := locked.credentials[id]
		if exists != target.StaffID.Valid || (exists && !credentialMatches(target, current)) {
			return ErrConflict
		}
		if err = q.StaffRevokeSessions(ctx, persistence.StaffRevokeSessionsParams{StaffID: id, Column2: stamp(at)}); err != nil {
			return err
		}
		if err = q.StaffSetCredential(ctx, persistence.StaffSetCredentialParams{StaffID: id, PasswordHash: hash, ChangedAt: stamp(at)}); err != nil {
			return err
		}
		return audit(ctx, q, actor.StaffID, id, "staff.password_replace", at, map[string]any{"changed_fields": []string{"credentials"}})
	})
}
