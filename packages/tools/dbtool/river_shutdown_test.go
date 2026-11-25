package dbtool

import (
	"context"
	"errors"
	"testing"
	"time"
)

type fakeRiverStopper struct {
	stopped    chan struct{}
	hardCtx    error
	stopErr    error
	stopOnHard bool
}

func (c *fakeRiverStopper) Stop(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return c.stopErr
}

func (c *fakeRiverStopper) StopAndCancel(ctx context.Context) error {
	c.hardCtx = ctx.Err()
	if c.hardCtx == nil && c.stopOnHard {
		close(c.stopped)
	}
	return c.hardCtx
}

func (c *fakeRiverStopper) Stopped() <-chan struct{} { return c.stopped }

func TestStopRiverClientAfterWorkContextExpires(t *testing.T) {
	workCtx, cancel := context.WithCancel(context.Background())
	cancel()
	if !errors.Is(workCtx.Err(), context.Canceled) {
		t.Fatal("work context did not expire")
	}

	gracefulErr := errors.New("graceful shutdown timed out")
	client := &fakeRiverStopper{stopped: make(chan struct{}), stopErr: gracefulErr, stopOnHard: true}
	cleaned := false
	err := withLifecycle(workCtx, func(context.Context) (string, func(context.Context) error, error) {
		return "fixture", func(context.Context) error {
			select {
			case <-client.Stopped():
				cleaned = true
				return nil
			default:
				return errors.New("River active during PostgreSQL cleanup")
			}
		}, nil
	}, func(ctx context.Context, _ string) error {
		if !errors.Is(ctx.Err(), context.Canceled) {
			return errors.New("work context was not expired")
		}
		return stopRiverClient(client)
	})
	if !errors.Is(err, gracefulErr) {
		t.Fatalf("graceful stop failure was lost: %v", err)
	}
	if !cleaned {
		t.Fatal("PostgreSQL cleanup did not follow confirmed River termination")
	}
	if client.hardCtx != nil {
		t.Fatalf("hard stop inherited canceled context: %v", client.hardCtx)
	}
	select {
	case <-client.Stopped():
	default:
		t.Fatal("River termination was not confirmed")
	}
}

func TestUnconfirmedRiverShutdownRetainsPostgres(t *testing.T) {
	gracefulErr := errors.New("graceful shutdown timed out")
	client := &fakeRiverStopper{stopped: make(chan struct{}), stopErr: gracefulErr}
	cleaned := false
	err := withLifecycle(context.Background(), func(context.Context) (string, func(context.Context) error, error) {
		return "fixture", func(context.Context) error { cleaned = true; return nil }, nil
	}, func(context.Context, string) error {
		return stopRiverClientWithin(client, time.Millisecond, time.Millisecond)
	})
	if cleaned {
		t.Fatal("owned PostgreSQL was removed while River may still be active")
	}
	if !errors.Is(err, errUnsafeOwnedCleanup) || !errors.Is(err, gracefulErr) || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("shutdown/cleanup failure was not preserved: %v", err)
	}
}

func TestPostgresLifecycleRetainsResourceWhenClientIsActive(t *testing.T) {
	cleaned := false
	activeErr := errors.New("River still active")
	err := withLifecycle(context.Background(), func(context.Context) (string, func(context.Context) error, error) {
		return "fixture", func(context.Context) error {
			cleaned = true
			return nil
		}, nil
	}, func(context.Context, string) error {
		return errors.Join(errUnsafeOwnedCleanup, activeErr)
	})
	if cleaned {
		t.Fatal("owned PostgreSQL was removed while River may still be active")
	}
	if !errors.Is(err, errUnsafeOwnedCleanup) || !errors.Is(err, activeErr) {
		t.Fatalf("unsafe cleanup failure was not surfaced: %v", err)
	}
}
