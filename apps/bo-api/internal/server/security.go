package server

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/vandordev/vkit-zenith/apps/bo-api/internal/security"
	"github.com/vandordev/vkit-zenith/packages/application/staff"
	db "github.com/vandordev/vkit-zenith/packages/db"
	"go.uber.org/fx"
)

func newBoundary(lc fx.Lifecycle, c runtimeConfig, e *echo.Echo, logger *slog.Logger) *security.Boundary {
	b := &security.Boundary{Clock: time.Now, Origin: c.API.BOOrigin, Development: c.API.Development, Secret: []byte(c.API.CSRFSecret), TrustedProxies: c.API.TrustedProxyCIDRs}
	e.Pre(b.Middleware)
	lc.Append(fx.Hook{OnStart: func(ctx context.Context) error {
		d := c.Database
		p, err := db.Open(ctx, db.Settings{DSN: d.DSN, MaxOpenConnections: d.MaxOpenConnections, ConnectionMaxLifetimeSeconds: d.ConnectionMaxLifetimeSeconds, ConnectionMaxIdleTimeSeconds: d.ConnectionMaxIdleTimeSeconds, ConnectTimeoutSeconds: d.ConnectTimeoutSeconds})
		if err != nil {
			logger.Error("BO database unavailable; check database availability and BO database configuration")
			return errors.New("BO database unavailable")
		}
		b.Pool = p
		b.Service = staff.NewService(p, time.Now, staff.NewPasswords())
		return nil
	}, OnStop: func(context.Context) error {
		if b.Pool != nil {
			b.Pool.Close()
		}
		return nil
	}})
	return b
}

// Offline composition registers schemas with nil dependencies, never executable
// fake usecases. The export Echo rejects all requests before route dispatch.
func exportBoundary() *security.Boundary { return &security.Boundary{} }
