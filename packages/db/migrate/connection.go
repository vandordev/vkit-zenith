package migrate

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	db "github.com/vandordev/vkit-zenith/packages/db"
)

// Open is the separate operational Goose adapter. The caller closes the handle.
func Open(ctx context.Context, s db.Settings) (*sql.DB, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c, err := db.ConnectionConfig(s)
	if err != nil {
		return nil, err
	}
	p := stdlib.OpenDB(*c)
	p.SetMaxOpenConns(s.MaxOpenConnections)
	p.SetMaxIdleConns(s.MaxOpenConnections)
	p.SetConnMaxLifetime(time.Duration(s.ConnectionMaxLifetimeSeconds) * time.Second)
	p.SetConnMaxIdleTime(time.Duration(s.ConnectionMaxIdleTimeSeconds) * time.Second)
	ping, cancel := context.WithTimeout(ctx, time.Duration(s.ConnectTimeoutSeconds)*time.Second)
	defer cancel()
	if err := p.PingContext(ping); err != nil {
		safe := errors.Join(errors.New("migration database: connection verification failed"), ping.Err())
		if p.Close() != nil {
			safe = errors.Join(safe, errors.New("migration database: partial pool close failed"))
		}
		return nil, safe
	}
	return p, nil
}
