package db

import (
	"errors"
	"strings"
)

type Settings struct {
	DSN                          string
	MaxOpenConnections           int
	ConnectionMaxLifetimeSeconds int
	ConnectionMaxIdleTimeSeconds int
	ConnectTimeoutSeconds        int
}

func (s Settings) Validate() error {
	if strings.TrimSpace(s.DSN) == "" {
		return errors.New("database dsn: required")
	}
	if s.MaxOpenConnections < 1 || s.MaxOpenConnections > 10000 {
		return errors.New("database maxOpenConnections: outside 1–10000")
	}
	if s.ConnectionMaxLifetimeSeconds < 0 || s.ConnectionMaxLifetimeSeconds > 604800 {
		return errors.New("database connectionMaxLifetimeSeconds: outside 0–604800")
	}
	if s.ConnectionMaxIdleTimeSeconds < 0 || s.ConnectionMaxIdleTimeSeconds > 604800 {
		return errors.New("database connectionMaxIdleTimeSeconds: outside 0–604800")
	}
	if s.ConnectTimeoutSeconds < 1 || s.ConnectTimeoutSeconds > 300 {
		return errors.New("database connectTimeoutSeconds: outside 1–300")
	}
	return nil
}
