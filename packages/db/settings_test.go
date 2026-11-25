package db

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"
	"time"
)

func validSettings() Settings {
	return Settings{DSN: "postgres://fixture:fake@127.0.0.1:1/fixture?sslmode=disable", MaxOpenConnections: 10, ConnectionMaxLifetimeSeconds: 1800, ConnectionMaxIdleTimeSeconds: 300, ConnectTimeoutSeconds: 1}
}

func TestSettingsBounds(t *testing.T) {
	if err := validSettings().Validate(); err != nil {
		t.Fatal(err)
	}
	for _, change := range []func(*Settings){
		func(s *Settings) { s.DSN = "" }, func(s *Settings) { s.MaxOpenConnections = 0 }, func(s *Settings) { s.MaxOpenConnections = 10001 },
		func(s *Settings) { s.ConnectionMaxLifetimeSeconds = 604801 }, func(s *Settings) { s.ConnectionMaxIdleTimeSeconds = -1 },
		func(s *Settings) { s.ConnectTimeoutSeconds = 0 }, func(s *Settings) { s.ConnectTimeoutSeconds = 301 },
	} {
		s := validSettings()
		change(&s)
		if s.Validate() == nil {
			t.Fatal("invalid settings accepted")
		}
	}
}

func TestConnectionExplicitParsing(t *testing.T) {
	s := validSettings()
	c, err := ConnectionConfig(s)
	if err != nil {
		t.Fatal(err)
	}
	if c.User != "fixture" || c.Password != "fake" || c.Host != "127.0.0.1" || c.Database != "fixture" || c.TLSConfig != nil {
		t.Fatal("explicit config lost")
	}
	for _, key := range []string{"PGHOST", "PGPASSWORD", "PGSERVICE", "PGPASSFILE", "PGSSLMODE"} {
		t.Run(key, func(t *testing.T) {
			t.Setenv(key, "sentinel-secret")
			_, err := ConnectionConfig(s)
			if err == nil || strings.Contains(err.Error(), "sentinel-secret") {
				t.Fatal("ambient config not safely refused")
			}
		})
	}
	for _, dsn := range []string{"sentinel-secret", "postgres://u:p@remote/db", "postgres://u:p@remote/db?sslmode=prefer", "postgres://u:p@remote/db?sslmode=allow", "postgres://u:p@localhost/db?service=sentinel-secret", "postgres://u:p@localhost/db?passfile=sentinel-secret"} {
		s.DSN = dsn
		_, err := ConnectionConfig(s)
		if err == nil || strings.Contains(err.Error(), "sentinel-secret") {
			t.Fatal("unsafe DSN accepted/leaked")
		}
	}
}

func TestConnectionExplicitInternalTransport(t *testing.T) {
	s := validSettings()
	s.DSN = "postgres://fixture:fake@postgres:5432/fixture?sslmode=disable"
	c, err := ConnectionConfig(s)
	if err != nil {
		t.Fatal(err)
	}
	if c.Host != "postgres" || c.TLSConfig != nil || len(c.Fallbacks) != 0 {
		t.Fatal("explicit internal plaintext transport not preserved")
	}
}

func TestOpenCanceledSafe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Open(ctx, validSettings())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
}

func TestPoolConfigurationMapping(t *testing.T) {
	for _, zero := range []bool{false, true} {
		s := validSettings()
		if zero {
			s.ConnectionMaxLifetimeSeconds, s.ConnectionMaxIdleTimeSeconds = 0, 0
		}
		c, err := poolConfig(s)
		if err != nil {
			t.Fatal(err)
		}
		if c.MaxConns != 10 || c.MinConns != 0 || c.MinIdleConns != 0 || c.MaxConnLifetimeJitter != 0 {
			t.Fatal("pool bounds changed")
		}
		idle := time.Duration(s.ConnectionMaxIdleTimeSeconds) * time.Second
		if zero {
			idle = time.Duration(math.MaxInt64)
		}
		if c.MaxConnLifetime != time.Duration(s.ConnectionMaxLifetimeSeconds)*time.Second || c.MaxConnIdleTime != idle {
			t.Fatal("duration contract changed")
		}
	}
}
