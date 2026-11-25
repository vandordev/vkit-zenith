//go:build integration

package dbtool

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pressly/goose/v3/lock"
	db "github.com/vandordev/vkit-zenith/packages/db"
	"github.com/vandordev/vkit-zenith/packages/db/migrate"
)

func settings(dsn string) db.Settings {
	return db.Settings{DSN: dsn, MaxOpenConnections: 4, ConnectionMaxLifetimeSeconds: 60, ConnectionMaxIdleTimeSeconds: 10, ConnectTimeoutSeconds: 5}
}
func isolated(t *testing.T, fn func(context.Context, string) error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if err := WithPostgres(ctx, fn); err != nil {
		t.Fatalf("isolated PostgreSQL boundary: %v", err)
	}
}

func TestPoolAndTransactionPostgres(t *testing.T) {
	isolated(t, func(ctx context.Context, dsn string) error {
		a, err := db.Open(ctx, settings(dsn))
		if err != nil {
			return err
		}
		defer a.Close()
		b, err := db.Open(ctx, settings(dsn))
		if err != nil {
			return err
		}
		defer b.Close()
		if a == b || a.Config().MaxConns != 4 {
			return errors.New("pool independence/settings")
		}
		operational, err := migrate.Open(ctx, settings(dsn))
		if err != nil {
			return err
		}
		defer operational.Close()
		if _, err := migrate.Run(ctx, operational, os.DirFS("testdata/migrations"), migrate.Request{Operation: migrate.Up}); err != nil {
			return err
		}
		sentinel := errors.New("callback failed")
		for _, mode := range []string{"commit", "rollback", "panic", "cancel"} {
			var got error
			panicked := false
			var panicErr error
			func() {
				defer func() {
					if p := recover(); p != nil {
						if p != sentinel {
							panicErr = errors.New("wrong panic")
						}
						panicked = true
					}
				}()
				got = db.InTx(ctx, a, pgx.TxOptions{}, func(c context.Context, tx pgx.Tx) error {
					if _, err := tx.Exec(c, "INSERT INTO fixture_entries(label) VALUES ($1)", mode); err != nil {
						return err
					}
					switch mode {
					case "rollback":
						return sentinel
					case "panic":
						panic(sentinel)
					case "cancel":
						cancelled, stop := context.WithCancel(c)
						stop()
						_, err := tx.Exec(cancelled, "SELECT 1")
						return err
					}
					return nil
				})
			}()
			if panicErr != nil {
				return panicErr
			}
			if mode == "commit" && got != nil {
				return got
			}
			if mode == "rollback" && !errors.Is(got, sentinel) {
				return errors.New("primary lost")
			}
			if mode == "panic" && !panicked {
				return errors.New("panic swallowed")
			}
			var count int
			if err := b.QueryRow(ctx, "SELECT count(*) FROM fixture_entries WHERE label=$1", mode).Scan(&count); err != nil {
				return err
			}
			want := 0
			if mode == "commit" {
				want = 1
			}
			if count != want {
				return fmt.Errorf("%s: rows %d want %d", mode, count, want)
			}
		}
		rows, err := a.Query(ctx, "SELECT label FROM fixture_entries")
		if err != nil {
			return err
		}
		if !rows.Next() {
			return errors.New("pgx query empty")
		}
		rows.Close()
		cancelled, stop := context.WithCancel(ctx)
		stop()
		if _, err := db.Open(cancelled, settings(dsn)); !errors.Is(err, context.Canceled) {
			return errors.New("cancelled open")
		}
		if _, err := a.Exec(cancelled, "SELECT 1"); !errors.Is(err, context.Canceled) {
			return errors.New("cancelled query")
		}
		short, stopQuery := context.WithTimeout(ctx, 30*time.Millisecond)
		_, queryErr := a.Exec(short, "SELECT pg_sleep(2)")
		stopQuery()
		if queryErr == nil {
			return errors.New("in-flight query cancellation ignored")
		}
		badURL, _ := url.Parse(dsn)
		badURL.User = url.UserPassword("fixture", "sentinel-invalid-password")
		failed, openErr := db.Open(ctx, settings(badURL.String()))
		if openErr == nil || failed != nil || strings.Contains(openErr.Error(), "sentinel-invalid-password") {
			return errors.New("failed open retained pool or leaked credentials")
		}
		a.Close()
		if a.Ping(ctx) == nil {
			return errors.New("closed pool ping")
		}
		return nil
	})
}

func TestMigrationPostgres(t *testing.T) {
	isolated(t, func(ctx context.Context, dsn string) error {
		a, err := migrate.Open(ctx, settings(dsn))
		if err != nil {
			return err
		}
		defer a.Close()
		b, err := migrate.Open(ctx, settings(dsn))
		if err != nil {
			return err
		}
		defer b.Close()
		source := os.DirFS("testdata/migrations")
		conn, err := a.Conn(ctx)
		if err != nil {
			return err
		}
		defer conn.Close()
		l, err := lock.NewPostgresSessionLocker()
		if err != nil {
			return err
		}
		if err := l.SessionLock(ctx, conn); err != nil {
			return err
		}
		short, stop := context.WithTimeout(ctx, 200*time.Millisecond)
		_, err = migrate.Run(short, b, source, migrate.Request{Operation: migrate.Up})
		stop()
		if !errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("lock contention did not time out: %w", err)
		}
		var exists bool
		if err := a.QueryRowContext(ctx, "SELECT to_regclass('public.fixture_entries') IS NOT NULL").Scan(&exists); err != nil {
			return err
		}
		if exists {
			return errors.New("contended migration applied")
		}
		if err := l.SessionUnlock(ctx, conn); err != nil {
			return err
		}
		var wg sync.WaitGroup
		results := make(chan error, 2)
		for _, pool := range []*sql.DB{a, b} {
			wg.Add(1)
			go func(p *sql.DB) {
				defer wg.Done()
				_, err := migrate.Run(ctx, p, source, migrate.Request{Operation: migrate.Up})
				results <- err
			}(pool)
		}
		wg.Wait()
		close(results)
		for err := range results {
			if err != nil {
				return err
			}
		}
		status, err := migrate.Run(ctx, a, source, migrate.Request{Operation: migrate.Status})
		if err != nil {
			return err
		}
		if status.Version != 1 || len(status.Entries) != 1 || !status.Entries[0].Applied {
			return errors.New("incorrect status")
		}
		if _, err := migrate.Run(ctx, a, source, migrate.Request{Operation: migrate.Up}); err != nil {
			return err
		}
		bad := fstest.MapFS{"00002_bad.sql": {Data: []byte("-- +goose Up\nTHIS IS NOT SQL;\n-- +goose Down\nSELECT 1;\n")}}
		if _, err := migrate.Run(ctx, a, bad, migrate.Request{Operation: migrate.Up}); err == nil {
			return errors.New("invalid SQL succeeded")
		}
		var count int
		if err := a.QueryRowContext(ctx, "SELECT count(*) FROM goose_db_version WHERE version_id=2 AND is_applied").Scan(&count); err != nil {
			return err
		}
		if count != 0 {
			return errors.New("failed migration recorded")
		}
		result, err := migrate.Run(ctx, a, source, migrate.Request{Operation: migrate.DownTo, Target: 0, ConfirmRollback: true})
		if err != nil {
			return err
		}
		if result.Version != 0 {
			return errors.New("rollback version")
		}
		if err := a.QueryRowContext(ctx, "SELECT to_regclass('public.fixture_entries') IS NOT NULL").Scan(&exists); err != nil {
			return err
		}
		if exists {
			return errors.New("Down not applied")
		}
		return nil
	})
}

func TestTransactionCommitFailureAndZeroDurationReuse(t *testing.T) {
	isolated(t, func(ctx context.Context, dsn string) error {
		s := settings(dsn)
		s.MaxOpenConnections = 1
		s.ConnectionMaxLifetimeSeconds, s.ConnectionMaxIdleTimeSeconds = 0, 0
		p, err := db.Open(ctx, s)
		if err != nil {
			return err
		}
		defer p.Close()
		var first, second int
		if err := p.QueryRow(ctx, "SELECT pg_backend_pid()").Scan(&first); err != nil {
			return err
		}
		if err := p.QueryRow(ctx, "SELECT pg_backend_pid()").Scan(&second); err != nil {
			return err
		}
		if first != second {
			return errors.New("zero-duration pool did not reuse connection")
		}
		if _, err := p.Exec(ctx, "CREATE TABLE deferred_fixture (value int CONSTRAINT deferred_fixture_unique UNIQUE DEFERRABLE INITIALLY DEFERRED)"); err != nil {
			return err
		}
		err = db.InTx(ctx, p, pgx.TxOptions{}, func(ctx context.Context, tx pgx.Tx) error {
			_, err := tx.Exec(ctx, "INSERT INTO deferred_fixture VALUES (1), (1)")
			return err
		})
		if err == nil {
			return errors.New("deferred commit failure hidden")
		}
		var count int
		if err := p.QueryRow(ctx, "SELECT count(*) FROM deferred_fixture").Scan(&count); err != nil {
			return err
		}
		if count != 0 {
			return errors.New("failed commit persisted partial effects")
		}
		return nil
	})
}
