package customers

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/packages/application/internal/persistence"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
	"time"
)

func audit(ctx context.Context, q *persistence.Queries, a staff.Actor, target, org pgtype.UUID, kind, action string, at time.Time, reason string, fields ...string) error {
	detail := map[string]any{"changed_fields": fields}
	if reason != "" {
		detail["reason"] = reason
	}
	b, err := json.Marshal(detail)
	if err != nil {
		return staff.ErrUnavailable
	}
	return q.CustomerAudit(ctx, persistence.CustomerAuditParams{OccurredAt: stamp(at), ActorStaffID: a.StaffID, OrganizationID: org, Action: action, TargetType: pgtype.Text{String: kind, Valid: true}, TargetID: pgtype.Text{String: target.String(), Valid: true}, Detail: b})
}
