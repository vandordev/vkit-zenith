package staff

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/packages/application/internal/persistence"
)

// ConfirmedMutation admits one Superadmin action. Password work happens before
// transaction locks; current session, role and credential version are rechecked
// after admission. The callback must not retain or escape the transaction and
// must map its domain/constraint errors before returning them to this boundary.
// No confirmation proof or credential material leaves staff ownership.
func (s *Service) ConfirmedMutation(ctx context.Context, actor Actor, currentPassword string, fn func(context.Context, pgx.Tx, time.Time) error) error {
	if fn == nil {
		return ErrValidation
	}
	confirmed, err := s.confirmation(ctx, actor, currentPassword)
	if err != nil {
		return err
	}
	return s.transaction(ctx, func(q *persistence.Queries, tx pgx.Tx, at time.Time) error {
		locked, err := lockGovernance(ctx, q, actor, pgtype.UUID{}, "superadmin", at)
		if err != nil {
			return err
		}
		if !credentialMatches(confirmed, locked.credentials[actor.StaffID]) {
			return ErrDenied
		}
		return fn(ctx, tx, at)
	})
}
