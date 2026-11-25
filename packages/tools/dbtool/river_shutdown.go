package dbtool

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type riverStopper interface {
	Stop(context.Context) error
	StopAndCancel(context.Context) error
	Stopped() <-chan struct{}
}

func stopRiverClient(client riverStopper) error {
	return stopRiverClientWithin(client, 10*time.Second, 30*time.Second)
}

func stopRiverClientWithin(client riverStopper, gracefulTimeout, hardTimeout time.Duration) error {
	stopCtx, cancel := context.WithTimeout(context.Background(), gracefulTimeout)
	stopErr := client.Stop(stopCtx)
	cancel()
	select {
	case <-client.Stopped():
		return stopErr
	default:
	}

	hardCtx, cancel := context.WithTimeout(context.Background(), hardTimeout)
	defer cancel()
	hardErr := client.StopAndCancel(hardCtx)
	select {
	case <-client.Stopped():
		return errors.Join(stopErr, hardErr)
	case <-hardCtx.Done():
		return errors.Join(stopErr, hardErr, errUnsafeOwnedCleanup, fmt.Errorf("River shutdown was not confirmed before PostgreSQL cleanup: %w", hardCtx.Err()))
	}
}
