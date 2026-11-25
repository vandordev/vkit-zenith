package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"go.uber.org/fx"
)

func exportEcho() *echo.Echo { e := echo.New(); e.HideBanner = true; e.HidePort = true; return e }
func offlineEcho() *echo.Echo {
	e := exportEcho()
	e.Pre(func(echo.HandlerFunc) echo.HandlerFunc { return func(c echo.Context) error { return c.NoContent(503) } })
	return e
}
func newAPI(e *echo.Echo) huma.API {
	c := huma.DefaultConfig("Zenith Backoffice API", "0.1.0")
	c.OpenAPIPath, c.DocsPath, c.SchemasPath = "", "", ""
	c.CreateHooks = nil
	return humaecho.NewV4(e, c)
}
func newEcho(_ runtimeConfig, logger *slog.Logger) *echo.Echo {
	return NewHTTP(logger)
}

// NewHTTP composes the real request stack without starting a listener or pool.
// Lifecycle and operation registration remain explicit callers' responsibilities.
func NewHTTP(logger *slog.Logger) *echo.Echo {
	e := exportEcho()
	e.HTTPErrorHandler = safeError
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			err := next(c)
			if err != nil {
				c.Error(err)
			}
			logger.Info("request", "method", c.Request().Method, "status", c.Response().Status, "durationMs", time.Since(start).Milliseconds())
			return nil
		}
	})
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{DisablePrintStack: true, LogErrorFunc: func(_ echo.Context, _ error, _ []byte) error { return echo.NewHTTPError(500) }}))
	e.Use(middleware.BodyLimit("1M"))
	return e
}
func safeError(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}
	status := 500
	var failure *echo.HTTPError
	if errors.As(err, &failure) && failure.Code >= 400 && failure.Code < 600 {
		status = failure.Code
	}
	c.Response().Header().Set(echo.HeaderContentType, "application/problem+json")
	_ = c.JSON(status, security.Problem(status))
}
func lifecycle(lc fx.Lifecycle, e *echo.Echo, api huma.API, value runtimeConfig, logger *slog.Logger, shutdown fx.Shutdowner, _ *security.Boundary) {
	c := value.API.BOAPI
	registerDocs(e, api, c.Docs.BasicAuth)
	s := &http.Server{Addr: net.JoinHostPort(c.Host, strconv.Itoa(c.Port)), Handler: e, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: time.Duration(c.ReadTimeoutSeconds) * time.Second, WriteTimeout: time.Duration(c.WriteTimeoutSeconds) * time.Second, IdleTimeout: time.Duration(c.IdleTimeoutSeconds) * time.Second, MaxHeaderBytes: 1 << 20}
	lc.Append(fx.Hook{OnStart: func(ctx context.Context) error {
		listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", s.Addr)
		if err != nil {
			logger.Error("BO listener unavailable; check configured host, port and whether the port is already in use")
			return errors.New("BO listener unavailable")
		}
		go func() {
			if err := s.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Error("listener stopped unexpectedly")
				_ = shutdown.Shutdown(fx.ExitCode(1))
			}
		}()
		logger.Info("BO API ready", "address", listener.Addr().String())
		return nil
	}, OnStop: func(ctx context.Context) error {
		bounded, cancel := context.WithTimeout(ctx, time.Duration(c.ShutdownTimeoutSeconds)*time.Second)
		defer cancel()
		if err := s.Shutdown(bounded); err != nil {
			_ = s.Close()
			return errors.New("BO graceful shutdown deadline exceeded")
		}
		return nil
	}})
}
