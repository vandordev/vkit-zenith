package dbtool

import (
	"bytes"
	"context"
	"debug/buildinfo"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const SQLCVersion = "v1.29.0"
const SQLCModuleSum = "h1:HQctoD7y/i29Bao53qXO7CZ/BV9NcvpGpsJWvz9nKWs="

type sqlcConfig struct {
	Version string       `yaml:"version"`
	SQL     []sqlcTarget `yaml:"sql"`
}
type sqlcTarget struct {
	Engine  string    `yaml:"engine"`
	Schema  string    `yaml:"schema"`
	Queries sqlcPaths `yaml:"queries"`
	Gen     struct {
		Go struct {
			Package    string `yaml:"package"`
			Out        string `yaml:"out"`
			SQLPackage string `yaml:"sql_package"`
		} `yaml:"go"`
	} `yaml:"gen"`
}

// sqlc accepts either one query path or an explicit list. Keep every path visible
// to the same source/output and symlink checks before invoking the offline CLI.
type sqlcPaths []string

func (p *sqlcPaths) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode && n.Tag == "!!str" {
		*p = sqlcPaths{n.Value}
		return nil
	}
	if n.Kind != yaml.SequenceNode || len(n.Content) == 0 {
		return errors.New("sqlc: expected query paths")
	}
	for _, child := range n.Content {
		if child.Kind != yaml.ScalarNode || child.Tag != "!!str" || child.Value == "" {
			return errors.New("sqlc: invalid query path")
		}
		*p = append(*p, child.Value)
	}
	return nil
}

func sqlcBinary(ctx context.Context) (string, error) {
	path, err := exec.LookPath("sqlc")
	if err != nil {
		return "", errors.New("sqlc: pinned CLI unavailable; see tooling provisioning guide")
	}
	info, err := buildinfo.ReadFile(path)
	if err != nil || info.Path != "github.com/sqlc-dev/sqlc/cmd/sqlc" || info.Main.Path != "github.com/sqlc-dev/sqlc" || info.Main.Version != SQLCVersion || info.Main.Sum != SQLCModuleSum || info.Main.Replace != nil {
		return "", errors.New("sqlc: binary module provenance does not match pinned release")
	}
	cmd := exec.CommandContext(ctx, path, "version")
	b, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(b)) != SQLCVersion {
		return "", errors.New("sqlc: binary version mismatch")
	}
	return path, nil
}

// Generate reads only reviewed offline SQL targets. No database selector, cloud,
// plugin, runtime YAML or inherited credential environment is supported.
func Generate(ctx context.Context, configPath string) (err error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	configPath, err = filepath.Abs(configPath)
	if err != nil {
		return err
	}
	if err := checkAncestors(configPath); err != nil {
		return err
	}
	b, err := os.ReadFile(configPath)
	if err != nil {
		return errors.New("sqlc: configuration unavailable")
	}
	var c sqlcConfig
	d := yaml.NewDecoder(bytes.NewReader(b))
	d.KnownFields(true)
	if err := d.Decode(&c); err != nil {
		return errors.New("sqlc: invalid offline configuration")
	}
	var extra any
	if d.Decode(&extra) != io.EOF || c.Version != "2" || len(c.SQL) == 0 {
		return errors.New("sqlc: expected one v2 document with targets")
	}
	base := filepath.Dir(configPath)
	outputs := make([]string, len(c.SQL))
	for i := range c.SQL {
		t := &c.SQL[i]
		if t.Engine != "postgresql" || t.Gen.Go.SQLPackage != "pgx/v5" || t.Gen.Go.Package == "" || t.Gen.Go.Out == "" || t.Schema == "" || len(t.Queries) == 0 {
			return errors.New("sqlc: unsupported target")
		}
		paths := []*string{&t.Schema, &t.Gen.Go.Out}
		for j := range t.Queries {
			paths = append(paths, &t.Queries[j])
		}
		for _, p := range paths {
			if *p == "" {
				return errors.New("sqlc: empty source path")
			}
			if !filepath.IsAbs(*p) {
				*p = filepath.Join(base, *p)
			}
			*p = filepath.Clean(*p)
			if err := checkAncestors(*p); err != nil {
				return err
			}
		}
		out := t.Gen.Go.Out
		// Outputs cannot contain source/config trees or each other.
		for _, source := range append([]string{configPath, t.Schema}, t.Queries...) {
			if containsPath(out, source) || containsPath(source, out) {
				return errors.New("sqlc: output overlaps source")
			}
		}
		for _, previous := range outputs[:i] {
			if containsPath(previous, out) || containsPath(out, previous) {
				return errors.New("sqlc: overlapping outputs")
			}
		}
		if err := validateOutput(out); err != nil {
			return err
		}
		outputs[i] = out
	}
	for _, out := range outputs {
		for _, target := range c.SQL {
			for _, source := range append([]string{target.Schema}, target.Queries...) {
				if containsPath(out, source) || containsPath(source, out) {
					return errors.New("sqlc: output overlaps another target source")
				}
				if err := filepath.WalkDir(source, func(_ string, entry fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					if entry.Type()&os.ModeSymlink != 0 || (!entry.IsDir() && !entry.Type().IsRegular()) {
						return errors.New("sqlc: nonregular source refused")
					}
					return nil
				}); err != nil {
					return err
				}
			}
		}
	}
	binary, err := sqlcBinary(ctx)
	if err != nil {
		return err
	}
	stage, err := os.MkdirTemp(base, ".sqlc-stage-")
	if err != nil {
		return err
	}
	defer func() {
		if os.RemoveAll(stage) != nil {
			err = errors.Join(err, errors.New("sqlc: staging cleanup failed"))
		}
	}()
	for i := range c.SQL {
		// sqlc v1.29 joins even absolute schema/query paths to its config dir.
		// Mirror resolution explicitly with paths relative to the staged config.
		c.SQL[i].Schema, err = filepath.Rel(stage, c.SQL[i].Schema)
		if err != nil {
			return err
		}
		for j := range c.SQL[i].Queries {
			c.SQL[i].Queries[j], err = filepath.Rel(stage, c.SQL[i].Queries[j])
			if err != nil {
				return err
			}
		}
		c.SQL[i].Gen.Go.Out = fmt.Sprint(i)
	}
	b, err = yaml.Marshal(c)
	if err != nil {
		return errors.New("sqlc: staging configuration failed")
	}
	stagedConfig := filepath.Join(stage, "sqlc.yaml")
	if err := os.WriteFile(stagedConfig, b, 0600); err != nil {
		return err
	}
	home := filepath.Join(stage, "home")
	if err := os.Mkdir(home, 0700); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, binary, "generate", "-f", stagedConfig)
	cmd.Dir = base
	cmd.Env = []string{"HOME=" + home, "PATH=" + os.Getenv("PATH"), "DOCKER_HOST=tcp://127.0.0.1:1"}
	if err := cmd.Run(); err != nil {
		return errors.Join(errors.New("sqlc: offline generation failed; previous output unchanged"), ctx.Err())
	}
	// Validate every generated tree before publishing any target.
	for _, t := range c.SQL {
		if err := ownedTree(filepath.Join(stage, t.Gen.Go.Out)); err != nil {
			return err
		}
	}
	for i, out := range outputs {
		err := publish(ctx, out, func(_ context.Context, destination string) error {
			entries, err := os.ReadDir(filepath.Join(stage, c.SQL[i].Gen.Go.Out))
			if err != nil {
				return err
			}
			for _, entry := range entries {
				if entry.IsDir() {
					return errors.New("sqlc: nested generated output refused")
				}
				b, err := os.ReadFile(filepath.Join(stage, c.SQL[i].Gen.Go.Out, entry.Name()))
				if err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(destination, entry.Name()), b, 0644); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("sqlc: publication failed after %d of %d targets: %w", i, len(outputs), err)
		}
	}
	return nil
}

func containsPath(parent, child string) bool {
	rel, err := filepath.Rel(parent, child)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}
func validateOutput(out string) error {
	if _, err := os.Lstat(out + ".backup"); !os.IsNotExist(err) {
		return errors.New("generated output: recovery backup requires inspection")
	}
	if _, err := os.Lstat(out); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	return ownedTree(out)
}
