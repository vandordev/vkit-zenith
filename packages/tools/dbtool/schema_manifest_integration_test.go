//go:build integration

package dbtool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

type platformManifest struct {
	Count              int                       `json:"count"`
	TableVersions      map[string]int            `json:"table_introduced_in"`
	ConstraintVersions map[string]map[string]int `json:"constraint_introduced_in"`
	ConstraintRemoved  map[string]map[string]int `json:"constraint_removed_in"`
	Excluded           []string                  `json:"excluded"`
	Groups             map[string]struct {
		Tests  []string          `json:"positive_negative_tests"`
		Tables map[string]string `json:"tables"`
	} `json:"groups"`
	UniqueIndexes []string `json:"unique_indexes"`
}

func loadManifest() (platformManifest, error) {
	var m platformManifest
	b, err := os.ReadFile("testdata/platform/tables.json")
	if err != nil {
		return m, err
	}
	err = json.Unmarshal(b, &m)
	return m, err
}

func TestSchemaManifest(t *testing.T) {
	versions := []int{}
	for v := 1; v <= latestSchemaVersion(t); v++ {
		versions = append(versions, v)
	}
	for _, version := range versions {
		t.Run(fmt.Sprintf("migration-%d", version), func(t *testing.T) {
			schemaPrefixDatabase(t, version, func(ctx context.Context, p *pgxpool.Pool) error {
				m, err := loadManifest()
				if err != nil {
					return err
				}
				var tables []string
				for _, group := range m.Groups {
					if len(group.Tests) == 0 {
						return errors.New("missing group coverage")
					}
					for table, suffixes := range group.Tables {
						if m.TableVersions[table] > version {
							continue
						}
						tables = append(tables, table)
						var expected []string
						for _, suffix := range strings.Fields(suffixes) {
							if removed := m.ConstraintRemoved[table][suffix]; removed != 0 && version >= removed {
								continue
							}
							if m.ConstraintVersions[table][suffix] > version {
								continue
							}
							name := table + "_" + suffix
							if strings.HasPrefix(table, "river_") && suffix != "pkey" && suffix != "pkey1" {
								name = suffix
							}
							expected = append(expected, name)
						}
						sort.Strings(expected)
						rows, err := p.Query(ctx, `SELECT conname FROM pg_constraint WHERE conrelid=$1::regclass AND contype<>'t' ORDER BY conname`, table)
						if err != nil {
							return err
						}
						var names []string
						for rows.Next() {
							var name string
							if err := rows.Scan(&name); err != nil {
								rows.Close()
								return err
							}
							names = append(names, name)
						}
						rows.Close()
						if err := rows.Err(); err != nil {
							return err
						}
						if !reflect.DeepEqual(names, expected) {
							return fmt.Errorf("constraint inventory %s: %v != %v", table, names, expected)
						}
					}
				}
				if len(tables) != 21 || m.Count != 21 || !reflect.DeepEqual(m.Excluded, []string{"goose_db_version"}) {
					return errors.New("manifest arithmetic/exclusions")
				}

				if err := checkTables(ctx, p, tables); err != nil {
					return err
				}
				// All explicit triggers and indexes must survive the actual complete migration chain.
				files, err := filepath.Glob("../../db/migrations/*.sql")
				if err != nil {
					return err
				}
				ddl := ""
				for _, file := range files {
					name, _, _ := strings.Cut(filepath.Base(file), "_")
					introduced, err := strconv.Atoi(name)
					if err != nil {
						return err
					}
					if introduced > version {
						continue
					}
					b, err := os.ReadFile(file)
					if err != nil {
						return err
					}
					ddl += strings.Split(string(b), "-- +goose Down")[0]
				}
				for _, inventory := range []struct{ pattern, drop, query string }{
					{`CREATE (?:CONSTRAINT )?TRIGGER (\w+)`, `DROP TRIGGER(?: IF EXISTS)?\s+(\w+)`, `SELECT tgname FROM pg_trigger WHERE NOT tgisinternal ORDER BY tgname`},
					{`CREATE (?:UNIQUE )?INDEX (?:IF NOT EXISTS )?(\w+)`, `DROP INDEX(?: IF EXISTS)?\s+(\w+)`, `SELECT indexname FROM pg_indexes WHERE schemaname='public' AND indexname NOT IN (SELECT conname FROM pg_constraint WHERE contype IN ('p','u','x')) AND tablename<>'goose_db_version' ORDER BY indexname`},
				} {
					var expected []string
					for _, match := range regexp.MustCompile(inventory.pattern).FindAllStringSubmatch(ddl, -1) {
						if strings.EqualFold(match[1], "ON") {
							continue // Unnamed indexes receive PostgreSQL-generated names.
						}
						expected = append(expected, match[1])
					}
					dropped := map[string]bool{}
					for _, match := range regexp.MustCompile(inventory.drop).FindAllStringSubmatch(ddl, -1) {
						dropped[match[1]] = true
					}
					expected = slices.DeleteFunc(expected, func(name string) bool { return dropped[name] })
					sort.Strings(expected)
					rows, err := p.Query(ctx, inventory.query)
					if err != nil {
						return err
					}
					var actual []string
					for rows.Next() {
						var s string
						if err := rows.Scan(&s); err != nil {
							rows.Close()
							return err
						}
						actual = append(actual, s)
					}
					rows.Close()
					if err := rows.Err(); err != nil {
						return err
					}
					if !reflect.DeepEqual(actual, expected) {
						return fmt.Errorf("trigger/index inventory mismatch %v != %v", actual, expected)
					}
				}
				for _, index := range m.UniqueIndexes {
					var unique bool
					if err := p.QueryRow(ctx, `SELECT indisunique FROM pg_index WHERE indexrelid=$1::regclass`, index).Scan(&unique); err != nil {
						return err
					}
					if !unique {
						return fmt.Errorf("not unique: %s", index)
					}
				}
				// Reverse FK prefix coverage: include PK/UNIQUE/exclusion backing indexes,
				// matching leading column sets (order within a composite equality key is free).
				rows, err := p.Query(ctx, `SELECT c.conrelid::regclass::text,c.conname FROM pg_constraint c WHERE c.contype='f' AND c.connamespace='public'::regnamespace AND NOT EXISTS (SELECT 1 FROM pg_index i WHERE i.indrelid=c.conrelid AND i.indisvalid AND i.indpred IS NULL AND i.indexprs IS NULL AND ((i.indnkeyatts>=cardinality(c.conkey) AND (SELECT array_agg(k ORDER BY k) FROM unnest(i.indkey::smallint[]) WITH ORDINALITY v(k,n) WHERE n<=cardinality(c.conkey))=(SELECT array_agg(k ORDER BY k) FROM unnest(c.conkey) k)) OR (i.indisunique AND ARRAY(SELECT k FROM unnest(i.indkey::smallint[]) WITH ORDINALITY v(k,n) WHERE n<=i.indnkeyatts)<@c.conkey)))`)
				if err != nil {
					return err
				}
				var missing []string
				for rows.Next() {
					var table, name string
					if err := rows.Scan(&table, &name); err != nil {
						rows.Close()
						return err
					}
					missing = append(missing, table+"/"+name)
				}
				rows.Close()
				if err := rows.Err(); err != nil {
					return err
				}
				if len(missing) > 0 {
					return fmt.Errorf("missing reverse FK indexes: %v", missing)
				}
				return nil
			})
		})
	}
}
