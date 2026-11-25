package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vandordev/vkit-zenith/apps/web-api/internal/identityhttp"
	"github.com/vandordev/vkit-zenith/packages/application/identity"
	config "github.com/vandordev/vkit-zenith/packages/config/go/config"
	"go.uber.org/fx"
)

func exportEcho() *echo.Echo { e := echo.New(); e.HideBanner = true; e.HidePort = true; return e }
func newAPI(e *echo.Echo) huma.API {
	return NewAPI(e)
}
func NewAPI(e *echo.Echo) huma.API {
	c := huma.DefaultConfig("Zenith Web API", "0.1.0")
	// Disable all automatic aliases and schema-link transformation. Only the
	// two explicit documentation endpoints below are exposed at runtime.
	c.OpenAPIPath, c.DocsPath, c.SchemasPath = "", "", ""
	c.CreateHooks = nil
	deadlines := &sync.Map{}
	e.Use(v1BodyLimit(deadlines))
	return boundedAPI{API: humaecho.NewV4(e, c), deadlines: deadlines}
}

type boundedAPI struct {
	huma.API
	deadlines *sync.Map
}

func (a boundedAPI) Adapter() huma.Adapter {
	return boundedAdapter{Adapter: a.API.Adapter(), deadlines: a.deadlines}
}

type boundedAdapter struct {
	huma.Adapter
	deadlines *sync.Map
}

func (a boundedAdapter) Handle(op *huma.Operation, handler func(huma.Context)) {
	if strings.HasPrefix(op.Path, "/v1/auth/") {
		if op.MaxBodyBytes == 0 || op.MaxBodyBytes > 16*1024 {
			op.MaxBodyBytes = 16 * 1024
		}
	} else if strings.HasPrefix(op.Path, "/v1/") && (op.MaxBodyBytes <= 0 || op.MaxBodyBytes >= 64*1024) {
		op.MaxBodyBytes = 64*1024 + 1
	}
	if strings.HasPrefix(op.Path, "/v1/") {
		timeout := op.BodyReadTimeout
		if timeout == 0 {
			timeout = 5 * time.Second
		}
		a.deadlines.Store(op.Method+" "+strings.NewReplacer("{", ":", "}", "").Replace(op.Path), timeout)
	}
	a.Adapter.Handle(op, handler)
}
func NewHTTP(logger *slog.Logger, origin string) *echo.Echo {
	return newEcho(runtimeConfig{Identity: config.IdentityConfig{Identity: config.IdentitySettings{AuthOrigin: origin}}}, logger)
}
func newEcho(value runtimeConfig, logger *slog.Logger) *echo.Echo {
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
	e.Use(IdentityBindings(value.Identity.Identity))
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{DisablePrintStack: true, LogErrorFunc: func(c echo.Context, _ error, _ []byte) error {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}}))
	e.Use(identityCORS(value.Identity.Identity.AuthOrigin))
	e.Use(middleware.BodyLimit("1M"))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			peer, err := identityhttp.ResolveSource(c.Request().RemoteAddr, c.Request().Header.Get(echo.HeaderXForwardedFor), value.Identity.Identity.TrustedProxyCIDRs)
			if err == nil {
				bindings := identityhttp.RequestBrowserBindings(c.Request().Context())
				ctx := identityhttp.WithRequestSource(c.Request().Context(), peer)
				ctx = identityhttp.WithBrowserBindings(ctx, bindings)
				c.SetRequest(c.Request().WithContext(ctx))
			}
			return next(c)
		}
	})
	return e
}

func v1BodyLimit(deadlines *sync.Map) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			if !strings.HasPrefix(path, "/v1/") {
				return next(c)
			}
			limit := int64(64 * 1024)
			if strings.HasPrefix(path, "/v1/auth/") {
				limit = 16 * 1024
			}
			request := c.Request()
			if request.ContentLength > limit {
				return echo.ErrStatusRequestEntityTooLarge
			}
			timeout := 5 * time.Second
			if value, ok := deadlines.Load(c.Request().Method + " " + c.Path()); ok {
				timeout = value.(time.Duration)
			}
			if timeout < 0 {
				_ = huma.SetReadDeadline(c.Response(), time.Time{})
			} else {
				_ = huma.SetReadDeadline(c.Response(), time.Now().Add(timeout))
				defer huma.SetReadDeadline(c.Response(), time.Time{})
			}
			body, err := io.ReadAll(io.LimitReader(request.Body, limit+1))
			if err != nil {
				var networkError net.Error
				if errors.As(err, &networkError) && networkError.Timeout() {
					return echo.NewHTTPError(http.StatusRequestTimeout)
				}
				return echo.NewHTTPError(http.StatusBadRequest)
			}
			if int64(len(body)) > limit {
				return echo.ErrStatusRequestEntityTooLarge
			}
			request.Body = io.NopCloser(bytes.NewReader(body))
			return next(c)
		}
	}
}

// IdentityBindings hashes only cookies trusted by the identity runtime.
func IdentityBindings(settings config.IdentitySettings) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			bindings := identityhttp.BrowserBindings{}
			sessionName, enrollmentName := "__Host-zenith_auth", "__Host-zenith_enrollment"
			if settings.Development {
				sessionName, enrollmentName = "zenith_auth_dev", "zenith_enrollment_dev"
			}
			if cookie, err := c.Cookie(sessionName); err == nil {
				bindings.Session, bindings.SessionSecret = identity.HashSecret(cookie.Value), cookie.Value
			}
			if cookie, err := c.Cookie(enrollmentName); err == nil {
				bindings.Enrollment = identity.HashSecret(cookie.Value)
			}
			c.SetRequest(c.Request().WithContext(identityhttp.WithBrowserBindings(c.Request().Context(), bindings)))
			return next(c)
		}
	}
}
func identityCORS(origin string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			isV1 := strings.HasPrefix(c.Request().URL.Path, "/v1/")
			isAuth := strings.HasPrefix(c.Request().URL.Path, "/v1/auth/")
			requestOrigin := c.Request().Header.Get(echo.HeaderOrigin)
			if isV1 {
				header := c.Response().Header()
				header.Set(echo.HeaderCacheControl, "no-store")
			}
			if isAuth {
				header := c.Response().Header()
				header.Set("Pragma", "no-cache")
				header.Set(echo.HeaderReferrerPolicy, "no-referrer")
			}
			if isV1 && requestOrigin != "" {
				if requestOrigin != origin {
					return echo.NewHTTPError(http.StatusForbidden)
				}
				h := c.Response().Header()
				h.Set(echo.HeaderAccessControlAllowOrigin, origin)
				h.Set(echo.HeaderAccessControlAllowCredentials, "true")
				h.Add(echo.HeaderVary, echo.HeaderOrigin)
				if c.Request().Method == http.MethodOptions {
					h.Set(echo.HeaderAccessControlAllowMethods, "GET, POST, PUT, PATCH, DELETE, OPTIONS")
					h.Set(echo.HeaderAccessControlAllowHeaders, "Content-Type, X-CSRF-Token")
					h.Set(echo.HeaderAccessControlMaxAge, "600")
					return c.NoContent(http.StatusNoContent)
				}
			}
			return next(c)
		}
	}
}

// IdentityCORS exposes the exact Auth origin policy to in-process acceptance;
// it does not construct a listener or resolve runtime configuration.
func IdentityCORS(origin string) echo.MiddlewareFunc { return identityCORS(origin) }
func safeError(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}
	status := http.StatusInternalServerError
	var failure *echo.HTTPError
	if errors.As(err, &failure) && failure.Code >= 400 && failure.Code < 600 {
		status = failure.Code
	}
	title := http.StatusText(status)
	c.Response().Header().Set(echo.HeaderContentType, "application/problem+json")
	_ = c.JSON(status, &huma.ErrorModel{Type: "about:blank", Title: title, Status: status, Detail: title})
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
