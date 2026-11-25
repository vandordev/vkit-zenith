package migrate

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

type Operation string

const (
	Status  Operation = "status"
	Version Operation = "version"
	Up      Operation = "up"
	DownTo  Operation = "down-to"
)

type Request struct {
	Operation       Operation
	Target          int64
	ConfirmRollback bool
}
type Entry struct {
	Version int64
	Applied bool
}
type Result struct {
	Version int64
	Entries []Entry
}

type boundedLocker struct{ lock.SessionLocker }

// A failed lock/unlock leaves session ownership uncertain. Never return that
// physical session to the caller's reusable pool with an advisory lock attached.
func (l boundedLocker) SessionLock(ctx context.Context, c *sql.Conn) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	err := l.SessionLocker.SessionLock(ctx, c)
	if err != nil {
		_ = c.Raw(func(any) error { return driver.ErrBadConn })
	}
	return err
}
func (l boundedLocker) SessionUnlock(_ context.Context, c *sql.Conn) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := l.SessionLocker.SessionUnlock(ctx, c)
	if err != nil {
		_ = c.Raw(func(any) error { return driver.ErrBadConn })
	}
	return err
}

// Run retains ownership of pool with the caller. Empty sources are version zero.
func Run(ctx context.Context, pool *sql.DB, source fs.FS, r Request) (Result, error) {
	if r.Operation != Status && r.Operation != Version && r.Operation != Up && r.Operation != DownTo {
		return Result{}, errors.New("migration operation: unsupported")
	}
	if r.Operation == DownTo && (!r.ConfirmRollback || r.Target < 0) {
		return Result{}, errors.New("migration down-to: nonnegative target and explicit confirmation required")
	}
	versions, err := Source(source)
	if err != nil {
		return Result{}, err
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if len(versions) == 0 {
		return Result{Entries: []Entry{}}, nil
	}
	if pool == nil {
		return Result{}, errors.New("migration operation: pool required")
	}
	l, err := lock.NewPostgresSessionLocker(lock.WithLockTimeout(1, 30), lock.WithUnlockTimeout(1, 10))
	if err != nil {
		return Result{}, errors.New("migration lock: configuration failed")
	}
	p, err := goose.NewProvider(goose.DialectPostgres, pool, source, goose.WithDisableGlobalRegistry(true), goose.WithSessionLocker(boundedLocker{l}), goose.WithSlog(slog.New(slog.NewTextHandler(io.Discard, nil))))
	safe := func(err error) error {
		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return errors.Join(errors.New("migration operation: failed"), ctx.Err())
		}
		return errors.New("migration operation: failed (source, database or advisory lock)")
	}
	if err != nil {
		return Result{}, safe(err)
	}
	switch r.Operation {
	case Up:
		_, err = p.Up(ctx)
	case DownTo:
		_, err = p.DownTo(ctx, r.Target)
	}
	if err != nil {
		return Result{}, safe(err)
	}
	v, err := p.GetDBVersion(ctx)
	if err != nil {
		return Result{}, safe(err)
	}
	result := Result{Version: v, Entries: []Entry{}}
	if r.Operation == Status {
		entries, err := p.Status(ctx)
		if err != nil {
			return Result{}, safe(err)
		}
		for _, entry := range entries {
			result.Entries = append(result.Entries, Entry{Version: entry.Source.Version, Applied: entry.State == goose.StateApplied})
		}
	}
	return result, nil
}
