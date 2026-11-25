// Command scaffold creates handler, usecase and migration files that follow the
// repository patterns. Run it through `task create:*` from the repository root.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vandordev/vkit-zenith/packages/tools/internal/scaffold"
)

func main() {
	if len(os.Args) < 2 {
		fail(fmt.Errorf("usage: scaffold handler|usecase|migration [flags]"))
	}
	set := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	root := set.String("root", ".", "Repository root")
	var written []string
	var err error
	switch os.Args[1] {
	case "handler":
		var o scaffold.HandlerOptions
		set.StringVar(&o.App, "app", "", "web-api, bo-api or api")
		set.StringVar(&o.Name, "name", "", "resource_action, e.g. projects_list")
		set.StringVar(&o.Kind, "kind", "", "query or mutation")
		set.StringVar(&o.Method, "method", "", "HTTP method")
		set.StringVar(&o.Path, "path", "", "/v1/... path")
		_ = set.Parse(os.Args[2:])
		written, err = scaffold.Handler(*root, o)
	case "usecase":
		var o scaffold.UsecaseOptions
		set.StringVar(&o.Domain, "domain", "", "application package, e.g. projects")
		set.StringVar(&o.Name, "name", "", "exported verb, e.g. Create")
		_ = set.Parse(os.Args[2:])
		written, err = scaffold.Usecase(*root, o)
		if err == nil {
			defer fmt.Printf("next: add %s to packages/application/sqlc.yaml queries if it is new, write the SQL, run `task gen:db`\n", o.Domain)
		}
	case "migration":
		name := set.String("name", "", "snake_case, e.g. add_projects")
		_ = set.Parse(os.Args[2:])
		var path string
		path, err = scaffold.Migration(*root, *name)
		written = []string{path}
		if err == nil {
			defer fmt.Println("next: update packages/db/schema.dbml in the same change, then `task check:db-schema`")
		}
	default:
		err = fmt.Errorf("unknown generator %q: want handler, usecase or migration", os.Args[1])
	}
	if err != nil {
		fail(err)
	}
	for _, path := range written {
		fmt.Println("created", path)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
