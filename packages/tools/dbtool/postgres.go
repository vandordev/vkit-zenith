// Package dbtool owns isolated PostgreSQL/codegen resources, never runtime pools.
package dbtool

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/containerd/errdefs"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const PostgresImage = "postgres:17.11-bookworm@sha256:858bb5e98eed02eb8cafbcdc007f9a65d1b613972028944a8ec2133ccaf80d27"

type starter func(context.Context) (string, func(context.Context) error, error)

var errUnsafeOwnedCleanup = errors.New("isolated PostgreSQL: owned resource is still in use")

type lifecyclePanic struct{ Original any }

func (lifecyclePanic) Error() string {
	return "isolated PostgreSQL: panic with owned resource cleanup failure"
}

func withLifecycle(ctx context.Context, start starter, fn func(context.Context, string) error) (err error) {
	dsn, cleanup, startErr := start(ctx)
	if cleanup != nil {
		defer func() {
			panicked := recover()
			if errors.Is(err, errUnsafeOwnedCleanup) {
				if panicked != nil {
					panic(lifecyclePanic{Original: panicked})
				}
				err = errors.Join(err, errors.New("isolated PostgreSQL: resource retained for the enabled reaper"))
				return
			}
			c, cancel := context.WithTimeout(context.Background(), 45*time.Second)
			defer cancel()
			if e := cleanup(c); e != nil {
				if panicked != nil {
					panic(lifecyclePanic{Original: panicked})
				}
				err = errors.Join(err, errors.New("isolated PostgreSQL: owned resource cleanup failed"))
			}
			if panicked != nil {
				panic(panicked)
			}
		}()
	}
	if startErr != nil {
		return errors.Join(errors.New("isolated PostgreSQL: startup/readiness failed"), ctx.Err())
	}
	return fn(ctx, dsn)
}

// WithPostgres has no DSN override: each invocation owns a fresh random-port
// container. Cleanup ignores cancellation of work. Resource reaper stays enabled.
func WithPostgres(ctx context.Context, fn func(context.Context, string) error) error {
	return withPostgresReceipt(ctx, nil, fn)
}

// PostgresReceipt contains safe owned-resource identity, never credentials/DSN.
// Removed is set only after the owned container's absence is confirmed.
type PostgresReceipt struct {
	ContainerID string
	Removed     bool
}

func WithPostgresReceipt(ctx context.Context, receipt *PostgresReceipt, fn func(context.Context, string) error) error {
	if receipt == nil {
		return errors.New("isolated PostgreSQL: receipt required")
	}
	*receipt = PostgresReceipt{}
	return withPostgresReceipt(ctx, receipt, fn)
}

func withPostgresReceipt(ctx context.Context, receipt *PostgresReceipt, fn func(context.Context, string) error) error {
	if testcontainers.ReadConfig().RyukDisabled {
		return errors.New("isolated PostgreSQL: resource reaper must remain enabled")
	}
	// Intent is durable BEFORE Docker acquisition, including Ryuk and any
	// PostgreSQL created before Run returns a handle. The outer owner can inspect
	// this exact session's labels after a hard Go timeout, never a guessed ID.
	if err := RecordOwnership(ctx, "postgres-session", testcontainers.SessionID()); err != nil {
		return err
	}
	return withLifecycle(ctx, func(c context.Context) (string, func(context.Context) error, error) {
		secret := make([]byte, 24)
		if _, err := rand.Read(secret); err != nil {
			return "", nil, err
		}
		startup, cancel := context.WithTimeout(c, 2*time.Minute)
		defer cancel()
		container, err := postgres.Run(startup, PostgresImage, postgres.WithDatabase("fixture"), postgres.WithUsername("fixture"), postgres.WithPassword(hex.EncodeToString(secret)), postgres.BasicWaitStrategies(), testcontainers.WithLogger(discardLogger{}))
		returnedErr := RecordOwnership(ctx, "postgres-returned", testcontainers.SessionID())
		err = errors.Join(err, returnedErr)
		var cleanup func(context.Context) error
		if container != nil {
			// Recording failure must not bypass cleanup of the acquired handle.
			journalErr := RecordOwnership(ctx, "postgres-acquired", container.GetContainerID())
			err = errors.Join(err, journalErr)
			if receipt != nil {
				receipt.ContainerID = container.GetContainerID()
			}
			cleanup = func(c context.Context) error {
				if err := container.Terminate(c); err != nil {
					return err
				}
				if _, err := container.Inspect(c); !errdefs.IsNotFound(err) {
					return errors.New("isolated PostgreSQL: removal could not be confirmed")
				}
				if receipt != nil {
					receipt.Removed = true
				}
				return nil
			}
		}
		if err != nil {
			return "", cleanup, err
		}
		dsn, err := container.ConnectionString(c, "sslmode=disable")
		return dsn, cleanup, err
	}, fn)
}

type discardLogger struct{}

func (discardLogger) Printf(string, ...any) {}
