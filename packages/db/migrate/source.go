// Package migrate exposes explicit Goose operations, never startup hooks.
package migrate

import (
	"bufio"
	"errors"
	"io/fs"
	"sort"
	"strings"

	"github.com/pressly/goose/v3"
)

// Source validates readable, uniquely versioned SQL inputs. Goose itself parses
// SQL/statement annotations at execution; the Up marker check is not a SQL parser.
func Source(source fs.FS) ([]int64, error) {
	bad := errors.New("migration source: unreadable or invalid SQL migration files")
	if source == nil {
		return nil, bad
	}
	files, err := fs.ReadDir(source, ".")
	if err != nil {
		return nil, bad
	}
	versions := []int64{}
	seen := map[int64]bool{}
	for _, file := range files {
		if file.Type()&fs.ModeSymlink != 0 {
			return nil, bad
		}
		if file.Name() == "README.md" && !file.IsDir() {
			continue
		}
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			return nil, bad
		}
		v, err := goose.NumericComponent(file.Name())
		if err != nil || v < 1 || seen[v] {
			return nil, bad
		}
		seen[v] = true
		f, err := source.Open(file.Name())
		if err != nil {
			return nil, bad
		}
		scan := bufio.NewScanner(f)
		up := false
		for scan.Scan() {
			if strings.TrimSpace(scan.Text()) == "-- +goose Up" {
				up = true
			}
		}
		scanErr := scan.Err()
		closeErr := f.Close()
		if scanErr != nil || closeErr != nil || !up {
			return nil, bad
		}
		versions = append(versions, v)
	}
	sort.Slice(versions, func(i, j int) bool { return versions[i] < versions[j] })
	return versions, nil
}
