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

func auditInsert(id int, actor string, user, staff string, detail string) string {
	return fixtureSQL(fmt.Sprintf(`INSERT INTO audit_events (id,occurred_at,actor_kind,actor_user_id,actor_staff_id,action,outcome,detail) VALUES (@%d@,'2026-01-01','%s',%s,%s,'fixture.mutation','success','%s')`, id, actor, user, staff, detail))
}

func TestSchemaAudit(t *testing.T) {
	schemaContractDatabase(t, 1, func(ctx context.Context, p *pgxpool.Pool) error {
		tables := identityTables
		if err := checkPrefixOrFullTables(ctx, p, tables); err != nil {
			return err
		}
		if err := identityFixture(ctx, p); err != nil {
			return err
		}
		if err := insertStaffFixture(ctx, p, 90, 1); err != nil {
			return err
		}
		for i, actor := range []string{"system", "anonymous", "user", "staff"} {
			u, s := "NULL", "NULL"
			switch actor {
			case "user":
				u = "@1@"
			case "staff":
				s = "@90@"
			}
			if _, err := p.Exec(ctx, auditInsert(100+i, actor, u, s, `{"reason":"synthetic fixture","changed_fields":["status"],"previous":{"status":"active"},"current":{"status":"disabled"}}`)); err != nil {
				return err
			}
		}
		if err := checkNegatives(ctx, p, []negativeSQL{
			{"invalid actor", auditInsert(110, "unknown", "NULL", "NULL", `{}`), "23514", "audit_events_actor_check"},
			{"missing actor", auditInsert(110, "user", "NULL", "NULL", `{}`), "23514", "audit_events_actor_check"},
			{"mixed actor", auditInsert(110, "user", "@1@", "@90@", `{}`), "23514", "audit_events_actor_check"},
			{"system fake user", auditInsert(110, "system", "@1@", "NULL", `{}`), "23514", "audit_events_actor_check"},
			{"user FK", auditInsert(110, "user", "@199@", "NULL", `{}`), "23503", "audit_events_actor_user_fkey"},
			{"staff FK", auditInsert(110, "staff", "NULL", "@199@", `{}`), "23503", "audit_events_actor_staff_fkey"},
			{"organization FK", `INSERT INTO audit_events (id,occurred_at,actor_kind,organization_id,action,outcome,detail) VALUES (@110@,'2026-01-01','system',@199@,'fixture','success','{}')`, "23503", "audit_events_organization_fkey"},
			{"detail object", auditInsert(110, "system", "NULL", "NULL", `[]`), "23514", "audit_events_detail_check"},
			{"detail unknown key", auditInsert(110, "system", "NULL", "NULL", `{"password":"fake"}`), "23514", "audit_events_detail_check"},
			{"detail reason type", auditInsert(110, "system", "NULL", "NULL", `{"reason":{}}`), "23514", "audit_events_detail_check"},
			{"detail nested secret", auditInsert(110, "system", "NULL", "NULL", `{"previous":{"token_hash":"fake"}}`), "23514", "audit_events_detail_check"},
			{"detail field shape", auditInsert(110, "system", "NULL", "NULL", `{"changed_fields":[1]}`), "23514", "audit_events_detail_check"},
			{"detail oversized", `INSERT INTO audit_events (id,occurred_at,actor_kind,action,outcome,detail) VALUES (@110@,'2026-01-01','system','fixture','success',jsonb_build_object('reason',repeat('a',17000)))`, "23514", "audit_events_detail_check"},
			{"outcome", `INSERT INTO audit_events (id,occurred_at,actor_kind,action,outcome,detail) VALUES (@110@,'2026-01-01','system','fixture','unknown','{}')`, "23514", "audit_events_outcome_check"},
			{"target shape", `INSERT INTO audit_events (id,occurred_at,actor_kind,action,outcome,target_type,detail) VALUES (@110@,'2026-01-01','system','fixture','success','user','{}')`, "23514", "audit_events_target_check"},
		}); err != nil {
			return err
		}
		return nil
	})
}

func TestSchemaAuditNonowner(t *testing.T) {
	schemaDatabase(t, func(ctx context.Context, p *pgxpool.Pool) error {
		if _, err := p.Exec(ctx, `CREATE ROLE audit_fixture NOSUPERUSER NOLOGIN; GRANT USAGE ON SCHEMA public TO audit_fixture; GRANT SELECT,INSERT ON audit_events TO audit_fixture;`); err != nil {
			return err
		}
		c, err := p.Acquire(ctx)
		if err != nil {
			return err
		}
		defer c.Release()
		if _, err := c.Exec(ctx, "SET ROLE audit_fixture"); err != nil {
			return err
		}
		var super, owner, create bool
		if err := c.QueryRow(ctx, `SELECT r.rolsuper,c.relowner=r.oid,has_schema_privilege(current_user,'public','CREATE') FROM pg_roles r,pg_class c WHERE r.rolname=current_user AND c.oid='audit_events'::regclass`).Scan(&super, &owner, &create); err != nil {
			return err
		}
		if super || owner || create {
			return errors.New("audit fixture has excessive authority")
		}
		if _, err := c.Exec(ctx, auditInsert(100, "system", "NULL", "NULL", `{}`)); err != nil {
			return err
		}
		var count int
		if err := c.QueryRow(ctx, "SELECT count(*) FROM audit_events").Scan(&count); err != nil {
			return err
		}
		if count != 1 {
			return errors.New("audit SELECT/INSERT unavailable")
		}
		for _, statement := range []string{"UPDATE audit_events SET action='changed'", "DELETE FROM audit_events", "TRUNCATE audit_events"} {
			_, err := c.Exec(ctx, statement)
			if err := expectPG(err, "42501", ""); err != nil {
				return err
			}
		}
		if _, err := c.Exec(ctx, "RESET ROLE"); err != nil {
			return err
		}
		if _, err := c.Exec(ctx, "GRANT ALL ON TABLE audit_events TO audit_fixture; SET ROLE audit_fixture"); err != nil {
			return err
		}
		for _, statement := range []string{"UPDATE audit_events SET action='changed'", "DELETE FROM audit_events", "TRUNCATE audit_events"} {
			_, err := c.Exec(ctx, statement)
			if err := expectPG(err, "42501", ""); err != nil {
				return err
			}
		}
		if err := c.QueryRow(ctx, "SELECT count(*) FROM audit_events WHERE action='fixture.mutation'").Scan(&count); err != nil {
			return err
		}
		if count != 1 {
			return errors.New("audit mutated")
		}
		return nil
	})
}

func TestSchemaAuditAtomicity(t *testing.T) {
	schemaDatabase(t, func(ctx context.Context, p *pgxpool.Pool) error {
		if err := identityFixture(ctx, p); err != nil {
			return err
		}
		sentinel := errors.New("injected after audit insert")
		for _, fail := range []bool{true, false} {
			err := schemaTx(ctx, p, func(tx pgx.Tx) error {
				if _, err := tx.Exec(ctx, fixtureSQL(`UPDATE users SET display_name='Audited' WHERE id=@1@`)); err != nil {
					return err
				}
				if _, err := tx.Exec(ctx, auditInsert(100, "user", "@1@", "NULL", `{"changed_fields":["display_name"]}`)); err != nil {
					return err
				}
				if fail {
					return sentinel
				}
				return nil
			})
			if fail && !errors.Is(err, sentinel) {
				return errors.New("atomicity fault not preserved")
			}
			if !fail && err != nil {
				return err
			}
			var name string
			var count int
			if err := p.QueryRow(ctx, fixtureSQL(`SELECT display_name FROM users WHERE id=@1@`)).Scan(&name); err != nil {
				return err
			}
			if err := p.QueryRow(ctx, "SELECT count(*) FROM audit_events").Scan(&count); err != nil {
				return err
			}
			if fail && (name != "One" || count != 0) {
				return errors.New("mutation/audit partial rollback")
			}
			if !fail && (name != "Audited" || count != 1) {
				return errors.New("mutation/audit not committed together")
			}
		}
		return nil
	})
}
