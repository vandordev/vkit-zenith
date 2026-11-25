//go:build integration

package scaffold

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestScaffoldCompiles proves generated code builds against the real APIs and packages,
// not only that it matches a golden file. It works on a copy of the tracked tree.
func TestScaffoldCompiles(t *testing.T) {
	repo, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	list, err := exec.Command("git", "-C", repo, "ls-files", "-z").Output()
	if err != nil {
		t.Fatalf("git ls-files: %v", err)
	}
	root := t.TempDir()
	for _, rel := range strings.Split(strings.TrimRight(string(list), "\x00"), "\x00") {
		if rel == "" || strings.HasPrefix(rel, "apps/") && strings.Contains(rel, "/node_modules/") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(repo, rel))
		if err != nil {
			continue // tracked but deleted in the working tree
		}
		dst := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, o := range []HandlerOptions{
		{App: "web-api", Name: "projects_list", Kind: "query", Method: "GET", Path: "/v1/organizations/{organizationId}/projects"},
		{App: "web-api", Name: "projects_create", Kind: "mutation", Method: "POST", Path: "/v1/organizations/{organizationId}/projects"},
		{App: "bo-api", Name: "projects_get", Kind: "query", Method: "GET", Path: "/v1/projects/{projectId}"},
		{App: "bo-api", Name: "projects_archive", Kind: "mutation", Method: "POST", Path: "/v1/projects/{projectId}/archive"},
		{App: "api", Name: "projects_list", Kind: "query", Method: "GET", Path: "/v1/projects"},
		{App: "api", Name: "projects_create", Kind: "mutation", Method: "PATCH", Path: "/v1/projects/{projectId}"},
	} {
		if _, err := Handler(root, o); err != nil {
			t.Fatalf("%+v: %v", o, err)
		}
	}
	for _, name := range []string{"Create", "Archive"} {
		if _, err := Usecase(root, UsecaseOptions{Domain: "projects", Name: name}); err != nil {
			t.Fatal(err)
		}
	}
	for _, step := range []struct{ dir, gen string }{{"apps/web-api", "apps/web-api"}, {"apps/bo-api", "apps/bo-api"}, {"apps/api", "apps/api"}} {
		// Register the new constructors exactly as `task gen:handlers` would.
		gen := exec.Command("go", "run", "./packages/tools/cmd/handler-gen", "--app", step.gen)
		gen.Dir = root
		if out, err := gen.CombinedOutput(); err != nil {
			t.Fatalf("handler-gen %s: %v\n%s", step.gen, err, out)
		}
		run(t, filepath.Join(root, step.dir), "go", "vet", "./...")
	}
	run(t, filepath.Join(root, "packages/application"), "go", "vet", "./projects/...")
}

func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v in %s: %v\n%s", name, args, dir, err, out)
	}
}
