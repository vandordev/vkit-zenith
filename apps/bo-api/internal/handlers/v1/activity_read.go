package v1

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/persistence"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
)

func activityRead(ctx context.Context, b *security.Boundary) (*persistence.Queries, error) {
	if _, err := b.Read(ctx, "superadmin"); err != nil {
		return nil, security.Error(ctx, err)
	}
	return persistence.New(b.Pool), nil
}

type ActivityInput struct {
	Page         int    `query:"page" default:"1" minimum:"1" maximum:"1000000"`
	PageSize     int    `query:"pageSize" default:"25" minimum:"1" maximum:"100"`
	From         string `query:"from" format:"date-time" maxLength:"64" required:"false"`
	Until        string `query:"until" format:"date-time" maxLength:"64" required:"false"`
	ActorStaffID string `query:"actorStaffId" format:"uuid" required:"false"`
	Action       string `query:"action" maxLength:"128"`
	Outcome      string `query:"outcome" enum:",success,failure,denied"`
}
type ActivityChoicesInput struct {
	Page       int    `query:"page" default:"1" minimum:"1" maximum:"1000000"`
	PageSize   int    `query:"pageSize" default:"25" minimum:"1" maximum:"100"`
	Q          string `query:"q" maxLength:"100"`
	SelectedID string `query:"selectedId" format:"uuid" required:"false"`
}

func activityFilters(in *ActivityInput) (p persistence.BOActivityRecordsParams, err error) {
	p.Actions = activityActions()
	p.PageLimit = int32(in.PageSize)
	p.PageOffset = int32((in.Page - 1) * in.PageSize)
	p.ActionFilter = in.Action
	p.OutcomeFilter = in.Outcome
	if in.Action != "" {
		if _, ok := activityPolicyFor(in.Action); !ok {
			return p, staff.ErrValidation
		}
	}
	if in.Outcome != "" && in.Outcome != "success" && in.Outcome != "failure" && in.Outcome != "denied" {
		return p, staff.ErrValidation
	}
	if p.ActorID, err = optionalID(in.ActorStaffID); err != nil {
		return p, err
	}
	parse := func(value string) (pgtype.Timestamptz, error) {
		if value == "" {
			return pgtype.Timestamptz{}, nil
		}
		t, e := time.Parse(time.RFC3339Nano, value)
		if e != nil || t.Year() < 1 || t.Year() > 9999 {
			return pgtype.Timestamptz{}, staff.ErrValidation
		}
		return pgtype.Timestamptz{Time: t.UTC(), Valid: true}, nil
	}
	if p.FromTime, err = parse(in.From); err != nil {
		return p, err
	}
	if p.UntilTime, err = parse(in.Until); err != nil {
		return p, err
	}
	if p.FromTime.Valid && p.UntilTime.Valid && !p.FromTime.Time.Before(p.UntilTime.Time) {
		return p, staff.ErrValidation
	}
	return p, nil
}
func activityCountFilters(p persistence.BOActivityRecordsParams) persistence.BOActivityCountParams {
	return persistence.BOActivityCountParams{Actions: p.Actions, EventID: p.EventID, FromTime: p.FromTime, UntilTime: p.UntilTime, ActorID: p.ActorID, ActionFilter: p.ActionFilter, OutcomeFilter: p.OutcomeFilter}
}
func activityFromRow(r persistence.BOActivityRecordsRow) activityRecord {
	return activityRecord{
		id: r.ID, occurredAt: r.OccurredAt.Time, action: r.Action, outcome: r.Outcome, actorKind: r.ActorKind,
		actorID: r.ActorStaffID, actorName: r.ActorName, targetType: r.TargetType,
		targetID: pgtype.Text{String: r.TargetID, Valid: r.TargetID != ""}, targetName: r.TargetName,
		targetOrganizationID: r.TargetOrganizationID,
		targetAvailable:      r.TargetAvailable,
		detail:               r.SafeDetail,
	}
}
func activitySnapshot(ctx context.Context, b *security.Boundary) (*persistence.Queries, pgx.Tx, error) {
	if _, err := b.Read(ctx, "superadmin"); err != nil {
		return nil, nil, security.Error(ctx, err)
	}
	tx, err := b.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, nil, catalogReadError(ctx, err)
	}
	return persistence.New(tx), tx, nil
}

func optionalID(raw string) (pgtype.UUID, error) {
	if raw == "" {
		return pgtype.UUID{}, nil
	}
	return parseID(raw)
}
