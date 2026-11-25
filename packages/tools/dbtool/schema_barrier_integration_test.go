//go:build integration

package dbtool

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

func waitBlocked(ctx context.Context, p *pgxpool.Pool, pid uint32) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tick := time.NewTicker(5 * time.Millisecond)
	defer tick.Stop()
	for {
		var blocked bool
		if err := p.QueryRow(ctx, "SELECT cardinality(pg_blocking_pids($1))>0", pid).Scan(&blocked); err != nil {
			return err
		}
		if blocked {
			return nil
		}
		select {
		case <-ctx.Done():
			return errors.New("race did not reach lock barrier")
		case <-tick.C:
		}
	}
}
