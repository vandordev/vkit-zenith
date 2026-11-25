package db

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// InTx commits only successful callbacks; no retries, nesting or savepoints.
func InTx(ctx context.Context, pool *pgxpool.Pool, opts pgx.TxOptions, fn func(context.Context, pgx.Tx) error) error {
	return inTx(ctx, pool.BeginTx, opts, fn)
}

func inTx(ctx context.Context, begin func(context.Context, pgx.TxOptions) (pgx.Tx, error), opts pgx.TxOptions, fn func(context.Context, pgx.Tx) error) (err error) {
	tx, e := begin(ctx, opts)
	if e != nil {
		return errors.Join(errors.New("database transaction: begin failed"), ctx.Err())
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		// Defer does not recover: original panic identity survives cleanup.
		if e := tx.Rollback(cleanup); e != nil && !errors.Is(e, pgx.ErrTxClosed) {
			err = errors.Join(err, errors.New("database transaction: rollback failed"))
		}
	}()
	if err = fn(ctx, tx); err != nil {
		return err
	}
	if e := tx.Commit(ctx); e != nil {
		return errors.Join(errors.New("database transaction: commit failed"), ctx.Err())
	}
	return nil
}
