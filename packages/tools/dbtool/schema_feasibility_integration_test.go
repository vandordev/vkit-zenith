//go:build integration

package dbtool

import (
	"context"
	"errors"
	"testing"

	db "github.com/vandordev/vkit-zenith/packages/db"
)

func TestSchemaFeasibility(t *testing.T) {
	isolated(t, func(ctx context.Context, dsn string) error {
		p, err := db.Open(ctx, settings(dsn))
		if err != nil {
			return err
		}
		defer p.Close()
		conn, err := p.Acquire(ctx)
		if err != nil {
			return err
		}
		defer conn.Release()
		_, err = conn.Exec(ctx, `CREATE ROLE extension_fixture NOSUPERUSER NOLOGIN;
GRANT CREATE ON DATABASE fixture TO extension_fixture;
GRANT USAGE, CREATE ON SCHEMA public TO extension_fixture;
SET ROLE extension_fixture;
SET lock_timeout='5s'; SET statement_timeout='15s';
CREATE EXTENSION btree_gist;
CREATE TABLE exclusion_fixture (id uuid, starts_at timestamptz, ends_at timestamptz,
CONSTRAINT exclusion_fixture_no_overlap EXCLUDE USING gist (id WITH =, tstzrange(starts_at,ends_at,'[)') WITH &&));
INSERT INTO exclusion_fixture VALUES ('00000000-0000-0000-0000-000000000001','2026-01-01','2026-02-01'),
('00000000-0000-0000-0000-000000000001','2026-02-01','2026-03-01');`)
		if err != nil {
			return err
		}
		var super bool
		if err := conn.QueryRow(ctx, "SELECT rolsuper FROM pg_roles WHERE rolname=current_user").Scan(&super); err != nil {
			return err
		}
		if super {
			return errors.New("extension probe remained superuser")
		}
		_, err = conn.Exec(ctx, "INSERT INTO exclusion_fixture VALUES ('00000000-0000-0000-0000-000000000001','2026-01-15','2026-02-15')")
		return expectPG(err, "23P01", "exclusion_fixture_no_overlap")
	})
}
