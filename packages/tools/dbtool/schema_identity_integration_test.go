//go:build integration

package dbtool

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func fixtureID(n int) string { return fmt.Sprintf("00000000-0000-0000-0000-%012d", n) }

// @n in SQL fixtures denotes a Go-supplied deterministic fake UUID.
func fixtureSQL(sql string) string {
	for n := 200; n >= 1; n-- {
		sql = strings.ReplaceAll(sql, fmt.Sprintf("@%d@", n), "'"+fixtureID(n)+"'")
	}
	return sql
}
func fixtureExec(ctx context.Context, p *pgxpool.Pool, sql string) error {
	_, err := p.Exec(ctx, fixtureSQL(sql))
	return err
}

const fixtureTime = "2026-01-01T00:00:00Z"

var identityTables = []string{"users", "password_credentials", "email_action_tokens", "auth_sessions", "organizations", "organization_memberships"}

func checkTables(ctx context.Context, p *pgxpool.Pool, want []string) error {
	rows, err := p.Query(ctx, "SELECT tablename FROM pg_tables WHERE schemaname='public' AND tablename <> 'goose_db_version' ORDER BY tablename")
	if err != nil {
		return err
	}
	defer rows.Close()
	var actual []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return err
		}
		actual = append(actual, s)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	expected := append([]string{}, want...)
	sort.Strings(expected)
	if !reflect.DeepEqual(actual, expected) {
		return fmt.Errorf("table prefix mismatch: %v", actual)
	}
	for _, table := range expected {
		var count int
		if err := p.QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(&count); err != nil {
			return err
		}
		expectedCount := 0
		if table == "platform_staff_auth_control" {
			expectedCount = 1
			var valid bool
			if err := p.QueryRow(ctx, `SELECT id=1 AND mutation_epoch=0 AND bootstrapped_at IS NULL FROM platform_staff_auth_control`).Scan(&valid); err != nil {
				return err
			}
			if !valid {
				return errors.New("invalid initial staff singleton")
			}
		}
		if table == "identity_limiter_control" {
			expectedCount = 1
			var valid bool
			if err := p.QueryRow(ctx, `SELECT id=1 AND bucket_count=0 FROM identity_limiter_control`).Scan(&valid); err != nil {
				return err
			}
			if !valid {
				return errors.New("invalid initial customer admission singleton")
			}
		}
		if count != expectedCount {
			return fmt.Errorf("unexpected seed in %s", table)
		}
		var pk int
		if err := p.QueryRow(ctx, "SELECT count(*) FROM pg_constraint WHERE conrelid=$1::regclass AND contype='p'", table).Scan(&pk); err != nil {
			return err
		}
		if pk != 1 {
			return fmt.Errorf("missing PK: %s", table)
		}
	}
	return nil
}

func identityFixture(ctx context.Context, p *pgxpool.Pool) error {
	return fixtureExec(ctx, p, `
INSERT INTO users VALUES (@1@,'One@example.test','one@example.test','One','active','2026-01-01','2026-01-01','2026-01-01'),(@2@,'two@example.test','two@example.test','Two','active',NULL,'2026-01-01','2026-01-01');
INSERT INTO password_credentials VALUES (@1@,'fake-encoded-hash','2026-01-01');
INSERT INTO email_action_tokens VALUES (@10@,@1@,'verify_email','one@example.test','\x01','2026-01-02',NULL,NULL,'2026-01-01');
INSERT INTO auth_sessions VALUES (@11@,@1@,'\x01','2026-01-01','2026-01-01','2026-01-02',NULL,'2026-01-01'),(@12@,@2@,'\x02','2026-01-01','2026-01-01','2026-01-02',NULL,'2026-01-01');
INSERT INTO organizations VALUES (@30@,'First','first','active','2026-01-01','2026-01-01'),(@31@,'Second','second','active','2026-01-01','2026-01-01');
INSERT INTO organization_memberships VALUES (@32@,@30@,@1@,'active','owner','2026-01-01','2026-01-01'),(@33@,@31@,@2@,'active','owner','2026-01-01','2026-01-01');`)
}

type negativeSQL struct{ name, sql, code, constraint string }

func checkNegatives(ctx context.Context, p *pgxpool.Pool, cases []negativeSQL) error {
	for _, c := range cases {
		_, err := p.Exec(ctx, fixtureSQL(c.sql))
		if err = expectPG(err, c.code, c.constraint); err != nil {
			return fmt.Errorf("%s: %w", c.name, err)
		}
	}
	return nil
}

func TestSchemaIdentity(t *testing.T) {
	schemaContractDatabase(t, 1, func(ctx context.Context, p *pgxpool.Pool) error {
		if err := checkPrefixOrFullTables(ctx, p, identityTables); err != nil {
			return err
		}
		if err := identityCatalog(ctx, p); err != nil {
			return err
		}
		if err := identityFixture(ctx, p); err != nil {
			return err
		}
		if err := checkNegatives(ctx, p, []negativeSQL{
			{"email duplicate", `INSERT INTO users SELECT @3@,email,email_normalized,display_name,status,email_verified_at,created_at,updated_at FROM users WHERE id=@1@`, "23505", "users_email_normalized_key"},
			{"email malformed", `UPDATE users SET email='', email_normalized='' WHERE id=@1@`, "23514", "users_email_check"},
			{"user state", `UPDATE users SET status='pending' WHERE id=@1@`, "23514", "users_status_check"},
			{"user time", `UPDATE users SET updated_at='infinity' WHERE id=@1@`, "23514", "users_times_check"},
			{"empty password hash", `UPDATE password_credentials SET password_hash=''`, "23514", "password_credentials_hash_check"},
			{"token terminal", `UPDATE email_action_tokens SET consumed_at='2026-01-01',revoked_at='2026-01-01'`, "23514", "email_action_tokens_times_check"},
			{"token purpose", `UPDATE email_action_tokens SET purpose='invite'`, "23514", "email_action_tokens_purpose_check"},
			{"token expiry", `UPDATE email_action_tokens SET expires_at=created_at`, "23514", "email_action_tokens_times_check"},
			{"session expiry", `UPDATE auth_sessions SET expires_at=authenticated_at`, "23514", "auth_sessions_times_check"},
			{"organization state", `UPDATE organizations SET status='deleted'`, "23514", "organizations_status_check"},
			{"membership role", `UPDATE organization_memberships SET role='unknown' WHERE id=@32@`, "23514", "organization_memberships_role_check"},
			{"membership owner state", `UPDATE organization_memberships SET status='removed' WHERE id=@32@`, "23514", "organization_memberships_owner_check"},
		}); err != nil {
			return err
		}
		return nil
	})
}

func identityCatalog(ctx context.Context, p *pgxpool.Pool) error {
	columns := map[string]int{"users": 8, "password_credentials": 3, "email_action_tokens": 9, "auth_sessions": 8, "organizations": 7, "organization_memberships": 7}
	var version int
	if err := p.QueryRow(ctx, `SELECT max(version_id) FROM goose_db_version WHERE is_applied`).Scan(&version); err != nil {
		return err
	}
	if version >= 1 {
		for table, extra := range map[string]int{"users": 2, "password_credentials": 1, "email_action_tokens": 3, "auth_sessions": 5} {
			columns[table] += extra
		}
	}
	for table, want := range columns {
		var actual, defaults int
		if err := p.QueryRow(ctx, `SELECT count(*),count(column_default) FROM information_schema.columns WHERE table_schema='public' AND table_name=$1`, table).Scan(&actual, &defaults); err != nil {
			return err
		}
		if actual != want || defaults != 0 {
			return fmt.Errorf("column/default contract %s: got %d/%d", table, actual, defaults)
		}
	}
	for _, column := range []struct{ table, name, kind, nullable string }{
		{"email_action_tokens", "token_hash", "bytea", "NO"}, {"auth_sessions", "expires_at", "timestamp with time zone", "NO"}, {"organization_memberships", "role", "text", "NO"},
	} {
		var kind, nullable string
		if err := p.QueryRow(ctx, `SELECT data_type,is_nullable FROM information_schema.columns WHERE table_schema='public' AND table_name=$1 AND column_name=$2`, column.table, column.name).Scan(&kind, &nullable); err != nil {
			return err
		}
		if kind != column.kind || nullable != column.nullable {
			return fmt.Errorf("column type/nullability mismatch: %s.%s", column.table, column.name)
		}
	}
	return nil
}
