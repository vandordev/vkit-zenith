//go:build integration

package dbtool

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestSchemaStaffAuth(t *testing.T) {
	schemaDatabase(t, func(ctx context.Context, p *pgxpool.Pool) error {
		var n int
		if err := p.QueryRow(ctx, `SELECT count(*) FROM pg_tables WHERE schemaname='public' AND tablename<>'goose_db_version'`).Scan(&n); err != nil {
			return err
		}
		if n != 21 {
			return errors.New("current schema including River requires 21 tables")
		}
		if err := fixtureExec(ctx, p, `INSERT INTO platform_staff(id,email_normalized,display_name,status,role,created_at,updated_at) VALUES
  (@1@,'one@example.test','One','active','superadmin','2026-01-01','2026-01-01'),
  (@2@,'two@example.test','Two','active','superadmin','2026-01-01','2026-01-01');`); err != nil {
			return err
		}
		if err := checkNegatives(ctx, p, []negativeSQL{
			{"last active", `UPDATE platform_staff SET status='disabled'`, "23514", "platform_staff_last_superadmin_check"},
			{"last role", `UPDATE platform_staff SET role='admin'`, "23514", "platform_staff_last_superadmin_check"},
			{"latch clear", `UPDATE platform_staff_auth_control SET bootstrapped_at=NULL`, "23514", "platform_staff_bootstrap_latch_check"},
			{"control delete", `DELETE FROM platform_staff_auth_control`, "42501", ""},
			{"staff truncate", `TRUNCATE platform_staff CASCADE`, "42501", ""},
			{"normalized email", `UPDATE platform_staff SET email_normalized='UPPER@example.test' WHERE id=@1@`, "23514", "platform_staff_email_check"},
			{"role detail", `INSERT INTO audit_events(id,occurred_at,actor_kind,action,outcome,detail) VALUES (@99@,now(),'system','test','success','{"current":{"role":"owner"}}')`, "23514", "audit_events_detail_check"},
		}); err != nil {
			return err
		}
		for _, isolation := range []pgx.TxIsoLevel{pgx.ReadCommitted, pgx.RepeatableRead} {
			if err := withConnections(ctx, p, func(a, b *pgxpool.Conn) error {
				ta, err := a.BeginTx(ctx, pgx.TxOptions{IsoLevel: isolation})
				if err != nil {
					return err
				}
				defer ta.Rollback(ctx)
				tb, err := b.BeginTx(ctx, pgx.TxOptions{IsoLevel: isolation})
				if err != nil {
					return err
				}
				defer tb.Rollback(ctx)
				// Establish both snapshots before first governance write.
				for _, tx := range []pgx.Tx{ta, tb} {
					var n int
					if err = tx.QueryRow(ctx, `SELECT count(*) FROM platform_staff`).Scan(&n); err != nil {
						return err
					}
				}
				if _, err = ta.Exec(ctx, fixtureSQL(`UPDATE platform_staff SET role='admin' WHERE id=@1@`)); err != nil {
					return err
				}
				done := make(chan error, 1)
				go func() {
					_, e := tb.Exec(ctx, fixtureSQL(`UPDATE platform_staff SET role='admin' WHERE id=@2@`))
					if e == nil {
						e = tb.Commit(ctx)
					}
					_ = tb.Rollback(ctx)
					done <- e
				}()
				if err = waitBlocked(ctx, p, b.Conn().PgConn().PID()); err != nil {
					_ = ta.Rollback(ctx)
					<-done
					return err
				}
				if err = ta.Commit(ctx); err != nil {
					<-done
					return err
				}
				loser := <-done
				code := "23514"
				constraint := "platform_staff_last_superadmin_check"
				if isolation == pgx.RepeatableRead {
					code = "40001"
					constraint = ""
				}
				if err = expectPG(loser, code, constraint); err != nil {
					return err
				}
				_, err = p.Exec(ctx, fixtureSQL(`UPDATE platform_staff SET role='superadmin' WHERE id=@1@`))
				return err
			}); err != nil {
				return err
			}
		}
		// SQL-level revocation does not depend on future login/HTTP adapters.
		if err := fixtureExec(ctx, p, `INSERT INTO platform_staff_credentials VALUES (@1@,'$argon2id$v=19$m=65536,t=3,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA','2026-01-01');`); err != nil {
			return err
		}
		for i, mutation := range []string{
			`UPDATE platform_staff SET role='admin' WHERE id=@1@`,
			`UPDATE platform_staff SET status='disabled' WHERE id=@1@`,
			`UPDATE platform_staff_credentials SET changed_at='2026-01-02' WHERE staff_id=@1@`,
			`DELETE FROM platform_staff_credentials WHERE staff_id=@1@`,
		} {
			if err := fixtureExec(ctx, p, `UPDATE platform_staff SET role='superadmin',status='active' WHERE id=@1@`); err != nil {
				return err
			}
			if _, err := p.Exec(ctx, fixtureSQL(`INSERT INTO platform_staff_sessions(id,staff_id,token_hash,created_at,expires_at) VALUES ($1,@1@,$2,'2026-01-01','2026-01-31')`), nativeUUID(100+i), makeHashFixture(i)); err != nil {
				return err
			}
			if err := fixtureExec(ctx, p, mutation); err != nil {
				return err
			}
			var active int
			if err := p.QueryRow(ctx, `SELECT count(*) FROM platform_staff_sessions WHERE revoked_at IS NULL`).Scan(&active); err != nil {
				return err
			}
			if active != 0 {
				return errors.New("direct SQL failed all-session revocation")
			}
			if err := checkNegatives(ctx, p, []negativeSQL{{"revival", `UPDATE platform_staff_sessions SET revoked_at=NULL`, "23514", "platform_staff_sessions_immutable_check"}}); err != nil {
				return err
			}
		}
		return nil
	})
}

func makeHashFixture(i int) []byte { b := make([]byte, 32); b[0] = byte(i + 1); return b }
