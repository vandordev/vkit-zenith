package dbtool

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func treeHashes(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, e fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if e.IsDir() {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		out[rel] = fmt.Sprintf("%x", sha256.Sum256(b))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func fixtureConfig(t *testing.T) (string, string) {
	t.Helper()
	root := tempRoot(t)
	for name, body := range map[string]string{
		"schema.sql":  "CREATE TABLE fixture_entries (id bigint PRIMARY KEY, label text NOT NULL);",
		"queries.sql": "-- name: FixtureListEntries :many\nSELECT id, label FROM fixture_entries ORDER BY id;\n",
		"sqlc.yaml":   "version: '2'\nsql:\n  - engine: postgresql\n    schema: schema.sql\n    queries: queries.sql\n    gen:\n      go:\n        package: fixturequeries\n        out: generated\n        sql_package: pgx/v5\n",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(root, "sqlc.yaml"), filepath.Join(root, "generated")
}

func TestSQLCOfflineDeterministicAndFailureSafety(t *testing.T) {
	t.Setenv("DB_DSN", "")
	if err := os.Unsetenv("DB_DSN"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", tempRoot(t))
	t.Setenv("DOCKER_HOST", "tcp://127.0.0.1:1")
	config, out := fixtureConfig(t)
	if err := Generate(t.Context(), config); err != nil {
		t.Fatal(err)
	}
	before := treeHashes(t, out)
	if len(before) != 3 {
		t.Fatalf("expected 3 generated files, got %d", len(before))
	}
	if err := Generate(t.Context(), config); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, treeHashes(t, out)) {
		t.Fatal("nondeterministic output")
	}
	for _, query := range []string{"-- name: Broken :one\nSELECT missing FROM fixture_entries;", "-- name: Duplicate :one\nSELECT 1;\n-- name: Duplicate :one\nSELECT 2;"} {
		if err := os.WriteFile(filepath.Join(filepath.Dir(config), "queries.sql"), []byte(query), 0600); err != nil {
			t.Fatal(err)
		}
		if Generate(t.Context(), config) == nil {
			t.Fatal("invalid SQL accepted")
		}
		if !reflect.DeepEqual(before, treeHashes(t, out)) {
			t.Fatal("failure changed old output")
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if Generate(ctx, config) == nil {
		t.Fatal("cancellation ignored")
	}
	if !reflect.DeepEqual(before, treeHashes(t, out)) {
		t.Fatal("cancel changed output")
	}
}

func TestSQLCRejectsUnreviewedConfiguration(t *testing.T) {
	config, _ := fixtureConfig(t)
	for _, source := range []string{"version: '2'\nplugins: []\n", "version: '2'\ncloud: {}\n", "version: '2'\nsql: []\n", "version: '2'\n---\nversion: '2'\n"} {
		if err := os.WriteFile(config, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		if Generate(t.Context(), config) == nil {
			t.Fatal("unreviewed config accepted")
		}
	}
}
