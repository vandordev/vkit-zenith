package db

import (
	"context"
	"errors"
	"math"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// connectionConfig deliberately accepts explicit PostgreSQL URLs only. pgx has no
// environment-isolated parser; fail closed on ambient PG settings rather than
// mutating process environment or discovering service/password files.
func ConnectionConfig(s Settings) (*pgx.ConnConfig, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}
	for _, entry := range os.Environ() {
		if strings.HasPrefix(entry, "PG") {
			_, value, _ := strings.Cut(entry, "=")
			if value != "" {
				return nil, errors.New("database configuration: ambient PG settings are forbidden; use explicit dsn")
			}
		}
	}
	u, err := url.Parse(s.DSN)
	invalid := func() (*pgx.ConnConfig, error) {
		return nil, errors.New("database dsn: invalid explicit PostgreSQL URL or TLS settings")
	}
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.Hostname() == "" || u.User == nil || u.User.Username() == "" || len(u.Path) < 2 || u.Fragment != "" || u.Opaque != "" {
		return invalid()
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return invalid()
	}
	for key, values := range q {
		if len(values) != 1 {
			return invalid()
		}
		switch key {
		case "sslmode", "sslrootcert", "sslcert", "sslkey", "application_name":
		default:
			return invalid()
		}
	}
	mode := q.Get("sslmode")
	// Plaintext must be an explicit operator choice, including private Docker
	// networks. Never infer it from the host or silently fall back from TLS.
	if mode != "verify-full" && mode != "verify-ca" && mode != "require" && mode != "disable" {
		return invalid()
	}
	// Empty explicit values suppress pgx's home-directory defaults. These options
	// are adapter-owned and cannot be supplied in the URL query.
	q.Set("passfile", "")
	for _, key := range []string{"sslcert", "sslkey", "sslrootcert"} {
		if !q.Has(key) {
			q.Set(key, "")
		}
	}
	q.Set("connect_timeout", strconv.Itoa(s.ConnectTimeoutSeconds))
	u.RawQuery = q.Encode()
	c, err := pgx.ParseConfig(u.String())
	if err != nil {
		return invalid()
	}
	return c, nil
}

// Open creates and pings an independent native pool. The caller closes it.
// It never reads YAML, runs migrations or starts containers.
func Open(ctx context.Context, s Settings) (*pgxpool.Pool, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c, err := poolConfig(s)
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.NewWithConfig(ctx, c)
	if err != nil {
		return nil, errors.New("database open: pool initialization failed")
	}
	pingCtx, cancel := context.WithTimeout(ctx, time.Duration(s.ConnectTimeoutSeconds)*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		safe := errors.New("database open: connection verification failed")
		if pingCtx.Err() != nil {
			safe = errors.Join(safe, pingCtx.Err())
		}
		return nil, safe
	}
	return pool, nil
}

func poolConfig(s Settings) (*pgxpool.Config, error) {
	c, err := ConnectionConfig(s)
	if err != nil {
		return nil, err
	}
	// Parse only after isolation validation; use the validated config itself.
	p, err := pgxpool.ParseConfig(c.ConnString())
	if err != nil {
		return nil, errors.New("database open: invalid pool configuration")
	}
	p.ConnConfig = c
	p.MaxConns = int32(s.MaxOpenConnections)
	p.MinConns, p.MinIdleConns = 0, 0
	p.MaxConnLifetime = time.Duration(s.ConnectionMaxLifetimeSeconds) * time.Second
	p.MaxConnLifetimeJitter = 0
	p.MaxConnIdleTime = time.Duration(s.ConnectionMaxIdleTimeSeconds) * time.Second
	// pgx treats zero idle time as immediate retirement, unlike our unlimited
	// contract. The maximum positive duration disables practical retirement.
	if p.MaxConnIdleTime == 0 {
		p.MaxConnIdleTime = time.Duration(math.MaxInt64)
	}
	return p, nil
}
