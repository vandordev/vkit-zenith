package dbtool

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func writeSQLCConfig(t *testing.T, path string, c sqlcConfig) {
	t.Helper()
	b, err := yaml.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
}

func multiTargetFixture(t *testing.T) (string, sqlcConfig, []string) {
	t.Helper()
	config, _ := fixtureConfig(t)
	root := filepath.Dir(config)
	b, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	var c sqlcConfig
	if err = yaml.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	first := c.SQL[0]
	first.Gen.Go.Out = "source2/generated1"
	second := first
	second.Queries = append(sqlcPaths(nil), first.Queries...)
	second.Gen.Go.Out = "generated2"
	second.Gen.Go.Package = "secondqueries"
	second.Schema = "second-schema.sql"
	second.Queries = sqlcPaths{"second-queries.sql"}
	for _, pair := range [][2]string{{"schema.sql", second.Schema}, {"queries.sql", second.Queries[0]}} {
		data, err := os.ReadFile(filepath.Join(root, pair[0]))
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(filepath.Join(root, pair[1]), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	c.SQL = []sqlcTarget{first, second}
	writeSQLCConfig(t, config, c)
	if err = Generate(t.Context(), config); err != nil {
		t.Fatal(err)
	}
	return config, c, []string{filepath.Join(root, first.Gen.Go.Out), filepath.Join(root, second.Gen.Go.Out)}
}

// R1 covers validation across targets, before intentionally sequential publication.
// It does not assert a transactional/atomic multi-output swap.
func TestSQLCMultiTargetBoundaries(t *testing.T) {
	for _, mode := range []string{"duplicate", "nested", "reverse-nested", "other-schema", "other-query", "source-contains-output", "malformed-second"} {
		t.Run(mode, func(t *testing.T) {
			config, c, outputs := multiTargetFixture(t)
			root := filepath.Dir(config)
			before := []map[string]string{treeHashes(t, outputs[0]), treeHashes(t, outputs[1])}
			want := "overlapping outputs"
			switch mode {
			case "duplicate":
				c.SQL[1].Gen.Go.Out = c.SQL[0].Gen.Go.Out
			case "nested":
				c.SQL[1].Gen.Go.Out = filepath.Join(c.SQL[0].Gen.Go.Out, "child")
			case "reverse-nested":
				c.SQL[0].Gen.Go.Out = filepath.Join(c.SQL[1].Gen.Go.Out, "child")
			case "other-schema":
				c.SQL[1].Schema = filepath.Join(c.SQL[0].Gen.Go.Out, "db.go")
				want = "another target source"
			case "other-query":
				c.SQL[1].Queries = sqlcPaths{filepath.Join(c.SQL[0].Gen.Go.Out, "db.go")}
				want = "another target source"
			case "source-contains-output":
				c.SQL[1].Schema = "source2"
				want = "another target source"
			case "malformed-second":
				// Target 1 would produce changed output if publication began prematurely.
				if err := os.WriteFile(filepath.Join(root, "queries.sql"), []byte("-- name: Changed :one\nSELECT 42::bigint;\n"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, c.SQL[1].Queries[0]), []byte("-- name: Broken :one\nSELECT missing FROM fixture_entries;\n"), 0600); err != nil {
					t.Fatal(err)
				}
				want = "offline generation failed; previous output unchanged"
			}
			writeSQLCConfig(t, config, c)
			err := Generate(t.Context(), config)
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("expected %s: %v", want, err)
			}
			for i, out := range outputs {
				if !reflect.DeepEqual(before[i], treeHashes(t, out)) {
					t.Fatalf("target %d changed before publication", i+1)
				}
			}
		})
	}
}

func TestSQLCPlatformCanonicalOffline(t *testing.T) {
	t.Setenv("DB_DSN", "")
	if err := os.Unsetenv("DB_DSN"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DOCKER_HOST", "tcp://127.0.0.1:1")
	t.Setenv("HOME", tempRoot(t))
	// Read the repository config; outputs are redirected only, schema stays canonical.
	b, err := os.ReadFile("../sqlc.yaml")
	if err != nil {
		t.Fatal(err)
	}
	var c sqlcConfig
	if err = yaml.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	base, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	root := tempRoot(t)
	for i := range c.SQL {
		target := &c.SQL[i]
		target.Schema = filepath.Join(base, target.Schema)
		for j := range target.Queries {
			target.Queries[j] = filepath.Join(base, target.Queries[j])
		}
		target.Gen.Go.Out = filepath.Join(root, target.Gen.Go.Package)
	}
	config := filepath.Join(root, "sqlc.yaml")
	writeSQLCConfig(t, config, c)
	if err = Generate(t.Context(), config); err != nil {
		t.Fatal(err)
	}
	for _, target := range c.SQL {
		before := treeHashes(t, target.Gen.Go.Out)
		if err = Generate(t.Context(), config); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(before, treeHashes(t, target.Gen.Go.Out)) {
			t.Fatal("canonical generation nondeterministic")
		}
		committed := filepath.Join(base, "dbtool/testdata/sqlc_fixture/generated")
		if target.Gen.Go.Package == "platformqueries" {
			committed = filepath.Join(base, "dbtool/testdata/platform/generated")
		}
		if !reflect.DeepEqual(before, treeHashes(t, committed)) {
			t.Fatal("committed generation drift")
		}
		entries, err := os.ReadDir(target.Gen.Go.Out)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			data, err := os.ReadFile(filepath.Join(target.Gen.Go.Out, e.Name()))
			if err != nil {
				t.Fatal(err)
			}
			for _, sentinel := range []string{"goose_db_version", "postgres://", "postgresql://", "DB_DSN", "DOCKER_HOST", "SECRET_SENTINEL", "Generated at"} {
				if strings.Contains(string(data), sentinel) {
					t.Fatalf("unexpected generated content %s", sentinel)
				}
			}
		}
	}
}

func TestSQLCNoOwnerQueries(t *testing.T) {
	config, out := fixtureConfig(t)
	if err := Generate(t.Context(), config); err != nil {
		t.Fatal(err)
	}
	before := treeHashes(t, out)
	// No operation owner means no scaffolding/query invented to make codegen pass.
	if err := os.WriteFile(filepath.Join(filepath.Dir(config), "queries.sql"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := Generate(t.Context(), config); err == nil {
		t.Fatal("empty query set unexpectedly generated")
	}
	if !reflect.DeepEqual(before, treeHashes(t, out)) {
		t.Fatal("empty query failure replaced prior output")
	}
}
