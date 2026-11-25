//go:build integration

package authtest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

const schemaUsers = `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at)
VALUES('10000000-0000-4000-8000-000000000001','one@example.test','one@example.test','One','active','2026-01-01','2026-01-01'),
('10000000-0000-4000-8000-000000000002','two@example.test','two@example.test','Two','active','2026-01-01','2026-01-01')`

func TestAuthSchemaTelephone(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		if _, err := f.Pool.Exec(ctx, schemaUsers); err != nil {
			return err
		}
		for _, phone := range []any{nil, "+6281234567890", "+12", "+123456789012345"} {
			if _, err := f.Pool.Exec(ctx, `UPDATE users SET telephone=$1`, phone); err != nil {
				return err
			}
		}
		if _, err := f.Pool.Exec(ctx, `UPDATE users SET telephone='+6281234567890'`); err != nil {
			return err
		}
		var nonunique bool
		if err := f.Pool.QueryRow(ctx, `SELECT count(*)=2 AND count(DISTINCT telephone)=1 FROM users`).Scan(&nonunique); err != nil {
			return err
		}
		if !nonunique {
			return errors.New("telephone unexpectedly unique")
		}
		for _, phone := range []string{"", "081234", "+0123", "+1", "+1234567890123456", "+62 8123", "+１２"} {
			_, err := f.Pool.Exec(ctx, `UPDATE users SET telephone=$1`, phone)
			var pg *pgconn.PgError
			if !errors.As(err, &pg) || pg.ConstraintName != "users_telephone_check" {
				return errors.New("canonical telephone constraint missing")
			}
		}
		var legacy bool
		if err := f.Pool.QueryRow(ctx, `SELECT bool_and(identity_version IS NULL) FROM users`).Scan(&legacy); err != nil {
			return err
		}
		if !legacy {
			return errors.New("legacy users were promoted")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("telephone schema contract failed: %s", schemaErrorKind(err))
	}
}

func TestAuthSchemaBoundedState(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		if _, err := f.Pool.Exec(ctx, schemaUsers); err != nil {
			return err
		}
		// Positive controls prevent a missing table/column from satisfying negatives.
		for _, query := range []string{
			`INSERT INTO identity_enrollment_receipts(secret_hash,created_at,expires_at) VALUES(decode(repeat('01',32),'hex'),'2026-01-01','2026-01-01 00:30Z')`,
			`INSERT INTO identity_limiter_buckets(key_hash,window_started_at,expires_at,attempts) VALUES(decode(repeat('01',32),'hex'),'2026-01-01','2026-01-01 01:00Z',1)`,
		} {
			if _, err := f.Pool.Exec(ctx, query); err != nil {
				return err
			}
		}
		for _, query := range []string{
			`INSERT INTO identity_enrollment_receipts(secret_hash,created_at,expires_at) VALUES(decode(repeat('02',32),'hex'),'2026-01-01','infinity')`,
			`INSERT INTO identity_limiter_buckets(key_hash,window_started_at,expires_at,attempts) VALUES(decode(repeat('02',32),'hex'),'2026-01-01','infinity',1)`,
			`UPDATE identity_enrollment_receipts SET user_id='10000000-0000-4000-8000-000000000001'`,
			`UPDATE identity_enrollment_receipts SET secret_hash=decode('01','hex')`,
			`UPDATE identity_enrollment_receipts SET expires_at='2026-01-01 00:30:01Z'`,
			`UPDATE identity_limiter_buckets SET attempts=0`,
			`UPDATE identity_limiter_buckets SET expires_at='2026-01-02 00:00:01Z'`,
			`UPDATE users SET identity_version=0`,
		} {
			_, err := f.Pool.Exec(ctx, query)
			var pg *pgconn.PgError
			if !errors.As(err, &pg) || pg.Code != "23514" {
				return errors.New("finite state check missing")
			}
		}
		var count int
		if err := f.Pool.QueryRow(ctx, `SELECT count(*) FROM identity_limiter_control WHERE id=1`).Scan(&count); err != nil {
			return err
		}
		if count != 1 {
			return errors.New("admission fence missing")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("bounded state schema contract failed: %s", schemaErrorKind(err))
	}
}

// Only fixed categories may enter test output; never relay SQL/driver values.
func schemaErrorKind(err error) string {
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		switch pg.Code {
		case "42703":
			return "missing column"
		case "42P01":
			return "missing relation"
		case "23514":
			return "unexpected check rejection"
		}
	}
	return "fixture or assertion failure"
}

func TestAuthSchemaGenerationBindings(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		if _, err := f.Pool.Exec(ctx, schemaUsers); err != nil {
			return err
		}
		for _, query := range []string{
			`INSERT INTO password_credentials(user_id,password_hash,password_changed_at) VALUES('10000000-0000-4000-8000-000000000001','fake legacy hash','2026-01-01')`,
			`INSERT INTO auth_sessions(id,user_id,secret_hash,authenticated_at,last_seen_at,expires_at,created_at) VALUES('20000000-0000-4000-8000-000000000001','10000000-0000-4000-8000-000000000001',decode('01','hex'),'2026-01-01','2026-01-01','2026-02-19','2026-01-01')`,
			`INSERT INTO email_action_tokens(id,user_id,purpose,target_email_normalized,token_hash,created_at,expires_at) VALUES('30000000-0000-4000-8000-000000000001','10000000-0000-4000-8000-000000000001','verify_email','one@example.test',decode('01','hex'),'2026-01-01','2026-01-02')`,
		} {
			if _, err := f.Pool.Exec(ctx, query); err != nil {
				return err
			}
		}
		var legacy bool
		if err := f.Pool.QueryRow(ctx, `SELECT c.generation IS NULL AND s.kind IS NULL AND t.identity_version IS NULL FROM password_credentials c JOIN auth_sessions s ON s.user_id=c.user_id JOIN email_action_tokens t ON t.user_id=c.user_id`).Scan(&legacy); err != nil {
			return err
		}
		if !legacy {
			return errors.New("legacy material promoted")
		}
		for _, query := range []string{
			`UPDATE users SET identity_version=1`,
			`UPDATE password_credentials SET generation=1`,
			`UPDATE auth_sessions SET kind='restricted',identity_version=1,credential_generation=1,last_activity_at='2026-01-01',browser_cookie_expires_at='2026-01-01 00:30Z',expires_at='2026-01-01 00:30Z',secret_hash=decode(repeat('01',32),'hex')`,
			`INSERT INTO email_action_tokens(id,user_id,purpose,target_email_normalized,token_hash,created_at,expires_at,identity_version,credential_generation) VALUES('30000000-0000-4000-8000-000000000002','10000000-0000-4000-8000-000000000002','setup_password','two@example.test',decode(repeat('02',32),'hex'),'2026-01-01','2026-01-02',1,0)`,
		} {
			if _, err := f.Pool.Exec(ctx, query); err != nil {
				return err
			}
		}
		for _, query := range []string{
			`UPDATE password_credentials SET generation=0`,
			`UPDATE auth_sessions SET kind='admin'`,
			`UPDATE auth_sessions SET expires_at='2026-02-19'`,
			`UPDATE auth_sessions SET browser_cookie_expires_at='infinity'`,
			`UPDATE email_action_tokens SET expires_at='2026-01-02 00:00:01Z' WHERE purpose='setup_password'`,
			`UPDATE email_action_tokens SET purpose='reset_password' WHERE purpose='setup_password'`,
			`UPDATE email_action_tokens SET credential_generation=NULL WHERE purpose='setup_password'`,
			`UPDATE email_action_tokens SET credential_generation=1 WHERE purpose='setup_password'`,
			`UPDATE email_action_tokens SET token_hash=decode('02','hex') WHERE purpose='setup_password'`,
		} {
			_, err := f.Pool.Exec(ctx, query)
			var pg *pgconn.PgError
			if !errors.As(err, &pg) || pg.Code != "23514" {
				return errors.New("generation or lifetime check missing")
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("generation schema contract failed: %s", schemaErrorKind(err))
	}
}
