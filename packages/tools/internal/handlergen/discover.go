package handlergen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const marker = "//vkit:handler"

type constructor struct{ Import, Name string }

var variant = regexp.MustCompile(`_(?:aix|android|darwin|dragonfly|freebsd|illumos|ios|js|linux|netbsd|openbsd|plan9|solaris|wasip1|windows|386|amd64|arm|arm64|loong64|mips|mipsle|mips64|mips64le|ppc64|ppc64le|riscv64|s390x|wasm)(?:_[a-z0-9]+)?\.go$`)

func discover(app, module string) ([]constructor, error) {
	root := filepath.Join(app, "internal", "handlers")
	var result []constructor
	seen := make(map[string]bool)
	packages := make(map[string]string)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%s: cannot inspect handler sources", path)
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("%s: symlinks are not supported", path)
		}
		if entry.IsDir() {
			if strings.HasPrefix(entry.Name(), ".") || entry.Name() == "testdata" {
				return filepath.SkipDir
			}
			if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
				return fmt.Errorf("%s: nested module is not supported", path)
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("%s: invalid Go syntax", path)
		}
		if ast.IsGenerated(file) {
			return nil
		}
		if variant.MatchString(entry.Name()) {
			return fmt.Errorf("%s: platform-specific handler files are unsupported", path)
		}
		for _, group := range file.Comments {
			for _, comment := range group.List {
				if strings.HasPrefix(comment.Text, "//go:build") || strings.HasPrefix(comment.Text, "// +build") {
					return fmt.Errorf("%s: build-tagged handler files are unsupported", fset.Position(comment.Pos()))
				}
			}
		}
		directory := filepath.Dir(path)
		if existing := packages[directory]; existing != "" && existing != file.Name.Name {
			return fmt.Errorf("%s: conflicting package names", path)
		}
		packages[directory] = file.Name.Name
		if file.Name.Name == "main" {
			return fmt.Errorf("%s: handlers must be importable packages", path)
		}
		attached := make(map[*ast.Comment]bool)
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Doc == nil {
				continue
			}
			marked := false
			for _, comment := range fn.Doc.List {
				if comment.Text == marker {
					if marked {
						return fmt.Errorf("%s: duplicate handler marker", fset.Position(comment.Pos()))
					}
					marked = true
					attached[comment] = true
				}
			}
			if !marked {
				continue
			}
			fail := func(reason string) error { return fmt.Errorf("%s: %s", fset.Position(fn.Pos()), reason) }
			if fn.Recv != nil || !ast.IsExported(fn.Name.Name) || fn.Type.TypeParams != nil || fn.Body == nil {
				return fail("handler requires an exported package-level, non-generic implementation")
			}
			if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
				return fail("handler must return only error")
			}
			output := fn.Type.Results.List[0]
			ident, ok := output.Type.(*ast.Ident)
			if !ok || ident.Name != "error" || len(output.Names) > 1 {
				return fail("handler must return only error")
			}
			for _, input := range fn.Type.Params.List {
				if _, variadic := input.Type.(*ast.Ellipsis); variadic {
					return fail("variadic handler constructors are unsupported")
				}
			}
			rel, err := filepath.Rel(app, directory)
			if err != nil {
				return fail("invalid handler package path")
			}
			importPath := module + "/" + filepath.ToSlash(rel)
			key := importPath + "." + fn.Name.Name
			if seen[key] {
				return fail("duplicate marked constructor")
			}
			seen[key] = true
			result = append(result, constructor{Import: importPath, Name: fn.Name.Name})
		}
		for _, group := range file.Comments {
			for _, comment := range group.List {
				if comment.Text == marker && !attached[comment] {
					return fmt.Errorf("%s: handler marker must directly document a constructor", fset.Position(comment.Pos()))
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Import != result[j].Import {
			return result[i].Import < result[j].Import
		}
		return result[i].Name < result[j].Name
	})
	if len(result) == 0 {
		return nil, fmt.Errorf("%s: no marked constructors", root)
	}
	return result, nil
}
