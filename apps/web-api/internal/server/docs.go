package server

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v4"
	config "github.com/vandordev/vkit-zenith/packages/config/go/config"
)

const scalarPage = `<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Zenith API reference</title></head><body><script id="api-reference" data-url="/openapi.json"></script><script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.49.2"></script></body></html>`

func registerDocs(e *echo.Echo, api huma.API, auth config.BasicAuthConfig) {
	protect := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Cache-Control", "no-store")
			if auth.Enabled {
				user, password, ok := c.Request().BasicAuth()
				actualUser, expectedUser := sha256.Sum256([]byte(user)), sha256.Sum256([]byte(auth.Username))
				actualPass, expectedPass := sha256.Sum256([]byte(password)), sha256.Sum256([]byte(auth.Password))
				validUser := subtle.ConstantTimeCompare(actualUser[:], expectedUser[:])
				validPass := subtle.ConstantTimeCompare(actualPass[:], expectedPass[:])
				if !ok || validUser&validPass != 1 {
					c.Response().Header().Set("WWW-Authenticate", `Basic realm="Zenith documentation", charset="UTF-8"`)
					return echo.NewHTTPError(http.StatusUnauthorized)
				}
			}
			return next(c)
		}
	}
	e.GET("/docs", func(c echo.Context) error { return c.HTML(http.StatusOK, scalarPage) }, protect)
	e.GET("/openapi.json", func(c echo.Context) error { return c.JSON(http.StatusOK, api.OpenAPI()) }, protect)
}
