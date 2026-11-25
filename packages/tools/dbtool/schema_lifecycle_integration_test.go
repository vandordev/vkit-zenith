//go:build integration

package dbtool

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var fixtureFault = errors.New("fixture injected transaction failure")

// Fixture-only algorithms: no protocol adapter, credentials, services or network effects.
func redeemFixture(ctx context.Context, tx pgx.Tx, kind, at, purpose, email, client string, fail bool) error {
	var id string
	var err error
	switch kind {
	case "email":
		err = tx.QueryRow(ctx, fixtureSQL(`UPDATE email_action_tokens e SET consumed_at=$1 FROM users u WHERE e.id=@10@ AND u.id=e.user_id AND u.status='active' AND e.purpose=$2 AND e.target_email_normalized=$3 AND u.email_normalized=e.target_email_normalized AND e.expires_at>$1 AND e.consumed_at IS NULL AND e.revoked_at IS NULL RETURNING e.user_id`), at, purpose, email).Scan(&id)
	default:
		return errors.New("unknown fixture kind")
	}
	if err != nil {
		return err
	}
	if fail {
		return fixtureFault
	}
	var effect string
	switch kind {
	case "email":
		effect = `UPDATE users SET email_verified_at='2026-01-01 12:00Z' WHERE id=@1@`
	}
	if _, err = tx.Exec(ctx, fixtureSQL(effect)); err != nil {
		return err
	}
	_, err = tx.Exec(ctx, fixtureSQL(`INSERT INTO audit_events VALUES (@190@,'2026-01-01 12:00Z','system',NULL,NULL,NULL,'fixture.redeem','success',NULL,NULL,NULL,'{}')`))
	return err
}

func snapshotFixture(ctx context.Context, p *pgxpool.Pool) (string, error) {
	rows, err := p.Query(ctx, `SELECT tablename FROM pg_tables WHERE schemaname='public' AND tablename<>'goose_db_version' ORDER BY tablename`)
	if err != nil {
		return "", err
	}
	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			rows.Close()
			return "", err
		}
		tables = append(tables, name)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return "", err
	}
	result := ""
	for _, table := range tables {
		var value string
		if err := p.QueryRow(ctx, `SELECT coalesce(jsonb_agg(v ORDER BY v::text),'[]'::jsonb)::text FROM (SELECT to_jsonb(t) v FROM `+pgx.Identifier{table}.Sanitize()+` t) s`).Scan(&value); err != nil {
			return "", err
		}
		result += table + value
	}
	return result, nil
}

func TestSchemaLifecycle(t *testing.T) {
	for _, kind := range []string{"email"} {
		t.Run(kind, func(t *testing.T) {
			schemaDatabase(t, func(ctx context.Context, p *pgxpool.Pool) error {
				if err := identityFixture(ctx, p); err != nil {
					return err
				}
				email := "two@example.test"
				if kind == "email" {
					email = "one@example.test"
				}
				before, err := snapshotFixture(ctx, p)
				if err != nil {
					return err
				}
				for _, c := range []struct {
					at, purpose, email, client string
					fail                       bool
					want                       error
				}{
					{"2026-01-02", "verify_email", email, fixtureID(40), false, pgx.ErrNoRows},
					{"2026-01-01 12:00Z", "verify_email", email, fixtureID(40), true, fixtureFault},
				} {
					err := schemaTx(ctx, p, func(tx pgx.Tx) error { return redeemFixture(ctx, tx, kind, c.at, c.purpose, c.email, c.client, c.fail) })
					if !errors.Is(err, c.want) {
						return fmt.Errorf("%s rollback/expiry: %w", kind, err)
					}
					after, e := snapshotFixture(ctx, p)
					if e != nil {
						return e
					}
					if before != after {
						return errors.New("failed redemption changed persistence")
					}
				}
				if kind == "email" {
					err := schemaTx(ctx, p, func(tx pgx.Tx) error {
						return redeemFixture(ctx, tx, kind, "2026-01-01 12:00Z", "reset_password", "wrong@example.test", fixtureID(40), false)
					})
					if !errors.Is(err, pgx.ErrNoRows) {
						return fmt.Errorf("wrong purpose/target accepted: %v", err)
					}
				}
				if kind == "email" {
					err := schemaTx(ctx, p, func(tx pgx.Tx) error {
						return redeemFixture(ctx, tx, kind, "2026-01-01 12:00Z", "reset_password", email, fixtureID(40), false)
					})
					if !errors.Is(err, pgx.ErrNoRows) {
						return errors.New("wrong email purpose accepted")
					}
					err = schemaTx(ctx, p, func(tx pgx.Tx) error {
						if _, e := tx.Exec(ctx, `UPDATE email_action_tokens SET revoked_at='2026-01-01'`); e != nil {
							return e
						}
						return redeemFixture(ctx, tx, kind, "2026-01-01 12:00Z", "verify_email", email, fixtureID(40), false)
					})
					if !errors.Is(err, pgx.ErrNoRows) {
						return errors.New("revoked email accepted")
					}
				}
				// Current disabled accounts are never redeemed, even with valid rows.
				err = schemaTx(ctx, p, func(tx pgx.Tx) error {
					uid := 2
					if kind == "email" {
						uid = 1
						if _, e := tx.Exec(ctx, fixtureSQL(`INSERT INTO organization_memberships VALUES (@183@,@30@,@2@,'active','owner','2026-01-01','2026-01-01')`)); e != nil {
							return e
						}
					}
					if uid == 2 {
						if _, e := tx.Exec(ctx, fixtureSQL(`INSERT INTO organization_memberships VALUES (@183@,@31@,@1@,'active','owner','2026-01-01','2026-01-01')`)); e != nil {
							return e
						}
					}
					if _, e := tx.Exec(ctx, `UPDATE users SET status='disabled' WHERE id=$1`, fixtureID(uid)); e != nil {
						return e
					}
					return redeemFixture(ctx, tx, kind, "2026-01-01 12:00Z", "verify_email", email, fixtureID(40), false)
				})
				if !errors.Is(err, pgx.ErrNoRows) {
					return fmt.Errorf("disabled account accepted: %v", err)
				}
				after, err := snapshotFixture(ctx, p)
				if err != nil {
					return err
				}
				if after != before {
					return errors.New("invalid lifecycle context changed persisted data")
				}
				return withConnections(ctx, p, func(a, b *pgxpool.Conn) error {
					ta, err := a.Begin(ctx)
					if err != nil {
						return err
					}
					defer ta.Rollback(ctx)
					tb, err := b.Begin(ctx)
					if err != nil {
						return err
					}
					defer tb.Rollback(ctx)
					if err = redeemFixture(ctx, ta, kind, "2026-01-01 12:00Z", "verify_email", email, fixtureID(40), false); err != nil {
						return err
					}
					done := make(chan error, 1)
					go func() {
						e := redeemFixture(ctx, tb, kind, "2026-01-01 12:00Z", "verify_email", email, fixtureID(40), false)
						if e == nil {
							e = tb.Commit(ctx)
						}
						_ = tb.Rollback(ctx)
						done <- e
					}()
					barrier := waitBlocked(ctx, p, b.Conn().PgConn().PID())
					if barrier != nil {
						_ = ta.Rollback(ctx)
						<-done
						return barrier
					}
					win := ta.Commit(ctx)
					lose := <-done
					if win != nil {
						return win
					}
					if !errors.Is(lose, pgx.ErrNoRows) {
						return fmt.Errorf("redemption loser: %v", lose)
					}
					var n int
					if err := p.QueryRow(ctx, `SELECT count(*) FROM audit_events`).Scan(&n); err != nil {
						return err
					}
					if n != 1 {
						return errors.New("one-time audit count")
					}
					return ownerCount(ctx, p)
				})
			})
		})
	}
}
