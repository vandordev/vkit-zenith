//go:build integration

package dbtool

import (
	"context"
	"errors"
	"testing"
)

func TestPostgresCancellationCleanup(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	err := WithPostgres(ctx, func(context.Context, string) error { cancel(); return context.Canceled })
	if err != context.Canceled {
		t.Fatalf("cancelled work or cleanup failed: %v", err)
	}
}

func TestPostgresFailureCleanup(t *testing.T) {
	sentinel := errors.New("fixture callback failed")
	if err := WithPostgres(t.Context(), func(context.Context, string) error { return sentinel }); err != sentinel {
		t.Fatalf("callback failure/cleanup: %v", err)
	}
	panicked := false
	func() {
		defer func() {
			if p := recover(); p != nil {
				if p != sentinel {
					t.Fatal("panic changed")
				}
				panicked = true
			}
		}()
		_ = WithPostgres(t.Context(), func(context.Context, string) error { panic(sentinel) })
	}()
	if !panicked {
		t.Fatal("panic swallowed")
	}
}
