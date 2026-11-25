package dbtool

import (
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSQLCQueryDirectories(t *testing.T) {
	config, out := fixtureConfig(t)
	root := filepath.Dir(config)
	data, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	var c sqlcConfig
	if err = yaml.Unmarshal(data, &c); err != nil {
		t.Fatal(err)
	}
	if err = os.Mkdir(filepath.Join(root, "second"), 0700); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(root, "second", "more.sql"), []byte("-- name: SecondPath :one\nSELECT 42::bigint;\n"), 0600); err != nil {
		t.Fatal(err)
	}
	c.SQL[0].Queries = sqlcPaths{"queries.sql", "second"}
	writeSQLCConfig(t, config, c)
	if err = Generate(t.Context(), config); err != nil {
		t.Fatal(err)
	}
	before := treeHashes(t, out)
	if err = Generate(t.Context(), config); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, treeHashes(t, out)) {
		t.Fatal("nondeterministic multi-path output")
	}
	for _, mode := range []string{"overlap", "symlink", "empty"} {
		t.Run(mode, func(t *testing.T) {
			c.SQL[0].Queries = sqlcPaths{"queries.sql", "second"}
			switch mode {
			case "overlap":
				c.SQL[0].Queries[1] = out
			case "symlink":
				if e := os.Symlink(filepath.Join(root, "queries.sql"), filepath.Join(root, "second", "linked.sql")); e != nil {
					t.Fatal(e)
				}
				defer os.Remove(filepath.Join(root, "second", "linked.sql"))
			case "empty":
				c.SQL[0].Queries[1] = ""
			}
			writeSQLCConfig(t, config, c)
			if e := Generate(t.Context(), config); e == nil {
				t.Fatal("unsafe query path accepted")
			}
			if !reflect.DeepEqual(before, treeHashes(t, out)) {
				t.Fatal("unsafe input replaced output")
			}
		})
	}
	for _, input := range []string{"queries: []", "queries: [42]", "queries: {path: second}"} {
		var target sqlcTarget
		if e := yaml.NewDecoder(strings.NewReader(input)).Decode(&target); e == nil {
			t.Fatalf("invalid path shape accepted: %s", input)
		}
	}
	c.SQL[0].Queries = nil
	writeSQLCConfig(t, config, c)
	if e := Generate(t.Context(), config); e == nil {
		t.Fatal("missing query paths accepted")
	}
}
