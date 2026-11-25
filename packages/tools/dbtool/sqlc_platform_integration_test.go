//go:build integration

package dbtool

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	platform "github.com/vandordev/vkit-zenith/packages/tools/dbtool/testdata/platform/generated"
)

func nativeUUID(n int) pgtype.UUID {
	var id pgtype.UUID
	if err := id.Scan(fixtureID(n)); err != nil {
		panic(err)
	}
	return id
}
func nativeTime(s string) pgtype.Timestamptz {
	v, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return pgtype.Timestamptz{Time: v, Valid: true}
}

func TestSQLCPlatformRoundTrip(t *testing.T) {
	schemaDatabase(t, func(ctx context.Context, p *pgxpool.Pool) error {
		if err := identityFixture(ctx, p); err != nil {
			return err
		}
		q := platform.New(p)
		if _, err := p.Exec(ctx, `INSERT INTO river_migration(line,version) VALUES ('fixture',1)`); err != nil {
			return err
		}
		migrations, err := q.PlatformRiverMigrationCatalog(ctx)
		if err != nil {
			return err
		}
		if len(migrations) != 1 || migrations[0].Line != "fixture" || migrations[0].Version != 1 || !migrations[0].CreatedAt.Valid {
			return errors.New("River migration catalog generated model")
		}
		user, err := q.PlatformUserByNormalizedEmail(ctx, "one@example.test")
		if err != nil {
			return err
		}
		if user.ID != nativeUUID(1) || !user.EmailVerifiedAt.Valid {
			return errors.New("UUID/timestamp native mapping")
		}
		token, err := q.PlatformEmailToken(ctx, nativeUUID(10))
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(token.TokenHash, []byte{1}) || !token.ExpiresAt.Valid || token.ConsumedAt.Valid || token.RevokedAt.Valid {
			return errors.New("bytea/nullable time mapping")
		}
		params := platform.PlatformInsertAuditParams{ID: nativeUUID(190), OccurredAt: nativeTime("2026-01-01T00:00:00Z"), Action: "fixture.generated", Detail: []byte(`{"reason":"fixture"}`)}
		err = schemaTx(ctx, p, func(tx pgx.Tx) error {
			bound := q.WithTx(tx)
			if err := bound.PlatformInsertAudit(ctx, params); err != nil {
				return err
			}
			row, err := bound.PlatformReadAudit(ctx, params.ID)
			if err != nil {
				return err
			}
			if !json.Valid(row.Detail) || row.ID != params.ID {
				return errors.New("JSONB/transaction round trip")
			}
			return fixtureFault
		})
		if !errors.Is(err, fixtureFault) {
			return err
		}
		if _, err = q.PlatformReadAudit(ctx, params.ID); !errors.Is(err, pgx.ErrNoRows) {
			return errors.New("generated mutation escaped InTx rollback")
		}
		if err := schemaTx(ctx, p, func(tx pgx.Tx) error { return q.WithTx(tx).PlatformInsertAudit(ctx, params) }); err != nil {
			return err
		}
		row, err := q.PlatformReadAudit(ctx, params.ID)
		if err != nil {
			return err
		}
		var detail map[string]string
		if err = json.Unmarshal(row.Detail, &detail); err != nil {
			return err
		}
		if detail["reason"] != "fixture" || !row.OccurredAt.Time.Equal(params.OccurredAt.Time) {
			return errors.New("JSONB/time committed round trip")
		}
		return nil
	})
}
