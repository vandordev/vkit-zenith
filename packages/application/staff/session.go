package staff

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/packages/application/internal/persistence"
)

func (s *Service) Authenticate(ctx context.Context, hash [32]byte) (Staff, error) {
	r, err := persistence.New(s.pool).StaffAuthenticate(ctx, persistence.StaffAuthenticateParams{TokenHash: hash[:], CreatedAt: stamp(s.clock())})
	if errors.Is(err, pgx.ErrNoRows) {
		return Staff{}, ErrUnauthenticated
	}
	if err != nil {
		return Staff{}, safeError(err)
	}
	return Staff{ID: r.ID, Email: r.EmailNormalized, DisplayName: r.DisplayName, Role: r.Role, Status: r.Status, UpdatedAt: r.UpdatedAt.Time}, nil
}

func (s *Service) Logout(ctx context.Context, actor Actor) error {
	return s.transaction(ctx, func(q *persistence.Queries, _ pgx.Tx, at time.Time) error {
		if _, err := lockGovernance(ctx, q, actor, pgtype.UUID{}, "any_staff", at); err != nil {
			return err
		}
		if err := q.StaffLogout(ctx, persistence.StaffLogoutParams{StaffID: actor.StaffID, TokenHash: actor.SessionHash[:], Column3: stamp(at)}); err != nil {
			return err
		}
		return audit(ctx, q, actor.StaffID, actor.StaffID, "staff.logout", at, map[string]any{})
	})
}

func (s *Service) ChangePassword(ctx context.Context, actor Actor, input PasswordChange) error {
	confirmed, err := s.confirmation(ctx, actor, input.CurrentPassword)
	if err != nil {
		return err
	}
	if !validPassword(input.NewPassword) || input.NewPassword == input.CurrentPassword {
		return ErrValidation
	}
	hash, err := s.passwords.Hash(ctx, input.NewPassword)
	if err != nil {
		return err
	}
	return s.transaction(ctx, func(q *persistence.Queries, _ pgx.Tx, at time.Time) error {
		locked, err := lockGovernance(ctx, q, actor, actor.StaffID, "any_staff", at)
		if err != nil {
			return err
		}
		if !credentialMatches(confirmed, locked.credentials[actor.StaffID]) {
			return ErrDenied
		}
		if err = q.StaffRevokeSessions(ctx, persistence.StaffRevokeSessionsParams{StaffID: actor.StaffID, Column2: stamp(at)}); err != nil {
			return err
		}
		if err = q.StaffSetCredential(ctx, persistence.StaffSetCredentialParams{StaffID: actor.StaffID, PasswordHash: hash, ChangedAt: stamp(at)}); err != nil {
			return err
		}
		return audit(ctx, q, actor.StaffID, actor.StaffID, "staff.password_change", at, map[string]any{"changed_fields": []string{"credentials"}})
	})
}
