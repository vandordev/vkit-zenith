package handlergen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var moduleLine = regexp.MustCompile(`(?m)^module\s+([a-zA-Z0-9._~/-]+)\s*$`)

func Generate(relative string) error {
	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot locate repository")
	}
	if _, err := os.Stat(filepath.Join(root, "go.work")); err != nil {
		return fmt.Errorf("invoke handler-gen from the repository root")
	}
	clean := filepath.Clean(relative)
	if relative == "" || filepath.IsAbs(relative) || clean == "." || strings.HasPrefix(clean, "..") || !strings.HasPrefix(filepath.ToSlash(clean), "apps/") {
		return fmt.Errorf("--app must select a repository-relative application")
	}
	app := filepath.Join(root, clean)
	real, err := filepath.EvalSymlinks(app)
	if err != nil || real != app {
		return fmt.Errorf("application must be a real directory, not a symlink")
	}
	goMod, err := os.ReadFile(filepath.Join(app, "go.mod"))
	if err != nil {
		return fmt.Errorf("application requires its own go.mod")
	}
	match := moduleLine.FindSubmatch(goMod)
	if len(match) != 2 {
		return fmt.Errorf("application go.mod requires an unquoted module declaration")
	}
	constructors, err := discover(app, string(match[1]))
	if err != nil {
		return err
	}
	source, err := render(string(match[1]), constructors)
	if err != nil {
		return fmt.Errorf("cannot format generated composition")
	}
	target := filepath.Join(app, "app.go")
	if info, err := os.Lstat(target); err == nil {
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%s: output must be a regular generated file", target)
		}
		existing, err := os.ReadFile(target)
		if err != nil {
			return fmt.Errorf("%s: cannot read output", target)
		}
		if !bytes.HasPrefix(existing, []byte(notice+"\n")) {
			return fmt.Errorf("%s: refusing to overwrite a manual file", target)
		}
		if bytes.Equal(existing, source) {
			return nil
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("%s: cannot inspect output", target)
	}
	file, err := os.CreateTemp(app, ".app-*.go.tmp")
	if err != nil {
		return fmt.Errorf("cannot create temporary composition")
	}
	defer os.Remove(file.Name())
	if _, err := file.Write(source); err != nil {
		_ = file.Close()
		return fmt.Errorf("cannot write composition")
	}
	if err := file.Chmod(0644); err != nil {
		_ = file.Close()
		return fmt.Errorf("cannot set composition permissions")
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("cannot close composition")
	}
	if err := os.Rename(file.Name(), target); err != nil {
		return fmt.Errorf("cannot atomically replace composition")
	}
	return nil
}
