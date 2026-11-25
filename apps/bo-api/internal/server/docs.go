package server

import (
	"crypto/sha256"
	"crypto/subtle"
	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v4"
	config "github.com/vandordev/vkit-zenith/packages/config/go/config"
	"net/http"
)

const scalarPage = `<!doctype html><html lang="en"><head><meta charset="utf-8"><title>Zenith BO API reference</title></head><body><script id="api-reference" data-url="/openapi.json"></script><script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.49.2"></script></body></html>`

func registerDocs(e *echo.Echo, api huma.API, auth config.BasicAuthConfig) {
	protect := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "no-store")
			if auth.Enabled {
				u, p, ok := c.Request().BasicAuth()
				a, b := sha256.Sum256([]byte(u)), sha256.Sum256([]byte(auth.Username))
				x, y := sha256.Sum256([]byte(p)), sha256.Sum256([]byte(auth.Password))
				valid := subtle.ConstantTimeCompare(a[:], b[:]) & subtle.ConstantTimeCompare(x[:], y[:])
				if !ok || valid != 1 {
					c.Response().Header().Set("WWW-Authenticate", `Basic realm="Zenith BO documentation", charset="UTF-8"`)
					return echo.NewHTTPError(http.StatusUnauthorized)
				}
			}
			return next(c)
		}
	}
	e.GET("/docs", func(c echo.Context) error { return c.HTML(200, scalarPage) }, protect)
	e.GET("/openapi.json", func(c echo.Context) error { return c.JSON(200, api.OpenAPI()) }, protect)
}
