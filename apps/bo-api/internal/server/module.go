package server

import (
	"flag"
	"log/slog"
	"os"

	config "github.com/vandordev/vkit-zenith/packages/config/go/config"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

var root = flag.String("root", "", "Trusted root containing config/config.yaml")
var development = flag.Bool("development", false, "Explicit root .env opt-in")

type runtimeConfig struct {
	API      config.BOAPIConfig
	Database config.DatabaseConfig
}

func loadConfig() (runtimeConfig, error) {
	trusted, err := config.ApplicationRoot(*root)
	if err != nil {
		return runtimeConfig{}, err
	}
	options := config.Options{Root: trusted}
	if *development {
		options.Environment, err = config.DevelopmentEnvironment(trusted, nil)
		if err != nil {
			return runtimeConfig{}, err
		}
	}
	api, err := config.LoadBOAPI(options)
	if err != nil {
		return runtimeConfig{}, err
	}
	database, err := config.LoadDatabase(options)
	if err != nil {
		return runtimeConfig{}, err
	}
	return runtimeConfig{api, database}, nil
}
func newLogger(c runtimeConfig) *slog.Logger {
	var level slog.Level
	_ = level.UnmarshalText([]byte(c.API.App.LogLevel))
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})).With("service", "bo-api")
}

// The runtime graph alone selects config and installs lifecycle hooks.
var Module = fx.Options(fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }), fx.Provide(loadConfig, newLogger, newEcho, newAPI, newBoundary), fx.Invoke(lifecycle))

// Export never constructs a pool, reads config/secrets or registers lifecycle.
// Export dependencies cannot serve requests; offlineEcho refuses route dispatch.
var ExportModule = fx.Options(fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }), fx.Provide(offlineEcho, newAPI, exportBoundary))
