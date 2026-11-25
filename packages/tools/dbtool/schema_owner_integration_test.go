//go:build integration

package dbtool

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ownerCount(ctx context.Context, p *pgxpool.Pool) error {
	var n int
	err := p.QueryRow(ctx, `SELECT count(*) FROM organizations o WHERE NOT EXISTS (SELECT 1 FROM organization_memberships m JOIN users u ON u.id=m.user_id WHERE m.organization_id=o.id AND m.role='owner' AND m.status='active' AND u.status='active')`).Scan(&n)
	if err != nil {
		return err
	}
	if n != 0 {
		return errors.New("ownerless organization persisted")
	}
	return nil
}

// Direct pgx commit retains SQLSTATE; the public InTx boundary deliberately sanitizes it.
func directFixtureTx(ctx context.Context, p *pgxpool.Pool, sql string, immediate bool) error {
	tx, err := p.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, fixtureSQL(sql)); err != nil {
		return err
	}
	if immediate {
		if _, err = tx.Exec(ctx, "SET CONSTRAINTS ALL IMMEDIATE"); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func TestSchemaOwner(t *testing.T) {
	schemaDatabase(t, func(ctx context.Context, p *pgxpool.Pool) error {
		if err := identityFixture(ctx, p); err != nil {
			return err
		}
		for _, immediate := range []bool{false, true} {
			for _, sql := range []string{
				`INSERT INTO organizations VALUES (@90@,'Empty','empty','active','2026-01-01','2026-01-01')`,
				`UPDATE organization_memberships SET role='member' WHERE id=@32@`,
				`UPDATE organization_memberships SET role='member',status='removed' WHERE id=@32@`,
				`DELETE FROM organization_memberships WHERE id=@32@`,
				`UPDATE users SET status='disabled' WHERE id=@1@`,
				`UPDATE organizations SET status='suspended' WHERE id=@30@; UPDATE organization_memberships SET role='member' WHERE id=@32@`,
			} {
				if err := expectPG(directFixtureTx(ctx, p, sql, immediate), "23514", "organization_active_owner_check"); err != nil {
					return err
				}
			}
		}
		if err := checkNegatives(ctx, p, []negativeSQL{
			{"membership identity", `UPDATE organization_memberships SET organization_id=@31@ WHERE id=@32@`, "23514", "organization_owner_identity_check"},
			{"user removal restricted", `DELETE FROM users WHERE id=@1@`, "23503", ""},
			{"truncate", `TRUNCATE organization_memberships CASCADE`, "23514", "organization_active_owner_check"},
		}); err != nil {
			return err
		}
		if err := directFixtureTx(ctx, p, `INSERT INTO organization_memberships VALUES (@34@,@30@,@2@,'active','owner','2026-01-01','2026-01-01'); UPDATE organization_memberships SET role='member' WHERE id=@32@`, false); err != nil {
			return err
		}
		// One account owns two organizations; neither is exempt from validation.
		if err := expectPG(directFixtureTx(ctx, p, `UPDATE users SET status='disabled' WHERE id=@2@`, false), "23514", "organization_active_owner_check"); err != nil {
			return err
		}
		return ownerCount(ctx, p)
	})
}

func TestSchemaOwnerRaces(t *testing.T) {
	for _, iso := range []pgx.TxIsoLevel{pgx.ReadCommitted, pgx.RepeatableRead} {
		for _, mode := range []string{"remove", "delete", "remove-state", "disable", "insert-disable", "disable-insert", "multi-org"} {
			t.Run(fmt.Sprintf("%s/%s", iso, mode), func(t *testing.T) {
				schemaDatabase(t, func(ctx context.Context, p *pgxpool.Pool) error {
					if err := identityFixture(ctx, p); err != nil {
						return err
					}
					if err := fixtureExec(ctx, p, `INSERT INTO organization_memberships VALUES (@34@,@30@,@2@,'active','owner','2026-01-01','2026-01-01'),(@35@,@31@,@1@,'active','owner','2026-01-01','2026-01-01')`); err != nil {
						return err
					}
					first := `UPDATE organization_memberships SET role='member' WHERE id=@32@`
					second := `UPDATE organization_memberships SET role='member' WHERE id=@34@`
					if mode == "delete" {
						first = `DELETE FROM organization_memberships WHERE id=@32@`
						second = `DELETE FROM organization_memberships WHERE id=@34@`
					}
					if mode == "remove-state" {
						first = `UPDATE organization_memberships SET role='member',status='removed' WHERE id=@32@`
						second = `UPDATE organization_memberships SET role='member',status='removed' WHERE id=@34@`
					}
					if mode == "disable" || mode == "multi-org" {
						first = `UPDATE users SET status='disabled' WHERE id=@1@`
						second = `UPDATE users SET status='disabled' WHERE id=@2@`
					}
					if mode == "insert-disable" || mode == "disable-insert" {
						if err := fixtureExec(ctx, p, `DELETE FROM organization_memberships WHERE id=@35@`); err != nil {
							return err
						}
						first = `INSERT INTO organization_memberships VALUES (@35@,@31@,@1@,'active','owner','2026-01-01','2026-01-01'); UPDATE organization_memberships SET role='member' WHERE id=@33@`
						second = `UPDATE users SET status='disabled' WHERE id=@1@`
						if mode == "disable-insert" {
							first, second = second, first
						}
					}
					return withConnections(ctx, p, func(a, b *pgxpool.Conn) error {
						ta, err := a.BeginTx(ctx, pgx.TxOptions{IsoLevel: iso})
						if err != nil {
							return err
						}
						defer ta.Rollback(ctx)
						tb, err := b.BeginTx(ctx, pgx.TxOptions{IsoLevel: iso})
						if err != nil {
							return err
						}
						defer tb.Rollback(ctx)
						if _, err = tb.Exec(ctx, "SELECT count(*) FROM organizations"); err != nil {
							return err
						}
						if _, err = ta.Exec(ctx, fixtureSQL(first)); err != nil {
							return err
						}
						done := make(chan error, 1)
						go func() {
							_, e := tb.Exec(ctx, fixtureSQL(second))
							if e == nil {
								e = tb.Commit(ctx)
							}
							c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
							defer cancel()
							_ = tb.Rollback(c)
							done <- e
						}()
						barrier := waitBlocked(ctx, p, b.Conn().PgConn().PID())
						if barrier != nil {
							_ = ta.Rollback(ctx)
							<-done
							return barrier
						}
						winner := ta.Commit(ctx)
						loser := <-done
						if winner != nil {
							return winner
						}
						if iso == pgx.RepeatableRead {
							if err := expectPG(loser, "40001", ""); err != nil {
								return err
							}
						} else {
							if err := expectPG(loser, "23514", "organization_active_owner_check"); err != nil {
								return err
							}
						}
						return ownerCount(ctx, p)
					})
				})
			})
		}
	}
}
