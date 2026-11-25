package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
	config "github.com/vandordev/vkit-zenith/packages/config/go/config"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

type runtimeConfig struct {
	API config.APIConfig
}

func loadConfig() (runtimeConfig, error) {
	root, err := config.ApplicationRoot(*root)
	if err != nil {
		return runtimeConfig{}, err
	}
	options := config.Options{Root: root}
	if *development {
		options.Environment, err = config.DevelopmentEnvironment(root, nil)
		if err != nil {
			return runtimeConfig{}, err
		}
	}
	api, err := config.LoadAPI(options)
	return runtimeConfig{API: api}, err
}

func newLogger(value runtimeConfig) *slog.Logger {
	var level slog.Level
	_ = level.UnmarshalText([]byte(value.API.App.LogLevel))
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})).With("service", "api")
}

func newEcho(value runtimeConfig, logger *slog.Logger) *echo.Echo {
	e := newBareEcho()
	e.Use(requestLogger(logger), boundaryMiddleware())
	return e
}

func newExportEcho() *echo.Echo {
	e := newBareEcho()
	e.Use(boundaryMiddleware())
	return e
}

func newBareEcho() *echo.Echo {
	e := echo.New()
	e.HideBanner, e.HidePort = true, true
	e.HTTPErrorHandler = safeEchoError
	return e
}

func newAPI(e *echo.Echo) huma.API {
	c := huma.DefaultConfig("Zenith Public API", "0.1.0")
	c.OpenAPIPath, c.DocsPath, c.SchemasPath = "", "", ""
	c.CreateHooks = nil
	c.Info.Description = publicAPIOverview
	c.Tags = []*huma.Tag{{Name: "Organization", Description: "Identify the organization that owns the API key."}}
	c.Components = &huma.Components{SecuritySchemes: map[string]*huma.SecurityScheme{
		"bearerAuth": {
			Type: "http", Scheme: "bearer", BearerFormat: "Zenith API key",
			Description: "Send an organization API key in `Authorization: Bearer <api_key>`. The key selects the organization. Keep it on your server.",
		},
	}}
	return humaecho.NewV4(e, c)
}

func lifecycle(lc fx.Lifecycle, e *echo.Echo, api huma.API, value runtimeConfig, logger *slog.Logger, shutdown fx.Shutdowner) {
	registerDocs(e, api, value.API.API.Docs.BasicAuth)
	c := value.API.API
	s := &http.Server{Addr: net.JoinHostPort(c.Host, strconv.Itoa(c.Port)), Handler: e, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: time.Duration(c.ReadTimeoutSeconds) * time.Second, WriteTimeout: time.Duration(c.WriteTimeoutSeconds) * time.Second, IdleTimeout: time.Duration(c.IdleTimeoutSeconds) * time.Second, MaxHeaderBytes: 1 << 20}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", s.Addr)
			if err != nil {
				return errors.New("api listener unavailable")
			}
			go func() {
				if err := s.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Error("listener stopped unexpectedly")
					_ = shutdown.Shutdown(fx.ExitCode(1))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			bounded, cancel := context.WithTimeout(ctx, time.Duration(c.ShutdownTimeoutSeconds)*time.Second)
			defer cancel()
			if err := s.Shutdown(bounded); err != nil {
				_ = s.Close()
				return errors.New("api graceful shutdown deadline exceeded")
			}
			return nil
		},
	})
}

var Module = fx.Options(
	fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
	fx.Provide(loadConfig, newLogger, newEcho, newAPI),
	fx.Invoke(lifecycle),
)
var ExportModule = fx.Options(
	fx.WithLogger(func() fxevent.Logger { return fxevent.NopLogger }),
	fx.Provide(newExportEcho, newAPI),
)
