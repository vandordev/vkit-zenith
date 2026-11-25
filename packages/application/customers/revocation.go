package customers

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/packages/application/internal/persistence"
	"time"
)

// Uses the caller's confirmed transaction, never a separate connection/commit.
func revoke(ctx context.Context, q *persistence.Queries, id pgtype.UUID, at time.Time) error {
	if _, err := q.CustomerLockSessions(ctx, id); err != nil {
		return err
	}
	if _, err := q.CustomerLockEmailTokens(ctx, id); err != nil {
		return err
	}
	if err := q.CustomerRevokeSessions(ctx, persistence.CustomerRevokeSessionsParams{UserID: id, At: stamp(at)}); err != nil {
		return err
	}
	return q.CustomerRevokeEmailTokens(ctx, persistence.CustomerRevokeEmailTokensParams{UserID: id, At: stamp(at)})
}
