package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const dbFixture = `
db:
  dsn: "${DB_DSN}"
  maxOpenConnections: 10
  connectionMaxLifetimeSeconds: 1800
  connectionMaxIdleTimeSeconds: 300
  connectTimeoutSeconds: 10
`

func fixture(t *testing.T, transform func(string) string) Options {
	t.Helper()
	source, err := os.ReadFile("../../../config/config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	if !strings.Contains(text, "db:\n") {
		text += dbFixture
	}
	if transform != nil {
		text = transform(text)
	}
	root, err := os.MkdirTemp("/tmp/opencode", "config-db-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(root) })
	if err := os.Mkdir(filepath.Join(root, "config"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "config/config.yaml"), []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
	return Options{Root: root, Environment: map[string]string{}}
}

func TestDatabaseSelectionIsolation(t *testing.T) {
	o := fixture(t, nil)
	gotPublic, err := LoadAPI(o)
	if err != nil {
		t.Fatal(err)
	}
	if gotPublic.API.Port != 4102 {
		t.Fatalf("public API port = %d, want 4102", gotPublic.API.Port)
	}
	gotWeb, err := LoadWebAPI(o)
	if err != nil {
		t.Fatal(err)
	}
	if gotWeb.API.Port != 4101 {
		t.Fatalf("web API port = %d, want 4101", gotWeb.API.Port)
	}
	if _, err := LoadDatabase(o); err == nil {
		t.Fatal("missing DB secret accepted")
	}
	if _, err := LoadBOAPI(o); err == nil {
		t.Fatal("BO selector accepted missing BO secrets")
	}
	o.Environment["DB_DSN"] = "postgres://fixture:sentinel-secret@localhost:5432/fixture?sslmode=disable"
	db, err := LoadDatabase(o)
	if err != nil {
		t.Fatal(err)
	}
	if db.DSN != o.Environment["DB_DSN"] || db.MaxOpenConnections != 10 || db.ConnectTimeoutSeconds != 10 {
		t.Fatal("incorrect selection")
	}
}

func TestDatabaseInvalidSettingsSafe(t *testing.T) {
	for _, pair := range [][2]string{
		{"maxOpenConnections: 10", "maxOpenConnections: 0"},
		{"maxOpenConnections: 10", "maxOpenConnections: 10001"},
		{"maxOpenConnections: 10", "maxOpenConnections: 10\n  maxIdleConnections: 5"},
		{"connectionMaxLifetimeSeconds: 1800", "connectionMaxLifetimeSeconds: 604801"},
		{"connectionMaxIdleTimeSeconds: 300", "connectionMaxIdleTimeSeconds: -1"},
		{"connectTimeoutSeconds: 10", "connectTimeoutSeconds: 301"},
		{"connectTimeoutSeconds: 10", "connectTimeoutSeconds: 1.5"},
		{"connectTimeoutSeconds: 10", "unknown: sentinel-secret"},
		{"connectTimeoutSeconds: 10", "connectTimeoutSeconds: 10\n  connectTimeoutSeconds: 2"},
		{"dsn: \"${DB_DSN}\"", "dsn: \"${broken\""},
	} {
		t.Run(pair[1], func(t *testing.T) {
			o := fixture(t, func(s string) string { return strings.Replace(s, pair[0], pair[1], 1) })
			o.Environment["DB_DSN"] = "sentinel-secret"
			_, err := LoadDatabase(o)
			if err == nil {
				t.Fatal("invalid config accepted")
			}
			if strings.Contains(err.Error(), "sentinel-secret") {
				t.Fatal("secret leaked")
			}
		})
	}
}
