package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/danielgtaylor/huma/v2"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/server"
	"go.uber.org/fx"
)

func main() {
	export := flag.String("export-openapi", "", "Offline schema only; no config, DB, secrets or listener")
	flag.Parse()
	if *export != "" {
		if err := exportOpenAPI(*export); err != nil {
			fmt.Fprintln(os.Stderr, "BO schema export failed")
			os.Exit(1)
		}
		return
	}
	app := newApp()
	if app.Err() != nil {
		fmt.Fprintln(os.Stderr, "BO API configuration or composition failed")
		os.Exit(1)
	}
	app.Run()
}
func exportOpenAPI(path string) error {
	var api huma.API
	app := fx.New(server.ExportModule, handlerOptions(), fx.Populate(&api))
	if err := app.Err(); err != nil {
		return err
	}
	data, err := json.MarshalIndent(api.OpenAPI(), "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}
