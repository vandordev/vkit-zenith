package server

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/identityhttp"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/orgaccess"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/persistence"
	"github.com/vandordev/vkit-zenith/packages/application/identity"
	config "github.com/vandordev/vkit-zenith/packages/config/go/config"
	db "github.com/vandordev/vkit-zenith/packages/db"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

type runtimeConfig struct {
	API      config.WebAPIConfig
	Identity config.IdentityConfig
	Database config.DatabaseConfig
}

func loadConfig() (runtimeConfig, error) {
	trustedRoot, err := config.ApplicationRoot(*root)
	if err != nil {
		return runtimeConfig{}, err
	}
	options := config.Options{Root: trustedRoot}
	if *development {
		options.Environment, err = config.DevelopmentEnvironment(trustedRoot, nil)
		if err != nil {
			return runtimeConfig{}, err
		}
	}
	api, err := config.LoadWebAPI(options)
	if err != nil {
		return runtimeConfig{}, err
	}
	identityConfig, err := config.LoadIdentity(options)
	if err != nil {
		return runtimeConfig{}, err
	}
	database, err := config.LoadDatabase(options)
	if err != nil {
		return runtimeConfig{}, err
	}
	return runtimeConfig{API: api, Identity: identityConfig, Database: database}, nil
}
func newLogger(value runtimeConfig) *slog.Logger {
	var level slog.Level
	_ = level.UnmarshalText([]byte(value.API.App.LogLevel))
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})).With("service", "web-api")
}

// Module owns only this service's runtime infrastructure. FX construction
// registers operations before OnStart can bind a listener.
var Module = fx.Options(
	fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
	fx.Provide(loadConfig, newLogger, newEcho, newAPI, openPool, newIdentityRuntime, newQueries, orgaccess.NewBoundary),
	fx.Invoke(lifecycle),
)

// ExportModule deliberately has no runtime config, auth credentials, logger or
// listener lifecycle. It registers the same Huma configuration and operations.
var ExportModule = fx.Options(
	fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
	fx.Provide(exportEcho, newAPI),
	fx.Provide(identityhttp.OfflineRuntime),
	fx.Provide(func() *persistence.Queries { return persistence.New(nil) }),
	fx.Provide(func() *orgaccess.Boundary { return orgaccess.NewBoundary(nil, nil) }),
)

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

func openPool(lifecycle fx.Lifecycle, value runtimeConfig) (*pgxpool.Pool, error) {
	c := value.Database
	pool, err := db.Open(context.Background(), db.Settings{DSN: c.DSN, MaxOpenConnections: c.MaxOpenConnections, ConnectionMaxLifetimeSeconds: c.ConnectionMaxLifetimeSeconds, ConnectionMaxIdleTimeSeconds: c.ConnectionMaxIdleTimeSeconds, ConnectTimeoutSeconds: c.ConnectTimeoutSeconds})
	if err != nil {
		return nil, errors.New("api database unavailable")
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error { pool.Close(); return nil }})
	return pool, nil
}

func newQueries(pool *pgxpool.Pool) *persistence.Queries { return persistence.New(pool) }

func newIdentityRuntime(pool *pgxpool.Pool, value runtimeConfig) (*identityhttp.Runtime, error) {
	idc := value.Identity.Identity
	sender, err := identityhttp.NewDeliverySender(idc.AuthOrigin, idc.Development, idc.Email.ResendKey, idc.Email.From, nil)
	if err != nil {
		return nil, errors.New("api email delivery configuration invalid")
	}
	service := identity.NewService(pool, systemClock{}, sender, idc.LimiterKey, idc.CSRFKey, idc.AuthOrigin, idc.Development)
	return identityhttp.NewRuntime(service, idc.AuthOrigin, idc.Development), nil
}
