package scaffold

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite golden files")

// fakeRoot creates the directories the generators write into, like a repository checkout.
func fakeRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, dir := range []string{"apps/web-api/internal/handlers/v1", "apps/bo-api/internal/handlers/v1", "apps/api/internal/handlers/v1", "packages/application", "packages/db/migrations"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "packages/db/migrations/00001_baseline.sql"), []byte("-- +goose Up\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

func golden(t *testing.T, root string, written []string) {
	t.Helper()
	for _, path := range written {
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		rel, _ := filepath.Rel(root, path)
		file := filepath.Join("testdata/golden", strings.ReplaceAll(rel, "/", "__")+".golden")
		if *update {
			if err := os.WriteFile(file, got, 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("missing golden %s (run with -update): %v", file, err)
		}
		if string(got) != string(want) {
			t.Errorf("%s differs from %s:\n%s", rel, file, got)
		}
	}
}

func TestHandlerGolden(t *testing.T) {
	root := fakeRoot(t)
	for _, o := range []HandlerOptions{
		{App: "web-api", Name: "projects_list", Kind: "query", Method: "GET", Path: "/v1/organizations/{organizationId}/projects"},
		{App: "web-api", Name: "projects_create", Kind: "mutation", Method: "POST", Path: "/v1/organizations/{organizationId}/projects"},
		{App: "bo-api", Name: "projects_get", Kind: "query", Method: "GET", Path: "/v1/projects/{projectId}"},
		{App: "bo-api", Name: "projects_archive", Kind: "mutation", Method: "POST", Path: "/v1/projects/{projectId}/archive"},
		{App: "api", Name: "projects_list", Kind: "query", Method: "GET", Path: "/v1/projects"},
		{App: "api", Name: "projects_create", Kind: "mutation", Method: "POST", Path: "/v1/projects"},
	} {
		written, err := Handler(root, o)
		if err != nil {
			t.Fatalf("%+v: %v", o, err)
		}
		golden(t, root, written)
	}
}

func TestUsecaseGolden(t *testing.T) {
	root := fakeRoot(t)
	first, err := Usecase(root, UsecaseOptions{Domain: "projects", Name: "Create"})
	if err != nil {
		t.Fatal(err)
	}
	// A second usecase in an existing domain adds only its own files.
	second, err := Usecase(root, UsecaseOptions{Domain: "projects", Name: "Archive"})
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 3 || len(second) != 2 {
		t.Fatalf("new domain wrote %d files, existing domain %d; want 3 and 2", len(first), len(second))
	}
	golden(t, root, append(first, second...))
}

func TestMigrationNumbering(t *testing.T) {
	root := fakeRoot(t)
	path, err := Migration(root, "add_projects")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "00002_add_projects.sql" {
		t.Fatalf("got %s", path)
	}
	next, err := Migration(root, "add_project_tags")
	if err != nil || filepath.Base(next) != "00003_add_project_tags.sql" {
		t.Fatalf("got %s, %v", next, err)
	}
	golden(t, root, []string{path})
}

func TestRefusesOverwrite(t *testing.T) {
	root := fakeRoot(t)
	o := HandlerOptions{App: "bo-api", Name: "projects_list", Kind: "query", Method: "GET", Path: "/v1/projects"}
	written, err := Handler(root, o)
	if err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(written[0])
	if err := os.WriteFile(written[0], append(before, []byte("// local edit\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Handler(root, o); err == nil || !strings.Contains(err.Error(), "exists") {
		t.Fatalf("second run error = %v, want exists", err)
	}
	after, _ := os.ReadFile(written[0])
	if !strings.HasSuffix(string(after), "// local edit\n") {
		t.Fatal("existing handler was overwritten")
	}
	if _, err := Usecase(root, UsecaseOptions{Domain: "projects", Name: "Create"}); err != nil {
		t.Fatal(err)
	}
	if _, err := Usecase(root, UsecaseOptions{Domain: "projects", Name: "Create"}); err == nil || !strings.Contains(err.Error(), "exists") {
		t.Fatalf("second usecase run error = %v, want exists", err)
	}
}

func TestRejectsInvalid(t *testing.T) {
	ok := HandlerOptions{App: "bo-api", Name: "projects_list", Kind: "query", Method: "GET", Path: "/v1/projects"}
	for _, tc := range []struct {
		name   string
		mutate func(*HandlerOptions)
		want   string
	}{
		{"unknown app", func(o *HandlerOptions) { o.App = "checkout-api" }, "app"},
		{"name case", func(o *HandlerOptions) { o.Name = "Projects-Create" }, "name"},
		{"single token name", func(o *HandlerOptions) { o.Name = "projects" }, "name"},
		{"missing kind", func(o *HandlerOptions) { o.Kind = "" }, "kind"},
		{"unknown kind", func(o *HandlerOptions) { o.Kind = "read" }, "kind"},
		{"unknown method", func(o *HandlerOptions) { o.Method = "FETCH" }, "method"},
		{"query must be GET", func(o *HandlerOptions) { o.Method = "POST" }, "query"},
		{"mutation must not be GET", func(o *HandlerOptions) { o.Kind = "mutation" }, "mutation"},
		{"unversioned path", func(o *HandlerOptions) { o.Path = "/projects" }, "path"},
		{"bad path parameter", func(o *HandlerOptions) { o.Path = "/v1/projects/{project-id}" }, "path"},
		{"web path without organization", func(o *HandlerOptions) { o.App = "web-api" }, "organizationId"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := fakeRoot(t)
			o := ok
			tc.mutate(&o)
			_, err := Handler(root, o)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want mention of %q", err, tc.want)
			}
			entries, _ := os.ReadDir(filepath.Join(root, "apps", "bo-api", "internal", "handlers", "v1"))
			web, _ := os.ReadDir(filepath.Join(root, "apps", "web-api", "internal", "handlers", "v1"))
			if len(entries)+len(web) != 0 {
				t.Fatal("rejected input wrote files")
			}
		})
	}
	for _, o := range []UsecaseOptions{{Domain: "Bad-Domain", Name: "Create"}, {Domain: "projects", Name: "create"}, {Domain: "staff", Name: ""}} {
		if _, err := Usecase(fakeRoot(t), o); err == nil {
			t.Errorf("usecase %+v accepted", o)
		}
	}
	for _, name := range []string{"", "AddProjects", "add projects", "../x"} {
		if _, err := Migration(fakeRoot(t), name); err == nil {
			t.Errorf("migration %q accepted", name)
		}
	}
}
