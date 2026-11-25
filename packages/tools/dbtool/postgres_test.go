package dbtool

import (
	"context"
	"errors"
	"testing"
)

func TestPostgresLifecycleCleanup(t *testing.T) {
	sentinel := errors.New("callback")
	for _, mode := range []string{"success", "callback", "cancel", "startup", "cleanup", "panic"} {
		t.Run(mode, func(t *testing.T) {
			cleaned := false
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var err error
			func() {
				defer func() {
					if p := recover(); p != nil && mode != "panic" {
						t.Fatal("unexpected panic")
					}
				}()
				err = withLifecycle(ctx, func(context.Context) (string, func(context.Context) error, error) {
					cleanup := func(c context.Context) error {
						if c.Err() != nil {
							t.Fatal("cancelled cleanup context")
						}
						cleaned = true
						if mode == "cleanup" {
							return sentinel
						}
						return nil
					}
					if mode == "startup" {
						return "", cleanup, sentinel
					}
					return "fixture-only", cleanup, nil
				}, func(context.Context, string) error {
					switch mode {
					case "callback":
						return sentinel
					case "cancel":
						cancel()
						return context.Canceled
					case "panic":
						panic(sentinel)
					}
					return nil
				})
			}()
			if !cleaned {
				t.Fatal("resource leaked")
			}
			if mode != "success" && mode != "panic" && err == nil {
				t.Fatal("failure hidden")
			}
		})
	}
}

func TestPostgresLifecyclePanicCleanupFailureSurfaced(t *testing.T) {
	sentinel := errors.New("original panic")
	defer func() {
		p, ok := recover().(lifecyclePanic)
		if !ok || p.Original != sentinel {
			t.Fatal("panic or cleanup failure hidden")
		}
	}()
	_ = withLifecycle(context.Background(), func(context.Context) (string, func(context.Context) error, error) {
		return "fixture", func(context.Context) error { return errors.New("cleanup failed") }, nil
	}, func(context.Context, string) error { panic(sentinel) })
}
