//go:build integration

package authtest

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Execute the actual application-owned SQL input, not an independently copied
// mock lock. sqlc generation/compilation separately proves its typed adapter.
func identitySchemaSQL(name string) (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("identity query source unavailable")
	}
	b, err := os.ReadFile(filepath.Join(filepath.Dir(file), "../../application/identity/locking.sql"))
	if err != nil {
		return "", errors.New("identity query source unavailable")
	}
	for _, part := range strings.Split(string(b), "-- name: ")[1:] {
		header, query, _ := strings.Cut(part, "\n")
		if strings.HasPrefix(header, name+" ") {
			return strings.TrimSpace(query), nil
		}
	}
	return "", errors.New("identity query missing")
}

func waitSchemaBlocked(ctx context.Context, p *pgxpool.Pool, pid uint32) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tick := time.NewTicker(5 * time.Millisecond)
	defer tick.Stop()
	for {
		var blocked bool
		if err := p.QueryRow(ctx, `SELECT cardinality(pg_blocking_pids($1))>0`, pid).Scan(&blocked); err != nil {
			return err
		}
		if blocked {
			return nil
		}
		select {
		case <-ctx.Done():
			return errors.New("writer did not reach lock barrier")
		case <-tick.C:
		}
	}
}

func TestAuthSchemaUserFence(t *testing.T) {
	lock, err := identitySchemaSQL("IdentityLockUser")
	if err != nil {
		t.Fatal("required identity user fence query missing")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	err = WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		if _, err := f.Pool.Exec(ctx, schemaUsers); err != nil {
			return err
		}
		a, err := f.Pool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer a.Release()
		b, err := f.Pool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer b.Release()
		ta, err := a.Begin(ctx)
		if err != nil {
			return err
		}
		defer ta.Rollback(ctx)
		if _, err = ta.Exec(ctx, lock, "10000000-0000-4000-8000-000000000001"); err != nil {
			return err
		}
		tb, err := b.Begin(ctx)
		if err != nil {
			return err
		}
		defer tb.Rollback(ctx)
		done := make(chan error, 1)
		go func() {
			_, e := tb.Exec(ctx, `UPDATE users SET status='disabled' WHERE id='10000000-0000-4000-8000-000000000001'`)
			done <- e
		}()
		barrier := waitSchemaBlocked(ctx, f.Pool, b.Conn().PgConn().PID())
		if err = ta.Commit(ctx); err != nil {
			cancel()
			<-done
			return err
		}
		if err = <-done; err != nil {
			return err
		}
		if barrier != nil {
			return barrier
		}
		if err = tb.Commit(ctx); err != nil {
			return err
		}
		// Next holder sees committed disable. This proves only SQL fence/recheck
		// feasibility; actual customer issuer/BO races remain Tasks 12–13 gates.
		ta, err = a.Begin(ctx)
		if err != nil {
			return err
		}
		defer ta.Rollback(ctx)
		if _, err = ta.Exec(ctx, lock, "10000000-0000-4000-8000-000000000001"); err != nil {
			return err
		}
		var status string
		if err = ta.QueryRow(ctx, `SELECT status FROM users WHERE id='10000000-0000-4000-8000-000000000001'`).Scan(&status); err != nil {
			return err
		}
		if status != "disabled" {
			return errors.New("fence admitted stale eligibility")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("user fence contract failed: %s", schemaErrorKind(err))
	}
}

func TestAuthSchemaLimiterCapacityRace(t *testing.T) {
	lock, err := identitySchemaSQL("IdentityLockLimiterControl")
	if err != nil {
		t.Fatal("required identity limiter fence query missing")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	err = WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		// Use real rows at the real production cap, not a reduced test constant.
		for start := 1; start <= 90001; start += 10000 {
			end := start + 9999
			if end > 99999 {
				end = 99999
			}
			_, err := f.Pool.Exec(ctx, `INSERT INTO identity_limiter_buckets(key_hash,window_started_at,expires_at,attempts) SELECT sha256(int4send(i)),'2026-01-01'::timestamptz,'2026-01-01 01:00Z'::timestamptz,1 FROM generate_series($1::int,$2::int) i`, start, end)
			if err != nil {
				return err
			}
		}
		a, err := f.Pool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer a.Release()
		b, err := f.Pool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer b.Release()
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
		if _, err = ta.Exec(ctx, lock); err != nil {
			return err
		}
		insert := `INSERT INTO identity_limiter_buckets VALUES(sha256(int4send($1::int)),'2026-01-01','2026-01-01 01:00Z',1)`
		if _, err = ta.Exec(ctx, insert, 100000); err != nil {
			return err
		}
		done := make(chan error, 1)
		go func() {
			_, e := tb.Exec(ctx, lock)
			if e == nil {
				_, e = tb.Exec(ctx, insert, 100001)
			}
			done <- e
		}()
		barrier := waitSchemaBlocked(ctx, f.Pool, b.Conn().PgConn().PID())
		if err = ta.Commit(ctx); err != nil {
			cancel()
			<-done
			return err
		}
		loser := <-done
		if barrier != nil {
			return barrier
		}
		var pg *pgconn.PgError
		if !errors.As(loser, &pg) || pg.Code != "23514" || pg.ConstraintName != "identity_limiter_capacity_check" {
			return errors.New("concurrent bucket capacity exceeded")
		}
		if err = tb.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
			return err
		}
		var count, actual int
		if err = f.Pool.QueryRow(ctx, `SELECT bucket_count,(SELECT count(*) FROM identity_limiter_buckets) FROM identity_limiter_control`).Scan(&count, &actual); err != nil {
			return err
		}
		if count != 100000 || actual != 100000 {
			return errors.New("bucket cap accounting diverged")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("capacity race contract failed: %s", schemaErrorKind(err))
	}
}

func TestAuthSchemaActionCapacityRace(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		if _, err := f.Pool.Exec(ctx, schemaUsers); err != nil {
			return err
		}
		insert := `INSERT INTO email_action_tokens(id,user_id,purpose,target_email_normalized,token_hash,created_at,expires_at,identity_version,credential_generation) VALUES(gen_random_uuid(),'10000000-0000-4000-8000-000000000001','reset_password','one@example.test',sha256(int4send($1::int)),'2026-01-01','2026-01-01 00:30Z',1,1)`
		for i := 1; i <= 9; i++ {
			if _, err := f.Pool.Exec(ctx, insert, i); err != nil {
				return err
			}
		}
		a, err := f.Pool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer a.Release()
		b, err := f.Pool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer b.Release()
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
		if _, err = ta.Exec(ctx, insert, 10); err != nil {
			return err
		}
		done := make(chan error, 1)
		go func() { _, e := tb.Exec(ctx, insert, 11); done <- e }()
		barrier := waitSchemaBlocked(ctx, f.Pool, b.Conn().PgConn().PID())
		if err = ta.Commit(ctx); err != nil {
			cancel()
			<-done
			return err
		}
		loser := <-done
		if barrier != nil {
			return barrier
		}
		var pg *pgconn.PgError
		if !errors.As(loser, &pg) || pg.Code != "23514" || pg.ConstraintName != "email_action_tokens_capacity_check" {
			return errors.New("concurrent action cap not enforced")
		}
		if err = tb.Rollback(ctx); err != nil {
			return err
		}
		var count int
		if err = f.Pool.QueryRow(ctx, `SELECT count(*) FROM email_action_tokens`).Scan(&count); err != nil {
			return err
		}
		if count != 10 {
			return errors.New("action cap overshot")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("action cap contract failed: %s", schemaErrorKind(err))
	}
}

// These are real SQL-lock barrier contracts, not claims that the future issuer
// already composes/authorizes the entire chain. Task 12/13 must exercise it again
// through actual commands and provider hooks, including BO/reset races.
func TestAuthSchemaDownstreamFences(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		_, err := f.Pool.Exec(ctx, schemaUsers+`;
INSERT INTO password_credentials VALUES('10000000-0000-4000-8000-000000000001','fake-legacy','2026-01-01',NULL);
INSERT INTO auth_sessions(id,user_id,secret_hash,authenticated_at,last_seen_at,expires_at,created_at) VALUES('20000000-0000-4000-8000-000000000001','10000000-0000-4000-8000-000000000001',decode('01','hex'),'2026-01-01','2026-01-01','2026-01-02','2026-01-01');
INSERT INTO email_action_tokens(id,user_id,purpose,target_email_normalized,token_hash,created_at,expires_at) VALUES('30000000-0000-4000-8000-000000000001','10000000-0000-4000-8000-000000000001','verify_email','one@example.test',decode('01','hex'),'2026-01-01','2026-01-02')`)
		if err != nil {
			return err
		}
		for _, item := range []struct {
			name, update string
			arg          any
		}{
			{"IdentityLockCredential", `UPDATE password_credentials SET password_hash='changed-fake'`, "10000000-0000-4000-8000-000000000001"},
			{"IdentityLockSession", `UPDATE auth_sessions SET revoked_at='2026-01-01'`, "20000000-0000-4000-8000-000000000001"},
			{"IdentityLockEmailActions", `UPDATE email_action_tokens SET revoked_at='2026-01-01'`, "10000000-0000-4000-8000-000000000001"},
		} {
			query, err := identitySchemaSQL(item.name)
			if err != nil {
				return err
			}
			// Each subcase rolls both transactions back so its complete fixture
			// remains intact and the next barrier cannot pass on zero matched rows.
			err = func() error {
				a, err := f.Pool.Acquire(ctx)
				if err != nil {
					return err
				}
				defer a.Release()
				b, err := f.Pool.Acquire(ctx)
				if err != nil {
					return err
				}
				defer b.Release()
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
				rows, err := ta.Query(ctx, query, item.arg)
				if err != nil {
					return err
				}
				n := 0
				for rows.Next() {
					n++
				}
				rows.Close()
				if rows.Err() != nil {
					return rows.Err()
				}
				if n != 1 {
					return errors.New("fence did not match exactly one fixture")
				}
				done := make(chan error, 1)
				go func() {
					tag, e := tb.Exec(ctx, item.update)
					if e == nil && tag.RowsAffected() != 1 {
						e = errors.New("writer missed fixture")
					}
					done <- e
				}()
				barrier := waitSchemaBlocked(ctx, f.Pool, b.Conn().PgConn().PID())
				if err = ta.Rollback(ctx); err != nil {
					cancel()
					<-done
					return err
				}
				if err = <-done; err != nil {
					return err
				}
				return barrier
			}()
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("downstream fence contract failed: %s", schemaErrorKind(err))
	}
}

func TestAuthSchemaActionStaleSnapshot(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	err := WithFixture(ctx, func(ctx context.Context, f *Fixture) error {
		if _, err := f.Pool.Exec(ctx, schemaUsers); err != nil {
			return err
		}
		a, err := f.Pool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer a.Release()
		b, err := f.Pool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer b.Release()
		ta, err := a.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
		if err != nil {
			return err
		}
		defer ta.Rollback(ctx)
		tb, err := b.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
		if err != nil {
			return err
		}
		defer tb.Rollback(ctx)
		for _, tx := range []pgx.Tx{ta, tb} {
			var n int
			if err = tx.QueryRow(ctx, `SELECT count(*) FROM users`).Scan(&n); err != nil {
				return err
			}
			if n != 2 {
				return errors.New("snapshot fixture missing")
			}
		}
		insert := `INSERT INTO email_action_tokens(id,user_id,purpose,target_email_normalized,token_hash,created_at,expires_at,identity_version,credential_generation) VALUES(gen_random_uuid(),'10000000-0000-4000-8000-000000000001','reset_password','one@example.test',sha256(int4send($1::int)),'2026-01-01','2026-01-01 00:30Z',1,1)`
		if _, err = ta.Exec(ctx, insert, 1); err != nil {
			return err
		}
		done := make(chan error, 1)
		go func() { _, e := tb.Exec(ctx, insert, 2); done <- e }()
		barrier := waitSchemaBlocked(ctx, f.Pool, b.Conn().PgConn().PID())
		if err = ta.Commit(ctx); err != nil {
			cancel()
			<-done
			return err
		}
		loser := <-done
		if barrier != nil {
			return barrier
		}
		var pg *pgconn.PgError
		if !errors.As(loser, &pg) || pg.Code != "40001" {
			return errors.New("stale action writer accepted obsolete snapshot")
		}
		if err = tb.Rollback(ctx); err != nil {
			return err
		}
		var count int
		if err = f.Pool.QueryRow(ctx, `SELECT count(*) FROM email_action_tokens`).Scan(&count); err != nil {
			return err
		}
		if count != 1 {
			return errors.New("obsolete action was committed")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot fence contract failed: %s", schemaErrorKind(err))
	}
}
