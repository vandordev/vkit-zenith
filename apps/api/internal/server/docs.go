package server

import (
	"crypto/sha256"
	"crypto/subtle"
	_ "embed"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v4"
	config "github.com/vandordev/vkit-zenith/packages/config/go/config"
)

//go:generate cp ../../../../packages/brand/assets/logo.svg favicon.svg
//go:embed favicon.svg
var docsFavicon []byte

const (
	docsTitle       = "Zenith Public API Reference"
	docsDescription = "Integrate your server with Zenith using organization API keys."
	scalarPage      = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>` + docsTitle + `</title>
  <meta name="description" content="` + docsDescription + `">
  <meta property="og:type" content="website">
  <meta property="og:site_name" content="Zenith">
  <meta property="og:title" content="` + docsTitle + `">
  <meta property="og:description" content="` + docsDescription + `">
  <link rel="icon" type="image/svg+xml" href="/favicon.svg">
</head>
<body>
  <script id="api-reference" data-url="/openapi.json"></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.49.2"></script>
</body>
</html>`
)

const publicAPIOverview = `Call this API from your server with an organization API key. Paths are relative to the origin serving this reference.

### Request and response conventions

- Send JSON with **Content-Type: application/json** and accept **application/json** and **application/problem+json**.
- Zenith resource IDs are UUID strings and timestamps are RFC 3339.
- Standard API errors use RFC 9457 Problem Details with a stable **code**. Use **code** and the HTTP status for error handling. Transport limits such as body size and read deadlines can return a problem without **code**.
- Responses carry **Cache-Control: no-store**.`

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
	e.Match([]string{http.MethodGet, http.MethodHead}, "/", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/docs")
	})
	e.Match([]string{http.MethodGet, http.MethodHead}, "/favicon.svg", func(c echo.Context) error {
		return c.Blob(http.StatusOK, "image/svg+xml", docsFavicon)
	})
	e.GET("/docs", func(c echo.Context) error { return c.HTML(http.StatusOK, scalarPage) }, protect)
	e.GET("/openapi.json", func(c echo.Context) error { return c.JSON(http.StatusOK, api.OpenAPI()) }, protect)
}
