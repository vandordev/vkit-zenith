package authtest

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/vandordev/vkit-zenith/packages/db"
	"github.com/vandordev/vkit-zenith/packages/db/migrate"
	"github.com/vandordev/vkit-zenith/packages/tools/dbtool"
)

// Fixture is test-only. The pool is reachable only inside WithFixture's callback;
// there is deliberately no operator DSN override or ambient configuration loader.
type Fixture struct {
	Pool *pgxpool.Pool
	Root string
}

func WithFixture(ctx context.Context, fn func(context.Context, *Fixture) error) error {
	return WithFixtureReceipt(ctx, &dbtool.PostgresReceipt{}, fn)
}

func WithFixtureReceipt(ctx context.Context, receipt *dbtool.PostgresReceipt, fn func(context.Context, *Fixture) error) error {
	return withFixtureDeadline(ctx, receipt, 3*time.Minute, fn)
}

// WithBrowserFixture keeps the longer journey budget out of persistence tests.
func WithBrowserFixture(ctx context.Context, fn func(context.Context, *Fixture) error) error {
	return withFixtureDeadline(ctx, &dbtool.PostgresReceipt{}, 15*time.Minute, fn)
}

func withFixtureDeadline(ctx context.Context, receipt *dbtool.PostgresReceipt, deadline time.Duration, fn func(context.Context, *Fixture) error) error {
	ctx = ownershipContext(ctx)
	ctx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()
	return dbtool.WithPostgresReceipt(ctx, receipt, func(ctx context.Context, dsn string) (err error) {
		root, err := ownedRoot("auth-fixture-")
		if err != nil {
			return errors.New("auth fixture: temporary root unavailable")
		}
		defer func() {
			if e := os.RemoveAll(root); e != nil {
				err = errors.Join(err, errors.New("auth fixture: temporary cleanup failed"))
			}
		}()
		settings := db.Settings{DSN: dsn, MaxOpenConnections: 10, ConnectionMaxLifetimeSeconds: 60, ConnectionMaxIdleTimeSeconds: 30, ConnectTimeoutSeconds: 10}
		m, err := migrate.Open(ctx, settings)
		if err != nil {
			return errors.New("auth fixture: migration connection failed")
		}
		_, source, _, ok := runtime.Caller(0)
		if !ok {
			_ = m.Close()
			return errors.New("auth fixture: canonical migration source unavailable")
		}
		_, runErr := migrate.Run(ctx, m, os.DirFS(filepath.Join(filepath.Dir(source), "../../db/migrations")), migrate.Request{Operation: migrate.Up})
		// Role defaults affect only NEW sessions, including db.Open's first ping.
		if runErr == nil {
			_, runErr = m.ExecContext(ctx, `ALTER ROLE fixture SET lock_timeout='5s'; ALTER ROLE fixture SET statement_timeout='15s'`)
		}
		closeErr := m.Close()
		if runErr != nil || closeErr != nil {
			return errors.New("auth fixture: canonical migration failed")
		}
		pool, err := db.Open(ctx, settings)
		if err != nil {
			return errors.New("auth fixture: pool open failed")
		}
		defer pool.Close()
		return fn(ctx, &Fixture{Pool: pool, Root: root})
	})
}

// ReserveLoopback keeps the kernel-assigned socket open until the owner either
// serves it directly or transfers its descriptor to a child. Never close/rebind
// a guessed free port during startup.
func ReserveLoopback() (net.Listener, error) {
	l, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	if err := recordOwnership("listener", strconv.Itoa(l.Addr().(*net.TCPAddr).Port)); err != nil {
		_ = l.Close()
		return nil, err
	}
	return l, nil
}

// ChildEnvironment is a fresh allowlist, never os.Environ. Tool paths are fixed;
// HOME points at the owned empty directory, excluding user RC/config/credentials.
func ChildEnvironment(root string) []string {
	return []string{"PATH=/usr/local/go/bin:/usr/local/bin:/usr/bin:/bin", "HOME=" + root, "TMPDIR=" + root, "LANG=C.UTF-8"}
}
