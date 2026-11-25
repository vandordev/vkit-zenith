package config

import (
	"strconv"
	"strings"
)

// DatabaseConfig is selected only by explicit database consumers. Units are seconds.
type DatabaseConfig struct {
	DSN                          string `validate:"required"`
	MaxOpenConnections           int    `validate:"min=1,max=10000"`
	ConnectionMaxLifetimeSeconds int    `validate:"min=0,max=604800"`
	ConnectionMaxIdleTimeSeconds int    `validate:"min=0,max=604800"`
	ConnectTimeoutSeconds        int    `validate:"min=1,max=300"`
}

func LoadDatabase(options Options) (DatabaseConfig, error) {
	d, err := loadDocument(options, "db")
	if err != nil {
		return DatabaseConfig{}, err
	}
	return d.selectedDatabase("db")
}
func (d *document) selectedDatabase(path ...string) (DatabaseConfig, error) {
	var out DatabaseConfig
	var err error
	if out.DSN, err = d.stringAt(append(path, "dsn")...); err != nil {
		return DatabaseConfig{}, err
	}
	for _, field := range []struct {
		key    string
		target *int
		max    int
	}{
		{"maxOpenConnections", &out.MaxOpenConnections, 10000},
		{"connectionMaxLifetimeSeconds", &out.ConnectionMaxLifetimeSeconds, 604800},
		{"connectionMaxIdleTimeSeconds", &out.ConnectionMaxIdleTimeSeconds, 604800},
		{"connectTimeoutSeconds", &out.ConnectTimeoutSeconds, 300},
	} {
		value, err := d.valueAt(append(path, field.key)...)
		if err != nil {
			return DatabaseConfig{}, err
		}
		n := -1
		switch v := value.(type) {
		case string:
			if digits.MatchString(v) {
				if parsed, e := strconv.Atoi(v); e == nil {
					n = parsed
				}
			}
		case float64:
			if v >= 0 && v <= float64(field.max) && v == float64(int(v)) {
				n = int(v)
			}
		}
		if n < 0 || n > field.max {
			return DatabaseConfig{}, d.fail(strings.Join(append(path, field.key), "."), "expected bounded integer")
		}
		*field.target = n
	}
	if err := d.validate(out); err != nil {
		return DatabaseConfig{}, err
	}
	return out, nil
}
