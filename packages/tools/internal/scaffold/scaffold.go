// Package scaffold writes new handler, usecase and migration files that follow the
// repository's patterns (ADR-0004). It only creates files and never overwrites one.
package scaffold

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
)

//go:embed templates/*.tmpl
var files embed.FS

var templates = template.Must(template.New("").ParseFS(files, "templates/*.tmpl"))

type HandlerOptions struct{ App, Name, Kind, Method, Path string }
type UsecaseOptions struct{ Domain, Name string }

type handlerData struct {
	HandlerOptions
	Constructor, OperationID, Type string
	MethodConst                    string
	List                           bool
	Params                         []param
	Module                         string
}
type param struct{ Field, Name string }

const module = "github.com/vandordev/vkit-zenith"

// Handler writes apps/<app>/internal/handlers/v1/<name>.go and returns the written path.
func Handler(root string, o HandlerOptions) ([]string, error) {
	params, err := o.validate()
	if err != nil {
		return nil, err
	}
	resource, action := split(o.Name)
	d := handlerData{HandlerOptions: o, MethodConst: "Method" + pascal(strings.ToLower(o.Method)), Constructor: "New" + pascal(action) + pascal(resource), Type: "V1" + pascal(o.Name), List: action == "list", Module: module}
	if prefix := apps[o.App]; prefix != "" {
		d.OperationID = prefix + pascal(action) + pascal(resource)
	} else {
		d.OperationID = action + pascal(resource)
	}
	for _, p := range params {
		if o.App == "web-api" && p == "organizationId" {
			continue // the web template always declares the organization scope
		}
		field := strings.ToUpper(p[:1]) + p[1:]
		if strings.HasSuffix(field, "Id") {
			field = strings.TrimSuffix(field, "Id") + "ID"
		}
		d.Params = append(d.Params, param{Field: field, Name: p})
	}
	path := filepath.Join(root, "apps", o.App, "internal", "handlers", "v1", o.Name+".go")
	if err := render(path, o.App+"-"+o.Kind+".go.tmpl", d, true); err != nil {
		return nil, err
	}
	return []string{path}, nil
}

// Usecase writes the usecase method and its SQL stub; a new domain also gets its Service.
func Usecase(root string, o UsecaseOptions) ([]string, error) {
	if err := o.validate(); err != nil {
		return nil, err
	}
	dir := filepath.Join(root, "packages", "application", o.Domain)
	snake := toSnake(o.Name)
	method, sql := filepath.Join(dir, snake+".go"), filepath.Join(dir, snake+".sql")
	for _, p := range []string{method, sql} {
		if _, err := os.Stat(p); err == nil {
			return nil, fmt.Errorf("%s exists; refusing to overwrite", p)
		}
	}
	data := struct {
		UsecaseOptions
		Snake string
	}{o, snake}
	var written []string
	service := filepath.Join(dir, "service.go")
	if _, err := os.Stat(service); errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		if err := render(service, "service.go.tmpl", data, true); err != nil {
			return nil, err
		}
		written = append(written, service)
	}
	if err := render(method, "usecase.go.tmpl", data, true); err != nil {
		return written, err
	}
	if err := render(sql, "usecase.sql.tmpl", data, false); err != nil {
		return append(written, method), err
	}
	return append(written, method, sql), nil
}

var migrationFile = regexp.MustCompile(`^(\d{5})_.*\.sql$`)

// Migration writes the next sequential Goose migration and returns its path.
func Migration(root, name string) (string, error) {
	if !snakeName.MatchString(name) {
		return "", fmt.Errorf("name %q: want snake_case, e.g. add_projects", name)
	}
	dir := filepath.Join(root, "packages", "db", "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var versions []int
	for _, e := range entries {
		if m := migrationFile.FindStringSubmatch(e.Name()); m != nil {
			n, _ := strconv.Atoi(m[1])
			versions = append(versions, n)
		}
	}
	sort.Ints(versions)
	next := 1
	if len(versions) > 0 {
		next = versions[len(versions)-1] + 1
	}
	path := filepath.Join(dir, fmt.Sprintf("%05d_%s.sql", next, name))
	return path, render(path, "migration.sql.tmpl", name, false)
}

func toSnake(exported string) string {
	var b strings.Builder
	for i, r := range exported {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}

// render executes a template and creates the file exclusively; Go sources are gofmt'd first.
func render(path, name string, data any, gofmt bool) error {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, name, data); err != nil {
		return err
	}
	out := buf.Bytes()
	if gofmt {
		formatted, err := format.Source(out)
		if err != nil {
			return fmt.Errorf("%s: generated Go does not parse: %w", path, err)
		}
		out = formatted
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if errors.Is(err, fs.ErrExist) {
		return fmt.Errorf("%s exists; refusing to overwrite", path)
	}
	if err != nil {
		return err
	}
	if _, err := f.Write(out); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
