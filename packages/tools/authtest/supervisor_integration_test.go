//go:build integration

package authtest

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vandordev/vkit-zenith/packages/tools/dbtool"
)

// Selected only by the dedicated outer-supervisor negative test, not the
// successful fixture selector. It intentionally cannot clean up its resources.
func TestSupervisorPostgresChild(t *testing.T) {
	if os.Args[len(os.Args)-1] != "auth-supervisor-probe" {
		return // helper only; no acceptance selector includes this name.
	}
	ctx, cancel := context.WithTimeout(ownershipContext(context.Background()), time.Minute)
	defer cancel()
	err := dbtool.WithPostgres(ctx, func(ctx context.Context, _ string) error {
		listener, err := ReserveLoopback()
		if err != nil {
			return err
		}
		defer listener.Close() // will NOT run on outer timeout.
		root, err := ownedRoot("auth-process-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(root)
		file, err := listener.(*net.TCPListener).File()
		if err != nil {
			return err
		}
		defer file.Close()
		executable, err := os.Executable()
		if err != nil {
			return err
		}
		process, err := StartOwned(ctx, root, executable, []string{"-test.run=^TestFixtureOwnedChild$", "auth-owned-child"}, []*os.File{file})
		if err != nil {
			return err
		}
		defer process.Close()
		deadline := time.Now().Add(3 * time.Second)
		for {
			if _, err := os.Stat(filepath.Join(root, "ready")); err == nil {
				break
			}
			if time.Now().After(deadline) {
				return context.DeadlineExceeded
			}
			time.Sleep(10 * time.Millisecond)
		}
		if err := os.WriteFile(filepath.Join(filepath.Dir(os.Getenv("AUTH_OWNERSHIP_MANIFEST")), "ready"), nil, 0600); err != nil {
			return err
		}
		time.Sleep(time.Minute)
		return nil
	})
	if err != nil {
		t.Fatal("required owned PG acquisition failed")
	}
	t.Fatal("outer timeout injection did not run")
}
