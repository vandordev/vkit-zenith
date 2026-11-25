//go:build integration

package dbtool

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/vandordev/vkit-zenith/packages/db"
	"github.com/vandordev/vkit-zenith/packages/db/migrate"
)

type prefixSource struct {
	fs.FS
	version int
}

func (p prefixSource) ReadDir(name string) ([]fs.DirEntry, error) {
	entries, err := fs.ReadDir(p.FS, name)
	if err != nil {
		return nil, err
	}
	var out []fs.DirEntry
	for _, e := range entries {
		n, _, _ := strings.Cut(e.Name(), "_")
		v, err := strconv.Atoi(n)
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") || err != nil || v <= p.version {
			out = append(out, e)
		}
	}
	return out, nil
}
func (p prefixSource) Open(name string) (fs.File, error) {
	n, _, _ := strings.Cut(path.Base(name), "_")
	v, err := strconv.Atoi(n)
	if strings.HasSuffix(name, ".sql") && (err != nil || v > p.version) {
		return nil, fs.ErrNotExist
	}
	return p.FS.Open(name)
}

func schemaDatabase(t *testing.T, fn func(context.Context, *pgxpool.Pool) error) {
	schemaPrefixDatabase(t, latestSchemaVersion(t), fn)
}

func latestSchemaVersion(t *testing.T) int {
	t.Helper()
	entries, err := os.ReadDir("../../db/migrations")
	if err != nil {
		t.Fatal(err)
	}
	latest := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		n, _, _ := strings.Cut(entry.Name(), "_")
		v, err := strconv.Atoi(n)
		if err != nil {
			t.Fatalf("invalid canonical migration: %s", entry.Name())
		}
		if v > latest {
			latest = v
		}
	}
	if latest == 0 {
		t.Fatal("no canonical migrations")
	}
	return latest
}

// Retain intermediate acceptance and separately exercise the same contract on
// the final migration chain. A prefix pass is never substituted for a full pass.
func schemaContractDatabase(t *testing.T, prefix int, fn func(context.Context, *pgxpool.Pool) error) {
	t.Helper()
	versions := []int{prefix}
	if latest := latestSchemaVersion(t); latest != prefix {
		versions = append(versions, latest)
	}
	for _, version := range versions {
		t.Run(fmt.Sprintf("migration-%d", version), func(t *testing.T) { schemaPrefixDatabase(t, version, fn) })
	}
}

func checkPrefixOrFullTables(ctx context.Context, p *pgxpool.Pool, prefix []string) error {
	var version int
	if err := p.QueryRow(ctx, `SELECT max(version_id) FROM goose_db_version WHERE is_applied`).Scan(&version); err != nil {
		return err
	}
	manifest, err := loadManifest()
	if err != nil {
		return err
	}
	var tables []string
	for _, g := range manifest.Groups {
		for table := range g.Tables {
			if manifest.TableVersions[table] <= version {
				tables = append(tables, table)
			}
		}
	}
	return checkTables(ctx, p, tables)
}
func schemaPrefixDatabase(t *testing.T, version int, fn func(context.Context, *pgxpool.Pool) error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	ctx = WithOwnershipJournal(ctx, os.Getenv("AUTH_OWNERSHIP_MANIFEST"))
	err := WithPostgres(ctx, func(ctx context.Context, dsn string) error {
		work, cancel := context.WithTimeout(ctx, 3*time.Minute)
		defer cancel()
		m, err := migrate.Open(work, settings(dsn))
		if err != nil {
			return err
		}
		source := prefixSource{os.DirFS("../../db/migrations"), version}
		_, runErr := migrate.Run(work, m, source, migrate.Request{Operation: migrate.Up})
		if runErr == nil {
			_, runErr = migrate.Run(work, m, source, migrate.Request{Operation: migrate.Up})
		}
		if runErr == nil {
			_, runErr = m.ExecContext(work, "ALTER ROLE fixture SET lock_timeout='5s'; ALTER ROLE fixture SET statement_timeout='15s'")
		}
		closeErr := m.Close()
		if runErr != nil || closeErr != nil {
			return errors.Join(runErr, closeErr)
		}
		p, err := db.Open(work, settings(dsn))
		if err != nil {
			return err
		}
		closePool := true
		defer func() {
			if closePool {
				p.Close()
			}
		}()
		fnErr := fn(work, p)
		if errors.Is(fnErr, errUnsafeOwnedCleanup) {
			closePool = false // Keep River's database handle alive while Ryuk owns cleanup.
		}
		return fnErr
	})
	if err != nil {
		t.Fatalf("schema fixture: %v", err)
	}
}

func schemaTx(ctx context.Context, p *pgxpool.Pool, fn func(pgx.Tx) error) error {
	return db.InTx(ctx, p, pgx.TxOptions{}, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, "SET LOCAL lock_timeout='5s'; SET LOCAL statement_timeout='15s'"); err != nil {
			return err
		}
		return fn(tx)
	})
}

func expectPG(err error, code, constraint string) error {
	var pe *pgconn.PgError
	if !errors.As(err, &pe) || pe.Code != code || (constraint != "" && pe.ConstraintName != constraint) {
		return fmt.Errorf("unexpected SQLSTATE/constraint: want %s/%s, got %v", code, constraint, err)
	}
	return nil
}

// withConnections always releases both handles after independent rollback,
// including when a barrier worker fails. Callers own explicit channels/barriers.
func withConnections(ctx context.Context, p *pgxpool.Pool, fn func(*pgxpool.Conn, *pgxpool.Conn) error) (err error) {
	a, err := p.Acquire(ctx)
	if err != nil {
		return err
	}
	defer a.Release()
	b, err := p.Acquire(ctx)
	if err != nil {
		return err
	}
	defer b.Release()
	defer func() {
		c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, e := a.Exec(c, "ROLLBACK"); e != nil {
			err = errors.Join(err, errors.New("first fixture connection rollback failed"))
		}
		if _, e := b.Exec(c, "ROLLBACK"); e != nil {
			err = errors.Join(err, errors.New("second fixture connection rollback failed"))
		}
	}()
	for _, conn := range []*pgxpool.Conn{a, b} {
		if _, err := conn.Exec(ctx, "SET lock_timeout='5s'; SET statement_timeout='15s'"); err != nil {
			return err
		}
	}
	return fn(a, b)
}
