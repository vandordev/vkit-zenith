//go:build integration

package dbtool

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"testing/fstest"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3/lock"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	db "github.com/vandordev/vkit-zenith/packages/db"
	"github.com/vandordev/vkit-zenith/packages/db/migrate"
)

func TestSchemaMigrationClosure(t *testing.T) {
	latest := latestSchemaVersion(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	err := WithPostgres(ctx, func(ctx context.Context, dsn string) error {
		m, err := migrate.Open(ctx, settings(dsn))
		if err != nil {
			return err
		}
		defer m.Close()
		p, err := db.Open(ctx, settings(dsn))
		if err != nil {
			return err
		}
		defer p.Close()
		source := os.DirFS("../../db/migrations")
		for v := 1; v <= latest; v++ {
			result, err := migrate.Run(ctx, m, prefixSource{source, v}, migrate.Request{Operation: migrate.Up})
			if err != nil {
				return fmt.Errorf("up %d: %w", v, err)
			}
			if result.Version != int64(v) {
				return errors.New("prefix version")
			}
		}
		for _, op := range []migrate.Operation{migrate.Up, migrate.Status, migrate.Version} {
			r, err := migrate.Run(ctx, m, source, migrate.Request{Operation: op})
			if err != nil {
				return err
			}
			if r.Version != int64(latest) {
				return errors.New("final version")
			}
			if op == migrate.Status {
				if len(r.Entries) != latest {
					return errors.New("status entries")
				}
				for _, e := range r.Entries {
					if !e.Applied {
						return errors.New("unapplied source")
					}
				}
			}
		}
		// The baseline supports a complete owned fixture rollback.
		if _, err := migrate.Run(ctx, m, source, migrate.Request{Operation: migrate.DownTo, Target: 0, ConfirmRollback: true}); err != nil {
			return err
		}
		var n int
		if err := p.QueryRow(ctx, `SELECT count(*) FROM pg_tables WHERE schemaname='public' AND tablename<>'goose_db_version'`).Scan(&n); err != nil {
			return err
		}
		if n != 0 {
			return errors.New("Down left owned tables")
		}
		if err := p.QueryRow(ctx, `SELECT count(*) FROM pg_proc p JOIN pg_namespace n ON n.oid=p.pronamespace WHERE n.nspname='public' AND NOT EXISTS(SELECT 1 FROM pg_depend d WHERE d.objid=p.oid AND d.deptype='e')`).Scan(&n); err != nil {
			return err
		}
		if n != 0 {
			return errors.New("Down left owned functions")
		}
		if err := p.QueryRow(ctx, `SELECT count(*) FROM pg_extension WHERE extname='btree_gist'`).Scan(&n); err != nil {
			return err
		}
		if n != 0 {
			return errors.New("baseline installed unneeded extension")
		}
		if _, err := migrate.Run(ctx, m, source, migrate.Request{Operation: migrate.Up}); err != nil {
			return err
		}
		manifest, err := loadManifest()
		if err != nil {
			return err
		}
		var tables []string
		for _, g := range manifest.Groups {
			for table := range g.Tables {
				tables = append(tables, table)
			}
		}
		if err := checkTables(ctx, p, tables); err != nil {
			return err
		}
		// A malformed migration transaction leaves neither side effects nor version.
		bad := fstest.MapFS{fmt.Sprintf("%05d_bad.sql", latest+1): {Data: []byte("-- +goose Up\nCREATE TABLE partial_fixture(id integer);\nTHIS IS NOT SQL;\n-- +goose Down\nDROP TABLE partial_fixture;\n")}}
		if _, err := migrate.Run(ctx, m, bad, migrate.Request{Operation: migrate.Up}); err == nil {
			return errors.New("malformed source accepted")
		}
		var exists bool
		if err := p.QueryRow(ctx, `SELECT to_regclass('public.partial_fixture') IS NOT NULL OR EXISTS(SELECT 1 FROM goose_db_version WHERE version_id=$1 AND is_applied)`, latest+1).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return errors.New("partial malformed migration persisted")
		}
		duplicate := fstest.MapFS{fmt.Sprintf("%05d_a.sql", latest+1): {Data: []byte("-- +goose Up\nSELECT 1;")}, fmt.Sprintf("%05d_b.sql", latest+1): {Data: []byte("-- +goose Up\nSELECT 1;")}}
		if _, err := migrate.Run(ctx, m, duplicate, migrate.Request{Operation: migrate.Up}); err == nil {
			return errors.New("duplicate version accepted")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

type schemaRiverArgs struct{}

func (schemaRiverArgs) Kind() string { return "schema_test" }

type schemaRiverWorker struct {
	river.WorkerDefaults[schemaRiverArgs]
	done chan struct{}
}

func (w *schemaRiverWorker) Work(context.Context, *river.Job[schemaRiverArgs]) error {
	w.done <- struct{}{}
	return nil
}

func TestSchemaRiver(t *testing.T) {
	schemaDatabase(t, func(ctx context.Context, p *pgxpool.Pool) (result error) {
		worked := make(chan struct{}, 1)
		workers := river.NewWorkers()
		river.AddWorker(workers, &schemaRiverWorker{done: worked})
		client, err := river.NewClient(riverpgxv5.New(p), &river.Config{
			Queues:  map[string]river.QueueConfig{river.QueueDefault: {MaxWorkers: 1}},
			Workers: workers,
		})
		if err != nil {
			return err
		}
		if err := client.Start(ctx); err != nil {
			return err
		}
		defer func() {
			result = errors.Join(result, stopRiverClient(client))
		}()

		tx, err := p.Begin(ctx)
		if err != nil {
			return err
		}
		inserted, err := client.InsertTx(ctx, tx, schemaRiverArgs{}, nil)
		if err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}

		select {
		case <-worked:
		case <-ctx.Done():
			return fmt.Errorf("River test job was not worked: %w", ctx.Err())
		}
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			var state string
			if err := p.QueryRow(ctx, `SELECT state::text FROM river_job WHERE id=$1`, inserted.Job.ID).Scan(&state); err != nil {
				return err
			}
			if state == "completed" {
				break
			}
			select {
			case <-ticker.C:
			case <-ctx.Done():
				return fmt.Errorf("River test job state is %q: %w", state, ctx.Err())
			}
		}
		return nil
	})
}

func TestSchemaMigrationLock(t *testing.T) {
	latest := latestSchemaVersion(t)
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
		c, err := a.Conn(ctx)
		if err != nil {
			return err
		}
		defer c.Close()
		locker, err := lock.NewPostgresSessionLocker()
		if err != nil {
			return err
		}
		if err = locker.SessionLock(ctx, c); err != nil {
			return err
		}
		source := os.DirFS("../../db/migrations")
		short, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
		_, err = migrate.Run(short, b, source, migrate.Request{Operation: migrate.Up})
		cancel()
		if !errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("canonical lock timeout: %w", err)
		}
		var exists bool
		if err := a.QueryRowContext(ctx, `SELECT to_regclass('public.users') IS NOT NULL`).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return errors.New("locked runner applied canonical source")
		}
		if err := locker.SessionUnlock(ctx, c); err != nil {
			return err
		}
		var wg sync.WaitGroup
		done := make(chan error, 2)
		for _, pool := range []*sql.DB{a, b} {
			wg.Add(1)
			go func(pool *sql.DB) {
				defer wg.Done()
				r, e := migrate.Run(ctx, pool, source, migrate.Request{Operation: migrate.Up})
				if e == nil && r.Version != int64(latest) {
					e = errors.New("concurrent canonical version")
				}
				done <- e
			}(pool)
		}
		wg.Wait()
		close(done)
		for err := range done {
			if err != nil {
				return err
			}
		}
		var locked bool
		if err := a.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM pg_locks WHERE locktype='advisory' AND granted AND pid IN (SELECT pid FROM pg_stat_activity WHERE datname=current_database()))`).Scan(&locked); err != nil {
			return err
		}
		if locked {
			return errors.New("runner retained advisory lock")
		}
		return nil
	})
}
