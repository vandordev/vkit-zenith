package server

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
)

const DummyUploadLimit int64 = 10 << 20

func boundaryMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set(echo.HeaderCacheControl, "no-store")
			return next(c)
		}
	}
}

func safeEchoError(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}
	status := http.StatusInternalServerError
	var failure *echo.HTTPError
	if errors.As(err, &failure) && failure.Code >= 400 && failure.Code < 500 {
		status = failure.Code
	}
	c.Response().Header().Set(echo.HeaderContentType, "application/problem+json")
	_ = c.JSON(status, map[string]any{"type": "about:blank", "title": http.StatusText(status), "status": status, "detail": http.StatusText(status), "code": errorCode(status)})
}

func requestLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				c.Error(err)
			}
			logger.Info("request", "method", c.Request().Method, "path", c.Request().URL.Path, "status", c.Response().Status)
			return nil
		}
	}
}

func errorCode(status int) string {
	switch status {
	case http.StatusNotFound:
		return "not_found"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusBadRequest, http.StatusRequestEntityTooLarge, http.StatusMethodNotAllowed, http.StatusUnsupportedMediaType, http.StatusUnprocessableEntity:
		return "validation_failed"
	default:
		return "internal_error"
	}
}
