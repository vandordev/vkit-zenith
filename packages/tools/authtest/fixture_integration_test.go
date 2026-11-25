//go:build integration

package authtest

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/packages/tools/dbtool"
)

func TestAuthFixturePostgres(t *testing.T) {
	for _, mode := range []string{"roundtrip", "callback-failure", "callback-cancel", "callback-panic"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			var root string
			var poolClosed func() bool
			marker := errors.New("injected callback failure")
			var caught any
			var err error
			var receipt dbtool.PostgresReceipt
			func() {
				defer func() { caught = recover() }()
				err = WithFixtureReceipt(ctx, &receipt, func(ctx context.Context, f *Fixture) error {
					root = f.Root
					poolClosed = func() bool { return f.Pool.Ping(context.Background()) != nil }
					// Hold the readiness session while acquiring a second session.
					first, err := f.Pool.Acquire(ctx)
					if err != nil {
						return errors.New("initial connection unavailable")
					}
					defer first.Release()
					second, err := f.Pool.Acquire(ctx)
					if err != nil {
						return errors.New("concurrent connection unavailable")
					}
					defer second.Release()
					for _, connection := range []*pgxpool.Conn{first, second} {
						for setting, want := range map[string]string{"lock_timeout": "5s", "statement_timeout": "15s"} {
							var got string
							if err := connection.QueryRow(ctx, "SHOW "+setting).Scan(&got); err != nil || got != want {
								return errors.New("connection timeout default missing")
							}
						}
					}
					var value int
					if err := f.Pool.QueryRow(ctx, "SELECT 42").Scan(&value); err != nil || value != 42 {
						return errors.New("owned database roundtrip failed")
					}
					// Canonical migration tables, not an ad-hoc substitute schema.
					if _, err := f.Pool.Exec(ctx, `INSERT INTO users(id,email,email_normalized,display_name,status,created_at,updated_at) VALUES(gen_random_uuid(),'fixture@example.test','fixture@example.test','Fixture','active',now(),now())`); err != nil {
						return errors.New("canonical fake identity insert failed")
					}
					if mode == "callback-failure" {
						return marker
					}
					if mode == "callback-cancel" {
						cancel()
						return ctx.Err()
					}
					if mode == "callback-panic" {
						panic(marker)
					}
					return nil
				})
			}()
			if root == "" {
				t.Fatalf("required owned PostgreSQL unavailable: %v", err)
			}
			if mode == "callback-panic" {
				if caught != marker {
					t.Fatal("callback panic or cleanup failed")
				}
			} else if caught != nil {
				t.Fatal("unexpected panic")
			}
			if mode == "roundtrip" && err != nil {
				t.Fatal(err)
			}
			if mode == "callback-failure" && err != marker {
				t.Fatal("callback error lost")
			}
			if mode == "callback-cancel" && err != context.Canceled {
				t.Fatal("cancellation lost")
			}
			if _, err := os.Stat(root); !os.IsNotExist(err) {
				t.Fatal("owned temporary root remains")
			}
			if poolClosed == nil || !poolClosed() {
				t.Fatal("owned pool remains open")
			}
			// WithPostgres returns only after Terminate + Inspect confirms removal.
			if receipt.ContainerID == "" || !receipt.Removed {
				t.Fatal("owned PostgreSQL removal not confirmed")
			}
			t.Logf("owned PG canonical roundtrip cleanup PASS container=%s removed=%v", receipt.ContainerID, receipt.Removed)
		})
	}
}

func TestAuthFixturePostgresStartupCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var receipt dbtool.PostgresReceipt
	called := false
	err := WithFixtureReceipt(ctx, &receipt, func(context.Context, *Fixture) error { called = true; return nil })
	if err == nil || called {
		t.Fatal("cancelled PG startup accepted")
	}
	if receipt.ContainerID != "" && !receipt.Removed {
		t.Fatal("partially acquired PG not removed")
	}
	t.Logf("cancelled startup cleanup PASS acquired=%v removed=%v", receipt.ContainerID != "", receipt.Removed)
}
