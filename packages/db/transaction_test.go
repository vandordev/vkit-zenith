package db

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

type fakeTx struct {
	pgx.Tx
	commitErr, rollbackErr error
	committed, rolled      bool
}

func (t *fakeTx) Commit(context.Context) error { t.committed = true; return t.commitErr }
func (t *fakeTx) Rollback(ctx context.Context) error {
	if ctx.Err() != nil {
		panic("cleanup inherited cancellation")
	}
	if _, ok := ctx.Deadline(); !ok {
		panic("unbounded cleanup")
	}
	t.rolled = true
	return t.rollbackErr
}

func TestTransactionOutcomes(t *testing.T) {
	sentinel := errors.New("callback failed")
	for _, mode := range []string{"success", "callback", "rollback", "commit", "panic", "cancel", "begin"} {
		t.Run(mode, func(t *testing.T) {
			tx := &fakeTx{}
			if mode == "rollback" {
				tx.rollbackErr = sentinel
			}
			if mode == "commit" {
				tx.commitErr = sentinel
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			panicked := false
			var err error
			func() {
				defer func() {
					if p := recover(); p != nil {
						if p != sentinel {
							t.Fatal("panic identity lost")
						}
						panicked = true
					}
				}()
				err = inTx(ctx, func(context.Context, pgx.TxOptions) (pgx.Tx, error) {
					if mode == "begin" {
						return nil, sentinel
					}
					return tx, nil
				}, pgx.TxOptions{}, func(context.Context, pgx.Tx) error {
					switch mode {
					case "callback", "rollback":
						return sentinel
					case "panic":
						panic(sentinel)
					case "cancel":
						cancel()
						return ctx.Err()
					}
					return nil
				})
			}()
			if mode != "begin" && !tx.rolled {
				t.Fatal("cleanup not attempted")
			}
			switch mode {
			case "success":
				if err != nil || !tx.committed {
					t.Fatal("commit failed")
				}
			case "callback", "rollback":
				if !errors.Is(err, sentinel) || tx.committed {
					t.Fatal("callback identity lost")
				}
			case "commit", "begin":
				if err == nil || errors.Is(err, sentinel) {
					t.Fatal("unsafe boundary error")
				}
			case "panic":
				if !panicked || tx.committed {
					t.Fatal("panic swallowed")
				}
			case "cancel":
				if !errors.Is(err, context.Canceled) || tx.committed {
					t.Fatal("cancellation lost")
				}
			}
		})
	}
}
